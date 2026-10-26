// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// What happens to a show is written to the log at info level, so "when did we go on air?" and "who changed the show file
// during the performance?" can be answered afterwards.
func TestShowEventsAreLogged(t *testing.T) {
	s, _, c := standbyServer(t)
	_ = s
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	_, r := c.do("POST", "/api/admin/shows", map[string]any{"name": "Our Town"})
	id := r["id"].(string)
	c.do("PUT", "/api/admin/shows/"+id+"/showfile", map[string]any{"text": "[settings]\ntitle: Our Town\n[/settings]\nHello\nWorld\n"})
	c.do("PUT", "/api/admin/shows/"+id+"/showfile?lang=es", map[string]any{"text": "Hola\nMundo\n"})
	c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"font": "serif"}})
	c.do("PUT", "/api/admin/shows/"+id, map[string]any{"name": "Our Town 2"})
	c.do("POST", "/api/admin/shows/"+id+"/duplicate", map[string]any{"name": "Copy"})
	c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{})
	c.do("POST", "/api/operator/claim", map[string]any{"takeover": true})
	c.do("POST", "/api/operator/onair", nil)
	c.do("PUT", "/api/admin/shows/"+id+"/showfile", map[string]any{"text": "[settings]\ntitle: Our Town 2\n[/settings]\nHello\nWorld\nAgain\n", "confirm": "Our Town 2"})
	c.do("POST", "/api/operator/offair", nil)
	c.do("POST", "/api/operator/reset", nil)
	out := buf.String()
	for _, want := range []string{
		`INFO show "Our Town" created`,
		`INFO the show file of "Our Town" was saved`,
		`INFO the es translation of "Our Town" was saved`,
		`INFO the look of show "Our Town" was saved`,
		`INFO show renamed to "Our Town 2"`,
		`INFO show "Copy" created as a copy of ` + id,
		`INFO show "Our Town 2" is on air`,
		`INFO the show file of "Our Town 2" was saved while the show is on air`,
		`INFO show "Our Town 2" is back in standby`,
		`INFO show "Our Town 2" was reset to the start, in standby`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in the log:\n%s", want, out)
		}
	}
}

// A failure of the program itself (not of the request) is logged, so the log shows what the person was told.
func TestServerFailuresAreLogged(t *testing.T) {
	s, _, c := standbyServer(t)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	// a settings file that cannot be written
	s.cfg.Path = filepath.Join(t.TempDir(), "gone", "config.yaml")
	if code, _ := c.do("PUT", "/api/admin/settings/standby", map[string]any{"message": "Back soon"}); code != 500 {
		t.Fatalf("code %d", code)
	}
	if out := buf.String(); !strings.Contains(out, "ERROR answered 500: Could not save the setting in config.yaml") {
		t.Errorf("the failure was not logged:\n%s", out)
	}
	// a request that was simply wrong is not a failure of the program
	buf.Reset()
	if code, _ := c.do("PUT", "/api/admin/settings/standby", map[string]any{"message": strings.Repeat("x", 500)}); code != 400 {
		t.Fatalf("code %d", code)
	}
	if buf.Len() != 0 {
		t.Errorf("a bad request was logged as an error: %q", buf.String())
	}
}
