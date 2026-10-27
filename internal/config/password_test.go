// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hexOf(sum []byte) string { return hex.EncodeToString(sum) }

// A password with no prefix is plain text; sha1:, sha256: and sha512: hold a digest; plain: is plain text on purpose.
func TestPasswordFormsAreChecked(t *testing.T) {
	s256, s512, s1 := sha256.Sum256([]byte("hunter22")), sha512.Sum512([]byte("hunter22")), sha1.Sum([]byte("hunter22"))
	for name, stored := range map[string]string{
		"plain text":       "hunter22",
		"plain: prefix":    "plain:hunter22",
		"sha256":           "sha256:" + hexOf(s256[:]),
		"sha256 uppercase": "sha256:" + strings.ToUpper(hexOf(s256[:])),
		"sha512":           "sha512:" + hexOf(s512[:]),
		"sha1":             "sha1:" + hexOf(s1[:]),
	} {
		if p := PasswordProblem(stored); p != "" {
			t.Errorf("%s: %s", name, p)
		}
		if !CheckPassword(stored, "hunter22") || CheckPassword(stored, "hunter23") || CheckPassword(stored, "") {
			t.Errorf("%s: the right password must pass and a wrong one not", name)
		}
		if IsHashed(stored) != (strings.HasPrefix(stored, "sha")) {
			t.Errorf("%s: IsHashed = %v", name, IsHashed(stored))
		}
	}
	if pw, ok := PlainPassword("plain:sha256:abc"); !ok || pw != "sha256:abc" {
		t.Errorf("plain: keeps what follows, even a digest-looking text: %q %v", pw, ok)
	}
	if _, ok := PlainPassword(HashPassword("x")); ok {
		t.Error("a digest has no plain text")
	}
	if !CheckPassword("plain:sha256:abc", "sha256:abc") {
		t.Error("a plain: password that looks like a digest is checked as text")
	}
}

func TestMalformedDigestsAndEmptyPasswordsAreProblems(t *testing.T) {
	for _, stored := range []string{"", "plain:", "sha256:", "sha256:abcd", "sha256:" + strings.Repeat("z", 64), "sha1:" + strings.Repeat("0", 64)} {
		if PasswordProblem(stored) == "" {
			t.Errorf("%q should be a problem", stored)
		}
		if stored != "" && stored != "plain:" && CheckPassword(stored, stored) {
			t.Errorf("%q must not be accepted as any password", stored)
		}
	}
	if p := PasswordProblem("sha256:abcd"); !strings.Contains(p, "plain:") {
		t.Errorf("the message should say how to keep a password that begins like that: %q", p)
	}
}

func TestSamePasswordComparesAcrossForms(t *testing.T) {
	h := HashPassword("hunter22")
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"hunter22", "hunter22", true}, {"hunter22", "plain:hunter22", true}, {"hunter22", h, true}, {h, "hunter22", true},
		{h, h, true}, {"hunter22", "hunter23", false}, {h, HashPassword("other"), false}, {h, "sha512:" + hexOf(make([]byte, 64)), false},
	} {
		if got := SamePassword(c.a, c.b); got != c.want {
			t.Errorf("SamePassword(%q, %q) = %v", c.a, c.b, got)
		}
	}
}

// A new password keeps the form of the one it replaces; plain text that looks like a digest gets the plain: prefix.
func TestStoredLikeKeepsTheForm(t *testing.T) {
	if got := StoredLike("old-plain", "newpass1"); got != "newpass1" {
		t.Errorf("plain stays plain: %q", got)
	}
	if got := StoredLike(HashPassword("old"), "newpass1"); got != HashPassword("newpass1") {
		t.Errorf("a digest stays a digest: %q", got)
	}
	for _, pw := range []string{"sha256:abc", "plain:abc"} {
		got := StoredLike("old-plain", pw)
		if !CheckPassword(got, pw) || PasswordProblem(got) != "" {
			t.Errorf("%q stored as %q must check against itself", pw, got)
		}
	}
}

// The file may hold digests; the program starts, warns about what is weak and refuses what is malformed.
func TestConfigWithDigestsLoadsAndWarns(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(p, []byte("admin_password: "+HashPassword("adminpw1")+"\noperator_password: operatorpw1\n"), 0o600)
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(c.AdminPassword, "adminpw1") || !CheckPassword(c.OperatorPassword, "operatorpw1") {
		t.Error("both passwords should check")
	}
	for _, w := range c.Warnings() {
		if strings.Contains(w, "shorter") || strings.Contains(w, "same") {
			t.Errorf("unexpected warning %q", w)
		}
	}
	s1 := sha1.Sum([]byte("operatorpw1"))
	os.WriteFile(p, []byte("admin_password: "+HashPassword("short")+"\noperator_password: sha1:"+hexOf(s1[:])+"\n"), 0o600)
	c, err = Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if w := strings.Join(c.Warnings(), "\n"); !strings.Contains(w, "operator_password is a sha1 digest") || strings.Contains(w, "shorter") {
		t.Errorf("warnings: %q (a digest's length says nothing; sha1 is weak)", w)
	}
	// the same password in two forms is still the same password
	os.WriteFile(p, []byte("admin_password: "+HashPassword("samepass1")+"\noperator_password: samepass1\n"), 0o600)
	if c, _ = Load(p); !strings.Contains(strings.Join(c.Warnings(), "\n"), "are the same") {
		t.Errorf("warnings: %v", c.Warnings())
	}
	os.WriteFile(p, []byte("admin_password: sha256:nothex\noperator_password: operatorpw1\n"), 0o600)
	if _, err := Load(p); err == nil || !strings.Contains(err.Error(), "admin_password: starts with \"sha256:\"") {
		t.Errorf("%v", err)
	}
}
