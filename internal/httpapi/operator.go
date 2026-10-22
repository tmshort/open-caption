// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmshort/open-caption/internal/logx"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/showfile"
	"github.com/tmshort/open-caption/web"
)

const (
	// frame-ancestors 'self' lets Setup embed the operator and audience pages for rehearsal.
	pageCSP    = "default-src 'self'; img-src 'self' data:; style-src 'self'; frame-ancestors 'self'"
	maxCmdBody = 8 << 10
)

// ItemView is a fullscreen item as the operator sees it: names only, never image URLs (REQ-031).
type ItemView struct {
	Kind    string  `json:"kind"`
	Label   string  `json:"label"`
	Seconds float64 `json:"seconds,omitempty"`
}

// PartView is one voice of a caption shown together with others.
type PartView struct {
	Speaker string `json:"speaker,omitempty"`
	Text    string `json:"text"`
	Sung    bool   `json:"sung,omitempty"`
}

// CueView is one row of the operator's show file list.
type CueView struct {
	Index   int        `json:"index"`
	Kind    string     `json:"kind"`
	Line    int        `json:"line"`
	Speaker string     `json:"speaker,omitempty"`
	Text    string     `json:"text,omitempty"`
	Stage   bool       `json:"stage,omitempty"`
	Sung    bool       `json:"sung,omitempty"`
	Parts   []PartView `json:"parts,omitempty"`
	Label   string     `json:"label,omitempty"`
	Period  float64    `json:"period,omitempty"`
	Items   []ItemView `json:"items,omitempty"`
}

func itemView(it showfile.Item) ItemView {
	v := ItemView{Kind: it.Kind, Seconds: it.Seconds}
	switch it.Kind {
	case "image":
		v.Label = it.Image
		if it.Caption != "" {
			v.Label += " — " + it.Caption
		}
	case "text":
		v.Label = it.Text
	default:
		v.Label = "blank"
	}
	return v
}

func cueViews(cues []showfile.Cue) []CueView {
	out := make([]CueView, len(cues))
	for i, c := range cues {
		v := CueView{Index: i, Kind: string(c.Kind), Line: c.Line, Speaker: c.Speaker, Text: c.Text, Stage: c.Stage, Sung: c.Sung, Period: c.Period}
		if c.Item != nil {
			v.Label = itemView(*c.Item).Label
		}
		for _, it := range c.Items {
			v.Items = append(v.Items, itemView(it))
		}
		for _, pt := range c.Parts {
			v.Parts = append(v.Parts, PartView{Speaker: pt.Speaker, Text: pt.Text, Sung: pt.Sung})
		}
		out[i] = v
	}
	return out
}

// resolver finds the runtime a request should act on (nil: none available).
type resolver func(r *http.Request) *Runtime

func (s *Server) operatorRoutes(mux *http.ServeMux) {
	sub, _ := fs.Sub(web.Operator, "operator")
	static := http.StripPrefix("/static/operator/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /operator", s.gate(RoleOperator, func(w http.ResponseWriter, r *http.Request) { servePage(w, sub, "index.html") }))
	mux.HandleFunc("GET /static/operator/", s.gate(RoleOperator, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		static.ServeHTTP(w, r)
	}))

	s.operatorAPI(mux, "/api/operator/", RoleOperator, func(*http.Request) *Runtime { return s.rt })

	// The audience report (REQ-070): a page, and its data as JSON and CSV. Only for the show on air, not rehearsals.
	mux.HandleFunc("GET /operator/report", s.gate(RoleOperator, func(w http.ResponseWriter, r *http.Request) { servePage(w, sub, "report.html") }))
	mux.HandleFunc("GET /api/operator/report", s.require(RoleOperator, s.reportJSON))
	mux.HandleFunc("GET /api/operator/reports", s.require(RoleOperator, s.reportList))
	mux.HandleFunc("GET /api/operator/report.csv", s.require(RoleOperator, s.reportCSV))
}

