package show

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testDefaults() Defaults {
	return Defaults{Language: "en", IdleMessage: "Soon", LoopSeconds: 10, ScrollbackLines: 200,
		Look: Theme{BG: "#000000", FG: "#ffffff", Accent: "#ffffff", Font: "system", ScalePct: 100, Speakers: "label"}}
}

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir(), testDefaults())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func pngBytes(t *testing.T, size int) []byte {
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, size, size))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func read(t *testing.T, s *Store, id, lang string) string {
	t.Helper()
	txt, err := s.ReadShowFile(id, lang)
	if err != nil {
		t.Fatal(err)
	}
	return txt
}

func TestCreateGetListUpdateDelete(t *testing.T) {
	s := newStore(t)
	m, err := s.Create("  The Odd Couple!  ", "Hello\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(m.ID, "the-odd-couple-") || !ValidID(m.ID) || m.Name != "The Odd Couple!" {
		t.Errorf("%+v", m)
	}
	// Nothing about the show lives outside its show file: the title is in a [settings] block.
	txt := read(t, s, m.ID, "")
	if !strings.Contains(txt, "[settings]\ntitle: The Odd Couple!\n[/settings]") || !strings.HasSuffix(txt, "Hello\n") {
		t.Errorf("show file:\n%s", txt)
	}
	if m.Theme.BG != "#000000" || m.Theme.FG != "#ffffff" || m.Theme.IdleMessage != "Soon" || m.LookSaved {
		t.Errorf("a new show uses the defaults and has no saved look: %+v", m)
	}
	if list, _ := s.List(); len(list) != 1 || list[0].ID != m.ID {
		t.Errorf("%+v", list)
	}
	name := "Renamed"
	th := Theme{BG: "#ffffff", FG: "#000000", ScalePct: 150, Speakers: "color", Font: "serif"}
	m2, err := s.Update(m.ID, &name, &th)
	if err != nil || m2.Name != "Renamed" || m2.Theme.ScalePct != 150 || m2.Theme.Accent != "#ffffff" || !m2.LookSaved {
		t.Fatalf("%v %+v", err, m2)
	}
	txt = read(t, s, m.ID, "")
	for _, want := range []string{"title: Renamed", "background: #ffffff", "text color: #000000", "font: serif", "text size: 150%", "speaker names: color", "Hello"} {
		if !strings.Contains(txt, want) {
			t.Errorf("show file lacks %q:\n%s", want, txt)
		}
	}
	if err := s.Delete(m.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(m.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v", err)
	}
	if _, err := s.Create("   ", ""); !errors.Is(err, ErrBadName) {
		t.Errorf("blank name: %v", err)
	}
	if _, err := s.Create("two\nlines", ""); !errors.Is(err, ErrBadName) {
		t.Errorf("multi-line name: %v", err)
	}
}

func TestTheLayersDefaultsThenConfigThenShowFile(t *testing.T) {
	// A show file that sets some things; config defaults supply the rest.
	cfg := testDefaults()
	cfg.Look.BG, cfg.Look.Font, cfg.IdleMessage, cfg.LoopSeconds = "#101010", "serif", "Config says soon", 7
	s, _ := NewStore(t.TempDir(), cfg)
	m, err := s.Create("", "[settings]\ntitle: Layers\nfont: mono\ntext color: yellow\nloop seconds: 20\n[/settings]\nHello")
	if err != nil {
		t.Fatal(err)
	}
	th := m.Theme
	if m.Name != "Layers" {
		t.Errorf("title from the show file: %q", m.Name)
	}
	if th.Font != "mono" || th.FG != "#ffee00" || th.LoopSeconds != 20 {
		t.Errorf("the show file overrides: %+v", th)
	}
	if th.BG != "#101010" || th.IdleMessage != "Config says soon" || th.ScrollbackLines != 200 || th.ScalePct != 100 {
		t.Errorf("config.yaml defaults show through where the show file is silent: %+v", th)
	}
	if !m.LookSaved {
		t.Error("a show file that sets the look counts as having a saved look")
	}
	// Changing a default later changes shows that never set it, and only those.
	cfg.Look.BG, cfg.Look.Font = "#222222", "dyslexic"
	s2 := &Store{dir: s.dir, def: cfg}
	m2, _ := s2.Get(m.ID)
	if m2.Theme.BG != "#222222" || m2.Theme.Font != "mono" {
		t.Errorf("%+v", m2.Theme)
	}
}

func TestCreateKeepsTheSettingsAnInitialFileAlreadyHas(t *testing.T) {
	s := newStore(t)
	initial := "<!-- banner -->\n[settings]\ntitle: Sample\nfont: dyslexic\n[/settings]\nHi"
	m, err := s.Create("", initial)
	if err != nil || m.Name != "Sample" || m.Theme.Font != "dyslexic" || read(t, s, m.ID, "") != initial {
		t.Fatalf("%v %+v\n%s", err, m, read(t, s, m.ID, ""))
	}
	m, err = s.Create("Mine", initial)
	if err != nil || m.Name != "Mine" || m.Theme.Font != "dyslexic" || !strings.Contains(read(t, s, m.ID, ""), "<!-- banner -->") {
		t.Errorf("a name replaces the title but keeps the other settings: %v %+v", err, m)
	}
}

func TestIDsCannotEscape(t *testing.T) {
	s := newStore(t)
	for _, id := range []string{"", "..", "../x", "a/b", "A", "-x", strings.Repeat("a", 80)} {
		if _, err := s.Get(id); !errors.Is(err, ErrBadID) {
			t.Errorf("Get(%q) = %v", id, err)
		}
		if err := s.Delete(id); !errors.Is(err, ErrBadID) {
			t.Errorf("Delete(%q) = %v", id, err)
		}
	}
	m, _ := s.Create("x", "")
	if _, err := s.ReadShowFile(m.ID, "../../etc"); err == nil {
		t.Error("language must be validated")
	}
}

func TestLookValidation(t *testing.T) {
	d := testDefaults()
	bad := []Theme{{BG: "red-ish"}, {FG: "#12345"}, {Font: "comic"}, {ScalePct: 10}, {ScalePct: 900}, {Speakers: "loud"}, {LoopSeconds: -1}, {ScrollbackLines: 5}, {IdleMessage: strings.Repeat("x", 300)}, {IdleMessages: map[string]string{"../x": "no"}}}
	for _, th := range bad {
		if _, err := th.Normalize(d); !errors.Is(err, ErrBadTheme) {
			t.Errorf("%+v accepted", th)
		}
	}
	th, err := (Theme{BG: "#FFF", Font: "OpenDyslexic", Speakers: "different colors", IdleMessages: map[string]string{"es": "  Pronto  ", "fr": ""}}).Normalize(d)
	if err != nil || th.BG != "#ffffff" || th.Font != "dyslexic" || th.Speakers != "color" || th.FG != "#ffffff" || th.IdleMessage != "Soon" ||
		len(th.IdleMessages) != 1 || th.IdleMessages["es"] != "Pronto" {
		t.Errorf("%v %+v", err, th)
	}
}

func TestAssets(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("x", "")
	a, err := s.SaveAsset(m.ID, "My Curtain (1).PNG", bytes.NewReader(pngBytes(t, 8)))
	if err != nil || a.Name != "My Curtain (1).PNG" {
		t.Fatalf("%v %+v", err, a)
	}
	if _, err := s.SaveAsset(m.ID, "../../evil.png", bytes.NewReader(pngBytes(t, 8))); err != nil {
		t.Errorf("path parts should be stripped, not rejected: %v", err)
	}
	for name, data := range map[string][]byte{
		"fake.png":    []byte("<svg onload=alert(1)>"),
		"logo.svg":    []byte("<svg/>"),
		"note.txt":    []byte("hi"),
		"script.js":   []byte("alert(1)"),
		"photo.jpg":   pngBytes(t, 8), // PNG bytes with a .jpg name
		".hidden.png": pngBytes(t, 8),
	} {
		if _, err := s.SaveAsset(m.ID, name, bytes.NewReader(data)); err == nil && name != ".hidden.png" {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := s.SaveAsset(m.ID, "big.png", bytes.NewReader(append(pngBytes(t, 8), make([]byte, MaxAssetBytes)...))); !errors.Is(err, ErrTooLarge) {
		t.Errorf("oversize: %v", err)
	}
	if names := s.AssetNames(m.ID); len(names) != 3 {
		t.Errorf("%v", names)
	}
	if err := s.DeleteAsset(m.ID, "evil.png"); err != nil {
		t.Error(err)
	}
	if err := s.DeleteAsset(m.ID, "../show.md"); err == nil {
		t.Error("traversal accepted")
	}
	if _, err := s.SaveAsset("nope-0000", "a.png", bytes.NewReader(pngBytes(t, 8))); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown show: %v", err)
	}
}

func TestDuplicateCopiesEverything(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Original", "Hello\nWorld\n")
	s.WriteShowFile(m.ID, "es", "[language: es | Español]\nHola\n")
	s.SaveAsset(m.ID, "a.png", bytes.NewReader(pngBytes(t, 8)))
	th := Theme{BG: "#001122", FG: "#ffee00", ScalePct: 150}
	s.Update(m.ID, nil, &th)

	cp, err := s.Duplicate(m.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if cp.ID == m.ID || cp.Name != "Original (copy)" || cp.Theme.BG != "#001122" || cp.Theme.ScalePct != 150 {
		t.Errorf("%+v", cp)
	}
	if txt := read(t, s, cp.ID, ""); !strings.Contains(txt, "Hello\nWorld") || !strings.Contains(txt, "title: Original (copy)") {
		t.Errorf("show file %q", txt)
	}
	if txt := read(t, s, cp.ID, "es"); !strings.Contains(txt, "Hola") {
		t.Errorf("translation %q", txt)
	}
	if names := s.AssetNames(cp.ID); len(names) != 1 || names[0] != "a.png" {
		t.Errorf("assets %v", names)
	}
	s.WriteShowFile(cp.ID, "", "changed")
	if txt := read(t, s, m.ID, ""); !strings.Contains(txt, "Hello\nWorld") {
		t.Error("original changed with the copy")
	}
	if c2, err := s.Duplicate(m.ID, "Mine"); err != nil || c2.Name != "Mine" {
		t.Errorf("named copy: %v %+v", err, c2)
	}
	if _, err := s.Duplicate("nope-0000", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing show: %v", err)
	}
}

func TestLanguagesAndTranslations(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Multi", "Hello")
	if l, _ := s.Languages(m.ID); len(l) != 1 || l[0] != "en" {
		t.Fatalf("%v", l)
	}
	s.WriteShowFile(m.ID, "fr", "Bonjour")
	s.WriteShowFile(m.ID, "es", "Hola")
	if l, _ := s.Languages(m.ID); len(l) != 3 || l[0] != "en" || l[1] != "es" || l[2] != "fr" {
		t.Errorf("main first, then alphabetical: %v", l)
	}
	if err := s.DeleteShowFile(m.ID, "en"); err == nil {
		t.Error("the main language must not be removable")
	}
	if err := s.DeleteShowFile(m.ID, "fr"); err != nil {
		t.Error(err)
	}
	if err := s.DeleteShowFile(m.ID, "fr"); !errors.Is(err, ErrNotFound) {
		t.Errorf("already gone: %v", err)
	}
	if err := s.DeleteShowFile(m.ID, "../x"); err == nil {
		t.Error("invalid language accepted")
	}
	// Asking for the main language by its code reads the main file.
	if txt := read(t, s, m.ID, "en"); !strings.Contains(txt, "Hello") {
		t.Errorf("%q", txt)
	}
}

func TestMainLanguageComesFromTheShowFile(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Français", "[language: fr | Français]\nBonjour")
	if m.Language != "fr" || m.LanguageName != "Français" {
		t.Fatalf("%+v", m)
	}
	s.WriteShowFile(m.ID, "en", "[language: en | English]\nHello")
	if l, _ := s.Languages(m.ID); len(l) != 2 || l[0] != "fr" || l[1] != "en" {
		t.Errorf("%v", l)
	}
	if txt := read(t, s, m.ID, "fr"); !strings.Contains(txt, "Bonjour") {
		t.Errorf("main by code: %q", txt)
	}
	if txt := read(t, s, m.ID, "en"); !strings.Contains(txt, "Hello") {
		t.Errorf("translation: %q", txt)
	}
}

func TestPreShowMessagesLiveInEachLanguagesOwnShowFile(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Msg", "Hello")
	s.WriteShowFile(m.ID, "es", "[language: es | Español]\nHola")
	th := Theme{IdleMessage: "Soon, friends", IdleMessages: map[string]string{"es": "Pronto, amigos", "xx": "no file for this one"}}
	m2, err := s.Update(m.ID, nil, &th)
	if err != nil || m2.Theme.IdleMessage != "Soon, friends" || m2.Theme.IdleMessages["es"] != "Pronto, amigos" {
		t.Fatalf("%v %+v", err, m2.Theme)
	}
	if es := read(t, s, m.ID, "es"); !strings.Contains(es, "idle message: Pronto, amigos") || strings.Contains(es, "font:") {
		t.Errorf("only the idle message goes in a translation:\n%s", es)
	}
	if _, has := m2.Theme.IdleMessages["xx"]; has {
		t.Error("no file, no message")
	}
	th.IdleMessages = nil
	m3, _ := s.Update(m.ID, nil, &th)
	if len(m3.Theme.IdleMessages) != 0 || strings.Contains(read(t, s, m.ID, "es"), "idle message") {
		t.Errorf("clearing removes the line: %+v", m3.Theme.IdleMessages)
	}
}

func TestOlderLayoutIsConvertedAndKept(t *testing.T) {
	dir := t.TempDir()
	id := "old-show-1234"
	sd := filepath.Join(dir, id)
	os.MkdirAll(filepath.Join(sd, "assets"), 0o755)
	os.WriteFile(filepath.Join(sd, "show.yaml"), []byte(`name: Old Show
created: 2026-01-01T00:00:00Z
updated: 2026-01-02T00:00:00Z
language: en
theme:
  bg: "#101010"
  fg: "#ffee00"
  accent: "#ffffff"
  font: serif
  scale_pct: 130
  speakers: color
  idle_message: Almost time
  idle_messages:
    es: Casi es la hora
look_saved: true
`), 0o644)
	os.WriteFile(filepath.Join(sd, "script.en.md"), []byte("[language: en | English]\nHello\n"), 0o644)
	os.WriteFile(filepath.Join(sd, "script.es.md"), []byte("[language: es | Español]\nHola\n"), 0o644)
	os.WriteFile(filepath.Join(sd, "assets", "a.png"), pngBytes(t, 4), 0o644)

	s, err := NewStore(dir, testDefaults())
	if err != nil {
		t.Fatal(err)
	}
	m, err := s.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	th := m.Theme
	if m.Name != "Old Show" || th.BG != "#101010" || th.FG != "#ffee00" || th.Font != "serif" || th.ScalePct != 130 || th.Speakers != "color" ||
		th.IdleMessage != "Almost time" || th.IdleMessages["es"] != "Casi es la hora" || !m.LookSaved {
		t.Errorf("%+v", m)
	}
	if txt := read(t, s, id, ""); !strings.Contains(txt, "Hello") || !strings.Contains(txt, "title: Old Show") {
		t.Errorf("main:\n%s", txt)
	}
	if txt := read(t, s, id, "es"); !strings.Contains(txt, "Hola") || !strings.Contains(txt, "idle message: Casi es la hora") {
		t.Errorf("translation:\n%s", txt)
	}
	for _, gone := range []string{"show.yaml", "script.en.md", "script.es.md"} {
		if _, err := os.Stat(filepath.Join(sd, gone)); err == nil {
			t.Errorf("%s should have been moved", gone)
		}
		if _, err := os.Stat(filepath.Join(sd, "old-format", gone)); err != nil {
			t.Errorf("%s should be kept in old-format/", gone)
		}
	}
	if names := s.AssetNames(id); len(names) != 1 {
		t.Errorf("images are untouched: %v", names)
	}
	// Opening the store again does nothing more.
	if _, err := NewStore(dir, testDefaults()); err != nil {
		t.Fatal(err)
	}
}

func TestOlderShowThatNeverSavedItsLookKeepsTheDefaults(t *testing.T) {
	dir := t.TempDir()
	sd := filepath.Join(dir, "plain-0001")
	os.MkdirAll(sd, 0o755)
	os.WriteFile(filepath.Join(sd, "show.yaml"), []byte("name: Plain\ntheme:\n  bg: \"#000000\"\n  fg: \"#ffffff\"\n  idle_message: Soon\n"), 0o644)
	os.WriteFile(filepath.Join(sd, "script.en.md"), []byte("Hello"), 0o644)
	s, _ := NewStore(dir, testDefaults())
	m, _ := s.Get("plain-0001")
	txt := read(t, s, "plain-0001", "")
	if m.LookSaved || strings.Contains(txt, "background") || strings.Contains(txt, "idle message") {
		t.Errorf("an untouched look must stay the default, not become explicit: %+v\n%s", m.LookSaved, txt)
	}
}

func TestImportFilesNewAndOldLayouts(t *testing.T) {
	s := newStore(t)
	m, err := s.ImportFiles(map[string]string{"show.md": "[settings]\ntitle: Exported\nfont: serif\n[/settings]\nHi", "show.es.md": "Hola", "readme.txt": "ignored"}, " (imported)")
	if err != nil || m.Name != "Exported (imported)" || m.Theme.Font != "serif" {
		t.Fatalf("%v %+v", err, m)
	}
	if l, _ := s.Languages(m.ID); len(l) != 2 {
		t.Errorf("%v", l)
	}
	old, err := s.ImportFiles(map[string]string{"show.yaml": "name: Old\nlanguage: en\n", "script.en.md": "Hello"}, " (imported)")
	if err != nil || old.Name != "Old (imported)" || !strings.Contains(read(t, s, old.ID, ""), "Hello") {
		t.Errorf("%v %+v", err, old)
	}
	if _, err := s.ImportFiles(map[string]string{"nothing.txt": "x"}, ""); err == nil {
		t.Error("an archive with no show file must be refused")
	}
	if _, err := s.ImportFiles(map[string]string{"show.yaml": ": not yaml ["}, ""); err == nil {
		t.Error("garbage must be refused")
	}
}
