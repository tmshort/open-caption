# 03 — Show file Format

A show file is a **Markdown** text file (`.md` or `.txt`; both are parsed the same
way). Design goals: an author can paste in a normal theater show file and it works
with zero markup; extra features are opt-in and read naturally to a human.

**No embedded HTML.** Formatting uses Markdown only. Any `<tag>` in the text is
shown literally (escaped) and produces an info diagnostic, never executed.
(HTML comments are the one exception; see rule 4.) This also means the server needs no HTML sanitizer: it escapes all text and
emits only the small set of tags it generates itself.

## Basic rules

1. UTF-8 text; CRLF or LF; a leading BOM is ignored. The file may have any name; Setup also reads Windows-1252 and UTF-16 (with a mark) files, converting them. Preparers are pointed to the help topic *Preparing a show file* (`web/help/admin-preparing-a-show-file.md`), which covers the workflow from script to finished file, including editing offline.
2. **Each non-blank line is one caption cue** — one press of the spacebar.
   Blank lines are ignored (use them freely for readability).
3. A caption longer than the screen **wraps automatically**. The author decides
   how to split speeches into lines to match the style of the performance;
   the system never splits or merges lines.
4. **Comments** are written `<!-- like this -->` and may span several lines.
   Comments are never shown to anyone. A comment may also appear inside a line; it is removed, and a line left empty is ignored. (This is the only HTML the format accepts.)
   To leave a message for the **operator**, use `[note: text]` instead; it is
   visible in the operator view but never sent to the audience.
   A line starting with one to six `#` and a space is a Markdown **heading** and is a section marker, the same as
   `[section: text]` (see the directive table); don't use `#` for comments. A `#` not followed by a space
   (`#hashtag`) is ignored with an info message. A caption that must begin with a literal `#` is written `\#`.
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
| `***very loud***` | ***bold italic*** (`<strong><em>`) |
| `~~old line~~` | ~~strikethrough~~ (`<del>`; GitHub-flavoured Markdown). Needs a non-space just inside each pair, so `~~ x ~~` stays literal; it cannot be confused with a sung line, which is `~` followed by a space |
| `(thunder rumbles)` — whole line in parentheses | automatically styled as a sound/stage direction |
| `![Act 1](set.jpg)` — alone on a line | fullscreen image (same as `[image: set.jpg \| Act 1]`) |
| `\*` `\_` `\#` `\[` `\:` | literal character |

Not supported (shown as plain text): links, lists, tables, blockquotes, code,
headings (ignored), raw HTML. **Underline, superscript and subscript deliberately have no markup**
(Markdown has none: `__x__` is bold in CommonMark, and `~sub~`/`^sup^` are Pandoc extensions that would clash with `~` for sung lines).
Any Unicode character works as typed, so H₂O, x², 1ˢᵗ and ½ need no syntax. The generated tags are only `<strong>`, `<em>` and `<del>`.

## Settings: the title, the look and other show-wide options

A show file also says how the show **looks and behaves**, so the file is self-contained: everything about a show except its
images is in its show file(s). Settings go in a `[settings]` block, usually at the top, one `name: value` per line, ended by
`[/settings]`. Names are not case-sensitive and `_`, `-` and spaces are interchangeable. Settings are not cues: they never
affect the operator's list or the structure comparison between languages. If `[/settings]` is missing, an error is reported and the block ends at the last known setting line: what follows is read as captions, and Setup's rewrites (Rename, Look, Archive) close the block and keep them.

