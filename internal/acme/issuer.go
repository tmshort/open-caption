//go:build !acme_off

package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"github.com/tmshort/open-caption/internal/config"
)

// legoIssuer is the real Issuer: the lego library does the protocol; its account (key and registration) is kept in dir.
type legoIssuer struct {
	cfg   config.ACME
	dir   string
	warnf func(string, ...any)
}

// NewIssuer returns the lego-based Issuer for the settings. It fails at once if the DNS provider is not built in.
func NewIssuer(cfg config.ACME, dir string, warnf func(string, ...any)) (Issuer, error) {
	if !HasProvider(cfg.DNS.Provider) {
		_, err := newProvider(cfg.DNS.Provider)
		return nil, err
	}
	return &legoIssuer{cfg: cfg, dir: dir, warnf: warnf}, nil
}

// account is the CA account: lego's registration.User.
type account struct {
	Email string                 `json:"email"`
	Reg   *registration.Resource `json:"registration"`
	key   crypto.PrivateKey
}

func (a *account) GetEmail() string                        { return a.Email }
func (a *account) GetRegistration() *registration.Resource { return a.Reg }
func (a *account) GetPrivateKey() crypto.PrivateKey        { return a.key }

func (l *legoIssuer) loadAccount(directory string) (*account, error) {
	if err := os.MkdirAll(l.dir, 0o700); err != nil {
		return nil, err
	}
	// one account per CA (the staging and the real one are different accounts)
	tag := "account-" + fmt.Sprintf("%x", sum(directory))
	keyFile, metaFile := filepath.Join(l.dir, tag+".key"), filepath.Join(l.dir, tag+".json")
	a := &account{Email: l.cfg.Email}
	if kb, err := os.ReadFile(keyFile); err == nil {
		if blk, _ := pem.Decode(kb); blk != nil {
			if k, err := x509.ParseECPrivateKey(blk.Bytes); err == nil {
				a.key = k
			}
		}
		if mb, err := os.ReadFile(metaFile); err == nil {
			var saved account
			if json.Unmarshal(mb, &saved) == nil && saved.Email == l.cfg.Email {
				a.Reg = saved.Reg
			}
		}
	}
	if a.key == nil {
		k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		der, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		if err := writeAtomic(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), 0o600); err != nil {
			return nil, err
		}
		a.key, a.Reg = k, nil
	}
	return a, nil
}

func (l *legoIssuer) saveAccount(directory string, a *account) error {
	b, err := json.MarshalIndent(a, "", " ")
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(l.dir, "account-"+fmt.Sprintf("%x", sum(directory))+".json"), b, 0o600)
}

func directoryURL(d string) string {
	switch d {
	case config.ACMEDirectoryProduction:
		return lego.LEDirectoryProduction
	case config.ACMEDirectoryStaging:
		return lego.LEDirectoryStaging
	}
	return d
}

// Obtain asks the CA for a certificate for the domain and answers its DNS challenge through the provider.
func (l *legoIssuer) Obtain(ctx context.Context) (Bundle, error) {
	if _, err := (Credentials{Inline: l.cfg.DNS.Credentials, File: l.cfg.DNS.CredentialsFile, Warnf: l.warnf}).Apply(); err != nil {
		return Bundle{}, err
	}
	prov, err := newProvider(l.cfg.DNS.Provider)
	if err != nil {
		return Bundle{}, fmt.Errorf("the DNS provider %s is not set up (check its credentials): %w", l.cfg.DNS.Provider, err)
	}
	url := directoryURL(l.cfg.Directory)
	acct, err := l.loadAccount(url)
	if err != nil {
		return Bundle{}, err
	}
	lc := lego.NewConfig(acct)
	lc.CADirURL = url
	lc.Certificate.KeyType = certcrypto.EC256
	client, err := lego.NewClient(lc)
	if err != nil {
		return Bundle{}, fmt.Errorf("cannot reach the certificate authority: %w", err)
	}
	var opts []dns01.ChallengeOption
	if w := l.cfg.DNS.PropagationWait; w != "" {
		if d, err := time.ParseDuration(w); err == nil {
			opts = append(opts, dns01.PropagationWait(d, true))
		}
	}
	if err := client.Challenge.SetDNS01Provider(prov, opts...); err != nil {
		return Bundle{}, err
	}
	if acct.Reg == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: l.cfg.AgreeToTerms})
		if err != nil {
			return Bundle{}, fmt.Errorf("cannot register with the certificate authority: %w", err)
		}
		acct.Reg = reg
		if err := l.saveAccount(url, acct); err != nil {
			return Bundle{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return Bundle{}, err
	}
	res, err := client.Certificate.Obtain(certificate.ObtainRequest{Domains: []string{l.cfg.Domain}, Bundle: true})
	if err != nil {
		return Bundle{}, err
	}
	if len(res.Certificate) == 0 || len(res.PrivateKey) == 0 {
		return Bundle{}, errors.New("the certificate authority returned no certificate")
	}
	return Bundle{Cert: res.Certificate, Key: res.PrivateKey}, nil
}
