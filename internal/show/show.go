// Package show stores shows on disk. A show is a folder holding its show file (show.md, in the show's main
// language), optional translations (show.<lang>.md), and its images (assets/). Everything about a show other
// than its images lives in the show file itself: its title, its look and other settings in a [settings] block,
// and the lines. The program's defaults come from config.yaml, and the show file overrides them.
package show

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"

	"github.com/tmshort/open-caption/internal/showfile"
)

// Limits.
const (
	MaxAssetBytes = 10 << 20
	LargeAsset    = 1 << 20 // warn above this (images are not rescaled)
	MaxShowFile   = 2 << 20
)

// mainFile is the show file of a show's main language.
const mainFile = "show.md"

// Errors.
var (
	ErrNotFound = errors.New("show not found")
	ErrBadID    = errors.New("invalid show id")
	ErrBadAsset = errors.New("not a supported image")
	ErrTooLarge = errors.New("file too large")
	ErrBadName  = errors.New("invalid name")
	ErrBadTheme = errors.New("invalid look")

	idRe          = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,60}$`)
	langRe        = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)
	allowedExt    = map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".webp": "image/webp"}
	unsafeNameRun = regexp.MustCompile(`[^\p{L}\p{N} ._()\-]+`)
)

// Theme is what the audience sees (the look) plus the other show-wide numbers. It is the result of
// layering the show file's settings over the defaults.
type Theme struct {
	BG              string  `json:"bg"`
	FG              string  `json:"fg"`
	Accent          string  `json:"accent"`
	Font            string  `json:"font"`
	ScalePct        int     `json:"scalePct"`
	Speakers        string  `json:"speakers"`
	IdleMessage     string  `json:"idleMessage"`
	LoopSeconds     float64 `json:"loopSeconds"`
	ScrollbackLines int     `json:"scrollbackLines"`
	// IdleMessages are the pre-show messages of translations (by language code), each set in that
	// language's own show file.
	IdleMessages map[string]string `json:"idleMessages,omitempty"`
}

// Defaults are the values used when a show file does not set them (they come from config.yaml,
// which starts from the program's built-in values).
type Defaults struct {
	Language        string
	IdleMessage     string
	LoopSeconds     float64
	ScrollbackLines int
	Look            Theme // BG, FG, Accent, Font, ScalePct, Speakers
}

// Resolve layers a show file's settings over the defaults.
func (d Defaults) Resolve(s showfile.Settings) Theme {
	t := d.Look
	t.IdleMessage, t.LoopSeconds, t.ScrollbackLines = d.IdleMessage, d.LoopSeconds, d.ScrollbackLines
	if s.Background != nil {
		t.BG = *s.Background
	}
	if s.TextColor != nil {
		t.FG = *s.TextColor
	}
	if s.SpeakerColor != nil {
		t.Accent = *s.SpeakerColor
	}
	if s.Font != nil {
		t.Font = *s.Font
	}
	if s.TextSize != nil {
		t.ScalePct = *s.TextSize
	}
	if s.SpeakerNames != nil {
		t.Speakers = *s.SpeakerNames
	}
	if s.IdleMessage != nil && strings.TrimSpace(*s.IdleMessage) != "" {
		t.IdleMessage = *s.IdleMessage
	}
	if s.LoopSeconds != nil {
		t.LoopSeconds = *s.LoopSeconds
	}
	if s.ScrollbackLines != nil {
		t.ScrollbackLines = *s.ScrollbackLines
	}
	return t
}

// Normalize validates a look coming from Setup, filling anything missing from the defaults.
func (t Theme) Normalize(d Defaults) (Theme, error) {
	def := d.Look
	pick := func(v, fallback string) string {
		if strings.TrimSpace(v) == "" {
			return fallback
		}
		return v
	}
	var err error
	bad := func(what string, e error) (Theme, error) { return t, fmt.Errorf("%w: %s: %v", ErrBadTheme, what, e) }
	if t.BG, err = showfile.NormalizeColor(pick(t.BG, def.BG)); err != nil {
		return bad("background", err)
	}
	if t.FG, err = showfile.NormalizeColor(pick(t.FG, def.FG)); err != nil {
		return bad("text color", err)
	}
	if t.Accent, err = showfile.NormalizeColor(pick(t.Accent, def.Accent)); err != nil {
		return bad("speaker name color", err)
	}
	if t.Font, err = showfile.NormalizeFont(pick(t.Font, def.Font)); err != nil {
		return bad("font", err)
	}
	if t.Speakers, err = showfile.NormalizeSpeakerNames(pick(t.Speakers, def.Speakers)); err != nil {
		return bad("speaker names", err)
	}
	if t.ScalePct == 0 {
		t.ScalePct = def.ScalePct
	}
	if t.ScalePct < 50 || t.ScalePct > 300 {
		return bad("text size", errors.New("must be 50 to 300 percent"))
	}
	if t.LoopSeconds == 0 {
		t.LoopSeconds = d.LoopSeconds
	}
	if t.LoopSeconds <= 0 || t.LoopSeconds > 3600 {
		return bad("loop seconds", errors.New("must be more than 0 and at most 3600"))
	}
	if t.ScrollbackLines == 0 {
		t.ScrollbackLines = d.ScrollbackLines
	}
	if t.ScrollbackLines < 20 || t.ScrollbackLines > 2000 {
		return bad("scrollback lines", errors.New("must be 20 to 2000"))
	}
	t.IdleMessage = strings.TrimSpace(pick(t.IdleMessage, d.IdleMessage))
	if len(t.IdleMessage) > showfile.MaxIdleMessage {
		return bad("idle message", errors.New("too long"))
	}
	clean := map[string]string{}
	for code, msg := range t.IdleMessages {
		msg = strings.TrimSpace(msg)
		if !langRe.MatchString(code) || len(msg) > showfile.MaxIdleMessage {
			return bad("idle message for "+code, errors.New("not usable"))
		}
		if msg != "" {
			clean[code] = msg
		}
	}
	t.IdleMessages = nil
	if len(clean) > 0 {
		t.IdleMessages = clean
	}
	return t, nil
}

// Meta describes a show, as read from its show file.
type Meta struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"` // the show file's title
	Updated      time.Time `json:"updated"`
	Language     string    `json:"language"` // the main language
	LanguageName string    `json:"languageName"`
	Theme        Theme     `json:"theme"`
	// LookSaved is true once the show file sets any look setting (so Setup can tick the Look step).
	LookSaved bool `json:"lookSaved"`
}

