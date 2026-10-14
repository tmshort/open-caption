# 02 — Architecture

## Overview

```
 Operator browser ──HTTP POST commands──┐
 Setup browser ────HTTP/REST───────────┤
                                        ▼
                               ┌─────────────────┐        ┌──────────────┐
                               │  Go server      │◄──────►│  data dir    │
                               │  - HTTP router  │        │  shows/…     │
                               │  - Show engine  │        │  assets/…    │
                               │  - Hub (fanout) │        │  state.json  │
                               └────────┬────────┘        └──────────────┘
                                        │ SSE (one-way, server → client)
              ┌─────────────────────────┼───────────────────────┐
              ▼                         ▼                       ▼
        Phone/iPad               Desktop browser          Projector browser
```

## Key decisions

1. **Transport: Server-Sent Events (SSE) for downstream, plain HTTP POST for commands.**
   Audience traffic is strictly server→client. SSE is simple, built into browsers
   with automatic reconnect (`Last-Event-ID`), works on iOS Safari, needs no
   third-party library in Go, and proxies cleanly. WebSocket is not needed.
   The operator's keypress is a tiny `POST /api/live/advance`.
2. **State-snapshot-first protocol.** The server holds a single authoritative
   `LiveState`. Every event carries a monotonically increasing `seq`. A new or
   reconnecting client first receives a full snapshot, then deltas. Dropped
   events never leave a client wrong because the snapshot is always sufficient.
3. **Server is the only source of truth.** Clients render state; they never
   decide what comes next. Operator position is held server-side.
4. **No CGO, no DB.** Persistence is JSON files plus uploaded assets on disk,
   written atomically (write temp + rename). Fine for the data volumes involved.
5. **Standard library first.** `net/http` (Go 1.22+ pattern routing), `html/template`
   not needed beyond a shell, `embed`. Allowed dependencies: an HTML sanitizer
   (`github.com/microcosm-cc/bluemonday`) and a QR generator
   (`github.com/skip2/go-qrcode`), or equivalents. Frontend is vanilla JS/CSS
   with no build step (keeps contributor barrier low and the binary self-contained).
6. **Fanout design for 200+ clients.** Hub keeps a set of client channels, each
   with a small buffered queue. A slow client that overflows its buffer is
   dropped and reconnects (receives a fresh snapshot). The event payload is
   serialized once and shared across all clients. 200 clients × few events/min
   is trivially small; the design goal is correctness under flaky WiFi.

## Domain model

- **Show**: id, name, created/updated, script source text, parsed `Cue[]`, `Assets[]`, `Theme`.
- **Cue**: `index`, `kind`, `payload`, `source line`, optional `section`/`speaker`. Kinds:
  `caption`, `clear`, `blank`, `image`, `fullscreen_text`, `fullscreen_html`, `section` (operator-only marker), `slideshow`.
  See doc 03.
- **LiveState** (the part broadcast to audience):
  - `mode`: `preshow | captions | fullscreen | blank`
  - `scrollback`: ordered list of caption items currently on the scrolling surface (bounded, e.g. last 200)
  - `fullscreen`: optional `{kind: text|html|image, content}` overlaying everything
  - `theme`
  - `seq`
- **OperatorState** (operator-only): `cursor` (index of the next cue), `history` stack for undo, controller id, audience count.

### Applying a cue (semantics)

| Cue kind | Effect on audience state |
|----------|--------------------------|
| `caption` | Append to scrollback; if fullscreen was active, fullscreen is dismissed and captions resume |
| `clear` | Empty the scrollback; dismiss fullscreen; show captions surface (empty) |
| `blank` | Mode `blank` (solid background, nothing shown); scrollback retained |
| `image` | Fullscreen image until next cue that changes it |
| `fullscreen_text/html` | Fullscreen text/HTML until next cue |
| `slideshow` | Fullscreen looping images, timer-driven client-side, until next cue |
| `section` | No audience effect; operator marker (e.g. "Act 2 Scene 1") |

Advance = apply `cue[cursor]`, then `cursor++`. Undo restores the previous
LiveState snapshot from the history stack (bounded, e.g. 100).
Jump-to-cue replays state: it computes the resulting state as if cues up to N
were applied in order (scrollback rebuilt from the most recent `clear`), so
jumping is deterministic.

## Package layout (proposed)

```
cmd/open-caption/        main: flags, wiring, signal handling
internal/
  config/                flags/env/data-dir resolution, first-run settings
  script/                parser, validator, cue model (pure, heavily tested)
  show/                  show store (CRUD, assets), persistence
  live/                  engine: LiveState, cursor, history, apply/jump/undo
  hub/                   SSE fanout, client registry, counters
  httpapi/               routes, auth middleware, handlers
  sanitize/              HTML allow-list sanitizer wrapper
  netinfo/               enumerate host IPs, QR code
web/                     embedded static assets
  setup/  operator/  audience/  shared/
docs/
```

`script`, `live`, and `hub` must not import `httpapi`; they are testable
without a network.

## Authentication

- On first run the server prompts (setup wizard) for an operator PIN; stored hashed (argon2id or bcrypt) in `config.json`.
- Login sets an HttpOnly session cookie. `/api/admin/*`, `/setup`, `/operator` require it.
- Audience routes (`/`, `/events`, `/assets/*`) are public. Audience cannot reach any mutating endpoint.
- Intended for trusted LAN; not hardened for the public internet (documented).

## Persistence layout

```
<data-dir>/
  config.json             pin hash, listen addr, active show id, defaults
  shows/<id>/
    show.json             metadata + theme
    script.txt            original imported text (source of truth)
    assets/<filename>     uploaded images
  live.json               active show id, cursor, scrollback, mode (written on each change, debounced ≤ 250 ms)
```

Default data dir: OS user-config location (`os.UserConfigDir()/open-caption`), overridable with `--data`.
The parsed cue list is derived from `script.txt` on load, not stored.

## Error handling principles

- Script parse errors never crash the server; invalid lines become warnings and are skipped or rendered as plain captions (spec in doc 03).
- Operator commands are idempotent where possible and carry the client's last-seen `seq`; a stale command (`seq` mismatch for jump/advance) returns 409 with the current state so two operator tabs can't double-advance.
- Never block the engine on a slow client.
