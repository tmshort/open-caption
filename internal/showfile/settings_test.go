package showfile

import (
	"strings"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestSettingsBlockIsReadAndIsNotACue(t *testing.T) {
	src := `<!-- banner -->
[language: en | English]
[settings]
title: The Garden Party
background: #000
Text Colour: yellow
speaker-name-color: #FFAA00
font: Dyslexia friendly
text size: 120%
speaker names: different colors
idle message: Captions will begin shortly
loop seconds: 12s
scrollback lines: 300
[/settings]
Hello there
`
	r := Parse(src, Options{})
	if len(r.Cues) != 1 || r.Cues[0].Text != "Hello there" || r.Cues[0].Line != 15 {
		t.Fatalf("settings must not become cues: %+v", r.Cues)
	}
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("unexpected %+v", d)
		}
	}
	s := r.Settings
	if *s.Title != "The Garden Party" || *s.Background != "#000000" || *s.TextColor != "#ffee00" || *s.SpeakerColor != "#ffaa00" ||
		*s.Font != "dyslexic" || *s.TextSize != 120 || *s.SpeakerNames != "color" || *s.IdleMessage != "Captions will begin shortly" ||
		*s.LoopSeconds != 12 || *s.ScrollbackLines != 300 {
		t.Errorf("%+v", s)
	}
	if !s.LookSet() {
		t.Error("look is set")
	}
	if Parse("Just a caption", Options{}).Settings.LookSet() {
		t.Error("an empty settings struct sets no look")
	}
}

func TestSettingsProblemsAreReportedWithLinesAndTheRestStillWorks(t *testing.T) {
	src := "[settings]\nfont: comic sans\ncolour: red\nbackgroud: #000000\njust some words\ntext size: huge\nloop seconds: 0\nscrollback lines: 5\nbackground: nope\nfont: serif\nfont: mono\n[/settings]\nA"
	r := Parse(src, Options{})
	if len(r.Cues) != 1 {
		t.Fatalf("%+v", r.Cues)
	}
	byLine := map[int]string{}
	for _, d := range r.Diagnostics {
		byLine[d.Line] += d.Severity + ": " + d.Message + " | "
	}
	for line, want := range map[int]string{
		2: "is not a font", 3: "Unknown setting \"colour\"", 4: "Did you mean \"background\"?", 5: "Expected \"name: value\"",
		6: "not a text size", 7: "not a time", 8: "not a line count", 9: "is not a color", 11: "already set on line 10",
	} {
		if !strings.Contains(byLine[line], want) {
			t.Errorf("line %d: want %q in %q", line, want, byLine[line])
		}
	}
	if r.Settings.Font == nil || *r.Settings.Font != "mono" {
		t.Errorf("the last valid value wins: %+v", r.Settings.Font)
	}
	if r.Settings.TextSize != nil || r.Settings.Background != nil {
		t.Errorf("an invalid value must not be kept: %+v", r.Settings)
	}
}

func TestSettingsBlockProblems(t *testing.T) {
	r := Parse("[settings]\nfont: serif\nA", Options{})
	// An unclosed block swallows the rest of the file (the error says so); what was read is kept.
	if r.Settings.Font == nil || !hasDiag(r, SevError, "never closed") {
		t.Errorf("unclosed: %+v %+v", r.Cues, r.Diagnostics)
	}
	if r := Parse("[/settings]", Options{}); !hasDiag(r, SevWarning, "no matching [settings]") {
		t.Errorf("%+v", r.Diagnostics)
	}
	if r := Parse("[settings]\n[clear]\n[/settings]", Options{}); !hasDiag(r, SevWarning, "Only settings") || len(r.Cues) != 0 {
		t.Errorf("a command inside: %+v %+v", r.Cues, r.Diagnostics)
	}
	if r := Parse("[loop]\n[settings]\n[/loop]", Options{}); !hasDiag(r, SevWarning, "can't go inside a loop") {
		t.Errorf("inside a loop: %+v", r.Diagnostics)
	}
	if r := Parse("<!--\n[settings]\nfont: serif\n[/settings]\n-->\nA", Options{}); r.Settings.Font != nil || len(r.Cues) != 1 {
		t.Errorf("a commented-out block must be ignored: %+v", r.Settings)
	}
}

func TestOnlyTheIdleMessageIsSetInATranslation(t *testing.T) {
	src := "[language: es | Español]\n[settings]\ntitle: Ignored\nfont: serif\nidle message: Pronto\n[/settings]\nHola"
	r := Parse(src, Options{Translation: true})
	if r.Settings.Font != nil || r.Settings.Title != nil || r.Settings.IdleMessage == nil || *r.Settings.IdleMessage != "Pronto" {
		t.Errorf("%+v", r.Settings)
	}
	if !hasDiag(r, SevWarning, "set once for the whole show") {
		t.Errorf("the ignored look setting should be reported: %+v", r.Diagnostics)
	}
	// the title in a translation is quietly ignored (no complaint)
	for _, d := range r.Diagnostics {
		if strings.Contains(d.Message, "title") {
			t.Errorf("title should be silent in a translation: %+v", d)
		}
	}
}

