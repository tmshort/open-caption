// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package show

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"strings"
	"testing"
	"time"
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

func TestImportFiles(t *testing.T) {
	s := newStore(t)
	m, err := s.ImportFiles(map[string]string{"show.md": "[settings]\ntitle: Exported\nfont: serif\n[/settings]\nHi", "show.es.md": "Hola", "readme.txt": "ignored"}, " (imported)")
	if err != nil || m.Name != "Exported (imported)" || m.Theme.Font != "serif" {
		t.Fatalf("%v %+v", err, m)
	}
	if l, _ := s.Languages(m.ID); len(l) != 2 {
		t.Errorf("%v", l)
	}
	if _, err := s.ImportFiles(map[string]string{"nothing.txt": "x"}, ""); err == nil {
		t.Error("an archive with no show file must be refused")
	}
}

func TestArchiveAndRestoreAreStoredInTheShowFile(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Old play", "Hello")
	a, err := s.SetArchived(m.ID, true)
	if err != nil || !a.Archived {
		t.Fatalf("%v %+v", err, a)
	}
	if txt := read(t, s, m.ID, ""); !strings.Contains(txt, "archived: yes") || !strings.Contains(txt, "title: Old play") || !strings.HasSuffix(txt, "Hello\n") && !strings.HasSuffix(txt, "Hello") {
		t.Errorf("show file:\n%s", txt)
	}
	cp, _ := s.Duplicate(m.ID, "")
	if cp.Archived {
		t.Error("a copy of an archived show is not archived")
	}
	r, err := s.SetArchived(m.ID, false)
	if err != nil || r.Archived || strings.Contains(read(t, s, m.ID, ""), "archived") {
		t.Errorf("restore: %v %+v", err, r)
	}
	if _, err := s.SetArchived("nope-0000", true); !errors.Is(err, ErrNotFound) {
		t.Errorf("%v", err)
	}
	// An archived show stays archived through a look change but an import of one is restored.
	s.SetArchived(m.ID, true)
	th := Theme{ScalePct: 120}
	if u, _ := s.Update(m.ID, nil, &th); !u.Archived {
		t.Error("changing the look must not restore the show")
	}
	imp, err := s.ImportFiles(map[string]string{"show.md": read(t, s, m.ID, "")}, " (imported)")
	if err != nil || imp.Archived {
		t.Errorf("import: %v %+v", err, imp)
	}
}

func TestEverySaveKeepsTheVersionItReplaced(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("History", "one")
	for _, text := range []string{"two", "two", "three\nlines", "four"} { // the repeated save changes nothing
		if err := s.WriteShowFile(m.ID, "", text); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)
	}
	v, err := s.History(m.ID, "")
	if err != nil || len(v) != 3 {
		t.Fatalf("%v %+v", err, v)
	}
	if v[0].ID <= v[1].ID {
		t.Errorf("newest first: %+v", v)
	}
	got := func(i int) string {
		txt, err := s.ReadHistory(m.ID, "", v[i].ID)
		if err != nil {
			t.Fatal(err)
		}
		return txt
	}
	if got(0) != "three\nlines" || got(1) != "two" || !strings.Contains(got(2), "title: History") || v[0].Lines != 2 {
		t.Errorf("%q %q %q", got(0), got(1), got(2))
	}
	// Translations have their own history, and nothing outside it can be read.
	s.WriteShowFile(m.ID, "es", "uno")
	s.WriteShowFile(m.ID, "es", "dos")
	if es, _ := s.History(m.ID, "es"); len(es) != 1 {
		t.Errorf("%+v", es)
	}
	for _, bad := range []string{"../show.md", "x", "20260101T000000.000Z/../../x"} {
		if _, err := s.ReadHistory(m.ID, "", bad); !errors.Is(err, ErrNotFound) {
			t.Errorf("%q: %v", bad, err)
		}
	}
	if _, err := s.History("nope-0000", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("%v", err)
	}
	// The history is not part of a copy or of the language list.
	cp, _ := s.Duplicate(m.ID, "")
	if h, _ := s.History(cp.ID, ""); len(h) != 0 {
		t.Errorf("a copy starts with a clean history: %+v", h)
	}
	if l, _ := s.Languages(m.ID); len(l) != 2 {
		t.Errorf("%v", l)
	}
}

