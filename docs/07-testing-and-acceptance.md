# 07 — Testing and Acceptance

## Test layers

| Layer | What | How |
|-------|------|-----|
| Unit | `show file` parser and validator (table-driven; the shipped sample must parse clean and its Spanish translation must match its structure), Markdown renderer/escaping, language structure check, config load/generate and access lists, `live` engine (apply/undo/jump determinism, persistence), show store, help renderer | `go test ./...` |
| Property/fuzz | `showfile.Parse` never panics and never emits unescaped HTML (`FuzzParse`); engine: `jump(n)` state equals sequential apply state, for random show files | `go test`, `make fuzz` |
| Consistency | Help covers every show file directive, escape and warning; every example in help parses clean; every operator key documented exists in `operator.js`; the guides in `docs/` render with working links and anchors; every bundled font is served | part of `go test ./...` |
| Automatic certificates | The manager (issue, renew window, failed renewal keeps the old certificate, wrong-name refusal), credentials precedence and file parsing, configuration rules, Setup endpoints, redaction of credentials in the support bundle; the real ACME path (lego, DNS-01) against Pebble | `go test` (fake issuer); `make pebble-test` (a pinned Pebble, doc 10); `make acme-variants` builds every provider choice |
| Multicast DNS | The responder on a fake network (claiming and renaming the name, address and IPv6 answers, goodbyes, a changed address, a question from the computer's own address); and on real multicast sockets on a spare port, asked from this same computer (skipped when the computer has no multicast interface or, on CI, when multicast does not loop back) | `go test ./internal/mdns` |
| Release packaging | The download holds the guides and no design documents, and has no dead links; the release workflow uses `make release` | `go test ./deploy`; `make release-snapshot` builds every platform locally |
| Static analysis | Unused code, misuse of the standard library, needless constructs (the code is kept free of findings) | `make lint` (`go vet` and a pinned staticcheck, run through `go run`) in CI on Linux |
| Static analysis of the JavaScript | Undefined or unused variables, mistakes in the page scripts, browser tests and tools (kept free of findings; `eslint.config.mjs`) | `make lint-js` (ESLint, installed with the browser tests' packages) in CI on Linux |
| Licence headers | Every source file (Go, JS, CSS, HTML, scripts, workflows, Makefile) starts with the copyright and SPDX lines | `go test ./deploy` |
| Supply chain | Known vulnerabilities in dependencies and the standard library | `make vulncheck` (govulncheck on the code and on the pinned Pebble programs) in CI weekly and on release tags; Dependabot proposes Go, npm and GitHub Actions updates weekly |
| Shutdown | `cmd/open-caption` starts the real binary, puts a show on air, sends it Ctrl+C (Ctrl+Break on Windows) and checks it saves its position and exits 0, on Linux, macOS and Windows |
| Integration | httptest servers: login/roles/rate limit/cookie flags, access lists, forged-request refusal, operator flow and takeover, audience SSE (snapshot-then-events, late joiner, loops, 503), languages in lockstep, rehearsal isolation, persistence across restart, settings/passwords, edit conflicts, export/import | `go test ./...` |
| Load | N SSE clients + the operator advancing; latency percentiles, dropped clients | `cmd/loadtest` (`./bin/open-caption --dev …`); CI runs it with 200 clients on every push |
| Browser end-to-end | Real Chrome against the built program: Setup (sample show, edit conflict, Look), rehearsal isolation, live show with two languages, Back/Blank/Clear, songs and voices together, speaker colors, fonts, help pauses keys, scroll-back, language menu, Settings/password change, restart survival | `make e2e` (needs Node + Chrome; runs in CI as a separate job) |
| Manual device | iPhone Safari, iPad Safari, Android Chrome, desktop Chrome/Firefox/Edge, a projector, a full house on real WiFi | Acceptance checklist below |

## Key test cases

- Loops: per-item duration, loop default, config default; loop first in file auto-starts; STOP→blank→NEXT; Back after ending a loop; reconnect mid-loop shows the current item; images in the next loop are preloaded.
- Skip (REQ-030): the next line is passed over with no audience event and "last sent" unchanged, operator-only cues are stepped over, Back undoes it, the end of the show and standby are refused; the API (409 in standby); a browser check (S key, the toast, Back).
- Auth: separate passwords; admin session works on `/operator`, operator session rejected on `/admin`; no Basic-auth challenge; cookie `Secure` only over HTTPS; HTTP and HTTPS serve the same routes with no redirect.
- Show file: no markup → N captions; comments/blank lines ignored; CRLF; BOM; very long line; unicode/emoji/RTL passthrough; unclosed `[loop]`, raw `<b>` shown literally with an info diagnostic, `<!-- -->` comments (single- and multi-line, unclosed) never reach any client, `[note: …]` reaches only the operator view, `#` lines ignored with an info diagnostic; unknown directive; missing image with fuzzy suggestion; `\#` escape; `[[ ]]` multi-line; speaker detection positives/negatives.
- Help: every parser directive/escape/diagnostic is documented in `web/help/` (test); every fenced example parses without errors; sample show file parses clean; opening help on `/operator` pauses show-control keys; help works with no network.
- Escaping: `<script>`, `onerror=`, `javascript:` links, `<img src=http…>` in show file text are rendered as literal text; Markdown links are not clickable; image refs only resolve to uploaded assets.
- Engine: advance past end (no-op + signal "end of show file"); back at start; jump forward/backward equals sequential; `clear` mid-history then undo restores scrollback; fullscreen followed by caption dismisses fullscreen; restart restores cursor and scrollback.
- Concurrency: two operator sessions advancing with the same `rev` → exactly one NEXT applied, other gets 409; a NEXT sent while a loop ticks (seq changed, rev unchanged) succeeds.
- Rehearsal: commands on `/api/admin/rehearsal/*` never reach audience clients or change `live.json`, even while a real show is live.
- Security: first-run passwords never appear in any HTTP response; beyond `audience.max_clients`, `/events` returns 503.
- SSE: slow client doesn't delay others; overflow closes the slow client; heartbeat arrives; late joiner receives scrollback.
- Images: replacing an image on screen changes its URL and snapshot; replacing an upcoming image updates its preload URL without rebuilding the screen; unrelated uploads do neither; deleting an image referenced by the show on air is refused, but standby permits it.

- Languages: a translation that does not match the main show file is reported with its own line number and is not offered; one keypress sends the same cue to every language and a language that drifts is moved back; negotiation by `?lang=`, `Accept-Language` quality values, regional fallback; devices on a language a new show lacks are told to reload; per-language state survives a restart.
- Songs and voices: `~` shorthand, `[song]`, `[together]` and ` || ` (including mixed spoken/sung, unclosed blocks, escapes).
- Security: access lists (listed, unlisted, bad entry denies all, cookie from an unlisted computer, admin address may run the show); forged `Origin`/`Sec-Fetch-Site` refused; sessions survive a restart but not a password change; the settings and support bundle never contain passwords.
- Persistence: undo steps and last-action time survive a restart; two people saving one show file is detected.
- **Audience report (REQ-070):** `internal/report` unit tests (devices counted once across reconnects and second tabs, early arrivals counted from the start, a different show ends the run, the graph's points and loop stretches (with their names), JSON seconds, CSV, same-second stored file names, readable CSV download names); `internal/httpapi` integration tests (devices waiting in standby, a tablet joining, a phone dropping and returning, rapid reset and same-show re-activation creating separate runs, the JSON, CSV and saved report, deleting a saved report but not the run in progress, a rehearsal not counted); browser tests (the report from START, the page, the banner link at the end of the show, no Delete for the run in progress).

## Performance targets (verified by load test)

- 200 concurrent audience connections: operator keypress → last client receives event p95 ≤ 250 ms, p99 ≤ 500 ms.
- Server RSS < 150 MB at 200 clients; CPU < 10% of one core while idle-connected.
- 500 clients degrade gracefully (no crash); connections beyond `audience.max_clients` (default 500, per language) receive 503 and retry with backoff.
- Measured on one computer (2026-10-14): 200 clients p95 6 ms, 500 clients p95 8 ms, about 32 MB of memory. Real WiFi is the limit, not the program.

## Acceptance checklist (pre-release, in venue conditions if possible)

1. [ ] Fresh install on macOS, Windows, Linux: first run generates `config.yaml` and reaches the `/admin` login with no errors.
2. [ ] A volunteer unfamiliar with the project imports a show file and images using only the UI hints within 15 minutes.
3. [ ] Malformed show file shows clear line-numbered diagnostics and still previews.
4. [ ] Operator advances through a full show with Space only; no missed or duplicated lines.
5. [ ] Image cue appears fullscreen on audience devices; operator sees only the indicator.
6. [ ] `[clear]`, `[blank]`, intermission fullscreen text behave as in doc 02 table.
7. [ ] A first-cue loop auto-starts on activation and repeats on the server timer; Shift+Q stops it to blank; Space ends it and sends the next caption; an intermission loop starts when Space reaches it.
8. [ ] 75+ real/simulated devices connected; all in sync; audience count correct.
9. [ ] Toggle airplane mode on a phone for 20 s mid-show: it recovers to the current state unaided.
10. [ ] Kill and restart the server mid-show: clients reconnect; operator resumes at the same cue.
11. [ ] Locking and unlocking an iPad mid-show returns to current state.
12. [ ] Projector mode readable from the back of the room.
13. [ ] A second operator tab is view-only until takeover.
14. [ ] Audience cannot reach any `/api/*` mutating endpoint or `/admin`.
15. [ ] A new volunteer finds how to add a comment, a note, an image and a loop using only the in-app help.
16. [ ] The same walk-through with a second language: both languages show the same line on one keypress; the language menu and `Accept-Language` pick the right one; a deliberately broken translation is reported and held back.
17. [ ] A song with lyrics, and two voices together, read correctly on a phone and a projector.
18. [ ] Scrolling up on a phone to re-read does not lose the live captions, and "back to live" works.
19. [ ] With `admin_allow: [localhost]`, Setup cannot be opened from a phone on the guest WiFi, and the audience page still works.
20. [ ] The oldest iPad/iPhone/Android device you expect in the audience shows captions, keeps its screen on (or shows the Auto-Lock tip), and reconnects after being locked.
21. [ ] Setup and the operator page work on the oldest device you intend to use for them (the floor is about Safari/iOS 14.1 or Chrome 87; an older browser shows a "too old" message). The audience page is expected to work from about iOS 12; try the oldest phone you can find.
22. [ ] On the venue's guest WiFi, `http://open-caption.local/` opens on an iPhone, an Android phone and a Windows laptop (or you know the network blocks multicast and the numeric address and QR code are what you will rely on).


## Accessibility audit (REQ-063)

`node test/e2e/a11y.js` (run in CI on Linux) drives real Chrome through every page and Setup step (audience waiting and with
captions, Setup sign-in and all seven steps, the Settings dialog, operator, Help), in **dark and light colour schemes** at a
desktop and a phone width, and runs **axe-core** with the WCAG 2.0/2.1/2.2 A and AA rules plus best practices. It also checks
**reflow** (no sideways scrolling at 320 px wide, WCAG 1.4.10) and that **Tab** walks the controls with a visible **focus
ring** (2.4.7). It also audits the audience report page (REQ-070), signed out and with a seeded report. It must report nothing.

The first run found 8 kinds of problem, all fixed: the audience "Aa" and language buttons were too faint (contrast 1.5:1;
now 6:1 or better); Setup, the operator page and Help had no `main` landmark, and the operator page no level-one heading;
the language tabs held a non-tab button inside the `tablist`; the file pickers and the two scrollable lists had no
accessible name or keyboard access; the key hints on buttons and the current Help topic had too little contrast; the
language menu was below the 24 px minimum target size; the number boxes in the Look step were unreadable in the dark
scheme; and the show list overflowed at 320 px.

**Not covered by the automated audit (do by hand before relying on it):**
- [ ] A screen reader (VoiceOver on iPhone and Mac, NVDA on Windows, TalkBack on Android) reads the audience page's caption
  log sensibly as lines arrive, and the operator page's list and status messages make sense. The audience page is a
  polite live region; whether a given screen reader speaks every line at show speed is untested.
- [ ] Windows High Contrast / forced-colors mode and iOS "Increase Contrast" (no special styles are provided; the pages use
  system colours only where the default styling does).
- [ ] Browser zoom to 200% and larger system text sizes on a real phone (the audience text scales with its own size control).
- [ ] The translated pages with a screen reader in each language.

## Running the browser test on a slow computer

`test/e2e/run.js` stretches every pause and time limit by `OC_E2E_SLOW` (default 2 on CI, 1 elsewhere). A section that makes no
progress for `OC_E2E_SECTION_LIMIT` seconds (default 120, also stretched) stops the run with `STUCK: the section "…"`, and the whole
run is limited by `OC_E2E_TIMEOUT` (default 900 s). Whatever happens (a failure, a stuck section, Ctrl+C) every Chrome and the
server it started are closed, so a bad run leaves nothing behind to disturb the next one.
