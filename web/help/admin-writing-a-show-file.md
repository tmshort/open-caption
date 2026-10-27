# Writing a show file

A show file is a plain text file (`.txt` or `.md`) written in a simple form of Markdown. If you paste in an ordinary theater show file, it already works: **every line is one caption**, and one press of the spacebar sends one line. Blank lines are ignored, so use them to keep the show file readable.

Long lines wrap on the screen automatically. You decide where to split a long speech into lines, to match the style of the performance. The program never splits or joins lines for you.

To check a show file, or to prepare one in your own text editor, see [Preparing a show file](admin-preparing-a-show-file.md).

## Formatting

| Write | You get |
|-------|---------|
| `**loud**` | **bold** text |
| `*aside*` or `_aside_` | *italic* text |
| `***very loud***` | bold and italic text |
| `~~old line~~` | text with a line through it (a correction, or a joke crossed out) |
| `(thunder rumbles)` | a whole line in parentheses is shown as a sound or stage direction |
| `HAMLET: To be, or not to be` | a speaker label above the line (the name must be in CAPITAL LETTERS followed by a colon; a group is fine too: `ALICE & BOB:`, `ALL (EXCEPT BOB):`) |

**Crossing text out.** Put two tildes on each side, `~~like this~~`, and the words are shown with a line through them. There must be a non-space character just inside each pair (`~~ x ~~` stays as typed). A sung line starts with a single `~` followed by a space, so the two never get mixed up; to start an ordinary line with a tilde, write `\~`.

Anything else is shown as typed. Links, lists, tables and headings are not supported. Raw HTML is not supported either: if you type `<b>`, it appears literally, and the page shows a note suggesting the Markdown form. Underline, superscript and subscript have no markup, but any character you can type works as it is, so write H₂O, x², 1ˢᵗ or ½ directly.

```showfile
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Checking the character names

Click **Characters…** above the editor to see every name used as a speaker, the most used first, with how many lines each speaks on. It is the quick way to find a misspelled name: **ALCE** with one line beside **ALICE** with forty is marked *Did you mean ALICE?*. Click a name to go to its first line. A group such as `ALICE & BOB:` counts for each of its names.

## Going back to an earlier version

Every time you save, the version it replaced is kept (the last 50 for each language). Click **Earlier versions…** above the editor, choose a time and press **Load**. That only puts the old text in the editor; nothing changes until you press **Save**, so you can look first. **Compare** shows, line by line, what loading a version would change. A version you give a name to (in the box beside it, or with **Keep** to name what is saved now) is never forgotten, however many saves follow. Export does not include the earlier versions.

## Show settings: the title and the look

Everything about a show other than its pictures is stored in the show file, including its **title** and how the captions **look**. These go in a `[settings]` block, usually at the top, with one `name: value` per line:

```showfile
[settings]
title: The Garden Party
background: #000000
text color: #ffffff
speaker name color: #ffffff
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
loop seconds: 10
scrollback lines: 200
[/settings]
```

| Setting | What it does |
|---------|--------------|
| `title` | The show's name, as listed in Setup |
| `background` | The screen color behind the captions, like `#000000` (a name such as `black` or `yellow` also works) |
| `text color` | The color of the captions, like `#ffffff` |
| `speaker name color` | The color of speaker names (when they are not given a color each) |
| `font` | `system`, `serif`, `mono`, `dyslexic` or `hyperlegible` |
| `text size` | A percentage from 50 to 300; 100 is normal |
| `speaker names` | `label` (one color), `color` (a different color for each character) or `hidden` |
| `speaker position` | `above` (the name over the line), `beside` (on the same line, before the text) or `column` (in a column to the left of the text) |
| `idle message` | The words shown before the first line is sent |
| `loop seconds` | How long each item of a `[loop]` stays when it does not say |
| `scrollback lines` | How many captions a viewer can scroll back through (20 to 2000) |
| `archived` | `yes` if the show is archived; Setup's **Archive** and **Restore** buttons add and remove it |

