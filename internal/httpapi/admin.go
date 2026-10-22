// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmshort/open-caption/internal/logx"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/tmshort/open-caption/internal/check"
	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/netinfo"
	"github.com/tmshort/open-caption/internal/show"
	"github.com/tmshort/open-caption/internal/showfile"
	"github.com/tmshort/open-caption/web"
)

const (
	maxUploadBody  = 48 << 20
	maxZipBody     = 64 << 20
	maxZipEntries  = 500
	maxZipUnpacked = 200 << 20 // all the files of an import together, once unpacked (a small zip can hold a great deal)
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
	return Theme{BG: t.BG, FG: t.FG, Accent: t.Accent, Font: t.Font, ScalePct: t.ScalePct, Speakers: t.Speakers, IdleMessage: t.IdleMessage,
		IdleMessages: t.IdleMessages, LoopSeconds: t.LoopSeconds, ScrollbackLines: t.ScrollbackLines}
}

// ---- analysis of a show file (what the editor shows) ----

// Missing names an image the show file uses but that was not uploaded (shared with the command line: internal/check).
type Missing = check.Missing

// Analysis is the result of parsing a show file against a show's images.
type Analysis struct {
	Speakers    []string              `json:"speakers"`
	Language    string                `json:"language"`
	Cues        []CueView             `json:"cues"`
	Diagnostics []showfile.Diagnostic `json:"diagnostics"`
	Used        map[string][]int      `json:"used"`
	Missing     []Missing             `json:"missing"`
	Assets      []show.Asset          `json:"assets"`
}

func analyze(text string, assets []show.Asset, translation bool) (Analysis, []showfile.Cue) {
	names := make([]string, 0, len(assets))
	have := map[string]bool{}
	for _, a := range assets {
		names = append(names, a.Name)
		have[a.Name] = true
	}
	res := showfile.Parse(text, showfile.Options{Assets: names, Translation: translation})
	used := check.ImageUse(res.Cues)
	missing := check.MissingImages(used, have)
	return Analysis{Speakers: res.Speakers, Language: res.Language, Cues: cueViews(res.Cues), Diagnostics: res.Diagnostics, Used: used,
		Missing: missing, Assets: assets}, res.Cues
}

// ---- loading shows into runtimes ----

type loaded struct {
	meta    show.Meta
	primary LangFile
	others  []LangFile // translations whose structure matches the primary
	assets  string
	// skipped are translations left out because their structure does not match the primary's.
	skipped map[string][]showfile.Diagnostic
}

func (s *Server) loadShow(id string) (loaded, error) {
	m, err := s.store.Get(id)
	if err != nil {
		return loaded{}, err
	}
	codes, err := s.store.Languages(id)
	if err != nil {
		return loaded{}, err
	}
	assets := mustAssets(s.store, id)
	l := loaded{meta: m, assets: s.store.AssetsDir(id), skipped: map[string][]showfile.Diagnostic{}}
	var primaryRes showfile.Result
	for i, code := range codes {
		text, err := s.store.ReadShowFile(id, code)
		if err != nil {
			return loaded{}, err
		}
		res := parseWith(text, assets, i > 0)
		name := res.LanguageName
		if name == "" {
			name = showfile.LanguageName(code)
		}
		if i == 0 {
			primaryRes = res
			l.primary = LangFile{Code: code, Name: name, Cues: res.Cues}
			continue
		}
		if issues := showfile.CompareStructure(primaryRes.Cues, res.Cues, l.primary.Name); len(issues) > 0 {
			l.skipped[code] = issues // not offered to the audience until it is fixed
			continue
		}
		l.others = append(l.others, LangFile{Code: code, Name: name, Cues: res.Cues})
	}
	return l, nil
}

func parseWith(text string, assets []show.Asset, translation bool) showfile.Result {
	names := make([]string, 0, len(assets))
	for _, a := range assets {
		names = append(names, a.Name)
	}
	return showfile.Parse(text, showfile.Options{Assets: names, Translation: translation})
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
	s.rt.SetShowLangs(id, l.meta.Name, l.assets, l.primary, l.others, themeToRuntime(l.meta.Theme))
	return true
}

// ---- error mapping ----

func storeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, show.ErrNotFound):
		writeErr(w, http.StatusNotFound, "show not found")
	case errors.Is(err, show.ErrBadID), errors.Is(err, show.ErrBadName), errors.Is(err, show.ErrBadTheme), errors.Is(err, show.ErrBadAsset), errors.Is(err, show.ErrBadLabel):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, show.ErrTooLarge):
		writeErr(w, http.StatusRequestEntityTooLarge, err.Error())
	default:
		writeErr(w, http.StatusInternalServerError, err.Error())
	}
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, show.MaxShowFile+(64<<10))).Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request")
		return false
	}
	return true
}

// ---- routes ----

