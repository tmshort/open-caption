package acme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// certFor makes a certificate for domain that is valid from notBefore until notAfter.
func certFor(t *testing.T, domain string, notBefore, notAfter time.Time) Bundle {
	t.Helper()
	k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: "test CA"},
		NotBefore: notBefore, NotAfter: notAfter, DNSNames: []string{domain}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &k.PublicKey, k)
	if err != nil {
		t.Fatal(err)
	}
	kb, _ := x509.MarshalECPrivateKey(k)
	return Bundle{Cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), Key: pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb})}
}

type fakeIssuer struct {
	mu    sync.Mutex
	calls int
	next  func(call int) (Bundle, error)
}

func (f *fakeIssuer) Obtain(context.Context) (Bundle, error) {
	f.mu.Lock()
	f.calls++
	n := f.calls
	f.mu.Unlock()
	return f.next(n)
}
func (f *fakeIssuer) count() int { f.mu.Lock(); defer f.mu.Unlock(); return f.calls }

func newMgr(t *testing.T, iss Issuer, now *time.Time) *Manager {
	t.Helper()
	return New(Options{Domain: "captions.example.org", Dir: t.TempDir(), Issuer: iss, RenewBefore: 30 * 24 * time.Hour,
		Now: func() time.Time { return *now }, Fallback: func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return &tls.Certificate{}, nil }})
}

func serial(t *testing.T, m *Manager) string {
	t.Helper()
	c, err := m.GetCertificate(&tls.ClientHelloInfo{})
	if err != nil || c.Leaf == nil {
		return "fallback"
	}
	return c.Leaf.SerialNumber.String()
}

func TestFirstCertificateIsFetchedKeptOnDiskAndServed(t *testing.T) {
	now := time.Now()
	iss := &fakeIssuer{next: func(int) (Bundle, error) {
		return certFor(t, "captions.example.org", now.Add(-time.Hour), now.Add(90*24*time.Hour)), nil
	}}
	m := newMgr(t, iss, &now)
	if serial(t, m) != "fallback" || m.Status().State != "waiting" {
		t.Fatal("until the first certificate arrives the fallback is served")
	}
	m.check(context.Background())
	if serial(t, m) == "fallback" {
		t.Fatal("the new certificate is not served")
	}
	st := m.Status()
	if st.State != "ok" || st.DaysLeft < 88 || st.LastError != "" {
		t.Errorf("%+v", st)
	}
	for _, f := range []string{"cert.pem", "key.pem"} {
		if _, err := os.Stat(filepath.Join(m.o.Dir, f)); err != nil {
			t.Errorf("%s was not kept: %v", f, err)
		}
	}
	if st, _ := os.Stat(filepath.Join(m.o.Dir, "key.pem")); st.Mode().Perm()&0o077 != 0 && os.PathSeparator == '/' {
		t.Errorf("the private key must be readable by this account only: %v", st.Mode())
	}
	// a restart uses what was kept, without asking the CA again
	m2 := New(Options{Domain: "captions.example.org", Dir: m.o.Dir, Issuer: iss, Now: func() time.Time { return now }})
	if m2.cert == nil {
		t.Error("the kept certificate was not loaded")
	}
	m2.check(context.Background())
	if iss.count() != 1 {
		t.Errorf("a valid certificate must not be fetched again: %d calls", iss.count())
	}
}

func TestRenewalStartsInsideTheRenewWindowAndAFailureKeepsTheOldCertificate(t *testing.T) {
	now := time.Now()
	issued := 0
	iss := &fakeIssuer{}
	iss.next = func(call int) (Bundle, error) {
		switch call {
		case 1:
			issued++
			return certFor(t, "captions.example.org", now.Add(-time.Hour), now.Add(60*24*time.Hour)), nil
		case 2:
			return Bundle{}, errors.New("DNS provider said no")
		}
		issued++
		return certFor(t, "captions.example.org", now.Add(-time.Hour), now.Add(90*24*time.Hour)), nil
	}
	m := newMgr(t, iss, &now)
	m.check(context.Background())
	first := serial(t, m)
	now = now.Add(10 * 24 * time.Hour) // 50 days left: outside the 30-day window
	m.check(context.Background())
	if iss.count() != 1 {
		t.Fatal("renewed too early")
	}
	now = now.Add(25 * 24 * time.Hour) // 25 days left: inside it
	if wait := m.check(context.Background()); wait != m.o.RetryAfter {
		t.Errorf("after a failure the next try should come after RetryAfter, not %v", wait)
	}
	if serial(t, m) != first || m.Status().State != "failed" || !strings.Contains(m.Status().LastError, "said no") {
		t.Errorf("a failed renewal must keep serving the old certificate and say why: %+v", m.Status())
	}
	m.check(context.Background())
	if serial(t, m) == first || m.Status().State != "ok" || m.Status().LastError != "" {
		t.Errorf("the next try should succeed and clear the error: %+v", m.Status())
	}
}

func TestACertificateForTheWrongNameOrAnUnusableOneIsRefused(t *testing.T) {
	now := time.Now()
	bad := []Bundle{
		certFor(t, "other.example.org", now.Add(-time.Hour), now.Add(90*24*time.Hour)),
		{Cert: []byte("nonsense"), Key: []byte("nonsense")},
	}
	for i, b := range bad {
		iss := &fakeIssuer{next: func(int) (Bundle, error) { return b, nil }}
		m := newMgr(t, iss, &now)
		m.check(context.Background())
		if serial(t, m) != "fallback" || m.Status().LastError == "" {
			t.Errorf("bundle %d was accepted: %+v", i, m.Status())
		}
	}
}

func TestRenewNowWorksOnceAtATime(t *testing.T) {
	now := time.Now()
	release := make(chan struct{})
	iss := &fakeIssuer{next: func(int) (Bundle, error) {
		<-release
		return certFor(t, "captions.example.org", now.Add(-time.Hour), now.Add(90*24*time.Hour)), nil
	}}
	m := newMgr(t, iss, &now)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)
	deadline := time.Now().Add(3 * time.Second)
	for m.Status().State != "renewing" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if m.RenewNow() {
		t.Error("a second renewal must be refused while one is running")
	}
	close(release)
	for m.Status().State != "ok" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if m.Status().State != "ok" {
		t.Fatal("never became ok")
	}
}

// A crash after the new key replaced the old one but before the new certificate did must not lose the
// certificate (fetching again spends Let's Encrypt's rate limit).
func TestAnInterruptedInstallIsFinishedAtTheNextStart(t *testing.T) {
	now := time.Now()
	dir := t.TempDir()
	old := certFor(t, "captions.example.org", now.Add(-48*time.Hour), now.Add(10*24*time.Hour))
	fresh := certFor(t, "captions.example.org", now.Add(-time.Hour), now.Add(90*24*time.Hour))
	os.WriteFile(filepath.Join(dir, "cert.pem"), old.Cert, 0o644)
	os.WriteFile(filepath.Join(dir, "key.pem"), fresh.Key, 0o600)
	os.WriteFile(filepath.Join(dir, "cert.pem.new"), fresh.Cert, 0o644)
	m := New(Options{Domain: "captions.example.org", Dir: dir, Now: func() time.Time { return now }})
	if m.cert == nil || m.leaf.NotAfter.Before(now.Add(60*24*time.Hour)) {
		t.Fatalf("the finished install was not picked up: %+v", m.leaf)
	}
	if _, err := os.Stat(filepath.Join(dir, "cert.pem.new")); err == nil {
		t.Error("the temporary file was left behind")
	}
}
