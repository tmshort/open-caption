// Package script parses open-caption scripts (a Markdown subset, see
// docs/03-script-format.md) into cues. It is pure: no I/O, no network.
package script

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
	Item    *Item   `json:"item,omitempty"`   // image / fullscreen
	Items   []Item  `json:"items,omitempty"`  // loop
	Period  float64 `json:"period,omitempty"` // loop default seconds; 0 = config default
}

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
	Language     string       `json:"language"`
	LanguageName string       `json:"languageName,omitempty"`
	Cues         []Cue        `json:"cues"`
	Diagnostics  []Diagnostic `json:"diagnostics"`
}

// Options control parsing.
type Options struct {
	// Assets lists the uploaded image filenames. When non-nil, image references
	// are checked against it. When nil, references are not checked.
	Assets []string
}

// DirectiveNames lists every command a script may use as [name] or [name: ...].
// Help text and tests are checked against this list, so adding a directive
// without documenting it fails the build's tests.
var DirectiveNames = []string{"clear", "blank", "image", "fullscreen", "loop", "section", "note", "song", "language"}