func (s *Server) adminRoutes(mux *http.ServeMux) {
	sub, _ := fs.Sub(web.Admin, "admin")
	static := http.StripPrefix("/static/admin/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /admin", s.gate(RoleAdmin, func(w http.ResponseWriter, r *http.Request) { servePage(w, sub, "index.html") }))
	mux.HandleFunc("GET /static/admin/", s.gate(RoleAdmin, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		static.ServeHTTP(w, r)
	}))

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
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "role": principal(r).Role, "auto": principal(r).Auto})
	}))
	mux.HandleFunc("GET /api/admin/shows", adm(s.listShows))
	mux.HandleFunc("POST /api/admin/shows", adm(s.createShow))
	mux.HandleFunc("POST /api/admin/shows/import", adm(s.importShow))
	mux.HandleFunc("GET /api/admin/shows/{id}", adm(s.getShow))
	mux.HandleFunc("PUT /api/admin/shows/{id}", adm(s.updateShow))
	mux.HandleFunc("POST /api/admin/shows/{id}/duplicate", adm(s.duplicateShow))
	mux.HandleFunc("DELETE /api/admin/shows/{id}", adm(s.deleteShow))
	mux.HandleFunc("DELETE /api/admin/reports", adm(s.reportDelete))
	mux.HandleFunc("GET /api/admin/shows/{id}/showfile", adm(s.getShowFile))
	mux.HandleFunc("PUT /api/admin/shows/{id}/showfile", adm(s.putShowFile))
	mux.HandleFunc("GET /api/admin/shows/{id}/showfile/download", adm(s.downloadShowFile))
	mux.HandleFunc("GET /api/admin/shows/{id}/showfile/history", adm(func(w http.ResponseWriter, r *http.Request) {
		v, err := s.store.History(r.PathValue("id"), r.URL.Query().Get("lang"))
		if err != nil {
			storeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"versions": v})
	}))
	mux.HandleFunc("GET /api/admin/shows/{id}/showfile/history/{version}", adm(func(w http.ResponseWriter, r *http.Request) {
		text, err := s.store.ReadHistory(r.PathValue("id"), r.URL.Query().Get("lang"), r.PathValue("version"))
		if err != nil {
			storeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"text": text})
	}))
	mux.HandleFunc("PUT /api/admin/shows/{id}/showfile/history/{version}/label", adm(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Label string `json:"label"`
		}
		if !decode(w, r, &req) {
			return
		}
		if err := s.store.LabelVersion(r.PathValue("id"), r.URL.Query().Get("lang"), r.PathValue("version"), req.Label); err != nil {
			storeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}))
	mux.HandleFunc("POST /api/admin/shows/{id}/showfile/history", adm(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Label string `json:"label"`
		}
		if !decode(w, r, &req) {
			return
		}
		v, err := s.store.KeepVersion(r.PathValue("id"), r.URL.Query().Get("lang"), req.Label)
		if err != nil {
			storeErr(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, v)
	}))
	mux.HandleFunc("POST /api/admin/shows/{id}/showfile/validate", adm(s.validateShowFile))
	mux.HandleFunc("GET /api/admin/shows/{id}/languages", adm(s.listLanguages))
	mux.HandleFunc("POST /api/admin/shows/{id}/languages", adm(s.addLanguage))
	mux.HandleFunc("DELETE /api/admin/shows/{id}/languages/{lang}", adm(s.deleteLanguage))
	mux.HandleFunc("GET /api/admin/shows/{id}/assets", adm(s.listAssets))
	mux.HandleFunc("POST /api/admin/shows/{id}/assets", adm(s.uploadAssets))
	mux.HandleFunc("GET /api/admin/shows/{id}/assets/{file}", adm(s.getAsset))
	mux.HandleFunc("DELETE /api/admin/shows/{id}/assets/{file}", adm(s.deleteAsset))
	mux.HandleFunc("GET /api/admin/shows/{id}/export", adm(s.exportShow))
	mux.HandleFunc("POST /api/admin/shows/{id}/archive", adm(s.archiveShow))
	mux.HandleFunc("POST /api/admin/shows/{id}/unarchive", adm(s.unarchiveShow))
	mux.HandleFunc("POST /api/admin/shows/{id}/activate", adm(s.activateShow))
	mux.HandleFunc("GET /api/admin/network", adm(s.network))
	mux.HandleFunc("GET /api/admin/network/qr.png", adm(s.qr))
	mux.HandleFunc("GET /api/admin/support-bundle", adm(s.supportBundle))
	mux.HandleFunc("GET /api/admin/settings", adm(s.settings))
	mux.HandleFunc("PUT /api/admin/settings/passwords", adm(s.changePasswords))
	mux.HandleFunc("PUT /api/admin/settings/log-level", adm(s.setLogLevel))
	mux.HandleFunc("PUT /api/admin/settings/mdns", adm(s.setMDNS))
	mux.HandleFunc("PUT /api/admin/settings/network", adm(s.setNetwork))
	mux.HandleFunc("PUT /api/admin/settings/standby", adm(s.setStandby))
	mux.HandleFunc("GET /api/admin/settings/certificate", adm(s.certificateInfo))
	mux.HandleFunc("POST /api/admin/settings/certificate/renew", adm(s.renewCertificate))
	mux.HandleFunc("POST /api/admin/settings/standby-logo", adm(s.setStandbyLogo))
	mux.HandleFunc("DELETE /api/admin/settings/standby-logo", adm(s.deleteStandbyLogo))
	mux.HandleFunc("GET /api/admin/log", adm(s.recentLog))
	mux.HandleFunc("GET /api/admin/settings/operator-password", adm(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		s.pwMu.Lock()
		defer s.pwMu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"password": s.cfg.OperatorPassword})
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
	Fresh  bool `json:"fresh"` // the active show is in standby with nothing sent: restarting it would change nothing
}

