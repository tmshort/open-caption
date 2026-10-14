package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/tmshort/open-caption/internal/netinfo"
	"github.com/tmshort/open-caption/internal/script"
	"github.com/tmshort/open-caption/internal/show"
	"github.com/tmshort/open-caption/web"
)

const (
	maxUploadBody = 48 << 20
	maxZipBody    = 64 << 20
	maxZipEntries = 500
)

// Listeners records what the server actually bound (set by main) for the network panel.
type Listeners struct {
	HTTPAddr, HTTPSAddr             string // actual bound addresses ("" = not listening)
	HTTPConfigured, HTTPSConfigured string // what config asked for
}

// SetListeners tells the setup UI which addresses are really in use.
func (s *Server) SetListeners(l Listeners) {
	s.lmu.Lock()
	s.listeners = l
	s.lmu.Unlock()
}

func themeToRuntime(t show.Theme) Theme {
	return Theme{BG: t.BG, FG: t.FG, Accent: t.Accent, Font: t.Font, ScalePct: t.ScalePct, Speakers: t.Speakers, IdleMessage: t.IdleMessage}
}

// ---- analysis of a script (what the editor shows) ----

// Missing names an image the script uses but that was not uploaded.
type Missing struct {
	Name  string `json:"name"`
	Lines []int  `json:"lines"`
}

// Analysis is the result of parsing a script against a show's images.
type Analysis struct {
	Language    string              `json:"language"`
	Cues        []CueView           `json:"cues"`
	Diagnostics []script.Diagnostic `json:"diagnostics"`
	Used        map[string][]int    `json:"used"`
	Missing     []Missing           `json:"missing"`
	Assets      []show.Asset        `json:"assets"`
}

func analyze(text string, assets []show.Asset) (Analysis, []script.Cue) {
	names := make([]string, 0, len(assets))
	have := map[string]bool{}
	for _, a := range assets {
		names = append(names, a.Name)
		have[a.Name] = true
	}
	res := script.Parse(text, script.Options{Assets: names})
	used := map[string][]int{}
	note := func(it script.Item, line int) {
		if it.Kind == "image" {
			used[it.Image] = append(used[it.Image], line)
		}
	}
	for _, c := range res.Cues {
		if c.Item != nil {
			note(*c.Item, c.Line)
		}
		for _, it := range c.Items {
			note(it, c.Line)
		}
	}
	var missing []Missing
	for name, lines := range used {
		if !have[name] {
			missing = append(missing, Missing{Name: name, Lines: lines})
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].Name < missing[j].Name })
	return Analysis{Language: res.Language, Cues: cueViews(res.Cues), Diagnostics: res.Diagnostics, Used: used,
		Missing: missing, Assets: assets}, res.Cues
}

// ---- loading shows into runtimes ----

type loaded struct {
	meta   show.Meta
	cues   []script.Cue
	assets string
}

func (s *Server) loadShow(id string) (loaded, error) {
	m, err := s.store.Get(id)
	if err != nil {
		return loaded{}, err
	}
	text, err := s.store.ReadScript(id, "")
	if err != nil {
		return loaded{}, err
	}
	_, cues := analyze(text, mustAssets(s.store, id))
	return loaded{meta: m, cues: cues, assets: s.store.AssetsDir(id)}, nil
}

func mustAssets(st *show.Store, id string) []show.Asset {
	a, _ := st.ListAssets(id)
	return a
}

// BootActive loads the show recorded in live.json at startup. It reports whether one was loaded.
func (s *Server) BootActive() bool {
	if s.store == nil {
		return false
	}
	id := s.rt.SavedShowID()
	if id == "" {
		return false
	}
	l, err := s.loadShow(id)
	if err != nil {
		return false
	}
	s.rt.SetShow(id, l.meta.Name, l.assets, l.cues, themeToRuntime(l.meta.Theme))
	return true
}

// ---- error mapping ----

func storeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, show.ErrNotFound):
		writeErr(w, http.StatusNotFound, "show not found")
	case errors.Is(err, show.ErrBadID), errors.Is(err, show.ErrBadName), errors.Is(err, show.ErrBadTheme), errors.Is(err, show.ErrBadAsset):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, show.ErrTooLarge):
		writeErr(w, http.StatusRequestEntityTooLarge, err.Error())
	default:
		writeErr(w, http.StatusInternalServerError, err.Error())
	}
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, show.MaxScript+(64<<10))).Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return false
	}
	return true
}

// ---- routes ----

