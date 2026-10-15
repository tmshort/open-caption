// Command open-caption is an open-captioning server for community theater.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/httpapi"
	"github.com/tmshort/open-caption/internal/logbuf"
	"github.com/tmshort/open-caption/internal/netinfo"
	"github.com/tmshort/open-caption/internal/showfile"
)

var version = "dev"

func main() {
	var (
		dataDir   = flag.String("data", env("OPEN_CAPTION_DATA", ""), "data directory (default: <user config dir>/open-caption)")
		cfgPath   = flag.String("config", env("OPEN_CAPTION_CONFIG", ""), "config file (default: <data dir>/config.yaml)")
		httpAddr  = flag.String("http-addr", env("OPEN_CAPTION_HTTP_ADDR", ""), "HTTP listen address, e.g. :80")
		httpsAddr = flag.String("https-addr", env("OPEN_CAPTION_HTTPS_ADDR", ""), "HTTPS listen address, e.g. :443 (enables HTTPS)")
		logFile   = flag.String("log-file", env("OPEN_CAPTION_LOG", ""), "also append log lines to this file")
		noBrowser = flag.Bool("no-browser", false, "do not open the setup page in a browser at startup")
		showVer   = flag.Bool("version", false, "print version and exit")
		dev       = flag.Bool("dev", false, "DEVELOPMENT ONLY: enable unauthenticated /dev/* control endpoints")
		devShow   = flag.String("show-file", "", "with --dev: load this show file as the active show")
		devAsset  = flag.String("assets", "", "with --dev: directory of images for --show-file")
	)
	flag.Parse()
	if *showVer {
		fmt.Println("open-caption", version)
		return
	}

	// Logs go to the console, to a short in-memory history (for the support bundle), and optionally a file.
	logs := logbuf.New(1000)
	writers := []io.Writer{os.Stderr, logs}
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("cannot open --log-file: %v", err)
		}
		defer f.Close()
		writers = append(writers, f)
	}
	log.SetOutput(io.MultiWriter(writers...))

	dir := *dataDir
	if dir == "" {
		d, err := config.DefaultDataDir()
		if err != nil {
			log.Fatalf("cannot determine data directory: %v (use --data)", err)
		}
		dir = d
	}
	path := *cfgPath
	if path == "" {
		path = filepath.Join(dir, "config.yaml")
	}
	cfg, err := config.Load(path)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.DataDir == "" {
		cfg.DataDir = dir
	}
	if *httpAddr != "" {
		cfg.HTTP.Addr = *httpAddr
	}
	if *dev {
		// The --dev controls have no password, so they must never be reachable from the network.
		cfg.HTTP.Addr = loopbackOf(cfg.HTTP.Addr)
		cfg.HTTP.FallbackAddr = ""
		cfg.HTTPS.Enabled = false
	}
	if *httpsAddr != "" && !*dev {
		cfg.HTTPS.Enabled, cfg.HTTPS.Addr = true, *httpsAddr
	}
	log.Printf("config file: %s (data directory: %s)", cfg.Path, cfg.DataDir)
	if cfg.Generated {
		fmt.Printf("\nFirst run: created %s\n  admin password:    %s\n  operator password: %s\n(Keep these; they are stored in plain text in that file.)\n\n",
			cfg.Path, cfg.AdminPassword, cfg.OperatorPassword)
	}

	api := httpapi.New(cfg, version)
	api.SetLogSource(logs)
	defer api.Close()
	rt := api.Runtime()
	rt.EnablePersistence(filepath.Join(cfg.DataDir, "live.json"))
	if *dev {
		loadDevShow(api, cfg, *devShow, *devAsset)
	} else if api.BootActive() {
		log.Printf("loaded the active show")
	}
	if rt.Restore() {
		log.Printf("resumed the previous show position")
	}
	handler := api.Handler()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		if s, ok := <-sigs; ok {
			log.Printf("received %v; shutting down", s)
			stop()
		}
	}()
	var wg sync.WaitGroup
	var lis httpapi.Listeners
	serve := func(scheme string, l config.Listener, srvTLS *tls.Config) {
		ln, err := listen(l)
		if err != nil {
			log.Fatalf("listen %s: %v", scheme, err)
		}
		if scheme == "http" {
			lis.HTTPAddr, lis.HTTPConfigured = ln.Addr().String(), l.Addr
		} else {
			lis.HTTPSAddr, lis.HTTPSConfigured = ln.Addr().String(), l.Addr
		}
		srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, TLSConfig: srvTLS}
		for _, u := range urls(scheme, ln.Addr()) {
			log.Printf("audience/operator URL: %s", u)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			srv.Shutdown(sctx)
		}()
		go func() {
			var err error
			if srvTLS != nil {
				err = srv.ServeTLS(ln, "", "")
			} else {
				err = srv.Serve(ln)
			}
			if err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}

	serve("http", cfg.HTTP, nil)
	if cfg.HTTPS.Enabled {
		serve("https", config.Listener{Addr: cfg.HTTPS.Addr, FallbackAddr: cfg.HTTPS.FallbackAddr}, tlsConfig(cfg))
	}
	api.SetListeners(lis)
	log.Printf("open-caption %s running; operator page: /operator, setup page: /admin. Press Ctrl+C to stop.", version)
	if cfg.Defaults.OpenBrowser && !*noBrowser && !*dev {
		openBrowser(firstURL("http", lis.HTTPAddr) + "admin")
	}
	<-ctx.Done()
	log.Printf("shutting down (saving show position)...")
	wg.Wait()
	log.Printf("stopped")
}

