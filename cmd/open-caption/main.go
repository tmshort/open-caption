// Command open-caption is an open-captioning server for community theater.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/httpapi"
)

var version = "dev"

func main() {
	var (
		dataDir  = flag.String("data", env("OPEN_CAPTION_DATA", ""), "data directory (default: <user config dir>/open-caption)")
		cfgPath  = flag.String("config", env("OPEN_CAPTION_CONFIG", ""), "config file (default: <data dir>/config.yaml)")
		httpAddr = flag.String("http-addr", env("OPEN_CAPTION_HTTP_ADDR", ""), "HTTP listen address, e.g. :80")
		showVer  = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()
	if *showVer {
		fmt.Println("open-caption", version)
		return
	}

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
	if cfg.Generated {
		fmt.Printf("\nFirst run: created %s\n  admin password:    %s\n  operator password: %s\n(Keep these; they are stored in plain text in that file.)\n\n",
			cfg.Path, cfg.AdminPassword, cfg.OperatorPassword)
	}

	srv := &http.Server{Handler: httpapi.New(cfg, version).Handler(), ReadHeaderTimeout: 10 * time.Second}
	ln, err := listen(cfg.HTTP)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("open-caption %s listening on http://%s/", version, ln.Addr())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(sctx)
	}()
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// listen binds the configured address, falling back (with a warning) if it is unavailable.
func listen(l config.Listener) (net.Listener, error) {
	ln, err := net.Listen("tcp", l.Addr)
	if err == nil {
		return ln, nil
	}
	if l.FallbackAddr == "" {
		return nil, err
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
