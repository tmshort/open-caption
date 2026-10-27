// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package show file parses open-caption show files (a Markdown subset, see
// docs/03-show-file-format.md) into cues. It is pure: no I/O, no network.
package showfile

// CueKind identifies what a cue does when the operator sends it.
type CueKind string

const (
	KindCaption    CueKind = "caption"
	KindClear      CueKind = "clear"
	KindBlank      CueKind = "blank"
	KindImage      CueKind = "image"
	KindFullscreen CueKind = "fullscreen"
	KindLoop       CueKind = "loop"
	KindSection    CueKind = "section" // operator-only
	KindNote       CueKind = "note"    // operator-only
)

// OperatorOnly reports whether the cue has no audience effect.
func (k CueKind) OperatorOnly() bool { return k == KindSection || k == KindNote }

// Item is one fullscreen item: used by image/fullscreen cues and inside loops.
type Item struct {
	Kind    string  `json:"kind"`              // "text" | "image" | "blank"
	HTML    string  `json:"html,omitempty"`    // generated, escaped HTML (text)
	Text    string  `json:"text,omitempty"`    // plain text
	Image   string  `json:"image,omitempty"`   // asset filename (image)
	Caption string  `json:"caption,omitempty"` // optional caption under an image
	Seconds float64 `json:"seconds,omitempty"` // display time inside a loop; 0 = default
	// Announcement marks the operator's one-off notice (Engine.Message), so the operator page can say one is on the screens.
	Announcement bool `json:"announcement,omitempty"`
	// Timer is set on a text item of a loop that has a timer, when its text contains {timer}: HTML then holds
	// a placeholder the page fills with the time left. Deadline (unix ms) is set by the live engine when the
	// loop starts, never by the parser.
	Timer    bool  `json:"timer,omitempty"`
	Deadline int64 `json:"deadline,omitempty"`
}

// Part is one voice of a caption sung or spoken together with others.
type Part struct {
	Speaker string `json:"speaker,omitempty"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
	Stage   bool   `json:"stage,omitempty"`
	Sung    bool   `json:"sung,omitempty"`
}

// Cue is one step the operator can send.
type Cue struct {
	Index   int     `json:"index"`
	Kind    CueKind `json:"kind"`
	Line    int     `json:"line"` // 1-based source line where the cue starts
	Speaker string  `json:"speaker,omitempty"`
	HTML    string  `json:"html,omitempty"` // caption: generated, escaped HTML
	Text    string  `json:"text,omitempty"` // plain text (operator display, search)
	Stage   bool    `json:"stage,omitempty"`
	Sung    bool    `json:"sung,omitempty"`   // a lyric: shown between musical notes
	Parts   []Part  `json:"parts,omitempty"`  // caption: two or more voices shown at the same time
	Item    *Item   `json:"item,omitempty"`   // image / fullscreen
	Items   []Item  `json:"items,omitempty"`  // loop
	Period  float64 `json:"period,omitempty"` // loop default seconds; 0 = config default
	Timer   *Timer  `json:"timer,omitempty"`  // loop: a countdown for {timer} (see timer.go)
	Name    string  `json:"name,omitempty"`   // loop: what the author calls it (name Intermission); used in the audience report
	Hint    float64 `json:"hint,omitempty"`   // seconds the line usually stays up, from [pace: 8s] before it: a reminder for the operator only
}

// MaxLoopNameLength is the longest name a loop can have, in characters.
const MaxLoopNameLength = 40

// Diagnostic is a problem or note found while parsing.
type Diagnostic struct {
	Line     int    `json:"line"`
	Severity string `json:"severity"` // "error" | "warning" | "info"
	Message  string `json:"message"`
}

// Severities.
const (
	SevError   = "error"
	SevWarning = "warning"
	SevInfo    = "info"
)

// Result is the outcome of Parse.
type Result struct {
	// Speakers lists every speaker name in order of first appearance. Devices use the
	// position in this list to give each character a stable color.
	Speakers []string `json:"speakers"`
	// Characters are the individual names behind the speaker labels, with the lines they speak on (most used first).
	Characters   []Character  `json:"characters"`
	Language     string       `json:"language"`
	LanguageName string       `json:"languageName,omitempty"`
	Cues         []Cue        `json:"cues"`
	Diagnostics  []Diagnostic `json:"diagnostics"`
	// Settings are the show-wide options written in a [settings] block.
	Settings Settings `json:"settings"`
}

// Options control parsing.
type Options struct {
	// Assets lists the uploaded image filenames. When non-nil, image references
	// are checked against it. When nil, references are not checked.
	Assets []string
	// ImageFolder names the folder the images are checked against, for the command line (open-caption --check-show).
	// Setup leaves it empty, and a missing image then says to upload it in the Images step.
	ImageFolder string
	// Translation is set for a translation's show file: only the pre-show message may be set there; the
	// look and the rest are set once, in the main show file.
	Translation bool
}

// SpeakersOf returns the speaker names used by cues, in order of first appearance.
func SpeakersOf(cues []Cue) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(n string) {
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	for _, c := range cues {
		add(c.Speaker)
		for _, p := range c.Parts {
			add(p.Speaker)
		}
	}
	return out
}

// DirectiveNames lists every command a show file may use as [name] or [name: ...].
// Help text and tests are checked against this list, so adding a directive
// without documenting it fails the build's tests.
var DirectiveNames = []string{"clear", "blank", "image", "fullscreen", "loop", "section", "note", "song", "together", "language", "settings"}
