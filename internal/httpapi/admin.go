package httpapi

import (
	"io/fs"
	"net/http"

	"github.com/tmshort/open-caption/web"
)

// adminRoutes serves the setup page. For now it only offers login; the setup UI arrives in M4.
func (s *Server) adminRoutes(mux *http.ServeMux) {
	sub, _ := fs.Sub(web.Admin, "admin")
	static := http.StripPrefix("/static/admin/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) { servePage(w, sub, "index.html") })
	mux.HandleFunc("GET /static/admin/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		static.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/admin/ping", s.require(RoleAdmin, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}))
}
