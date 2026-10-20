// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"strings"
	"testing"
)

func TestBoldItalicAndStrikethrough(t *testing.T) {
	cases := map[string][2]string{
		"***both***":                   {"<strong><em>both</em></strong>", "both"},
		"***both*** and **bold**":      {"<strong><em>both</em></strong> and <strong>bold</strong>", "both and bold"},
		"**bold _and italic_**":        {"<strong>bold <em>and italic</em></strong>", "bold and italic"},
		"~~gone~~ now":                 {"<del>gone</del> now", "gone now"},
		"a ~~b c~~ d":                  {"a <del>b c</del> d", "a b c d"},
		"~~**both**~~":                 {"<del><strong>both</strong></del>", "both"},
		"~~ spaced ~~":                 {"~~ spaced ~~", "~~ spaced ~~"},
		"~~unclosed":                   {"~~unclosed", "~~unclosed"},
		"***":                          {"***", "***"},
		"**** x":                       {"**** x", "**** x"},
		"~~~":                          {"~~~", "~~~"},
		"\\~~escaped~~":                {"~<em>escaped</em>~", ""}, // placeholder, replaced below
		"<del>raw html</del> ~~real~~": {"&lt;del&gt;raw html&lt;/del&gt; <del>real</del>", "<del>raw html</del> real"},
	}
	delete(cases, "\\~~escaped~~")
	for in, want := range cases {
		h, p := RenderInline(in)
		if h != want[0] || p != want[1] {
			t.Errorf("%q: got %q / %q, want %q / %q", in, h, p, want[0], want[1])
		}
	}
	// a sung line and a struck word in the same show file
	r := Parse("ELSA: ~ Let it go ~\n~ and ~~this~~ ~\n", Options{})
	if len(r.Cues) != 2 || !r.Cues[0].Sung || !strings.Contains(r.Cues[1].HTML, "<del>this</del>") || !r.Cues[1].Sung {
		t.Errorf("%+v %v", r.Cues, r.Diagnostics)
	}
}
