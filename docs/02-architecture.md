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
4. **No CGO, no DB.** Persistence is JSON/YAML files plus uploaded assets on disk,
   written atomically (write temp + rename). Fine for the data volumes involved.
5. **Standard library first.** `net/http` (Go 1.22+ pattern routing), `crypto/tls`, `embed`.
   Allowed dependencies (all Apache-2.0-compatible): a YAML parser
   (`gopkg.in/yaml.v3`) and a QR generator (`github.com/skip2/go-qrcode`), or equivalents.
   No HTML sanitizer is needed: show files have no raw HTML; the server escapes all
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

- **Show**: id, name, created/updated, one show file per language (primary language required; others planned), parsed `Cue[]`, `Assets[]`, `Theme`.
- **Cue**: `index`, `kind`, payload, source line. Kinds:
  `caption`, `clear`, `blank`, `image`, `fullscreen`, `loop`, `section` and `note` (operator-only). See doc 03.
- **LiveState** (the part broadcast to audience, per language):
  - `mode`: `idle | captions | blank`
  - `scrollback`: ordered caption items on the scrolling surface (bounded, e.g. last 200)
  - `fullscreen`: optional item (text or image) overlaying everything
  - `preload`: upcoming image URLs
  - `theme`, `lang`, `seq`
- **OperatorState** (operator-only): `cursor` (index of the next cue), `history` stack for undo, loop status, controller id, audience count.

**Live** means: the active show's engine has applied at least one cue (or a loop is
running) since it was activated or last reset. Activating a show or `reset` returns it
to its initial state (not live). REQ-052's guard rails and the activate-`force` rule use this definition.

**Revision (`rev`)**: `LiveState.seq` increases on *every* audience-visible change,
including each loop item change. Operator commands therefore use a separate counter,
`OperatorState.rev`, which increases only when an operator command is applied
(NEXT, Back, STOP, jump, blank, clear, message, reset, activate, takeover) — never on loop ticks.

### Applying a cue (semantics)

| Cue kind | Effect on audience state |
|----------|--------------------------|
| `caption` | Append to scrollback; any fullscreen item or running loop is dismissed and captions resume |
| `clear` | Empty the scrollback; dismiss fullscreen/loop; show the (empty) captions surface |
| `blank` | Mode `blank` (background only); scrollback retained |
| `image`, `fullscreen` | Fullscreen item until the next cue |
| `loop` | Start the loop: show item 0 fullscreen, advance on the server timer per item duration, wrap around, until the next NEXT or STOP |
| `section`, `note` | No audience effect; operator-only; auto-passed (never consume a keypress) |

Applying **any** audience-visible cue other than `loop` itself (caption, clear, blank, image, fullscreen) ends a running loop.

- **NEXT** = apply `cue[cursor]` (skipping over `section`/`note`), then move
  `cursor` past it. Because a loop cue is consumed when it starts, the following
  NEXT sends the cue after the loop and thereby ends it.
- **STOP** (only meaningful while a loop runs): stop the loop timer, set `blank`.
  Cursor is unchanged.
- **Auto-start:** when a show is activated or reset, if the first audience-visible cue is a
  `loop` it is applied immediately (pre-show). Otherwise state is `idle` (idle message shown).
- **Ad-hoc operator commands** (blank toggle, clear, one-off message) change audience state without moving the cursor; they are pushed on the history stack like cues, so Back undoes them.
- **Undo (Back)** restores the previous LiveState snapshot from the history stack
  (bounded, e.g. 100); a loop that was running at that point restarts from item 0. This is how the operator restarts a loop: Back after STOP, or after NEXT ended it, "backs into" it.
- **Jump-to-cue** computes the state as if cues up to N were applied in order
  (scrollback rebuilt from the most recent `clear`; a loop cue at N-1 starts), so
  jumping is deterministic and never floods clients with intermediate events.

### Rehearsal

