package config

import (
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateThenReload(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "config.yaml")
	c, err := Load(p)
	if err != nil || !c.Generated {
		t.Fatalf("%v generated=%v", err, c.Generated)
	}
	if len(c.AdminPassword) != 12 || c.AdminPassword == c.OperatorPassword {
		t.Errorf("passwords: %q %q", c.AdminPassword, c.OperatorPassword)
	}
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(p); fi.Mode().Perm() != 0o600 {
			t.Errorf("mode %v", fi.Mode().Perm())
		}
	}
	c2, err := Load(p)
	if err != nil || c2.Generated || c2.AdminPassword != c.AdminPassword || c2.HTTP.Addr != ":80" {
		t.Errorf("%v %+v", err, c2)
	}
}

func TestPartialFileKeepsDefaults(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\nhttp:\n  addr: \":9000\"\n"), 0o600)
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.HTTP.Addr != ":9000" || c.HTTP.FallbackAddr != ":8080" || c.Defaults.LoopItemSeconds != 10 || c.Audience.MaxClients != 500 {
		t.Errorf("%+v", c)
	}
}

func TestValidation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(p, []byte("admin_password: a\n"), 0o600)
	if _, err := Load(p); err == nil {
		t.Error("expected error for missing operator password")
	}
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\nhttps:\n  enabled: true\n  cert_file: x.pem\n"), 0o600)
	if _, err := Load(p); err == nil {
		t.Error("expected error for cert without key")
	}
}

func TestAllowLists(t *testing.T) {
	ip := func(s string) netip.Addr { return netip.MustParseAddr(s) }
	cases := []struct {
		entries []string
		ip      string
		want    bool
	}{
		{nil, "8.8.8.8", true}, // empty: anyone
		{[]string{"localhost"}, "127.0.0.1", true},
		{[]string{"localhost"}, "::1", true},
		{[]string{"localhost"}, "192.168.1.5", false},
		{[]string{"private"}, "192.168.1.5", true},
		{[]string{"private"}, "10.20.30.40", true},
		{[]string{"private"}, "172.20.0.1", true},
		{[]string{"private"}, "172.32.0.1", false},
		{[]string{"private"}, "8.8.8.8", false},
		{[]string{"192.168.1.0/24"}, "192.168.1.200", true},
		{[]string{"192.168.1.0/24"}, "192.168.2.1", false},
		{[]string{"192.168.1.50", "localhost"}, "192.168.1.50", true},
		{[]string{"192.168.1.50"}, "192.168.1.51", false},
		{[]string{"192.168.1.0/24"}, "::ffff:192.168.1.7", true}, // IPv4 seen on a dual-stack socket
		{[]string{" LOCALHOST ", ""}, "127.0.0.1", true},
		{[]string{"fe80::/10"}, "fe80::1", true},
	}
	for _, c := range cases {
		l, err := ParseAllowList(c.entries)
		if err != nil {
			t.Fatalf("%v: %v", c.entries, err)
		}
		if got := l.Allows(ip(c.ip)); got != c.want {
			t.Errorf("%v allows %s = %v, want %v", c.entries, c.ip, got, c.want)
		}
	}
	for _, bad := range []string{"everyone", "192.168.1.0/99", "300.1.1.1", "192.168.1"} {
		l, err := ParseAllowList([]string{bad})
		if err == nil || l.Allows(ip("127.0.0.1")) {
			t.Errorf("%q should be an error and deny everything", bad)
		}
	}
	if l, _ := ParseAllowList(nil); l.Restricted() {
		t.Error("an empty list restricts nobody")
	}
	if l, _ := ParseAllowList([]string{"localhost"}); !l.Restricted() {
		t.Error("localhost is a restriction")
	}
}

func TestBadAccessSettingStopsStartupWithAClearMessage(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\naccess:\n  admin_allow: [\"everyone\"]\n"), 0o600)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "access.admin_allow") || !strings.Contains(err.Error(), `"everyone"`) {
		t.Errorf("%v", err)
	}
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\naccess:\n  admin_allow: [localhost]\n  operator_allow: [private, 192.168.1.0/24]\n"), 0o600)
	c, err := Load(p)
	if err != nil || len(c.Access.AdminAllow) != 1 || len(c.Access.OperatorAllow) != 2 {
		t.Errorf("%v %+v", err, c.Access)
	}
}