func TestOnlyTheNewestVersionsAreKept(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Many", "0")
	for i := 1; i <= MaxHistory+8; i++ {
		s.WriteShowFile(m.ID, "", strings.Repeat("x", i))
		time.Sleep(time.Millisecond)
	}
	v, _ := s.History(m.ID, "")
	if len(v) != MaxHistory {
		t.Fatalf("%d versions", len(v))
	}
	if txt, _ := s.ReadHistory(m.ID, "", v[0].ID); len(txt) != MaxHistory+7 {
		t.Errorf("newest kept version has %d characters", len(txt))
	}
}

func TestNamedVersionsAreKeptForGood(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Named", "v0")
	s.WriteShowFile(m.ID, "", "v1")
	hist, _ := s.History(m.ID, "")
	if len(hist) != 1 {
		t.Fatalf("%+v", hist)
	}
	if err := s.LabelVersion(m.ID, "", hist[0].ID, "  Before the cast change "); err != nil {
		t.Fatal(err)
	}
	if h, _ := s.History(m.ID, ""); h[0].Label != "Before the cast change" {
		t.Errorf("%+v", h)
	}
	// Keep what is saved now under a name, and push far more than MaxHistory ordinary versions through.
	kept, err := s.KeepVersion(m.ID, "", "Opening night")
	if err != nil || kept.Label != "Opening night" {
		t.Fatalf("%v %+v", err, kept)
	}
	for i := 0; i < MaxHistory+10; i++ {
		s.WriteShowFile(m.ID, "", strings.Repeat("x", i+2))
	}
	h, _ := s.History(m.ID, "")
	named := map[string]bool{}
	for _, v := range h {
		if v.Label != "" {
			named[v.Label] = true
		}
	}
	if !named["Before the cast change"] || !named["Opening night"] || len(h) != MaxHistory+2 {
		t.Errorf("named versions must survive pruning: %v (%d versions)", named, len(h))
	}
	if txt, _ := s.ReadHistory(m.ID, "", kept.ID); !strings.HasSuffix(txt, "v1") {
		t.Errorf("%q", txt)
	}
	// Removing the name makes it an ordinary version again; bad names and unknown versions are refused.
	if err := s.LabelVersion(m.ID, "", kept.ID, ""); err != nil {
		t.Error(err)
	}
	for _, bad := range []string{strings.Repeat("a", MaxLabel+1), "two\nlines"} {
		if err := s.LabelVersion(m.ID, "", h[0].ID, bad); !errors.Is(err, ErrBadLabel) {
			t.Errorf("%q: %v", bad, err)
		}
	}
	if _, err := s.KeepVersion(m.ID, "", "  "); !errors.Is(err, ErrBadLabel) {
		t.Errorf("%v", err)
	}
	if err := s.LabelVersion(m.ID, "", "20260101T000000.000Z", "x"); !errors.Is(err, ErrNotFound) {
		t.Errorf("%v", err)
	}
}

// Two shows whose ids would be the same must not share a folder: the second gets another suffix and the
// first keeps its show file.
func TestCreateNeverReusesAnExistingShowFolder(t *testing.T) {
	s := newStore(t)
	calls := 0
	old := randRead
	randRead = func(b []byte) {
		calls++
		if calls <= 2 {
			b[0], b[1] = 0xab, 0xcd
		} else {
			b[0], b[1] = 0x12, 0x34
		}
	}
	defer func() { randRead = old }()
	first, err := s.Create("Same Title", "First show\n")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Create("Same Title", "Second show\n")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatalf("both shows got the id %q", first.ID)
	}
	if got := read(t, s, first.ID, ""); !strings.Contains(got, "First show") {
		t.Errorf("the first show was overwritten:\n%s", got)
	}
}