func (s *Server) summary(m show.Meta) showSummary {
	active := s.rt.ShowID() == m.ID
	var op live.OperatorState
	if active {
		op = s.rt.Engine().Operator()
	}
	return showSummary{Meta: m, Active: active, Live: active && op.Live, Fresh: active && op.AtStart}
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

func (s *Server) createSample(req string) (show.Meta, error) {
	text, err := fs.ReadFile(web.Sample, "sample/sample-show.md")
	if err != nil {
		return show.Meta{}, err
	}
	m, err := s.store.Create(req, string(text)) // an empty name keeps the sample's own title
	if err != nil {
		return m, err
	}
	// The sample ships with a Spanish translation to show how languages work.
	if es, err := fs.ReadFile(web.Sample, "sample/sample-show.es.md"); err == nil {
		if err := s.store.WriteShowFile(m.ID, "es", string(es)); err != nil {
			return m, err
		}
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
	s.applyActive(id)
	writeJSON(w, http.StatusOK, s.summary(m))
}

// archiveShow puts a show away: kept in full, but listed apart in Setup and unable to go on air.
func (s *Server) archiveShow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.rt.ShowID() == id {
		writeErr(w, http.StatusConflict, "This show is the one on air. Activate another show (or close this one) before archiving it.")
		return
	}
	m, err := s.store.SetArchived(id, true)
	if err != nil {
		storeErr(w, err)
		return
	}
	logx.Infof("show %q archived", m.Name)
	writeJSON(w, http.StatusOK, s.summary(m))
}

func (s *Server) unarchiveShow(w http.ResponseWriter, r *http.Request) {
	m, err := s.store.SetArchived(r.PathValue("id"), false)
	if err != nil {
		storeErr(w, err)
		return
	}
	logx.Infof("show %q restored from the archive", m.Name)
	writeJSON(w, http.StatusOK, s.summary(m))
}

func (s *Server) duplicateShow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req)
	m, err := s.store.Duplicate(r.PathValue("id"), req.Name)
	if err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, s.summary(m))
}

func (s *Server) deleteShow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.rt.ShowID() == id {
		if s.rt.Engine().Operator().Live {
			writeErr(w, http.StatusConflict, "This show is on air. Restart it from the Run page, or activate another show first.")
			return
		}
		s.rt.SetShow("", "", "", nil, DefaultTheme(s.cfg.Defaults.IdleMessage))
		s.rt.Flush()
	}
	if err := s.store.Delete(id); err != nil {
		storeErr(w, err)
		return
	}
	logx.Infof("show %s deleted", id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// showFileResponse analyses one language's show file text. For a translation it also reports whether the
// structure matches the main language's show file (as errors, listed first).
func (s *Server) showFileResponse(id, lang, text string) (Analysis, []showfile.Cue) {
	assets := mustAssets(s.store, id)
	m, err := s.store.Get(id)
	translation := err == nil && lang != "" && lang != m.Language
	a, cues := analyze(text, assets, translation)
	if !translation {
		if !check.HasTitle(text) {
			a.Diagnostics = append(a.Diagnostics, showfile.Diagnostic{Line: 1, Severity: showfile.SevInfo,
				Message: "This show has no title, so it is listed as \"Untitled show\". Use Rename in the Shows step, or add a [settings] block with a title: line."})
		}
		return a, cues
	}
	ptext, _ := s.store.ReadShowFile(id, m.Language)
	pres := parseWith(ptext, assets, false)
	pname := pres.LanguageName
	if pname == "" {
		pname = showfile.LanguageName(m.Language)
	}
	if issues := showfile.CompareStructure(pres.Cues, cues, pname); len(issues) > 0 {
		a.Diagnostics = append(issues, a.Diagnostics...)
	}
	return a, cues
}

// fileVersion identifies a show file's text, so two people editing the same show file are noticed.
func fileVersion(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:8])
}

func (s *Server) getShowFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	text, err := s.store.ReadShowFile(id, r.URL.Query().Get("lang"))
	if err != nil {
		storeErr(w, err)
		return
	}
	a, _ := s.showFileResponse(id, r.URL.Query().Get("lang"), text)
	writeJSON(w, http.StatusOK, map[string]any{"text": text, "version": fileVersion(text), "analysis": a})
}

