package httpapi

import (
	"io/fs"
	"net/http"
	"sort"

	"github.com/tmshort/open-caption/docs"
	"github.com/tmshort/open-caption/internal/help"
	"github.com/tmshort/open-caption/web"
)

// guideTopics maps the guides in docs/ to help topic ids (the same mapping help.Aliases uses for links).
var guideTopics = help.Aliases

// helpRoutes serves the in-app help (REQ-037). Help is embedded in the binary, so it works offline.
// The pages need an admin or operator session; the audience never sees them.
func (s *Server) helpRoutes(mux *http.ServeMux) {
	topics, err := help.Load(web.Help, "help")
	if err != nil {
		panic("help: " + err.Error()) // embedded content; covered by tests
	}
	// The full guides ship inside the program too, so help never points at files a volunteer may not have.
	for file, id := range guideTopics {
		b, err := fs.ReadFile(docs.Guides, file)
		if err != nil {
			panic("help: " + err.Error())
		}
		t, err := help.FromMarkdown(id, "admin", string(b))
		if err != nil {
			panic("help: " + err.Error())
		}
		topics = append(topics, t)
	}
	sort.Slice(topics, func(i, j int) bool { return topics[i].ID < topics[j].ID })
	byID := map[string]help.Topic{}
	for _, t := range topics {
		byID[t.ID] = t
	}
	sub, _ := fs.Sub(web.HelpUI, "helpui")
	static := http.StripPrefix("/static/help/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /help", func(w http.ResponseWriter, r *http.Request) { servePage(w, sub, "index.html") })
	mux.HandleFunc("GET /static/help/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		static.ServeHTTP(w, r)
	})
	op := func(h http.HandlerFunc) http.HandlerFunc { return s.require(RoleOperator, h) }
	mux.HandleFunc("GET /api/help/topics", op(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"topics": topics})
	}))
	mux.HandleFunc("GET /api/help/search", op(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"hits": help.Search(topics, r.URL.Query().Get("q"))})
	}))
	mux.HandleFunc("GET /help/sample-script.md", op(func(w http.ResponseWriter, r *http.Request) {
		b, err := fs.ReadFile(web.Sample, "sample/sample-script.md")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="sample-script.md"`)
		w.Write(b)
	}))
	mux.HandleFunc("GET /help/{topic}", op(func(w http.ResponseWriter, r *http.Request) {
		t, ok := byID[r.PathValue("topic")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.URL.Query().Get("format") == "md" {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Write([]byte(t.Markdown))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(help.Render(t.Markdown)))
	}))
}
