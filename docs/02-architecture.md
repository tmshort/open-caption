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
                               │  - Hub (fanout) │        │  live.json   │
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
   The operator's keypress is a tiny `POST /api/operator/next`.
2. **State-snapshot-first protocol.** The server holds a single authoritative
   `LiveState`. Every event carries a monotonically increasing `seq`. A new or
   reconnecting client first receives a full snapshot, then deltas. Dropped
   events never leave a client wrong because the snapshot is always sufficient.
3. **Server is the only source of truth.** Clients render state; they never
   decide what comes next. Operator position is held server-side.
4. **No CGO, no DB.** Persistence is JSON files plus uploaded assets on disk,
   written atomically (write temp + rename). Fine for the data volumes involved.
5. **Standard library first.** `net/http` (Go 1.22+ pattern routing), `crypto/tls`, `embed`.
   Allowed dependencies (all Apache-2.0-compatible): a YAML parser
   (`gopkg.in/yaml.v3`) and a QR generator (`github.com/skip2/go-qrcode`), or equivalents.
   No HTML sanitizer is needed: scripts have no raw HTML; the server escapes all
   text and emits only the tags it generates from the Markdown subset (doc 03).
   Frontend is vanilla JS/CSS with no build step (low contributor barrier,
   self-contained binary).
6. **Loops are server-driven.** The engine owns the loop timer and publishes each
   item change as an event, so all devices stay in step and STOP is immediate.
7. **Plain HTTP and optional HTTPS, side by side.** Both listeners serve the same
   routes (doc 06).
8. **Fanout design for 200+ clients.** Hub keeps a set of client channels, each
   with a small buffered queue. A slow client that overflows its buffer is
   dropped and reconnects (receives a fresh snapshot). The event payload is
   serialized once and shared across all clients. 200 clients × few events/min
   is trivially small; the design goal is correctness under flaky WiFi.

## Domain model

- **Show**: id, name, created/updated, one script per language (primary language required; others planned), parsed `Cue[]`, `Assets[]`, `Theme`.
- **Cue**: `index`, `kind`, payload, source line. Kinds:
  `caption`, `clear`, `blank`, `image`, `fullscreen`, `loop`, `section` and `note` (operator-only). See doc 03.
- **LiveState** (the part broadcast to audience, per language):
  - `mode`: `idle | captions | blank`
  - `scrollback`: ordered caption items on the scrolling surface (bounded, e.g. last 200)
  - `fullscreen`: optional item (text or image) overlaying everything
  - `preload`: upcoming image URLs
  - `theme`, `lang`, `seq`
- **OperatorState** (operator-only): `cursor` (index of the next cue), `history` stack for undo, loop status, controller id, audience count.

### Applying a cue (semantics)

| Cue kind | Effect on audience state |
|----------|--------------------------|
| `caption` | Append to scrollback; any fullscreen item or running loop is dismissed and captions resume |
| `clear` | Empty the scrollback; dismiss fullscreen/loop; show the (empty) captions surface |
| `blank` | Mode `blank` (background only); scrollback retained |
| `image`, `fullscreen` | Fullscreen item until the next cue |
| `loop` | Start the loop: show item 0 fullscreen, advance on the server timer per item duration, wrap around, until the next NEXT or STOP |
| `section`, `note` | No audience effect; operator-only; auto-passed (never consume a keypress) |

- **NEXT** = apply `cue[cursor]` (skipping over `section`/`note`), then move
  `cursor` past it. Because a loop cue is consumed when it starts, the following
  NEXT sends the cue after the loop and thereby ends it.
- **STOP** (only meaningful while a loop runs): stop the loop timer, set `blank`.
  Cursor is unchanged.
- **Auto-start:** when a show is activated or reset, if the first audience-visible cue is a
  `loop` it is applied immediately (pre-show). Otherwise state is `idle` (idle message shown).
- **Undo (Back)** restores the previous LiveState snapshot from the history stack
  (bounded, e.g. 100); a loop that was running at that point restarts from item 0. This is how the operator restarts a loop: Back after STOP, or after NEXT ended it, "backs into" it.