// downloadShowFile sends the show file itself (not a .zip) for editing in the preparer's own editor; the
// editor's Upload a show file puts it back. ?lang= picks a translation.
func (s *Server) downloadShowFile(w http.ResponseWriter, r *http.Request) {
	id, lang := r.PathValue("id"), r.URL.Query().Get("lang")
	text, err := s.store.ReadShowFile(id, lang)
	if err != nil {
		storeErr(w, err)
		return
	}
	name := id + ".md"
	if lang != "" {
		name = id + "." + lang + ".md"
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	// Many Windows editors (Notepad before Windows 10 1903) need the mark to know a file is UTF-8, and
	// the parser ignores it, so it is harmless everywhere else.
	io.WriteString(w, "\ufeff"+strings.ReplaceAll(text, "\n", "\r\n"))
}

func (s *Server) validateShowFile(w http.ResponseWriter, r *http.Request) {
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
	a, _ := s.showFileResponse(id, r.URL.Query().Get("lang"), req.Text)
	writeJSON(w, http.StatusOK, map[string]any{"analysis": a})
}

func (s *Server) putShowFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text    string `json:"text"`
		Confirm string `json:"confirm"`
		// BaseVersion is the version the editor was based on. If the saved show file has changed since
		// (another person saved), the save is refused so nobody's work is silently lost.
		BaseVersion string `json:"baseVersion"`
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
	if req.BaseVersion != "" {
		cur, _ := s.store.ReadShowFile(id, r.URL.Query().Get("lang"))
		if fileVersion(cur) != req.BaseVersion && cur != req.Text {
			writeJSON(w, http.StatusConflict, map[string]any{"conflict": true, "text": cur, "version": fileVersion(cur),
				"error": "This show file was changed by someone else after you opened it."})
			return
		}
	}
	active := s.rt.ShowID() == id
	if lang := r.URL.Query().Get("lang"); active && (lang == "" || lang == m.Language) {
		if st := showfile.ReadSettings(req.Text, false); st.Archived != nil && *st.Archived {
			writeErr(w, http.StatusConflict, "This show is the one on air, so it cannot be archived. Activate another show first, or remove the \"archived\" line.")
			return
		}
	}
	if active && s.rt.Engine().Operator().Live && strings.TrimSpace(req.Confirm) != m.Name {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "This show is on air. Type the show's name to confirm the change.",
			"needConfirm": true, "showName": m.Name})
		return
	}
	if err := s.store.WriteShowFile(id, r.URL.Query().Get("lang"), req.Text); err != nil {
		storeErr(w, err)
		return
	}
	a, _ := s.showFileResponse(id, r.URL.Query().Get("lang"), req.Text)
	resp := map[string]any{"analysis": a, "saved": true, "version": fileVersion(req.Text)}
	if active {
		if changed, wasLive, skipped, ok := s.applyActive(id); ok {
			resp["applied"], resp["countChanged"], resp["wasLive"] = true, changed, wasLive
			if len(skipped) > 0 {
				resp["skippedLanguages"] = skipped
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// applyActive pushes a changed show to the audience if it is the show on air: the show files of every
// language, the look and other settings, and the title all come from the show files.
func (s *Server) applyActive(id string) (countChanged, wasLive bool, skipped []string, applied bool) {
	if s.rt.ShowID() != id {
		return
	}
	l, err := s.loadShow(id)
	if err != nil {
		return
	}
	s.rt.SetTheme(themeToRuntime(l.meta.Theme)) // first: engines created next use the new loop time and scrollback size
	s.rt.SetName(l.meta.Name)
	countChanged, wasLive = s.rt.ReplaceLangs(l.primary, l.others)
	s.rt.Flush()
	return countChanged, wasLive, sortedKeys(l.skipped), true
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ---- languages ----

type langSummary struct {
	Code      string                `json:"code"`
	Name      string                `json:"name"`
	Primary   bool                  `json:"primary"`
	Cues      int                   `json:"cues"`
	Errors    int                   `json:"errors"`
	Warnings  int                   `json:"warnings"`
	Structure []showfile.Diagnostic `json:"structure"` // why this translation is not offered to the audience (empty = fine)
	Available bool                  `json:"available"`
}

func (s *Server) listLanguages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	codes, err := s.store.Languages(id)
	if err != nil {
		storeErr(w, err)
		return
	}
	assets := mustAssets(s.store, id)
	var out []langSummary
	var primary showfile.Result
	primaryName := ""
	for i, code := range codes {
		text, _ := s.store.ReadShowFile(id, code)
		res := parseWith(text, assets, i > 0)
		sum := langSummary{Code: code, Name: res.LanguageName, Primary: i == 0, Cues: len(res.Cues), Structure: []showfile.Diagnostic{}}
		if sum.Name == "" {
			sum.Name = showfile.LanguageName(code)
		}
		for _, d := range res.Diagnostics {
			switch d.Severity {
			case showfile.SevError:
				sum.Errors++
			case showfile.SevWarning:
				sum.Warnings++
			}
		}
		if i == 0 {
			primary, primaryName = res, sum.Name
		} else if issues := showfile.CompareStructure(primary.Cues, res.Cues, primaryName); len(issues) > 0 {
			sum.Structure = issues
		}
		sum.Available = len(sum.Structure) == 0
		out = append(out, sum)
	}
	known := make([]LangInfo, 0, len(showfile.KnownLanguages))
	for _, l := range showfile.KnownLanguages {
		known = append(known, LangInfo{Code: l.Code, Name: l.Name})
	}
	writeJSON(w, http.StatusOK, map[string]any{"languages": out, "known": known})
}

// addLanguage starts a translation: an empty show file, or (to translate line by line) a copy of the main
// show file, which already has the right structure.
func (s *Server) addLanguage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code         string `json:"code"`
		Name         string `json:"name"`
		CopyFromMain bool   `json:"copyFromMain"`
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
	code := strings.ToLower(strings.TrimSpace(req.Code))
	if !showfile.LangCodeRe.MatchString(code) {
		writeErr(w, http.StatusBadRequest, "Use a short language code such as es, fr or pt-br.")
		return
	}
	if existing, _ := s.store.Languages(id); containsStr(existing, code) {
		writeErr(w, http.StatusConflict, "This show already has that language.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = showfile.LanguageName(code)
	}
	if strings.ContainsAny(name, "|\n[]") || len(name) > 40 {
		writeErr(w, http.StatusBadRequest, "The language name cannot contain | [ ] or line breaks, and must be under 40 characters.")
		return
	}
	header := "[language: " + code + " | " + name + "]\n"
	body := header
	if req.CopyFromMain {
		main, _ := s.store.ReadShowFile(id, m.Language)
		// The look and other settings belong to the main show file only, so the copy leaves them out.
		body += "<!-- Translate the words. Keep every line, command and blank screen in the same order as the main show file: one press of Space shows the same line in every language. -->\n" + showfile.ReplaceSettings(stripLanguageLine(main), showfile.Settings{})
	}
	if err := s.store.WriteShowFile(id, code, body); err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"code": code, "name": name})
}

// stripLanguageLine removes a leading [language: ...] directive so a copy does not carry the main language's.
func stripLanguageLine(text string) string {
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		if t := strings.ToLower(strings.TrimSpace(ln)); strings.HasPrefix(t, "[language") && strings.HasSuffix(t, "]") {
			return strings.Join(append(lines[:i:i], lines[i+1:]...), "\n")
		}
	}
	return text
}

