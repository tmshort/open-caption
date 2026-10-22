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
2. Wizard stepper across the top: **① Shows → ② Show file → ③ Images → ④ Look → ⑤ Test → ⑥ Share → ⑦ Report**. The Report step is for after the show: it does not need a show to be selected, and has no tick.
   Steps are also reachable directly for experienced users. Each step shows a green check when complete.

### Steps
- **Shows**: name; list of existing shows with Rename… / Duplicate… / Export / **Archive** / Delete (delete asks first). The show on air cannot be archived, and a show that is live (a line has been sent) cannot be deleted: Delete explains why in a window the user has to dismiss (not a toast), and its tooltip says so. Archived shows are listed apart under "Archived shows (n)" with Restore / Export / Delete; they are not offered for the other steps and cannot be activated. The archived state is the `archived: yes` line in the show file's `[settings]`, so it travels with an export (an import is always restored).
- **Show file**: (a **Help** link here opens the "Writing a show file" topic). **Language tabs** sit above the editor: the main language, each translation, and ＋ Add a language. A ⚠ marks a translation held back for not matching the main show file; translations show the main show file for reference, and structure errors appear in the problem list.
  - The Show file step uses the whole width of the window: the editor has a minimum width (28 rem) and no maximum, taking everything the side panel (problems and cue list, up to 24 rem) does not need, and its height follows the window (62 % of its height, at least 22 rem).
  - Drag-and-drop file, **Upload a show file…** (a button that opens the computer's file chooser; any `.txt` / `.md`, any name; UTF-8 or Windows-1252), or paste into an editor. Opening a file only fills the editor (and a toast says so): the problems list updates and nothing is stored until Save, which keeps the previous version in history. **Download the show file** saves the current file (for the language tab chosen) as a plain `.md` to edit in any program; this is the offline round trip.
  - The Shows step's **Import a show…** button (drawn like **Upload a show file…**; it opens the file chooser) takes a `.zip` or a plain `.md` / `.txt` show file and makes a new show. With no shows yet the page is titled **Add a show** and offers Create show, Create a sample show and Import a show…; once there are shows it is titled **Choose a show** and the same three sit under *Add another show*.
  - Editor pane (monospace, line numbers) + Diagnostics pane. Clicking a diagnostic jumps to the line.
  - Live preview of the parsed cue list (kinds shown as icons: 💬 caption, 🖼 image, ⛶ fullscreen, 🔁 loop, ⬛ blank, 🧹 clear, 📑 section, 📝 note).
  - Help for the show file is the header **Help** button (it opens "Writing a show file" on this step); there is no separate Show file help button. The editor's *Show file cheat sheet* summarizes the syntax with copy-able examples.
  - Insert-directive toolbar buttons (Clear, Blank, Image…, Fullscreen text, Section, Loop, 🎵 Song, and ♪ / ♫ to insert music notes at the cursor) so authors never type brackets by hand.
  - Re-import warns if cue count differs and the show is live.
- **Images**: PNG/JPG/GIF/WebP only (no SVG); files over 1 MB get a "large — may load slowly for many devices" warning, since images are sent unscaled. Grid of uploads with thumbnails; drag-drop multi-upload, or the **Upload images…** button (and **Upload this image…** beside each missing one); no page shows the browser's own "Choose File" control, every file chooser is a button that opens it (a hidden `input[type=file]`, checked by the end-to-end test); shows "used in lines 12, 40" or "unused"; missing-but-referenced images listed at top with an upload-in-place button.
- **Earlier versions** (button above the editor): every save keeps the version it replaced (the last 50 per language, in `<show>/history/`, not exported or duplicated). The dialog lists them by time and size; *Load* only fills the editor, so nothing changes until Save (and the live-show and conflict checks still apply). *Compare* shows, line by line in the browser, what loading that version would add and remove compared with the editor (long unchanged runs are folded). Each row has a box to **name** the version, and **Keep** names the show file as it is saved now (for example "Opening night"); named versions are never pruned (`history/labels.json`).
- **Look** (writes the show file's `[settings]` block; refused while the Show file editor has unsaved changes): theme (defaults to pure black background / white text for accessibility; background/text/accent colors, system font choices, base size), speaker display (hidden / one speaker-name color / a different color for each character, chosen automatically in order of first appearance from a color-blind-friendly palette and adjusted to stay readable on the background; default: one color, white), idle message (overrides the config default). Live preview in a phone-shaped and a projector-shaped frame. Layout: the controls are on the left in two columns (Background and Text side by side with the contrast ratio straight under them; then Speaker name color and Font; Text size, Speaker names and the pre-show message across the full width, the messages for every language one after another), and the Phone and Projector previews are stacked in a narrow column on the right; below 1000 px everything is one column.
- **Test**: "Rehearsal mode" runs the operator view against an **isolated rehearsal engine** (doc 02) with a preview pane side-by-side. Nothing is sent to real audience devices and the live show is unaffected; position resets afterward.
- **Share**: shows audience URL (large), QR code (printable page for lobby signs), host IPs and current audience count. It no longer starts anything in front of the audience: the button at the top, on the right of the *Ready?* checklist, is **Open the Run page**, which loads the show **in standby** (the audience keeps seeing the venue's standby screen) and goes to `/operator`; a different show that is on air is replaced only after asking. When the show is already loaded the button is a plain link, with **Restart show…** beside it (back to the beginning and to standby, confirmed when on air), and the hint says whether it is on air or waiting in standby. **Print a QR sign for the lobby** is a button, like Copy address. The addresses are to the **right of the QR code** (below it on a narrow screen), one per row, each row `[radio] link (Ethernet|Wi-Fi)` plus what belongs to that address; the `<name>.local` row shows only while the name is being announced (its on/off box and its name are in Settings → Network), with a note for phones that cannot open it, and a warning under it when the name was taken or could not be announced. The projector row lines up with the others. Under the address card a hint links to **Settings → Network** ("Two networks, or the wrong address?"); when every network interface is turned off the step says so. The list of audience addresses ends with a **projector link** (`<this computer's address>/?display=projector`, in a new window): the audience page for a shared screen (larger text, no controls, no pointer), to open on the computer the projector is plugged into, so it uses this computer's own address.

### Key hints
The operator page's buttons show their key as a keycap symbol and a name in the page's language: `␣ Space`, `⌫ Backspace`, `⇧ Shift+Q` (Spazio, Retroceso, Mayús+Q …). Keys are matched by the character typed (`event.key`), not by position, so QWERTY, AZERTY and QWERTZ layouts all work; the Mac's *delete* key is Backspace. The help topic *Keyboard* says so in every language.

### Interface language
Setup, the operator page and Help each have a language menu (a `<select>` in the header, and on the sign-in form) offering English, Español, Français, Deutsch, Italiano and Português. The first visit follows the browser's language; the choice is kept in the browser (`localStorage oc.ui.lang`) and `?ui=es` overrides it for one page. Changing it reloads the page in the new language (Setup asks first if the show file editor has unsaved changes). It is separate from the audience's language and from the show's languages.

Translation is done in the browser (`web/shared/i18n.js`): the English sentence is the key, `web/shared/i18n/<code>.js` holds each language's dictionary, and a sentence without an entry stays in English. Messages that come from the server (errors, notes about the show file) are translated the same way: a dictionary key may contain `{0}`, `{1}` placeholders, which match the text the server produced and are put back into the translation. The help topics (`web/help/<code>/<topic>.md`) and the two guides (`docs/i18n/<code>/`) are translated as whole documents and served by language (`?lang=` on `/api/help/topics`, `/api/help/search` and `/help/<topic>`); a topic without a translation would show in English. `test/e2e/check-i18n.js` lists every sentence the pages and server can show and fails if any language lacks one, so a new sentence cannot be added untranslated; the end-to-end test also walks every Setup step in each language looking for English left on screen.

### Settings (header button)
A window of separate sections, each with its own control, closed by a plain **Close** button (no section's button doubles as the window's main button). It has **four tabs** (Audience screen, **Network**, Security, About) (an ARIA tab list: click, or the arrow keys, Home and End), remembers the last one used, and has a fixed height: the title, the tab row and **Close** stay in view while a tab scrolls, and every tab stays in the page, so text typed but not yet saved survives switching tabs. Every dialog in Setup keeps its title and buttons in view and scrolls only the middle. The tabs: **Audience screen** (Standby screen), **Security** (the open-WiFi warning, Change passwords) and **About** (the About table, Log detail). All the panels share one grid cell (the unselected ones only invisible), so the window is as tall as the tallest tab, the About tab fits without scrolling wherever the screen allows, and it never changes size between tabs. The sections:
- **Network interfaces** (the *Network* tab; a setting of this computer, not of a show, so it is not on the Share step): one row for each connected network interface, with a **tick** (a tick is saved the moment it is clicked; a toast confirms it), the interface's name and kind (*Ethernet*, *Wi-Fi*), its addresses and what is done with it: *used*, *not used*, or, when two ticked interfaces are on the same network, *not used: same network as <interface> (<address>)* (the wired (Ethernet) one is used whatever its address, else the main connection's; untick the other to choose). The interface Setup is being used through says so and its tick is disabled (the program also refuses, 409). A box **Also refuse connections to interfaces that are not ticked** (with a hint that this computer itself is always served) saves at once too. Until a first change every interface is ticked; the first change records the ticked names in `config.yaml`, after which a new interface stays off until it is ticked. The explanation, an error under the list (for example when nothing would be ticked) and, when the program was started with `--network-interfaces`, a note that it wins at the next start, are text in the tab. When Settings is closed on the Share step it redraws.
- **About this copy** (read-only): the program version, the settings file, the data folder, the web ports, and **who may open Setup and the operator page** (with a warning when it is any computer on the network). It says these are fixed in `config.yaml` and need a restart to change.
- The **network name** (`<name>.local`) is in the *Network* tab, under **Name on the network** (it is also a network setting): a **Use a name** checkbox that turns the announcement on or off at once, the name (`http://` `<name>` `.local`) with an **Apply** button, and, below, *Devices on this network can open http://<name>.local/* while it is being announced or a warning when the name was taken (another computer already uses it) or could not be announced. The **Share** step lists the name only as a plain link row (radio and link) while it is announced, with a note for phones that cannot open it, and no controls of its own.
- **Log detail:** how much is logged; takes effect as soon as it is chosen.
- **Standby screen:** the venue's standby screen, shown to the audience while no show is on air: a **logo**: the built-in **[oc]** mark (the favicon's art) until you upload your own (PNG, JPEG, GIF or WebP, up to 2 MB; shown as it will look; **Remove logo** brings the built-in one back) and an optional one-line **message** (up to 200 characters; empty uses "Captions will begin shortly" in each viewer's language), each saved as soon as chosen / with **Save message**. Venue-level, never in a show file.
- **Change passwords:** new admin and/or operator password (at least 6 characters, different from each other; blank keeps one) with its own **Change passwords** button, and a button that shows the current operator password to hand over and the same button hides it again. People signed in with a changed password are signed out; the admin making the change stays signed in.

**Audience report (REQ-070).** When the show reaches its end the operator page's banner reads *End of the show.* followed by an **Audience report** link (a new window; not shown in a rehearsal). The page `/operator/report` (sign-in like the operator page) shows: the show's name; a status line (*The show is on air now: these are the numbers so far (from …)*, or *Final report: from … to …*); a row of cards (**Devices**, **Phones**, **Tablets**, **Computers**, **Projector screens**, **Most at once** with its time, **Connected now** / **Still connected at the end**, **Connections** with how many were reconnections, **Average time connected**, **Longest stay**); a note that devices are told apart by address and browser type, so it is an estimate; the graph **Devices over time** (a step line with the area under it: the vertical axis counts whole devices, the horizontal one is the clock with the start and the end (or *Now*, while the show is on air) labelled on it; a dot marks the busiest moment; the stretches during which a loop was running are hatched bands named as the show file names the loop (`name`), else *Pre-show*, *Intermission* or *Post-show*, with a legend; it scrolls sideways, reachable with the keyboard, on a narrow screen and fills the page when printed) and, for those who cannot tell the colours apart or use a screen reader, a **Loops** table (loop, from, to, duration); a **Languages** table; and a **Devices** table (device code, address, type, languages, first seen, last seen, connections, time connected, longest stay, at the end; it scrolls sideways on a narrow screen and can be reached with the keyboard). A **Report** menu opens the run in progress and earlier saved ones, **Refresh**, **Print** (the controls are left out of the print-out, which is black on white) and **Download CSV** are buttons, and **Back to Run** returns to the operator page. The page refreshes itself every few seconds while the show is on air. Setup shows the same report as its seventh step, **Report**: a heading, a toolbar (the **Report** menu, **Refresh**, **Print**, **Download CSV**, **Delete…** (for a saved report, after asking; the run in progress cannot be deleted) and a link **Open the report page** for the stand-alone page), and the report below; **Print** prints the report alone, without Setup's header and steps. It refreshes itself every few seconds while the show is on air. The **Share** step shows a hint link *Audience report (how many devices used the audience page during the last run of the show)* to it once a report exists.

The Help button opens the help panel and F1 does the same. The Share step has a **Recent log…** link: a viewer for the last lines the program logged (the same 1000 it keeps for the support bundle), with a level filter (everything / normal and up / problems only / errors only), refresh every few seconds, and Copy.

### Network panel (always visible in the header)
`Audience URL: http://192.168.1.10/ ▸ copy ▸ QR` (plus the `https://` URL when HTTPS is enabled, with a note that a self-signed certificate shows a browser warning) with a warning banner if no non-loopback address is detected or a fallback port is in use.

### Guard rails
- While the show is on air, Setup's header shows the red ON AIR pill, and saving a change to the show file asks "This show is on air" and requires typing the show name to confirm (REQ-052).
- Every destructive action uses an in-app confirm dialog, never silently.

## 2. Operator interface — `/operator`

Layout (desktop): three columns of the show file as a vertical list.

**One interface, two palettes.** Setup and the operator page share `web/shared/ui.css`: the same type (16 px, headings 20/19/16 px), links, buttons (one primary style: the accent colour, with `--accent-fg` text), text boxes and menus (2.4 rem high), focus ring (3 px accent), hints (14.4 px), disabled look, sign-in panel, dialogs, help panel and toast. Each page supplies only its palette; the operator page is always dark (a booth), uses Setup's dark palette with a darker page background, and keeps yellow (`--mark`) for the next line, speaker names and warnings. Its one deliberate difference is the **stage scale**: the cue list (18 px), the banner and the footer controls are a size larger than the interface around them.

**Operator status strip.** Who is in control, the Take over button and "last sent" sit in a strip under the operator header, so the header itself stays on one row. The strip turns amber when another operator is in control. Messages about a problem (not in control, out of date, connection trouble) stay on screen for six seconds; ordinary confirmations for two and a half.

**Progress and the way on.** The status strip also shows "Item n of total" (only items the audience sees; notes and section headings are not counted), or the item count while in standby. Under every Setup step but the last, a **Next: <step> →** button leads on (highlighted when the current step is complete). When the program gives no reason for a failure, both pages say what the status means and what to do (not allowed here, file too large, too many tries, program problem, check the connection) instead of "Something went wrong". The Share step tells people that a phone that cannot open the `.local` name should use the numeric address or the QR code, and the operator help has a short glossary of standby, on air, off air, restart, blank and clear.

**Announcements.** While the operator's one-off announcement is on the audience's screens, the operator page shows a blue banner quoting it ("ANNOUNCEMENT ON SCREEN: …"); Back or the next line removes both. After choosing a section in **Go to section…** the menu gives the keyboard back to the page, so Enter confirms the jump and Esc cancels it.

**Show actions.** Each show in the list has **Continue with this show** (primary) and a **More ▾** menu holding Rename…, Duplicate…, Export, Archive and Delete… (Delete last, separated and red; archived shows have Export and Delete…). The menu opens with Enter, Space or Down on the button, moves with the arrow keys, Home and End, and closes with Esc (focus returns to the button), Tab or a click elsewhere. An action that is refused stays listed and says why under its name (for example Archive on the show that is on air). Export is fetched and saved by the page (not a plain download link), so a failure is explained in a message — the program's reason, or "Could not reach the program…" — rather than the browser's own "download failed".

**Share step.** The button that opens the operator page is named **Open the Run page** (the page is called Run everywhere). The projector link is always the computer's numeric address (never the `.local` name), so it can be copied; like the other addresses it has a radio button that chooses it for **Copy address** and the QR code. **Print a QR sign for the lobby** sits right beside **Copy address**.

**Words.** The interface uses one pair of words for the show's state: **STANDBY** (loaded; the audience sees the standby screen) and **ON AIR** (the audience sees the show). **Go to standby…** returns to it; **Restart show…** goes to the start and to standby. "Live" is not used for the state. Three kinds of text are named apart: the **Standby message** (the venue's, in Settings), the **Pre-show message** (a show's, in Look; the `idle message` setting) and an **Announcement…** (the operator's one-off notice).

**Shortcut hints.** Setup's Help button says F1 and its Save button in the show-file editor says Ctrl+S / Cmd+S in their tooltips, as the operator page's buttons name their keys.

**Connection notice.** If a page cannot reach the program for more than two seconds (the live stream dropped and the retries fail), both Setup and the operator page show a red bar under the header, "Connection to the program lost. Reconnecting…", in addition to the red dot; it goes away by itself when the stream is back. (An ordinary one-second reconnect never flashes it.) A button in Setup pressed while the program is not answering shows "Could not reach the program. Check the connection, then try again." (rather than leaving the page unchanged); the operator page, opened while it is not answering, shows the bar and keeps trying.

**Sign-in and titles.** The two sign-in panels have the same layout: language menu, "Open Caption — Setup" / "Open Caption — Run", one sentence saying which password to enter (the admin password works on the operator page too), the password box and a primary **Sign in**. The operator page's tab title is "Run — <show> (ON AIR | STANDBY)" and Setup's is "Setup — <step>". The same action has the same name on both pages: **Restart show…** (it asks first), **Create a sample show**.

**Headers.** Both headers use one stylesheet, `web/shared/header.css`, so type size (16 px, titles 20 px bold), spacing, the status pills (audience count, ON AIR/STANDBY), the quiet links and buttons (2.4 rem high), the language menu and the connection dot are identical (a browser test compares them). The right side of the Setup and operator headers is laid out the same way, left to right: audience devices (👥 n), language, the page's own actions, a link to the other page, Help, Sign out, and the connection dot (green while the page is receiving updates from the server, red otherwise). Setup: devices, language, **Settings**, **Run** (the operator page), Help, Sign out, dot. Operator: devices, language, **Go to standby…**, **Restart show…**, **Setup** (back to Setup), Help, Sign out, dot; the left side holds the show name, the STANDBY/ON AIR badge, who is in control, last sent and Take over.

**Setup link (REQ-067).** A session with the admin role (the admin password, or a browser on this computer) shows the **Setup** link, back to `/admin`; a plain operator session and the rehearsal frame do not. On this computer there is no password to sign out of, so **Sign out** is hidden (and the login form never appears).

**Audience count.** In both headers the count (`👥 5 (en 3 · es 2)`) is a button. A language nobody is using is left out of the brackets (and with nobody connected there are no brackets). Clicking it opens **Connected devices**, a table of **Device** (`Phone (iPhone, Safari)`, `Computer (Mac, Chrome)`, `Projector screen`; the kind is translated, brand names are not), **IP address** and **Language**, refreshed every 3 seconds while it is open. The label and table are `web/shared/clients.js`, shared by both pages. The ON AIR pill is the same red on both pages (`header.css` fixes the colour; each page's own `--bad` differs).

**Standby and on air (REQ-066).** A loaded show opens in **standby**: the header badge says `STANDBY` (blue; `● ON AIR` once on air), a blue banner says `STANDBY — the audience sees the standby screen. Press Space (START) to put the show on air; it carries on from where it is.`, and the **NEXT ▶** button becomes **START ▶** in the same place with the same key (Space and the other advance keys, or a click). Back, Blank, Clear, Announcement and Stop are disabled and their keys do nothing; search, Restart and Help still work. **Go to standby…** (header, disabled in standby) asks first, then returns the audience to the standby screen with the position kept; START resumes. **Restart show…** goes back to the beginning *and* to standby; it is disabled (as is Home, and Setup's **Restart show…**) while the show is already at the start in standby with nothing sent (`atStart`, `fresh`), since there is nothing to restart.

```
┌ header: Show name • ● ON AIR • 83 viewers • Controller: you • [Takeover] ──────┐
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
│  [◀ Back]   [   NEXT ▶  (Space)   ]   [Skip (S)] [Blank (B)] [Clear (C)] [Announcement…]      │
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
  | `Backspace`, `←`, `PageUp` | **Back**: undo the last action (NEXT, STOP, blank, clear, an announcement); backing into a loop restarts it |
  | `S` | **Skip**: pass over the next line without sending it (the audience sees nothing; a toast names the line; Back undoes it). Not available in standby |
  | `B` | Toggle blank |
  | `C` | Clear screen |
  | `/` or `Ctrl+F` | Focus search/jump |
  | `Esc` | Cancel search / close dialog |
  | `Home` | Reset (confirm) |

  Space must call `preventDefault` to avoid page scroll. Key repeat is ignored; a command is debounced (~120 ms) to prevent double-fire, while the UI optimistically shows pending state and reconciles with the server response.
- **Safety**: clicking a cue in the list selects it and shows a bar, "Jump so that line n is next: …", with a **Jump here (Enter)** button (single click selects; Enter or a second click confirms) so accidental taps don't change the screen. Jumps never replay intermediate lines onto the audience view as a flood — they set the resulting state.
- **Loops**: in the list, the item of a running loop that is on the audience screens now has a small green ▶ beside it (bold text, `aria-current`), moving along as the loop does. While a loop runs, the header shows `🔁 LOOP RUNNING — item 2/4` with a STOP button, and, when the loop has a timer, `⏱ 13 min left on the timer` counting down (the button is always on the page, in the same place, and is simply disabled when no loop is running, so Back, Next and the rest never move during the show); after STOP it shows `⬛ BLANK (loop stopped)`. The cue after the loop is already highlighted as NEXT. Back restarts a loop: after NEXT ended it, or after STOP, Back puts it back running from its first item (there is no separate restart button). A loop that is the first cue auto-starts when the show is activated, so the operator opens on `🔁 LOOP RUNNING` with the first caption as NEXT.
- **Mini-map and timing**: a **Go to section…** menu lists the show file's sections as jump targets (with the usual confirmation), and the header shows **Last sent n s ago** (turns amber after 2 minutes).
- **Status indicators**: connection state (green/yellow/red); audience count; warning if another session claims control; "Last sent n seconds ago".
- **Touch layout** (phone/tablet; the operator page adapts to small screens but is not recommended for running a show, see REQ-035): large bottom Next/Back buttons; swipe not used (accidental). Wake lock requested.
- **Dark theme default** (theater is dark; low glare). The operator page uses its own login (operator password). Optional "stage-dark" red-tinted mode.
- **Announcement…** (an ad-hoc message): dialog to send one-off fullscreen text (e.g., "Show paused — please remain seated") without editing the show file.

## 3. Audience interface — `/`

Principle: **no interaction needed.** Controls, if any, are hidden/minimal.

- **Captions surface**: full-viewport, pure black background with white text by default (high contrast, accessibility-first), large text. New captions appear at the bottom and the surface scrolls smoothly (instant if `prefers-reduced-motion`). Latest caption full brightness; the previous ~2 slightly dimmed; older progressively dimmer and eventually clipped off the top. Speaker label small-caps accent color above/before the text.
- **Re-reading:** the caption area scrolls (touch, wheel, keyboard). While the reader is scrolled up, new lines are added below without moving the view, older lines are shown at full brightness, and a **↓ Back to live (n)** button counts the lines that arrived; scrolling to the bottom or tapping it resumes following. Trimming old lines and snapshots keep the reader's place. A `?display=projector` screen is never scrollable. While following live, each new line **comes up from below the bottom edge** with everything above it moving up together over about 0.3 s, as if the whole column were scrolling (the caption column is clipped at its bottom padding, so a line that has not scrolled up yet is out of sight; it never pops on in place) (when the lines do not yet fill the screen and sit at the bottom, the older lines slide up by the new line's height in the same time) (touching or wheeling the page stops it; `prefers-reduced-motion` turns it off).
- **Fullscreen overlay**: image (contain-fit, letterboxed on the theme background; a picture **with no caption fills the whole screen edge to edge**, with no margin, while one with a caption keeps its margin and shows the caption beneath it) or centered large text, covering everything, with fade transition (≤ 250 ms). On the next caption it is removed and the scrolling surface returns with its prior content, unless a `[clear]` occurred.
- **Loops**: the audience page does not run loop timers. The server sends one `fullscreen` event per item (doc 04), and the page cross-fades between them; a reconnect snapshot carries the current item. Upcoming images are prefetched in the background (`preload`), since each device downloads full-size files. The one exception is the **countdown**: an item that contains `{timer}` carries the end time (`deadline`), and the page (`web/audience/audience.js`, `paintTimers`) updates the `[data-timer]` elements four times a second in whole minutes, rounded up (`15 min`, `1 hr 5 min`; no seconds) using the server's clock (snapshot `ts`) so a phone with a wrong clock is still right. It stays at `0 min` when it reaches zero.
- **Blank**: pure theme background.
- **Standby** (`mode: standby`): the venue's logo (the built-in [oc] mark unless replaced) and message (or the built-in words in the viewer's language), full screen on the show's background colour; nothing of the show, ever. This is what the audience sees before the operator presses START, after Go to standby, and when no show is loaded.
- **Idle**: once the show is on air and before its first line (no leading loop), shows the show's pre-show message (`idle message`) and a subtle "connected" indicator.
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
- **Operator safety:** while the help viewer is open on `/operator`, show-control keys (Space, Shift+Q, etc.) are disabled, so reading help can't advance the show. An amber bar at the top of the help panel reminds the operator "Keys are paused. Close help to resume." beside a **Close (Esc)** button. Help opens as a panel beside the show file where space allows so the operator can keep an eye on the show.

**Format and source.** Topics are Markdown files in `web/help/`, rendered with the same renderer as show files. Examples are fenced code blocks with a **Copy** button.

**Topics for admins** (in the order of the list, which opens on the first)
1. *Getting started* — the six wizard steps; how to find the audience URL/QR; where passwords are (`config.yaml`).
2. *Preparing a show file* — for whoever turns the script into captions, including editing the file in their own program and checking it offline (`--check-show`).
3. *Writing a show file* — the Markdown conventions (below).
4. *Images* — formats, sizes, the 1 MB warning, naming, what "missing image" means.
5. *Look and feel* — themes, contrast, speaker colors, projector mode.
6. *Languages* — translated show files, identical structure, how audiences pick.
7. *Loops: pre-show and intermission* — syntax, durations, timers, how the operator starts/stops them.
8. *Musicals* — lyrics, voices together, clearing between numbers.
9. *Network and joining* — URL, QR, the `.local` name, HTTPS caveats, venue prerequisites.
10. *Troubleshooting* — port in use, nobody can connect, images slow, show file warnings explained, the log and the support bundle.

**Topics for operators**
1. *Running the show* — one-page summary: NEXT, Back, STOP, what the highlighted line means, the words used on the page.
2. *Keyboard reference* — the table in section 2 (generated from the same key map the page uses, so it can't drift).
3. *Indicators* — connection dot, audience count, controller/takeover, banners, image/loop chips.
4. *Loops for the operator* — how they start, STOP (blank), NEXT (continue), Back (restart).
5. *If something goes wrong* — wrong line sent (Back), lost connection, someone else took control, audience can't see captions, send an announcement, blank/clear.

**Guides and reference.** After the topics, under their own heading (**Guides and reference**), the longer documents from `docs/` that ship inside the program (`docs/embed.go`): the *usage guide*, the *venue network guide*, the *trusted HTTPS certificate guide* and the *`config.yaml` reference*, each translated into the five other languages (`docs/i18n/<code>/`). The order of all of it is `topicOrder` in `internal/help`; a test fails when a topic is not in that list. The manual page and the certificate design notes (doc 10) are English only and are not in the app.

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
