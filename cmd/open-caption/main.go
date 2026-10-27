// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Command open-caption is an open-captioning server for community theater.
package main

import (
	"context"
	"crypto/tls"
	"errors"
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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/tmshort/open-caption/internal/acme"
	"github.com/tmshort/open-caption/internal/check"
	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/httpapi"
	"github.com/tmshort/open-caption/internal/logbuf"
	"github.com/tmshort/open-caption/internal/logx"
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
		logLevel  = flag.String("log-level", env("OPEN_CAPTION_LOG_LEVEL", ""), "log detail: debug, info, warn or error (default: log.level in config.yaml, else info)")
		logFile   = flag.String("log-file", env("OPEN_CAPTION_LOG", ""), "also append log lines to this file")
		netIfs    = flag.String("network-interfaces", env("OPEN_CAPTION_NETWORK_INTERFACES", ""), "use only these network interfaces, comma separated, e.g. Ethernet or en0,en4 (default: network.interfaces in config.yaml, else all)")
		netRestr  = flag.Bool("network-restrict", envYes("OPEN_CAPTION_NETWORK_RESTRICT"), "also refuse connections made to addresses that are not offered (same as network.restrict_listening: yes)")
		noBrowser = flag.Bool("no-browser", false, "do not open the setup page in a browser at startup")
		showVer   = flag.Bool("version", false, "print version and exit")
		dev       = flag.Bool("dev", false, "DEVELOPMENT ONLY: enable unauthenticated /dev/* control endpoints")
		devShow   = flag.String("show-file", "", "with --dev: load this show file as the active show")
		devAsset  = flag.String("assets", "", "folder of images: for --check-show (default: the assets folder beside the show file), or with --dev for --show-file")
		checkShow = flag.String("check-show", "", "check a show file and print a report, like a lint (more files after it are translations of it); starts no server")
		hashPw    = flag.Bool("hash-password", false, "read a password from standard input and print it as a sha256: digest, for admin_password or operator_password in config.yaml; starts no server")
		strict    = flag.Bool("strict", false, "with --check-show: exit with status 1 on warnings too, not only on errors")
	)
	flag.Usage = usage
	flag.Parse()
	if *hashPw {
		fi, err := os.Stdin.Stat()
		os.Exit(hashPassword(os.Stdin, os.Stdout, os.Stderr, err == nil && fi.Mode()&os.ModeCharDevice != 0))
	}
	if *checkShow != "" {
		// The same analysis as Setup's editor (internal/check, internal/showfile); no config, no server.
		os.Exit(check.Run(os.Stdout, check.Options{Files: append([]string{*checkShow}, flag.Args()...), AssetsDir: *devAsset, Strict: *strict}))
	}
	if *showVer {
		fmt.Println("open-caption", version)
		fmt.Println("automatic certificates (ACME DNS-01), providers built in:", acmeProviders())
		return
	}

	// Logs go to the console, to a short in-memory history (for the support bundle), and optionally a file.
	logs := logbuf.New(1000)
	writers := []io.Writer{os.Stderr, logs}
	if *logFile != "" {
		f, err := logx.OpenFile(*logFile, logx.DefaultMaxBytes, logx.DefaultKeep) // starts a new file at 5 MB; keeps 3 old ones
		if err != nil {
			logx.Fatalf("cannot open --log-file: %v", err)
		}
		defer f.Close()
		writers = append(writers, f)
	}
	log.SetOutput(io.MultiWriter(writers...))

	dir := *dataDir
	if dir == "" {
		d, err := config.DefaultDataDir()
		if err != nil {
			logx.Fatalf("cannot determine data directory: %v (use --data)", err)
		}
		dir = d
	}
	path := *cfgPath
	if path == "" {
		path = filepath.Join(dir, "config.yaml")
	}
	cfg, err := config.Load(path)
	if err != nil {
		var pe *config.ProblemsError
		if errors.As(err, &pe) { // already says which file and what to fix: no timestamp noise
			fmt.Fprintln(os.Stderr, "config: "+err.Error()+"\nFix the file and start again.")
			os.Exit(1)
		}
		logx.Fatalf("config: %v", err)
	}
	level, overrides := cfg.Log.Level, ""
	if *logLevel != "" {
		if !strings.EqualFold(strings.TrimSpace(*logLevel), cfg.Log.Level) {
			overrides = fmt.Sprintf("--log-level (or OPEN_CAPTION_LOG_LEVEL) %s overrides log.level %s in config.yaml at every start; a change made in Setup lasts only until the program restarts", *logLevel, cfg.Log.Level)
		}
		level = *logLevel
	}
	lv, err := logx.Parse(level)
	if err != nil {
		logx.Fatalf("--log-level: %v", err)
	}
	logx.SetLevel(lv)
	cfg.Log.Level = lv.String()
	if overrides != "" {
		logx.Warnf("%s", overrides)
	}
	if cfg.DataDir == "" {
		cfg.DataDir = dir
	}
	if *httpAddr != "" {
		cfg.HTTP.Addr = *httpAddr
	}
	if *netIfs != "" {
		cfg.Network.Interfaces = splitList(*netIfs)
	}
	if *netRestr {
		cfg.Network.RestrictListening = true
	}
	if cfg.HTTPS.Enabled && cfg.HTTPS.ACME.Enabled {
		if !acme.Available() {
			logx.Fatalf("https.acme: this program was built without automatic certificates (ACME_PROVIDERS=none); see docs/https-certificate-guide.md")
		}
		if !acme.HasProvider(cfg.HTTPS.ACME.DNS.Provider) {
			logx.Fatalf("https.acme: the DNS provider %q is not built into this program (built with: %s); see docs/https-certificate-guide.md for how to build with others", cfg.HTTPS.ACME.DNS.Provider, acmeProviders())
		}
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
	logx.Infof("config file: %s (data directory: %s; log level: %s)", cfg.Path, cfg.DataDir, lv)
	for _, w := range cfg.Warnings() {
		logx.Warnf("config: %s", w)
	}
	if cfg.ViewerPassword != "" {
		logx.Infof("a viewer password is set, so the operator page can also be opened for watching only")
	}
	if cfg.Generated {
		fmt.Printf("\nFirst run: created %s\n  admin password:    %s\n  operator password: %s\n(Keep these; they are stored in plain text in that file.)\n\n",
			cfg.Path, cfg.AdminPassword, cfg.OperatorPassword)
	}

	api := httpapi.New(cfg, version)
	api.ApplyNetwork(cfg.Network) // before any address is printed or put in a certificate
	if *netIfs != "" || *netRestr {
		api.PinNetwork()
	}
	if *logLevel != "" {
		api.PinLogLevel(strings.ToLower(strings.TrimSpace(*logLevel)))
	}
	api.SetLogSource(logs)
	if !*dev {
		api.EnableStandby() // a show waits in standby until the operator starts it; --dev goes on air at once
	}
	defer api.Close()
	rt := api.Runtime()
	rt.EnablePersistence(filepath.Join(cfg.DataDir, "live.json"))
	if *dev {
		loadDevShow(api, cfg, *devShow, *devAsset)
	} else if api.BootActive() {
		logx.Infof("loaded the active show %q", rt.ShowName())
	}
	if rt.Restore() {
		logx.Infof("resumed the position of show %q from before the restart", rt.ShowName())
	}
	handler := api.Handler()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		if s, ok := <-sigs; ok {
			logx.Infof("received %v; shutting down", s)
			stop()
		}
	}()
	var wg sync.WaitGroup
	var lis httpapi.Listeners
	serve := func(scheme string, l config.Listener, srvTLS *tls.Config) {
		ln, err := listen(l)
		if err != nil {
			logx.Fatalf("listen %s: %v", scheme, err)
		}
		if scheme == "http" {
			lis.HTTPAddr, lis.HTTPConfigured = ln.Addr().String(), l.Addr
		} else {
			lis.HTTPSAddr, lis.HTTPSConfigured = ln.Addr().String(), l.Addr
		}
		ln = netinfo.LimitListener(ln, func(local net.Addr) { // while network.restrict_listening is on
			logx.Debugf("refused a %s connection to %s: that address is not offered (network.restrict_listening)", scheme, local)
		})
		srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second, TLSConfig: srvTLS, ErrorLog: httpErrorLog()}
		for _, u := range urls(scheme, ln.Addr()) {
			logx.Infof("audience/operator URL: %s", u)
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
				logx.Fatalf("%v", err)
			}
		}()
	}

	serve("http", cfg.HTTP, nil)
	if cfg.HTTPS.Enabled {
		serve("https", config.Listener{Addr: cfg.HTTPS.Addr, FallbackAddr: cfg.HTTPS.FallbackAddr}, tlsConfig(ctx, cfg, api))
	}
	api.SetListeners(lis)
	if !*dev { // the --dev controls are for this computer only
		api.ApplyMDNS(cfg.MDNS.Enabled, cfg.MDNS.Name)
	}
	logx.Infof("open-caption %s running; operator page: /operator, setup page: /admin. Press Ctrl+C to stop.", version)
	if cfg.Defaults.OpenBrowser && !*noBrowser && !*dev {
		openBrowser(firstURL("http", lis.HTTPAddr) + "admin")
	}
	<-ctx.Done()
	logx.Infof("shutting down (saving show position)...")
	wg.Wait()
	logx.Infof("stopped")
}

