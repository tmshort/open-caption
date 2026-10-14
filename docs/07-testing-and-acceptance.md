# 07 — Testing and Acceptance

## Test layers

| Layer | What | How |
|-------|------|-----|
| Unit | `script` parser/validator (table-driven, golden files in `testdata/`), sanitizer, `live` engine (apply/undo/jump determinism), speaker detection | `go test ./...` |
| Property/fuzz | `script.Parse` never panics, never emits unsanitized HTML; engine: `jump(n)` state == sequential apply state | `go test -fuzz` |
| Integration | httptest server: login, upload, advance, SSE stream receives events in order; reconnect with `Last-Event-ID` yields snapshot | Go tests |
| Load | N SSE clients (200, 500) + operator advancing at 1/s; measure delivery latency, memory, dropped clients | `cmd/loadtest` Go tool |
| Browser E2E | Playwright (or chromedp): operator spacebar → audience DOM updates; fullscreen/clear behavior; offline/online reconnect | CI optional/nightly |
| Manual device | iPhone Safari, iPad Safari, Android Chrome, desktop Chrome/Firefox/Edge, projector | Release checklist below |

## Key test cases

- Script: no markup → N captions; comments/blank lines ignored; CRLF; BOM; very long line; unicode/emoji/RTL passthrough; unclosed `[html]`; unknown directive; missing image with fuzzy suggestion; `\#` escape; `[[ ]]` multi-line; speaker detection positives/negatives.
- Sanitizer: `<script>`, `onerror=`, `javascript:` URLs, `style` abuse, external `<img src=http…>` all stripped.
- Engine: advance past end (no-op + signal "end of script"); back at start; jump forward/backward equals sequential; `clear` mid-history then undo restores scrollback; fullscreen followed by caption dismisses fullscreen; restart restores cursor and scrollback.
- Concurrency: two operator sessions advancing simultaneously → exactly one advance applied, other gets 409.
- SSE: slow client doesn't delay others; overflow closes the slow client; heartbeat arrives; late joiner receives scrollback.

## Performance targets (verified by load test)

- 200 concurrent audience connections: operator keypress → last client receives event p95 ≤ 250 ms, p99 ≤ 500 ms.
- Server RSS < 150 MB at 200 clients; CPU < 10% of one core while idle-connected.
- 500 clients degrade gracefully (no crash).

## Acceptance checklist (pre-release, in venue conditions if possible)

1. [ ] Fresh install on macOS, Windows, Linux: first run reaches the PIN screen with no errors.
2. [ ] A volunteer unfamiliar with the project imports a script and images using only the UI hints within 15 minutes.
3. [ ] Malformed script shows clear line-numbered diagnostics and still previews.
4. [ ] Operator advances through a full show with Space only; no missed or duplicated lines.
5. [ ] Image cue appears fullscreen on audience devices; operator sees only the indicator.
6. [ ] `[clear]`, `[blank]`, intermission fullscreen text behave as in doc 02 table.
7. [ ] Pre-show slideshow loops, then operator moves into the show.
8. [ ] 75+ real/simulated devices connected; all in sync; audience count correct.
9. [ ] Toggle airplane mode on a phone for 20 s mid-show: it recovers to the current state unaided.
10. [ ] Kill and restart the server mid-show: clients reconnect; operator resumes at the same cue.
11. [ ] Locking and unlocking an iPad mid-show returns to current state.
12. [ ] Projector mode readable from the back of the room.
13. [ ] A second operator tab is view-only until takeover.
14. [ ] Audience cannot reach any `/api/*` mutating endpoint or `/admin`.
