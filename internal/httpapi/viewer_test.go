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
	if code, r := v.do("POST", "/api/login", map[string]string{"role": "operator", "password": "viewerpw"}); code != 200 || r["role"] != RoleViewer {
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

// The other passwords still work as before, and the viewer password does not open Setup or the admin sign-in.
func TestViewerPasswordDoesNotOpenAnythingElse(t *testing.T) {
	s, _ := viewerServer(t, nil)
	for role, pw := range map[string]string{"admin": "viewerpw", "operator": "wrong"} {
		if _, _, ok := s.auth.login(role, pw); ok {
			t.Errorf("%s accepted %q", role, pw)
		}
	}
	if _, got, ok := s.auth.login("operator", "operatorpw"); !ok || got != RoleOperator {
		t.Errorf("operator: %v %v", got, ok)
	}
	if _, got, ok := s.auth.login("operator", "adminpw"); !ok || got != RoleAdmin {
		t.Errorf("admin on the operator form: %v %v", got, ok)
	}
	// no viewer password: nothing signs in as a viewer, whatever is typed (an empty password included)
	off, _ := viewerServer(t, func(c *config.Config) { c.ViewerPassword = "" })
	for _, pw := range []string{"", "viewerpw", "viewer"} {
		if _, _, ok := off.auth.login("operator", pw); ok {
			t.Errorf("with viewer sign-in off, %q signed in", pw)
		}
	}
	// the same password for viewer and operator: it signs in as the operator, never less
	same, _ := viewerServer(t, func(c *config.Config) { c.ViewerPassword = c.OperatorPassword })
	if _, got, _ := same.auth.login("operator", "operatorpw"); got != RoleOperator {
		t.Errorf("same password: %v", got)
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
	if v.login("operator", "tablet-pw") != 200 {
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
	if _, _, ok := s.auth.login("operator", "tablet-pw"); !ok {
		t.Error("the viewer password still works after another password changed")
	}
	v2 := newClient(t, ts.URL)
	v2.login("operator", "tablet-pw")
	if code, _ := admin.do("PUT", "/api/admin/settings/passwords", map[string]any{"removeViewer": true}); code != 200 {
		t.Fatal("remove viewer")
	}
	if code, _ := v2.do("GET", "/api/operator/state", nil); code != 401 {
		t.Errorf("turning viewer sign-in off signs viewers out: %d", code)
	}
	if _, _, ok := s.auth.login("operator", "tablet-pw"); ok {
		t.Error("the viewer password must stop working")
	}
	if b, _ := os.ReadFile(cfgPath); strings.Contains(string(b), "tablet-pw") {
		t.Errorf("the removed password is still in config.yaml:\n%s", b)
	}
	if _, st := admin.do("GET", "/api/admin/settings", nil); st["viewerPasswordSet"] != false {
		t.Errorf("settings after removing: %v", st)
	}
}
