package acme

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCredentialsPrecedenceEnvironmentThenFileThenConfig(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "dns.env")
	os.WriteFile(f, []byte("# a comment\n\nexport OCT_A=from-file\nOCT_B=\"quoted value\"\nOCT_C='single'\nOCT_D=from-file\n"), 0o600)
	t.Setenv("OCT_D", "from-env")                   // the environment wins over the file
	t.Setenv("OCT_E_FILE", filepath.Join(dir, "x")) // a NAME_FILE variable counts as set
	for _, k := range []string{"OCT_A", "OCT_B", "OCT_C", "OCT_F", "OCT_E"} {
		os.Unsetenv(k)
		t.Cleanup(func() { os.Unsetenv(k) })
	}
	set, err := Credentials{Inline: map[string]string{"OCT_A": "from-config", "OCT_E": "from-config", "OCT_F": "only-config"}, File: f}.Apply()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"OCT_A": "from-file", "OCT_B": "quoted value", "OCT_C": "single", "OCT_D": "from-env", "OCT_F": "only-config", "OCT_E": ""}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if strings.Contains(strings.Join(set, ","), "OCT_D") || strings.Contains(strings.Join(set, ","), "OCT_E,") {
		t.Errorf("names set: %v", set)
	}
}

func TestEnvFileErrorsNameTheLineAndOpenPermissionsAreWarnedAbout(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.env")
	os.WriteFile(f, []byte("OK=1\nthis is not valid\n"), 0o644)
	var warned string
	_, err := ParseEnvFile(f, func(format string, a ...any) { warned = format })
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Errorf("%v", err)
	}
	if runtime.GOOS != "windows" && !strings.Contains(warned, "chmod 600") {
		t.Errorf("a world-readable credentials file should be warned about: %q", warned)
	}
	if _, err := ParseEnvFile(filepath.Join(t.TempDir(), "missing.env"), nil); err == nil {
		t.Error("a missing file is an error")
	}
}
