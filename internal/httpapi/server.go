// Package httpapi wires HTTP routes (docs/04-protocol-and-api.md).
package httpapi

import (
	"fmt"
	"net/http"

	"github.com/tmshort/open-caption/internal/config"
)

// Server holds shared dependencies for handlers.
type Server struct {
	cfg     config.Config
	version string
}

// New creates a Server.
func New(cfg config.Config, version string) *Server { return &Server{cfg: cfg, version: version} }

// Handler returns the router.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"version":%q,"protocol":1}`+"\n", s.version)
	})
	return mux
}
