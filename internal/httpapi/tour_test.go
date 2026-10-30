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

// The tour of Setup is offered until it has been finished or skipped; the flag is kept in config.yaml.
func TestTourFlagIsReportedAndSaved(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("# kept\nadmin_password: adminpw\noperator_password: operatorpw\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	if _, r := c.do("GET", "/api/admin/ping", nil); r["tourDone"] != false {
		t.Fatalf("a new copy has not had the tour: %v", r)
	}
	if code, r := c.do("PUT", "/api/admin/settings/tour", map[string]any{"done": true}); code != 200 || r["saved"] != true {
		t.Fatalf("%d %v", code, r)
	}
	if _, r := c.do("GET", "/api/admin/ping", nil); r["tourDone"] != true {
		t.Fatalf("tourDone should be true once saved: %v", r)
	}
	b, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(b), "# kept") || !strings.Contains(string(b), "tour_done: true") {
		t.Fatalf("config.yaml:\n%s", b)
	}
	if _, err := config.Load(cfgPath); err != nil {
		t.Errorf("the file no longer loads: %v", err)
	}
	c.do("PUT", "/api/admin/settings/tour", map[string]any{"done": false})
	if _, r := c.do("GET", "/api/admin/ping", nil); r["tourDone"] != false {
		t.Fatalf("tourDone should be false again: %v", r)
	}
}