func TestLookDefaultsAreValidatedAndNormalised(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\ndefaults:\n  look:\n    background: \"#001\"\n    font: Dyslexia Friendly\n    text_size_percent: 130\n    speaker_names: colour\n"), 0o600)
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	l := c.Defaults.Look
	if l.Background != "#000011" || l.Font != "dyslexic" || l.TextSizePercent != 130 || l.SpeakerNames != "color" || l.TextColor != "#ffffff" {
		t.Errorf("%+v", l)
	}
	for yaml, want := range map[string]string{
		"defaults:\n  look:\n    font: comic\n":          "defaults.look.font",
		"defaults:\n  look:\n    background: nope\n":     "defaults.look.background",
		"defaults:\n  look:\n    text_size_percent: 5\n": "defaults.look.text_size_percent",
		"defaults:\n  loop_item_seconds: -1\n":           "defaults.loop_item_seconds",
		"audience:\n  scrollback_lines: 3\n":             "audience.scrollback_lines",
	} {
		os.WriteFile(p, []byte("admin_password: a\noperator_password: b\n"+yaml), 0o600)
		if _, err := Load(p); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: %v", yaml, err)
		}
	}
	// A config file that says nothing about the look uses the built-in defaults.
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\n"), 0o600)
	c, _ = Load(p)
	if c.Defaults.Look.Background != "#000000" || c.Defaults.Look.TextSizePercent != 100 || c.Defaults.Look.Font != "system" {
		t.Errorf("%+v", c.Defaults.Look)
	}
}

func load(t *testing.T, yamlText string) (Config, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "c.yaml")
	if err := os.WriteFile(p, []byte("admin_password: adminpw\noperator_password: operatorpw\n"+yamlText), 0o600); err != nil {
		t.Fatal(err)
	}
	return Load(p)
}

func TestMisspelledAndUnknownSettingsAreReported(t *testing.T) {
	_, err := load(t, "htp:\n  addr: \":9000\"\naudience:\n  max_client: 5\n")
	var pe *ProblemsError
	if !errors.As(err, &pe) || len(pe.Problems) != 2 {
		t.Fatalf("%v", err)
	}
	msg := err.Error()
	for _, want := range []string{"2 problems", "unknown setting htp", "unknown setting max_client", "line "} {
		if !strings.Contains(msg, want) {
			t.Errorf("message lacks %q:\n%s", want, msg)
		}
	}
}

func TestWrongTypesAndSyntaxAreReported(t *testing.T) {
	_, err := load(t, "audience:\n  max_clients: lots\n")
	if err == nil || !strings.Contains(err.Error(), "line 4") {
		t.Errorf("%v", err)
	}
	_, err = load(t, "http: [oops\n")
	if err == nil || !strings.Contains(err.Error(), "problem") {
		t.Errorf("%v", err)
	}
}

func TestEverythingWrongIsListedAtOnce(t *testing.T) {
	_, err := load(t, `http:
  addr: "localhost"
  fallback_addr: ":99999"
https:
  enabled: true
  cert_file: /nonexistent/cert.pem
  key_file: /nonexistent/key.pem
audience:
  max_clients: 0
  scrollback_lines: 5
defaults:
  language: English
  loop_item_seconds: 0
  look:
    background: banana
    font: comic
log:
  level: loud
access:
  admin_allow: [not-an-address]
`)
	var pe *ProblemsError
	if !errors.As(err, &pe) {
		t.Fatalf("%v", err)
	}
	for _, key := range []string{"http.addr", "http.fallback_addr", "https.cert_file", "https.key_file", "audience.max_clients", "audience.scrollback_lines",
		"defaults.language", "defaults.loop_item_seconds", "defaults.look.background", "defaults.look.font", "log.level", "access.admin_allow"} {
		found := false
		for _, p := range pe.Problems {
			found = found || strings.HasPrefix(p, key+":")
		}
		if !found {
			t.Errorf("no problem reported for %s:\n%v", key, err)
		}
	}
}

func TestValidFilesAndWarnings(t *testing.T) {
	c, err := load(t, "http:\n  addr: \"127.0.0.1:8080\"\n  fallback_addr: \"\"\nlog:\n  level: DEBUG\n")
	if err != nil || c.Log.Level != "DEBUG" {
		t.Fatalf("%v %+v", err, c.Log)
	}
	if w := strings.Join(c.Warnings(), "|"); !strings.Contains(w, "access.admin_allow is empty") || strings.Contains(w, "same") {
		t.Errorf("%s", w)
	}
	c, _ = load(t, "access:\n  admin_allow: [localhost]\n")
	c.OperatorPassword, c.AdminPassword = "abc", "abc"
	if w := strings.Join(c.Warnings(), "|"); !strings.Contains(w, "are the same") || !strings.Contains(w, "shorter than 6") {
		t.Errorf("%s", w)
	}
	if c, err := load(t, ""); err != nil || c.Log.Level != "info" {
		t.Errorf("default level: %v %q", err, c.Log.Level)
	}
}

