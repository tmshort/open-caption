// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"strings"
	"testing"
)

func kinds(r Result) []CueKind {
	var k []CueKind
	for _, c := range r.Cues {
		k = append(k, c.Kind)
	}
	return k
}

func hasDiag(r Result, sev, sub string) bool {
	for _, d := range r.Diagnostics {
		if d.Severity == sev && strings.Contains(d.Message, sub) {
			return true
		}
	}
	return false
}

func TestPlainTextIsCaptions(t *testing.T) {
	r := Parse("Hello there.\r\n\r\nHow are you?\n", Options{})
	if len(r.Cues) != 2 || r.Cues[0].Text != "Hello there." || r.Cues[1].Index != 1 {
		t.Fatalf("got %+v", r.Cues)
	}
	if r.Cues[0].Line != 1 || r.Cues[1].Line != 3 {
		t.Errorf("lines: %d %d", r.Cues[0].Line, r.Cues[1].Line)
	}
}

func TestBOMAndLineEndings(t *testing.T) {
	r := Parse("\uFEFFone\rtwo\r\nthree", Options{})
	if len(r.Cues) != 3 || r.Cues[0].Text != "one" {
		t.Fatalf("got %+v", r.Cues)
	}
}

func TestCommentsNeverReachCues(t *testing.T) {
	src := "<!-- title -->\nA\n<!-- multi\nline -->\nB <!-- inline --> C\n<!-- gone -->\n"
	r := Parse(src, Options{})
	if len(r.Cues) != 2 || r.Cues[0].Text != "A" || r.Cues[1].Text != "B  C" {
		t.Fatalf("got %+v", r.Cues)
	}
	if r.Cues[1].Line != 5 {
		t.Errorf("line numbers drifted: %d", r.Cues[1].Line)
	}
}

func TestUnclosedComment(t *testing.T) {
	r := Parse("A\n<!-- oops\nB\n", Options{})
	if len(r.Cues) != 1 || !hasDiag(r, SevError, "never closed") {
		t.Fatalf("got %+v %+v", r.Cues, r.Diagnostics)
	}
	for _, d := range r.Diagnostics {
		if d.Severity == SevError && d.Line != 2 {
			t.Errorf("error line = %d, want 2", d.Line)
		}
	}
}

func TestHeadingsReserved(t *testing.T) {
	r := Parse("# Title\nA\n\\# literal\n", Options{})
	if len(r.Cues) != 2 || r.Cues[1].Text != "# literal" {
		t.Fatalf("got %+v", r.Cues)
	}
	if !hasDiag(r, SevInfo, "reserved for headings") {
		t.Error("missing heading diagnostic")
	}
}

func TestSpeakerWithNothingAfterTheColonIsWarnedAbout(t *testing.T) {
	for _, in := range []string{"BARRY:", "DEE DEE & BARRY: ", "ALL (EXCEPT BARRY):"} {
		r := Parse(in, Options{})
		if !hasDiag(r, SevWarning, "nothing after the colon") {
			t.Errorf("%q: no warning: %+v", in, r.Diagnostics)
		}
	}
	for _, in := range []string{"BARRY: Hi", "BARRY: ~", "Note:", "BARRY\\:", "TIME: 10:30"} {
		if r := Parse(in, Options{}); hasDiag(r, SevWarning, "nothing after the colon") {
			t.Errorf("%q: needless warning", in)
		}
	}
	// the line is still a caption, as typed, and the warning names its line
	r := Parse("A: x\n\nBARRY:\n~ la la", Options{})
	if len(r.Cues) != 3 || r.Cues[1].Text != "BARRY:" {
		t.Errorf("cues: %+v", r.Cues)
	}
	for _, d := range r.Diagnostics {
		if strings.Contains(d.Message, "nothing after the colon") && d.Line != 3 {
			t.Errorf("line %d, want 3", d.Line)
		}
	}
}

