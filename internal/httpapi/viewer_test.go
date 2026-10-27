// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/showfile"
)

func viewerServer(t *testing.T, mutate func(*config.Config)) (*Server, *httptest.Server) {
	t.Helper()
	cfg := testConfig()
	cfg.ViewerPassword = "viewerpw"
	if mutate != nil {
		mutate(&cfg)
	}
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	s.Runtime().SetShow("demo", "Demo Show", t.TempDir(), showfile.Parse("[note: hold]\nA\nB\n", showfile.Options{}).Cues, DefaultTheme("Soon"))
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}

// With viewer_password set, the operator sign-in form accepts it and opens the page for watching only: the state and its
// stream can be read, and nothing that sends, changes, takes control or belongs to Setup.
func TestViewerCanWatchAndNothingElse(t *testing.T) {
	s, ts := viewerServer(t, nil)
	v := newClient(t, ts.URL)
	if code, r := v.do("POST", "/api/login", map[string]string{"role": "viewer", "password": "viewerpw"}); code != 200 || r["role"] != RoleViewer {
		t.Fatalf("viewer sign-in: %d %v", code, r)
	}
	code, st := v.do("GET", "/api/operator/state", nil)
	if code != 200 || st["role"] != RoleViewer || st["cues"] == nil {
		t.Fatalf("a viewer reads the state, notes included: %d %v", code, st)
	}
	if code, r := v.do("GET", "/api/me", nil); code != 200 || r["role"] != RoleViewer {
		t.Errorf("me: %d %v", code, r)
	}
	if code, _ := v.do("GET", "/api/help/topics", nil); code != 200 {
		t.Errorf("a viewer may read the help: %d", code)
	}
	for _, c := range [][2]string{
		{"POST", "/api/operator/claim"}, {"POST", "/api/operator/next"}, {"POST", "/api/operator/back"}, {"POST", "/api/operator/skip"},
		{"POST", "/api/operator/stop"}, {"POST", "/api/operator/jump"}, {"POST", "/api/operator/blank"}, {"POST", "/api/operator/clear"},
		{"POST", "/api/operator/reset"}, {"POST", "/api/operator/onair"}, {"POST", "/api/operator/offair"}, {"POST", "/api/operator/message"},
		{"GET", "/api/operator/clients"}, {"GET", "/api/operator/report"}, {"GET", "/api/operator/reports"},
		{"GET", "/api/admin/ping"}, {"GET", "/api/admin/shows"}, {"GET", "/api/admin/settings"}, {"PUT", "/api/admin/settings/passwords"},
		{"GET", "/api/admin/settings/operator-password"}, {"POST", "/api/admin/rehearsal/start"},
	} {
		if code, _ := v.do(c[0], c[1], map[string]any{"text": "x", "takeover": true}); code != 403 {
			t.Errorf("%s %s as a viewer: %d, want 403", c[0], c[1], code)
		}
	}
	if s.Runtime().ControllerSID() != "" {
		t.Error("a viewer never takes control")
	}
	// the stream opens for a viewer (the first event arrives), and an open stream does not make it the controller
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/operator/stream", nil)
	resp, err := v.hc.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("stream: %v %v", resp, err)
	}
	buf := make([]byte, 256)
	n, _ := resp.Body.Read(buf)
	if !strings.Contains(string(buf[:n]), "opstate") {
		t.Errorf("the first thing on the stream is the state: %q", buf[:n])
	}
	if s.Runtime().ControllerSID() != "" {
		t.Error("watching the stream must not give a viewer control")
	}
}

