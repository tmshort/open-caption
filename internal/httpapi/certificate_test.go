// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/acme"
)

type fakeCerts struct {
	st   acme.Status
	busy bool
	hit  int
}

func (f *fakeCerts) Status() acme.Status { return f.st }
func (f *fakeCerts) RenewNow() bool {
	if f.busy {
		return false
	}
	f.hit++
	return true
}

func TestCertificateInfoSaysWhereHTTPSGetsItsCertificate(t *testing.T) {
	s, _, c := standbyServer(t)
	if code, r := c.do("GET", "/api/admin/settings/certificate", nil); code != 200 || r["source"] != "none" {
		t.Errorf("HTTPS off: %d %v", code, r)
	}
	s.cfg.HTTPS.Enabled = true
	if _, r := c.do("GET", "/api/admin/settings/certificate", nil); r["source"] != "self-signed" {
		t.Errorf("self-signed: %v", r)
	}
	s.cfg.HTTPS.CertFile = "/x/cert.pem"
	if _, r := c.do("GET", "/api/admin/settings/certificate", nil); r["source"] != "file" {
		t.Errorf("file: %v", r)
	}
}

func TestAutomaticCertificateStateIsShownAndRenewNowIsGuarded(t *testing.T) {
	s, _, c := standbyServer(t)
	s.cfg.HTTPS.Enabled = true
	s.cfg.HTTPS.ACME.Enabled = true
	s.cfg.HTTPS.ACME.Directory, s.cfg.HTTPS.ACME.DNS.Provider = "letsencrypt", "easydns"
	if code, _ := c.do("POST", "/api/admin/settings/certificate/renew", nil); code != 409 {
		t.Errorf("without a manager renewing is refused: %d", code)
	}
	f := &fakeCerts{st: acme.Status{Domain: "captions.example.org", State: "ok", NotAfter: time.Now().Add(80 * 24 * time.Hour), DaysLeft: 80, LastError: ""}}
	s.SetCertificates(f)
	code, r := c.do("GET", "/api/admin/settings/certificate", nil)
	st, _ := r["status"].(map[string]any)
	if code != 200 || r["source"] != "acme" || r["provider"] != "easydns" || st["domain"] != "captions.example.org" || st["state"] != "ok" {
		t.Errorf("%d %v", code, r)
	}
	if _, has := r["credentials"]; has {
		t.Error("credentials must never be sent to the browser")
	}
	if code, _ := c.do("POST", "/api/admin/settings/certificate/renew", nil); code != 200 || f.hit != 1 {
		t.Errorf("renew: %d hit=%d", code, f.hit)
	}
	f.busy = true
	if code, _ := c.do("POST", "/api/admin/settings/certificate/renew", nil); code != 409 {
		t.Errorf("a second renewal while one runs is refused: %d", code)
	}
}

func TestCertificateEndpointsAreForTheAdminOnly(t *testing.T) {
	_, ts, _ := standbyServer(t)
	anon := newClient(t, ts.URL)
	if code, _ := anon.do("GET", "/api/admin/settings/certificate", nil); code != 401 {
		t.Errorf("anonymous: %d", code)
	}
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	if code, _ := op.do("POST", "/api/admin/settings/certificate/renew", nil); code != 401 && code != 403 {
		t.Errorf("the operator password must not renew certificates: %d", code)
	}
}