func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (s *Server) deleteLanguage(w http.ResponseWriter, r *http.Request) {
	id, code := r.PathValue("id"), r.PathValue("lang")
	if s.rt.ShowID() == id && s.rt.Engine().Operator().Live {
		writeErr(w, http.StatusConflict, "This show is on air. Restart it before removing a language.")
		return
	}
	if err := s.store.DeleteShowFile(id, code); err != nil {
		storeErr(w, err)
		return
	}
	s.applyActive(id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
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
		uploadErr(w, err, "That upload was too large. Images can be up to 10 MB each.", "That upload could not be read. Try choosing the images again.")
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

// uploadErr answers a form that could not be read: 413 if the body went over its limit, otherwise 400.
func uploadErr(w http.ResponseWriter, err error, tooLarge, unreadable string) {
	var big *http.MaxBytesError
	if errors.As(err, &big) {
		writeErr(w, http.StatusRequestEntityTooLarge, tooLarge)
		return
	}
	writeErr(w, http.StatusBadRequest, unreadable)
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
		Start bool `json:"start"` // also put it on air at once (for automation; Setup loads it into standby)
	}
	json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&req)
	id := r.PathValue("id")
	l, err := s.loadShow(id)
	if err != nil {
		storeErr(w, err)
		return
	}
	if l.meta.Archived {
		writeErr(w, http.StatusConflict, "This show is archived. Restore it from the archive first.")
		return
	}
	if s.rt.Engine().Operator().Live && !req.Force {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "A show is on air right now. Activating a show restarts it from the beginning.",
			"live": true, "activeId": s.rt.ShowID()})
		return
	}
	s.rt.SetShowLangs(id, l.meta.Name, l.assets, l.primary, l.others, themeToRuntime(l.meta.Theme))
	if req.Start {
		s.rt.Do(func(e *live.Engine) (live.OperatorState, error) { return e.Start(-1) })
	}
	s.rt.Flush()
	logx.Infof("show %q activated (%d language(s))", l.meta.Name, 1+len(l.others))
	skipped := []string{}
	for code := range l.skipped {
		skipped = append(skipped, code)
	}
	sort.Strings(skipped)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "activeId": id, "skippedLanguages": skipped})
}

// ---- export / import ----

func (s *Server) exportShow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.store.Get(id); err != nil {
		storeErr(w, err)
		return
	}
	// Built in memory first, so a failure is an error the person sees rather than a truncated download.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	var werr error
	put := func(name string, data []byte) {
		if werr != nil {
			return
		}
		f, err := zw.Create(name)
		if err == nil {
			_, err = f.Write(data)
		}
		werr = err
	}
	// The show file(s) hold everything but the images: show.md (main language) and show.<lang>.md.
	codes, _ := s.store.Languages(id)
	for i, code := range codes {
		name := "show." + code + ".md"
		if i == 0 {
			name = "show.md"
		}
		text, err := s.store.ReadShowFile(id, code)
		if err != nil {
			logx.Errorf("export of show %s: could not read %s: %v", id, name, err)
			writeErr(w, http.StatusInternalServerError, "The show could not be exported. Try again; if it keeps happening, look at the log.")
			return
		}
		put(name, []byte(text))
	}
	for _, a := range mustAssets(s.store, id) {
		b, err := os.ReadFile(filepath.Join(s.store.AssetsDir(id), a.Name))
		if err != nil {
			logx.Errorf("export of show %s: could not read image %s: %v", id, a.Name, err)
			writeErr(w, http.StatusInternalServerError, "The show could not be exported. Try again; if it keeps happening, look at the log.")
			return
		}
		put("assets/"+a.Name, b)
	}
	if err := zw.Close(); werr == nil {
		werr = err
	}
	if werr != nil {
		logx.Errorf("export of show %s: %v", id, werr)
		writeErr(w, http.StatusInternalServerError, "The show could not be exported. Try again; if it keeps happening, look at the log.")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, id))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Write(buf.Bytes())
}

