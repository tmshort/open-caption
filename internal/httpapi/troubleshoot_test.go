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
