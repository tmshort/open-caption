// Package httpapi wires HTTP routes (docs/04-protocol-and-api.md).
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/live"
)

// Server holds shared dependencies for handlers.
type Server struct {
	cfg     config.Config
	version string
	rt      *Runtime
	dev     bool
}

// New creates a Server with an empty runtime.
func New(cfg config.Config, version string) *Server {
	return &Server{cfg: cfg, version: version, rt: NewRuntime(cfg)}
}

// Runtime exposes the show runtime (for wiring and tests).
func (s *Server) Runtime() *Runtime { return s.rt }

// EnableDev turns on unauthenticated development endpoints (POST /dev/...).
// They exist only until the real operator API (M3) lands; never enable on a venue network.
func (s *Server) EnableDev() { s.dev = true }

// Handler returns the router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"version": s.version, "protocol": 1})
	})
	s.audienceRoutes(mux)
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
