//go:build !acme_off && pebble

package acme

// An end-to-end check of the real ACME path (lego, the DNS-01 challenge, our Manager) against Pebble, Let's Encrypt's
// test CA, with Pebble's test DNS server standing in for the venue's DNS provider. Nothing leaves this computer.
//
//	make pebble-test
//
// (it installs the pinned Pebble version, PEBBLE_VERSION in the Makefile, and sets the variables below.)
//
// (PEBBLE_DIR is the folder holding Pebble's test/certs; PEBBLE_BIN, if set, is the folder with the two programs.)

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"

	"github.com/tmshort/open-caption/internal/config"
)

// challTestProvider publishes the TXT record by telling Pebble's test DNS server, as a real provider does through its API.
type challTestProvider struct{ api string }

func (p challTestProvider) post(path string, body any) error {
	b, _ := json.Marshal(body)
	r, err := http.Post(p.api+path, "application/json", bytes.NewReader(b))
	if err == nil {
		r.Body.Close()
	}
	return err
}
func (p challTestProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	return p.post("/set-txt", map[string]string{"host": fqdn, "value": value})
}
func (p challTestProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)
	return p.post("/clear-txt", map[string]string{"host": fqdn})
}

func start(t *testing.T, bin string, env []string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), env...)
	if err := cmd.Start(); err != nil {
		t.Fatalf("%s: %v", bin, err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
}

func TestPebbleIssuesACertificateWithTheDNSChallenge(t *testing.T) {
	pd := os.Getenv("PEBBLE_DIR")
	if pd == "" {
		t.Skip("set PEBBLE_DIR (see the top of this file)")
	}
	bin := func(n string) string {
		if d := os.Getenv("PEBBLE_BIN"); d != "" {
			return filepath.Join(d, n)
		}
		return n
	}
	tmp := t.TempDir()
	cfg, _ := json.Marshal(map[string]any{"pebble": map[string]any{
		"listenAddress": "127.0.0.1:14000", "managementListenAddress": "127.0.0.1:15000",
		"certificate": filepath.Join(pd, "test/certs/localhost/cert.pem"), "privateKey": filepath.Join(pd, "test/certs/localhost/key.pem"),
		"httpPort": 5002, "tlsPort": 5001, "ocspResponderURL": "", "keyAlgorithm": "ecdsa",
		"profiles": map[string]any{"default": map[string]any{"description": "d", "validityPeriod": 7776000}}}})
	cfgFile := filepath.Join(tmp, "pebble.json")
	os.WriteFile(cfgFile, cfg, 0o600)
	start(t, bin("pebble-challtestsrv"), nil, "-http01", "", "-https01", "", "-tlsalpn01", "", "-doh", "", "-dnsserver", "127.0.0.1:8053", "-management", "127.0.0.1:8055")
	start(t, bin("pebble"), []string{"PEBBLE_VA_NOSLEEP=1"}, "-config", cfgFile, "-dnsserver", "127.0.0.1:8053")
	t.Setenv("LEGO_CA_CERTIFICATES", filepath.Join(pd, "test/certs/pebble.minica.pem"))

	hc := &http.Client{Timeout: time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	deadline := time.Now().Add(15 * time.Second)
	for {
		if r, err := hc.Get("https://127.0.0.1:14000/dir"); err == nil {
			r.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Pebble did not start")
		}
		time.Sleep(100 * time.Millisecond)
	}

	registry["pebble"] = func() (challenge.Provider, error) { return challTestProvider{api: "http://127.0.0.1:8055"}, nil }
	t.Cleanup(func() { delete(registry, "pebble") })
	a := config.ACME{Enabled: true, Domain: "captions.example.org", Email: "me@example.org", AgreeToTerms: true,
		Directory: "https://localhost:14000/dir", RenewBeforeDays: 30,
		DNS: config.ACMEDNS{Provider: "pebble", PropagationWait: "1s", Credentials: map[string]string{"OCT_PEBBLE_TOKEN": "not-a-real-secret"}}}
	dir := filepath.Join(tmp, "acme")
	iss, err := NewIssuer(a, dir, t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	m := New(Options{Domain: a.Domain, Dir: dir, Issuer: iss, RenewBefore: 30 * 24 * time.Hour, Infof: t.Logf, Warnf: t.Logf})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	m.check(ctx)
	st := m.Status()
	if st.State != "ok" || st.LastError != "" || st.DaysLeft < 80 {
		t.Fatalf("no certificate: %+v", st)
	}
	c, err := m.GetCertificate(&tls.ClientHelloInfo{ServerName: a.Domain})
	if err != nil || c.Leaf == nil || c.Leaf.VerifyHostname(a.Domain) != nil {
		t.Fatalf("the served certificate is not for the name: %v", err)
	}
	if os.Getenv("OCT_PEBBLE_TOKEN") != "not-a-real-secret" {
		t.Error("the credentials from the settings were not given to the provider")
	}
	// the account is kept: a second certificate needs no new registration
	files, _ := filepath.Glob(filepath.Join(dir, "account-*.json"))
	if len(files) != 1 {
		t.Errorf("account files: %v", files)
	}
	if !m.RenewNow() {
		t.Fatal("RenewNow refused")
	}
	m.check(ctx)
	if m.Status().State != "ok" {
		t.Errorf("a forced renewal failed: %+v", m.Status())
	}
}