func tlsConfig(cfg config.Config) *tls.Config {
	var cert tls.Certificate
	var err error
	if cfg.HTTPS.CertFile != "" {
		cert, err = tls.LoadX509KeyPair(cfg.HTTPS.CertFile, cfg.HTTPS.KeyFile)
	} else {
		host, _ := os.Hostname()
		cert, err = netinfo.EnsureSelfSigned(filepath.Join(cfg.DataDir, "tls"), host, netinfo.LocalIPs())
		log.Printf("HTTPS is using a self-signed certificate: browsers will show a warning until it is accepted")
	}
	if err != nil {
		log.Fatalf("https: %v", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
}

func urls(scheme string, addr net.Addr) []string {
	_, port, _ := net.SplitHostPort(addr.String())
	def := map[string]string{"http": "80", "https": "443"}[scheme]
	suffix := ""
	if port != def {
		suffix = ":" + port
	}
	out := []string{fmt.Sprintf("%s://localhost%s/", scheme, suffix)}
	for _, ip := range netinfo.LocalIPs() {
		out = append(out, fmt.Sprintf("%s://%s%s/", scheme, ip, suffix))
	}
	return out
}

func loadDevShow(api *httpapi.Server, cfg config.Config, showFilePath, assetsDir string) {
	api.EnableDev()
	log.Printf("WARNING: --dev enabled: /dev/* control endpoints are UNAUTHENTICATED; listening on this computer only (%s)", cfg.HTTP.Addr)
	if showFilePath == "" {
		return
	}
	b, err := os.ReadFile(showFilePath)
	if err != nil {
		log.Fatalf("cannot read the --show-file file: %v", err)
	}
	var names []string
	if assetsDir != "" {
		ents, _ := os.ReadDir(assetsDir)
		for _, e := range ents {
			names = append(names, e.Name())
		}
	}
	res := showfile.Parse(string(b), showfile.Options{Assets: names})
	for _, d := range res.Diagnostics {
		if d.Severity != showfile.SevInfo {
			log.Printf("show file line %d: %s: %s", d.Line, d.Severity, d.Message)
		}
	}
	api.Runtime().SetShow("dev", filepath.Base(showFilePath), assetsDir, res.Cues, httpapi.DefaultTheme(cfg.Defaults.IdleMessage))
	log.Printf("loaded %d cues from %s", len(res.Cues), showFilePath)
}

// listen binds the configured address, falling back (with a warning) if it is unavailable.
func listen(l config.Listener) (net.Listener, error) {
	ln, err := net.Listen("tcp", l.Addr)
	if err == nil {
		return ln, nil
	}
	if l.FallbackAddr == "" || l.FallbackAddr == l.Addr {
		return nil, fmt.Errorf("%w\n  (is another program, or another copy of open-caption, already using that port? try --http-addr :8081)", err)
	}
	log.Printf("WARNING: cannot listen on %s (%v); falling back to %s", l.Addr, err, l.FallbackAddr)
	return net.Listen("tcp", l.FallbackAddr)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// loopbackOf rewrites ":8080" or "0.0.0.0:8080" to "127.0.0.1:8080".
func loopbackOf(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		return "127.0.0.1:8080"
	}
	return "127.0.0.1:" + port
}

// firstURL returns the local URL (ending in "/") for a bound address.
func firstURL(scheme, bound string) string {
	_, port, err := net.SplitHostPort(bound)
	if err != nil || port == "" || (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		return scheme + "://localhost/"
	}
	return scheme + "://localhost:" + port + "/"
}

// openBrowser opens url in the default browser; failure is harmless (for example on a headless server).
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("could not open a browser (%v); open %s yourself", err, url)
		return
	}
	go cmd.Wait()
}