func tlsConfig(ctx context.Context, cfg config.Config, api *httpapi.Server) *tls.Config {
	if cfg.HTTPS.CertFile != "" {
		// the certificate files are watched: a renewed pair is picked up without a restart
		rl, err := netinfo.NewReloader(cfg.HTTPS.CertFile, cfg.HTTPS.KeyFile, logx.Infof, logx.Warnf)
		if err != nil {
			logx.Fatalf("HTTPS: %v", err)
		}
		return &tls.Config{GetCertificate: rl.GetCertificate, MinVersion: tls.VersionTLS12}
	}
	host, _ := os.Hostname()
	cert, err := netinfo.EnsureSelfSigned(filepath.Join(cfg.DataDir, "tls"), host, netinfo.LocalIPs())
	if err != nil {
		logx.Fatalf("HTTPS: %v", err)
	}
	if cfg.HTTPS.ACME.Enabled {
		// A certificate from a public CA, kept current in the background (docs/10). Until the first one arrives
		// (and if it never does) connections get the self-signed certificate, so HTTPS is never down.
		a := cfg.HTTPS.ACME
		iss, err := acme.NewIssuer(a, filepath.Join(cfg.DataDir, "tls", "acme"), logx.Warnf)
		if err != nil {
			logx.Fatalf("https.acme: %v", err)
		}
		m := acme.New(acme.Options{Domain: a.Domain, Dir: filepath.Join(cfg.DataDir, "tls", "acme"), Issuer: iss,
			RenewBefore: time.Duration(a.RenewBeforeDays) * 24 * time.Hour, Infof: logx.Infof, Warnf: logx.Warnf,
			Fallback: func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return &cert, nil }})
		api.SetCertificates(m)
		go m.Run(ctx)
		logx.Infof("HTTPS: automatic certificate for %s from %s with DNS provider %s", a.Domain, a.Directory, a.DNS.Provider)
		return &tls.Config{GetCertificate: m.GetCertificate, MinVersion: tls.VersionTLS12}
	}
	logx.Infof("HTTPS is using a self-signed certificate: browsers will show a warning until it is accepted (Setup's Share step says so too; docs/https-certificate-guide.md explains how to get a trusted one)")
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
}

