// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"log"
	"os"
	"path/filepath"
	"testing"

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