| Setting (aliases) | Value | Default comes from |
|-------------------|-------|--------------------|
| `title` (name) | 1–100 characters | "Untitled show" |
| `background` | `#rgb`, `#rrggbb`, or a name (black, white, yellow, red, green, blue, cyan, magenta, orange, gray) | `defaults.look.background` |
| `text color` (foreground) | a color | `defaults.look.text_color` |
| `speaker name color` (speaker color, accent) | a color | `defaults.look.speaker_name_color` |
| `font` | `system` (default, clean, sans), `serif`, `mono` (monospace), `dyslexic` (dyslexia friendly, OpenDyslexic), `hyperlegible` (high legibility, Atkinson) | `defaults.look.font` |
| `text size` (size, scale) | 50–300, optionally with `%` | `defaults.look.text_size_percent` |
| `speaker names` (speakers) | `label` (one color), `color` (a different color per character), `hidden` | `defaults.look.speaker_names` |
| `speaker position` (speaker placement) | `above` (the name over the line), `beside` (on the same line, before the text, bold), `column` (a column left of the text); voices shown together always use `beside` | `defaults.look.speaker_position` |
| `idle message` (idle) | up to 200 characters; shown before the first line is sent | `defaults.idle_message` |
| `loop seconds` (loop time) | seconds, more than 0 and at most 3600 | `defaults.loop_item_seconds` |
| `scrollback lines` (scrollback) | 20–2000 | `audience.scrollback_lines` |
| `archived` (archive) | `yes` or `no`; written only when yes | no |

**Layering.** A value comes from the most specific place that sets it: **the show file's `[settings]`, then
`config.yaml`'s `defaults`, then the program's built-in values** (config.yaml starts from the built-in values, so a
config file that says nothing changes nothing). What is *not* show configuration stays in `config.yaml` only: ports,
HTTPS, passwords, access lists, the data folder, audience limits.

**Translations.** Only `idle message` may be set in a translation's show file (each language has its own pre-show
message). Other settings there are reported as warnings and ignored, and `title` is ignored silently, because the look
and the title are set once for the whole show, in the main show file.

**Problems** are never fatal: an unknown name (with a suggestion), a value that does not make sense, a line without
`name: value`, or an unclosed block is reported with its line number, and the rest of the show file still works. Later
lines win over earlier ones, and several blocks are merged. An unclosed `[settings]` reads the rest of the file as
settings (the error says so).

**Setup edits the block.** The Look step and Rename rewrite one `[settings]` block near the top (after any leading comments
and the `[language: …]` line), replacing existing blocks, keeping unknown lines, and leaving the rest of the file exactly
as it was. Comments *inside* the block are not kept; put them above it. A saved look writes every look setting, so a
show file records exactly how it looks wherever it is copied. A show whose file sets none of the look settings counts as
"look not saved" in Setup. `archived: yes` marks an archived show (Setup's Archive and Restore buttons add and remove the line).

## Directives

A directive is a single line `[name]` or `[name: arguments]`. Arguments are
separated by `|`. A final argument that looks like a duration (`15s`, `2.5s`)
is the item's display time (only meaningful inside a loop; on an `[image]` or `[fullscreen]` outside a loop the time is ignored
and a warning says so).

| Directive | Effect |
|-----------|--------|
| `[clear]` | Clear the scrolling captions; show an empty caption screen |
| `[blank]` | Blank screen (background only); captions are kept |
| `[image: file]` / `[image: file \| caption]` | Fullscreen image until the next cue |
| `[fullscreen: text]` | Fullscreen centered large text until the next cue (Markdown formatting allowed) |
| `[fullscreen]` … `[/fullscreen]` | Same, multi-line text |
| `[loop]` … `[/loop]` / `[loop: 10s]` | Looping fullscreen items (see *Loops*) |
| `[together]` … `[/together]` (aliases `[duet]`, `[counterpoint]`) | Voices shown at the same time, one cue |
| `[song]` … `[/song]` / `[song: title]` | Every caption inside is a sung lyric (see *Songs and lyrics*) |
| `[settings]` … `[/settings]` | Show-wide settings (see *Settings*); never a cue |
| `[section: name]` or `# name` | Operator-only marker/bookmark; no audience effect. A Markdown heading (`#` to `######` and a space) is the same cue; a closing run of `#` is dropped |
| `[note: text]` | Operator-only note shown inline in the operator view (e.g. "wait for door slam") |
| `[language: es]` or `[language: es \| Español]` | First line of a translated show file, with an optional display name for the picker; see *Languages* |

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
  text** (shown as fullscreen text). `[blank: 3s]` is also allowed as a pause item.
