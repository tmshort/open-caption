# 08 — Roadmap and Task Breakdown

Each milestone is shippable and ends with a demoable result. Tasks reference requirements.

## M0 — Skeleton
- [x] Add Apache-2.0 headers/`NOTICE` as needed (REQ-007)
- [x] `go mod init`, repo layout per doc 02, Makefile/`just` with build/test/lint, cross-compile targets (REQ-001)
- [x] GitHub Actions: test + build matrix (macOS, Windows, Linux)
- [x] `GET /healthz`, config/data-dir resolution (REQ-004); [x] embedded static assets (REQ-002)

## M1 — Script engine (no UI)
- [x] `script.Parse` with cues, directives, comments, speakers, multi-line, diagnostics (REQ-020–023, 025)
- [x] Markdown-subset renderer (bold/italic/images; escapes all HTML) (REQ-022)
- [x] `config.yaml` load/generate with random passwords; flag/env overrides (REQ-005)
- [x] `live` engine: apply/undo/jump, loops with server-side timer and STOP, snapshots, export/restore (file persistence still to wire up) (REQ-012, 025, 030)
- [x] Table-driven, property (jump == sequential) and fuzz tests (golden files still to add)

## M2 — Audience stream + display
- [x] Hub + SSE `/events` with snapshot/heartbeat/reconnect semantics (REQ-010, 011, 013)
- [x] Audience page: scroll, fullscreen text/image, blank, server-driven loops, image preload, reconnect overlay (REQ-040–042)
- [x] Temporary dev endpoint (`--dev`, `--script`, `--assets`) to advance via curl → first end-to-end demo
- [x] Load-test tool (`cmd/loadtest`); 200 clients p95 5 ms, 500 clients p95 8 ms, 32 MB RSS (local, 2026-10-14)

## M3 — Operator
- [x] Auth: separate admin/operator passwords, login form + session cookie, rate limit (REQ-051)
- [x] HTTP + optional HTTPS listeners (provided cert or self-signed) (REQ-006)
- [x] Operator UI at `/operator`: list, keyboard control (NEXT/Space, STOP/Shift+Q), indicators, jump, ad-hoc message, controller claim/takeover (REQ-030–036)
- [x] Audience count and health (REQ-033)
- [x] Live state persisted to `live.json` and restored on restart (REQ-012)

## M4 — Setup
- [x] Shows CRUD, script editor + diagnostics, asset upload/management (REQ-020, 024, 054)
- [x] Theme editor with live previews and contrast check (REQ-045)
- [x] Wizard flow, network panel, QR page (REQ-050, 053)
- [x] Rehearsal mode: isolated engine, `/api/admin/rehearsal/*`, preview pane (wizard Test step)
- [x] Live-mode guard rails (REQ-052)
- [x] Export/import show zip

## M5 — Polish and hardening
- [x] Wake-lock handling with a one-time tip when unavailable; optional per-device text size and light/dark controls; projector mode (REQ-043, 044, 045). *(The hidden-video fallback for iOS is not implemented.)*
- [x] Help viewer and topics for admins and operators, incl. script conventions and sample script; help-coverage tests (REQ-037)
- [~] Accessibility pass (REQ-063): readable dimming floor, focus outlines, labelled controls, reduced motion; a screen-reader and contrast audit with real devices is still to do
- [x] `/captive-api`, venue network guide (`docs/venue-network-guide.md`)
- [x] Support bundle, logging (REQ-061)
- [ ] Device matrix testing and acceptance checklist (doc 07)

## Housekeeping after M5
- [x] Help ships the guides (`docs/usage-guide.md`, `docs/venue-network-guide.md`) inside the program, rendered with nested lists and in-page links
- [x] Rename and duplicate shows
- [x] Per-character speaker colors; right-to-left caption direction
- [x] Logins survive a restart; the operator page takes control back automatically
- [x] Operational niceties: undo survives a restart (last 10 steps), script edit-conflict detection, change passwords from Setup, Look tick kept on the server, expired-session/rehearsal cleanup, `--dev` listens on this computer only, operator section menu and "last sent" timer

## Stretch / Later
- mDNS advertisement; signed/notarized macOS build; Windows installer
- Multi-language scripts (identical structure, `[language: xx]`, `?lang=`) (REQ-026); engine/protocol already carry `lang`
- ACME/Let's Encrypt certificates (needs Internet + public name)
- Markdown headings as section markers
- Hashed passwords in config
- Cue timing hints / countdowns for operator; stage-manager "standby" cues
- Remote operator via second device with restricted view
- Optional auto-follow (speech recognition) as an assistant to operator
- Print-friendly script/cue sheet export
