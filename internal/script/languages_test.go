package script

import (
	"strings"
	"testing"
)

func cues(src string) []Cue { return Parse(src, Options{}).Cues }

func TestLanguageNames(t *testing.T) {
	cases := map[string]string{"es": "Español", "en": "English", "pt-br": "Português (BR)", "xx": "xx", "zh": "中文"}
	for code, want := range cases {
		if got := LanguageName(code); got != want {
			t.Errorf("%s: %q, want %q", code, got, want)
		}
	}
	for _, ok := range []string{"en", "es", "fil", "pt-br", "zh-hans"} {
		if !LangCodeRe.MatchString(ok) {
			t.Errorf("%s should be valid", ok)
		}
	}
	for _, bad := range []string{"", "E", "english", "../x", "es_ES", "es-", "1a"} {
		if LangCodeRe.MatchString(bad) {
			t.Errorf("%q should be invalid", bad)
		}
	}
}

func TestSameStructureDifferentWordsIsFine(t *testing.T) {
	en := cues("[loop: 5s]\nWelcome\n[image: a.png]\n[/loop]\n[section: Act 1]\nHELLO: Hi there\n[together]\nA: one\nB: two\n[/together]\n[clear]\n[fullscreen: Intermission]")
	es := cues("[loop: 5s]\nBienvenidos\n[image: b.png]\n[/loop]\n[section: Acto 1]\nHOLA: Hola\n[together]\nA: uno\nB: dos\n[/together]\n[clear]\n[fullscreen: Intermedio]")
	if d := CompareStructure(en, es, "English"); len(d) != 0 {
		t.Errorf("unexpected: %+v", d)
	}
}

func TestStructureMismatchesAreReportedWithTheTranslationsLine(t *testing.T) {
	en := cues("One\nTwo\n[clear]\nThree")
	cases := []struct{ name, src, want string }{
		{"different kind", "Uno\nDos\nTres\nCuatro", "Cue 3 does not match the English script: that script has [clear] here, and this one has a caption"},
		{"too short", "Uno\nDos\n[clear]", "has 3 cues but the English script has 4. Missing: a caption after cue 3"},
		{"too long", "Uno\nDos\n[clear]\nTres\nCuatro", "has 5 cues but the English script has 4. This line has no match there"},
		{"loop items", "Uno\nDos\n[loop]\nx\n[/loop]\nTres", "Cue 3"},
	}
	for _, c := range cases {
		d := CompareStructure(en, cues(c.src), "English")
		if len(d) == 0 || !strings.Contains(d[0].Message, c.want) || d[0].Severity != SevError {
			t.Errorf("%s: %+v", c.name, d)
		}
	}
	d := CompareStructure(en, cues("Uno\nDos\nTres\nCuatro"), "English")
	if d[0].Line != 3 {
		t.Errorf("should point at the translation's own line 3, got %d", d[0].Line)
	}
	// Loops must have the same number and kind of items; voices shown together the same number of voices.
	a := cues("[loop]\nx\n[image: a.png]\n[/loop]")
	b := cues("[loop]\nx\n[/loop]")
	if len(CompareStructure(a, b, "English")) == 0 {
		t.Error("different item counts should differ")
	}
	c := cues("[loop]\nx\n[blank: 3s]\n[/loop]")
	if len(CompareStructure(a, c, "English")) == 0 {
		t.Error("different item kinds should differ")
	}
	if len(CompareStructure(cues("[together]\nA: x\nB: y\n[/together]"), cues("A: x"), "English")) == 0 {
		t.Error("voices together vs one caption should differ")
	}
}
