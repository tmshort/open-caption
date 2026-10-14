# 03 — Script Format

A script is a **Markdown** text file (`.md` or `.txt`; both are parsed the same
way). Design goals: an author can paste in a normal theater script and it works
with zero markup; extra features are opt-in and read naturally to a human.

**No embedded HTML.** Formatting uses Markdown only. Any `<tag>` in the text is
shown literally (escaped) and produces an info diagnostic, never executed.
(HTML comments are the one exception; see rule 4.) This also means the server needs no HTML sanitizer: it escapes all text and
emits only the small set of tags it generates itself.

## Basic rules

1. UTF-8 text; CRLF or LF; a leading BOM is ignored.
2. **Each non-blank line is one caption cue** — one press of the spacebar.
   Blank lines are ignored (use them freely for readability).
3. A caption longer than the screen **wraps automatically**. The author decides
   how to split speeches into lines to match the style of the performance;
   the system never splits or merges lines.
4. **Comments** are written `<!-- like this -->` and may span several lines.
   Comments are never shown to anyone. (This is the only HTML the format accepts.)
   To leave a message for the **operator**, use `[note: text]` instead; it is
   visible in the operator view but never sent to the audience.
   Lines starting with `#` are Markdown headings: they are **reserved for future
   use** (e.g. as section markers) and ignored for now, so don't use `#` for
   comments. A caption that must begin with a literal `#` is written `\#`.
5. Lines in **square brackets** are **directives** (commands), case-insensitive.
6. An optional upper-case speaker prefix `NAME: text` is rendered as a label
   (see *Speaker detection*).
7. To put **several lines in one cue** (a song couplet), end each line except
   the last with `\` (Markdown's hard line break), or use a `[[ … ]]` block.

## Text formatting (Markdown subset)

| Write | Result |
|-------|--------|
| `**loud**` | **bold** |
| `*aside*` or `_aside_` | *italic* |
| `(thunder rumbles)` — whole line in parentheses | automatically styled as a sound/stage direction |
| `![Act 1](set.jpg)` — alone on a line | fullscreen image (same as `[image: set.jpg \| Act 1]`) |
| `\*` `\_` `\#` `\[` `\:` | literal character |

Not supported (shown as plain text): links, lists, tables, blockquotes, code,
headings (ignored), raw HTML.

## Directives

A directive is a single line `[name]` or `[name: arguments]`. Arguments are
separated by `|`. A final argument that looks like a duration (`15s`, `2.5s`)
is the item's display time (only meaningful inside a loop).

| Directive | Effect |
|-----------|--------|
| `[clear]` | Clear the scrolling captions; show an empty caption screen |
| `[blank]` | Blank screen (background only); captions are kept |
| `[image: file]` / `[image: file \| caption]` | Fullscreen image until the next cue |
| `[fullscreen: text]` | Fullscreen centered large text until the next cue (Markdown formatting allowed) |
| `[fullscreen]` … `[/fullscreen]` | Same, multi-line text |
| `[loop]` … `[/loop]` / `[loop: 10s]` | Looping fullscreen items (see *Loops*) |
| `[section: name]` | Operator-only marker/bookmark; no audience effect |
| `[note: text]` | Operator-only note shown inline in the operator view (e.g. "wait for door slam") |
| `[language: es]` or `[language: es \| Español]` | First line of a translated script, with an optional display name for the picker; see *Languages* |

Names are forgiving: `[Clear]`, `[CLEAR SCREEN]`, `[clear screen]` all work.
An unknown directive produces a warning and the line is skipped (never shown
to the audience).

`section` and `note` cues never consume a keypress: when the operator reaches
them they are passed over automatically.

### Multi-line captions

```
[[
Row, row, row your boat,
Gently down the stream
]]
```
Everything between `[[` and `]]` is one cue, line breaks preserved.

### Loops

A loop is a block of fullscreen items shown one after another, repeating, until
the operator advances. It is used for pre-show and intermission.

```
[loop: 10s]
[image: sponsors.png | 15s]
[fullscreen: **Welcome to The Odd Couple**]
Please silence your phones
[image: cast.jpg | The Cast | 8s]
[/loop]
```

- Each item is `[image: …]`, `[fullscreen: …]`, `![](file)`, or a **plain line of
  text** (shown as fullscreen text). `[blank | 3s]` is also allowed as a pause item.
- An item's display time is its trailing duration (`| 15s`); otherwise the
  loop's period (`[loop: 10s]`); otherwise `defaults.loop_item_seconds` from
  `config.yaml` (default 10 s).