func (s *Server) adminRoutes(mux *http.ServeMux) {
	sub, _ := fs.Sub(web.Admin, "admin")
	static := http.StripPrefix("/static/admin/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) { servePage(w, sub, "index.html") })
	mux.HandleFunc("GET /static/admin/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		static.ServeHTTP(w, r)
	})

	adm := func(h http.HandlerFunc) http.HandlerFunc {
		return s.require(RoleAdmin, func(w http.ResponseWriter, r *http.Request) {
			if s.store == nil {
				writeErr(w, http.StatusServiceUnavailable, "no data directory configured")
				return
			}
			h(w, r)
		})
	}
	mux.HandleFunc("GET /api/admin/ping", s.require(RoleAdmin, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}))
	mux.HandleFunc("GET /api/admin/shows", adm(s.listShows))
	mux.HandleFunc("POST /api/admin/shows", adm(s.createShow))
	mux.HandleFunc("POST /api/admin/shows/import", adm(s.importShow))
	mux.HandleFunc("GET /api/admin/shows/{id}", adm(s.getShow))
	mux.HandleFunc("PUT /api/admin/shows/{id}", adm(s.updateShow))
	mux.HandleFunc("DELETE /api/admin/shows/{id}", adm(s.deleteShow))
	mux.HandleFunc("GET /api/admin/shows/{id}/script", adm(s.getScript))
	mux.HandleFunc("PUT /api/admin/shows/{id}/script", adm(s.putScript))
	mux.HandleFunc("POST /api/admin/shows/{id}/script/validate", adm(s.validateScript))
	mux.HandleFunc("GET /api/admin/shows/{id}/assets", adm(s.listAssets))
	mux.HandleFunc("POST /api/admin/shows/{id}/assets", adm(s.uploadAssets))
	mux.HandleFunc("GET /api/admin/shows/{id}/assets/{file}", adm(s.getAsset))
	mux.HandleFunc("DELETE /api/admin/shows/{id}/assets/{file}", adm(s.deleteAsset))
	mux.HandleFunc("GET /api/admin/shows/{id}/export", adm(s.exportShow))
	mux.HandleFunc("POST /api/admin/shows/{id}/activate", adm(s.activateShow))
	mux.HandleFunc("GET /api/admin/network", adm(s.network))
	mux.HandleFunc("GET /api/admin/network/qr.png", adm(s.qr))
	mux.HandleFunc("GET /api/admin/settings", adm(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.cfg.Redacted())
	}))

	// Rehearsal: the operator API against a private, isolated runtime (docs/02, "Rehearsal").
	mux.HandleFunc("POST /api/admin/rehearsal/start", adm(s.startRehearsal))
	s.operatorAPI(mux, "/api/admin/rehearsal/", RoleAdmin, func(r *http.Request) *Runtime {
		if c, err := r.Cookie(cookieName); err == nil {
			return s.rehearsal(c.Value)
		}
		return nil
	})
	mux.HandleFunc("GET /api/admin/rehearsal/events", s.require(RoleAdmin, func(w http.ResponseWriter, r *http.Request) {
		rt := s.rehearsal(principal(r).Token)
		if rt == nil {
			writeErr(w, http.StatusNotFound, "no rehearsal is running")
			return
		}
		s.serveEvents(w, r, rt)
	}))
}

type showSummary struct {
	show.Meta
	Active bool `json:"active"`
	Live   bool `json:"live"`
}

func (s *Server) summary(m show.Meta) showSummary {
	active := s.rt.ShowID() == m.ID
	return showSummary{Meta: m, Active: active, Live: active && s.rt.Engine().Operator().Live}
}