The setup wizard's Test step runs the operator UI against an **isolated engine
instance** owned by the admin session. It uses the same code and endpoints shape
as live control (`/api/admin/rehearsal/*`, doc 04) but publishes only to that
session's preview pane, never to audience clients, and never touches `live.json`.
It can't affect, or be affected by, the live show.

### Languages

Each language of the active show has its **own `Engine`** (its own cues, state, loop timer and sequence
numbers) and its **own audience hub**. The operator drives the primary engine; `Runtime.Do` applies the
same command to every other engine immediately afterwards, so all languages stay on the same cue index.
Their rev, scrollback and loop positions stay equal because the structure is identical and the commands
are the same; after each command, any translation whose cursor differs from the primary's is moved to it
with `Jump`, so a language can never drift. Hubs are keyed by language code and kept across shows, so a
connected device stays on its stream when the next show also has that language; a device whose language
the new show lacks is sent a `reload` event and picks a language again.

Edits apply to every language together (`Runtime.ReplaceLangs`): when the show is live, each engine keeps
its place by cue index; a language added mid-show is created and jumped to the current cue; a removed one
closes and its devices are told to reload. Live state is persisted per language and restored the same way.

**Language selection** (per request to `/events`): an explicit `?lang=xx` wins; otherwise the
server matches the `Accept-Language` header (quality values respected, regional tags
fall back to the base language, e.g. `es-MX` → `es`, and a plain `pt` is served `pt-br` if that is
the only Portuguese) against the languages the active show provides; otherwise it uses the show's main
language. Responses carry `Vary: Accept-Language`.

## Package layout

```
cmd/open-caption/        main: flags, wiring, signal handling, listeners (HTTP and optional HTTPS)
cmd/loadtest/            audience load generator (needs --dev on the server)
internal/
  config/                config.yaml load/generate/validate, access allow-lists, data-dir resolution
  show file/                parser, validator, Markdown-subset renderer, cue model, language structure check (pure, heavily tested)
  show/                  show store: shows, show files per language, validated image assets, duplicate
  live/                  engine: state, cursor, undo history, loop timer, apply/jump/back, export/restore
  hub/                   SSE fanout, client registry, non-blocking broadcast
  httpapi/               routes, auth + sessions, access/origin guards, runtime (engines + hubs per language), admin/operator/audience handlers
  help/                  help topic loader, safe Markdown renderer, search
  netinfo/               host addresses, self-signed certificate
  logbuf/                in-memory recent log lines (support bundle)
web/                     embedded browser assets (go:embed)
  admin/ operator/ audience/ helpui/   the four pages
  shared/                speaker colors and the bundled fonts (OpenDyslexic, Atkinson Hyperlegible)
  help/                  help topics (Markdown); docs/ guides are added to them at build time
  sample/                the sample show (show file, Spanish translation, placeholder images)
docs/                    these specs, the usage guide and venue network guide (docs/embed.go ships the guides in the program)
deploy/                  systemd / launchd / Windows service files
test/e2e/                browser end-to-end test (Node + Chrome)
```

`show file`, `live`, `hub`, `show` and `help` must not import `httpapi`; they are testable without a network.

## Authentication

