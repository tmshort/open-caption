// Package httpapi wires HTTP routes (docs/04-protocol-and-api.md).
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/show"
)

// Server holds shared dependencies for handlers.
type Server struct {
	cfg     config.Config
	version string
	rt      *Runtime
	auth    *auth
	dev     bool
	store   *show.Store
	logs    interface{ String() string } // recent log lines for the support bundle (may be nil)
	started time.Time

	lmu       sync.Mutex
	listeners Listeners

	rmu        sync.Mutex
	rehearsals map[string]*Runtime // admin session token -> isolated rehearsal

	quit chan struct{}
	done sync.WaitGroup
}

// New creates a Server with an empty runtime.
func New(cfg config.Config, version string) *Server {
	s := &Server{cfg: cfg, version: version, rt: NewRuntime(cfg), auth: newAuth(cfg.AdminPassword, cfg.OperatorPassword, sessionsPath(cfg)),
		rehearsals: map[string]*Runtime{}, started: time.Now(), quit: make(chan struct{})}
	if cfg.DataDir != "" {
		if st, err := show.NewStore(filepath.Join(cfg.DataDir, "shows"), cfg.Defaults.Language, cfg.Defaults.IdleMessage); err == nil {
			s.store = st
		}
	}
	s.done.Add(1)
	go s.janitor()
	return s
}

// janitor drops rehearsals whose login has expired or been revoked, and forgets expired logins.
func (s *Server) janitor() {
	defer s.done.Done()
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-s.quit:
			return
		case <-t.C:
			s.sweep()
		}
	}
}

func (s *Server) sweep() {
	s.auth.prune()
	s.rmu.Lock()
	var stale []string
	for tok := range s.rehearsals {
		if _, ok := s.auth.lookup(tok); !ok {
			stale = append(stale, tok)
		}
	}
	s.rmu.Unlock()
	for _, tok := range stale {
		s.dropRehearsal(tok)
	}
}

func sessionsPath(cfg config.Config) string {
	if cfg.DataDir == "" {
		return ""
	}
	return filepath.Join(cfg.DataDir, "sessions.json")
}

// Close stops background work and flushes live state.
func (s *Server) Close() {
	close(s.quit)
	s.done.Wait()
	s.rmu.Lock()
	rts := s.rehearsals
	s.rehearsals = map[string]*Runtime{}
	s.rmu.Unlock()
	for _, rt := range rts {
		rt.Close()
	}
	s.rt.Close()
}

// SetLogSource supplies recent log lines for the support bundle.
func (s *Server) SetLogSource(l interface{ String() string }) { s.logs = l }

// captiveAPI answers RFC 8908 clients (DHCP option 114): this network is not a
// captive portal, but here is the page to show people (docs/06).
func (s *Server) captiveAPI(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	w.Header().Set("Content-Type", "application/captive+json")
	w.Header().Set("Cache-Control", "private, max-age=60")
	json.NewEncoder(w).Encode(map[string]any{"captive": false, "user-portal-url": scheme + "://" + r.Host + "/"})
}

// Runtime exposes the show runtime (for wiring and tests).
func (s *Server) Runtime() *Runtime { return s.rt }

// EnableDev turns on unauthenticated development endpoints (POST /dev/...).
// They exist only until the real operator API (M3) lands; never enable on a venue network.
func (s *Server) EnableDev() { s.dev = true }

// Handler returns the router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /captive-api", s.captiveAPI)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"version": s.version, "protocol": 1})
	})
	s.audienceRoutes(mux)
	s.authRoutes(mux)
	s.operatorRoutes(mux)
	s.adminRoutes(mux)
	s.helpRoutes(mux)
	if s.dev {
		s.devRoutes(mux)
	}
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) devRoutes(mux *http.ServeMux) {
	cmd := func(name string, fn func(*live.Engine) (live.OperatorState, error)) {
		mux.HandleFunc("POST /dev/"+name, func(w http.ResponseWriter, r *http.Request) {
			st, err := s.rt.Do(fn)
			code := http.StatusOK
			if err != nil {
				code = http.StatusConflict
			}
			writeJSON(w, code, map[string]any{"state": st, "error": errString(err)})
		})
	}
	cmd("next", func(e *live.Engine) (live.OperatorState, error) { return e.Next(-1) })
	cmd("back", func(e *live.Engine) (live.OperatorState, error) { return e.Back(-1) })
	cmd("stop", func(e *live.Engine) (live.OperatorState, error) { return e.Stop(-1) })
	cmd("reset", func(e *live.Engine) (live.OperatorState, error) { return e.Reset(-1) })
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
