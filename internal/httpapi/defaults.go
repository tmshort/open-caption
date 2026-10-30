// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"
	"strings"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logx"
)

// savedDefaults are the venue-wide defaults and audience limits as saved in config.yaml from Setup.
type savedDefaults struct {
	Defaults config.Defaults
	Audience config.Audience
}

// defaultsView is the defaults as Setup shows them: what is saved (which the running program uses only after its next
// start), and whether that differs from what is running now.
func (s *Server) defaultsView() map[string]any {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	cur := savedDefaults{s.cfg.Defaults, s.cfg.Audience}
	saved := cur
	if s.savedDefaults != nil {
		saved = *s.savedDefaults
	}
	d, l := saved.Defaults, saved.Defaults.Look
	return map[string]any{
		"language": d.Language, "idleMessage": d.IdleMessage, "loopSeconds": d.LoopItemSeconds, "openBrowser": d.OpenBrowser,
		"background": l.Background, "textColor": l.TextColor, "speakerNameColor": l.SpeakerNameColor, "font": l.Font,
		"textSizePercent": l.TextSizePercent, "speakerNames": l.SpeakerNames, "speakerPosition": l.SpeakerPosition,
		"scrollbackLines": saved.Audience.ScrollbackLines, "maxClients": saved.Audience.MaxClients,
		"restartNeeded": saved != cur,
	}
}

// setDefaults is PUT /api/admin/settings/defaults: the values a show uses where its show file says nothing, and the audience
// limits. They are checked, written to config.yaml (comments kept) and used from the next start; nothing running changes.
func (s *Server) setDefaults(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Language         string  `json:"language"`
		IdleMessage      string  `json:"idleMessage"`
		LoopSeconds      float64 `json:"loopSeconds"`
		OpenBrowser      bool    `json:"openBrowser"`
		Background       string  `json:"background"`
		TextColor        string  `json:"textColor"`
		SpeakerNameColor string  `json:"speakerNameColor"`
		Font             string  `json:"font"`
		TextSizePercent  int     `json:"textSizePercent"`
		SpeakerNames     string  `json:"speakerNames"`
		SpeakerPosition  string  `json:"speakerPosition"`
		ScrollbackLines  int     `json:"scrollbackLines"`
		MaxClients       int     `json:"maxClients"`
	}
	if !decode(w, r, &req) {
		return
	}
	if s.cfg.Path == "" {
		writeErr(w, http.StatusConflict, "This copy was started without a config file, so the defaults cannot be saved here.")
		return
	}
	d := config.Defaults{Language: strings.ToLower(strings.TrimSpace(req.Language)), LoopItemSeconds: req.LoopSeconds, IdleMessage: strings.TrimSpace(req.IdleMessage), OpenBrowser: req.OpenBrowser,
		Look: config.Look{Background: req.Background, TextColor: req.TextColor, SpeakerNameColor: req.SpeakerNameColor, Font: req.Font,
			TextSizePercent: req.TextSizePercent, SpeakerNames: req.SpeakerNames, SpeakerPosition: req.SpeakerPosition}}
	d, a, probs := config.CheckDefaults(d, config.Audience{MaxClients: req.MaxClients, ScrollbackLines: req.ScrollbackLines})
	if len(probs) > 0 {
		writeErr(w, http.StatusBadRequest, strings.Join(probs, "; "))
		return
	}
	kv := map[string]any{
		"defaults.language": d.Language, "defaults.loop_item_seconds": d.LoopItemSeconds, "defaults.idle_message": d.IdleMessage, "defaults.open_browser": d.OpenBrowser,
		"defaults.look.background": d.Look.Background, "defaults.look.text_color": d.Look.TextColor, "defaults.look.speaker_name_color": d.Look.SpeakerNameColor,
		"defaults.look.font": d.Look.Font, "defaults.look.text_size_percent": d.Look.TextSizePercent, "defaults.look.speaker_names": d.Look.SpeakerNames,
		"defaults.look.speaker_position": d.Look.SpeakerPosition, "audience.max_clients": a.MaxClients, "audience.scrollback_lines": a.ScrollbackLines,
	}
	if err := config.SetValues(s.cfg.Path, kv); err != nil {
		writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
		return
	}
	s.cfgMu.Lock()
	s.savedDefaults = &savedDefaults{d, a}
	s.cfgMu.Unlock()
	logx.Infof("the venue defaults and audience limits were saved in Setup; they apply from the next start")
	writeJSON(w, http.StatusOK, map[string]any{"saved": true, "defaults": s.defaultsView()})
}
