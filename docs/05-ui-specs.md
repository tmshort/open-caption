# 05 — UI Specifications

Three interfaces served from one binary. Vanilla JS/CSS, no build step,
responsive. Setup (`/admin`), Operator (`/operator`) and Audience (`/`) are
separate pages. Setup and Operator each have their own password (login form,
session cookie; no HTTP Basic auth). The admin password also opens `/operator`
so the setup user can rehearse.

## 1. Setup interface — `/admin`

Goal: a first-time volunteer can finish without reading documentation.

### First run
1. Login form. On the very first run the generated admin/operator passwords are printed to the **console only** (never to a web page, since the login page is reachable from the guest network); the login page says where to find them (console output or `config.yaml`) and how to change them.
2. Wizard stepper across the top: **① Show → ② Show file → ③ Images → ④ Look → ⑤ Test → ⑥ Go Live**.
   Steps are also reachable directly for experienced users. Each step shows a green check when complete.

### Steps
- **Show**: name; list of existing shows with Rename / Duplicate / Export / **Archive** / Delete (delete asks first). The show on air cannot be archived. Archived shows are listed apart under "Archived shows (n)" with Restore / Export / Delete; they are not offered for the other steps and cannot be activated. The archived state is the `archived: yes` line in the show file's `[settings]`, so it travels with an export (an import is always restored).
- **Show file**: (a **Help** link here opens the "Writing a show file" topic). **Language tabs** sit above the editor: the main language, each translation, and ＋ Add a language. A ⚠ marks a translation held back for not matching the main show file; translations show the main show file for reference, and structure errors appear in the problem list.
  - Drag-and-drop file, **Upload a show file…** (a button that opens the computer's file chooser; any `.txt` / `.md`, any name; UTF-8 or Windows-1252), or paste into an editor. Opening a file only fills the editor (and a toast says so): the problems list updates and nothing is stored until Save, which keeps the previous version in history. **Download the show file** saves the current file (for the language tab chosen) as a plain `.md` to edit in any program; this is the offline round trip.
  - The Show step's **Import a show** takes a `.zip` or a plain `.md` / `.txt` show file and makes a new show.
  - Editor pane (monospace, line numbers) + Diagnostics pane. Clicking a diagnostic jumps to the line.
  - Live preview of the parsed cue list (kinds shown as icons: 💬 caption, 🖼 image, ⛶ fullscreen, 🔁 loop, ⬛ blank, 🧹 clear, 📑 section, 📝 note).
  - A "Show file help" slide-out showing the syntax cheatsheet with copy-able examples (content of doc 03: Markdown subset, directives, loops).
  - Insert-directive toolbar buttons (Clear, Blank, Image…, Fullscreen text, Section, Loop, 🎵 Song, and ♪ / ♫ to insert music notes at the cursor) so authors never type brackets by hand.
  - Re-import warns if cue count differs and the show is live.
- **Images**: PNG/JPG/GIF/WebP only (no SVG); files over 1 MB get a "large — may load slowly for many devices" warning, since images are sent unscaled. Grid of uploads with thumbnails; drag-drop multi-upload; shows "used in lines 12, 40" or "unused"; missing-but-referenced images listed at top with an upload-in-place button.
- **Earlier versions** (button above the editor): every save keeps the version it replaced (the last 50 per language, in `<show>/history/`, not exported or duplicated). The dialog lists them by time and size; *Load* only fills the editor, so nothing changes until Save (and the live-show and conflict checks still apply). *Compare* shows, line by line in the browser, what loading that version would add and remove compared with the editor (long unchanged runs are folded). Each row has a box to **name** the version, and **Keep** names the show file as it is saved now (for example "Opening night"); named versions are never pruned (`history/labels.json`).
- **Look** (writes the show file's `[settings]` block; refused while the Show file editor has unsaved changes): theme (defaults to pure black background / white text for accessibility; background/text/accent colors, system font choices, base size), speaker display (hidden / one speaker-name color / a different color for each character, chosen automatically in order of first appearance from a color-blind-friendly palette and adjusted to stay readable on the background; default: one color, white), idle message (overrides the config default). Live preview in a phone-shaped and a projector-shaped frame.
- **Test**: "Rehearsal mode" runs the operator view against an **isolated rehearsal engine** (doc 02) with a preview pane side-by-side. Nothing is sent to real audience devices and the live show is unaffected; position resets afterward.
- **Go Live**: shows audience URL (large), QR code (printable page for lobby signs), host IPs, current audience count, and a "Start Show" button that activates the show and opens `/operator`.

### Key hints
The operator page's buttons show their key as a keycap symbol and a name in the page's language: `␣ Space`, `⌫ Backspace`, `⇧ Shift+Q` (Spazio, Retroceso, Mayús+Q …). Keys are matched by the character typed (`event.key`), not by position, so QWERTY, AZERTY and QWERTZ layouts all work; the Mac's *delete* key is Backspace. The help topic *Keyboard* says so in every language.

### Interface language
Setup, the operator page and Help each have a language menu (a `<select>` in the header, and on the sign-in form) offering English, Español, Français, Deutsch, Italiano and Português. The first visit follows the browser's language; the choice is kept in the browser (`localStorage oc.ui.lang`) and `?ui=es` overrides it for one page. Changing it reloads the page in the new language (Setup asks first if the show file editor has unsaved changes). It is separate from the audience's language and from the show's languages.

Translation is done in the browser (`web/shared/i18n.js`): the English sentence is the key, `web/shared/i18n/<code>.js` holds each language's dictionary, and a sentence without an entry stays in English. Messages that come from the server (errors, notes about the show file) are translated the same way: a dictionary key may contain `{0}`, `{1}` placeholders, which match the text the server produced and are put back into the translation. The help topics (`web/help/<code>/<topic>.md`) and the two guides (`docs/i18n/<code>/`) are translated as whole documents and served by language (`?lang=` on `/api/help/topics`, `/api/help/search` and `/help/<topic>`); a topic without a translation would show in English. `test/e2e/check-i18n.js` lists every sentence the pages and server can show and fails if any language lacks one, so a new sentence cannot be added untranslated; the end-to-end test also walks every Setup step in each language looking for English left on screen.

### Settings (header button)
A window of separate sections, each with its own control, closed by a plain **Close** button (no section's button doubles as the window's main button):
- **About this copy** (read-only): the program version, the settings file, the data folder, the web ports, and **who may open Setup and the operator page** (with a warning when it is any computer on the network). It says these are fixed in `config.yaml` and need a restart to change.
- **Network name:** turn the `<name>.local` announcement on or off and choose the name (**Apply**); shows whether another computer already had the name.
- **Log detail:** how much is logged; takes effect as soon as it is chosen.
- **Change passwords:** new admin and/or operator password (at least 6 characters, different from each other; blank keeps one) with its own **Change passwords** button, and a button that shows the current operator password to hand over and the same button hides it again. People signed in with a changed password are signed out; the admin making the change stays signed in.

The Help button opens the help panel and F1 does the same. The Go Live step has a **Recent log…** link: a viewer for the last lines the program logged (the same 1000 it keeps for the support bundle), with a level filter (everything / normal and up / problems only / errors only), refresh every few seconds, and Copy.

### Network panel (always visible in the header)
`Audience URL: http://192.168.1.10/ ▸ copy ▸ QR` (plus the `https://` URL when HTTPS is enabled, with a note that a self-signed certificate shows a browser warning) with a warning banner if no non-loopback address is detected or a fallback port is in use.

### Guard rails
- While live, Setup tabs show a banner "Show is live" and actions that change the show file require typing the show name to confirm (REQ-052).
- Every destructive action uses an in-app confirm dialog, never silently.

## 2. Operator interface — `/operator`

Layout (desktop): three columns of the show file as a vertical list.

```
┌ header: Show name • ● LIVE • 83 viewers • Controller: you • [Takeover] ──────┐
│  SENT (dimmed, scrolls up)                                                   │
│   …                                                                           │
│   OSCAR: Felix, you're late.                                                  │
│  ─────────────── on screen now ───────────────                                │
│ ▶ FELIX: I'm so sorry,  (just sent)                                           │
│  ═════════════ NEXT (press Space) ═════════════                               │
│  ► (lights fade)                                                              │
│    [🖼 IMAGE: curtain.jpg]                                                     │
│    📑 Act 2, Scene 1                                                          │
│  UPCOMING (normal)                                                            │
│   …                                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│  [◀ Back]   [   NEXT ▶  (Space)   ]   [Blank (B)] [Clear (C)] [Message…]      │
│  Search/jump: [____________]   Mini-map of sections                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

Behavior:
- **Next cue is visually unmistakable** (highlighted block with large text). The list auto-scrolls to keep the next cue ~⅓ from the top.
- **Image and fullscreen cues** render as a labeled chip: `🖼 IMAGE: curtain.jpg`, `⛶ FULLSCREEN: INTERMISSION`, `🔁 LOOP (3 items, 10 s each)` (click to expand the item list; images shown by name only); no image rendering. Operator-only `note` and `section` cues are styled distinctly and are skipped over by Space (they don't consume a keypress, they auto-pass when reached).
- **Keyboard** (ignored when typing in an input):

  | Key | Action |
  |-----|--------|
  | `Space`, `→`, `Enter`, `PageDown` | **NEXT**: send the next caption; start a loop when it is next; end a running loop and send the cue after it (clickers typically send PageDown/Right) |
  | `Shift+Q` | **STOP**: stop a running loop and leave a blank screen (NEXT then continues normally) |
  | `Backspace`, `←`, `PageUp` | **Back**: undo the last action (NEXT, STOP, blank, clear, message); backing into a loop restarts it |
  | `B` | Toggle blank |
  | `C` | Clear screen |
  | `/` or `Ctrl+F` | Focus search/jump |
  | `Esc` | Cancel search / close dialog |
  | `Home` | Reset (confirm) |

  Space must call `preventDefault` to avoid page scroll. Key repeat is ignored; a command is debounced (~120 ms) to prevent double-fire, while the UI optimistically shows pending state and reconciles with the server response.
- **Safety**: clicking a cue in the list opens a "Jump here?" confirmation (single click select, Enter/second-click confirm) so accidental taps don't change the screen. Jumps never replay intermediate lines onto the audience view as a flood — they set the resulting state.
- **Loops**: while a loop runs, the header shows `🔁 LOOP RUNNING — item 2/4` with a STOP button; after STOP it shows `⬛ BLANK (loop stopped)`. The cue after the loop is already highlighted as NEXT. Back restarts a loop: after NEXT ended it, or after STOP, Back puts it back running from its first item (there is no separate restart button). A loop that is the first cue auto-starts when the show is activated, so the operator opens on `🔁 LOOP RUNNING` with the first caption as NEXT.
- **Mini-map and timing**: a **Go to section…** menu lists the show file's sections as jump targets (with the usual confirmation), and the header shows **Last sent n s ago** (turns amber after 2 minutes).
- **Status indicators**: connection state (green/yellow/red); audience count; warning if another session claims control; "Last sent n seconds ago".
- **Touch layout** (phone/tablet): large bottom Next/Back buttons; swipe not used (accidental). Wake lock requested.
- **Dark theme default** (theater is dark; low glare). The operator page uses its own login (operator password). Optional "stage-dark" red-tinted mode.
- **Ad-hoc message**: dialog to send one-off caption or fullscreen text (e.g., "Show paused — please remain seated") without editing the show file.

## 3. Audience interface — `/`

Principle: **no interaction needed.** Controls, if any, are hidden/minimal.

- **Captions surface**: full-viewport, pure black background with white text by default (high contrast, accessibility-first), large text. New captions appear at the bottom and the surface scrolls smoothly (instant if `prefers-reduced-motion`). Latest caption full brightness; the previous ~2 slightly dimmed; older progressively dimmer and eventually clipped off the top. Speaker label small-caps accent color above/before the text.
- **Re-reading:** the caption area scrolls (touch, wheel, keyboard). While the reader is scrolled up, new lines are added below without moving the view, older lines are shown at full brightness, and a **↓ Back to live (n)** button counts the lines that arrived; scrolling to the bottom or tapping it resumes following. Trimming old lines and snapshots keep the reader's place. A `?display=projector` screen is never scrollable. While following live, each new line **glides** into place over about 0.3 s rather than jumping (touching or wheeling the page stops it; `prefers-reduced-motion` turns it off).
- **Fullscreen overlay**: image (contain-fit, letterboxed on theme background) or centered large text, covering everything, with fade transition (≤ 250 ms). On the next caption it is removed and the scrolling surface returns with its prior content, unless a `[clear]` occurred.
- **Loops**: the audience page does not run loop timers. The server sends one `fullscreen` event per item (doc 04), and the page cross-fades between them; a reconnect snapshot carries the current item. Upcoming images are prefetched in the background (`preload`), since each device downloads full-size files.
- **Blank**: pure theme background.
- **Idle**: before the show starts, shows the configured message and a subtle "connected" indicator.
- **Connection**: a small unobtrusive status dot; if offline > 5 s shows "Reconnecting…" overlay while retaining the last text visible underneath, and auto-recovers on reconnect (re-snapshot). Reconnect backoff 0.5 s → 5 s with jitter.
- **Keep awake**: request Screen Wake Lock on load and on first user visibility change; where it is unavailable (plain HTTP, older iPhones) or refused, a silent 2-pixel video (`nosleep.mp4` / `nosleep.webm`, about 1 KB each, looping, muted, `playsinline`) keeps the screen on; browsers that need a tap to start it start it at the first touch. The page shows the one-time hint "Tip: turn off Auto-Lock to keep captions on" only if the video could not be started. `<html data-awake>` says which method is active (`wakelock`, `video`, `none` or `off`). A button in the display panel (the "Aa" button) turns *Screen stays on* off for that device (kept in the browser, translated into every supported language) for people on a low battery; turning it off releases the wake lock and pauses the video.
- **Language** (when the show has more than one language): the page starts in the language matched from the browser's `Accept-Language` header, or the default (`en`). A small picker in a corner lists the available languages; choosing one navigates to `/?lang=xx`, so the URL form (e.g. on a lobby QR code) works identically. With a single language there is no picker. The page never blocks on a choice.
- **Language menu** (🌐, top left, only when the show has several languages): lists the languages by their own names and navigates to `/?lang=xx`. The page's own labels are translated for es, fr, de, it and pt; other languages use English labels.
- **Optional local controls** (collapsed, appear on tap/click in a corner): text size − / +, light/dark toggle. Stored in `localStorage`; never required.
- **Projector mode** (`?display=projector`): font scale ×1.5 by default, hidden cursor, no controls, no connection dot unless offline.
- **Responsive**: font size uses `clamp()` with `vw`/`vh` units relative to the theme `scalePct`; works portrait and landscape; safe-area insets for iPhone.
- **Accessibility**: captions container `role="log" aria-live="polite"`; contrast ≥ 7:1 by default; no flashing.
- **Size budget**: audience page + JS + CSS ≤ 50 KB gzipped, no external requests.

## 4. Help (operators and admins)

Help is part of the product, not an afterthought (REQ-037). It is written for
volunteers: plain language, short, with examples.

**Access**
- Every `/admin` and `/operator` screen has a **Help** button in the header and responds to `?` / `F1`. It opens a viewer with a topic list, search, and the current topic's content; it is context-sensitive (opened from the Show file step → "Writing a show file").
- Both pages link to the same viewer (`/help`), with the topic list ordered for the user's role.
- Works offline; content is embedded in the binary.
- **Operator safety:** while the help viewer is open on `/operator`, show-control keys (Space, Shift+Q, etc.) are disabled, so reading help can't advance the show. A banner reminds the operator "Keys paused — close help (Esc) to resume". Help opens as a panel beside the show file where space allows so the operator can keep an eye on the show.

**Format and source.** Topics are Markdown files in `web/help/`, rendered with the same renderer as show files. Examples are fenced code blocks with **Copy** (and, in the show file editor, **Insert at cursor**) buttons.

**Topics for admins**
1. *Getting started* — the six wizard steps; how to find the audience URL/QR; where passwords are (`config.yaml`).
2. *Writing a show file* — the Markdown conventions (below).
3. *Images* — formats, sizes, the 1 MB warning, naming, what "missing image" means.
4. *Loops: pre-show and intermission* — syntax, durations, how the operator starts/stops them.
5. *Look and feel* — themes, contrast, speaker colors, projector mode.
6. *Languages* — translated show files, identical structure, how audiences pick.
7. *Network and joining* — URL, QR, HTTPS caveats, venue prerequisites.
8. *Troubleshooting* — port in use, nobody can connect, images slow, show file warnings explained.

**Topics for operators**
1. *Running the show* — one-page summary: NEXT, Back, STOP, what the highlighted line means.
2. *Keyboard reference* — the table in section 2 (generated from the same key map the page uses, so it can't drift).
3. *Loops* — how they start, STOP (blank), NEXT (continue), Back (restart).
4. *Indicators* — connection dot, audience count, controller/takeover, image/loop chips.
5. *If something goes wrong* — wrong line sent (Back), lost connection, someone else took control, audience can't see captions, send an ad-hoc message, blank/clear.

**"Writing a show file" must cover every show file convention in doc 03**, each with a short example and the result:
- one line = one caption (and wrapping; you decide where to break lines)
- `**bold**`, `*italic*`, `(stage directions)`, `NAME: speaker` labels, `\` line continuation and `[[ … ]]` multi-line captions
- comments `<!-- … -->` (never shown) versus `[note: …]` (operator only); `#` is reserved for future headings
- `[clear]`, `[blank]`, `[image: file | caption]`, `![caption](file)`, `[fullscreen: text]`, `[section: name]`
- `[loop] … [/loop]` with durations
- `[language: xx | Name]`
- escaping (`\*`, `\#`, `\[`, `\:`) and that raw HTML isn't supported
- a "what the warnings mean" list mirroring the diagnostics
- a link to download the sample show file

**As built.** Help opens as a side panel (an embedded viewer at `/help`) from a **Help** button, `?`, or `F1`; on Setup the topic follows the current step. On `/operator`, while the panel is open every show-control key is ignored except `Esc`/`?`/`F1`, which close it. The audience page has an optional, low-key **Aa** control (top right) for per-device text size and dark-on-light, stored only on that device.

**Keeping help correct.** The parser's directive/syntax registry is the single list of conventions; a unit test fails if any directive, escape, or diagnostic code has no mention in the help topics, and another parses every fenced example in help to check it produces no errors.

## Visual design

Clean, minimal, theater-friendly: pure black backgrounds, high-contrast
white text (colors are configurable in the Look step). Setup UI is lighter and friendlier (light theme
default) and uses plain-language labels, not jargon (say "Captions", not "cues", in the UI).