func TestSpeaker(t *testing.T) {
	cases := []struct{ in, speaker, text string }{
		{"HAMLET: To be", "HAMLET", "To be"},
		{"MRS. O'NEIL: Hi", "MRS. O'NEIL", "Hi"},
		{"Note: this", "", "Note: this"},
		{"HAMLET\\: escaped", "", "HAMLET: escaped"},
		{"TIME: 10:30", "TIME", "10:30"},
		{"DEE DEE & BARRY: Yes!", "DEE DEE & BARRY", "Yes!"},
		{"BARRY, ANGIE, SHELDON & TRENT: ~ Not, not not", "BARRY, ANGIE, SHELDON & TRENT", "♪ Not, not not ♪"},
		{"ALL (EXCEPT BARRY): That's right!", "ALL (EXCEPT BARRY)", "That's right!"},
		{"PARENT #1: No.", "PARENT #1", "No."},
		{"KAYLEE & SHELBY / FEMALE STUDENTS: Hi", "KAYLEE & SHELBY / FEMALE STUDENTS", "Hi"},
		{"THIRD MALE STUDENT & FIFTH FEMALE STUDENT: Hi", "THIRD MALE STUDENT & FIFTH FEMALE STUDENT", "Hi"},
		{"(SFX): thunder", "", "(SFX): thunder"},
		{"A LABEL THAT IS FAR TOO LONG TO BE A SPEAKER NAME, SO IT IS NOT ONE AT ALL: x", "", "A LABEL THAT IS FAR TOO LONG TO BE A SPEAKER NAME, SO IT IS NOT ONE AT ALL: x"},
	}
	for _, c := range cases {
		r := Parse(c.in, Options{})
		if len(r.Cues) != 1 || r.Cues[0].Speaker != c.speaker || r.Cues[0].Text != c.text {
			t.Errorf("%q: got %+v", c.in, r.Cues)
		}
	}
}

func TestInlineMarkdownAndEscaping(t *testing.T) {
	cases := []struct{ in, html string }{
		{"He's **always** late", "He&#39;s <strong>always</strong> late"},
		{"an *aside* and _this_", "an <em>aside</em> and <em>this</em>"},
		{"snake_case_name stays", "snake_case_name stays"},
		{"2 * 3 * 4", "2 * 3 * 4"},
		{"**bold *nested* bold**", "<strong>bold <em>nested</em> bold</strong>"},
		{`literal \*stars\*`, "literal *stars*"},
		{"<script>alert(1)</script>", "&lt;script&gt;alert(1)&lt;/script&gt;"},
		{"[link](javascript:alert(1))", "[link](javascript:alert(1))"},
		{"unclosed **bold", "unclosed **bold"},
	}
	for _, c := range cases {
		h, _ := RenderInline(c.in)
		if h != c.html {
			t.Errorf("%q:\n got %q\nwant %q", c.in, h, c.html)
		}
	}
}

func TestHTMLShownLiterallyWithInfo(t *testing.T) {
	r := Parse("a <b>bold</b> move", Options{})
	if !strings.Contains(r.Cues[0].HTML, "&lt;b&gt;") || !hasDiag(r, SevInfo, "HTML is not supported") {
		t.Fatalf("%+v %+v", r.Cues, r.Diagnostics)
	}
}

func TestStageDirection(t *testing.T) {
	r := Parse("(thunder rumbles)\nBOB: (aside) hi", Options{})
	if !r.Cues[0].Stage || r.Cues[1].Stage {
		t.Fatalf("%+v", r.Cues)
	}
}

func TestMultiLineCaptions(t *testing.T) {
	r := Parse("[[\nRow, row,\nrow your boat\n]]\nline one \\\nline two\nlast", Options{})
	if len(r.Cues) != 3 {
		t.Fatalf("%+v", r.Cues)
	}
	if r.Cues[0].Text != "Row, row,\nrow your boat" || r.Cues[0].HTML != "Row, row,<br>row your boat" {
		t.Errorf("block: %+v", r.Cues[0])
	}
	if r.Cues[1].Text != "line one\nline two" || r.Cues[2].Text != "last" {
		t.Errorf("continuation: %+v %+v", r.Cues[1], r.Cues[2])
	}
	r = Parse("[[\nnever closed\n", Options{})
	if len(r.Cues) != 0 || !hasDiag(r, SevError, "never closed") {
		t.Errorf("%+v", r)
	}
}

