// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/showfile"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func standbyServer(t *testing.T) (*Server, *httptest.Server, *client) {
	t.Helper()
	cfg := testConfig()
	cfg.DataDir = t.TempDir()
	cfg.Path = filepath.Join(cfg.DataDir, "config.yaml")
	os.WriteFile(cfg.Path, []byte("admin_password: adminpw\noperator_password: operatorpw\n"), 0o600)
	s := New(cfg, "test")
	s.EnableStandby()
	t.Cleanup(s.Close)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	c := newClient(t, ts.URL)
	if c.login("admin", "adminpw") != 200 {
		t.Fatal("admin login")
	}
	return s, ts, c
}

// audienceState is the snapshot a device would get when it connects.
func audienceState(t *testing.T, s *Server) map[string]any {
	t.Helper()
	b, _ := json.Marshal(s.rt.State("en"))
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

// A loaded show waits in standby: the audience sees the venue's screen until the operator starts it, and Live
// means on air.
func TestShowWaitsInStandbyUntilTheOperatorStartsIt(t *testing.T) {
	s, _, c := standbyServer(t)
	if m := audienceState(t, s); m["mode"] != "standby" {
		t.Fatalf("nothing loaded: %v", m["mode"])
	}
	_, r := c.do("POST", "/api/admin/shows", map[string]any{"name": "Our Town"})
	id := r["id"].(string)
	c.do("PUT", "/api/admin/shows/"+id+"/showfile", map[string]any{"text": "[loop: 5s]\nWelcome\n[/loop]\nHello\nWorld\n"})
	if code, _ := c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{}); code != 200 {
		t.Fatal(code)
	}
	c.do("POST", "/api/operator/claim", map[string]any{"takeover": true})
	m := audienceState(t, s)
	if m["mode"] != "standby" || m["fullscreen"] != nil || len(m["scrollback"].([]any)) != 0 {
		t.Fatalf("a loaded show must not be seen: %v", m)
	}
	// not live: it can be deleted without any fuss
	_, list := c.do("GET", "/api/admin/shows", nil)
	for _, x := range list["shows"].([]any) {
		if x.(map[string]any)["live"] == true {
			t.Errorf("standby is not live: %v", x)
		}
		if x.(map[string]any)["fresh"] != true {
			t.Errorf("a show in standby with nothing sent has nothing to restart: %v", x)
		}
	}
	// the operator cannot send lines yet
	if code, _ := c.do("POST", "/api/operator/next", nil); code != 409 {
		t.Errorf("next in standby: %d", code)
	}
	code, p := c.do("POST", "/api/operator/onair", nil)
	if code != 200 {
		t.Fatalf("on air: %d %v", code, p)
	}
	if m := audienceState(t, s); m["mode"] != "captions" || m["fullscreen"] == nil {
		t.Errorf("the pre-show should be on: %v", m)
	}
	_, list = c.do("GET", "/api/admin/shows", nil)
	if sh := list["shows"].([]any)[0].(map[string]any); sh["live"] != true || sh["fresh"] == true {
		t.Errorf("on air is live and can be restarted: %v", sh)
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+id, nil); code != 409 {
		t.Errorf("a show on air cannot be deleted: %d", code)
	}
	c.do("POST", "/api/operator/next", nil)
	if m := audienceState(t, s); len(m["scrollback"].([]any)) != 1 {
		t.Errorf("a line was sent: %v", m)
	}
	// off air: the audience sees standby again, the position is kept, and the show can be deleted
	if code, _ := c.do("POST", "/api/operator/offair", nil); code != 200 {
		t.Fatal(code)
	}
	if m := audienceState(t, s); m["mode"] != "standby" || len(m["scrollback"].([]any)) != 0 {
		t.Errorf("off air: %v", m)
	}
	if st := s.rt.Engine().Operator(); st.Cursor != 2 {
		t.Errorf("position lost: %+v", st)
	}
	c.do("POST", "/api/operator/onair", nil)
	if m := audienceState(t, s); len(m["scrollback"].([]any)) != 1 {
		t.Errorf("back on air as it was: %v", m)
	}
	// "start: true" goes straight on air (automation); and restarting returns to standby
	c.do("POST", "/api/operator/offair", nil)
	c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{"start": true})
	if m := audienceState(t, s); m["mode"] == "standby" {
		t.Errorf("start:true: %v", m["mode"])
	}
	c.do("POST", "/api/operator/reset", nil)
	if m := audienceState(t, s); m["mode"] != "standby" {
		t.Errorf("restart goes back to standby: %v", m["mode"])
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+id, nil); code != 200 {
		t.Errorf("a show in standby can be deleted: %d", code)
	}
	if m := audienceState(t, s); m["mode"] != "standby" {
		t.Errorf("after deleting: %v", m["mode"])
	}
}

