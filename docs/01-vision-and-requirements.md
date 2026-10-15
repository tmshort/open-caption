# 01 — Vision and Requirements

## Vision

A community theater can set up open captioning for a show in under an hour,
with volunteers who are not technical, and run it from one laptop. Audience
members join the guest WiFi, open one URL, and read the show's captions in sync
with the stage. Nothing to install, nothing to click.

## Users

| User | Description | Key needs |
|------|-------------|-----------|
| **Admin / Setup user** | Several volunteers over time who prepare a show | Easy, forgiving, guided; clear errors when a script is malformed |
| **Operator** | Follows the show (script/prompt book) and releases captions | One-key advance, confidence about position, can't get lost, can recover from mistakes fast |
| **Audience member** | BYOD phone/tablet/desktop, or a shared projector | Zero interaction, readable, resilient to WiFi drops and sleeping screens |

## Scope

In scope: one server process on one host; script import; live operation;
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
- **REQ-005** Settings (ports/addresses, TLS, passwords, defaults) live in a human-editable `config.yaml` (generated on first run); flags/env vars override it.
- **REQ-006** The server listens on plain HTTP and, optionally, HTTPS (user-supplied cert or generated self-signed) at the same time. Automatic certificates (ACME/Let's Encrypt) are a possible later addition; they need Internet access and a public name and are out of scope for v1.
- **REQ-007** Licensed under Apache-2.0; dependencies must be license-compatible.

### Capacity and resilience
- **REQ-010** Support at least 200 concurrent audience clients (venue has ~150 seats; requirement is >75). Caption latency from operator keypress to render ≤ 250 ms at p95 on a LAN under that load.
- **REQ-011** Audience clients reconnect automatically after network loss or device sleep and show the *current* state without user action.
- **REQ-012** Server restart resumes at the previously active show, cue position and on-screen state; a loop that was running restarts from its first item.
- **REQ-013** A client joining mid-show immediately sees current state plus recent scrollback.

### Script
- **REQ-020** Import a Markdown/text script (`.md` or `.txt`, both parsed as Markdown, UTF-8) via upload or paste; re-import without losing assets.
- **REQ-021** Each non-blank script line is a caption (long lines wrap; the author controls splitting); non-technical authors can add human-readable directives (clear screen, images, fullscreen text, loops) (doc 03).
- **REQ-022** Formatting is a Markdown subset (bold/italic, images). Raw HTML is **not** supported; it is displayed as literal text with a diagnostic. Exception: `<!-- -->` comments. `[note: …]` is an operator-only note; `#` is reserved for future headings.
- **REQ-023** Import validates the script and reports errors/warnings with line numbers *without* blocking the preview.
- **REQ-024** Images (PNG/JPG/GIF/WebP; no SVG) are uploaded/managed in the setup UI and referenced by filename. Missing references are flagged. The server does not rescale; browsers scale to fit. Audience pages prefetch upcoming images.
- **REQ-025** **Loops**: a block of mixed fullscreen items (images/text), each shown for its own duration (default from config), repeating. Used for pre-show and intermission. A loop that is the first cue auto-starts on activation; any other loop starts when the operator presses NEXT on it.
- **REQ-027** Musicals: lyrics can be marked with a `~` shorthand or a `[song]` block, are shown in italics between ♪ notes, and music notes can be inserted from the editor (doc 03, *Songs and lyrics*). Two or more voices sung or spoken at the same time (duets, counterpoint) are shown together from one operator keypress, using `[together]` or ` || ` (doc 03, *Voices at the same time*).
- **REQ-026** Multiple languages: one script per language, identical in structure (checked cue by cue), each identified by a `[language: code | Name]` line. One operator keypress sends the same cue to every language. The audience's language is detected from `Accept-Language` (default: the main language), with a picker and a `?lang=` URL form. A translation that does not match the main script is reported and not shown to the audience.

### Operator
- **REQ-030** Spacebar (NEXT) advances to the next cue — the same command for captions, for starting a loop, and for ending a loop. **STOP** (Shift+Q) stops a running loop and leaves a blank screen. Other keys: back, blank, clear, jump (doc 05).
- **REQ-031** Operator view shows already-sent cues, the current cue, and upcoming cues. Image/fullscreen cues appear as an indicator (icon + name), never the image itself.
- **REQ-032** Operator can jump to any cue (click/search) and undo (Back) the last operator action; going forward again is just NEXT.
- **REQ-033** Operator view shows live audience count and connection health.
- **REQ-034** Only one operator controls at a time; others are view-only, with explicit takeover.
- **REQ-035** Operator can work from a laptop keyboard, and also from a tablet/phone (large Next button; optional Bluetooth clicker maps to keys).
- **REQ-036** Setup (`/admin`) and Operator (`/operator`) are separate pages with separate passwords; the operator page exposes no setup actions.
- **REQ-037** Built-in help for operators and admins, available from every `/admin` and `/operator` screen, works offline, and includes the script/Markdown conventions with copy-able examples (doc 05, section 4).

### Audience
- **REQ-040** One URL, no interaction needed. Page auto-scrolls as captions arrive, newest at bottom, older text dimmed.
- **REQ-046** A viewer can scroll up to re-read earlier lines without losing the live captions (a "back to live" control with a count of new lines); shared screens are not scrollable.
- **REQ-041** Fullscreen mode (text or image) takes over the entire display, e.g. "Intermission", and later returns to scrolling captions.
- **REQ-042** Works on current Safari/iOS, Chrome/Android, desktop Chrome/Firefox/Edge/Safari.
- **REQ-043** Keeps the screen awake where the browser allows (Wake Lock; fallback video trick as a documented option).
- **REQ-044** Projector display mode: URL parameter for larger type, no chrome, cursor hidden.
- **REQ-045** Accessibility-first defaults: pure black background / white text, high contrast, large type, responsive to orientation and screen size. Venue-level theme (colors, font size) and optional per-speaker colors set in setup; optional per-device local font-size control that is never required. Two readability fonts are bundled and available as a show-wide choice and as a per-device option: a dyslexia-friendly one (OpenDyslexic) and a high-legibility one (Atkinson Hyperlegible).

### Configuration and access
- **REQ-050** Setup is a guided wizard: create show → import script → upload images → look → test → go live.
- **REQ-051** `/admin` and `/operator` each have their own password (plaintext in `config.yaml`, generated on first run), entered via a login form with a session cookie — not HTTP Basic auth. An admin session also works on `/operator`. The audience interface is open.
- **REQ-052** Script edits during a live show require confirmation (typing the show name); the cursor is preserved by cue index, with a warning if the cue count changes.
- **REQ-053** The setup UI displays the audience URL and a QR code, plus the list of host IP addresses.
- **REQ-054** Multiple shows can be stored; one is "active".

## Non-functional requirements
- **REQ-060** No audience data collection; no cookies required; no analytics.
- **REQ-061** Server logs to the console (and optionally a file via `--log-file`); the last 1000 lines are kept in memory. A downloadable support bundle (settings without passwords, recent log, system facts, show summaries) is available from the Go Live step. *(Log levels are not implemented; messages are plain lines.)*
- **REQ-062** Cold start < 2 s; binary < 30 MB.
- **REQ-063** Accessibility: semantic markup, `aria-live="polite"` region on audience page, respects `prefers-reduced-motion`, color-contrast ≥ WCAG AA.
