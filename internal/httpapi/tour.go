// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logx"
)

// tourDone says whether the first-run tour of Setup has already been finished or skipped (setup.tour_done).
func (s *Server) tourDone() bool {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.cfg.Setup.TourDone
}

// setTour is PUT /api/admin/settings/tour: {done} records that the tour of Setup has been seen (true) or is to be offered
// again (false), in config.yaml (setup.tour_done). A copy with no settings file keeps it only until it stops.
func (s *Server) setTour(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Done bool `json:"done"`
	}
	if !decode(w, r, &req) {
		return
	}
	saved := false
	if s.cfg.Path != "" {
		if err := config.SetValues(s.cfg.Path, map[string]any{"setup.tour_done": req.Done}); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	s.cfgMu.Lock()
	s.cfg.Setup.TourDone = req.Done
	s.cfgMu.Unlock()
	logx.Debugf("setup.tour_done set to %v", req.Done)
	writeJSON(w, http.StatusOK, map[string]any{"tourDone": req.Done, "saved": saved})
}