func TestStandbyScreenIsTheVenuesOwn(t *testing.T) {
	s, ts, c := standbyServer(t)
	if st, _ := audienceState(t, s)["standby"].(map[string]any); st == nil {
		t.Fatal("the audience gets the standby screen")
	}
	// the message
	if code, r := c.do("PUT", "/api/admin/settings/standby", map[string]any{"message": "  Welcome to the Little Theatre  "}); code != 200 || r["saved"] != true {
		t.Fatalf("message: %d %v", code, r)
	}
	if got := audienceState(t, s)["standby"].(map[string]any)["message"]; got != "Welcome to the Little Theatre" {
		t.Errorf("message: %v", got)
	}
	if b, _ := os.ReadFile(s.cfg.Path); !strings.Contains(string(b), "Welcome to the Little Theatre") {
		t.Errorf("not saved in config.yaml: %s", b)
	}
	if code, _ := c.do("PUT", "/api/admin/settings/standby", map[string]any{"message": strings.Repeat("x", 300)}); code != 400 {
		t.Errorf("too long: %d", code)
	}
	if code, _ := c.do("PUT", "/api/admin/settings/standby", map[string]any{"message": "two\nlines"}); code != 400 {
		t.Errorf("two lines: %d", code)
	}
	// the logo
	upload := func(name string, data []byte) int {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		fw, _ := mw.CreateFormFile("file", name)
		fw.Write(data)
		mw.Close()
		req, _ := http.NewRequest("POST", ts.URL+"/api/admin/settings/standby-logo", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		resp, err := c.hc.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if code := upload("evil.svg", []byte("<svg xmlns='http://www.w3.org/2000/svg'><script>alert(1)</script></svg>")); code != 400 {
		t.Errorf("svg: %d", code)
	}
	if code := upload("logo.png", pngData(t)); code != 200 {
		t.Fatalf("png: %d", code)
	}
	logo, _ := audienceState(t, s)["standby"].(map[string]any)["logo"].(string)
	if !strings.HasPrefix(logo, "/standby-logo?v=") {
		t.Fatalf("logo url: %q", logo)
	}
	resp, err := http.Get(ts.URL + logo) // public: no login
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/png") || !bytes.Equal(b, pngData(t)) {
		t.Errorf("served logo: %d %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	if code, _ := c.do("DELETE", "/api/admin/settings/standby-logo", nil); code != 200 {
		t.Error("delete")
	}
	if st := audienceState(t, s)["standby"].(map[string]any); st["logo"] != defaultLogo || st["default"] != true {
		t.Errorf("the built-in logo comes back: %v", st)
	}
	if resp, _ := http.Get(ts.URL + "/standby-logo"); resp.StatusCode != 404 {
		t.Errorf("no logo: %d", resp.StatusCode)
	}
}

// A translation added to a show that is on air catches up with the primary language: on air and at the same
// cue, not stuck on the standby screen (a new engine starts in standby, where a jump is refused).
func TestLanguageAddedToAShowOnAirCatchesUp(t *testing.T) {
	s := New(testConfig(), "t")
	s.EnableStandby()
	defer s.Close()
	en, es := showfile.Parse("A\nB\nC", showfile.Options{}), showfile.Parse("a\nb\nc", showfile.Options{})
	rt := s.Runtime()
	rt.SetShowLangs("demo", "Demo", "", LangFile{"en", "English", en.Cues}, nil, DefaultTheme(""))
	rt.Do(func(e *live.Engine) (live.OperatorState, error) { return e.Start(-1) })
	for i := 0; i < 2; i++ {
		rt.Do(func(e *live.Engine) (live.OperatorState, error) { return e.Next(-1) })
	}
	rt.ReplaceLangs(LangFile{"en", "English", en.Cues}, []LangFile{{"es", "Español", es.Cues}})
	st := rt.State("es")
	var got []string
	for _, c := range st.Scrollback {
		got = append(got, c.Text)
	}
	if st.Mode == "standby" || len(got) != 2 || got[1] != "b" {
		t.Fatalf("the new language shows mode %q and %v; want on air with a, b", st.Mode, got)
	}
	// and it follows the operator from here on
	rt.Do(func(e *live.Engine) (live.OperatorState, error) { return e.Next(-1) })
	if got := rt.State("es").Scrollback; len(got) != 3 || got[2].Text != "c" {
		t.Fatalf("the new language does not follow the operator: %+v", got)
	}
}
