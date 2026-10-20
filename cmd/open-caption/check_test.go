// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// --check-show is a lint: it prints the show file's problems and its exit status says whether it passed. It must
// work with no config file, no data folder and no network.
func TestCheckShowReportsAndSetsTheExitStatus(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "open-caption")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	good := filepath.Join(dir, "good.md")
	os.WriteFile(good, []byte("[settings]\ntitle: T\n[/settings]\n\n[clear]\nHello there.\n"), 0o644)
	warn := filepath.Join(dir, "warn.md")
	os.WriteFile(warn, []byte("[settings]\ntitle: T\n[/settings]\n\nHello there.\n[bogus: x]\n"), 0o644)
	bad := filepath.Join(dir, "bad.md")
	os.WriteFile(bad, []byte("[settings]\ntitle: T\n[/settings]\n\n[song: S]\nla la\n"), 0o644)

	status := func(args ...string) (int, string) {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "HOME="+dir, "XDG_CONFIG_HOME="+dir, "APPDATA="+dir) // no settings file anywhere
		out, err := cmd.CombinedOutput()
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), string(out)
		}
		if err != nil {
			t.Fatal(err)
		}
		return 0, string(out)
	}
	if code, out := status("--check-show", good); code != 0 || !strings.Contains(out, "0 errors, 0 warnings") {
		t.Errorf("a clean file: %d\n%s", code, out)
	}
	if code, out := status("--check-show", warn); code != 0 || !strings.Contains(out, "warning: Unknown command") {
		t.Errorf("warnings pass: %d\n%s", code, out)
	}
	if code, _ := status("--strict", "--check-show", warn); code != 1 {
		t.Errorf("--strict fails on warnings: %d", code)
	}
	if code, out := status("--check-show", bad); code != 1 || !strings.Contains(out, "bad.md:5: error:") {
		t.Errorf("errors fail, with the line: %d\n%s", code, out)
	}
	if code, out := status("--check-show", filepath.Join(dir, "missing.md")); code != 2 || !strings.Contains(out, "cannot read") {
		t.Errorf("an unreadable file: %d\n%s", code, out)
	}
	// nothing was created: no settings file, no data folder
	if ents, _ := os.ReadDir(dir); len(ents) != 4 {
		t.Errorf("--check-show must not write anything: %v", ents)
	}
}