func (s *Server) importShow(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxZipBody)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		uploadErr(w, err, "That file is too large.", "That file could not be read. Try choosing it again.")
		return
	}
	fh, hdr, err := r.FormFile("file")
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
	if !bytes.HasPrefix(data, []byte("PK")) { // not a .zip: a plain show file, whatever it is called
		if strings.HasSuffix(strings.ToLower(hdr.Filename), ".zip") { // named like an export, but is not one
			writeErr(w, http.StatusBadRequest, "That is not a show export (.zip).")
			return
		}
		s.importPlainFile(w, hdr.Filename, data)
		return
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(zr.File) > maxZipEntries {
		writeErr(w, http.StatusBadRequest, "That is not a show export (.zip).")
		return
	}
	var declared uint64
	for _, f := range zr.File {
		declared += f.UncompressedSize64
	}
	if declared > maxZipUnpacked {
		writeErr(w, http.StatusRequestEntityTooLarge, "That file is too large.")
		return
	}
	var unpacked int64
	read := func(f *zip.File, limit int64) ([]byte, error) {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, limit+1))
		if unpacked += int64(len(b)); int64(len(b)) > limit || unpacked > maxZipUnpacked {
			return nil, show.ErrTooLarge
		}
		return b, err
	}
	files := map[string]string{}
	for _, f := range zr.File {
		name := path.Clean(f.Name)
		if strings.Contains(name, "/") || !strings.HasSuffix(name, ".md") {
			continue
		}
		if b, err := read(f, show.MaxShowFile); err == nil {
			files[name] = string(b)
		}
	}
	m, err := s.store.ImportFiles(files, " (imported)")
	if err != nil {
		if errors.Is(err, show.ErrNotFound) || errors.Is(err, show.ErrBadName) {
			writeErr(w, http.StatusBadRequest, "That zip does not contain a show file (show.md).")
			return
		}
		storeErr(w, err)
		return
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

// importPlainFile makes a new show from one plain text show file (.md or .txt, any name). The title comes
// from its [settings] block, or else from the file's name. Files saved by older Windows programs in
// Windows-1252 rather than UTF-8 are converted, so accented letters survive.
func (s *Server) importPlainFile(w http.ResponseWriter, filename string, data []byte) {
	if len(data) > show.MaxShowFile {
		writeErr(w, http.StatusRequestEntityTooLarge, "That file is too large.")
		return
	}
	if bytes.IndexByte(data, 0) >= 0 {
		writeErr(w, http.StatusBadRequest, "That is not a text file. Choose a show file (.md or .txt) or an exported show (.zip).")
		return
	}
	text := decodeText(data)
	if strings.TrimSpace(text) == "" {
		writeErr(w, http.StatusBadRequest, "That file is empty.")
		return
	}
	title := strings.TrimSpace(showfile.ReadSettings(text, false).TitleOr(""))
	if title == "" {
		base := path.Base(strings.ReplaceAll(filename, "\\", "/"))
		title = strings.TrimSpace(strings.TrimSuffix(base, path.Ext(base)))
		title = strings.TrimSpace(strings.NewReplacer("_", " ").Replace(title))
		if title == "" || title == "." {
			title = "Imported show"
		}
	}
	if r := []rune(title); len(r) > 100 {
		title = string(r[:100])
	}
	m, err := s.store.Create(title, text)
	if err != nil {
		storeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"show": s.summary(m), "skippedFiles": 0})
}

// decodeText turns uploaded bytes into UTF-8: UTF-8 (with or without a byte order mark) as it is, UTF-16
// with a mark converted, and anything else read as Windows-1252.
func decodeText(b []byte) string {
	switch {
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}), bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		be := b[0] == 0xFE
		u := make([]uint16, 0, len(b)/2)
		for i := 2; i+1 < len(b); i += 2 {
			if be {
				u = append(u, uint16(b[i])<<8|uint16(b[i+1]))
			} else {
				u = append(u, uint16(b[i+1])<<8|uint16(b[i]))
			}
		}
		return string(utf16.Decode(u))
	case utf8.Valid(b):
		return strings.TrimPrefix(string(b), "\ufeff")
	}
	var sb strings.Builder
	for _, c := range b {
		if c >= 0x80 && c <= 0x9F {
			sb.WriteRune(cp1252[c-0x80])
		} else {
			sb.WriteRune(rune(c))
		}
	}
	return sb.String()
}

// cp1252 holds what Windows-1252 puts at 0x80-0x9F; the rest of that code page matches Unicode.
var cp1252 = [32]rune{'€', 0x81, '‚', 'ƒ', '„', '…', '†', '‡', 'ˆ', '‰', 'Š', '‹', 'Œ', 0x8D, 'Ž', 0x8F, 0x90, '‘', '’', '“', '”', '•', '–', '—', '˜', '™', 'š', '›', 'œ', 0x9D, 'ž', 'Ÿ'}

// ---- network panel ----

