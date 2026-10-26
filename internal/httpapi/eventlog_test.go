// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"bytes"
	"log"
	"os"
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