// /follow takes any of the three passwords and each keeps its own role; the operator and admin forms are as they were,
// and the viewer password opens neither.
func TestFollowTakesEveryPasswordAndTheViewerPasswordOnlyThat(t *testing.T) {
	s, ts := viewerServer(t, nil)
	for pw, want := range map[string]string{"viewerpw": RoleViewer, "operatorpw": RoleOperator, "adminpw": RoleAdmin} {
		if _, got, ok := s.auth.login("viewer", pw); !ok || got != want {
			t.Errorf("/follow with the %s password: %q %v", want, got, ok)
		}
	}
	if _, _, ok := s.auth.login("viewer", "wrong"); ok {
		t.Error("a wrong password signed in on /follow")
	}
	for role, pw := range map[string]string{"admin": "viewerpw", "operator": "viewerpw"} {
		if _, _, ok := s.auth.login(role, pw); ok {
			t.Errorf("the %s form accepted the viewer password", role)
		}
	}
	if _, got, ok := s.auth.login("operator", "adminpw"); !ok || got != RoleAdmin {
		t.Errorf("admin on the operator form: %v %v", got, ok)
	}
	// no viewer password: /follow still takes the operator and admin passwords, and nothing else, an empty password included
	off, _ := viewerServer(t, func(c *config.Config) { c.ViewerPassword = "" })
	for _, pw := range []string{"", "viewerpw", "viewer"} {
		if _, _, ok := off.auth.login("viewer", pw); ok {
			t.Errorf("with no viewer password, %q signed in on /follow", pw)
		}
	}
	if _, got, ok := off.auth.login("viewer", "operatorpw"); !ok || got != RoleOperator {
		t.Errorf("operator password on /follow without a viewer password: %v %v", got, ok)
	}
	// the same password for viewer and operator: it signs in as the operator, never as less
	same, _ := viewerServer(t, func(c *config.Config) { c.ViewerPassword = c.OperatorPassword })
	if _, got, _ := same.auth.login("viewer", "operatorpw"); got != RoleOperator {
		t.Errorf("same password: %v", got)
	}
	// the page itself, from the same access list as the operator page
	c := newClient(t, ts.URL)
	resp, err := c.hc.Get(ts.URL + "/follow")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("GET /follow: %v %v", resp, err)
	}
	resp.Body.Close()
	closed, cts := viewerServer(t, func(c *config.Config) {
		c.Access.AdminAllow, c.Access.OperatorAllow = []string{"192.0.2.0/24"}, []string{"192.0.2.0/24"}
	})
	_ = closed
	if resp, err := newClient(t, cts.URL).hc.Get(cts.URL + "/follow"); err != nil || resp.StatusCode != 403 {
		t.Errorf("/follow follows operator_allow: %v %v", resp, err)
	}
}

// A saved viewer session survives a restart; and the fingerprint is what it was before viewers existed when there is no
// viewer password, so the upgrade signs nobody out.
func TestViewerSessionsAndTheFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")
	a := newAuth("a", "o", "v", path)
	tok, _ := a.issue(RoleViewer, false)
	b := newAuth("a", "o", "v", path)
	if s, ok := b.lookup(tok); !ok || s.role != RoleViewer {
		t.Errorf("a viewer session is kept across a restart: %v %v", s, ok)
	}
	if c := newAuth("a", "o", "other", path); func() bool { _, ok := c.lookup(tok); return ok }() {
		t.Error("changing the viewer password signs viewers out")
	}
	ia, io := config.PasswordID("a"), config.PasswordID("o")
	old := sha256.Sum256(append(append([]byte("fp:"), ia[:]...), io[:]...))
	if got := fingerprint("a", "o", ""); got != hex.EncodeToString(old[:]) {
		t.Errorf("without a viewer the fingerprint must be unchanged: %s", got)
	}
	if fingerprint("a", "o", "v") == fingerprint("a", "o", "") {
		t.Error("a viewer password must change the fingerprint")
	}
}