type urlInfo struct {
	URL    string `json:"url"`
	Kind   string `json:"kind"` // "http" | "https"
	Host   string `json:"host"`
	Public bool   `json:"public"` // reachable from other devices (not localhost)
	Name   bool   `json:"name"`   // a local network name (open-caption.local) rather than a numeric address
	// Adapter is the kind of network adapter a numeric address belongs to ("ethernet", "wifi" or "other"), so Setup can say which.
	Adapter string `json:"adapter,omitempty"`
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
	adapterOf := map[string]string{}
	for _, a := range netinfo.AllAddresses() {
		adapterOf[a.IP.String()] = a.Kind
	}
	var urls []urlInfo
	add := func(scheme, bound, def string) {
		if bound == "" {
			return
		}
		for _, ip := range ips {
			urls = append(urls, urlInfo{URL: hostURL(scheme, ip.String(), bound, def), Kind: scheme, Host: ip.String(), Public: true, Adapter: adapterOf[ip.String()]})
		}
		urls = append(urls, urlInfo{URL: hostURL(scheme, "localhost", bound, def), Kind: scheme, Host: "localhost"})
	}
	add("http", l.HTTPAddr, "80")
	if st := s.MDNSStatus(); st.Active && l.HTTPAddr != "" {
		urls = append([]urlInfo{{URL: hostURL("http", st.Host, l.HTTPAddr, "80"), Kind: "http", Host: st.Host, Public: true, Name: true}}, urls...)
	}
	add("https", l.HTTPSAddr, "443")

	var warnings []string
	if len(ips) == 0 {
		if len(netinfo.AllAddresses()) > 0 {
			warnings = append(warnings, "Every network interface of this computer is turned off in the network settings. Tick one in Settings, Network.")
		} else {
			warnings = append(warnings, "No network address was found. Connect this computer to the venue's network, then reload.")
		}
	}
	if l.HTTPAddr != "" && l.HTTPConfigured != "" && portOf(l.HTTPAddr) != portOf(l.HTTPConfigured) {
		warnings = append(warnings, fmt.Sprintf("Port %s was not available, so the server is using port %s. Audience addresses must include the port.", portOf(l.HTTPConfigured), portOf(l.HTTPAddr)))
	}
	if l.HTTPSAddr != "" && s.cfg.HTTPS.CertFile == "" {
		warnings = append(warnings, "HTTPS is using a self-signed certificate. Each device shows a security warning the first time. Plain http:// avoids this.")
	}
	writeJSON(w, http.StatusOK, map[string]any{"hostname": host, "ips": ipStrings(ips), "urls": urls, "warnings": warnings,
		"audience": s.rt.AudienceCount(), "mdns": s.MDNSStatus()})
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
	if st := s.MDNSStatus(); st.Active && strings.EqualFold(h, st.Host) {
		return true
	}
	for _, a := range netinfo.AllAddresses() { // every address, also one that is not offered: it still reaches this computer
		if a.IP.String() == h {
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
	// The old rehearsal of this session is replaced under the same lock that inserts the new one: two starts
	// at once (a double click, two tabs) would otherwise both find nothing to drop and the second would
	// overwrite the first in the map, leaving its runtime running with nothing able to close it.
	s.rmu.Lock()
	old := s.rehearsals[tok]
	delete(s.rehearsals, tok)
	if len(s.rehearsals) >= maxRehearsals {
		s.rmu.Unlock()
		if old != nil {
			old.Close()
		}
		writeErr(w, http.StatusTooManyRequests, "too many rehearsals are open")
		return
	}
	cfg := s.snapshot()
	cfg.Audience.MaxClients = 8
	rt := NewRuntime(cfg)
	if s.standby {
		rt.UseStandby()
	}
	rt.SetStandbyScreen(s.standbyView())
	rt.SetAssetBase("/api/admin/shows/" + url.PathEscape(req.ShowID) + "/assets/")
	rt.SetShowLangs(req.ShowID, l.meta.Name, l.assets, l.primary, l.others, themeToRuntime(l.meta.Theme))
	rt.Claim(tok, true)
	s.rehearsals[tok] = rt
	s.rmu.Unlock()
	if old != nil {
		old.Close()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "showId": req.ShowID})
}

// ---- settings ----

func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	s.lmu.Lock()
	l := s.listeners
	s.lmu.Unlock()
	ns := s.networkSettings()
	writeJSON(w, http.StatusOK, map[string]any{
		"config": s.snapshot().Redacted(), "configPath": s.cfg.Path, "dataDir": s.cfg.DataDir,
		"httpAddr": l.HTTPAddr, "httpsAddr": l.HTTPSAddr, "version": s.version,
		"canChangePasswords": s.cfg.Path != "", "standby": s.standbyView(), "logLevel": logx.Current().String(), "logLevelPinned": s.pinnedLevel, "mdns": s.MDNSStatus(),
		"adminAllow": s.cfg.Access.AdminAllow, "operatorAllow": s.cfg.Access.OperatorAllow,
		"adminRestricted": s.adminAllow.Restricted(), "operatorRestricted": s.operatorAllow.Restricted(),
		"network": map[string]any{"interfaces": ns.Interfaces, "restrictListening": ns.RestrictListening}, "networkPinned": s.pinnedNetwork, "interfaces": s.networkInterfaces(r),
	})
}

var logLevelRe = regexp.MustCompile(`^\d{4}/\d\d/\d\d \d\d:\d\d:\d\d(?:\.\d+)? (DEBUG|INFO|WARN|ERROR) `)

// recentLog returns the last lines of the in-memory log (the same lines the support bundle holds), for the
// log viewer in Setup. ?level= is the lowest severity to show (debug, info, warn or error) and ?lines= how
// many of the newest matching lines to return (up to 1000).
func (s *Server) recentLog(w http.ResponseWriter, r *http.Request) {
	min := logx.Debug
	if q := r.URL.Query().Get("level"); q != "" {
		lv, err := logx.Parse(q)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "The level must be debug, info, warn or error.")
			return
		}
		min = lv
	}
	n := 300
	if q := r.URL.Query().Get("lines"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			n = v
		}
	}
	if n > 1000 {
		n = 1000
	}
	out := []string{}
	if s.logs != nil {
		for _, ln := range s.logs.Lines() {
			lv := logx.Info // a line without a level (from another writer) counts as normal
			if m := logLevelRe.FindStringSubmatch(ln); m != nil {
				lv, _ = logx.Parse(m[1])
			}
			if lv >= min {
				out = append(out, ln)
			}
		}
	}
	if len(out) > n {
		out = out[len(out)-n:]
	}
	writeJSON(w, http.StatusOK, map[string]any{"lines": out, "level": logx.Current().String()})
}

