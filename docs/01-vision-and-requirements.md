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
audience display; image/fullscreen/pre-show support.

Out of scope (v1): cloud hosting, user accounts, automatic speech
recognition / auto-follow of the show, audio, multi-language selection,
multiple simultaneous shows, internet-facing deployment.

## Functional requirements

### Platform and deployment
- **REQ-001** Single Go binary, no runtime dependencies, no CGO; builds for macOS (arm64/amd64), Windows (amd64), Linux (amd64/arm64).
- **REQ-002** All web assets embedded in the binary (`go:embed`). Nothing fetched from the internet at runtime (the venue network may have no uplink).
- **REQ-003** The whole system (setup, operator, audience) runs on a single host.
- **REQ-004** Persistent data lives in one data directory (configurable); no external database.

### Capacity and resilience
- **REQ-010** Support at least 200 concurrent audience clients (venue has ~150 seats; requirement is >75). Caption latency from operator keypress to render ≤ 250 ms at p95 on a LAN under that load.
- **REQ-011** Audience clients reconnect automatically after network loss or device sleep and show the *current* state without user action.
- **REQ-012** Server restart resumes at the previously active show and cue position.
- **REQ-013** A client joining mid-show immediately sees current state plus recent scrollback.

### Script
- **REQ-020** Import a plain-text script file (UTF-8) via upload or paste; re-import without losing assets.
- **REQ-021** Each script line is a caption by default; non-technical authors can add directives (clear screen, images, fullscreen text, pre-show) using a human-readable syntax (see doc 03).
- **REQ-022** Inline HTML is supported in captions, sanitized against an allow-list.
- **REQ-023** Import validates the script and reports errors/warnings with line numbers *without* blocking the preview.
- **REQ-024** Images are uploaded/managed in the setup UI and referenced by filename from the script. Missing references are flagged.
- **REQ-025** Pre-show content (images/text, optionally looping slideshow) is expressible in the script.

### Operator
- **REQ-030** Spacebar advances to the next cue. Other keys: back, blank, clear, jump (doc 05).
- **REQ-031** Operator view shows already-sent cues, the current cue, and upcoming cues. Image/fullscreen cues appear as an indicator (icon + name), never the image itself.
- **REQ-032** Operator can jump to any cue (click/search) and undo/redo advances.
- **REQ-033** Operator view shows live audience count and connection health.
- **REQ-034** Only one operator controls at a time; others are view-only, with explicit takeover.
- **REQ-035** Operator can work from a laptop keyboard, and also from a tablet/phone (large Next button; optional Bluetooth clicker maps to keys).
- **REQ-036** Setup and Operator may be one app with two modes/tabs, but the live surface must not expose destructive setup actions (see REQ-052).

### Audience
- **REQ-040** One URL, no interaction needed. Page auto-scrolls as captions arrive, newest at bottom, older text dimmed.
- **REQ-041** Fullscreen mode (text or image) takes over the entire display, e.g. "Intermission", and later returns to scrolling captions.
- **REQ-042** Works on current Safari/iOS, Chrome/Android, desktop Chrome/Firefox/Edge/Safari.
- **REQ-043** Keeps the screen awake where the browser allows (Wake Lock; fallback video trick as a documented option).
- **REQ-044** Projector display mode: URL parameter for larger type, no chrome, cursor hidden.
- **REQ-045** Readable defaults: high contrast, large type, responsive to orientation and screen size. Venue-level theme (colors, font size) set in setup; optional per-device local font-size control that is never required.

### Configuration and access
- **REQ-050** First-run experience is a guided wizard: create show → import script → upload images → preview → go live.
- **REQ-051** Setup and operator interfaces are protected by a PIN/password (set at first run); audience interface is open.
- **REQ-052** Setup actions that change the script are disabled (or require confirmation) while a show is live.
- **REQ-053** The setup UI displays the audience URL and a QR code, plus the list of host IP addresses.
- **REQ-054** Multiple shows can be stored; one is "active".

## Non-functional requirements
- **REQ-060** No audience data collection; no cookies required; no analytics.
- **REQ-061** Server logs to stdout/file with levels; a downloadable support bundle from the UI.
- **REQ-062** Cold start < 2 s; binary < 30 MB.
- **REQ-063** Accessibility: semantic markup, `aria-live="polite"` region on audience page, respects `prefers-reduced-motion`, color-contrast ≥ WCAG AA.