func TestDirectives(t *testing.T) {
	src := "[Clear]\n[CLEAR SCREEN]\n[blank]\n[image: curtain.jpg]\n[image: set.jpg | Act 1]\n![The Garden](g.png)\n" +
		"[fullscreen: **INTERMISSION**]\n[section: Act 2]\n[note: wait for door]\n[language: es | Español]\n"
	r := Parse(src, Options{})
	want := []CueKind{KindClear, KindClear, KindBlank, KindImage, KindImage, KindImage, KindFullscreen, KindSection, KindNote}
	if got := kinds(r); strings.Join(strs(got), ",") != strings.Join(strs(want), ",") {
		t.Fatalf("got %v want %v (%+v)", got, want, r.Diagnostics)
	}
	if r.Cues[4].Item.Caption != "Act 1" || r.Cues[5].Item.Image != "g.png" || r.Cues[5].Item.Caption != "The Garden" {
		t.Errorf("image items: %+v %+v", r.Cues[4].Item, r.Cues[5].Item)
	}
	if r.Cues[6].Item.HTML != "<strong>INTERMISSION</strong>" {
		t.Errorf("fullscreen: %+v", r.Cues[6].Item)
	}
	if r.Language != "es" || r.LanguageName != "Español" {
		t.Errorf("language %q %q", r.Language, r.LanguageName)
	}
}

func strs(k []CueKind) []string {
	var s []string
	for _, x := range k {
		s = append(s, string(x))
	}
	return s
}

func TestUnknownDirective(t *testing.T) {
	r := Parse("[fade]\nok", Options{})
	if len(r.Cues) != 1 || !hasDiag(r, SevWarning, "Unknown command [fade]") {
		t.Fatalf("%+v", r)
	}
}

func TestFullscreenBlock(t *testing.T) {
	r := Parse("[fullscreen]\nThanks for\ncoming!\n[/fullscreen]\n[fullscreen]\nnever closed", Options{})
	if len(r.Cues) != 1 || r.Cues[0].Item.HTML != "Thanks for<br>coming!" || !hasDiag(r, SevError, "never closed") {
		t.Fatalf("%+v %+v", r.Cues, r.Diagnostics)
	}
}

func TestLoop(t *testing.T) {
	src := "[loop: 12s]\n[image: sponsors.png | 15s]\nWelcome to **The Show**\n![cast](cast.jpg)\n[fullscreen: Hello | 5s]\n[blank: 3s]\n[/loop]\nHi"
	r := Parse(src, Options{})
	if len(r.Cues) != 2 || r.Cues[0].Kind != KindLoop {
		t.Fatalf("%+v %+v", r.Cues, r.Diagnostics)
	}
	l := r.Cues[0]
	if l.Period != 12 || len(l.Items) != 5 {
		t.Fatalf("%+v", l)
	}
	checks := []struct {
		kind string
		secs float64
	}{{"image", 15}, {"text", 0}, {"image", 0}, {"text", 5}, {"blank", 3}}
	for i, c := range checks {
		if l.Items[i].Kind != c.kind || l.Items[i].Seconds != c.secs {
			t.Errorf("item %d = %+v, want %v", i, l.Items[i], c)
		}
	}
	if l.Items[1].HTML != "Welcome to <strong>The Show</strong>" {
		t.Errorf("text item html: %q", l.Items[1].HTML)
	}
}

func TestLoopProblems(t *testing.T) {
	r := Parse("[loop]\n[clear]\n[section: x]\n[loop]\nok\n[/loop]", Options{})
	if len(r.Cues) != 1 || len(r.Cues[0].Items) != 1 {
		t.Fatalf("%+v %+v", r.Cues, r.Diagnostics)
	}
	if !hasDiag(r, SevWarning, "[clear] can't go inside a loop") || !hasDiag(r, SevWarning, "nested") {
		t.Errorf("%+v", r.Diagnostics)
	}
	r = Parse("[loop]\nitem\n", Options{})
	if len(r.Cues) != 0 || !hasDiag(r, SevError, "[loop] was opened but never closed") {
		t.Errorf("%+v", r)
	}
	r = Parse("[/loop]", Options{})
	if !hasDiag(r, SevWarning, "no matching") {
		t.Errorf("%+v", r.Diagnostics)
	}
}

