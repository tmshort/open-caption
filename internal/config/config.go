// Package config loads, generates and validates config.yaml (see
// docs/02-architecture.md, "Configuration").
package config

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/tmshort/open-caption/internal/logx"
	"github.com/tmshort/open-caption/internal/mdns"
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
	// ACME gets and renews a certificate from a public CA (Let's Encrypt) with the DNS-01 challenge (docs/10).
	ACME ACME `yaml:"acme,omitempty"`
}

// ACME is the automatic-certificate setting. The name must be a public DNS name that the venue owns.
type ACME struct {
	Enabled bool   `yaml:"enabled"`
	Domain  string `yaml:"domain"`
	Email   string `yaml:"email"`
	// AgreeToTerms must be true: it records that the person accepts the certificate authority's terms of service.
	AgreeToTerms bool `yaml:"agree_to_terms"`
	// Directory is "letsencrypt" (the default), "letsencrypt-staging" (for trying it out), or the address of another ACME CA.
	Directory string `yaml:"directory,omitempty"`
	// RenewBeforeDays is how long before it expires the certificate is renewed (default 30).
	RenewBeforeDays int     `yaml:"renew_before_days,omitempty"`
	DNS             ACMEDNS `yaml:"dns"`
}

// ACMEDNS says how the DNS challenge answers: which provider, and its credentials.
type ACMEDNS struct {
	// Provider is a lego DNS provider name (for example "cloudflare", "godaddy", "easydns"); it must be built into the program.
	Provider string `yaml:"provider"`
	// Credentials are the provider's own settings (for example EASYDNS_TOKEN). The environment, including NAME_FILE variables
	// that name a file holding the value, wins over CredentialsFile, which wins over these.
	Credentials map[string]string `yaml:"credentials,omitempty"`
	// CredentialsFile is a file of NAME=value lines with the same settings, read again at every attempt.
	CredentialsFile string `yaml:"credentials_file,omitempty"`
	// PropagationWait, if set (for example "2m"), is a fixed wait for the new DNS record to spread instead of checking for it.
	PropagationWait string `yaml:"propagation_wait,omitempty"`
}

// Built-in values of the ACME setting.
const (
	ACMEDirectoryProduction = "letsencrypt"
	ACMEDirectoryStaging    = "letsencrypt-staging"
	DefaultRenewBeforeDays  = 30
)

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

// MDNS makes the server findable by name on the local network (multicast DNS): http://open-caption.local/.
type MDNS struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"` // one label; ".local" is added
}

// Network chooses which of this computer's network addresses are offered to the audience: listed in Setup,
// printed at startup, put in the HTTPS certificate and announced by name. Entries are an address
// ("192.168.1.50"), a network ("192.168.1.0/24"), or "private" (the usual home and office networks). Include
// empty means every address; an address that matches Exclude is never offered, whatever Include says.
// Prefer settles which adapter is used when two are on the same network (a cable and Wi-Fi at once): the first
// entry that matches wins. An entry is an adapter's name ("Ethernet", "en4") or a kind ("ethernet", "wifi");
// with none, the main connection is used, then a wired adapter, then any other, then Wi-Fi.
// RestrictListening goes further than offering: the server refuses a connection made to an address that is not
// offered (connections to this computer itself, at localhost, are always served).
type Network struct {
	Include           []string `yaml:"include"`
	Exclude           []string `yaml:"exclude"`
	Prefer            []string `yaml:"prefer"`
	RestrictListening bool     `yaml:"restrict_listening"`
}

// Standby is what the audience sees while no show is on air: before the operator starts one, after one goes
// off air, and when no show is loaded. It belongs to the venue, not to a show. (The logo is a file in the data
// folder, set in Setup.)
type Standby struct {
	Message string `yaml:"message"` // optional; empty means the built-in "captions will begin shortly" in the viewer's language
}

// MaxStandbyMessage is the longest standby message, in characters.
const MaxStandbyMessage = 200

// Log controls how much the program writes to its log.
type Log struct {
	Level string `yaml:"level"` // debug | info | warn | error
}

// Config is the contents of config.yaml.
// CurrentVersion is the config_version this program writes and fully understands.
const CurrentVersion = 1

