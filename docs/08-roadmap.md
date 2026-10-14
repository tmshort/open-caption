# 08 — Roadmap and Task Breakdown

Each milestone is shippable and ends with a demoable result. Tasks reference requirements.

## M0 — Skeleton
- [ ] Add Apache-2.0 headers/`NOTICE` as needed (REQ-007)
- [ ] `go mod init`, repo layout per doc 02, Makefile/`just` with build/test/lint, cross-compile targets (REQ-001)
- [ ] GitHub Actions: test + build matrix (macOS, Windows, Linux)
- [ ] Embedded static assets, `GET /healthz`, config/data-dir resolution (REQ-002, 004)

## M1 — Script engine (no UI)
- [ ] `script.Parse` with cues, directives, comments, speakers, multi-line, diagnostics (REQ-020–023, 025)
- [ ] Markdown-subset renderer (bold/italic/images; escapes all HTML) (REQ-022)
- [ ] `config.yaml` load/generate with random passwords; flag/env overrides (REQ-005)
- [ ] `live` engine: apply/undo/jump, loops with server-side timer and STOP, snapshots, persistence of cursor (REQ-012, 025, 030)
- [ ] Golden-file + fuzz tests

## M2 — Audience stream + display
- [ ] Hub + SSE `/events` with snapshot/heartbeat/reconnect semantics (REQ-010, 011, 013)
- [ ] Audience page: scroll, fullscreen text/image, blank, server-driven loops, image preload, reconnect overlay (REQ-040–042)
- [ ] Temporary dev endpoint to advance via curl → first end-to-end demo
- [ ] Load-test tool; verify 200-client target (REQ-010)

## M3 — Operator
- [ ] Auth: separate admin/operator passwords, login form + session cookie, rate limit (REQ-051)
- [ ] HTTP + optional HTTPS listeners (provided cert or self-signed) (REQ-006)
- [ ] Operator UI at `/operator`: list, keyboard control (NEXT/Space, STOP/Shift+Q), indicators, jump, ad-hoc message, controller claim/takeover (REQ-030–036)
- [ ] Audience count and health (REQ-033)

## M4 — Setup
- [ ] Shows CRUD, script editor + diagnostics, asset upload/management (REQ-020, 024, 054)
- [ ] Theme editor with live previews (REQ-045)
- [ ] Wizard flow, network panel, QR page (REQ-050, 053)
- [ ] Live-mode guard rails (REQ-052)
- [ ] Export/import show zip

## M5 — Polish and hardening
- [ ] Wake-lock handling, iOS quirks, projector mode (REQ-043, 044)
- [ ] Accessibility pass (REQ-063)
- [ ] Rehearsal mode
- [ ] `/captive-api`, venue network guide (doc 06)
- [ ] Support bundle, logging (REQ-061)
- [ ] Device matrix testing and acceptance checklist (doc 07)

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