func TestAssetChecks(t *testing.T) {
	assets := []string{"curtain.jpg", "set.png"}
	r := Parse("[image: curtian.jpg]\n[image: nothing-like-it.gif]\n[image: set.png]", Options{Assets: assets})
	if !hasDiag(r, SevWarning, `Did you mean "curtain.jpg"`) || !hasDiag(r, SevWarning, "Upload it") {
		t.Errorf("%+v", r.Diagnostics)
	}
	if len(r.Cues) != 3 {
		t.Errorf("cues should still be kept: %d", len(r.Cues))
	}
	r = Parse("[image: whatever.jpg]", Options{})
	for _, d := range r.Diagnostics {
		if d.Severity == SevWarning {
			t.Errorf("assets unchecked when nil: %+v", d)
		}
	}
}

func TestExample(t *testing.T) {
	src := `<!-- The Odd Couple -->
[loop: 10s]
[image: sponsors.png | 15s]
Welcome to **The Odd Couple**
[/loop]

[section: Act 1, Scene 1]
[clear]
(Poker night.)
SPEED: Where's Felix?
[note: Felix enters stage left]
[[
FELIX: I'm so sorry,
I got lost.
]]
`
	r := Parse(src, Options{})
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("unexpected %+v", d)
		}
	}
	if got := len(r.Cues); got != 7 {
		t.Errorf("cues = %d: %v", got, kinds(r))
	}
}

func FuzzParse(f *testing.F) {
	for _, s := range []string{"a", "[loop]\nx", "<!-- x", "[[\n", "**", "![a](b)", "[image: |]", "\\", "[fullscreen]"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		r := Parse(s, Options{})
		for _, c := range r.Cues {
			if strings.Contains(c.HTML, "<script") || strings.Contains(c.HTML, "<img") {
				t.Fatalf("unescaped html: %q", c.HTML)
			}
		}
	})
}

// Every registered directive must be understood by the parser.
func TestDirectiveRegistryIsHandled(t *testing.T) {
	for _, name := range DirectiveNames {
		for _, src := range []string{"[" + name + "]", "[" + name + ": x]"} {
			r := Parse(src, Options{})
			if hasDiag(r, SevWarning, "Unknown command") {
				t.Errorf("%s: parser does not know directive %q", src, name)
			}
		}
	}
}

func TestSingShorthand(t *testing.T) {
	cases := []struct {
		in, text, html string
		sung           bool
	}{
		{"~ Row, row, row your boat ~", "♪ Row, row, row your boat ♪", "♪ Row, row, row your boat ♪", true},
		{"~ no closing marker", "♪ no closing marker ♪", "", true},
		{"~", "♪", "♪", true},
		{"~ ~", "♪", "♪", true},
		{"~ *Let* it go ~", "♪ Let it go ♪", "♪ <em>Let</em> it go ♪", true},
		{"about ~5 minutes", "about ~5 minutes", "", false},
		{"~5 minutes", "~5 minutes", "", false},
		{`\~ not a lyric`, "~ not a lyric", "", false},
		{"♪ typed directly ♪", "♪ typed directly ♪", "", false},
	}
	for _, c := range cases {
		r := Parse(c.in, Options{})
		if len(r.Cues) != 1 || r.Cues[0].Text != c.text || r.Cues[0].Sung != c.sung || (c.html != "" && r.Cues[0].HTML != c.html) {
			t.Errorf("%q: got %+v", c.in, r.Cues)
		}
	}
	r := Parse("ELSA: ~ Let it go ~", Options{})
	if r.Cues[0].Speaker != "ELSA" || !r.Cues[0].Sung || r.Cues[0].Text != "♪ Let it go ♪" {
		t.Errorf("speaker + lyric: %+v", r.Cues[0])
	}
	r = Parse("[[\n~ Row, row,\nrow your boat ~\n]]", Options{})
	if r.Cues[0].Text != "♪ Row, row, ♪\n♪ row your boat ♪" || !r.Cues[0].Sung {
		t.Errorf("multi-line lyric: %q", r.Cues[0].Text)
	}
	if r := Parse("(~ hums ~)", Options{}); r.Cues[0].Sung {
		t.Error("a parenthetical is not a lyric")
	}
}