type Config struct {
	// ConfigVersion lets a newer program mark its file. A file with a newer version than this program knows
	// may have settings it does not recognise; those are ignored with a warning instead of stopping it.
	ConfigVersion    int      `yaml:"config_version"`
	AdminPassword    string   `yaml:"admin_password"`
	OperatorPassword string   `yaml:"operator_password"`
	DataDir          string   `yaml:"data_dir"`
	HTTP             Listener `yaml:"http"`
	HTTPS            HTTPS    `yaml:"https"`
	Defaults         Defaults `yaml:"defaults"`
	Audience         Audience `yaml:"audience"`
	Access           Access   `yaml:"access"`
	Log              Log      `yaml:"log"`
	MDNS             MDNS     `yaml:"mdns"`
	Network          Network  `yaml:"network"`
	Standby          Standby  `yaml:"standby"`

	// Path is where the config was loaded from or generated (not serialized).
	Path string `yaml:"-"`
	// Generated is true if Load created the file this run (not serialized).
	Generated bool `yaml:"-"`
	// notes are warnings found while loading (not serialized).
	notes []string `yaml:"-"`
}

// Default returns the built-in defaults (without passwords).
func Default() Config {
	return Config{
		ConfigVersion: CurrentVersion,
		HTTP:          Listener{Addr: ":80", FallbackAddr: ":8080"},
		HTTPS:         HTTPS{Addr: ":443", FallbackAddr: ":8443", SelfSigned: true},
		Defaults: Defaults{Language: "en", LoopItemSeconds: 10,
			IdleMessage: "Captions will begin shortly", OpenBrowser: true,
			Look: Look{Background: "#000000", TextColor: "#ffffff", SpeakerNameColor: "#ffffff", Font: "system", TextSizePercent: 100, SpeakerNames: "label"}},
		Audience: Audience{MaxClients: 500, ScrollbackLines: 200},
		Access:   Access{LocalAutoLogin: true},
		Log:      Log{Level: "info"},
		MDNS:     MDNS{Enabled: true, Name: "open-caption"},
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
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true) // a misspelled key is an error, not silently ignored
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		probs := yamlProblems(err)
		if cfg.ConfigVersion > CurrentVersion { // written by a newer program: unknown settings are expected
			var real []string
			for _, p := range probs {
				if strings.Contains(p, "unknown setting") {
					cfg.notes = append(cfg.notes, p+" (ignored: this file was written by a newer version)")
				} else {
					real = append(real, p)
				}
			}
			probs = real
		}
		if len(probs) > 0 {
			return cfg, &ProblemsError{Path: path, Problems: probs}
		}
	}
	if cfg.ConfigVersion > CurrentVersion {
		cfg.notes = append(cfg.notes, fmt.Sprintf("config_version is %d but this program understands %d: update the program if settings seem to be ignored", cfg.ConfigVersion, CurrentVersion))
	}
	if probs := cfg.normalize(); len(probs) > 0 {
		return cfg, &ProblemsError{Path: path, Problems: probs}
	}
	return cfg, nil
}

// ProblemsError lists everything wrong with a config file, so it can all be fixed in one go.
type ProblemsError struct {
	Path     string
	Problems []string
}

func (e *ProblemsError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s has %d problem", e.Path, len(e.Problems))
	if len(e.Problems) != 1 {
		b.WriteString("s")
	}
	b.WriteString(":")
	for _, p := range e.Problems {
		b.WriteString("\n  - " + p)
	}
	return b.String()
}

// yamlProblems turns a YAML decoding error (syntax error, unknown key, wrong type) into one line each.
func yamlProblems(err error) []string {
	var te *yaml.TypeError
	if errors.As(err, &te) {
		out := make([]string, 0, len(te.Errors))
		sections := map[string]string{"Listener": "http", "HTTPS": "https", "ACME": "https.acme", "ACMEDNS": "https.acme.dns", "Defaults": "defaults", "Look": "defaults.look", "Audience": "audience", "Access": "access", "Log": "log"}
		for _, e := range te.Errors {
			e = strings.TrimPrefix(strings.TrimSpace(e), "yaml: ")
			if i := strings.Index(e, " not found in type config."); i >= 0 { // line N: field X not found in type config.T
				typ := strings.TrimPrefix(e[i:], " not found in type config.")
				e = strings.Replace(e[:i], "field ", "unknown setting ", 1)
				if sec := sections[typ]; sec != "" {
					e += " in " + sec
				}
			} else if i := strings.Index(e, "cannot unmarshal !!"); i >= 0 { // line N: cannot unmarshal !!str `lots` into int
				rest := e[i+len("cannot unmarshal !!"):]
				if j := strings.Index(rest, " "); j >= 0 {
					what, into := rest[j+1:], ""
					if k := strings.LastIndex(what, " into "); k >= 0 {
						what, into = what[:k], what[k+len(" into "):]
					}
					e = e[:i] + what + " is not " + map[string]string{"int": "a whole number", "float64": "a number", "bool": "yes or no", "string": "text"}[into]
					if into == "" || map[string]string{"int": "x", "float64": "x", "bool": "x", "string": "x"}[into] == "" {
						e = e[:i] + what + " is the wrong kind of value here"
					}
				}
			}
			out = append(out, e)
		}
		return out
	}
	return []string{strings.TrimPrefix(err.Error(), "yaml: ")}
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
	kv := map[string]any{}
	if admin != "" {
		kv["admin_password"] = admin
	}
	if operator != "" {
		kv["operator_password"] = operator
	}
	return SetValues(path, kv)
}