// Asset describes an uploaded image.
type Asset struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Large bool   `json:"large"`
}

// Store manages <dir>/<id>/.
type Store struct {
	dir string
	def Defaults
}

// NewStore creates the shows directory if needed and converts any shows saved in the older layout
// (show.yaml plus script files) to show files.
func NewStore(dir string, def Defaults) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, def: def}
	if ents, err := os.ReadDir(dir); err == nil {
		for _, e := range ents {
			if e.IsDir() && ValidID(e.Name()) {
				if err := s.migrate(e.Name()); err != nil {
					log.Printf("show %s: could not convert to the new layout: %v", e.Name(), err)
				}
			}
		}
	}
	return s, nil
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

func validTitle(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 || strings.ContainsAny(name, "\r\n") {
		return "", fmt.Errorf("%w: give the show a name (up to 100 characters)", ErrBadName)
	}
	return name, nil
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Create makes a new show. initial, if not empty, is the text of its show file; if that text has no
// [settings] block one is added with the title. A name replaces the title if one is given.
func (s *Store) Create(name, initial string) (Meta, error) {
	title := strings.TrimSpace(name)
	if title != "" || strings.TrimSpace(initial) == "" {
		var err error
		if title, err = validTitle(name); err != nil {
			return Meta{}, err
		}
	}
	text := initial
	if title != "" {
		text = showfile.UpdateSettings(initial, false, func(st *showfile.Settings) { st.Title = &title })
	}
	if len(text) > MaxShowFile {
		return Meta{}, ErrTooLarge
	}
	base := title
	if base == "" {
		base = showfile.ReadSettings(initial, false).TitleOr("show")
	}
	suffix := make([]byte, 2)
	rand.Read(suffix)
	id := slug(base) + "-" + hex.EncodeToString(suffix)
	dir, _ := s.path(id)
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		return Meta{}, err
	}
	if err := atomicWrite(filepath.Join(dir, mainFile), []byte(text), 0o644); err != nil {
		return Meta{}, err
	}
	return s.Get(id)
}

