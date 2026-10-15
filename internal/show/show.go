// Package show stores shows on disk: metadata + theme, one script per
// language, and uploaded images (docs/02-architecture.md, "Persistence layout").
package show

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
)

// Limits.
const (
	MaxAssetBytes = 10 << 20
	LargeAsset    = 1 << 20 // warn above this (images are not rescaled)
	MaxScript     = 2 << 20
)

// Errors.
var (
	ErrNotFound   = errors.New("show not found")
	ErrBadID      = errors.New("invalid show id")
	ErrBadAsset   = errors.New("not a supported image")
	ErrTooLarge   = errors.New("file too large")
	ErrBadName    = errors.New("invalid name")
	ErrBadTheme   = errors.New("invalid theme")
	idRe          = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,60}$`)
	langRe        = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)
	colorRe       = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	allowedExt    = map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".webp": "image/webp"}
	allowedFonts  = map[string]bool{"system": true, "serif": true, "mono": true, "dyslexic": true}
	allowedSpeak  = map[string]bool{"label": true, "color": true, "hidden": true}
	unsafeNameRun = regexp.MustCompile(`[^\p{L}\p{N} ._()\-]+`)
)

// Theme is the per-show look. Zero values are filled by Normalize.
type Theme struct {
	BG          string `yaml:"bg" json:"bg"`
	FG          string `yaml:"fg" json:"fg"`
	Accent      string `yaml:"accent" json:"accent"`
	Font        string `yaml:"font" json:"font"`
	ScalePct    int    `yaml:"scale_pct" json:"scalePct"`
	Speakers    string `yaml:"speakers" json:"speakers"`
	IdleMessage string `yaml:"idle_message" json:"idleMessage"`
}

// Normalize fills defaults and validates.
func (t Theme) Normalize(defIdle string) (Theme, error) {
	def := Theme{BG: "#000000", FG: "#ffffff", Accent: "#ffffff", Font: "system", ScalePct: 100, Speakers: "label", IdleMessage: defIdle}
	if t.BG == "" {
		t.BG = def.BG
	}
	if t.FG == "" {
		t.FG = def.FG
	}
	if t.Accent == "" {
		t.Accent = def.Accent
	}
	if t.Font == "" {
		t.Font = def.Font
	}
	if t.ScalePct == 0 {
		t.ScalePct = def.ScalePct
	}
	if t.Speakers == "" {
		t.Speakers = def.Speakers
	}
	if strings.TrimSpace(t.IdleMessage) == "" {
		t.IdleMessage = defIdle
	}
	switch {
	case !colorRe.MatchString(t.BG), !colorRe.MatchString(t.FG), !colorRe.MatchString(t.Accent):
		return t, fmt.Errorf("%w: colors must look like #rrggbb", ErrBadTheme)
	case !allowedFonts[t.Font]:
		return t, fmt.Errorf("%w: unknown font", ErrBadTheme)
	case t.ScalePct < 50 || t.ScalePct > 300:
		return t, fmt.Errorf("%w: text size must be 50–300%%", ErrBadTheme)
	case !allowedSpeak[t.Speakers]:
		return t, fmt.Errorf("%w: unknown speaker style", ErrBadTheme)
	case len(t.IdleMessage) > 200:
		return t, fmt.Errorf("%w: idle message too long", ErrBadTheme)
	}
	return t, nil
}

// Meta is show.yaml.
type Meta struct {
	ID       string    `yaml:"-" json:"id"`
	Name     string    `yaml:"name" json:"name"`
	Created  time.Time `yaml:"created" json:"created"`
	Updated  time.Time `yaml:"updated" json:"updated"`
	Language string    `yaml:"language" json:"language"`
	Theme    Theme     `yaml:"theme" json:"theme"`
}

// Asset describes an uploaded image.
type Asset struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Large bool   `json:"large"`
}

// Store manages <dir>/<id>/.
type Store struct {
	dir         string
	defLanguage string
	defIdle     string
}

// NewStore creates the shows directory if needed.
func NewStore(dir, defaultLanguage, defaultIdle string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir, defLanguage: defaultLanguage, defIdle: defaultIdle}, nil
}

// ValidID reports whether id is a well-formed show id.
func ValidID(id string) bool { return idRe.MatchString(id) }

func (s *Store) path(id string, elem ...string) (string, error) {
	if !ValidID(id) {
		return "", ErrBadID
	}
	return filepath.Join(append([]string{s.dir, id}, elem...)...), nil
}

// AssetsDir returns the directory holding a show's images.
func (s *Store) AssetsDir(id string) string {
	p, err := s.path(id, "assets")
	if err != nil {
		return ""
	}
	return p
}

func slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	if out == "" {
		out = "show"
	}
	return out
}

// Create makes a new show with an empty script. initial, if non-empty, is the script text.
func (s *Store) Create(name, initial string) (Meta, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 {
		return Meta{}, fmt.Errorf("%w: give the show a name (up to 100 characters)", ErrBadName)
	}
	suffix := make([]byte, 2)
	rand.Read(suffix)
	id := slug(name) + "-" + hex.EncodeToString(suffix)
	th, _ := Theme{}.Normalize(s.defIdle)
	m := Meta{ID: id, Name: name, Created: time.Now().UTC(), Updated: time.Now().UTC(), Language: s.defLanguage, Theme: th}
	dir, _ := s.path(id)
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		return Meta{}, err
	}
	if err := s.writeMeta(m); err != nil {
		return Meta{}, err
	}
	if err := s.WriteScript(id, "", initial); err != nil {
		return Meta{}, err
	}
	return m, nil
}

func (s *Store) writeMeta(m Meta) error {
	p, err := s.path(m.ID, "show.yaml")
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return atomicWrite(p, b, 0o644)
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Get loads a show's metadata.
func (s *Store) Get(id string) (Meta, error) {
	p, err := s.path(id, "show.yaml")
	if err != nil {
		return Meta{}, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Meta{}, ErrNotFound
	} else if err != nil {
		return Meta{}, err
	}
	var m Meta
	if err := yaml.Unmarshal(b, &m); err != nil {
		return Meta{}, fmt.Errorf("%s: %w", p, err)
	}
	m.ID = id
	if m.Language == "" {
		m.Language = s.defLanguage
	}
	if th, err := m.Theme.Normalize(s.defIdle); err == nil {
		m.Theme = th
	}
	return m, nil
}

// List returns all shows, most recently updated first.
func (s *Store) List() ([]Meta, error) {
	ents, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	out := []Meta{}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		if m, err := s.Get(e.Name()); err == nil {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Updated.After(out[j].Updated) })
	return out, nil
}

// Update changes the name and/or theme. Nil fields are left alone.
func (s *Store) Update(id string, name *string, theme *Theme) (Meta, error) {
	m, err := s.Get(id)
	if err != nil {
		return m, err
	}
	if name != nil {
		n := strings.TrimSpace(*name)
		if n == "" || len(n) > 100 {
			return m, fmt.Errorf("%w: give the show a name (up to 100 characters)", ErrBadName)
		}
		m.Name = n
	}
	if theme != nil {
		th, err := theme.Normalize(s.defIdle)
		if err != nil {
			return m, err
		}
		m.Theme = th
	}
	m.Updated = time.Now().UTC()
	return m, s.writeMeta(m)
}

// Delete removes a show and everything in it.
func (s *Store) Delete(id string) error {
	p, err := s.path(id)
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); err != nil {
		return ErrNotFound
	}
	return os.RemoveAll(p)
}

func (s *Store) scriptName(lang string) (string, error) {
	if lang == "" {
		lang = s.defLanguage
	}
	if !langRe.MatchString(lang) {
		return "", fmt.Errorf("invalid language %q", lang)
	}
	return "script." + lang + ".md", nil
}

// ReadScript returns the script text for a language ("" = default).
func (s *Store) ReadScript(id, lang string) (string, error) {
	n, err := s.scriptName(lang)
	if err != nil {
		return "", err
	}
	p, err := s.path(id, n)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		if _, gerr := s.Get(id); gerr != nil {
			return "", gerr
		}
		return "", nil
	}
	return string(b), err
}

// WriteScript saves the script verbatim and bumps the show's updated time.
func (s *Store) WriteScript(id, lang, text string) error {
	if len(text) > MaxScript {
		return ErrTooLarge
	}
	n, err := s.scriptName(lang)
	if err != nil {
		return err
	}
	p, err := s.path(id, n)
	if err != nil {
		return err
	}
	if err := atomicWrite(p, []byte(text), 0o644); err != nil {
		return err
	}
	if m, err := s.Get(id); err == nil {
		m.Updated = time.Now().UTC()
		s.writeMeta(m)
	}
	return nil
}

// ListAssets returns uploaded images, sorted by name.
func (s *Store) ListAssets(id string) ([]Asset, error) {
	dir, err := s.path(id, "assets")
	if err != nil {
		return nil, err
	}
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Asset{}, nil
	} else if err != nil {
		return nil, err
	}
	out := []Asset{}
	for _, e := range ents {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if _, ok := allowedExt[strings.ToLower(filepath.Ext(e.Name()))]; !ok {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Asset{Name: e.Name(), Size: fi.Size(), Large: fi.Size() > LargeAsset})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

// AssetNames returns just the file names.
func (s *Store) AssetNames(id string) []string {
	as, _ := s.ListAssets(id)
	names := make([]string, 0, len(as))
	for _, a := range as {
		names = append(names, a.Name)
	}
	return names
}

// CleanAssetName turns an uploaded file name into a safe stored name.
func CleanAssetName(name string) (string, error) {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimSpace(unsafeNameRun.ReplaceAllString(name, "_"))
	name = strings.TrimLeft(name, ".")
	if name == "" || len(name) > 100 {
		return "", ErrBadName
	}
	if _, ok := allowedExt[strings.ToLower(filepath.Ext(name))]; !ok {
		return "", fmt.Errorf("%w: use PNG, JPG, GIF or WebP", ErrBadAsset)
	}
	return name, nil
}

// SaveAsset stores an image (replacing one of the same name) after checking its
// extension, size and actual content type. It returns the stored name.
func (s *Store) SaveAsset(id, filename string, r io.Reader) (Asset, error) {
	name, err := CleanAssetName(filename)
	if err != nil {
		return Asset{}, err
	}
	dir, err := s.path(id, "assets")
	if err != nil {
		return Asset{}, err
	}
	if _, err := s.Get(id); err != nil {
		return Asset{}, err
	}
	data, err := io.ReadAll(io.LimitReader(r, MaxAssetBytes+1))
	if err != nil {
		return Asset{}, err
	}
	if len(data) > MaxAssetBytes {
		return Asset{}, fmt.Errorf("%w: images can be up to %d MB", ErrTooLarge, MaxAssetBytes>>20)
	}
	sniffed := http.DetectContentType(data)
	if want := allowedExt[strings.ToLower(filepath.Ext(name))]; sniffed != want {
		return Asset{}, fmt.Errorf("%w: %q is not really a %s file", ErrBadAsset, name, strings.TrimPrefix(want, "image/"))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Asset{}, err
	}
	if err := atomicWrite(filepath.Join(dir, name), data, 0o644); err != nil {
		return Asset{}, err
	}
	return Asset{Name: name, Size: int64(len(data)), Large: len(data) > LargeAsset}, nil
}

// Duplicate copies a show (scripts in every language, images and look) under a new name.
// If name is empty, " (copy)" is added to the original name.
func (s *Store) Duplicate(id, name string) (Meta, error) {
	src, err := s.Get(id)
	if err != nil {
		return Meta{}, err
	}
	if strings.TrimSpace(name) == "" {
		name = src.Name + " (copy)"
	}
	dst, err := s.Create(name, "")
	if err != nil {
		return Meta{}, err
	}
	fail := func(err error) (Meta, error) { s.Delete(dst.ID); return Meta{}, err }
	srcDir, _ := s.path(id)
	dstDir, _ := s.path(dst.ID)
	ents, err := os.ReadDir(srcDir)
	if err != nil {
		return fail(err)
	}
	for _, e := range ents {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "script.") && strings.HasSuffix(e.Name(), ".md") {
			b, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
			if err != nil {
				return fail(err)
			}
			if err := atomicWrite(filepath.Join(dstDir, e.Name()), b, 0o644); err != nil {
				return fail(err)
			}
		}
	}
	for _, a := range s.assetNamesOnDisk(id) {
		b, err := os.ReadFile(filepath.Join(srcDir, "assets", a))
		if err != nil {
			return fail(err)
		}
		if err := atomicWrite(filepath.Join(dstDir, "assets", a), b, 0o644); err != nil {
			return fail(err)
		}
	}
	th := src.Theme
	return s.Update(dst.ID, nil, &th)
}

func (s *Store) assetNamesOnDisk(id string) []string { return s.AssetNames(id) }

// DeleteAsset removes an image.
func (s *Store) DeleteAsset(id, name string) error {
	dir, err := s.path(id, "assets")
	if err != nil {
		return err
	}
	if name != filepath.Base(name) || strings.HasPrefix(name, ".") {
		return ErrBadName
	}
	if err := os.Remove(filepath.Join(dir, name)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