// SetValues changes individual settings in the config file, keyed by dotted name ("log.level"). The file is
// edited in place: comments, key order and every other setting stay exactly as the person left them, and
// a key that is missing is added. The file must be valid before and after.
func SetValues(path string, kv map[string]any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return errors.New("config file not found")
	}
	if _, err := Load(path); err != nil {
		return err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return err
	}
	if doc.Kind == 0 { // an empty file
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return errors.New("config file is not a list of settings")
	}
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		node := root
		parts := strings.Split(k, ".")
		for i, part := range parts {
			child := mapValue(node, part)
			last := i == len(parts)-1
			if child == nil {
				kind := yaml.MappingNode
				tag := "!!map"
				if last {
					kind, tag = yaml.ScalarNode, "!!str"
				}
				child = &yaml.Node{Kind: kind, Tag: tag}
				node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: part}, child)
			}
			if last {
				n := yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", HeadComment: child.HeadComment, LineComment: child.LineComment, FootComment: child.FootComment}
				switch v := kv[k].(type) {
				case bool:
					n.Tag, n.Value = "!!bool", strconv.FormatBool(v)
				case int:
					n.Tag, n.Value = "!!int", strconv.Itoa(v)
				case []string: // a list of text values; an empty list is written as []
					n.Kind, n.Tag, n.Value = yaml.SequenceNode, "!!seq", ""
					if len(v) == 0 {
						n.Style = yaml.FlowStyle
					}
					for _, item := range v {
						n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: item})
					}
				default:
					n.Value = fmt.Sprint(v)
				}
				*child = n
			} else {
				if child.Kind != yaml.MappingNode {
					*child = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"} // for example "log:" with nothing under it
				}
				node = child
			}
		}
	}
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return err
	}
	enc.Close()
	mode := os.FileMode(0o600)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out.Bytes(), mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	if _, err := Load(path); err != nil { // never leave a file that will not start
		os.WriteFile(path, b, mode)
		return fmt.Errorf("the change would make %s invalid and was not saved: %w", path, err)
	}
	return nil
}

