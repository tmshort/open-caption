package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// The options are described in three places: --help (generated from the code), the man page and the
// configuration reference. This keeps them in step: every option and environment variable the program has must be
// in all three, and --help must say where the rest is documented.
func TestEveryOptionIsDocumentedInTheManPageAndTheReference(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "open-caption")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("--help should exit 0: %v\n%s", err, out)
	}
	help := string(out)
	for _, w := range []string{"docs/configuration-reference.md", "man open-caption", "Usage: open-caption"} {
		if !strings.Contains(help, w) {
			t.Errorf("--help does not mention %q", w)
		}
	}
	flags := regexp.MustCompile(`(?m)^  -([a-z][a-z-]*)`).FindAllStringSubmatch(help, -1)
	envs := regexp.MustCompile(`(?m)^  (OPEN_CAPTION_[A-Z_]+)`).FindAllStringSubmatch(help, -1)
	if len(flags) < 10 || len(envs) < 6 {
		t.Fatalf("could not read the options from --help (%d flags, %d variables):\n%s", len(flags), len(envs), help)
	}
	man, err := os.ReadFile("../../docs/open-caption.1")
	if err != nil {
		t.Fatal(err)
	}
	manText := strings.ReplaceAll(string(man), `\-`, "-") // roff writes a hyphen as \-
	ref, err := os.ReadFile("../../docs/configuration-reference.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range flags {
		if !strings.Contains(manText, "--"+f[1]) {
			t.Errorf("the man page does not describe --%s", f[1])
		}
		if !strings.Contains(string(ref), "`--"+f[1]) {
			t.Errorf("docs/configuration-reference.md does not describe --%s", f[1])
		}
	}
	for _, e := range envs {
		if !strings.Contains(string(man), e[1]) || !strings.Contains(string(ref), e[1]) {
			t.Errorf("%s is missing from the man page or the reference", e[1])
		}
	}
	if !strings.Contains(manText, "--help") || !strings.Contains(string(ref), "`--help`") {
		t.Error("--help itself is documented in the man page and the reference")
	}
}
