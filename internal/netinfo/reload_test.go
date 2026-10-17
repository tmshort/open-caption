package netinfo

import (
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// pair makes a certificate pair in its own folder and returns the files' contents.
func pair(t *testing.T) (cert, key []byte) {
	t.Helper()
	dir := t.TempDir()
	if _, err := EnsureSelfSigned(dir, "host.test", []net.IP{net.ParseIP("10.0.0.5")}); err != nil {
		t.Fatal(err)
	}
	c, _ := os.ReadFile(filepath.Join(dir, "cert.pem"))
	k, _ := os.ReadFile(filepath.Join(dir, "key.pem"))
	return c, k
}

func serial(t *testing.T, r *Reloader) string {
	t.Helper()
	c, err := r.GetCertificate(&tls.ClientHelloInfo{})
	if err != nil {
		t.Fatal(err)
	}
	return leafOf(c).SerialNumber.String()
}

func write(t *testing.T, path string, b []byte, mod time.Time) {
	t.Helper()
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	os.Chtimes(path, mod, mod)
}

func TestReloaderPicksUpARenewedCertificateWithoutARestart(t *testing.T) {
	dir := t.TempDir()
	cf, kf := filepath.Join(dir, "c.pem"), filepath.Join(dir, "k.pem")
	c1, k1 := pair(t)
	c2, k2 := pair(t)
	t0 := time.Now().Add(-time.Hour)
	write(t, cf, c1, t0)
	write(t, kf, k1, t0)
	var logs []string
	r, err := NewReloader(cf, kf, func(f string, a ...any) { logs = append(logs, f) }, nil)
	if err != nil {
		t.Fatal(err)
	}
	clock := time.Now()
	r.now = func() time.Time { return clock }
	first := serial(t, r)

	write(t, cf, c2, t0.Add(time.Minute))
	write(t, kf, k2, t0.Add(time.Minute))
	if serial(t, r) != first {
		t.Error("the files are only looked at every 30 seconds")
	}
	clock = clock.Add(reloadEvery + time.Second)
	if got := serial(t, r); got == first {
		t.Error("a renewed certificate was not picked up")
	}
}

func TestReloaderKeepsServingWhenTheNewFilesAreBadAndRecoversWhenFixed(t *testing.T) {
	dir := t.TempDir()
	cf, kf := filepath.Join(dir, "c.pem"), filepath.Join(dir, "k.pem")
	c1, k1 := pair(t)
	c2, k2 := pair(t)
	t0 := time.Now().Add(-time.Hour)
	write(t, cf, c1, t0)
	write(t, kf, k1, t0)
	warnings := 0
	r, err := NewReloader(cf, kf, nil, func(f string, a ...any) { warnings++ })
	if err != nil {
		t.Fatal(err)
	}
	clock := time.Now()
	r.now = func() time.Time { return clock }
	first := serial(t, r)

	// the new certificate has arrived but its key has not: they do not match
	write(t, cf, c2, t0.Add(time.Minute))
	for i := 0; i < 3; i++ {
		clock = clock.Add(reloadEvery + time.Second)
		if serial(t, r) != first {
			t.Fatal("a mismatched pair replaced the working certificate")
		}
	}
	if warnings != 1 {
		t.Errorf("the problem should be logged once, not at every check: %d", warnings)
	}
	write(t, kf, k2, t0.Add(2*time.Minute))
	clock = clock.Add(reloadEvery + time.Second)
	if serial(t, r) == first {
		t.Error("the completed pair was not picked up")
	}
}

func TestNewReloaderRefusesAnUnusablePair(t *testing.T) {
	dir := t.TempDir()
	if _, err := NewReloader(filepath.Join(dir, "no.pem"), filepath.Join(dir, "no.key"), nil, nil); err == nil {
		t.Error("missing files should be an error at start-up")
	}
}