- **Jump-to-cue** computes the state as if cues up to N were applied in order
  (scrollback rebuilt from the most recent `clear`; a loop cue at N-1 starts), so
  jumping is deterministic and never floods clients with intermediate events.

### Languages (planned)

Each language has its own parsed script of identical structure. The engine keeps
one cursor; applying a cue updates one `LiveState` per language, and the hub
routes events by the client's `lang`. v1 ships with a single language but the
event stream, `LiveState.lang`, and the show model already carry it.

**Language selection** (per request to `/`): an explicit `?lang=xx` wins; otherwise the
server matches the `Accept-Language` header (quality values respected, regional tags
fall back to the base language, e.g. `es-MX` → `es`) against the languages the active
show provides; otherwise it uses `defaults.language` (`en`). Responses carry
`Vary: Accept-Language`.

## Package layout (proposed)

```
cmd/open-caption/        main: flags, wiring, signal handling
internal/
  config/                config.yaml load/generate, flags/env overrides, data-dir resolution
  script/                parser, validator, Markdown-subset renderer, cue model (pure, heavily tested)
  show/                  show store (CRUD, assets), persistence
  live/                  engine: LiveState, cursor, history, loop timer, apply/jump/undo
  hub/                   SSE fanout, client registry, counters
  httpapi/               routes, auth middleware, handlers, HTTP/HTTPS listeners
  netinfo/               enumerate host IPs, QR code, self-signed cert helper
web/                     embedded static assets
  setup/  operator/  audience/  shared/
  help/                  help topics as Markdown (embedded), rendered by the same Markdown code as scripts
docs/
```

`script`, `live`, and `hub` must not import `httpapi`; they are testable
without a network.

## Authentication

- Two roles with **separate passwords**: `admin` (pages `/admin`) and `operator` (page `/operator`). An admin session also satisfies operator checks.
- Passwords live in **plaintext in `config.yaml`** (initial design; the file is created with mode 0600). On first run, if no config exists, the server generates it with random passwords and prints them to the console and shows them on the `/admin` login page the first time.
- Login is an HTML form posting to `/api/login`, producing an HttpOnly, SameSite=Strict cookie (`Secure` over HTTPS). **No HTTP Basic auth** (poor password-manager support).
- Audience routes (`/`, `/events`, `/assets/*`) are public and cannot reach any mutating endpoint.
- Intended for a trusted LAN; not hardened for the public internet (documented).

## Configuration (`config.yaml`)

Everything environment-specific is configurable here; command-line flags and
`OPEN_CAPTION_*` env vars override it.

```yaml
admin_password: "…generated…"
operator_password: "…generated…"
data_dir: ""                       # default: <user config dir>/open-caption
http:
  addr: ":80"
  fallback_addr: ":8080"           # used with a visible warning if addr can't be bound
https:
  enabled: false
  addr: ":443"
  fallback_addr: ":8443"
  cert_file: ""                    # supply both, or leave empty with self_signed: true
  key_file: ""
  self_signed: true
defaults:
  language: en
  loop_item_seconds: 10
  idle_message: "Captions will begin shortly"
  open_browser: true
audience:
  max_clients: 500
  scrollback_lines: 200
```

## Persistence layout

```
<data-dir>/
  config.yaml             settings and passwords (above)
  shows/<id>/
    show.yaml             metadata + theme
    script.<lang>.md      imported script(s), verbatim (source of truth)
    assets/<filename>     uploaded images
  live.json               active show id, cursor, scrollback, mode (written on each change, debounced ≤ 250 ms)
  tls/                    generated self-signed cert/key (if used)
```

Default data dir: OS user-config location (`os.UserConfigDir()/open-caption`).
The parsed cue list is derived from the script on load, not stored.

## Error handling principles

- Script parse errors never crash the server; invalid lines become warnings and are skipped or rendered as plain captions (spec in doc 03).
- Operator commands are idempotent where possible and carry the client's last-seen `seq`; a stale command (`seq` mismatch for jump/advance) returns 409 with the current state so two operator tabs can't double-advance.
- Never block the engine on a slow client.
