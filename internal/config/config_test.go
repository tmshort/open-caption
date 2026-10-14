package config

import (
	"os"
	"path/filepath"
	"runtime"
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