// mapValue returns the value node for key in a mapping node, or nil.
func mapValue(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// Validate checks the settings and returns every problem found (nil if the configuration is usable).
func (c Config) Validate() error {
	if probs := c.normalize(); len(probs) > 0 {
		return &ProblemsError{Path: c.Path, Problems: probs}
	}
	return nil
}

// Warnings are things that work but are probably not what was meant (for the start-up log).
func (c Config) Warnings() []string {
	w := append([]string(nil), c.notes...)
	if c.AdminPassword != "" && c.AdminPassword == c.OperatorPassword {
		w = append(w, "admin_password and operator_password are the same: anyone with the operator password can open Setup")
	}
	for k, v := range map[string]string{"admin_password": c.AdminPassword, "operator_password": c.OperatorPassword} {
		if v != "" && len(v) < MinPasswordLength {
			w = append(w, fmt.Sprintf("%s is shorter than %d characters", k, MinPasswordLength))
		}
	}
	if len(c.Access.AdminAllow) == 0 {
		w = append(w, "access.admin_allow is empty: any computer on the network may try the Setup password")
	}
	sort.Strings(w)
	return w
}

// normalize checks every setting, writes the look in its canonical form (for example "#FFF" becomes
// "#ffffff"), and returns all the problems found, each naming the setting.
func (c *Config) normalize() []string {
	var probs []string
	add := func(key string, format string, args ...any) {
		probs = append(probs, key+": "+fmt.Sprintf(format, args...))
	}
	if c.AdminPassword == "" {
		add("admin_password", "must not be empty")
	}
	if c.OperatorPassword == "" {
		add("operator_password", "must not be empty")
	}
	if c.DataDir != "" {
		if fi, err := os.Stat(c.DataDir); err == nil && !fi.IsDir() {
			add("data_dir", "%s is a file, not a folder", c.DataDir)
		}
	}
	addr := func(key, v string, required bool) {
		if v == "" {
			if required {
				add(key, "must not be empty (for example \":80\")")
			}
			return
		}
		if err := checkAddr(v); err != nil {
			add(key, "%q is not a listen address: %v (for example \":80\" or \"127.0.0.1:8080\")", v, err)
		}
	}
	addr("http.addr", c.HTTP.Addr, true)
	addr("http.fallback_addr", c.HTTP.FallbackAddr, false)
	if c.HTTPS.Enabled {
		addr("https.addr", c.HTTPS.Addr, true)
		addr("https.fallback_addr", c.HTTPS.FallbackAddr, false)
		switch {
		case (c.HTTPS.CertFile == "") != (c.HTTPS.KeyFile == ""):
			add("https", "cert_file and key_file must be set together")
		case c.HTTPS.CertFile == "" && !c.HTTPS.SelfSigned:
			add("https", "enabled, but there is no cert_file/key_file and self_signed is false")
		case c.HTTPS.CertFile != "" && c.HTTPS.ACME.Enabled:
			add("https.acme", "cannot be used together with cert_file/key_file: choose one way to get the certificate")
		case c.HTTPS.CertFile != "":
			for k, f := range map[string]string{"https.cert_file": c.HTTPS.CertFile, "https.key_file": c.HTTPS.KeyFile} {
				if _, err := os.Stat(f); err != nil {
					add(k, "cannot read %s", f)
				}
			}
		}
	}
	c.checkACME(add)
	if _, err := ParseAllowList(c.Access.AdminAllow); err != nil {
		add("access.admin_allow", "%v", err)
	}
	if _, err := ParseAllowList(c.Access.OperatorAllow); err != nil {
		add("access.operator_allow", "%v", err)
	}
	if _, err := NewAddressFilter(c.Network.Include, nil); err != nil {
		add("network.include", "%v", err)
	}
	if _, err := NewAddressFilter(nil, c.Network.Exclude); err != nil {
		add("network.exclude", "%v", err)
	}
	for i, e := range c.Network.Prefer {
		if c.Network.Prefer[i] = strings.TrimSpace(e); c.Network.Prefer[i] == "" {
			add("network.prefer", "entry %d is empty: use an adapter's name such as Ethernet, or ethernet or wifi", i+1)
		}
	}
	if n, err := mdns.NormalizeName(c.MDNS.Name); err != nil {
		add("mdns.name", "%v", err)
	} else {
		c.MDNS.Name = n
	}
	c.Standby.Message = strings.TrimSpace(c.Standby.Message)
	if len([]rune(c.Standby.Message)) > MaxStandbyMessage || strings.ContainsAny(c.Standby.Message, "\r\n") {
		add("standby.message", "use one line of at most %d characters", MaxStandbyMessage)
	}
	if _, err := logx.Parse(c.Log.Level); err != nil {
		add("log.level", "%v", err)
	}
	if c.Audience.MaxClients < 1 || c.Audience.MaxClients > 100000 {
		add("audience.max_clients", "%d is not a client count: use 1 to 100000", c.Audience.MaxClients)
	}
	if n := c.Audience.ScrollbackLines; n < 20 || n > 2000 {
		add("audience.scrollback_lines", "%d is not a line count: use 20 to 2000", n)
	}
	d := &c.Defaults
	if !langCode.MatchString(d.Language) {
		add("defaults.language", "%q is not a language code (for example en, es or pt-br)", d.Language)
	}
	if d.LoopItemSeconds <= 0 || d.LoopItemSeconds > 3600 {
		add("defaults.loop_item_seconds", "%v is not a time: use a number of seconds from 1 to 3600", d.LoopItemSeconds)
	}
	if len(d.IdleMessage) > showfile.MaxIdleMessage {
		add("defaults.idle_message", "is longer than %d characters", showfile.MaxIdleMessage)
	}
	l := &d.Look
	var err error
	if l.Background, err = showfile.NormalizeColor(l.Background); err != nil {
		add("defaults.look.background", "%v", err)
	}
	if l.TextColor, err = showfile.NormalizeColor(l.TextColor); err != nil {
		add("defaults.look.text_color", "%v", err)
	}
	if l.SpeakerNameColor, err = showfile.NormalizeColor(l.SpeakerNameColor); err != nil {
		add("defaults.look.speaker_name_color", "%v", err)
	}
	if l.Font, err = showfile.NormalizeFont(l.Font); err != nil {
		add("defaults.look.font", "%v", err)
	}
	if l.SpeakerNames, err = showfile.NormalizeSpeakerNames(l.SpeakerNames); err != nil {
		add("defaults.look.speaker_names", "%v", err)
	}
	if l.TextSizePercent < 50 || l.TextSizePercent > 300 {
		add("defaults.look.text_size_percent", "%d is not a text size: use 50 to 300", l.TextSizePercent)
	}
	return probs
}

var langCode = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)