// usage is what --help prints: the options (each may also be written with two dashes), the environment variables that
// stand in for them, and where the rest is documented. The man page (docs/open-caption.1) and the reference
// (docs/configuration-reference.md) describe the same options; a test keeps the three in step.
func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, `open-caption %s: open-captioning server for community theater

Usage: open-caption [options]

To check a show file without starting anything (a lint): open-caption --check-show show.md [translation.es.md ...]
To keep a password in config.yaml as a digest, not in plain text: printf '%%s' 'the password' | open-caption --hash-password

Run with no options to start with the defaults: the first run writes config.yaml with random passwords and prints
them, then prints the addresses to open (Setup at /admin, the operator page at /operator, the audience page at /).

Options (one or two leading dashes both work, for example -data or --data):
`, version)
	flag.PrintDefaults()
	fmt.Fprint(out, `
Environment variables (a command-line option wins over its variable, and both win over config.yaml):
  OPEN_CAPTION_CONFIG       same as --config
  OPEN_CAPTION_DATA         same as --data
  OPEN_CAPTION_HTTP_ADDR    same as --http-addr
  OPEN_CAPTION_HTTPS_ADDR   same as --https-addr
  OPEN_CAPTION_LOG_LEVEL    same as --log-level
  OPEN_CAPTION_LOG          same as --log-file
  OPEN_CAPTION_NETWORK_INTERFACES same as --network-interfaces
  OPEN_CAPTION_NETWORK_RESTRICT same as --network-restrict (yes, true, on or 1)

More: man open-caption (docs/open-caption.1), docs/configuration-reference.md (every setting in config.yaml),
docs/usage-guide.md, and https://github.com/tmshort/open-caption
`)
}

// acmeProviders is the list of DNS providers built into this program, for --version and the log.
func acmeProviders() string {
	if !acme.Available() {
		return "none (built without automatic certificates)"
	}
	return strings.Join(acme.Providers(), ", ")
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
	logx.Warnf("--dev enabled: /dev/* control endpoints are UNAUTHENTICATED; listening on this computer only (%s)", cfg.HTTP.Addr)
	if showFilePath == "" {
		return
	}
	b, err := os.ReadFile(showFilePath)
	if err != nil {
		logx.Fatalf("cannot read the --show-file file: %v", err)
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
		switch d.Severity {
		case showfile.SevError:
			logx.Errorf("show file line %d: %s", d.Line, d.Message)
		case showfile.SevWarning:
			logx.Warnf("show file line %d: %s", d.Line, d.Message)
		default:
			logx.Debugf("show file line %d: %s", d.Line, d.Message)
		}
	}
	api.Runtime().SetShow("dev", filepath.Base(showFilePath), assetsDir, res.Cues, httpapi.DefaultTheme(cfg.Defaults.IdleMessage))
	logx.Infof("loaded %d cues from %s", len(res.Cues), showFilePath)
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
	logx.Warnf("cannot listen on %s (%v); falling back to %s", l.Addr, err, l.FallbackAddr)
	return net.Listen("tcp", l.FallbackAddr)
}

// splitList turns "a, b,c" into its entries.
func splitList(s string) []string {
	var out []string
	for _, e := range strings.Split(s, ",") {
		if e = strings.TrimSpace(e); e != "" {
			out = append(out, e)
		}
	}
	return out
}

// envYes reports whether the environment variable is set to yes, true, on or 1.
func envYes(k string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(k))) {
	case "yes", "true", "on", "1":
		return true
	}
	return false
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
		logx.Warnf("could not open a browser (%v); open %s yourself", err, url)
		return
	}
	go cmd.Wait()
}