// Get loads a show's details from its show file(s).
func (s *Store) Get(id string) (Meta, error) {
	p, err := s.path(id, mainFile)
	if err != nil {
		return Meta{}, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Meta{}, ErrNotFound
	} else if err != nil {
		return Meta{}, err
	}
	res := showfile.Parse(string(b), showfile.Options{})
	m := Meta{ID: id, Language: res.Language, LanguageName: res.LanguageName, LookSaved: res.Settings.LookSet()}
	if m.Language == "" {
		m.Language = s.def.Language
	}
	if m.LanguageName == "" {
		m.LanguageName = showfile.LanguageName(m.Language)
	}
	m.Name = "Untitled show"
	if res.Settings.Title != nil {
		m.Name = *res.Settings.Title
	}
	m.Theme = s.def.Resolve(res.Settings)
	if fi, err := os.Stat(p); err == nil {
		m.Updated = fi.ModTime().UTC()
	}
	// Translations carry their own pre-show message.
	dir, _ := s.path(id)
	if ents, err := os.ReadDir(dir); err == nil {
		for _, e := range ents {
			code, ok := translationCode(e.Name())
			if !ok || e.IsDir() || code == m.Language {
				continue
			}
			if tb, err := os.ReadFile(filepath.Join(dir, e.Name())); err == nil {
				if im := showfile.ReadSettings(string(tb), true).IdleMessage; im != nil && strings.TrimSpace(*im) != "" {
					if m.Theme.IdleMessages == nil {
						m.Theme.IdleMessages = map[string]string{}
					}
					m.Theme.IdleMessages[code] = *im
				}
			}
			if fi, err := e.Info(); err == nil && fi.ModTime().After(m.Updated) {
				m.Updated = fi.ModTime().UTC()
			}
		}
	}
	return m, nil
}