// Setup sets, replaces and turns off the viewer password; it must differ from the others, and it is written to config.yaml.
func TestSetupManagesTheViewerPassword(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.yaml"
	os.WriteFile(cfgPath, []byte("admin_password: adminpw\noperator_password: operatorpw\n"), 0o600)
	s, ts := viewerServer(t, func(c *config.Config) { c.ViewerPassword = ""; c.DataDir, c.Path = dir, cfgPath })
	admin := newClient(t, ts.URL)
	admin.login("admin", "adminpw")
	if _, st := admin.do("GET", "/api/admin/settings", nil); st["viewerPasswordSet"] != false {
		t.Errorf("settings before: %v", st)
	}
	for body, want := range map[string]int{
		`{"viewer":"short"}`: 400, `{"viewer":"operatorpw"}`: 400, `{"viewer":"adminpw"}`: 400,
		`{"viewer":"viewerpw","removeViewer":true}`: 400, `{"operator":"viewerpw","viewer":"viewerpw"}`: 400,
	} {
		var m map[string]any
		json.Unmarshal([]byte(body), &m)
		code, _ := admin.do("PUT", "/api/admin/settings/passwords", m)
		if code != want {
			t.Errorf("%s: %d, want %d", body, code, want)
		}
	}
	if code, r := admin.do("PUT", "/api/admin/settings/passwords", map[string]any{"viewer": "tablet-pw"}); code != 200 || r["changed"].(map[string]any)["viewer"] != true {
		t.Fatalf("set viewer: %d %v", code, r)
	}
	if b, _ := os.ReadFile(cfgPath); !strings.Contains(string(b), "viewer_password: tablet-pw") || !strings.Contains(string(b), "admin_password: adminpw") {
		t.Errorf("config.yaml:\n%s", b)
	}
	if _, st := admin.do("GET", "/api/admin/settings", nil); st["viewerPasswordSet"] != true || strings.Contains(toJSON(st), "tablet-pw") {
		t.Errorf("settings after (the password itself is never sent): %v", st)
	}
	v := newClient(t, ts.URL)
	if v.login("viewer", "tablet-pw") != 200 {
		t.Fatal("the new viewer password signs in")
	}
	if code, _ := v.do("GET", "/api/operator/state", nil); code != 200 {
		t.Errorf("viewer state: %d", code)
	}
	// changing a password that is not the viewer's leaves viewers signed in; replacing the viewer's signs them out
	if code, _ := admin.do("PUT", "/api/admin/settings/passwords", map[string]any{"operator": "operatorpw2"}); code != 200 {
		t.Fatal("change operator")
	}
	if code, _ := v.do("GET", "/api/operator/state", nil); code != 200 {
		t.Errorf("a viewer stays signed in when another password changes: %d", code)
	}
	if _, _, ok := s.auth.login("viewer", "tablet-pw"); !ok {
		t.Error("the viewer password still works after another password changed")
	}
	v2 := newClient(t, ts.URL)
	v2.login("viewer", "tablet-pw")
	if code, _ := admin.do("PUT", "/api/admin/settings/passwords", map[string]any{"removeViewer": true}); code != 200 {
		t.Fatal("remove viewer")
	}
	if code, _ := v2.do("GET", "/api/operator/state", nil); code != 401 {
		t.Errorf("turning viewer sign-in off signs viewers out: %d", code)
	}
	if _, _, ok := s.auth.login("viewer", "tablet-pw"); ok {
		t.Error("the viewer password must stop working")
	}
	if b, _ := os.ReadFile(cfgPath); strings.Contains(string(b), "tablet-pw") {
		t.Errorf("the removed password is still in config.yaml:\n%s", b)
	}
	if _, st := admin.do("GET", "/api/admin/settings", nil); st["viewerPasswordSet"] != false {
		t.Errorf("settings after removing: %v", st)
	}
}

