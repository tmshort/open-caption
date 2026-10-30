// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/config"
)

// Setup can change who may open each page; the change is kept in config.yaml (with its comments) and applies at once.
func TestAccessListsCanBeChangedInSetupAndAreKept(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("# kept\nadmin_password: adminpw\noperator_password: operatorpw\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	code, r := c.do("PUT", "/api/admin/settings/access", map[string]any{
		"adminAllow": []string{"localhost"}, "operatorAllow": []string{"private", "192.168.1.50"}, "viewerAllow": []string{"192.168.1.60"}, "localAutoLogin": true})
	if code != 200 || r["saved"] != true || r["adminRestricted"] != true {
		t.Fatalf("%d %v", code, r)
	}
	b, _ := os.ReadFile(cfgPath)
	for _, want := range []string{"# kept", "admin_allow:", "- localhost", "operator_allow:", "- private", "192.168.1.50", "viewer_allow:", "192.168.1.60", "local_auto_login: true"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("config.yaml lacks %q:\n%s", want, b)
		}
	}
	a, o, v := s.allowLists()
	if !a.Allows(netip.MustParseAddr("127.0.0.1")) || a.Allows(netip.MustParseAddr("192.168.1.9")) {
		t.Error("the Setup list did not apply")
	}
	if !o.Allows(netip.MustParseAddr("10.1.2.3")) || o.Allows(netip.MustParseAddr("8.8.8.8")) {
		t.Error("the operator list did not apply")
	}
	if !v.Allows(netip.MustParseAddr("192.168.1.60")) || v.Allows(netip.MustParseAddr("192.168.1.61")) {
		t.Error("the viewer list did not apply")
	}
	_, g := c.do("GET", "/api/admin/settings", nil)
	acc, _ := g["access"].(map[string]any)
	if acc["localAutoLogin"] != true || toJSON(acc["viewerAllow"]) != `["192.168.1.60"]` || g["adminRestricted"] != true {
		t.Errorf("%v", g)
	}
}

// A list that would shut out the computer making the change is refused and nothing is saved.
func TestAccessChangeCannotLockOutThisComputer(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	orig := "admin_password: adminpw\noperator_password: operatorpw\n"
	if err := os.WriteFile(cfgPath, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}
	s, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	// the test client is 127.0.0.1
	code, r := c.do("PUT", "/api/admin/settings/access", map[string]any{"adminAllow": []string{"192.168.1.50"}})
	if code != 409 || r["error"] == nil {
		t.Fatalf("%d %v", code, r)
	}
	if b, _ := os.ReadFile(cfgPath); string(b) != orig {
		t.Errorf("config.yaml changed:\n%s", b)
	}
	if s.adminRestricted() {
		t.Error("the refused list applied")
	}
	// an empty Setup list lets anyone try, and so includes this computer
	if code, _ := c.do("PUT", "/api/admin/settings/access", map[string]any{}); code != 200 {
		t.Errorf("an empty list: %d", code)
	}
}

func TestAccessChangeRefusesAnEntryThatIsNotAnAddress(t *testing.T) {
	_, _, c := serverWith(t, nil)
	for _, bad := range []string{"not-an-address", "300.1.1.1", "192.168.1.0/99"} {
		if code, r := c.do("PUT", "/api/admin/settings/access", map[string]any{"operatorAllow": []string{bad}}); code != 400 || r["error"] == nil {
			t.Errorf("%q: %d %v", bad, code, r)
		}
	}
	tooMany := make([]string, maxAccessEntries+1)
	for i := range tooMany {
		tooMany[i] = "10.0.0." + strings.Repeat("1", 1) + string(rune('a'+i%26))
	}
	if code, _ := c.do("PUT", "/api/admin/settings/access", map[string]any{"operatorAllow": tooMany}); code != 400 {
		t.Errorf("too many: %d", code)
	}
}

// Turning local_auto_login off ends the sign-in a browser on this computer was given without a password.
func TestTurningLocalAutoLoginOffEndsTheAutomaticSignIn(t *testing.T) {
	cfg := testConfig()
	cfg.Access.LocalAutoLogin = true
	cfg.DataDir = t.TempDir()
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	c := newClient(t, ts.URL)
	if code, _ := c.do("GET", "/api/admin/settings", nil); code != 200 {
		t.Fatalf("automatic sign-in: %d", code)
	}
	if code, _ := c.do("PUT", "/api/admin/settings/access", map[string]any{"localAutoLogin": false}); code != 200 {
		t.Fatalf("change: %d", code)
	}
	if s.localAutoLogin() {
		t.Error("still on")
	}
	if code, _ := c.do("GET", "/api/admin/settings", nil); code != 401 {
		t.Errorf("after turning it off: %d", code)
	}
}
