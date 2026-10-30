// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logbuf"
	"github.com/tmshort/open-caption/internal/logx"
)

func findings(m map[string]any) map[string]map[string]any {
	out := map[string]map[string]any{}
	cs, _ := m["checks"].([]any)
	for _, c := range cs {
		f := c.(map[string]any)
		out[f["id"].(string)] = f
	}
	return out
}

func TestTroubleshootReportsWhatTheProgramKnows(t *testing.T) {
	s, _, c := standbyServer(t)
	lb := logbuf.New(100)
	s.SetLogSource(lb)
	log.SetOutput(lb)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	code, m := c.do("GET", "/api/admin/troubleshoot", nil)
	if code != 200 {
		t.Fatal(code)
	}
	f := findings(m)
	for _, id := range []string{"address", "https", "name", "settings", "show", "devices", "problems"} {
		if f[id] == nil {
			t.Errorf("no %q check in %v", id, m["checks"])
		}
	}
	if f["settings"]["status"] != "ok" || f["show"]["status"] != "info" || f["problems"]["status"] != "ok" {
		t.Errorf("a fresh server: %v", f)
	}

	// the settings file can no longer be written
	s.cfg.Path = filepath.Join(t.TempDir(), "gone", "config.yaml")
	if f = findings(func() map[string]any { _, m := c.do("GET", "/api/admin/troubleshoot", nil); return m }()); f["settings"]["status"] != "bad" {
		t.Errorf("an unwritable settings file: %v", f["settings"])
	}

	// problems in the log are counted, for the button's badge, and shown as a check
	logx.Warnf("something to look at")
	logx.Errorf("something broke")
	logx.Infof("all fine")
	_, brief := c.do("GET", "/api/admin/troubleshoot?brief=1", nil)
	if brief["problems"] != float64(2) || brief["checks"] != nil {
		t.Errorf("brief = %v", brief)
	}
	_, m = c.do("GET", "/api/admin/troubleshoot", nil)
	if f = findings(m); f["problems"]["status"] != "warn" || f["problems"]["params"].(map[string]any)["n"] != float64(2) {
		t.Errorf("problems check: %v", f["problems"])
	}
}

func TestPortCheckComparesPortsNotAddressSpelling(t *testing.T) {
	for _, c := range []struct {
		wanted, actual string
		same           bool
	}{
		{":80", "[::]:80", true},
		{":80", "0.0.0.0:80", true},
		{"127.0.0.1:8080", "127.0.0.1:8080", true},
		{"127.0.0.1:0", "127.0.0.1:51234", true},
		{":80", "[::]:8080", false},
		{":443", "[::]:8443", false},
	} {
		if got := samePort(c.wanted, c.actual); got != c.same {
			t.Errorf("samePort(%q, %q) = %v, want %v", c.wanted, c.actual, got, c.same)
		}
	}
	s, _, c := standbyServer(t)
	check := func(wanted, actual string) string {
		s.SetListeners(Listeners{HTTPConfigured: wanted, HTTPAddr: actual})
		_, m := c.do("GET", "/api/admin/troubleshoot", nil)
		return findings(m)["port"]["status"].(string)
	}
	if got := check(":80", "[::]:80"); got != "ok" {
		t.Errorf("getting the port asked for was reported as %q", got)
	}
	if got := check(":80", "[::]:8080"); got != "warn" {
		t.Errorf("falling back to another port was reported as %q", got)
	}
	// HTTPS has its own port, and its own check (when HTTPS is on)
	s.cfg.HTTPS.Enabled = true
	checkTLS := func(wanted, actual string) string {
		s.SetListeners(Listeners{HTTPConfigured: ":80", HTTPAddr: "[::]:80", HTTPSConfigured: wanted, HTTPSAddr: actual})
		_, m := c.do("GET", "/api/admin/troubleshoot", nil)
		f := findings(m)["port-https"]
		if f == nil {
			return "missing"
		}
		return f["status"].(string)
	}
	if got := checkTLS(":443", "[::]:443"); got != "ok" {
		t.Errorf("the HTTPS port asked for was reported as %q", got)
	}
	if got := checkTLS(":443", "[::]:8443"); got != "warn" {
		t.Errorf("the HTTPS fallback port was reported as %q", got)
	}
	s.cfg.HTTPS.Enabled = false
	if got := checkTLS(":443", "[::]:443"); got != "missing" {
		t.Errorf("an HTTPS port was reported while HTTPS is off: %q", got)
	}
}

// With a show loaded, Troubleshoot says whether the audience sees the standby screen or the captions.
func TestTroubleshootSaysWhetherTheShowIsOnAir(t *testing.T) {
	_, _, c := standbyServer(t)
	_, m := c.do("GET", "/api/admin/troubleshoot", nil)
	if findings(m)["onair"] != nil {
		t.Fatalf("with no show loaded there is nothing to say: %v", m["checks"])
	}
	_, sh := c.do("POST", "/api/admin/shows", map[string]any{"name": "Tiny"})
	id, _ := sh["id"].(string)
	c.do("PUT", "/api/admin/shows/"+id+"/showfile", map[string]any{"text": "One\nTwo\n"})
	c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{})
	_, m = c.do("GET", "/api/admin/troubleshoot", nil)
	if f := findings(m)["onair"]; f == nil || f["params"].(map[string]any)["state"] != "standby" {
		t.Fatalf("a loaded show waits in standby: %v", m["checks"])
	}
	c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{"start": true})
	_, m = c.do("GET", "/api/admin/troubleshoot", nil)
	if f := findings(m)["onair"]; f == nil || f["params"].(map[string]any)["state"] != "live" {
		t.Fatalf("a started show is on air: %v", m["checks"])
	}
}