// postFrom is a JSON POST made as if from remote, with the session cookie if there is one.
func postFrom(s *Server, path, remote, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://localhost"+path, strings.NewReader(body))
	r.RemoteAddr, r.Host = remote, "localhost"
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

// access.viewer_allow limits where the viewer-only password works. The operator and admin passwords are not limited by it
// (they follow operator_allow and admin_allow), and an empty list means anywhere.
func TestViewerAllowLimitsOnlyTheViewerPassword(t *testing.T) {
	s, _ := viewerServer(t, func(c *config.Config) { c.Access.ViewerAllow = []string{"192.0.2.0/24"} })
	login := func(remote, pw string) *httptest.ResponseRecorder {
		return postFrom(s, "/api/login", remote, `{"role":"viewer","password":"`+pw+`"}`, nil)
	}
	if w := login("192.0.2.9:5000", "viewerpw"); w.Code != 200 {
		t.Fatalf("a listed computer signs in as a viewer: %d %s", w.Code, w.Body)
	}
	if w := login("198.51.100.4:5000", "viewerpw"); w.Code != 403 || len(w.Result().Cookies()) != 0 {
		t.Errorf("an unlisted computer is refused the viewer password: %d %s", w.Code, w.Body)
	}
	s.auth.mu.Lock()
	n := len(s.auth.sessions)
	s.auth.mu.Unlock()
	if n != 1 {
		t.Errorf("a refused viewer sign-in leaves no session behind: %d sessions", n)
	}
	for pw, want := range map[string]string{"operatorpw": RoleOperator, "adminpw": RoleAdmin} {
		w := login("198.51.100.4:5000", pw)
		if w.Code != 200 || !strings.Contains(w.Body.String(), want) {
			t.Errorf("the %s password is not limited by viewer_allow: %d %s", want, w.Code, w.Body)
		}
	}
	// a viewer cookie is only good on the listed computers
	ck := login("192.0.2.9:5000", "viewerpw").Result().Cookies()[0]
	get := func(remote string) int {
		r := httptest.NewRequest("GET", "http://localhost/api/operator/state", nil)
		r.RemoteAddr, r.Host = remote, "localhost"
		r.AddCookie(ck)
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w.Code
	}
	if get("192.0.2.9:5000") != 200 || get("198.51.100.4:5000") != 403 {
		t.Error("a viewer session works from a listed computer and not from another")
	}
	// empty = anywhere
	open, _ := viewerServer(t, nil)
	if w := postFrom(open, "/api/login", "198.51.100.4:5000", `{"role":"viewer","password":"viewerpw"}`, nil); w.Code != 200 {
		t.Errorf("no viewer_allow: %d", w.Code)
	}
}

// A page that only watches (/follow) appears in the connected-devices list, marked as a viewer, and in the report of the
// run, but is not an audience device: the audience count and the list's total leave it out. An ordinary operator is not listed.
func TestAViewerIsListedWithTheDevicesButNotCounted(t *testing.T) {
	s, ts := viewerServer(t, nil)
	s.syncReport()
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	op2 := newClient(t, ts.URL) // an ordinary operator page: not a viewer
	op2.login("operator", "operatorpw")
	v := newClient(t, ts.URL)
	v.do("POST", "/api/login", map[string]string{"role": "viewer", "password": "viewerpw"})
	open := func(c *client, query string) context.CancelFunc {
		ctx, cancel := context.WithCancel(context.Background())
		req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/operator/stream"+query, nil)
		resp, err := c.hc.Do(req)
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("stream: %v %v", resp, err)
		}
		buf := make([]byte, 64)
		resp.Body.Read(buf)
		return cancel
	}
	stopOp := open(op, "")
	defer stopOp()
	stopV := open(v, "?follow=1")
	stopFollowingOperator := open(op2, "?follow=1") // an operator who opened /follow is a watcher too
	defer stopFollowingOperator()
	code, r := op.do("GET", "/api/operator/clients", nil)
	rows, _ := r["clients"].([]any)
	if code != 200 || len(rows) != 2 || r["total"] != float64(0) {
		t.Fatalf("clients: %d %v", code, r)
	}
	for _, x := range rows {
		if x.(map[string]any)["role"] != "viewer" {
			t.Errorf("row %v is marked a viewer", x)
		}
	}
	if s.Runtime().AudienceCount() != 0 {
		t.Error("a viewer is not an audience device")
	}
	stopV()
	waitFor(t, func() bool { _, r := op.do("GET", "/api/operator/clients", nil); return len(r["clients"].([]any)) == 1 })
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if ok() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out")
}
