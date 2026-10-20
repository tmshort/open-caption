// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/tmshort/open-caption/internal/logx"
	"github.com/tmshort/open-caption/internal/report"
	"github.com/tmshort/open-caption/internal/showfile"
)

// The audience report (REQ-070): how many devices used the audience page during a run of the show, however long
// each stayed. The counting is internal/report; this connects it to the audience connections, to the show going on
// and off the air, to the folder that keeps the finished reports, and to the operator pages that show them.

// newReporting makes the tracker and, when there is a data folder, saves each finished run in <data>/reports.
func (s *Server) newReporting() {
	s.tracker = report.New(nil)
	if s.cfg.DataDir == "" {
		return
	}
	s.reports = report.Store{Dir: filepath.Join(s.cfg.DataDir, "reports")}
	s.tracker.OnClose = func(r report.Report) {
		if err := s.reports.Save(r); err != nil {
			logx.Warnf("could not save the audience report: %v", err)
			return
		}
		logx.Infof("audience report for %q saved: %d devices", r.Show, r.Totals.Devices)
	}
}

// syncReport tells the tracker whether the show is on the air: it starts a run when it goes on and ends it when it
// goes off the air, is restarted or is replaced. It is cheap, and called whenever a device connects and every
// couple of seconds.
func (s *Server) syncReport() {
	p := s.rt.OperatorPayload()
	var loop *report.LoopInfo
	if p.Op.Live && p.Op.Loop != nil {
		cues := s.rt.Engine().Cues()
		loop = &report.LoopInfo{Cue: p.Op.Loop.Cue, Kind: loopKind(cues, p.Op.Loop.Cue)}
		if i := p.Op.Loop.Cue; i >= 0 && i < len(cues) {
			loop.Name = cues[i].Name // the main language's name for it, when the show file gives one
		}
	}
	s.tracker.Sync(p.Op.Live && p.Show.ID != "", p.Show.ID, p.Show.Name, loop)
}

// loopKind says what a loop is by where it stands in the show file: before the first line the audience sees it is the
// pre-show, after the last it is the post-show, and between lines an intermission.
func loopKind(cues []showfile.Cue, at int) string {
	first, last := -1, -1
	for i, c := range cues {
		switch c.Kind {
		case showfile.KindCaption, showfile.KindImage, showfile.KindFullscreen:
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	switch {
	case first < 0:
		return report.LoopOther
	case at < first:
		return report.LoopPreShow
	case at > last:
		return report.LoopPostShow
	}
	return report.LoopIntermission
}

// reportLoop keeps the tracker in step with the show and saves the run in progress now and then, so a crash during
// the show does not lose the count.
func (s *Server) reportLoop() {
	defer s.done.Done()
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	saved := time.Now()
	for {
		select {
		case <-s.quit:
			s.tracker.Sync(false, "", "", nil) // closes a run in progress, which saves it
			return
		case <-t.C:
			s.syncReport()
			if s.reports.Dir != "" && time.Since(saved) >= 30*time.Second {
				saved = time.Now()
				if cur, ok := s.tracker.Current(); ok {
					if err := s.reports.Save(cur); err != nil {
						logx.Debugf("could not save the audience report so far: %v", err)
					}
				}
			}
		}
	}
}

type reportListItem struct {
	ID      string    `json:"id"`
	Show    string    `json:"show"`
	Started time.Time `json:"started"`
	Ended   time.Time `json:"ended"`
	Running bool      `json:"running"`
}

// chooseReport finds the report asked for: the one in progress or, failing that, the newest saved one ("" or
// "current"), or a saved one by its id.
func (s *Server) chooseReport(id string) (report.Report, string, bool) {
	if id == "" || id == "current" {
		if cur, ok := s.tracker.Current(); ok {
			return cur, "current", true
		}
		if r, name, ok := s.reports.Latest(); ok {
			return r, name, true
		}
		return report.Report{}, "", false
	}
	r, err := s.reports.Load(id)
	if err != nil {
		return report.Report{}, "", false
	}
	return r, id, true
}

// GET /api/operator/report?id=: a report (the one in progress, else the newest) as JSON.
func (s *Server) reportJSON(w http.ResponseWriter, r *http.Request) {
	s.syncReport()
	rep, id, ok := s.chooseReport(r.URL.Query().Get("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "There is no audience report yet. One is made each time the show is put on air.")
		return
	}
	names := map[string]string{}
	for c := range rep.Languages {
		names[c] = showfile.LanguageName(c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "report": rep, "languageNames": names})
}

// GET /api/operator/reports: the saved reports, newest first (the run in progress is listed first, as "current").
func (s *Server) reportList(w http.ResponseWriter, r *http.Request) {
	s.syncReport()
	out := []reportListItem{}
	cur, running := s.tracker.Current()
	if running {
		out = append(out, reportListItem{ID: "current", Show: cur.Show, Started: cur.Started, Running: true})
	}
	for _, n := range s.reports.Names() {
		rep, err := s.reports.Load(n)
		if err != nil || (running && report.Name(rep) == report.Name(cur)) {
			continue // unreadable, or the run in progress (saved as it goes)
		}
		out = append(out, reportListItem{ID: n, Show: rep.Show, Started: rep.Started, Ended: rep.Ended, Running: false})
	}
	writeJSON(w, http.StatusOK, map[string]any{"reports": out})
}

// GET /api/operator/report.csv?id=: one row for each device, for a spreadsheet.
func (s *Server) reportCSV(w http.ResponseWriter, r *http.Request) {
	s.syncReport()
	rep, _, ok := s.chooseReport(r.URL.Query().Get("id"))
	if !ok {
		writeErr(w, http.StatusNotFound, "There is no audience report yet. One is made each time the show is put on air.")
		return
	}
	name := "audience-report-" + strings.TrimSuffix(report.Name(rep), ".json") + ".csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte("\xef\xbb\xbf")) // so a spreadsheet reads the accents in a show's name
	report.WriteCSV(w, rep)
}