- Loops can't be nested. Captions, `[clear]`, `[section]` and `[note]` inside a
  loop produce a warning and are skipped.
- The loop is **one cue** with one position in the operator list.

Behavior (see doc 02 for the state model):
- **Starting.** When the operator presses NEXT on a loop cue, the loop starts
  automatically and the cursor moves to the cue after the loop.
- **Pre-show.** If a loop is the first cue in the script (comments, sections and
  notes don't count), it starts automatically as soon as the show is activated.
  Before that (and if the script doesn't start with a loop), the audience sees
  the idle message.
- **Ending.** NEXT ends the loop and sends the cue after it (same key as for any
  caption). Back (undo) after that returns to the loop.
- **Stopping.** The operator's STOP command (Shift+Q) stops the loop and leaves a
  blank screen. NEXT then proceeds as usual.
- **Restarting.** There is no restart command. Back "backs into" the loop: after
  NEXT ended it, or after STOP, Back restores the loop **running from its first item**.

### Speaker detection

A leading label is a speaker if it matches `^[A-Z][A-Z0-9 .'\-]{0,29}:\s+\S`
(upper-case word(s), colon, space). `HAMLET: To be, or not to be` → speaker
`HAMLET`, text `To be, or not to be`. Mixed-case `Note: this` is not a speaker.
Escape with `\:`. A show-level theme setting controls whether speakers are
hidden, shown as a label, or shown in per-character colors (default: label, no color).

### Languages (planned; design reserved now)

Additional languages are separate script files **identical in structure** to the
primary one (same cues in the same order and of the same kinds), with a
language identifier as the first directive:

```
[language: es]
```
The primary-language script may also declare `[language: en]`. The validator
compares each translation against the primary and reports structural mismatches
(cue count/kind) with line numbers. One operator NEXT advances *every*
language to the same cue index; the operator sees the primary language.
Image assets may differ per language by using different filenames at the same
position. The audience gets a language automatically from the browser's `Accept-Language`
header, can switch with a picker, and can be sent directly to `/?lang=es`
(see doc 05). The default is `en` (the `defaults.language` setting) when nothing matches.

## Parsed model (Go)

```go
type CueKind string // caption, clear, blank, image, fullscreen, loop, section, note

type Item struct {      // a fullscreen item (used by image/fullscreen cues and in loops)
    Kind    string      // "text" | "image" | "blank"
    HTML    string      // generated from Markdown by the server (text)
    Text    string      // plain text
    Image   string      // asset filename
    Caption string
    Seconds float64     // 0 = use default
}

type Cue struct {
    Index   int
    Kind    CueKind
    Line    int         // 1-based source line (start)
    Speaker string      // caption only
    HTML    string      // caption: generated, escaped HTML
    Text    string      // plain text (operator display, search)
    Item    *Item       // image / fullscreen
    Items   []Item      // loop
    Period  float64     // loop default period (0 = use config default)
}

type Diagnostic struct {
    Line     int
    Severity string // "error" | "warning" | "info"
    Message  string
}

type ParseResult struct {
    Language    string
    Cues        []Cue
    Diagnostics []Diagnostic
}
```

`script.Parse(src string, assets AssetLookup) ParseResult` is a pure function.

## Diagnostics (validation)

Each is reported with a line number and a plain-English message:
- warning: `Line 42: image "curtian.jpg" not found. Did you mean "curtain.jpg"?`
- warning: `Line 17: unknown command [fade]. This line was skipped.`
- warning: `Line 63: captions can't go inside a loop. This line was skipped.`
- error: `Line 88: [loop] was opened but never closed.`
- error: `Line 12: comment <!-- was opened but never closed.`
- info: `Line 30: lines starting with # are reserved for headings and were ignored. Use <!-- comment --> for comments or [note: …] for operator notes.`
- info: `Line 5: HTML is not supported; "<b>" will appear as plain text. Use **bold** instead.`
- info: summary — `312 captions, 4 images, 2 loops, 3 sections`.

Errors don't prevent saving; the affected construct is skipped. The setup UI
makes fixing easy (line list, click to jump to the line in the editor).

## Complete example

```
<!-- The Odd Couple — captions v3 -->
[loop: 10s]
[image: sponsors.png | 15s]
Welcome to **The Odd Couple**
Please silence your phones
[/loop]

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
[loop: 12s]
[fullscreen: INTERMISSION — 15 minutes]
![Please visit the lobby](lobby.jpg)
[/loop]
[section: Act 2]
[clear]
```

## Round-tripping

The original text is stored verbatim as the show's script file; the setup UI
edits that text. There is no lossy intermediate format.