func TestNormalizers(t *testing.T) {
	for in, want := range map[string]string{"#FFF": "#ffffff", "#aBc123": "#abc123", " White ": "#ffffff", "yellow": "#ffee00"} {
		if got, err := NormalizeColor(in); err != nil || got != want {
			t.Errorf("color %q: %q %v", in, got, err)
		}
	}
	for _, bad := range []string{"", "red-ish", "#12", "#gggggg", "rgb(1,2,3)"} {
		if _, err := NormalizeColor(bad); err == nil {
			t.Errorf("color %q accepted", bad)
		}
	}
	for in, want := range map[string]string{"System": "system", "mono": "mono", "OpenDyslexic": "dyslexic", "high legibility": "hyperlegible", "Atkinson Hyperlegible": "hyperlegible"} {
		if got, err := NormalizeFont(in); err != nil || got != want {
			t.Errorf("font %q: %q %v", in, got, err)
		}
	}
	if n, err := NormalizeTextSize("150 %"); err != nil || n != 150 {
		t.Errorf("%d %v", n, err)
	}
	for _, bad := range []string{"49", "301", "big", ""} {
		if _, err := NormalizeTextSize(bad); err == nil {
			t.Errorf("size %q accepted", bad)
		}
	}
}

func TestUpdateSettingsKeepsEverythingElse(t *testing.T) {
	src := "<!-- my banner -->\n[language: en | English]\n\n[settings]\nfont: serif\nmystery: value\n[/settings]\n\n[section: Act 1]\n<!-- a comment -->\nHello\n"
	out := UpdateSettings(src, false, func(s *Settings) { s.Font = ptr("dyslexic"); s.Title = ptr("Hamlet") })
	for _, want := range []string{"<!-- my banner -->", "[language: en | English]", "title: Hamlet", "font: dyslexic", "mystery: value", "[section: Act 1]", "<!-- a comment -->", "Hello"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if strings.Count(out, "[settings]") != 1 || strings.Contains(out, "font: serif") {
		t.Errorf("exactly one block with the new value:\n%s", out)
	}
	// the result still parses to the same cues and the new settings
	a, b := Parse(src, Options{}), Parse(out, Options{})
	if len(a.Cues) != len(b.Cues) || *b.Settings.Title != "Hamlet" || *b.Settings.Font != "dyslexic" {
		t.Errorf("%+v", b.Settings)
	}
	if len(b.Diagnostics) > 0 {
		for _, d := range b.Diagnostics {
			if d.Severity != SevInfo && !strings.Contains(d.Message, "mystery") {
				t.Errorf("%+v", d)
			}
		}
	}
	// idempotent
	if again := UpdateSettings(out, false, func(*Settings) {}); again != out {
		t.Errorf("a no-change update must not alter the file:\n%s\n---\n%s", out, again)
	}
}

func TestSettingsBlockIsInsertedAfterTheBannerAndLanguageLine(t *testing.T) {
	cases := map[string]string{
		"Hello":                               "[settings]\ntitle: T\n[/settings]\n\nHello",
		"":                                    "[settings]\ntitle: T\n[/settings]\n",
		"<!-- a\nb -->\n[language: es]\nHola": "<!-- a\nb -->\n[language: es]\n\n[settings]\ntitle: T\n[/settings]\n\nHola",
	}
	for in, want := range cases {
		if got := UpdateSettings(in, false, func(s *Settings) { s.Title = ptr("T") }); got != want {
			t.Errorf("%q:\n got %q\nwant %q", in, got, want)
		}
	}
	// an empty settings value removes the block entirely
	got := ReplaceSettings("[settings]\ntitle: T\n[/settings]\n\nHello", Settings{})
	if got != "Hello" {
		t.Errorf("%q", got)
	}
	// multiple blocks collapse into one
	multi := "[settings]\nfont: serif\n[/settings]\nA\n[settings]\ntext size: 90%\n[/settings]\nB"
	out := UpdateSettings(multi, false, func(s *Settings) {})
	r := Parse(out, Options{})
	if strings.Count(out, "[settings]") != 1 || *r.Settings.Font != "serif" || *r.Settings.TextSize != 90 || len(r.Cues) != 2 {
		t.Errorf("%s", out)
	}
}

func TestEveryGeneratedBlockRoundTrips(t *testing.T) {
	s := Settings{Title: ptr("A: title [with] stuff"), Background: ptr("#101010"), TextColor: ptr("#fafafa"), SpeakerColor: ptr("#ffcc00"),
		Font: ptr("hyperlegible"), TextSize: ptr(140), SpeakerNames: ptr("hidden"), IdleMessage: ptr("Soon: very soon"), LoopSeconds: ptr(7.5), ScrollbackLines: ptr(500)}
	text := ReplaceSettings("Line", s)
	r := Parse(text, Options{})
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("%+v", d)
		}
	}
	g := r.Settings
	if *g.Title != *s.Title || *g.Background != *s.Background || *g.Font != *s.Font || *g.TextSize != 140 || *g.IdleMessage != "Soon: very soon" || *g.LoopSeconds != 7.5 || *g.ScrollbackLines != 500 || *g.SpeakerNames != "hidden" {
		t.Errorf("%+v\n%s", g, text)
	}
}

func TestSettingNamesAreStable(t *testing.T) {
	names := strings.Join(SettingNames(), ",")
	if names != "title,background,text color,speaker name color,font,text size,speaker names,idle message,loop seconds,scrollback lines" {
		t.Errorf("%s", names)
	}
	for _, n := range SettingNames() {
		if SettingHelp(n) == "" {
			t.Errorf("%s has no description", n)
		}
	}
}
