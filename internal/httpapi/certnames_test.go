// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTestCert(t *testing.T, names ...string) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}, DNSNames: names,
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "cert.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// The name in the HTTPS certificate is offered as an address of its own (a wildcard cannot be typed, so it is not).
func TestTheCertificateNameIsOfferedAsAnAddress(t *testing.T) {
	s, _, c := adminServer(t)
	s.SetListeners(Listeners{HTTPAddr: "[::]:80", HTTPSAddr: "[::]:8443"})
	certURLs := func() []string {
		_, r := c.do("GET", "/api/admin/network", nil)
		var out []string
		for _, u := range r["urls"].([]any) {
			if m := u.(map[string]any); m["cert"] == true {
				out = append(out, m["url"].(string))
			}
		}
		return out
	}
	if got := certURLs(); len(got) != 0 {
		t.Fatalf("no HTTPS: %v", got)
	}
	s.cfg.HTTPS.Enabled = true
	if got := certURLs(); len(got) != 0 {
		t.Fatalf("self-signed has no name: %v", got)
	}
	s.cfg.HTTPS.CertFile = writeTestCert(t, "Captions.Example.org", "*.example.org")
	if got := certURLs(); len(got) != 1 || got[0] != "https://captions.example.org:8443/" {
		t.Fatalf("certificate file: %v", got)
	}
	s.cfg.HTTPS.CertFile = ""
	s.cfg.HTTPS.ACME.Enabled, s.cfg.HTTPS.ACME.Domain = true, "show.example.org"
	if got := certURLs(); len(got) != 1 || got[0] != "https://show.example.org:8443/" {
		t.Fatalf("acme: %v", got)
	}
	// A QR code can be made for it, and an automatic certificate is not called self-signed.
	if code, _ := c.do("GET", "/api/admin/network/qr.png?url="+url.QueryEscape("https://show.example.org:8443/"), nil); code != 200 {
		t.Errorf("QR code for the certificate name: %d", code)
	}
	if code, _ := c.do("GET", "/api/admin/network/qr.png?url="+url.QueryEscape("https://other.example.org/"), nil); code != 400 {
		t.Errorf("QR code for a name that is not this server's: %d", code)
	}
	if _, r := c.do("GET", "/api/admin/network", nil); strings.Contains(toJSON(r["warnings"]), "self-signed") {
		t.Errorf("an automatic certificate was called self-signed: %v", r["warnings"])
	}
	s.cfg.HTTPS.ACME.Enabled = false
	if _, r := c.do("GET", "/api/admin/network", nil); !strings.Contains(toJSON(r["warnings"]), "self-signed") {
		t.Errorf("a self-signed certificate was not reported: %v", r["warnings"])
	}
}