// operatorAPI registers the live-control API under prefix, acting on whichever
// runtime resolve returns. The real show uses it at /api/operator/, and each
// admin's rehearsal at /api/admin/rehearsal/, so both behave identically.
func (s *Server) operatorAPI(mux *http.ServeMux, prefix, role string, resolve resolver) {
	guard := func(h func(http.ResponseWriter, *http.Request, *Runtime)) http.HandlerFunc {
		return s.require(role, func(w http.ResponseWriter, r *http.Request) {
			rt := resolve(r)
			if rt == nil {
				writeErr(w, http.StatusNotFound, "no rehearsal is running; start one from Setup")
				return
			}
			h(w, r, rt)
		})
	}
	mux.HandleFunc("GET "+prefix+"state", guard(s.operatorState))
	mux.HandleFunc("GET "+prefix+"stream", guard(s.operatorStream))
	mux.HandleFunc("GET "+prefix+"clients", guard(s.listClients))
	mux.HandleFunc("POST "+prefix+"claim", guard(func(w http.ResponseWriter, r *http.Request, rt *Runtime) {
		var req struct{ Takeover bool }
		json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCmdBody)).Decode(&req)
		ok := rt.Claim(principal(r).Token, req.Takeover)
		if ok {
			logx.Infof("operator control taken (takeover: %v) by session %s", req.Takeover, SID(principal(r).Token))
		} else {
			logx.Debugf("operator control refused: another operator is in control")
		}
		code := http.StatusOK
		if !ok {
			code = http.StatusConflict
		}
		writeJSON(w, code, map[string]any{"ok": ok, "payload": rt.OperatorPayload()})
	}))

	command := func(name string, fn func(*live.Engine, cmdBody) (live.OperatorState, error)) {
		mux.HandleFunc("POST "+prefix+name, guard(func(w http.ResponseWriter, r *http.Request, rt *Runtime) {
			runCommand(w, r, rt, fn)
		}))
	}
	command("next", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Next(b.rev()) })
	command("skip", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Skip(b.rev()) })
	command("back", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Back(b.rev()) })
	command("stop", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Stop(b.rev()) })
	command("jump", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Jump(b.Index, b.rev()) })
	command("blank", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.SetBlank(b.On, b.rev()) })
	command("clear", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Clear(b.rev()) })
	command("reset", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Reset(b.rev()) })
	command("onair", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.Start(b.rev()) })
	command("offair", func(e *live.Engine, b cmdBody) (live.OperatorState, error) { return e.OffAir(b.rev()) })
	command("message", func(e *live.Engine, b cmdBody) (live.OperatorState, error) {
		if len(b.Text) == 0 || len(b.Text) > 500 {
			return e.Operator(), errBadRequest
		}
		return e.Message(b.Text, b.rev())
	})
}

var errBadRequest = errors.New("bad request")

type cmdBody struct {
	Rev   *int64 `json:"rev"`
	Index int    `json:"index"`
	On    bool   `json:"on"`
	Text  string `json:"text"`
}

// rev returns the client's revision, or -1 (skip the check) if it sent none.
func (b cmdBody) rev() int64 {
	if b.Rev == nil {
		return -1
	}
	return *b.Rev
}

func servePage(w http.ResponseWriter, sub fs.FS, name string) {
	b, err := fs.ReadFile(sub, name)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Content-Security-Policy", pageCSP)
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-store")
	w.Write(b)
}

func (s *Server) operatorState(w http.ResponseWriter, r *http.Request, rt *Runtime) {
	p := rt.OperatorPayload()
	writeJSON(w, http.StatusOK, map[string]any{
		"payload": p, "cues": cueViews(rt.Engine().Cues()), "me": SID(principal(r).Token),
		"role": principal(r).Role, "auto": principal(r).Auto,
	})
}

func runCommand(w http.ResponseWriter, r *http.Request, rt *Runtime, fn func(*live.Engine, cmdBody) (live.OperatorState, error)) {
	p := principal(r)
	if !rt.IsController(p.Token) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "another operator is in control", "payload": rt.OperatorPayload()})
		return
	}
	var body cmdBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCmdBody)).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeErr(w, http.StatusBadRequest, "bad request")
		return
	}
	_, err := rt.Do(func(e *live.Engine) (live.OperatorState, error) { return fn(e, body) })
	logx.Debugf("operator command %s: %v", r.URL.Path, err)
	resp := map[string]any{"payload": rt.OperatorPayload()}
	code := http.StatusOK
	switch {
	case err == nil:
	case errors.Is(err, live.ErrStale):
		code, resp["error"] = http.StatusConflict, "the show moved on; your view was out of date"
	case errors.Is(err, live.ErrEnd):
		resp["notice"] = "End of the show"
	case errors.Is(err, live.ErrNoUndo):
		resp["notice"] = "Nothing to undo"
	case errors.Is(err, live.ErrNoLoop):
		resp["notice"] = "No loop is running"
	case errors.Is(err, live.ErrStandby):
		code, resp["error"] = http.StatusConflict, "the show is not on air; start it first"
	case errors.Is(err, live.ErrBadCue), errors.Is(err, errBadRequest):
		code, resp["error"] = http.StatusBadRequest, err.Error()
	default:
		code, resp["error"] = http.StatusInternalServerError, err.Error()
	}
	writeJSON(w, code, resp)
}

// operatorStream pushes OperatorPayload on every change (SSE).
func (s *Server) operatorStream(w http.ResponseWriter, r *http.Request, rt *Runtime) {
	rc := http.NewResponseController(w)
	tok := principal(r).Token
	cl := rt.OpHub.Subscribe()
	rt.streamOpen(tok)
	defer func() { rt.OpHub.Unsubscribe(cl); rt.streamClose(tok) }()

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no")
	rc.SetWriteDeadline(time.Time{})

	b, _ := json.Marshal(rt.OperatorPayload())
	fmt.Fprintf(w, "retry: 1000\n\nevent: opstate\ndata: %s\n\n", b)
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
				return
			}
			if _, err := w.Write(f.Data); err != nil || rc.Flush() != nil {
				return
			}
		case <-tick.C:
			if _, err := w.Write([]byte("event: ping\ndata: {}\n\n")); err != nil || rc.Flush() != nil {
				return
			}
		}
	}
}