- An item's display time is its trailing duration (`| 15s`); otherwise the
  loop's period (`[loop: 10s]`); otherwise `defaults.loop_item_seconds` from
  `config.yaml` (default 10 s).
- A time can be at most 3600 seconds; a longer one is cut to 3600 with a warning.
- Loops can't be nested. Captions, `[clear]`, `[section]` (and so a `#` heading) and `[note]` inside a
  loop produce a warning and are skipped.
- The loop is **one cue** with one position in the operator list.
- A loop can have a **name** (`[loop: 10s, name Intermission]`; see *Loop names* below).
- **Options** go in the brackets, comma-separated, in any order: an item time (`10s`) and a timer
  (`timer 15 minutes`). See *Loop timers* below.

Behavior (see doc 02 for the state model):
- **Starting.** When the operator presses NEXT on a loop cue, the loop starts
  automatically and the cursor moves to the cue after the loop.
- **Pre-show.** If a loop is the first cue in the show file (comments, sections and
  notes don't count), it starts automatically as soon as the show is activated.
  Before that (and if the show file doesn't start with a loop), the audience sees
  the idle message.
- **Ending.** NEXT ends the loop and sends the cue after it (same key as for any
  caption). Back (undo) after that returns to the loop.
- **Stopping.** The operator's STOP command (Shift+Q) stops the loop and leaves a
  blank screen. NEXT then proceeds as usual.
- **Restarting.** There is no restart command. Back "backs into" the loop: after
  NEXT ended it, or after STOP, Back restores the loop **running from its first item**.

### Loop names (REQ-070)

A loop can be given a name, which the **audience report** uses to label the stretch during which the loop ran (on its graph and in its list of loops).

```
[loop: 12s, name Intermission, timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
[/loop]
```

- `name <text>` (a colon after the word is allowed) is one of the loop's options, which share the brackets, separated by commas, in any order (`[loop: 10s, name Pre-show]`). The text is free, at most 40 characters (longer is cut, with a warning), and has no commas. The audience never sees it. An empty name (`[loop: name]`) is a warning and the loop has no name.
- Only the name in the main language's show file is used. A translation may name its loops differently or not at all; `CompareStructure` ignores names.
- Without a name the report guesses from the loop's place among the lines the audience sees (captions, images, fullscreen messages): before the first is the *pre-show*, after the last the *post-show*, between them an *intermission*.

### Loop timers (REQ-065)

A loop can carry a countdown, so the audience can see about how long an intermission has left or how
long until the show starts. It has two parts: the **option** on the loop and the **placeholder** in its text.

```
[loop: 12s, timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
[/loop]
```

- `timer <value>` (or `countdown <value>`; a colon after the word is allowed) takes either
  - **minutes**: `15`, `15 min`, `15 minutes`, `7.5m` (more than 0, at most 600). The countdown **starts when the loop starts**, which is each time it is entered: NEXT onto the loop cue, a leading (pre-show) loop starting at activation, and Back into the loop (which restarts the loop from its first item and the countdown from its full length). Saving the show file while the loop runs, and restarting the program (the end time is saved in `live.json`), do **not** restart it.
  - a **time of day**: `7:30pm`, `7pm`, `12:15 a.m.`, or 24-hour `19:30` (a time with a colon and no am/pm is 24-hour). The countdown runs to that time **today**, by the local clock and time zone of the computer running the program; a time already past stays past, so the timer reads `0:00` rather than counting towards tomorrow. A bare number is minutes, never a time.
- `{timer}` in any text item of that loop (plain line, `[fullscreen: …]`, but not an image caption) becomes the time left in **whole minutes, rounded up, without seconds** (`15 min`; `1 hr 5 min` from one hour), with the unit in the device's language (`Intl.NumberFormat`; an old browser shows `15 min`). At zero it stays `0 min` (the loop carries on until the operator moves on). Only items containing it show it.
- The page counts down locally once a second from an end time (unix ms) the server sends with the item, and corrects for a device clock that is wrong using the server's time in the snapshot.
- Diagnostics: a bad value (`timer soon`, `timer 9999`) is a warning and the loop has no timer; `{timer}` in a loop without a timer is a warning and is shown as typed; a loop with a timer whose items never use `{timer}` is a warning (the countdown would never be shown). A `{timer}` outside a loop is ordinary text.
- Translations: every language's loop must have the **same timer** (minutes or time of day; only the words around `{timer}` differ), otherwise the translation is held back like any other structure difference.

### Songs and lyrics (musicals)

Lyrics are captions flagged as *sung*. The server shows a sung caption in italics between music notes: `♪ lyric ♪`. Three ways to write one:

| Write | Result |
|-------|--------|
| `~ lyric ~` (closing `~` optional) | a sung caption; the `~` markers are removed and ♪ added at both ends |
| `[song]` … `[/song]` or `[song: Title]` … `[/song]` | every caption inside is sung; the title becomes an operator-only `section` cue `♪ Title` |
| a literal `♪` or `♫` anywhere | ordinary text, unchanged (no italics) |

Details: `~` alone (or `~ ~`) is a bare `♪`. The marker works after a speaker (`ELSA: ~ Let it go`). A `~` that is not followed by a space, or is not at the start of the line, is plain text; `\~` escapes it. Inside `[song]`, `[note]`, `[section]`, `[clear]`, `[blank]` and images still work, and a speaker label is kept. If `[song]` is never closed, the lyrics are kept and an error is reported. Songs cannot nest or appear inside a loop. For multi-line captions (`[[ … ]]` or a trailing backslash), a leading `~` makes every line of the caption sung.

### Voices at the same time (counterpoint, duets)

Several voices shown together are **one cue** (one press of Space) with `Parts`. Two ways to write them:

| Write | Result |
|-------|--------|
| `[together]` … `[/together]` (aliases `[duet]`, `[counterpoint]`) | every line inside is one voice; each keeps its own speaker label and `~` / song styling |
| `A: ~ one ~ \|\| B: ~ two ~` on a single line (voices separated by ` \|\| `) | the same, for short exchanges |

A block needs at least two lines (with one, it is shown as an ordinary caption and a warning is given). Only lines of text may appear inside; other commands are skipped with a warning. An unclosed block keeps its lines and reports an error. Inside `[song]`, every voice is sung. The cue's `Text` is the voices joined with newlines, each prefixed with its speaker, for search; the audience gets each voice separately so it can lay them out side by side. Voices that enter one after another are separate cues.

### Speaker detection

A leading label is a speaker if it matches `^[A-Z][A-Z0-9 .'\-&,/()#]{0,59}:\s+\S`
(upper-case word(s), colon, space; up to 60 characters). Besides letters, digits, spaces, `.`, `'` and `-`, a label
may contain `&`, `,`, `/`, `(`, `)` and `#`, so a group can sing together (`ALICE & BOB:`, `ALL (EXCEPT BOB):`,
`BOB, DAVE, ERIN & CAROL:`) and a part can be numbered (`PARENT #1:`). `HAMLET: To be, or not to be` → speaker
`HAMLET`, text `To be, or not to be`. A label is one speaker, whole: `ALICE & BOB` is not split into two, so with
`speaker names: color` it has a color of its own. Mixed-case `Note: this` is not a speaker.
A name and colon with nothing after them (`BOB:` alone on a line) is not a speaker: the line is shown as typed, and a warning points it out, since the words were probably meant to follow. A label in mixed case (`Bob: Hi`) is plain text, with a warning when the same name is a speaker elsewhere in the file. Escape with `\:`. The `speaker names` setting controls whether speakers are
hidden, shown as a label, or shown in per-character colors (default: label, no color).

### Languages

A show has one show file per language: `show.md` for the **main** language (named by its `[language: …]` line, by default `defaults.language`, usually `en`), which is what the operator follows, and `show.<code>.md` for each **translation**. A translation starts with a language line:

```
[language: es | Español]
```

The code should match the file name, and the optional name after `|` (written in its own language) is
what the audience sees in the language menu. Without a name, a built-in name for the code is used.

**Identical structure.** `showfile.CompareStructure(main, translation)` requires the same number of cues,
in the same order, of the same kinds; captions with the same number of simultaneous voices; loops with
the same number and kinds of items. Words, speaker names, section and note text, image names, and loop
timings may differ. This is checked cue by cue, and the first difference is reported as an error with the
translation's own line number (everything after it would be shifted, so later cues are not listed).
One operator NEXT sends the same cue number in every language, so a mismatch would show the wrong line
or run off the end. A translation that does not match is **not offered to the audience** (the other
languages are unaffected) until it is fixed.

The Setup editor can start a translation as a copy of the main show file, which has the right structure
already. Images may differ per language by using different filenames at the same position. The audience's
language is chosen as described in doc 02 (`?lang=`, then `Accept-Language`, then the main language).

## Parsed model (Go)

```go
type CueKind string // caption, clear, blank, image, fullscreen, loop, section, note

type Item struct {      // a fullscreen item (used by image/fullscreen cues and in loops); also Timer bool, Deadline int64 (unix ms, set by the engine)
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
    Sung    bool        // lyric: Text/HTML already include the surrounding ♪ notes
    Parts   []Part      // two or more voices shown at the same time (caption only)
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

`showfile.Parse(src string, assets AssetLookup) ParseResult` is a pure function.

## Diagnostics (validation)

Each is reported with a line number and a plain-English message:
- warning: `Line 42: image "curtian.jpg" not found. Did you mean "curtain.jpg"?`
- warning: `Line 17: unknown command [fade]. This line was skipped.`
- warning: `Line 63: captions can't go inside a loop. This line was skipped.`
- warning: `Line 12: "BOB:" is a speaker's name with nothing after the colon. Put the words on the same line, or write \: if the colon is part of the text.` (the line is still shown as typed)
- warning: `Line 12: The speaker label "BOB, ALICE, CAROL & BOB:" names BOB twice. Is one of them a different character?` (a group label that gives the same name twice is probably a typo)
- warning: `Line 40: "Bob:" is not in capitals, so it is shown as text, not as the speaker BOB:. Write the name in capitals.` (only when the file also uses that name as a speaker, so ordinary words such as `Note:` are left alone)
- info: `Line 52: This caption line is 132 characters long. On a phone it wraps over several lines; consider splitting it into shorter lines.` (a caption line over `LongCaption`, 100 characters; for a `[[ ]]` block or voices shown together, the longest line)
- warning: `Line 60: This caption shows 5 voices at once. A phone stacks them and a shared screen may not have room for them all; consider splitting it into two captions.` (voices shown together, `[together]` or ` || `, from `ManyVoices`, 5)
- error: `Line 88: [loop] was opened but never closed.`
- error: `Line 12: comment <!-- was opened but never closed.`
- info: `Line 30: A line starting with # is a heading only when a space follows the # (# Act 2), so this line was ignored. Write \# to start a caption with a #.` (`#hashtag`, seven or more `#`, or a bare `#`)
- warning: `Line 31: A heading can't go inside a loop. This line was skipped.` (a heading is a section; inside `[together]` the usual "can't go inside [together]" applies)
- info: `Line 5: HTML is not supported; "<b>" will appear as plain text. Use **bold** instead.`
- info: summary — `312 captions, 4 images, 2 loops, 3 sections`. The image count is the number of different image files the show file uses (those inside a loop too, a file used twice once), the same number the Images step shows as in use; fullscreen messages inside loops are counted as well.

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

The original text is stored verbatim as the show's main show file (`show.md`, with translations in `show.<lang>.md`); the setup UI
edits that text. There is no lossy intermediate format.

## Checking a show file from the command line

`open-caption --check-show show.md [translation.es.md ...]` parses a show file with the same code as Setup's editor and prints its problems as `file:line: severity: message`, then exits 0 (no errors), 1 (errors) or 2 (unreadable file). It starts no server and needs no `config.yaml`. Translations are compared with the main file's structure; images are checked against the `assets` folder beside the file (or `--assets`). See the [configuration reference](configuration-reference.md#checking-a-show-file).
