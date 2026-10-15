package show

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"strings"
	"testing"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir(), "en", "Soon")
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

func TestCreateGetListUpdateDelete(t *testing.T) {
	s := newStore(t)
	m, err := s.Create("  The Odd Couple!  ", "Hello\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(m.ID, "the-odd-couple-") || !ValidID(m.ID) || m.Name != "The Odd Couple!" {
		t.Errorf("%+v", m)
	}
	if m.Theme.BG != "#000000" || m.Theme.FG != "#ffffff" || m.Theme.IdleMessage != "Soon" {
		t.Errorf("default theme should be black/white: %+v", m.Theme)
	}
	if txt, _ := s.ReadScript(m.ID, ""); txt != "Hello\n" {
		t.Errorf("script %q", txt)
	}
	list, _ := s.List()
	if len(list) != 1 || list[0].ID != m.ID {
		t.Errorf("%+v", list)
	}
	name := "Renamed"
	th := Theme{BG: "#ffffff", FG: "#000000", ScalePct: 150, Speakers: "color"}
	m2, err := s.Update(m.ID, &name, &th)
	if err != nil || m2.Name != "Renamed" || m2.Theme.ScalePct != 150 || m2.Theme.Accent != "#ffffff" {
		t.Fatalf("%v %+v", err, m2)
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
	if _, err := s.ReadScript(m.ID, "../../etc"); err == nil {
		t.Error("language must be validated")
	}
}

func TestThemeValidation(t *testing.T) {
	if _, err := (Theme{Font: "dyslexic"}).Normalize(""); err != nil {
		t.Errorf("dyslexic font rejected: %v", err)
	}
	bad := []Theme{{BG: "red"}, {FG: "#12345"}, {Font: "comic"}, {ScalePct: 10}, {ScalePct: 900}, {Speakers: "loud"}}
	for _, th := range bad {
		if _, err := th.Normalize(""); !errors.Is(err, ErrBadTheme) {
			t.Errorf("%+v accepted", th)
		}
	}
	if th, err := (Theme{}).Normalize("Hi"); err != nil || th.IdleMessage != "Hi" {
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
	names := s.AssetNames(m.ID)
	if len(names) != 3 { // My Curtain (1).PNG, evil.png, hidden.png
		t.Errorf("%v", names)
	}
	if err := s.DeleteAsset(m.ID, "evil.png"); err != nil {
		t.Error(err)
	}
	if err := s.DeleteAsset(m.ID, "../show.yaml"); err == nil {
		t.Error("traversal accepted")
	}
	if _, err := s.SaveAsset("nope-0000", "a.png", bytes.NewReader(pngBytes(t, 8))); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown show: %v", err)
	}
}

func TestDuplicateCopiesEverything(t *testing.T) {
	s := newStore(t)
	m, _ := s.Create("Original", "Hello\nWorld\n")
	s.WriteScript(m.ID, "es", "Hola\n")
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
	if txt, _ := s.ReadScript(cp.ID, ""); txt != "Hello\nWorld\n" {
		t.Errorf("script %q", txt)
	}
	if txt, _ := s.ReadScript(cp.ID, "es"); txt != "Hola\n" {
		t.Errorf("translation %q", txt)
	}
	if names := s.AssetNames(cp.ID); len(names) != 1 || names[0] != "a.png" {
		t.Errorf("assets %v", names)
	}
	// Independent afterwards.
	s.WriteScript(cp.ID, "", "changed")
	if txt, _ := s.ReadScript(m.ID, ""); txt != "Hello\nWorld\n" {
		t.Error("original changed with the copy")
	}
	if c2, err := s.Duplicate(m.ID, "Mine"); err != nil || c2.Name != "Mine" {
		t.Errorf("named copy: %v %+v", err, c2)
	}
	if _, err := s.Duplicate("nope-0000", ""); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing show: %v", err)
	}
}
