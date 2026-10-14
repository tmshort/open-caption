# 07 — Testing and Acceptance

## Test layers

| Layer | What | How |
|-------|------|-----|
| Unit | `script` parser/validator (table-driven, golden files in `testdata/`), Markdown renderer/escaping, config load/generate, `live` engine (apply/undo/jump determinism), speaker detection | `go test ./...` |
| Property/fuzz | `script.Parse` never panics, never emits unescaped HTML from input text; engine: `jump(n)` state == sequential apply state | `go test -fuzz` |
| Integration | httptest server: login, upload, advance, SSE stream receives events in order; reconnect with `Last-Event-ID` yields snapshot | Go tests |
| Load | N SSE clients (200, 500) + operator advancing at 1/s; measure delivery latency, memory, dropped clients | `cmd/loadtest` Go tool |
| Browser E2E | Playwright (or chromedp): operator spacebar → audience DOM updates; fullscreen/clear behavior; offline/online reconnect | CI optional/nightly |
| Manual device | iPhone Safari, iPad Safari, Android Chrome, desktop Chrome/Firefox/Edge, projector | Release checklist below |

## Key test cases

- Loops: per-item duration, loop default, config default; loop first in file auto-starts; STOP→blank→NEXT; Back after ending a loop; reconnect mid-loop shows the current item; images in the next loop are preloaded.
- Auth: separate passwords; admin session works on `/operator`, operator session rejected on `/admin`; no Basic-auth challenge; cookie `Secure` only over HTTPS; HTTP and HTTPS serve the same routes with no redirect.
- Script: no markup → N captions; comments/blank lines ignored; CRLF; BOM; very long line; unicode/emoji/RTL passthrough; unclosed `[loop]`, raw `<b>` shown literally with an info diagnostic, `<!-- -->` comments (single- and multi-line, unclosed) never reach any client, `[note: …]` reaches only the operator view, `#` lines ignored with an info diagnostic; unknown directive; missing image with fuzzy suggestion; `\#` escape; `[[ ]]` multi-line; speaker detection positives/negatives.
- Help: every parser directive/escape/diagnostic is documented in `web/help/` (test); every fenced example parses without errors; sample script parses clean; opening help on `/operator` pauses show-control keys; help works with no network.
- Escaping: `<script>`, `onerror=`, `javascript:` links, `<img src=http…>` in script text are rendered as literal text; Markdown links are not clickable; image refs only resolve to uploaded assets.
- Engine: advance past end (no-op + signal "end of script"); back at start; jump forward/backward equals sequential; `clear` mid-history then undo restores scrollback; fullscreen followed by caption dismisses fullscreen; restart restores cursor and scrollback.
- Concurrency: two operator sessions advancing with the same `rev` → exactly one NEXT applied, other gets 409; a NEXT sent while a loop ticks (seq changed, rev unchanged) succeeds.
- Rehearsal: commands on `/api/admin/rehearsal/*` never reach audience clients or change `live.json`, even while a real show is live.
- Security: first-run passwords never appear in any HTTP response; beyond `audience.max_clients`, `/events` returns 503.
- SSE: slow client doesn't delay others; overflow closes the slow client; heartbeat arrives; late joiner receives scrollback.

## Performance targets (verified by load test)

- 200 concurrent audience connections: operator keypress → last client receives event p95 ≤ 250 ms, p99 ≤ 500 ms.
- Server RSS < 150 MB at 200 clients; CPU < 10% of one core while idle-connected.
- 500 clients degrade gracefully (no crash); connections beyond `audience.max_clients` (default 500) receive 503 and retry with backoff.

## Acceptance checklist (pre-release, in venue conditions if possible)

1. [ ] Fresh install on macOS, Windows, Linux: first run generates `config.yaml` and reaches the `/admin` login with no errors.
2. [ ] A volunteer unfamiliar with the project imports a script and images using only the UI hints within 15 minutes.
3. [ ] Malformed script shows clear line-numbered diagnostics and still previews.
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