func (s *Server) listShows(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.List()
	if err != nil {
		storeErr(w, err)
		return
	}
	out := make([]showSummary, 0, len(list))
	for _, m := range list {
		out = append(out, s.summary(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"shows": out, "activeId": s.rt.ShowID()})
}

func (s *Server) createShow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Sample bool   `json:"sample"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.Sample {
		m, err := s.createSample(req.Name)
		if err != nil {
			storeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, s.summary(m))
		return
	}
	m, err := s.store.Create(req.Name, "")
	if err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, s.summary(m))
}

func (s *Server) createSample(name string) (show.Meta, error) {
	if strings.TrimSpace(name) == "" {
		name = "Sample show: The Garden Party"
	}
	text, err := fs.ReadFile(web.Sample, "sample/sample-script.md")
	if err != nil {
		return show.Meta{}, err
	}
	m, err := s.store.Create(name, string(text))
	if err != nil {
		return m, err
	}
	ents, _ := fs.ReadDir(web.Sample, "sample/assets")
	for _, e := range ents {
		f, err := web.Sample.Open("sample/assets/" + e.Name())
		if err != nil {
			continue
		}
		_, err = s.store.SaveAsset(m.ID, e.Name(), f)
		f.Close()
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

func (s *Server) getShow(w http.ResponseWriter, r *http.Request) {
	m, err := s.store.Get(r.PathValue("id"))
	if err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.summary(m))
}

func (s *Server) updateShow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  *string     `json:"name"`
		Theme *show.Theme `json:"theme"`
	}
	if !decode(w, r, &req) {
		return
	}
	id := r.PathValue("id")
	m, err := s.store.Update(id, req.Name, req.Theme)
	if err != nil {
		storeErr(w, err)
		return
	}
	if s.rt.ShowID() == id {
		if req.Theme != nil {
			s.rt.SetTheme(themeToRuntime(m.Theme))
		}
		if req.Name != nil {
			s.rt.SetName(m.Name)
		}
	}
	writeJSON(w, http.StatusOK, s.summary(m))
}

func (s *Server) deleteShow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.rt.ShowID() == id {
		if s.rt.Engine().Operator().Live {
			writeErr(w, http.StatusConflict, "This show is live. Restart it from the operator page, or activate another show first.")
			return
		}
		s.rt.SetShow("", "", "", nil, DefaultTheme(s.cfg.Defaults.IdleMessage))
		s.rt.Flush()
	}
	if err := s.store.Delete(id); err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) scriptResponse(id, text string) (Analysis, []script.Cue) {
	return analyze(text, mustAssets(s.store, id))
}

func (s *Server) getScript(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	text, err := s.store.ReadScript(id, r.URL.Query().Get("lang"))
	if err != nil {
		storeErr(w, err)
		return
	}
	a, _ := s.scriptResponse(id, text)
	writeJSON(w, http.StatusOK, map[string]any{"text": text, "analysis": a})
}

func (s *Server) validateScript(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &req) {
		return
	}
	id := r.PathValue("id")
	if _, err := s.store.Get(id); err != nil {
		storeErr(w, err)
		return
	}
	a, _ := s.scriptResponse(id, req.Text)
	writeJSON(w, http.StatusOK, map[string]any{"analysis": a})
}

func (s *Server) putScript(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text    string `json:"text"`
		Confirm string `json:"confirm"`
	}
	if !decode(w, r, &req) {
		return
	}
	id := r.PathValue("id")
	m, err := s.store.Get(id)
	if err != nil {
		storeErr(w, err)
		return
	}
	active := s.rt.ShowID() == id
	if active && s.rt.Engine().Operator().Live && strings.TrimSpace(req.Confirm) != m.Name {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "This show is live. Type the show's name to confirm the change.",
			"needConfirm": true, "showName": m.Name})
		return
	}
	if err := s.store.WriteScript(id, r.URL.Query().Get("lang"), req.Text); err != nil {
		storeErr(w, err)
		return
	}
	a, cues := s.scriptResponse(id, req.Text)
	resp := map[string]any{"analysis": a, "saved": true}
	if active && r.URL.Query().Get("lang") == "" {
		changed, wasLive := s.rt.ReplaceCues(cues)
		resp["applied"], resp["countChanged"], resp["wasLive"] = true, changed, wasLive
		s.rt.Flush()
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- assets ----

func (s *Server) listAssets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.store.Get(id); err != nil {
		storeErr(w, err)
		return
	}
	a, err := s.store.ListAssets(id)
	if err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assets": a})
}

func (s *Server) uploadAssets(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.store.Get(id); err != nil {
		storeErr(w, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBody)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "That upload was too large. Images can be up to 10 MB each.")
		return
	}
	type fail struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	}
	var saved []show.Asset
	var failed []fail
	for _, fh := range r.MultipartForm.File["files"] {
		f, err := fh.Open()
		if err != nil {
			failed = append(failed, fail{fh.Filename, err.Error()})
			continue
		}
		a, err := s.store.SaveAsset(id, fh.Filename, f)
		f.Close()
		if err != nil {
			failed = append(failed, fail{fh.Filename, err.Error()})
			continue
		}
		saved = append(saved, a)
	}
	if len(saved) > 0 && s.rt.ShowID() == id {
		s.rt.refreshPreload()
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": saved, "failed": failed})
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	id, file := r.PathValue("id"), r.PathValue("file")
	if !show.ValidID(id) || file != filepath.Base(file) || strings.HasPrefix(file, ".") || strings.ContainsAny(file, `/\`) {
		http.NotFound(w, r)
		return
	}
	if _, err := show.CleanAssetName(file); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=60")
	http.ServeFile(w, r, filepath.Join(s.store.AssetsDir(id), file))
}

func (s *Server) deleteAsset(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteAsset(r.PathValue("id"), r.PathValue("file")); err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---- activate ----

func (s *Server) activateShow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Force bool `json:"force"`
	}
	json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req)
	id := r.PathValue("id")
	l, err := s.loadShow(id)
	if err != nil {
		storeErr(w, err)
		return
	}
	if s.rt.Engine().Operator().Live && !req.Force {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "A show is live right now. Activating a show restarts it from the beginning.",
			"live": true, "activeId": s.rt.ShowID()})
		return
	}
	s.rt.SetShow(id, l.meta.Name, l.assets, l.cues, themeToRuntime(l.meta.Theme))
	s.rt.Flush()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "activeId": id})
}

// ---- export / import ----

func (s *Server) exportShow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	m, err := s.store.Get(id)
	if err != nil {
		storeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, id))
	zw := zip.NewWriter(w)
	defer zw.Close()
	put := func(name string, data []byte) {
		if f, err := zw.Create(name); err == nil {
			f.Write(data)
		}
	}
	meta, _ := os.ReadFile(filepath.Join(filepath.Dir(s.store.AssetsDir(id)), "show.yaml"))
	put("show.yaml", meta)
	if text, err := s.store.ReadScript(id, m.Language); err == nil {
		put("script."+m.Language+".md", []byte(text))
	}
	for _, a := range mustAssets(s.store, id) {
		if b, err := os.ReadFile(filepath.Join(s.store.AssetsDir(id), a.Name)); err == nil {
			put("assets/"+a.Name, b)
		}
	}
}

func (s *Server) importShow(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxZipBody)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "That file is too large.")
		return
	}
	fh, _, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "choose a show .zip file")
		return
	}
	defer fh.Close()
	data, err := io.ReadAll(io.LimitReader(fh, maxZipBody))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "could not read the file")
		return
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(zr.File) > maxZipEntries {
		writeErr(w, http.StatusBadRequest, "That is not a show export (.zip).")
		return
	}
	read := func(f *zip.File, limit int64) ([]byte, error) {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, limit+1))
		if int64(len(b)) > limit {
			return nil, show.ErrTooLarge
		}
		return b, err
	}
	var metaYAML []byte
	scripts := map[string]string{}
	for _, f := range zr.File {
		name := path.Clean(f.Name)
		switch {
		case name == "show.yaml":
			metaYAML, _ = read(f, 64<<10)
		case strings.HasPrefix(name, "script.") && strings.HasSuffix(name, ".md") && !strings.Contains(name, "/"):
			if b, err := read(f, show.MaxScript); err == nil {
				scripts[strings.TrimSuffix(strings.TrimPrefix(name, "script."), ".md")] = string(b)
			}
		}
	}
	if metaYAML == nil || len(scripts) == 0 {
		writeErr(w, http.StatusBadRequest, "That zip is missing show.yaml or a script.")
		return
	}
	var meta show.Meta
	if err := yamlUnmarshal(metaYAML, &meta); err != nil || strings.TrimSpace(meta.Name) == "" {
		writeErr(w, http.StatusBadRequest, "That zip has an unreadable show.yaml.")
		return
	}
	m, err := s.store.Create(meta.Name+" (imported)", "")
	if err != nil {
		storeErr(w, err)
		return
	}
	for lang, text := range scripts {
		if lang == m.Language {
			s.store.WriteScript(m.ID, "", text)
		} else if err := s.store.WriteScript(m.ID, lang, text); err != nil {
			continue
		}
	}
	if th, err := meta.Theme.Normalize(s.cfg.Defaults.IdleMessage); err == nil {
		s.store.Update(m.ID, nil, &th)
	}
	skipped := 0
	for _, f := range zr.File {
		name := path.Clean(f.Name)
		if !strings.HasPrefix(name, "assets/") || strings.Count(name, "/") != 1 {
			continue
		}
		b, err := read(f, show.MaxAssetBytes)
		if err == nil {
			_, err = s.store.SaveAsset(m.ID, path.Base(name), bytes.NewReader(b))
		}
		if err != nil {
			skipped++
		}
	}
	m, _ = s.store.Get(m.ID)
	writeJSON(w, http.StatusCreated, map[string]any{"show": s.summary(m), "skippedFiles": skipped})
}

// ---- network panel ----

type urlInfo struct {
	URL    string `json:"url"`
	Kind   string `json:"kind"` // "http" | "https"
	Host   string `json:"host"`
	Public bool   `json:"public"` // reachable from other devices (not localhost)
}

func hostURL(scheme, host, bound, def string) string {
	_, port, err := net.SplitHostPort(bound)
	if err != nil || port == def || port == "" {
		return fmt.Sprintf("%s://%s/", scheme, host)
	}
	return fmt.Sprintf("%s://%s:%s/", scheme, host, port)
}

func (s *Server) network(w http.ResponseWriter, r *http.Request) {
	s.lmu.Lock()
	l := s.listeners
	s.lmu.Unlock()
	host, _ := os.Hostname()
	ips := netinfo.LocalIPs()
	var urls []urlInfo
	add := func(scheme, bound, def string) {
		if bound == "" {
			return
		}
		for _, ip := range ips {
			urls = append(urls, urlInfo{URL: hostURL(scheme, ip.String(), bound, def), Kind: scheme, Host: ip.String(), Public: true})
		}
		urls = append(urls, urlInfo{URL: hostURL(scheme, "localhost", bound, def), Kind: scheme, Host: "localhost"})
	}
	add("http", l.HTTPAddr, "80")
	add("https", l.HTTPSAddr, "443")

	var warnings []string
	if len(ips) == 0 {
		warnings = append(warnings, "No network address was found. Connect this computer to the venue's network, then reload.")
	}
	if l.HTTPAddr != "" && l.HTTPConfigured != "" && portOf(l.HTTPAddr) != portOf(l.HTTPConfigured) {
		warnings = append(warnings, fmt.Sprintf("Port %s was not available, so the server is using port %s. Audience addresses must include the port.", portOf(l.HTTPConfigured), portOf(l.HTTPAddr)))
	}
	if l.HTTPSAddr != "" && s.cfg.HTTPS.CertFile == "" {
		warnings = append(warnings, "HTTPS is using a self-signed certificate. Each device shows a security warning the first time. Plain http:// avoids this.")
	}
	writeJSON(w, http.StatusOK, map[string]any{"hostname": host, "ips": ipStrings(ips), "urls": urls, "warnings": warnings,
		"audience": s.rt.Hub.Count()})
}

func portOf(addr string) string {
	_, p, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return p
}

func ipStrings(ips []net.IP) []string {
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip.String())
	}
	return out
}

// qr renders a QR code for one of this server's own addresses.
func (s *Server) qr(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || len(raw) > 200 || !s.ownHost(u.Hostname()) {
		writeErr(w, http.StatusBadRequest, "that is not an address of this server")
		return
	}
	size := 512
	if n, err := strconv.Atoi(r.URL.Query().Get("size")); err == nil && n >= 128 && n <= 1024 {
		size = n
	}
	png, err := qrcode.Encode(raw, qrcode.Medium, size)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}

func (s *Server) ownHost(h string) bool {
	if h == "localhost" {
		return true
	}
	if name, _ := os.Hostname(); h == name {
		return true
	}
	for _, ip := range netinfo.LocalIPs() {
		if ip.String() == h {
			return true
		}
	}
	return false
}

// ---- rehearsal ----

const maxRehearsals = 8

func (s *Server) rehearsal(token string) *Runtime {
	s.rmu.Lock()
	defer s.rmu.Unlock()
	return s.rehearsals[token]
}

func (s *Server) dropRehearsal(token string) {
	s.rmu.Lock()
	rt := s.rehearsals[token]
	delete(s.rehearsals, token)
	s.rmu.Unlock()
	if rt != nil {
		rt.Close()
	}
}

func (s *Server) startRehearsal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShowID string `json:"showId"`
	}
	if !decode(w, r, &req) {
		return
	}
	l, err := s.loadShow(req.ShowID)
	if err != nil {
		storeErr(w, err)
		return
	}
	tok := principal(r).Token
	s.dropRehearsal(tok)
	s.rmu.Lock()
	if len(s.rehearsals) >= maxRehearsals {
		s.rmu.Unlock()
		writeErr(w, http.StatusTooManyRequests, "too many rehearsals are open")
		return
	}
	cfg := s.cfg
	cfg.Audience.MaxClients = 8
	rt := NewRuntime(cfg)
	rt.SetAssetBase("/api/admin/shows/" + url.PathEscape(req.ShowID) + "/assets/")
	rt.SetShow(req.ShowID, l.meta.Name, l.assets, l.cues, themeToRuntime(l.meta.Theme))
	rt.Claim(tok, true)
	s.rehearsals[tok] = rt
	s.rmu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "showId": req.ShowID})
}
