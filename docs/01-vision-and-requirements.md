# 01 — Vision and Requirements

## Vision

A community theater can set up open captioning for a show in under an hour,
with volunteers who are not technical, and run it from one laptop. Audience
members join the guest WiFi, open one URL, and read the show's captions in sync
with the stage. Nothing to install, nothing to click.

## Users

| User | Description | Key needs |
|------|-------------|-----------|
| **Admin / Setup user** | Several volunteers over time who prepare a show | Easy, forgiving, guided; clear errors when a show file is malformed |
| **Operator** | Follows the show (show file/prompt book) and releases captions | One-key advance, confidence about position, can't get lost, can recover from mistakes fast |
| **Audience member** | BYOD phone/tablet/desktop, or a shared projector | Zero interaction, readable, resilient to WiFi drops and sleeping screens |

## Scope

In scope: one server process on one host; show file import; live operation;
audience display; image, fullscreen and loop (pre-show/intermission) support.

Out of scope (v1): cloud hosting, user accounts, automatic speech
recognition / auto-follow of the show, audio, multiple
simultaneous shows, internet-facing deployment.

## Functional requirements

### Platform and deployment
- **REQ-001** Single Go binary, no runtime dependencies, no CGO; builds for macOS (arm64/amd64), Windows (amd64), Linux (amd64/arm64).
- **REQ-002** All web assets embedded in the binary (`go:embed`). Nothing fetched from the internet at runtime (the venue network may have no uplink).
- **REQ-003** The whole system (setup, operator, audience) runs on a single host.
- **REQ-004** Persistent data lives in one data directory (configurable); no external database.
- **REQ-005** Settings (ports/addresses, TLS, passwords, defaults) live in a human-editable `config.yaml` (generated on first run); flags/env vars override it. The file is validated when the program starts: a misspelled or unknown setting, a value of the wrong kind or out of range, a bad address, an unreadable certificate and similar are all reported together, each with its line or setting name, and the program does not start until they are fixed. A generated file starts with `config_version: 1`; a file marked with a newer version than the program knows starts anyway, ignoring (and warning about) settings it does not recognise, so going back to an older program does not strand a venue. Things that work but look unintended (short or equal passwords, an empty access list) are logged as warnings.
- **REQ-009** The server announces itself on the local network as `<name>.local` using multicast DNS (default `open-caption.local`), after checking that no other computer has the name; the admin can switch it off or choose another name in Setup. The numeric address and QR code remain the primary way to reach the page, because some networks block multicast.
- **REQ-006** The server listens on plain HTTP and, optionally, HTTPS (user-supplied cert or generated self-signed) at the same time. Automatic certificates (ACME/Let's Encrypt) are a possible later addition; they need Internet access and a public name and are out of scope for v1.
- **REQ-008** All configuration of a show — its title, look, pre-show message(s), loop time, scrollback size and main language, as well as its lines — is stored in the show file(s) (a `[settings]` block); only the show's images are stored separately. Defaults are layered: built-in values, overridden by `config.yaml`, overridden by the show file. Environment settings (ports, HTTPS, passwords, access lists, data folder) are in `config.yaml` only.
- **REQ-007** Licensed under Apache-2.0; dependencies must be license-compatible.

### Capacity and resilience
- **REQ-010** Support at least 200 concurrent audience clients (venue has ~150 seats; requirement is >75). Caption latency from operator keypress to render ≤ 250 ms at p95 on a LAN under that load.
- **REQ-011** Audience clients reconnect automatically after network loss or device sleep and show the *current* state without user action.
- **REQ-012** Server restart resumes at the previously active show, cue position and on-screen state; a loop that was running restarts from its first item.
- **REQ-013** A client joining mid-show immediately sees current state plus recent scrollback.

### Show file
- **REQ-020** Import a Markdown/text show file (`.md` or `.txt`, both parsed as Markdown, UTF-8) via upload or paste; re-import without losing assets.
- **REQ-021** Each non-blank show file line is a caption (long lines wrap; the author controls splitting); non-technical authors can add human-readable directives (clear screen, images, fullscreen text, loops) (doc 03).
- **REQ-022** Formatting is a Markdown subset (bold/italic, images). Raw HTML is **not** supported; it is displayed as literal text with a diagnostic. Exception: `<!-- -->` comments. `[note: …]` is an operator-only note; `#` is reserved for future headings.
- **REQ-023** Import validates the show file and reports errors/warnings with line numbers *without* blocking the preview.
- **REQ-024** Images (PNG/JPG/GIF/WebP; no SVG) are uploaded/managed in the setup UI and referenced by filename. Missing references are flagged. The server does not rescale; browsers scale to fit. Audience pages prefetch upcoming images.
- **REQ-025** **Loops**: a block of mixed fullscreen items (images/text), each shown for its own duration (default from config), repeating. Used for pre-show and intermission. A loop that is the first cue auto-starts on activation; any other loop starts when the operator presses NEXT on it.
- **REQ-027** Musicals: lyrics can be marked with a `~` shorthand or a `[song]` block, are shown in italics between ♪ notes, and music notes can be inserted from the editor (doc 03, *Songs and lyrics*). Two or more voices sung or spoken at the same time (duets, counterpoint) are shown together from one operator keypress, using `[together]` or ` || ` (doc 03, *Voices at the same time*).
- **REQ-026** Multiple languages: one show file per language, identical in structure (checked cue by cue), each identified by a `[language: code | Name]` line. One operator keypress sends the same cue to every language. The audience's language is detected from `Accept-Language` (default: the main language), with a picker and a `?lang=` URL form. A translation that does not match the main show file is reported and not shown to the audience.

### Operator
- **REQ-030** Spacebar (NEXT) advances to the next cue — the same command for captions, for starting a loop, and for ending a loop. **STOP** (Shift+Q) stops a running loop and leaves a blank screen. Other keys: back, blank, clear, jump (doc 05).
- **REQ-031** Operator view shows already-sent cues, the current cue, and upcoming cues. Image/fullscreen cues appear as an indicator (icon + name), never the image itself.
- **REQ-032** Operator can jump to any cue (click/search) and undo (Back) the last operator action; going forward again is just NEXT.
- **REQ-033** Operator view shows live audience count and connection health.
- **REQ-034** Only one operator controls at a time; others are view-only, with explicit takeover.
- **REQ-035** Operator can work from a laptop keyboard, and also from a tablet/phone (large Next button; optional Bluetooth clicker maps to keys).
- **REQ-036** Setup (`/admin`) and Operator (`/operator`) are separate pages with separate passwords; the operator page exposes no setup actions.
- **REQ-037** Built-in help for operators and admins, available from every `/admin` and `/operator` screen, works offline, and includes the show file/Markdown conventions with copy-able examples (doc 05, section 4).

### Audience
- **REQ-040** One URL, no interaction needed. Page auto-scrolls as captions arrive, newest at bottom, older text dimmed.
- **REQ-046** A viewer can scroll up to re-read earlier lines without losing the live captions (a "back to live" control with a count of new lines); shared screens are not scrollable.
- **REQ-041** Fullscreen mode (text or image) takes over the entire display, e.g. "Intermission", and later returns to scrolling captions.
- **REQ-042** Works on current Safari/iOS, Chrome/Android, desktop Chrome/Firefox/Edge/Safari. The **browser floor** is: audience page, about Safari/iOS 12, Chrome 66, Firefox 60 (ES2018 syntax, `EventSource`, CSS variables; newer CSS has fallbacks); Setup, operator page and Help, about Safari/iOS 14.1, Chrome/Edge 87, Firefox 66 (they lay out with flex gaps and `inset`). `<dialog>` and `replaceChildren` are filled in for browsers without them (`web/shared/compat.js`). A browser below the floor is told so in plain words instead of showing a blank page. These are derived from the features used and checked by `test/e2e/check-compat.js`, not from testing old devices.
- **REQ-043** Keeps the screen awake where the browser allows (Wake Lock). Where it is unavailable (every browser over plain HTTP except on `localhost`, and older iPhones), a tiny silent, invisible looping video keeps the screen on instead; if the browser will not start it, a one-time tip asks the viewer to turn off Auto-Lock.
- **REQ-044** Projector display mode: URL parameter for larger type, no chrome, cursor hidden.
- **REQ-045** Accessibility-first defaults: pure black background / white text, high contrast, large type, responsive to orientation and screen size. Venue-level theme (colors, font size) and optional per-speaker colors set in setup; optional per-device local font-size control that is never required. Two readability fonts are bundled and available as a show-wide choice and as a per-device option: a dyslexia-friendly one (OpenDyslexic) and a high-legibility one (Atkinson Hyperlegible).

### Configuration and access
- **REQ-050** Setup is a guided wizard: create show → import show file → upload images → look → test → go live.
- **REQ-051** `/admin` and `/operator` each have their own password (plaintext in `config.yaml`, generated on first run), entered via a login form with a session cookie — not HTTP Basic auth. An admin session also works on `/operator`. The audience interface is open.
- **REQ-052** Show file edits during a live (on air) show require confirmation (typing the show name); the cursor is preserved by cue index, with a warning if the cue count changes.
- **REQ-055** Setup and operator pages can be limited to listed computers (`access` in `config.yaml`); the audience page is always open. Forged cross-site requests to the API are refused.
- **REQ-056** Passwords can be changed from the Setup page, without a restart.
- **REQ-059** Every save of a show file keeps the version it replaced (the last 50 per language), and an earlier version can be compared with the editor and loaded back into it. A version can be given a name, and a named version is never forgotten.
- **REQ-057** Two people editing a show file do not silently overwrite each other.
- **REQ-058** Logins, the operator's control, the show position and the last few undo steps survive a restart.
- **REQ-053** The setup UI displays the audience URL and a QR code, plus the list of host IP addresses.
- **REQ-054** Multiple shows can be stored; one is "active". A show can be **archived** (kept in full, listed apart, unable to go on air, restorable) or deleted; the show on air can be neither.

## Non-functional requirements
- **REQ-060** No audience data collection; no cookies required; no analytics.
- **REQ-061** Server logs to the console (and optionally a file via `--log-file`); the last 1000 lines are kept in memory. The recent log can be read in Setup (Go Live → Recent log…) with a level filter. A downloadable support bundle (settings without passwords, recent log, system facts, show summaries) is available from the Go Live step. Log lines carry a level (`DEBUG`, `INFO`, `WARN`, `ERROR`) and lines below the chosen level (`log.level` in `config.yaml`, or `--log-level`; default `info`) are dropped. The level can be changed while the program runs, from Setup → Settings (it is also written to `config.yaml`). A log file kept with `--log-file` starts a new file at 5 MB and keeps three old ones (`.1` to `.3`), so it cannot fill the disk. `warn` and `error` are for problems; `info` for events such as sign-ins, takeovers and activations (never passwords); `debug` adds every request and operator command.
- **REQ-065** A loop can have a **countdown timer** (`[loop: timer 15 minutes]` or `[loop: timer 7:30pm]`) that the loop's text shows with `{timer}`, counting down live on every audience screen and for the operator. A number of minutes starts when the loop is entered; a time of day is today's. Devices with wrong clocks still show the right time (doc 03 *Loop timers*).
- **REQ-066** A show that is loaded waits in **standby**: the audience sees the venue's standby screen (an optional logo and message set once for the venue in Setup, or the built-in words in the viewer's language) and nothing of the show until the operator presses **START**; **Go off air** returns to it without losing the place, and Restart goes back to the beginning and to standby. **Live** means on air, so a show in standby can be edited, deleted or replaced freely. A pre-show loop starts at START, not at activation. (Doc 02 *On air, standby, off air*.)
- **REQ-067** On the computer that runs the program, Setup and the operator page open without a password: a browser at `localhost`, `127.0.0.1` or `[::1]` is signed in as admin (`access.local_auto_login`, on by default). It requires a loopback connection, a loopback `Host`, no proxy headers and no cross-site `Sec-Fetch-Site`, so a web page rebinding its own name to 127.0.0.1, or a reverse proxy, never gets it; the access lists still apply first. An admin session (password or automatic) gets a **Setup** link at the top of the operator page.
- **REQ-062** Cold start < 2 s; binary < 30 MB.
- **REQ-064** Setup, the operator page and Help are available in the six languages the audience page has (English, Spanish, French, German, Italian, Portuguese). The language is chosen from a menu on each page (remembered in the browser; the first visit follows the browser's language). A sentence with no translation shows in English rather than breaking. The built-in documentation (every help topic and the two guides) is translated into the same languages; a topic always has a translation, and the tests check that each translation has the same headings, tables, links and (byte for byte) the same show file examples as the English.
- **REQ-063** Accessibility: semantic markup, `aria-live="polite"` region on audience page, respects `prefers-reduced-motion`, color-contrast ≥ WCAG AA.