// translationCode returns the language code of a file named show.<code>.md.
func translationCode(name string) (string, bool) {
	if !strings.HasPrefix(name, "show.") || !strings.HasSuffix(name, ".md") || name == mainFile {
		return "", false
	}
	code := strings.TrimSuffix(strings.TrimPrefix(name, "show."), ".md")
	return code, langRe.MatchString(code)
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

// Update changes the title and/or the look by rewriting the [settings] of the show file(s). Nil
// arguments are left alone. A saved look writes every look setting, so the show file says exactly
// how the show looks wherever it is copied.
func (s *Store) Update(id string, name *string, look *Theme) (Meta, error) {
	m, err := s.Get(id)
	if err != nil {
		return m, err
	}
	var title string
	if name != nil {
		if title, err = validTitle(*name); err != nil {
			return m, err
		}
	}
	var th Theme
	if look != nil {
		if th, err = look.Normalize(s.def); err != nil {
			return m, err
		}
	}
	mainText, err := s.ReadShowFile(id, "")
	if err != nil {
		return m, err
	}
	mainText = showfile.UpdateSettings(mainText, false, func(st *showfile.Settings) {
		if name != nil {
			st.Title = &title
		}
		if look != nil {
			st.Background, st.TextColor, st.SpeakerColor, st.Font = &th.BG, &th.FG, &th.Accent, &th.Font
			st.TextSize, st.SpeakerNames, st.IdleMessage = &th.ScalePct, &th.Speakers, &th.IdleMessage
			st.LoopSeconds, st.ScrollbackLines = &th.LoopSeconds, &th.ScrollbackLines
		}
	})
	if err := s.WriteShowFile(id, "", mainText); err != nil {
		return m, err
	}
	if look != nil {
		langs, _ := s.Languages(id)
		for _, code := range langs[1:] {
			text, err := s.ReadShowFile(id, code)
			if err != nil {
				continue
			}
			msg, has := th.IdleMessages[code]
			text = showfile.UpdateSettings(text, true, func(st *showfile.Settings) {
				if has {
					st.IdleMessage = &msg
				} else {
					st.IdleMessage = nil
				}
			})
			if err := s.WriteShowFile(id, code, text); err != nil {
				return m, err
			}
		}
	}
	return s.Get(id)
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

// fileName is the file holding a language's show file ("" or the main language: show.md).
func (s *Store) fileName(id, lang string) (string, error) {
	if lang != "" {
		if !langRe.MatchString(lang) {
			return "", fmt.Errorf("invalid language %q", lang)
		}
		if m, err := s.Get(id); err == nil && lang == m.Language {
			lang = ""
		}
	}
	if lang == "" {
		return mainFile, nil
	}
	return "show." + lang + ".md", nil
}

// ReadShowFile returns the show file text for a language ("" = the main language). A translation that
// does not exist yet reads as empty.
func (s *Store) ReadShowFile(id, lang string) (string, error) {
	n, err := s.fileName(id, lang)
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

// WriteShowFile saves the show file verbatim.
func (s *Store) WriteShowFile(id, lang, text string) error {
	if len(text) > MaxShowFile {
		return ErrTooLarge
	}
	if _, err := s.Get(id); err != nil {
		return err
	}
	n, err := s.fileName(id, lang)
	if err != nil {
		return err
	}
	p, err := s.path(id, n)
	if err != nil {
		return err
	}
	return atomicWrite(p, []byte(text), 0o644)
}

// Languages lists the languages a show has show files for, the main language first.
func (s *Store) Languages(id string) ([]string, error) {
	m, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	dir, _ := s.path(id)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var others []string
	for _, e := range ents {
		if code, ok := translationCode(e.Name()); ok && !e.IsDir() && code != m.Language {
			others = append(others, code)
		}
	}
	sort.Strings(others)
	return append([]string{m.Language}, others...), nil
}

// DeleteShowFile removes a translation. The main language cannot be removed.
func (s *Store) DeleteShowFile(id, lang string) error {
	m, err := s.Get(id)
	if err != nil {
		return err
	}
	if lang == "" || lang == m.Language {
		return fmt.Errorf("%w: the main language cannot be removed", ErrBadName)
	}
	if !langRe.MatchString(lang) {
		return fmt.Errorf("invalid language %q", lang)
	}
	p, _ := s.path(id, "show."+lang+".md")
	if err := os.Remove(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return err
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

// Duplicate copies a show (its show files in every language, and its images) under a new title.
// If name is empty, " (copy)" is added to the original title.
func (s *Store) Duplicate(id, name string) (Meta, error) {
	src, err := s.Get(id)
	if err != nil {
		return Meta{}, err
	}
	title := strings.TrimSpace(name)
	if title == "" {
		title = src.Name + " (copy)"
	}
	if title, err = validTitle(title); err != nil {
		return Meta{}, err
	}
	dst, err := s.Create(title, "")
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
		_, isTrans := translationCode(e.Name())
		if e.IsDir() || (e.Name() != mainFile && !isTrans) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
		if err != nil {
			return fail(err)
		}
		if e.Name() == mainFile {
			b = []byte(showfile.UpdateSettings(string(b), false, func(st *showfile.Settings) { st.Title = &title }))
		}
		if err := atomicWrite(filepath.Join(dstDir, e.Name()), b, 0o644); err != nil {
			return fail(err)
		}
	}
	for _, a := range s.AssetNames(id) {
		b, err := os.ReadFile(filepath.Join(srcDir, "assets", a))
		if err != nil {
			return fail(err)
		}
		if err := atomicWrite(filepath.Join(dstDir, "assets", a), b, 0o644); err != nil {
			return fail(err)
		}
	}
	return s.Get(dst.ID)
}

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

// ---- the older layout (show.yaml plus script files) ----

type oldMeta struct {
	Name      string `yaml:"name"`
	Language  string `yaml:"language"`
	LookSaved bool   `yaml:"look_saved"`
	Theme     struct {
		BG           string            `yaml:"bg"`
		FG           string            `yaml:"fg"`
		Accent       string            `yaml:"accent"`
		Font         string            `yaml:"font"`
		ScalePct     int               `yaml:"scale_pct"`
		Speakers     string            `yaml:"speakers"`
		IdleMessage  string            `yaml:"idle_message"`
		IdleMessages map[string]string `yaml:"idle_messages"`
	} `yaml:"theme"`
}

// convertOld turns the old layout's metadata and scripts into show file contents (file name to text).
func (d Defaults) convertOld(old oldMeta, scripts map[string]string) map[string]string {
	main := old.Language
	if main == "" {
		main = d.Language
	}
	out := map[string]string{}
	mainText := scripts[main]
	if main != d.Language && !strings.Contains(strings.ToLower(mainText), "[language") {
		mainText = "[language: " + main + " | " + showfile.LanguageName(main) + "]\n" + mainText
	}
	mainText = showfile.UpdateSettings(mainText, false, func(st *showfile.Settings) {
		if t := strings.TrimSpace(old.Name); t != "" {
			st.Title = &t
		}
		th := old.Theme
		if old.LookSaved {
			for _, f := range []struct {
				dst **string
				val string
				fn  func(string) (string, error)
			}{{&st.Background, th.BG, showfile.NormalizeColor}, {&st.TextColor, th.FG, showfile.NormalizeColor}, {&st.SpeakerColor, th.Accent, showfile.NormalizeColor},
				{&st.Font, th.Font, showfile.NormalizeFont}, {&st.SpeakerNames, th.Speakers, showfile.NormalizeSpeakerNames}} {
				if v, err := f.fn(f.val); err == nil && f.val != "" {
					v := v
					*f.dst = &v
				}
			}
			if th.ScalePct >= 50 && th.ScalePct <= 300 {
				n := th.ScalePct
				st.TextSize = &n
			}
		}
		if im := strings.TrimSpace(th.IdleMessage); im != "" && (old.LookSaved || im != d.IdleMessage) {
			st.IdleMessage = &im
		}
	})
	out[mainFile] = mainText
	for code, text := range scripts {
		if code == main || !langRe.MatchString(code) {
			continue
		}
		if im := strings.TrimSpace(old.Theme.IdleMessages[code]); im != "" {
			text = showfile.UpdateSettings(text, true, func(st *showfile.Settings) { st.IdleMessage = &im })
		}
		out["show."+code+".md"] = text
	}
	return out
}

// migrate converts a show saved in the older layout. The old files are kept in an old-format folder.
func (s *Store) migrate(id string) error {
	dir, _ := s.path(id)
	if _, err := os.Stat(filepath.Join(dir, mainFile)); err == nil {
		return nil // already in the current layout
	}
	yb, err := os.ReadFile(filepath.Join(dir, "show.yaml"))
	if err != nil {
		return nil // nothing recognisable to convert
	}
	var old oldMeta
	if err := yaml.Unmarshal(yb, &old); err != nil {
		return err
	}
	scripts := map[string]string{}
	var oldFiles []string
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		n := e.Name()
		var code string
		switch {
		case strings.HasPrefix(n, "script.") && strings.HasSuffix(n, ".md"):
			code = strings.TrimSuffix(strings.TrimPrefix(n, "script."), ".md")
		case strings.HasPrefix(n, "show.") && strings.HasSuffix(n, ".md") && n != mainFile:
			code = strings.TrimSuffix(strings.TrimPrefix(n, "show."), ".md")
		default:
			continue
		}
		if b, err := os.ReadFile(filepath.Join(dir, n)); err == nil && langRe.MatchString(code) {
			if _, dup := scripts[code]; !dup || strings.HasPrefix(n, "script.") {
				scripts[code] = string(b)
			}
			oldFiles = append(oldFiles, n)
		}
	}
	files := s.def.convertOld(old, scripts)
	for name, text := range files {
		if err := atomicWrite(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			return err
		}
	}
	keep := filepath.Join(dir, "old-format")
	os.MkdirAll(keep, 0o755)
	for _, n := range append(oldFiles, "show.yaml") {
		if _, isNew := files[n]; isNew {
			continue // a stage-1 name that the new layout also uses was just rewritten in place
		}
		os.Rename(filepath.Join(dir, n), filepath.Join(keep, n))
	}
	return nil
}

// ImportFiles creates a show from exported files (file name to text): either the current layout
// (show.md, show.<lang>.md) or the older one (show.yaml with script.<lang>.md). The new show's title
// gets suffix appended.
func (s *Store) ImportFiles(files map[string]string, suffix string) (Meta, error) {
	out := map[string]string{}
	if _, current := files[mainFile]; current {
		for n, t := range files {
			if _, isT := translationCode(n); n == mainFile || isT {
				out[n] = t
			}
		}
	} else if y, ok := files["show.yaml"]; ok {
		var old oldMeta
		if err := yaml.Unmarshal([]byte(y), &old); err != nil || strings.TrimSpace(old.Name) == "" {
			return Meta{}, fmt.Errorf("%w: unreadable show.yaml", ErrBadName)
		}
		scripts := map[string]string{}
		for n, t := range files {
			if strings.HasPrefix(n, "script.") && strings.HasSuffix(n, ".md") {
				scripts[strings.TrimSuffix(strings.TrimPrefix(n, "script."), ".md")] = t
			}
		}
		out = s.def.convertOld(old, scripts)
	}
	if out[mainFile] == "" && len(out) == 0 {
		return Meta{}, fmt.Errorf("%w: no show file found", ErrNotFound)
	}
	title := showfile.ReadSettings(out[mainFile], false).TitleOr("Imported show") + suffix
	m, err := s.Create(title, out[mainFile])
	if err != nil {
		return m, err
	}
	for n, t := range out {
		if n != mainFile {
			if err := s.WriteShowFile(m.ID, strings.TrimSuffix(strings.TrimPrefix(n, "show."), ".md"), t); err != nil {
				s.Delete(m.ID)
				return Meta{}, err
			}
		}
	}
	return s.Get(m.ID)
}