// checkAddr accepts ":80", "127.0.0.1:8080", "[::1]:80" and "host:80".
func checkAddr(a string) error {
	_, port, err := net.SplitHostPort(a)
	if err != nil {
		return err
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("the port must be a number from 1 to 65535")
	}
	return nil
}

// Redacted returns a copy with passwords masked, for display/support bundles.
func (c Config) Redacted() Config {
	c.AdminPassword, c.OperatorPassword = "********", "********"
	if len(c.HTTPS.ACME.DNS.Credentials) > 0 { // the DNS provider's secrets (a new map: the original must not change)
		hidden := make(map[string]string, len(c.HTTPS.ACME.DNS.Credentials))
		for k := range c.HTTPS.ACME.DNS.Credentials {
			hidden[k] = "********"
		}
		c.HTTPS.ACME.DNS.Credentials = hidden
	}
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

// checkACME validates (and fills in the defaults of) the acme section. It only looks at the file system for credentials_file.
func (c *Config) checkACME(add func(key, format string, args ...any)) {
	a := &c.HTTPS.ACME
	if !a.Enabled {
		return
	}
	if !c.HTTPS.Enabled {
		add("https.acme", "needs https.enabled: yes")
	}
	a.Domain = strings.ToLower(strings.TrimSpace(a.Domain))
	switch {
	case a.Domain == "":
		add("https.acme.domain", "must be set: the public name of this computer, such as captions.example.org")
	case strings.HasSuffix(a.Domain, ".local") || strings.HasSuffix(a.Domain, ".lan") || strings.HasSuffix(a.Domain, ".internal") || !strings.Contains(a.Domain, "."):
		add("https.acme.domain", "%q is not a public DNS name: a certificate authority cannot issue certificates for it", a.Domain)
	case net.ParseIP(a.Domain) != nil:
		add("https.acme.domain", "must be a name, not an IP address")
	case !validHostname(a.Domain):
		add("https.acme.domain", "%q is not a valid DNS name (one name, with no wildcard, port or path)", a.Domain)
	}
	a.Email = strings.TrimSpace(a.Email)
	if i := strings.Index(a.Email, "@"); i < 1 || !strings.Contains(a.Email[i:], ".") || strings.ContainsAny(a.Email, " ,;") {
		add("https.acme.email", "must be an email address (the certificate authority sends expiry notices to it)")
	}
	if !a.AgreeToTerms {
		add("https.acme.agree_to_terms", "must be yes: it records that you accept the certificate authority's terms of service")
	}
	switch d := strings.TrimSpace(a.Directory); {
	case d == "":
		a.Directory = ACMEDirectoryProduction
	case d == ACMEDirectoryProduction, d == ACMEDirectoryStaging, strings.HasPrefix(d, "https://"):
		a.Directory = d
	default:
		add("https.acme.directory", "%q must be %q, %q or an https:// address", d, ACMEDirectoryProduction, ACMEDirectoryStaging)
	}
	switch {
	case a.RenewBeforeDays == 0:
		a.RenewBeforeDays = DefaultRenewBeforeDays
	case a.RenewBeforeDays < 1 || a.RenewBeforeDays > 60:
		add("https.acme.renew_before_days", "must be between 1 and 60 days")
	}
	d := &a.DNS
	d.Provider = strings.ToLower(strings.TrimSpace(d.Provider))
	if d.Provider == "" {
		add("https.acme.dns.provider", "must be set (for example cloudflare, godaddy or easydns)")
	}
	for k := range d.Credentials {
		if !validCredentialName(k) {
			add("https.acme.dns.credentials", "%q is not a setting name (capital letters, digits and underscores, such as EASYDNS_TOKEN)", k)
		}
	}
	if d.CredentialsFile != "" {
		if _, err := os.Stat(d.CredentialsFile); err != nil {
			add("https.acme.dns.credentials_file", "cannot read %s", d.CredentialsFile)
		}
	}
	if d.PropagationWait != "" {
		if w, err := time.ParseDuration(d.PropagationWait); err != nil || w <= 0 || w > 30*time.Minute {
			add("https.acme.dns.propagation_wait", "%q must be a time such as 90s or 2m, at most 30m", d.PropagationWait)
		}
	}
}

func validCredentialName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return true
}

func validHostname(s string) bool {
	if len(s) > 253 {
		return false
	}
	for _, label := range strings.Split(s, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
				return false
			}
		}
	}
	return true
}
