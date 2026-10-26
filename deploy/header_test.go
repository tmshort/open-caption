// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every source file starts with the copyright and licence lines (CONTRIBUTING.md); this keeps a new file from
// skipping them. Documents (Markdown), data, images and fonts have none, and neither has the one generated file.
func TestSourceFilesHaveCopyrightHeader(t *testing.T) {
	covered := map[string]bool{".go": true, ".js": true, ".mjs": true, ".css": true, ".html": true, ".sh": true, ".ps1": true,
		".bat": true, ".yml": true, ".yaml": true}
	skipDir := map[string]bool{".git": true, "node_modules": true, "bin": true, "dist": true, ".tools": true, "coverage": true}
	skipFile := map[string]bool{"zz_providers_gen.go": true, "dependabot.yml": true}
	checked := 0
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if skipDir[name] {
				return filepath.SkipDir
			}
			return nil
		}
		if skipFile[name] || !(covered[filepath.Ext(name)] || name == "Makefile") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		head := string(b)
		if len(head) > 400 {
			head = head[:400]
		}
		if !strings.Contains(head, "Copyright the open-caption authors") || !strings.Contains(head, "SPDX-License-Identifier: Apache-2.0") {
			t.Errorf("%s: missing the copyright header (\"Copyright the open-caption authors\" and \"SPDX-License-Identifier: Apache-2.0\")", path)
		}
		checked++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 100 {
		t.Errorf("only %d source files found; is the walk wrong?", checked)
	}
}