// setLogLevel changes how much is logged, at once, and records it in config.yaml so it is kept at the next start.
func (s *Server) setLogLevel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Level string `json:"level"`
	}
	if !decode(w, r, &req) {
		return
	}
	lv, err := logx.Parse(req.Level)
	if err != nil || strings.TrimSpace(req.Level) == "" {
		writeErr(w, http.StatusBadRequest, "The log level must be debug, info, warn or error.")
		return
	}
	saved := false
	if s.cfg.Path != "" {
		if err := config.SetValues(s.cfg.Path, map[string]any{"log.level": lv.String()}); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	logx.Infof("log level changed to %s in Setup", lv)
	logx.SetLevel(lv)
	resp := map[string]any{"level": lv.String(), "saved": saved}
	if s.pinnedLevel != "" {
		resp["note"] = "This copy was started with --log-level " + s.pinnedLevel + " (or OPEN_CAPTION_LOG_LEVEL), which overrides the saved setting each time it starts. The new level applies now, and lasts until the program is restarted."
	}
	writeJSON(w, http.StatusOK, resp)
}

// changePasswords sets new admin and/or operator passwords (blank = keep). The new values are written to
// config.yaml, take effect immediately, and sign out anyone using a password that changed (except the caller).
func (s *Server) changePasswords(w http.ResponseWriter, r *http.Request) {
	var req struct{ Admin, Operator string }
	if !decode(w, r, &req) {
		return
	}
	req.Admin, req.Operator = strings.TrimSpace(req.Admin), strings.TrimSpace(req.Operator)
	if s.cfg.Path == "" {
		writeErr(w, http.StatusConflict, "This copy was started without a config file, so passwords cannot be changed here.")
		return
	}
	if req.Admin == "" && req.Operator == "" {
		writeErr(w, http.StatusBadRequest, "Enter a new password in at least one box.")
		return
	}
	for _, pw := range []string{req.Admin, req.Operator} {
		if pw != "" && len(pw) < config.MinPasswordLength {
			writeErr(w, http.StatusBadRequest, fmt.Sprintf("Passwords must be at least %d characters.", config.MinPasswordLength))
			return
		}
	}
	s.pwMu.Lock()
	defer s.pwMu.Unlock()
	newAdmin, newOp := req.Admin, req.Operator
	effAdmin, effOp := s.cfg.AdminPassword, s.cfg.OperatorPassword
	if newAdmin != "" {
		effAdmin = newAdmin
	}
	if newOp != "" {
		effOp = newOp
	}
	if effAdmin == effOp {
		writeErr(w, http.StatusBadRequest, "The admin and operator passwords must be different.")
		return
	}
	if err := config.SetPasswords(s.cfg.Path, newAdmin, newOp); err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save config.yaml: "+err.Error())
		return
	}
	s.cfgMu.Lock()
	s.cfg.AdminPassword, s.cfg.OperatorPassword = effAdmin, effOp
	s.cfgMu.Unlock()
	s.auth.setPasswords(newAdmin, newOp, principal(r).Token)
	logx.Infof("passwords changed in Setup (admin: %v, operator: %v); other sessions were signed out", newAdmin != "", newOp != "")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "changed": map[string]bool{"admin": newAdmin != "", "operator": newOp != ""}})
}

// ---- support bundle (REQ-061) ----

// supportBundle returns a zip with settings (passwords redacted), recent log
// lines and basic facts about this computer, for sending to whoever helps.
func (s *Server) supportBundle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="open-caption-support.zip"`)
	zw := zip.NewWriter(w)
	defer zw.Close()
	put := func(name, text string) {
		if f, err := zw.Create(name); err == nil {
			io.WriteString(f, text)
		}
	}

	host, _ := os.Hostname()
	op := s.rt.Engine().Operator()
	s.lmu.Lock()
	l := s.listeners
	s.lmu.Unlock()
	var info strings.Builder
	fmt.Fprintf(&info, "open-caption %s\n", s.version)
	fmt.Fprintf(&info, "generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&info, "log level: %s\n", logx.Current())
	fmt.Fprintf(&info, "uptime: %s\n", time.Since(s.started).Round(time.Second))
	fmt.Fprintf(&info, "go: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&info, "host: %s\naddresses: %s\n", host, strings.Join(ipStrings(netinfo.LocalIPs()), ", "))
	fmt.Fprintf(&info, "http: %s (configured %s)\nhttps: %s (configured %s)\n", l.HTTPAddr, l.HTTPConfigured, l.HTTPSAddr, l.HTTPSConfigured)
	fmt.Fprintf(&info, "data dir: %s\n", s.cfg.DataDir)
	fmt.Fprintf(&info, "audience devices connected: %d\n", s.rt.AudienceCount())
	fmt.Fprintf(&info, "active show: %q live=%v cursor=%d of %d mode=%s\n", s.rt.ShowID(), op.Live, op.Cursor, op.Total, op.Mode)
	put("info.txt", info.String())

	if b, err := yaml.Marshal(s.snapshot().Redacted()); err == nil {
		put("config.redacted.yaml", string(b))
	}
	if s.logs != nil {
		put("log.txt", s.logs.String())
	}
	var shows strings.Builder
	if list, err := s.store.List(); err == nil {
		for _, m := range list {
			text, _ := s.store.ReadShowFile(m.ID, "")
			a, _ := analyze(text, mustAssets(s.store, m.ID), false)
			var errs, warns int
			for _, d := range a.Diagnostics {
				switch d.Severity {
				case "error":
					errs++
				case "warning":
					warns++
				}
			}
			fmt.Fprintf(&shows, "%s  %q  updated %s  cues=%d errors=%d warnings=%d images=%d missing=%d\n",
				m.ID, m.Name, m.Updated.Format(time.RFC3339), len(a.Cues), errs, warns, len(a.Assets), len(a.Missing))
		}
	}
	put("shows.txt", shows.String())
}