func TestSongBlock(t *testing.T) {
	src := "[song: Memory]\nMidnight, not a sound\n\nGRIZABELLA: Has the moon lost her memory?\n[note: hold the last chord]\n[/song]\nAfter the song"
	r := Parse(src, Options{})
	if got := kinds(r); strings.Join(strs(got), ",") != "section,caption,caption,note,caption" {
		t.Fatalf("kinds %v %+v", got, r.Diagnostics)
	}
	if r.Cues[0].Text != "♪ Memory" || !r.Cues[1].Sung || r.Cues[1].Text != "♪ Midnight, not a sound ♪" ||
		!r.Cues[2].Sung || r.Cues[2].Speaker != "GRIZABELLA" || r.Cues[4].Sung || r.Cues[4].Text != "After the song" {
		t.Errorf("%+v", r.Cues)
	}
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("unexpected %+v", d)
		}
	}
	// untitled song, closing without opening, nesting, unclosed (lyrics are kept)
	if r := Parse("[song]\nla\n[/song]", Options{}); len(r.Cues) != 1 || !r.Cues[0].Sung {
		t.Errorf("untitled: %+v", r.Cues)
	}
	if r := Parse("[/song]", Options{}); !hasDiag(r, SevWarning, "no matching [song]") {
		t.Errorf("%+v", r.Diagnostics)
	}
	if r := Parse("[song]\n[song]\nla\n[/song]", Options{}); !hasDiag(r, SevWarning, "inside another song") || len(r.Cues) != 1 {
		t.Errorf("%+v", r)
	}
	r = Parse("[song: X]\nla la\nli li", Options{})
	if !hasDiag(r, SevError, "[song] was opened but never closed") || len(r.Cues) != 3 {
		t.Errorf("unclosed: %+v %+v", r.Cues, r.Diagnostics)
	}
	if r := Parse("[loop]\n[song]\n[/loop]", Options{}); !hasDiag(r, SevWarning, "[song] can't go inside a loop") {
		t.Errorf("song in loop: %+v", r.Diagnostics)
	}
}

func TestTogetherBlockShowsVoicesAtOnce(t *testing.T) {
	src := "[together]\nANNA: ~ I love you ~\nKRIS: ~ Me too, me too ~\n[/together]\nNext line"
	r := Parse(src, Options{})
	if len(r.Cues) != 2 {
		t.Fatalf("one cue for both voices expected: %+v %+v", r.Cues, r.Diagnostics)
	}
	c := r.Cues[0]
	if len(c.Parts) != 2 || c.Parts[0].Speaker != "ANNA" || c.Parts[1].Speaker != "KRIS" || !c.Parts[0].Sung || !c.Parts[1].Sung ||
		c.Parts[1].Text != "♪ Me too, me too ♪" || !c.Sung {
		t.Errorf("%+v", c)
	}
	if c.Text != "ANNA: ♪ I love you ♪\nKRIS: ♪ Me too, me too ♪" || c.Line != 1 {
		t.Errorf("text/line: %q %d", c.Text, c.Line)
	}
	if r.Cues[1].Text != "Next line" || len(r.Cues[1].Parts) != 0 {
		t.Errorf("%+v", r.Cues[1])
	}
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("unexpected %+v", d)
		}
	}
}

func TestTogetherMixesSpokenAndSungAndAliases(t *testing.T) {
	r := Parse("[duet]\nMARY: (spoken) Hello\nJOE: ~ la la ~\n[/duet]", Options{})
	if len(r.Cues) != 1 || len(r.Cues[0].Parts) != 2 || r.Cues[0].Parts[0].Sung || !r.Cues[0].Parts[1].Sung || r.Cues[0].Sung {
		t.Fatalf("%+v %+v", r.Cues, r.Diagnostics)
	}
	// Inside a song every voice is sung.
	r = Parse("[song]\n[together]\nA: one\nB: two\n[/together]\n[/song]", Options{})
	if len(r.Cues) != 1 || !r.Cues[0].Sung || r.Cues[0].Parts[1].Text != "♪ two ♪" {
		t.Errorf("%+v %+v", r.Cues, r.Diagnostics)
	}
}

