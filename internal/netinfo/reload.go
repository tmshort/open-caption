package netinfo

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"
)

// reloadEvery is how often a connection may trigger a look at the certificate files (a stat, not a read).
const reloadEvery = 30 * time.Second

// Reloader serves a certificate from files and picks up a renewed one without a restart (for example after
// certbot or lego replaced the files). It checks the files' modification time at most every reloadEvery, on a
// new connection. A renewal that cannot be loaded (a half-written pair, a wrong key) keeps the certificate
// already in use and is tried again at the next check, so a bad file never takes HTTPS down.
type Reloader struct {
	certFile, keyFile string
	now               func() time.Time
	infof, warnf      func(format string, args ...any) // either may be nil

	mu       sync.Mutex
	cert     *tls.Certificate
	checked  time.Time
	seen     fileStamp
	failedAt fileStamp // the stamp of the pair that failed to load, so the failure is logged once
}

type fileStamp struct {
	certMod, keyMod   time.Time
	certSize, keySize int64
}

// NewReloader loads the pair once; it fails if the pair cannot be used. infof and warnf may be nil.
func NewReloader(certFile, keyFile string, infof, warnf func(string, ...any)) (*Reloader, error) {
	r := &Reloader{certFile: certFile, keyFile: keyFile, now: time.Now, infof: infof, warnf: warnf}
	stamp, err := r.stamp()
	if err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	r.cert, r.seen, r.checked = &cert, stamp, r.now()
	r.logExpiry("loaded")
	return r, nil
}

func (r *Reloader) stamp() (fileStamp, error) {
	c, err := os.Stat(r.certFile)
	if err != nil {
		return fileStamp{}, err
	}
	k, err := os.Stat(r.keyFile)
	if err != nil {
		return fileStamp{}, err
	}
	return fileStamp{c.ModTime(), k.ModTime(), c.Size(), k.Size()}, nil
}

// GetCertificate is tls.Config.GetCertificate.
func (r *Reloader) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if now := r.now(); now.Sub(r.checked) >= reloadEvery {
		r.checked = now
		r.refreshLocked()
	}
	return r.cert, nil
}

// Expires is when the certificate in use stops being valid (zero if unknown).
func (r *Reloader) Expires() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return leafOf(r.cert).NotAfter
}

func leafOf(c *tls.Certificate) *x509.Certificate {
	if c == nil {
		return &x509.Certificate{}
	}
	if c.Leaf != nil {
		return c.Leaf
	}
	if len(c.Certificate) > 0 {
		if l, err := x509.ParseCertificate(c.Certificate[0]); err == nil {
			return l
		}
	}
	return &x509.Certificate{}
}

func (r *Reloader) refreshLocked() {
	st, err := r.stamp()
	if err != nil {
		r.fail(r.seen, fmt.Errorf("cannot read the certificate files: %w", err))
		return
	}
	if st == r.seen {
		return
	}
	cert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		r.fail(st, err)
		return
	}
	r.cert, r.seen, r.failedAt = &cert, st, fileStamp{}
	r.logExpiry("reloaded")
}

func (r *Reloader) fail(st fileStamp, err error) {
	if r.failedAt == st { // already said so for this pair of files
		return
	}
	r.failedAt = st
	if r.warnf != nil {
		r.warnf("HTTPS: the certificate files changed but cannot be used yet (%v); still serving the previous certificate", err)
	}
}

func (r *Reloader) logExpiry(what string) {
	if r.infof == nil {
		return
	}
	exp := leafOf(r.cert).NotAfter
	if exp.IsZero() {
		r.infof("HTTPS: certificate %s", what)
		return
	}
	days := int(exp.Sub(r.now()).Hours() / 24)
	r.infof("HTTPS: certificate %s; it is valid until %s (%d days)", what, exp.Format("2006-01-02"), days)
}