- Two roles with **separate passwords**: `admin` (pages `/admin`) and `operator` (page `/operator`). An admin session also satisfies operator checks.
- Passwords live in **plaintext in `config.yaml`** (initial design; the file is created with mode 0600). On first run, if no config exists, the server generates it with random passwords and prints them to the console.
- Login is an HTML form posting to `/api/login`, producing an HttpOnly, SameSite=Strict cookie (`Secure` over HTTPS). **No HTTP Basic auth** (poor password-manager support).
- Audience routes (the public table in doc 04: `/`, `/events`, `/assets/*`, `/healthz`, `/captive-api`) are open and cannot reach any mutating endpoint. Help content requires a session.
- **Sessions survive a restart** (`sessions.json`, hashed tokens, discarded when a password changes), so an operator is not signed out if the program restarts mid-show. Passwords can be changed from the Setup page (written to `config.yaml`, effective at once).
- **Access lists** (`access.admin_allow`, `access.operator_allow`): which computers may open Setup / the operator page (and sign in to them). Entries: `localhost`, `private`, an address, or a network. An empty list lets anyone try the password (the default); an unreadable list denies everything. The check is made on pages, JavaScript, APIs and login, so a session cookie does not work from an unlisted computer. An admin address may also use the operator pages. The client address is the TCP peer; proxy headers are not trusted.
- **Forged requests:** state-changing `/api/*` requests are refused if the browser says they came from another site or another port of this computer (`Origin` not equal to `Host`, or `Sec-Fetch-Site` cross-site/same-site). Requests with no `Origin` (command-line tools) are allowed; the cookie is also SameSite=Strict.
- The generated first-run passwords are printed to the **console only**; they are never sent to a browser, because the login page is reachable from the guest network. (The config file itself is the other place to find them.)
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
defaults:                          # used when a show file does not say otherwise (a show file's [settings] override these)
  language: en
  loop_item_seconds: 10
  idle_message: "Captions will begin shortly"
  open_browser: true               # (environment: not a show setting)
  look:
    background: "#000000"
    text_color: "#ffffff"
    speaker_name_color: "#ffffff"
    font: system                   # system | serif | mono | dyslexic | hyperlegible
    text_size_percent: 100
    speaker_names: label           # label | color | hidden
audience:
  max_clients: 500                 # per language
  scrollback_lines: 200
access:
  admin_allow: []                  # [] = any computer; e.g. [localhost] or [192.168.1.0/24, private]
  operator_allow: []
```

## Where configuration lives (layers)

| What | Where | Notes |
|------|-------|-------|
| Environment: ports, HTTPS, passwords, access lists, data folder, audience limits, whether to open a browser | `config.yaml` only | Flags and `OPEN_CAPTION_*` override |
| Show defaults: look, idle message, loop time, scrollback size, main language | built-in values, overridden by `config.yaml` (`defaults`, `audience.scrollback_lines`) | `config.Load` starts from the built-in values, so a file that omits a key keeps the built-in one |
| The show itself: title, look, idle message(s), loop time, scrollback size, main language, and all lines | the show file(s) | Overrides the defaults above. Nothing else about a show is stored outside its show file(s) except its images |

`show.Defaults.Resolve(showfile.Settings)` layers a show file's settings over the defaults. Setup's Look step and Rename write
the show file's `[settings]` block (`showfile.UpdateSettings`); the server then re-reads the files and applies the result to the
show on air (`Server.applyActive`: look to every device, title to the operator, loop time and scrollback size to the engines,
and every language's lines).

## Persistence layout

```
<data-dir>/
  config.yaml             settings and passwords (above)
  shows/<id>/
    show.md               the show file, main language: title, [settings] (look, loop time, …) and the lines — the source of truth
    show.<lang>.md        one show file per translation (its own idle message, and the lines)
    assets/<filename>     uploaded images
  sessions.json           signed-in sessions (hashed), so a restart does not sign anyone out
  live.json               active show id and, per language: cursor, scrollback, mode, the last 10 undo steps (debounced ≤ 250 ms)
  tls/                    generated self-signed cert/key (if used)
```

Default data dir: OS user-config location (`os.UserConfigDir()/open-caption`).
The parsed cue list is derived from the show file on load, not stored.

## Error handling principles

- Show file parse errors never crash the server; invalid lines become warnings and are skipped or rendered as plain captions (spec in doc 03).
- Operator commands carry the client's last-seen `rev` (not `seq`, which changes on every loop tick); a stale `rev` returns 409 with the current state so two operator tabs can't double-advance, while a loop ticking in the background never causes a spurious 409.
- Beyond `audience.max_clients`, `/events` returns 503 and the page retries with backoff.
- Never block the engine on a slow client.
- The show-level idle message (Look step) overrides `defaults.idle_message`.
