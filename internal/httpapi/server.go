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
	certMu  sync.Mutex
	certs   CertificateManager
	cfg     config.Config
	version string
	rt      *Runtime
	auth    *auth
	dev     bool
	store   *show.Store

	adminAllow, operatorAllow config.AllowList // which computers may open Setup / the operator page
	mdns                      *mdnsManager     // the local network name (open-caption.local)
	standby                   bool             // shows start in standby (production); off for --dev
	pinnedLevel               string           // the --log-level / OPEN_CAPTION_LOG_LEVEL value, if one was given
	logs                      interface {
		String() string
		Lines() []string
	} // recent log lines for the support bundle (may be nil)
	started time.Time

	lmu       sync.Mutex
	listeners Listeners

	pwMu sync.Mutex // guards password changes (s.cfg passwords)

	rmu        sync.Mutex
	rehearsals map[string]*Runtime // admin session token -> isolated rehearsal

	quit      chan struct{}
	done      sync.WaitGroup
	closeOnce sync.Once
}

// New creates a Server with an empty runtime.
func New(cfg config.Config, version string) *Server {
	s := &Server{cfg: cfg, version: version, rt: NewRuntime(cfg), auth: newAuth(cfg.AdminPassword, cfg.OperatorPassword, sessionsPath(cfg)),
		rehearsals: map[string]*Runtime{}, started: time.Now(), quit: make(chan struct{}), mdns: newMDNSManager(cfg.MDNS)}
	// An unreadable list denies everything: never fall back to "anyone" by accident.
	s.adminAllow, _ = config.ParseAllowList(cfg.Access.AdminAllow)
	s.operatorAllow, _ = config.ParseAllowList(cfg.Access.OperatorAllow)
	if cfg.DataDir != "" {
		if st, err := show.NewStore(filepath.Join(cfg.DataDir, "shows"), showDefaults(cfg)); err == nil {
			s.store = st
		}
	}
	s.applyStandby()
	s.done.Add(1)
	go s.janitor()
	return s
}

// EnableStandby makes shows start in standby: nothing of a show reaches the audience until the operator
// starts it, and Live means on air. (--dev keeps the older start-at-once behavior.)
func (s *Server) EnableStandby() {
	s.standby = true
	s.rt.UseStandby()
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

// showDefaults are what a show uses where its show file says nothing: the built-in values as changed
// by config.yaml (config.Load starts from the built-in values, so cfg already has both layers).
func showDefaults(cfg config.Config) show.Defaults {
	l := cfg.Defaults.Look
	return show.Defaults{Language: cfg.Defaults.Language, IdleMessage: cfg.Defaults.IdleMessage, LoopSeconds: cfg.Defaults.LoopItemSeconds,
		ScrollbackLines: cfg.Audience.ScrollbackLines,
		Look:            show.Theme{BG: l.Background, FG: l.TextColor, Accent: l.SpeakerNameColor, Font: l.Font, ScalePct: l.TextSizePercent, Speakers: l.SpeakerNames}}
}

func sessionsPath(cfg config.Config) string {
	if cfg.DataDir == "" {
		return ""
	}
	return filepath.Join(cfg.DataDir, "sessions.json")
}

// Close stops background work and flushes live state.
func (s *Server) Close() {
	s.closeOnce.Do(func() {
		close(s.quit)
		s.done.Wait()
		s.stopMDNS()
		s.closeAll()
	})
}

func (s *Server) closeAll() {
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
// PinLogLevel records that the log level was given on the command line (or environment) and so overrides
// config.yaml at every start; Setup says so when the level is changed there.
func (s *Server) PinLogLevel(level string) { s.pinnedLevel = level }

func (s *Server) SetLogSource(l interface {
	String() string
	Lines() []string
}) {
	s.logs = l
}

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
	return s.originGuard(mux)
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