**You do not have to write any of it.** The **Look** step in Setup fills in the block for you (and **Rename** sets the title), and anything you leave out simply uses a default. The defaults come in two layers: the program has built-in values, which the settings file `config.yaml` can change for every show (its `defaults` section), and then each show file overrides whatever it sets. The most specific one wins: **show file, then config.yaml, then built-in**.

A few rules: a translation's show file may set only `idle message` (the look is set once, for the whole show, in the main show file); a setting that is misspelled or has a value that does not make sense is reported with its line number and ignored; and comments inside the block are not kept when Setup rewrites it, so put comments above it. Several `[settings]` blocks are combined, with later values winning.

## Songs and music

For a musical, show lyrics between music notes (♪). There are three ways, from quickest to most flexible:

- **Type or paste the notes yourself.** ♪ and ♫ work anywhere in the show file. In the editor, the **♪** and **♫** buttons insert them at the cursor.
- **Start a line with `~`.** The line is treated as a sung lyric: it is shown in italics with a ♪ added at each end, and the `~` itself is not shown. Only the opening `~` is needed: the ♪ at the end is added for you, and a closing `~` is allowed but changes nothing. This also works after a speaker's name.
- **Wrap a whole song in `[song]` and `[/song]`.** Every line inside is sung, so you do not need a `~` on each one. The title after the colon is a bookmark for the operator only. Notes, clears and blanks can still go inside.

```showfile
OLIVER: ~ Where is love? ~
~ Row, row, row your boat

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[note: hold the last chord]
[/song]
```

### Two or more voices at once (duets and counterpoint)

When singers sing different lines **at the same time**, put them in `[together]` … `[/together]`. Every line inside is shown at once, side by side on a wide screen and stacked on a phone, and the operator sends them with a single press of Space. Each keeps its own speaker name. For a short exchange you can write the voices on one line, separated by ` || ` (a space, two vertical bars, a space). `[duet]` and `[counterpoint]` mean the same as `[together]`.

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

Voices that join in **one after another** are separate lines: press Space for each. Use `[together]` only for lines that appear at the same moment. A `[together]` block holds only lines of text, needs at least two, and inside a `[song]` every voice is sung.

A tilde in the middle of a line, such as `about ~5 minutes`, is just text. To start an ordinary line with one, write `\~`. See [Musicals](admin-musicals.md) for more advice.

## Several lines in one caption

End a line with a backslash to join it with the next line, or put the lines between `[[` and `]]`.

```showfile
GHOST: Remember me. \
Remember me.

[[
Row, row, row your boat,
Gently down the stream
]]
```

## Comments and notes

- `<!-- like this -->` is a **comment**. Nobody sees it, and it may span several lines.
- `[note: text]` is a reminder shown to the **operator only**, for example `[note: hold for applause]`.
- `[section: Act 2, Scene 1]` adds a bookmark in the operator's list. Notes and sections never use up a press of the spacebar.
- `[pace: 8s]` says how long the next line usually stays on screen, as a reminder for the **operator only**. The Run page shows it beside that line (`⏱ about 8 s`), and once the line is sent it counts down to when the next press is due, turning yellow when it is. It never sends anything by itself. The time is more than 0 seconds. It belongs to the next line the operator sends (sections and notes are passed over) and cannot go inside a loop or a `[together]` block.
- A line that starts with `#` and a space is a **heading** and works like `[section: …]`: `# Act 2` (or `## Scene 1`, with up to six `#`) adds a bookmark in the operator's list and is never shown to the audience. A `#` with no space after it (a hashtag, say) is ignored, with a note; to start a caption with `#`, write `\#`. Headings cannot go inside a loop.

```showfile
<!-- Act 1 captions, version 3 -->
[section: Act 1, Scene 1]
[note: lights come up slowly, wait for the first line]
HAMLET: Who's there?
```

