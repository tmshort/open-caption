# 07 — Testing and Acceptance

## Test layers

| Layer | What | How |
|-------|------|-----|
| Unit | `show file` parser and validator (table-driven; the shipped sample must parse clean and its Spanish translation must match its structure), Markdown renderer/escaping, language structure check, config load/generate and access lists, `live` engine (apply/undo/jump determinism, persistence), show store, help renderer | `go test ./...` |
| Property/fuzz | `showfile.Parse` never panics and never emits unescaped HTML (`FuzzParse`); engine: `jump(n)` state equals sequential apply state, for random show files | `go test`, `make fuzz` |
| Consistency | Help covers every show file directive, escape and warning; every example in help parses clean; every operator key documented exists in `operator.js`; the guides in `docs/` render with working links and anchors; every bundled font is served | part of `go test ./...` |
| Supply chain | Known vulnerabilities in dependencies and the standard library | `govulncheck ./...` in CI; Dependabot proposes Go, npm and GitHub Actions updates weekly |
| Shutdown | `cmd/open-caption` starts the real binary, puts a show on air, sends it Ctrl+C (Ctrl+Break on Windows) and checks it saves its position and exits 0, on Linux, macOS and Windows |
| Integration | httptest servers: login/roles/rate limit/cookie flags, access lists, forged-request refusal, operator flow and takeover, audience SSE (snapshot-then-events, late joiner, loops, 503), languages in lockstep, rehearsal isolation, persistence across restart, settings/passwords, edit conflicts, export/import | `go test ./...` |
| Load | N SSE clients + the operator advancing; latency percentiles, dropped clients | `cmd/loadtest` (`./bin/open-caption --dev …`); CI runs it with 200 clients on every push |
| Browser end-to-end | Real Chrome against the built program: Setup (sample show, edit conflict, Look), rehearsal isolation, live show with two languages, Back/Blank/Clear, songs and voices together, speaker colors, fonts, help pauses keys, scroll-back, language menu, Settings/password change, restart survival | `make e2e` (needs Node + Chrome; runs in CI as a separate job) |
| Manual device | iPhone Safari, iPad Safari, Android Chrome, desktop Chrome/Firefox/Edge, a projector, a full house on real WiFi | Acceptance checklist below |

## Key test cases

- Loops: per-item duration, loop default, config default; loop first in file auto-starts; STOP→blank→NEXT; Back after ending a loop; reconnect mid-loop shows the current item; images in the next loop are preloaded.
- Auth: separate passwords; admin session works on `/operator`, operator session rejected on `/admin`; no Basic-auth challenge; cookie `Secure` only over HTTPS; HTTP and HTTPS serve the same routes with no redirect.
- Show file: no markup → N captions; comments/blank lines ignored; CRLF; BOM; very long line; unicode/emoji/RTL passthrough; unclosed `[loop]`, raw `<b>` shown literally with an info diagnostic, `<!-- -->` comments (single- and multi-line, unclosed) never reach any client, `[note: …]` reaches only the operator view, `#` lines ignored with an info diagnostic; unknown directive; missing image with fuzzy suggestion; `\#` escape; `[[ ]]` multi-line; speaker detection positives/negatives.
- Help: every parser directive/escape/diagnostic is documented in `web/help/` (test); every fenced example parses without errors; sample show file parses clean; opening help on `/operator` pauses show-control keys; help works with no network.
- Escaping: `<script>`, `onerror=`, `javascript:` links, `<img src=http…>` in show file text are rendered as literal text; Markdown links are not clickable; image refs only resolve to uploaded assets.
- Engine: advance past end (no-op + signal "end of show file"); back at start; jump forward/backward equals sequential; `clear` mid-history then undo restores scrollback; fullscreen followed by caption dismisses fullscreen; restart restores cursor and scrollback.
- Concurrency: two operator sessions advancing with the same `rev` → exactly one NEXT applied, other gets 409; a NEXT sent while a loop ticks (seq changed, rev unchanged) succeeds.
- Rehearsal: commands on `/api/admin/rehearsal/*` never reach audience clients or change `live.json`, even while a real show is live.
- Security: first-run passwords never appear in any HTTP response; beyond `audience.max_clients`, `/events` returns 503.
- SSE: slow client doesn't delay others; overflow closes the slow client; heartbeat arrives; late joiner receives scrollback.

- Languages: a translation that does not match the main show file is reported with its own line number and is not offered; one keypress sends the same cue to every language and a language that drifts is moved back; negotiation by `?lang=`, `Accept-Language` quality values, regional fallback; devices on a language a new show lacks are told to reload; per-language state survives a restart.
- Songs and voices: `~` shorthand, `[song]`, `[together]` and ` || ` (including mixed spoken/sung, unclosed blocks, escapes).
- Security: access lists (listed, unlisted, bad entry denies all, cookie from an unlisted computer, admin address may run the show); forged `Origin`/`Sec-Fetch-Site` refused; sessions survive a restart but not a password change; the settings and support bundle never contain passwords.
- Persistence: undo steps and last-action time survive a restart; two people saving one show file is detected.

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
