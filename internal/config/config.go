// Package config loads, generates and validates config.yaml (see
// docs/02-architecture.md, "Configuration").
package config

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Listener is one HTTP(S) listen address with a fallback.
type Listener struct {
	Addr         string `yaml:"addr"`
	FallbackAddr string `yaml:"fallback_addr"`
}

// HTTPS adds TLS settings to a listener.
type HTTPS struct {
	Enabled      bool   `yaml:"enabled"`
	Addr         string `yaml:"addr"`
	FallbackAddr string `yaml:"fallback_addr"`
	CertFile     string `yaml:"cert_file"`
	KeyFile      string `yaml:"key_file"`
	SelfSigned   bool   `yaml:"self_signed"`
}

// Defaults are show-independent defaults.
type Defaults struct {
	Language        string  `yaml:"language"`
	LoopItemSeconds float64 `yaml:"loop_item_seconds"`
	IdleMessage     string  `yaml:"idle_message"`
	OpenBrowser     bool    `yaml:"open_browser"`
}

// Audience limits.
type Audience struct {
	MaxClients      int `yaml:"max_clients"`
	ScrollbackLines int `yaml:"scrollback_lines"`
}

// Config is the contents of config.yaml.
type Config struct {
	AdminPassword    string   `yaml:"admin_password"`
	OperatorPassword string   `yaml:"operator_password"`
	DataDir          string   `yaml:"data_dir"`
	HTTP             Listener `yaml:"http"`
	HTTPS            HTTPS    `yaml:"https"`
	Defaults         Defaults `yaml:"defaults"`
	Audience         Audience `yaml:"audience"`

	// Path is where the config was loaded from or generated (not serialized).
	Path string `yaml:"-"`
	// Generated is true if Load created the file this run (not serialized).
	Generated bool `yaml:"-"`
}

// Default returns the built-in defaults (without passwords).
func Default() Config {
	return Config{
		HTTP:  Listener{Addr: ":80", FallbackAddr: ":8080"},
		HTTPS: HTTPS{Addr: ":443", FallbackAddr: ":8443", SelfSigned: true},
		Defaults: Defaults{Language: "en", LoopItemSeconds: 10,
			IdleMessage: "Captions will begin shortly", OpenBrowser: true},
		Audience: Audience{MaxClients: 500, ScrollbackLines: 200},
	}
}

// DefaultDataDir is <user config dir>/open-caption.
func DefaultDataDir() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "open-caption"), nil
}

// Load reads path, or creates it (mode 0600) with random passwords if it
// doesn't exist. Missing keys keep their defaults.
func Load(path string) (Config, error) {
	cfg := Default()
	cfg.Path = path
	b, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		cfg.AdminPassword, cfg.OperatorPassword = RandomPassword(), RandomPassword()
		if err := cfg.Save(); err != nil {
			return cfg, err
		}
		cfg.Generated = true
		return cfg, nil
	case err != nil:
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, cfg.Validate()
}

// Save writes the config atomically with mode 0600.
func (c Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	header := "# open-caption configuration. Passwords are stored in plain text; keep this file private.\n"
	tmp := c.Path + ".tmp"
	if err := os.WriteFile(tmp, append([]byte(header), b...), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.Path)
}

// Validate checks required settings.
func (c Config) Validate() error {
	switch {
	case c.AdminPassword == "":
		return errors.New("config: admin_password must not be empty")
	case c.OperatorPassword == "":
		return errors.New("config: operator_password must not be empty")
	case c.HTTPS.Enabled && (c.HTTPS.CertFile == "") != (c.HTTPS.KeyFile == ""):
		return errors.New("config: https.cert_file and https.key_file must be set together")
	case c.HTTPS.Enabled && c.HTTPS.CertFile == "" && !c.HTTPS.SelfSigned:
		return errors.New("config: https enabled but no cert_file/key_file and self_signed is false")
	}
	return nil
}

// Redacted returns a copy with passwords masked, for display/support bundles.
func (c Config) Redacted() Config {
	c.AdminPassword, c.OperatorPassword = "********", "********"
	return c
}

const pwAlphabet = "abcdefghjkmnpqrstuvwxyz23456789" // no look-alike characters

// RandomPassword returns a 12-character password that is easy to read aloud.
func RandomPassword() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = pwAlphabet[int(b[i])%len(pwAlphabet)]
	}
	return string(b)
}
