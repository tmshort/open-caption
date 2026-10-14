# 03 — Script Format

Design goals: an author can paste in a normal theater script and it works with
zero markup; extra features are opt-in and read naturally to a human.

## Basic rules

1. UTF-8 text file (`.txt`; `.md` accepted, treated as plain text). CRLF or LF.
2. **Each non-blank line is one caption cue** — one press of the spacebar.
   Blank lines are ignored (use them freely for readability).
3. Lines starting with `#` are **comments** (ignored). A line that must begin
   with a literal `#` is written `\#`.
4. Lines in **square brackets** are **directives** (commands), case-insensitive.
5. An optional upper-case speaker prefix `NAME: text` is recognized and
   rendered as a label (see *Speaker detection* below).
6. Long text wraps automatically; the author never hand-wraps.
7. To put **several lines in one cue** (e.g. a song lyric couplet), end each
   line except the last with ` \` (space + backslash), or use a `[[ … ]]` block (below).

## Directives

All directives are a single line `[name]` or `[name: argument]`.

| Directive | Effect | Example |
|-----------|--------|---------|
| `[clear]` | Clear the scrolling captions; show empty caption screen | `[clear]` |
| `[blank]` | Blank screen (background only), captions retained | `[blank]` |
| `[image: file]` | Fullscreen image until next cue | `[image: curtain.jpg]` |
| `[image: file, caption text]` | Image with optional caption beneath | `[image: set.jpg, Act 1 – The Garden]` |
| `[fullscreen: text]` | Fullscreen centered large text | `[fullscreen: INTERMISSION — 15 minutes]` |
| `[slideshow: a.jpg, b.jpg, c.jpg; 8s]` | Looping fullscreen images, N seconds each (default 8s) | pre-show |
| `[section: name]` | Operator-only marker/bookmark, no audience effect | `[section: Act 2, Scene 1]` |
| `[note: text]` | Operator-only note shown inline in the operator view (e.g. "wait for door slam") | `[note: hold for applause]` |
| `[preshow]` … `[show]` | Marks pre-show region; see below | |
| `[html]` … `[/html]` | Block of raw HTML shown as a caption | see below |

Directive names are forgiving: `[Clear]`, `[CLEAR SCREEN]`, `[clear screen]`
all work. Unknown directive → warning, line is skipped (never shown to the audience).

### Multi-line captions

```
[[
Row, row, row your boat,
Gently down the stream
]]
```
Everything between `[[` and `]]` is one cue, line breaks preserved.

### Inline HTML

Within any caption line, a limited set of HTML tags is allowed:
`<b> <i> <u> <em> <strong> <br> <span class="…"> <small> <big>`, plus
`<img src="file">` (resolved against uploaded assets).
Everything else is stripped. `<script>`, event handlers, external URLs and
`style` attributes are always removed. Predefined span classes: `loud`, `whisper`,
`sfx`, `lyric`, `red`, `green`, `blue`, `yellow`.

Friendly shorthand (Markdown-lite) also works in captions:
`*italic*`, `**bold**`, and `(parenthetical stage directions)` auto-styled
as sound/stage text when the whole line is a parenthetical, e.g. `(thunder rumbles)`.

Full HTML block (for fullscreen layouts):

```
[html]
<h1>Thanks for coming!</h1>
<p>Please silence your phones.</p>
[/html]
```
An `[html]` block is a **fullscreen** cue (sanitized with a wider allow-list:
headings, paragraphs, lists, images, inline styles limited to color/alignment/size).
Use `[html: caption]` ... `[/html]` to make it an in-scroll caption instead.

### Speaker detection

A leading label is recognized as a speaker if it matches `^[A-Z][A-Z0-9 .'\-]{0,29}:\s+\S`
(upper-case word(s) then colon). `HAMLET: To be, or not to be` → speaker
`HAMLET`, text `To be, or not to be`. Mixed-case `Note: this` is not a speaker.
A show-level setting controls whether speakers are shown, hidden, or colored per-character.
Speaker detection can be escaped with `\:`.

### Pre-show

Everything before the first `[show]` directive is **pre-show** if the file
contains a `[preshow]` directive at the top. The pre-show is displayed
automatically when the show is loaded/idle; the operator ends it by advancing
past `[show]`. Typical:

```
[preshow]
[slideshow: sponsor1.png, sponsor2.png, cast.jpg; 10s]
[fullscreen: Welcome to The Odd Couple — Please silence your phones]
[show]
[clear]
(lights up)
OSCAR: Felix, you're late.
```

If no `[preshow]` is present, the show begins at the first cue and the audience
sees a configurable idle message ("Captions will begin shortly") until then.

## Parsed model (Go)

```go
type CueKind string // caption, clear, blank, image, fullscreen_text, fullscreen_html, slideshow, section, note

type Cue struct {
    Index   int
    Kind    CueKind
    Line    int      // 1-based source line (start)
    Speaker string   // caption only
    HTML    string   // sanitized HTML (caption/fullscreen_html)
    Text    string   // plain-text version (also used for operator display and search)
    Image   string   // asset filename (image)
    Images  []string // slideshow
    Seconds int      // slideshow per-slide
    Preshow bool
}

type Diagnostic struct {
    Line     int
    Severity string // "error" | "warning" | "info"
    Message  string
}

type ParseResult struct {
    Cues        []Cue
    Diagnostics []Diagnostic
}
```

`script.Parse(src string, assets AssetLookup) ParseResult` is a pure function.

## Diagnostics (validation)

Examples, each reported with line number and a plain-English message:
- warning: `Line 42: image "curtian.jpg" not found. Did you mean "curtain.jpg"?` (fuzzy match)
- warning: `Line 17: unknown command [fade]. This line was skipped.`
- error: `Line 88: [html] block opened but never closed.`
- info: `Line 5: HTML tag <marquee> is not allowed and was removed.`
- info: summary — `312 captions, 4 images, 3 sections, estimated runtime n/a`.

Errors don't prevent saving; the affected construct is skipped. The setup UI
must make fixing easy (inline line list, click to jump to line in editor).

## Complete example

```
# The Odd Couple — captions v3
[preshow]
[slideshow: sponsors.png, cast.jpg; 10s]
[show]

[section: Act 1, Scene 1]
[clear]
(Poker night. Cigar smoke fills the room.)
SPEED: Where's Felix?
MURRAY: He's always late. **Always.**
[note: Felix enters stage left]
[[
FELIX: I'm so sorry,
I got lost in the kitchen.
]]

[section: Intermission]
[fullscreen: INTERMISSION — 15 minutes]
[image: clock.png]
```

## Round-tripping

The original text is stored verbatim as `script.txt`; the setup UI edits that
text. There is no lossy intermediate format.
