# 08 — Roadmap and Task Breakdown

Each milestone is shippable and ends with a demoable result. Tasks reference requirements.

## M0 — Skeleton
- [x] Add Apache-2.0 headers/`NOTICE` as needed (REQ-007)
- [x] `go mod init`, repo layout per doc 02, Makefile/`just` with build/test/lint, cross-compile targets (REQ-001)
- [x] GitHub Actions: test + build matrix (macOS, Windows, Linux)
- [x] `GET /healthz`, config/data-dir resolution (REQ-004); [x] embedded static assets (REQ-002)

## M1 — Show file engine (no UI)
- [x] `showfile.Parse` with cues, directives, comments, speakers, multi-line, diagnostics (REQ-020–023, 025)
- [x] Markdown-subset renderer (bold/italic/images; escapes all HTML) (REQ-022)
- [x] `config.yaml` load/generate with random passwords; flag/env overrides (REQ-005)
- [x] `live` engine: apply/undo/jump, loops with server-side timer and STOP, snapshots, export/restore (file persistence still to wire up) (REQ-012, 025, 030)
- [x] Table-driven, property (jump == sequential) and fuzz tests (golden files still to add)

## M2 — Audience stream + display
- [x] Hub + SSE `/events` with snapshot/heartbeat/reconnect semantics (REQ-010, 011, 013)
- [x] Audience page: scroll, fullscreen text/image, blank, server-driven loops, image preload, reconnect overlay (REQ-040–042)
- [x] Temporary dev endpoint (`--dev`, `--show-file`, `--assets`) to advance via curl → first end-to-end demo
- [x] Load-test tool (`cmd/loadtest`); 200 clients p95 5 ms, 500 clients p95 8 ms, 32 MB RSS (local, 2026-10-14)

## M3 — Operator
- [x] Auth: separate admin/operator passwords, login form + session cookie, rate limit (REQ-051)
- [x] HTTP + optional HTTPS listeners (provided cert or self-signed) (REQ-006)
- [x] Operator UI at `/operator`: list, keyboard control (NEXT/Space, STOP/Shift+Q), indicators, jump, ad-hoc message, controller claim/takeover (REQ-030–036)
- [x] Audience count and health (REQ-033)
- [x] Live state persisted to `live.json` and restored on restart (REQ-012)

## M4 — Setup
- [x] Shows CRUD, show file editor + diagnostics, asset upload/management (REQ-020, 024, 054)
- [x] Theme editor with live previews and contrast check (REQ-045)
- [x] Wizard flow, network panel, QR page (REQ-050, 053)
- [x] Rehearsal mode: isolated engine, `/api/admin/rehearsal/*`, preview pane (wizard Test step)
- [x] Live-mode guard rails (REQ-052)
- [x] Export/import show zip

## M5 — Polish and hardening
- [x] Wake-lock handling with a one-time tip when unavailable; optional per-device text size and light/dark controls; projector mode (REQ-043, 044, 045). The hidden-video fallback keeps older iPhones awake where the Wake Lock API is missing (untested on real iPhones).
- [x] Help viewer and topics for admins and operators, incl. show file conventions and sample show file; help-coverage tests (REQ-037)
- [~] Accessibility pass (REQ-063): readable dimming floor, focus outlines, labelled controls, reduced motion; a screen-reader and contrast audit with real devices is still to do
- [x] `/captive-api`, venue network guide (`docs/venue-network-guide.md`)
- [x] Support bundle, logging (REQ-061)
- [ ] Device matrix testing and acceptance checklist (doc 07)

## After M5: hardening and polish
- [x] Access lists for Setup/operator, forged-request check, scroll-back for the audience, rewritten end-to-end test (`make e2e`, also a CI job), specs reconciled (2026-10-15)
- [ ] Real-device rehearsal (the largest remaining risk); minimum browser versions to be set from it
- [ ] Signed/notarized builds, installers, release pipeline
- [ ] Screen-reader and contrast audit

## Housekeeping after M5
- [x] Help ships the guides (the usage guide, the venue network guide, the HTTPS certificate guide and the `config.yaml` reference) inside the program, translated into five more languages, under a **Guides and reference** heading
- [x] Rename and duplicate shows
- [x] Per-character speaker colors; right-to-left caption direction
- [x] Logins survive a restart; the operator page takes control back automatically
- [x] Operational niceties: undo survives a restart (last 10 steps), show file edit-conflict detection, change passwords from Setup, Look tick kept on the server, expired-session/rehearsal cleanup, `--dev` listens on this computer only, operator section menu and "last sent" timer

- [x] Multiple languages (REQ-026): per-language engines and hubs driven in lockstep, structure checking, language negotiation and menu, Setup language tabs, sample Spanish translation

- [x] mDNS: the server announces itself as `open-caption.local` (or a name chosen in Setup)

- [x] Automatic HTTPS certificates (REQ-068): ACME DNS-01 with build-time choice of DNS providers, credentials from the environment/file/config, Setup status and renew-now, a Pebble end-to-end test (doc 10)
- [x] A supplied HTTPS certificate is watched and a renewed one used without a restart
- [x] Standby and on air (REQ-066): a loaded show waits in standby until the operator starts it; Go to standby…, Restart show…
- [x] Loop countdown timers (REQ-065); announcements for the operator's one-off notice
- [x] Local automatic sign-in on the computer that runs the program (REQ-067)
- [x] One interface for Setup and the operator page (`web/shared/ui.css`, `header.css`), a connection-lost notice, the show list's More menu, the status strip
- [x] Command line: `--check-show` (a lint for show files, sharing its analysis with Setup's editor), `--help`, the man page (`docs/open-caption.1`), the configuration reference (`docs/configuration-reference.md`)

## Stretch / Later
- Signed/notarized macOS build; Windows installer
- Markdown headings as section markers
- Hashed passwords in config
- Cue timing hints for the operator (loops already have a countdown timer)
- Remote operator via second device with restricted view
- Optional auto-follow (speech recognition) as an assistant to operator
- Print-friendly show file/cue sheet export