## Commands

A line in square brackets is a command. Names are not case-sensitive.

| Command | What it does |
|---------|--------------|
| `[clear]` | Clear the captions from the screen |
| `[blank]` | Blank the screen (the captions are kept) |
| `[image: file.png]` | Show a picture full screen. Add a caption with `[image: file.png \| text]` |
| `![caption](file.png)` | The same, written the Markdown way |
| `[fullscreen: text]` | Show large text full screen. For several lines use `[fullscreen]` … `[/fullscreen]` |
| `[loop]` … `[/loop]` | A repeating set of full-screen items for pre-show or intermission. See [Loops](admin-loops.md) |
| `[together]` … `[/together]` | Two or more voices shown at the same time. See [Songs and music](admin-musicals.md) |
| `[song: title]` … `[/song]` | A song: every line inside is shown as a sung lyric. See [Songs and music](admin-musicals.md) |
| `[section: name]` | A bookmark for the operator |
| `[note: text]` | A note for the operator |
| `[pace: 8s]` | A pacing reminder for the operator (see above) |
| `[language: es]` | Marks a translated show file. See [Languages](admin-languages.md) |
| `[settings]` … `[/settings]` | The show's title and look. See *Show settings* above |

The time options only mean something inside a [loop](admin-loops.md): `| 15s` after an image or fullscreen text, and `[blank: 3s]`. Elsewhere an image or fullscreen text stays until the operator moves on, a time on it is ignored and reported as a warning, as is a time on `[blank]`.

```showfile
[clear]
(Lights up on the garden.)
[image: garden.png | The garden at dusk]
[fullscreen: **Please remain seated**]
[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
[blank]
```

An image, fullscreen text or blank stays on the audience screens until the operator sends the next line.

## Special characters

Put a backslash before a character to show it literally instead of treating it as formatting: `\*`, `\_`, `\#`, `\[`, `\]`, `\:`, `\(`, `\)`, `\!`, `\~` and `\\` (a backslash itself). For example, a caption that must begin with a number sign is written `\# 1 hit`, and `HAMLET\: not a speaker` shows the colon without making a speaker label.

```showfile
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## What the warnings mean

The Show file step lists problems with line numbers. Click one to jump to the line.

- **"Unknown command [x]"** — a line in square brackets that is not a command. If it was meant to be a caption such as a sound effect, write it with parentheses, like `(laughter)`.
- **"Image … not found"** — the picture is not uploaded, or the name is spelled differently. Upload it in the Images step.
- **"[loop] was opened but never closed"** — add `[/loop]` after the last item. The same applies to `[fullscreen]` blocks, `[[` blocks and comments.
- **"HTML is not supported"** — use `**bold**` and `*italic*` instead of tags.
- **"can't go inside a loop"** — loops hold only pictures and text.
- **"[together] needs at least two lines"** — a `[together]` block with one line is shown as an ordinary caption. **"[together] was opened but never closed"** — add `[/together]`; the lines you wrote are kept.
- **"This caption shows 5 voices at once"** — five or more voices in one `[together]` block (or ` || ` line) are a lot for a phone, which stacks them, and for a shared screen, which has room for about four and never scrolls. Split it into two captions.
- **"Unknown setting" / "is not a font" / "is not a color"** — a line in the `[settings]` block that Setup cannot use; it is ignored and the default applies. The message says what to write instead.
- **"[settings] was opened but never closed"** — add `[/settings]`; everything after an unclosed block is read as settings.
- **"[song] was opened but never closed"** — add `[/song]` after the last lyric. The lyrics you wrote are still kept.

Errors never stop you from saving.

## If two people edit the same show file

If someone else saves the show file while you have it open, the program will not quietly overwrite their work. When you save, it tells you the show file was changed by someone else and lets you **use their version** (replacing what is in your editor) or **save yours anyway**. Copy anything you want to keep before choosing. The part with the error is skipped, and the rest of the show file still works.
