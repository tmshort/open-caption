# Writing a script

A script is a plain text file (`.txt` or `.md`) written in a simple form of Markdown. If you paste in an ordinary theater script, it already works: **every line is one caption**, and one press of the spacebar sends one line. Blank lines are ignored, so use them to keep the script readable.

Long lines wrap on the screen automatically. You decide where to split a long speech into lines, to match the style of the performance. The program never splits or joins lines for you.

## Formatting

| Write | You get |
|-------|---------|
| `**loud**` | **bold** text |
| `*aside*` or `_aside_` | *italic* text |
| `(thunder rumbles)` | a whole line in parentheses is shown as a sound or stage direction |
| `HAMLET: To be, or not to be` | a speaker label above the line (the name must be in CAPITAL LETTERS followed by a colon) |

Anything else is shown as typed. Links, lists, tables and headings are not supported. Raw HTML is not supported either: if you type `<b>`, it appears literally, and the page shows a note suggesting the Markdown form.

```script
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Songs and music

For a musical, show lyrics between music notes (♪). There are three ways, from quickest to most flexible:

- **Type or paste the notes yourself.** ♪ and ♫ work anywhere in the script. In the editor, the **♪** and **♫** buttons insert them at the cursor.
- **Start a line with `~`.** The line is treated as a sung lyric: it is shown in italics with a ♪ added at each end, and the `~` itself is not shown. A closing `~` is optional. This also works after a speaker's name.
- **Wrap a whole song in `[song]` and `[/song]`.** Every line inside is sung, so you do not need a `~` on each one. The title after the colon is a bookmark for the operator only. Notes, clears and blanks can still go inside.

```script
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

```script
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

Voices that join in **one after another** are separate lines: press Space for each. Use `[together]` only for lines that appear at the same moment. A `[together]` block holds only lines of text, needs at least two, and inside a `[song]` every voice is sung.

A tilde in the middle of a line, such as `about ~5 minutes`, is just text. To start an ordinary line with one, write `\~`. See [Musicals](topic:admin-musicals) for more advice.

## Several lines in one caption

End a line with a backslash to join it with the next line, or put the lines between `[[` and `]]`.

```script
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
- Lines beginning with `#` are reserved for headings in a future version and are ignored. Use a comment instead.

```script
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
| `[loop]` … `[/loop]` | A repeating set of full-screen items for pre-show or intermission. See [Loops](topic:admin-loops) |
| `[together]` … `[/together]` | Two or more voices shown at the same time. See [Songs and music](topic:admin-musicals) |
| `[song: title]` … `[/song]` | A song: every line inside is shown as a sung lyric. See [Songs and music](topic:admin-musicals) |
| `[section: name]` | A bookmark for the operator |
| `[note: text]` | A note for the operator |
| `[language: es]` | Marks a translated script. See [Languages](topic:admin-languages) |

```script
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

```script
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## What the warnings mean

The Script step lists problems with line numbers. Click one to jump to the line.

- **"Unknown command [x]"** — a line in square brackets that is not a command. If it was meant to be a caption such as a sound effect, write it with parentheses, like `(laughter)`.
- **"Image … not found"** — the picture is not uploaded, or the name is spelled differently. Upload it in the Images step.
- **"[loop] was opened but never closed"** — add `[/loop]` after the last item. The same applies to `[fullscreen]` blocks, `[[` blocks and comments.
- **"HTML is not supported"** — use `**bold**` and `*italic*` instead of tags.
- **"can't go inside a loop"** — loops hold only pictures and text.
- **"[together] needs at least two lines"** — a `[together]` block with one line is shown as an ordinary caption. **"[together] was opened but never closed"** — add `[/together]`; the lines you wrote are kept.
- **"[song] was opened but never closed"** — add `[/song]` after the last lyric. The lyrics you wrote are still kept.

Errors never stop you from saving. The part with the error is skipped, and the rest of the script still works.