func TestEmptyFileNamesTheMissingPasswords(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(p, nil, 0o600)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "admin_password: must not be empty") || !strings.Contains(err.Error(), "operator_password") {
		t.Errorf("%v", err)
	}
}

func TestChangingPasswordsKeepsCommentsAndOtherSettings(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.yaml")
	orig := `# My theater's server. Do not share this file.
admin_password: oldadmin   # the boss
operator_password: oldoperator

# Only the front-of-house laptop may run Setup.
access:
  admin_allow: [192.168.1.50]   # box office
http:
    addr: ":9000"
`
	os.WriteFile(p, []byte(orig), 0o640)
	if err := SetPasswords(p, "123456", ""); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	got := string(b)
	for _, want := range []string{"# My theater's server. Do not share this file.", "# the boss", "# Only the front-of-house laptop may run Setup.", "# box office", "operator_password: oldoperator", "192.168.1.50", `"123456"`} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "defaults:") || strings.Contains(got, "scrollback") {
		t.Errorf("defaults must not be written into a file that did not have them:\n%s", got)
	}
	c, err := Load(p)
	if err != nil || c.AdminPassword != "123456" || c.OperatorPassword != "oldoperator" || c.HTTP.Addr != ":9000" {
		t.Fatalf("%v %+v", err, c)
	}
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(p); fi.Mode().Perm() != 0o640 {
			t.Errorf("the file's mode is kept: %v", fi.Mode().Perm())
		}
	}
}

func TestSetValuesAddsMissingKeysAndRefusesToBreakTheFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\nlog:\n"), 0o600)
	if err := SetValues(p, map[string]string{"log.level": "debug", "audience.note": ""}); err == nil {
		t.Error("an unknown key must be refused (the file would not start)")
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "log:\n") || strings.Contains(string(b), "note") {
		t.Errorf("the file must be left as it was:\n%s", b)
	}
	if err := SetValues(p, map[string]string{"log.level": "debug"}); err != nil {
		t.Fatal(err)
	}
	if c, err := Load(p); err != nil || c.Log.Level != "debug" {
		t.Errorf("%v %+v", err, c.Log)
	}
	if err := SetValues(p, map[string]string{"log.level": "loud"}); err == nil {
		t.Error("an invalid value must be refused")
	}
	if c, _ := Load(p); c.Log.Level != "debug" {
		t.Error("the file must keep its last good value")
	}
	// A file that has none of the keys (and one that is empty of settings) gets them added.
	os.WriteFile(p, []byte("admin_password: a\noperator_password: b\n"), 0o600)
	if err := SetValues(p, map[string]string{"log.level": "warn"}); err != nil {
		t.Fatal(err)
	}
	if c, _ := Load(p); c.Log.Level != "warn" {
		t.Errorf("%+v", c.Log)
	}
}

func TestFileFromANewerVersionStartsWithWarnings(t *testing.T) {
	c, err := load(t, "config_version: 7\nsparkles:\n  level: 11\nhttp:\n  addr: \":9001\"\n")
	if err != nil {
		t.Fatalf("a newer file's unknown settings must not stop the program: %v", err)
	}
	if c.HTTP.Addr != ":9001" {
		t.Errorf("known settings still apply: %+v", c.HTTP)
	}
	w := strings.Join(c.Warnings(), "|")
	if !strings.Contains(w, "unknown setting sparkles") || !strings.Contains(w, "config_version is 7") {
		t.Errorf("%s", w)
	}
	// Real mistakes in a newer file are still errors, and an ordinary file is still strict.
	if _, err := load(t, "config_version: 7\nsparkles: 1\naudience:\n  max_clients: lots\n"); err == nil || !strings.Contains(err.Error(), "whole number") || strings.Contains(err.Error(), "sparkles") {
		t.Errorf("%v", err)
	}
	if _, err := load(t, "sparkles: 1\n"); err == nil || !strings.Contains(err.Error(), "unknown setting sparkles") {
		t.Errorf("%v", err)
	}
	if c, err := load(t, "config_version: 1\n"); err != nil || len(c.Warnings()) == 0 {
		t.Errorf("%v %v", err, c.Warnings())
	}
}

func TestGeneratedFileSaysWhichVersionWroteIt(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	if _, err := Load(p); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), "config_version: 1") {
		t.Errorf("%s", b)
	}
}
