// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/config"
)

func defaultsBody() map[string]any {
	return map[string]any{"language": "es", "idleMessage": "Pronto", "loopSeconds": 7.5, "openBrowser": false,
		"background": "#102030", "textColor": "white", "speakerNameColor": "#ffcc00", "font": "serif", "textSizePercent": 150,
		"speakerNames": "color", "speakerPosition": "beside", "scrollbackLines": 300, "maxClients": 250}
}

// The venue defaults are saved in config.yaml with the comments kept, and the running program keeps its own until it restarts.
func TestVenueDefaultsAreSavedForTheNextStart(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("# kept\nadmin_password: adminpw\noperator_password: operatorpw\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	_, g := c.do("GET", "/api/admin/settings", nil)
	if d, _ := g["defaults"].(map[string]any); d["restartNeeded"] != false || d["language"] != "en" {
		t.Fatalf("%v", g["defaults"])
	}
	code, r := c.do("PUT", "/api/admin/settings/defaults", defaultsBody())
	if code != 200 || r["saved"] != true {
		t.Fatalf("%d %v", code, r)
	}
	b, _ := os.ReadFile(cfgPath)
	for _, want := range []string{"# kept", "language: es", "idle_message: Pronto", "loop_item_seconds: 7.5", "open_browser: false", "background: '#102030'", "text_color: '#ffffff'", "font: serif", "text_size_percent: 150", "speaker_names: color", "speaker_position: beside", "scrollback_lines: 300", "max_clients: 250"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("config.yaml lacks %q:\n%s", want, b)
		}
	}
	if _, err := config.Load(cfgPath); err != nil {
		t.Errorf("the file no longer loads: %v", err)
	}
	if got := s.snapshot().Defaults.Language; got != "en" {
		t.Errorf("the running program changed its default language to %q", got)
	}
	_, g = c.do("GET", "/api/admin/settings", nil)
	d, _ := g["defaults"].(map[string]any)
	if d["restartNeeded"] != true || d["language"] != "es" || d["maxClients"] != float64(250) || d["loopSeconds"] != 7.5 {
		t.Errorf("%v", d)
	}
}

func TestVenueDefaultsRefuseWhatIsNotValidAndSaveNothing(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	orig := "admin_password: adminpw\noperator_password: operatorpw\n"
	if err := os.WriteFile(cfgPath, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	for k, v := range map[string]any{"language": "english!", "loopSeconds": 0, "background": "not a colour", "font": "comic", "textSizePercent": 400,
		"speakerNames": "loud", "speakerPosition": "up", "scrollbackLines": 5, "maxClients": 0, "idleMessage": strings.Repeat("x", 500)} {
		body := defaultsBody()
		body[k] = v
		if code, r := c.do("PUT", "/api/admin/settings/defaults", body); code != 400 || r["error"] == nil {
			t.Errorf("%s=%v: %d %v", k, v, code, r)
		}
	}
	if b, _ := os.ReadFile(cfgPath); string(b) != orig {
		t.Errorf("config.yaml changed:\n%s", b)
	}
}

func TestVenueDefaultsNeedAConfigFile(t *testing.T) {
	_, _, c := serverWith(t, nil)
	if code, _ := c.do("PUT", "/api/admin/settings/defaults", defaultsBody()); code != 409 {
		t.Errorf("without a settings file: %d", code)
	}
}
