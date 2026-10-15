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

	"github.com/tmshort/open-caption/internal/showfile"
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

// Look is the default appearance of the captions. A show file's [settings] override each value.
type Look struct {
	Background       string `yaml:"background"`
	TextColor        string `yaml:"text_color"`
	SpeakerNameColor string `yaml:"speaker_name_color"`
	Font             string `yaml:"font"`
	TextSizePercent  int    `yaml:"text_size_percent"`
	SpeakerNames     string `yaml:"speaker_names"`
}

// Defaults are the values a show uses when its show file does not say otherwise. The layers are the
// program's built-in values, then this section of config.yaml, then the show file's [settings].
type Defaults struct {
	Language        string  `yaml:"language"`
	LoopItemSeconds float64 `yaml:"loop_item_seconds"`
	IdleMessage     string  `yaml:"idle_message"`
	OpenBrowser     bool    `yaml:"open_browser"`
	Look            Look    `yaml:"look"`
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
	Access           Access   `yaml:"access"`

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
			IdleMessage: "Captions will begin shortly", OpenBrowser: true,
			Look: Look{Background: "#000000", TextColor: "#ffffff", SpeakerNameColor: "#ffffff", Font: "system", TextSizePercent: 100, SpeakerNames: "label"}},
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
	if err := cfg.normalizeLook(); err != nil {
		return cfg, err
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
	header := "# open-caption configuration. Passwords are stored in plain text; keep this file private.\n" +
		"# access: limits which computers may open the Setup (admin_allow) and operator (operator_allow) pages.\n" +
		"# An empty list lets anyone on the network try the password. Entries: localhost, private,\n" +
		"# an address (192.168.1.50) or a network (192.168.1.0/24). Example: admin_allow: [localhost]\n"
	tmp := c.Path + ".tmp"
	if err := os.WriteFile(tmp, append([]byte(header), b...), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.Path)
}

// MinPasswordLength is the shortest password the Setup page accepts.
const MinPasswordLength = 6

// SetPasswords changes the passwords in the config file at path, leaving every other setting as it is
// on disk (so command-line overrides are never written into the file). Empty arguments are left alone.
func SetPasswords(path, admin, operator string) error {
	if _, err := os.Stat(path); err != nil {
		return errors.New("config file not found")
	}
	c, err := Load(path)
	if err != nil {
		return err
	}
	if admin != "" {
		c.AdminPassword = admin
	}
	if operator != "" {
		c.OperatorPassword = operator
	}
	return c.Save()
}

// Validate checks required settings.
func (c Config) Validate() error {
	switch {
	case c.AdminPassword == "":
		return errors.New("config: admin_password must not be empty")
	case c.OperatorPassword == "":
		return errors.New("config: operator_password must not be empty")
	case c.accessError() != nil:
		return c.accessError()
	case c.HTTPS.Enabled && (c.HTTPS.CertFile == "") != (c.HTTPS.KeyFile == ""):
		return errors.New("config: https.cert_file and https.key_file must be set together")
	case c.HTTPS.Enabled && c.HTTPS.CertFile == "" && !c.HTTPS.SelfSigned:
		return errors.New("config: https enabled but no cert_file/key_file and self_signed is false")
	}
	return nil
}

// normalizeLook checks the default look and writes each value in its canonical form (for example
// "#FFF" becomes "#ffffff").
func (c *Config) normalizeLook() error {
	l := &c.Defaults.Look
	var err error
	bad := func(key string, e error) error { return fmt.Errorf("config: defaults.look.%s: %w", key, e) }
	if l.Background, err = showfile.NormalizeColor(l.Background); err != nil {
		return bad("background", err)
	}
	if l.TextColor, err = showfile.NormalizeColor(l.TextColor); err != nil {
		return bad("text_color", err)
	}
	if l.SpeakerNameColor, err = showfile.NormalizeColor(l.SpeakerNameColor); err != nil {
		return bad("speaker_name_color", err)
	}
	if l.Font, err = showfile.NormalizeFont(l.Font); err != nil {
		return bad("font", err)
	}
	if l.SpeakerNames, err = showfile.NormalizeSpeakerNames(l.SpeakerNames); err != nil {
		return bad("speaker_names", err)
	}
	if l.TextSizePercent < 50 || l.TextSizePercent > 300 {
		return bad("text_size_percent", fmt.Errorf("%d is not a text size: use 50 to 300", l.TextSizePercent))
	}
	if c.Defaults.LoopItemSeconds <= 0 || c.Defaults.LoopItemSeconds > 3600 {
		return fmt.Errorf("config: defaults.loop_item_seconds: %v is not a time: use a number of seconds from 1 to 3600", c.Defaults.LoopItemSeconds)
	}
	if len(c.Defaults.IdleMessage) > showfile.MaxIdleMessage {
		return fmt.Errorf("config: defaults.idle_message is longer than %d characters", showfile.MaxIdleMessage)
	}
	if n := c.Audience.ScrollbackLines; n < 20 || n > 2000 {
		return fmt.Errorf("config: audience.scrollback_lines: %d is not a line count: use 20 to 2000", n)
	}
	return nil
}

func (c Config) accessError() error {
	if _, err := ParseAllowList(c.Access.AdminAllow); err != nil {
		return fmt.Errorf("config: access.admin_allow: %w", err)
	}
	if _, err := ParseAllowList(c.Access.OperatorAllow); err != nil {
		return fmt.Errorf("config: access.operator_allow: %w", err)
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