func TestVoiceSeparatorOnOneLine(t *testing.T) {
	r := Parse("ANNA: ~ I love you ~ || KRIS: ~ Me too ~ || BOB: Hello", Options{})
	if len(r.Cues) != 1 || len(r.Cues[0].Parts) != 3 || r.Cues[0].Parts[2].Speaker != "BOB" || r.Cues[0].Parts[2].Sung {
		t.Fatalf("%+v", r.Cues)
	}
	for _, in := range []string{"a||b", "a | | b", "this || ", "|| that"} {
		r := Parse(in, Options{})
		if len(r.Cues) != 1 || len(r.Cues[0].Parts) != 0 {
			t.Errorf("%q should stay one ordinary caption: %+v", in, r.Cues)
		}
	}
}

func TestTogetherProblems(t *testing.T) {
	r := Parse("[together]\nonly one\n[/together]", Options{})
	if len(r.Cues) != 1 || len(r.Cues[0].Parts) != 0 || !hasDiag(r, SevWarning, "at least two lines") {
		t.Errorf("single: %+v %+v", r.Cues, r.Diagnostics)
	}
	if r := Parse("[together]\n[/together]", Options{}); len(r.Cues) != 0 || !hasDiag(r, SevWarning, "empty") {
		t.Errorf("empty: %+v", r.Diagnostics)
	}
	r = Parse("[together]\nA: one\nB: two", Options{})
	if len(r.Cues) != 1 || len(r.Cues[0].Parts) != 2 || !hasDiag(r, SevError, "never closed") {
		t.Errorf("unclosed keeps the lines: %+v %+v", r.Cues, r.Diagnostics)
	}
	if r := Parse("[/together]", Options{}); !hasDiag(r, SevWarning, "no matching [together]") {
		t.Errorf("%+v", r.Diagnostics)
	}
	r = Parse("[together]\nA: one\n[clear]\nB: two\n[/together]", Options{})
	if !hasDiag(r, SevWarning, "[clear] can't go inside [together]") || len(r.Cues) != 1 || len(r.Cues[0].Parts) != 2 {
		t.Errorf("directive inside: %+v %+v", r.Cues, r.Diagnostics)
	}
	if r := Parse("[loop]\n[together]\n[/loop]", Options{}); !hasDiag(r, SevWarning, "can't go inside a loop") {
		t.Errorf("loop: %+v", r.Diagnostics)
	}
}

func TestCounterpointAlias(t *testing.T) {
	r := Parse("[counterpoint]\nA: x\nB: y\n[/counterpoint]", Options{})
	if len(r.Cues) != 1 || len(r.Cues[0].Parts) != 2 {
		t.Errorf("%+v %+v", r.Cues, r.Diagnostics)
	}
}

func TestSpeakersInOrderOfFirstAppearance(t *testing.T) {
	r := Parse("B: one\nA: two\nB: again\n[together]\nC: x\nA: y\n[/together]\nD: ~ la ~ || E: ~ la ~", Options{})
	if got := strings.Join(r.Speakers, ","); got != "B,A,C,D,E" {
		t.Errorf("speakers %q", got)
	}
	if r := Parse("no speakers here", Options{}); r.Speakers == nil || len(r.Speakers) != 0 {
		t.Errorf("want an empty, non-null list: %#v", r.Speakers)
	}
}

func TestAMissingImageSaysWhereItWasLookedForOnlyWhenTheCallerNamesAFolder(t *testing.T) {
	src := "[image: curtain.jpg]\n"
	msg := func(o Options) string {
		for _, d := range Parse(src, o).Diagnostics {
			if strings.Contains(d.Message, "curtain.jpg") {
				return d.Message
			}
		}
		return ""
	}
	// Setup (no folder named): the advice is for the Images step
	if got := msg(Options{Assets: []string{"logo.png"}}); got != `Image "curtain.jpg" not found. Upload it in the Images step.` {
		t.Errorf("%q", got)
	}
	// the command line: the folder that was searched
	if got := msg(Options{Assets: []string{"logo.png"}, ImageFolder: "/shows/x/assets"}); got != `Image "curtain.jpg" not found in /shows/x/assets.` {
		t.Errorf("%q", got)
	}
	// a near miss still gets the suggestion, in both
	if got := msg(Options{Assets: []string{"curtain.png"}, ImageFolder: "/f"}); !strings.Contains(got, `Did you mean "curtain.png"?`) {
		t.Errorf("%q", got)
	}
}

