package config

import (
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
