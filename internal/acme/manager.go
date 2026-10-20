// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package acme gets and renews the server's HTTPS certificate from a public certificate authority (Let's Encrypt)
// with the DNS-01 challenge: the program publishes a TXT record in the venue's public DNS through its provider's API,
// so it never has to be reachable from the Internet. See docs/10-acme-lets-encrypt.md.
//
// Manager (this file) has no knowledge of the ACME protocol: an Issuer does that. The real Issuer uses the lego
// library and is compiled in only with the DNS providers the builder chose (build tags, see the Makefile).
package acme

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Bundle is a certificate chain and its private key, both PEM.
type Bundle struct{ Cert, Key []byte }

// Issuer gets a new certificate (it may take minutes: it waits for DNS to spread).
type Issuer interface {
	Obtain(ctx context.Context) (Bundle, error)
}

// Options configure a Manager.
type Options struct {
	Domain      string
	Dir         string // where the certificate and key are kept (<data-dir>/tls/acme)
	RenewBefore time.Duration
	Issuer      Issuer
	// Fallback serves connections until the first certificate arrives (and if there is none); typically a self-signed one.
	Fallback     func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	CheckEvery   time.Duration // how often to look at the expiry (default 12 hours)
	RetryAfter   time.Duration // after a failure (default 1 hour: CAs limit failed attempts)
	Now          func() time.Time
	Infof, Warnf func(string, ...any)
}

// Status is what Setup shows.
type Status struct {
	Domain      string    `json:"domain"`
	State       string    `json:"state"` // waiting | ok | renewing | failed
	NotAfter    time.Time `json:"notAfter,omitzero"`
	DaysLeft    int       `json:"daysLeft"`
	Issuer      string    `json:"issuer,omitempty"`
	LastAttempt time.Time `json:"lastAttempt,omitzero"`
	LastError   string    `json:"lastError,omitempty"`
	NextCheck   time.Time `json:"nextCheck,omitzero"`
}

// Manager keeps one certificate current.
type Manager struct {
	o Options

	mu       sync.Mutex
	cert     *tls.Certificate
	leaf     *x509.Certificate
	renewing bool
	attempt  time.Time
	lastErr  string
	next     time.Time
	kick     chan struct{}
	forced   atomic.Bool // RenewNow was asked for
}

// New loads the certificate kept from an earlier run, if any (an unusable one is ignored: a new one will be fetched).
func New(o Options) *Manager {
	if o.CheckEvery <= 0 {
		o.CheckEvery = 12 * time.Hour
	}
	if o.RetryAfter <= 0 {
		o.RetryAfter = time.Hour
	}
	if o.RenewBefore <= 0 {
		o.RenewBefore = 30 * 24 * time.Hour
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	m := &Manager{o: o, kick: make(chan struct{}, 1)}
	if _, err := os.Stat(m.certPath() + ".new"); err == nil { // an interrupted install: the key was replaced, the certificate not yet
		if _, _, err := loadPair(m.certPath()+".new", m.keyPath()); err == nil {
			os.Rename(m.certPath()+".new", m.certPath())
		}
	}
	if cert, leaf, err := loadPair(m.certPath(), m.keyPath()); err == nil && leaf.VerifyHostname(o.Domain) == nil && o.Now().Before(leaf.NotAfter) {
		m.cert, m.leaf = cert, leaf
		m.infof("HTTPS: using the saved certificate for %s, valid until %s", o.Domain, leaf.NotAfter.Format("2006-01-02"))
	}
	return m
}

func (m *Manager) certPath() string { return filepath.Join(m.o.Dir, "cert.pem") }
func (m *Manager) keyPath() string  { return filepath.Join(m.o.Dir, "key.pem") }

func (m *Manager) infof(f string, a ...any) {
	if m.o.Infof != nil {
		m.o.Infof(f, a...)
	}
}
func (m *Manager) warnf(f string, a ...any) {
	if m.o.Warnf != nil {
		m.o.Warnf(f, a...)
	}
}

func loadPair(certFile, keyFile string) (*tls.Certificate, *x509.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, nil, err
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, nil, err
	}
	cert.Leaf = leaf
	return &cert, leaf, nil
}

// GetCertificate is tls.Config.GetCertificate: the ACME certificate, or the fallback until there is one.
func (m *Manager) GetCertificate(h *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.mu.Lock()
	c := m.cert
	m.mu.Unlock()
	if c != nil {
		return c, nil
	}
	if m.o.Fallback != nil {
		return m.o.Fallback(h)
	}
	return nil, errors.New("no certificate yet")
}

