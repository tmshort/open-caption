// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"encoding/json"
	"github.com/tmshort/open-caption/internal/logx"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tmshort/open-caption/internal/hub"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/web"
)

const (
	heartbeat = 15 * time.Second
	// audienceCSP: same-origin only; no external requests (doc 06).
	audienceCSP = "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self'"
)

// viewerLimitLog keeps a crowd past the limit, all asking again every 2 seconds, from filling the log.
var viewerLimitLog logx.Throttle

func init() {
	// Not in Go's built-in table, and the system's table differs between computers (and Windows).
	for ext, typ := range map[string]string{".mp4": "video/mp4", ".webm": "video/webm", ".ico": "image/x-icon", ".woff2": "font/woff2", ".otf": "font/otf"} {
		mime.AddExtensionType(ext, typ)
	}
}

func (s *Server) audienceRoutes(mux *http.ServeMux) {
	sub, _ := fs.Sub(web.Audience, "audience")
	static := http.StripPrefix("/static/audience/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /standby-logo", s.standbyLogo)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", audienceCSP)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Vary", "Accept-Language")
		w.Header().Set("Cache-Control", "no-cache")
		b, _ := fs.ReadFile(sub, "index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(b)
	})
	mux.HandleFunc("GET /static/audience/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		static.ServeHTTP(w, r)
	})
	shared, _ := fs.Sub(web.Shared, "shared")
	sharedFS := http.StripPrefix("/static/shared/", http.FileServer(http.FS(shared)))
	mux.HandleFunc("GET /static/shared/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/shared/fonts/") {
			w.Header().Set("Cache-Control", "public, max-age=86400") // large and never changes between releases
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		sharedFS.ServeHTTP(w, r)
	})
	// Browsers ask for /favicon.ico on their own (and some ignore the SVG), so serve it at the root.
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		b, err := fs.ReadFile(shared, "favicon.ico")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write(b)
	})
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) { s.serveEvents(w, r, s.rt) })
	mux.HandleFunc("GET /assets/{show}/{file}", s.handleAsset)
}

// handleEvents streams audience events (docs/04). Subscribe first, then take
// the snapshot, then drop buffered frames the snapshot already covers; that
// ordering guarantees no event is missed or applied twice.
func (s *Server) serveEvents(w http.ResponseWriter, r *http.Request, rt *Runtime) {
	rc := http.NewResponseController(w)
	lang := rt.Negotiate(r.URL.Query().Get("lang"), r.Header.Get("Accept-Language"))
	hb := rt.HubFor(lang)
	cl := hb.SubscribeInfo(hub.ClientInfo{IP: clientIP(r), UserAgent: r.UserAgent(), Projector: r.URL.Query().Get("display") == "projector"})
	if cl == nil {
		viewerLimitLog.Warnf("refusing viewers: the limit of %d connected devices is reached (audience.max_clients); refused devices ask again every 2 seconds", s.cfg.Audience.MaxClients)
		w.Header().Set("Retry-After", "2")
		http.Error(w, "too many viewers", http.StatusServiceUnavailable)
		return
	}
	defer hb.Unsubscribe(cl)
	if rt == s.rt { // the show on air, not an administrator's rehearsal: it is counted for the audience report
		s.syncReport()
		projector := r.URL.Query().Get("display") == "projector"
		kind, _, _ := describeDevice(r.UserAgent(), false) // a projector is still a computer or a tablet: the report counts it by its own flag
		defer s.tracker.Connect(clientIP(r), r.UserAgent(), kind, lang, projector)()
	}
	w.Header().Set("Vary", "Accept-Language")

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	rc.SetWriteDeadline(time.Time{})

	snap := rt.State(lang)
	b, _ := json.Marshal(snap)
	w.Write([]byte("retry: 1000\n\n"))
	w.Write(formatSnapshot(snap.Seq, b))
	if rc.Flush() != nil {
		return
	}

	tick := time.NewTicker(heartbeat)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case f, ok := <-cl.C:
			if !ok {
				return // dropped as too slow, or hub closed: client reconnects
			}
			if f.Seq != 0 && f.Seq <= snap.Seq {
				continue
			}
			if _, err := w.Write(f.Data); err != nil || rc.Flush() != nil {
				return
			}
		case <-tick.C:
			// A real event (not a comment) so EventSource clients can watch for it.
			if _, err := w.Write([]byte("event: ping\ndata: {}\n\n")); err != nil || rc.Flush() != nil {
				return
			}
		}
	}
}

func formatSnapshot(seq int64, data []byte) []byte {
	return []byte("id: " + strconv.FormatInt(seq, 10) + "\nevent: " + live.EvSnapshot + "\ndata: " + string(data) + "\n\n")
}

// handleAsset serves images of the active show only.
func (s *Server) handleAsset(w http.ResponseWriter, r *http.Request) {
	show, file := r.PathValue("show"), r.PathValue("file")
	dir := s.rt.AssetsDir()
	if dir == "" || show != s.rt.ShowID() || file == "" || file != filepath.Base(file) ||
		strings.HasPrefix(file, ".") || strings.ContainsAny(file, `/\`) {
		http.NotFound(w, r)
		return
	}
	switch strings.ToLower(filepath.Ext(file)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, filepath.Join(dir, file))
}