// A loop can be given a name, which the audience report uses.
func TestLoopNameOption(t *testing.T) {
	for in, want := range map[string]string{
		"[loop: 10s, name Intermission]":            "Intermission",
		"[loop: name Act 1   break, 12s]":           "Act 1 break",
		"[loop: NAME: Pre-show, timer 7:30pm, 10s]": "Pre-show",
		"[loop: timer 15 minutes, name Interval]":   "Interval",
		"[loop: 10s]": "",
	} {
		body := "Hello"
		if strings.Contains(in, "timer") {
			body += " {timer}" // a loop with a timer shows it somewhere
		}
		r := Parse(in+"\n"+body+"\n[/loop]\nA line\n", Options{})
		if len(r.Cues) == 0 || r.Cues[0].Kind != KindLoop || r.Cues[0].Name != want {
			t.Errorf("%q: %+v", in, r.Cues)
			continue
		}
		for _, d := range r.Diagnostics {
			if d.Severity != SevInfo {
				t.Errorf("%q: unexpected %s: %s", in, d.Severity, d.Message)
			}
		}
	}
	// The name does not disturb the timer or the item time.
	r := Parse("[loop: 12s, name Break, timer 15 minutes]\n{timer}\n[/loop]\n", Options{})
	if c := r.Cues[0]; c.Name != "Break" || c.Period != 12 || c.Timer == nil || c.Timer.Minutes != 15 {
		t.Errorf("%+v", c)
	}
	// An empty or too long name is a warning, and the loop is still a loop.
	problems := func(r Result) []Diagnostic {
		var out []Diagnostic
		for _, d := range r.Diagnostics {
			if d.Severity != SevInfo {
				out = append(out, d)
			}
		}
		return out
	}
	r = Parse("[loop: name]\nHi\n[/loop]\n", Options{})
	if d := problems(r); r.Cues[0].Name != "" || len(d) != 1 || !strings.Contains(d[0].Message, "loop name is empty") {
		t.Errorf("%+v %+v", r.Cues[0], d)
	}
	long := strings.Repeat("x", MaxLoopNameLength+10)
	r = Parse("[loop: name "+long+"]\nHi\n[/loop]\n", Options{})
	if d := problems(r); len([]rune(r.Cues[0].Name)) != MaxLoopNameLength || len(d) != 1 || !strings.Contains(d[0].Message, "was cut") {
		t.Errorf("%q %+v", r.Cues[0].Name, d)
	}
}

// A translation may name its loops in its own words, or not at all: only the main language's name is used.
func TestLoopNamesDoNotHaveToMatchInATranslation(t *testing.T) {
	main := Parse("[loop: 10s, name Intermission]\nBack soon\n[/loop]\nHello\n", Options{})
	tr := Parse("[loop: 10s, name Entracte]\nÀ tout à l'heure\n[/loop]\nBonjour\n", Options{})
	if errs := CompareStructure(main.Cues, tr.Cues, "English"); len(errs) != 0 {
		t.Errorf("%v", errs)
	}
}

func TestHugeItemTimeIsCutWithAWarning(t *testing.T) {
	r := Parse("[loop: 99999999999s]\nHello | 99999999999s\n[blank: 7200s]\n[/loop]\n", Options{})
	if len(r.Cues) != 1 {
		t.Fatalf("cues: %+v", r.Cues)
	}
	c := r.Cues[0]
	if c.Period != MaxItemSeconds || c.Items[0].Seconds != MaxItemSeconds || c.Items[1].Seconds != MaxItemSeconds {
		t.Errorf("times not cut: period %v, items %v %v", c.Period, c.Items[0].Seconds, c.Items[1].Seconds)
	}
	if !hasDiag(r, SevWarning, "cut to the most allowed") {
		t.Errorf("no warning: %+v", r.Diagnostics)
	}
}