// Troubleshoot points out access lists that are likely to surprise: the Follow page left open and pages limited to this computer.
func TestTroubleshootExplainsAccessLists(t *testing.T) {
	s, _, c := standbyServer(t)
	state := func() string {
		_, m := c.do("GET", "/api/admin/troubleshoot", nil)
		f := findings(m)["access"]
		if f == nil {
			return ""
		}
		return f["params"].(map[string]any)["state"].(string)
	}
	if got := state(); got != "" && got != "setup-local" && got != "limited" {
		t.Fatalf("a fresh server: %q", got)
	}
	set := func(admin, operator, viewer []string, viewerPw string) {
		s.accessMu.Lock()
		s.adminAllow, s.operatorAllow, s.viewerAllow = mustList(t, admin), mustList(t, operator), mustList(t, viewer)
		s.accessMu.Unlock()
		s.cfgMu.Lock()
		s.cfg.Access.AdminAllow, s.cfg.Access.OperatorAllow, s.cfg.Access.ViewerAllow, s.cfg.ViewerPassword = admin, operator, viewer, viewerPw
		s.cfgMu.Unlock()
	}
	set(nil, []string{"192.168.1.0/24"}, nil, "viewerpw")
	if got := state(); got != "follow-open" {
		t.Errorf("a viewer password with only the operator page limited leaves Follow open: %q", got)
	}
	set(nil, []string{"localhost"}, nil, "")
	if got := state(); got != "operator-local" {
		t.Errorf("the operator page limited to this computer: %q", got)
	}
	set([]string{"localhost"}, nil, nil, "")
	if got := state(); got != "setup-local" {
		t.Errorf("Setup limited to this computer: %q", got)
	}
	set(nil, []string{"192.168.1.0/24"}, []string{"192.168.1.0/24"}, "viewerpw")
	if got := state(); got != "limited" {
		t.Errorf("lists that admit a network: %q", got)
	}
	set(nil, nil, nil, "")
	if got := state(); got != "" {
		t.Errorf("no lists, nothing to say: %q", got)
	}
}

func mustList(t *testing.T, e []string) config.AllowList {
	t.Helper()
	l, err := config.ParseAllowList(e)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// Troubleshoot warns when the Setup and operator passwords are the same, and notes a password kept as plain text.
func TestTroubleshootNotesWeakPasswordSetups(t *testing.T) {
	s, _, c := standbyServer(t)
	state := func() string {
		_, m := c.do("GET", "/api/admin/troubleshoot", nil)
		f := findings(m)["passwords"]
		if f == nil {
			return ""
		}
		return f["status"].(string) + ":" + f["params"].(map[string]any)["state"].(string)
	}
	set := func(admin, operator string) {
		s.cfgMu.Lock()
		s.cfg.AdminPassword, s.cfg.OperatorPassword = admin, operator
		s.cfgMu.Unlock()
	}
	set(config.HashPassword("one"), config.HashPassword("two"))
	if got := state(); got != "" {
		t.Errorf("two digests that differ need no remark: %q", got)
	}
	set(config.HashPassword("same"), "same")
	if got := state(); got != "warn:same" {
		t.Errorf("the same password for both: %q", got)
	}
	set(config.HashPassword("one"), "plain-two")
	if got := state(); got != "info:plain" {
		t.Errorf("a plain-text password: %q", got)
	}
}

// Troubleshoot checks that the data folder can be written to and has room.
func TestTroubleshootChecksTheDataFolder(t *testing.T) {
	s, _, c := standbyServer(t)
	get := func() map[string]any {
		_, m := c.do("GET", "/api/admin/troubleshoot", nil)
		return findings(m)["data"]
	}
	if f := get(); f == nil || f["status"] != "ok" {
		t.Fatalf("a writable data folder is fine: %v", f)
	}
	dir := filepath.Join(t.TempDir(), "gone")
	s.cfgMu.Lock()
	s.cfg.DataDir = dir
	s.cfgMu.Unlock()
	if f := get(); f == nil || f["status"] != "bad" || f["params"].(map[string]any)["state"] != "unwritable" {
		t.Errorf("a folder that cannot be written to is a problem: %v", f)
	}
	if free, ok := freeBytes(t.TempDir()); !ok || free == 0 {
		t.Errorf("the free space of a temporary folder should be known and not zero: %d %v", free, ok)
	}
}

// Troubleshoot says when the loaded show's file has errors or uses an image that was not uploaded.
func TestTroubleshootChecksTheLoadedShowFile(t *testing.T) {
	_, _, c := standbyServer(t)
	_, sh := c.do("POST", "/api/admin/shows", map[string]any{"name": "Checked"})
	id, _ := sh["id"].(string)
	state := func(text string) string {
		c.do("PUT", "/api/admin/shows/"+id+"/showfile", map[string]any{"text": text})
		c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{})
		_, m := c.do("GET", "/api/admin/troubleshoot", nil)
		f := findings(m)["showfile"]
		if f == nil {
			return ""
		}
		return f["status"].(string) + ":" + f["params"].(map[string]any)["state"].(string)
	}
	if got := state("One\nTwo\n"); got != "" {
		t.Errorf("a clean show file needs no remark: %q", got)
	}
	if got := state("One\n[image: nothere.png]\n"); got != "warn:missing" {
		t.Errorf("an image that was not uploaded: %q", got)
	}
	if got := state("One\n[loop: 10s]\nunclosed\n"); !strings.HasPrefix(got, "bad:") {
		t.Errorf("an error in the show file: %q", got)
	}
}