// Run keeps the certificate current until ctx ends: it checks at once, then every CheckEvery (sooner after a failure).
func (m *Manager) Run(ctx context.Context) {
	for {
		wait := m.check(ctx)
		m.mu.Lock()
		m.next = m.o.Now().Add(wait)
		m.mu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-m.kick:
		case <-time.After(wait):
		}
	}
}

// RenewNow asks Run to get a new certificate now. It returns false if one is already being fetched.
func (m *Manager) RenewNow() bool {
	m.mu.Lock()
	busy := m.renewing
	m.mu.Unlock()
	if busy {
		return false
	}
	m.forced.Store(true)
	select {
	case m.kick <- struct{}{}:
	default:
	}
	return true
}

// due says whether the certificate has to be fetched now.
func (m *Manager) due() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.leaf == nil || m.o.Now().Add(m.o.RenewBefore).After(m.leaf.NotAfter)
}

// check renews if due and returns how long to wait before looking again.
func (m *Manager) check(ctx context.Context) time.Duration {
	if !m.forced.Swap(false) && !m.due() {
		return m.jitter(m.o.CheckEvery)
	}
	if err := m.renew(ctx); err != nil {
		return m.o.RetryAfter
	}
	return m.jitter(m.o.CheckEvery)
}

// jitter spreads the checks of many installations a little.
func (m *Manager) jitter(d time.Duration) time.Duration {
	return d - time.Duration(rand.Int63n(int64(d/10)+1))
}

func (m *Manager) renew(ctx context.Context) error {
	m.mu.Lock()
	m.renewing, m.attempt = true, m.o.Now()
	m.mu.Unlock()
	m.infof("HTTPS: getting a certificate for %s (this can take a few minutes)", m.o.Domain)
	b, err := m.o.Issuer.Obtain(ctx)
	if err == nil {
		err = m.install(b)
	}
	m.mu.Lock()
	m.renewing = false
	if err != nil {
		m.lastErr = err.Error()
	} else {
		m.lastErr = ""
	}
	m.mu.Unlock()
	if err != nil {
		m.warnf("HTTPS: could not get a certificate for %s: %v (the previous certificate stays in use; trying again later)", m.o.Domain, err)
		return err
	}
	m.infof("HTTPS: got a certificate for %s, valid until %s", m.o.Domain, m.leaf.NotAfter.Format("2006-01-02"))
	return nil
}

// install checks the bundle, keeps it on disk and starts serving it.
func (m *Manager) install(b Bundle) error {
	if err := os.MkdirAll(m.o.Dir, 0o700); err != nil {
		return err
	}
	cert, err := tls.X509KeyPair(b.Cert, b.Key)
	if err != nil {
		return fmt.Errorf("the certificate authority returned an unusable certificate: %w", err)
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}
	if err := leaf.VerifyHostname(m.o.Domain); err != nil {
		return fmt.Errorf("the certificate is not for %s: %w", m.o.Domain, err)
	}
	cert.Leaf = leaf
	// Both files are written completely under temporary names first, then moved into place (key, then
	// certificate). A crash in between leaves the finished certificate beside the new key; New moves it in.
	if err := os.WriteFile(m.keyPath()+".new", b.Key, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(m.certPath()+".new", b.Cert, 0o644); err != nil {
		return err
	}
	if err := os.Rename(m.keyPath()+".new", m.keyPath()); err != nil {
		return err
	}
	if err := os.Rename(m.certPath()+".new", m.certPath()); err != nil {
		return err
	}
	m.mu.Lock()
	m.cert, m.leaf = &cert, leaf
	m.mu.Unlock()
	return nil
}

func writeAtomic(path string, b []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Status reports the state of the certificate.
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Status{Domain: m.o.Domain, State: "waiting", LastAttempt: m.attempt, LastError: m.lastErr, NextCheck: m.next}
	if m.leaf != nil {
		s.State, s.NotAfter, s.Issuer = "ok", m.leaf.NotAfter, m.leaf.Issuer.CommonName
		s.DaysLeft = int(m.leaf.NotAfter.Sub(m.o.Now()).Hours() / 24)
	}
	switch {
	case m.renewing:
		s.State = "renewing"
	case m.lastErr != "" && m.leaf == nil:
		s.State = "failed"
	case m.lastErr != "":
		s.State = "failed" // still serving the previous certificate; the error says why it was not renewed
	}
	return s
}
