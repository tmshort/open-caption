package httpapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logx"
)

// The standby screen is the venue's, not a show's: a logo (a file in the data folder) and a short message.
// It is what the audience sees while no show is on air, so it has to work when no show exists.

const maxStandbyLogo = 2 << 20

// defaultLogo is shown until the venue uploads its own: the same [oc] mark as the favicon (open captioning).
const defaultLogo = "/static/shared/favicon.svg"

var logoExts = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}

// logoPath is the standby logo file, or "" when there is none.
func (s *Server) logoPath() string {
	if s.cfg.DataDir == "" {
		return ""
	}
	for _, ext := range logoExts {
		p := filepath.Join(s.cfg.DataDir, "standby-logo"+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// standbyView is the standby screen as the audience page gets it.
func (s *Server) standbyView() StandbyView {
	v := StandbyView{Message: s.cfg.Standby.Message, Logo: defaultLogo, Default: true}
	if p := s.logoPath(); p != "" {
		if st, err := os.Stat(p); err == nil {
			v.Logo, v.Default = fmt.Sprintf("/standby-logo?v=%d", st.ModTime().UnixNano()), false
		}
	}
	return v
}

// applyStandby pushes the standby screen to the show on air and to every open rehearsal.
func (s *Server) applyStandby() {
	v := s.standbyView()
	s.rt.SetStandbyScreen(v)
	s.rmu.Lock()
	for _, rt := range s.rehearsals {
		rt.SetStandbyScreen(v)
	}
	s.rmu.Unlock()
}

// standbyLogo is GET /standby-logo (public: the audience page shows it).
func (s *Server) standbyLogo(w http.ResponseWriter, r *http.Request) {
	p := s.logoPath()
	if p == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600") // the address carries ?v=<modified time>
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, p)
}

// setStandby is PUT /api/admin/settings/standby: the standby message ("" for the built-in words).
func (s *Server) setStandby(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}
	if !decode(w, r, &req) {
		return
	}
	msg := strings.TrimSpace(req.Message)
	if len([]rune(msg)) > config.MaxStandbyMessage || strings.ContainsAny(msg, "\r\n") {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("use one line of at most %d characters", config.MaxStandbyMessage))
		return
	}
	saved := false
	if s.cfg.Path != "" {
		if err := config.SetValues(s.cfg.Path, map[string]any{"standby.message": msg}); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	s.cfg.Standby.Message = msg
	s.applyStandby()
	logx.Infof("standby message changed in Setup")
	writeJSON(w, http.StatusOK, map[string]any{"standby": s.standbyView(), "saved": saved})
}

// setStandbyLogo is POST /api/admin/settings/standby-logo: a PNG, JPEG, GIF or WebP up to 2 MB.
func (s *Server) setStandbyLogo(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DataDir == "" {
		writeErr(w, http.StatusConflict, "This copy has no data folder, so a logo cannot be kept.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxStandbyLogo+(64<<10))
	if err := r.ParseMultipartForm(maxStandbyLogo); err != nil {
		uploadErr(w, err, "The logo can be up to 2 MB.", "That file could not be read. Try choosing it again.")
		return
	}
	fh, _, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "Choose an image file.")
		return
	}
	defer fh.Close()
	data, err := io.ReadAll(io.LimitReader(fh, maxStandbyLogo+1))
	if err != nil || len(data) > maxStandbyLogo {
		writeErr(w, http.StatusRequestEntityTooLarge, "The logo can be up to 2 MB.")
		return
	}
	ext, ok := logoExts[strings.SplitN(http.DetectContentType(data), ";", 2)[0]]
	if !ok || bytes.HasPrefix(bytes.TrimSpace(data), []byte("<")) {
		writeErr(w, http.StatusBadRequest, "The logo must be a PNG, JPEG, GIF or WebP image.")
		return
	}
	s.removeLogo()
	if err := os.WriteFile(filepath.Join(s.cfg.DataDir, "standby-logo"+ext), data, 0o644); err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save the logo: "+err.Error())
		return
	}
	s.applyStandby()
	logx.Infof("standby logo changed in Setup")
	writeJSON(w, http.StatusOK, map[string]any{"standby": s.standbyView()})
}

func (s *Server) removeLogo() {
	for _, ext := range logoExts {
		os.Remove(filepath.Join(s.cfg.DataDir, "standby-logo"+ext))
	}
}

// deleteStandbyLogo is DELETE /api/admin/settings/standby-logo.
func (s *Server) deleteStandbyLogo(w http.ResponseWriter, r *http.Request) {
	if s.cfg.DataDir != "" {
		s.removeLogo()
	}
	s.applyStandby()
	writeJSON(w, http.StatusOK, map[string]any{"standby": s.standbyView()})
}
