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
- **Image URLs**: each active show's referenced image has an in-memory version. Replacing one changes its URL so browser caches fetch the new bytes. Upcoming-image hints change only when their URLs change; viewers receive a new snapshot only when that image is on their screen. An image referenced by the show on air cannot be deleted until its show file references are removed; deletion in standby is allowed.
- **OperatorState** (operator-only): `cursor` (index of the next cue), `history` stack for undo, loop status, controller id, audience count.

**On air, standby, off air (REQ-066).** A show that is loaded (activated) waits in **standby**: the audience is
told `mode: "standby"` and sees the *venue's* standby screen (a logo and a message, see below), never anything of
the show, so Setup, rehearsal and editing are never in front of the audience. The operator's **START** (`onair`)
puts the show on air; **Go to standby** (`offair`) takes it off again **without losing its place** (the lines,
cursor and undo history are kept underneath, and START puts it back as it was; a loop that was running restarts
from its first item). `reset` (Restart show) and activating a show return it to the beginning **and to standby**.
While in standby every command that changes what the audience sees (next, back, stop, jump, blank, clear,
message) is refused with 409 `the show is not on air`; only START, Restart and editing work.

**Live** means **on air**: the engine is not in standby. REQ-052's guard rails (type the show name to confirm edits,
no delete or archive, the activate-`force` rule) apply only then, so a show in standby can be edited, deleted or
replaced freely. (Engines created without `Config.Standby`, used by `--dev` and the engine's unit tests, are on air at
once and mean by Live that a cue has been sent, as before.) The standby flag and the loop to resume are saved in
`live.json`; a restart keeps a show in standby in standby and a show on air on air (files from before this
existed carry on as they were, on air).

**The standby screen** is venue-level, not part of any show file (a show file can't be read when no show is loaded):
the optional message is `standby.message` in `config.yaml` (empty means the built-in "Captions will begin shortly"
in each viewer's language) and the logo is the built-in **[oc]** mark (open captioning, the same art as the favicon, served from `/static/shared/favicon.svg`) until the venue uploads its own PNG, JPEG, GIF or WebP of up to 2 MB, kept in the data folder
(`standby-logo.<ext>`), both set in Setup → Settings → Audience screen and served publicly at `/standby-logo`. It is
shown whenever the audience mode is `standby`, including before any show has been loaded.

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
- **SKIP** = move `cursor` past the next audience-visible cue *without* applying it: nothing is emitted, the audience's screen and "last sent" are unchanged, and the step is pushed on the undo history so Back takes it back. The cue is remembered in `skipped` (shown by the operator page, saved in `live.json`, cleared by a jump back past it, a restart of the show or an edit that changes the number of cues). A running loop keeps running; at the end of the show it is `ErrEnd` like NEXT.
- **STOP** (only meaningful while a loop runs): stop the loop timer, set `blank`.
  Cursor is unchanged.
- **Pre-show:** if the first audience-visible cue is a `loop`, it starts when the show **goes on air** (START): the
  show is in standby until then, so nothing is shown at activation or reset. Without a leading loop, START shows the
  show's own idle message (`mode: idle`) until the first line. (Without `Config.Standby`, as in `--dev`, the loop
  starts at activation.)
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
  acme/                  automatic HTTPS certificates (ACME DNS-01): the certificate manager, the lego-based issuer, the DNS providers chosen at build time (doc 10)
  check/                 the show file analysis shared by Setup's editor and `--check-show` (image use, the command-line report)
  config/                config.yaml load/generate/validate (strict: unknown keys and all bad values reported together), access allow-lists, data-dir resolution
  help/                  help topic loader, safe Markdown renderer, search
  httpapi/               routes, auth + sessions, access/origin guards, runtime (engines + hubs per language), admin/operator/audience handlers
  hub/                   SSE fanout, client registry, non-blocking broadcast
  live/                  engine: state, cursor, undo history, loop timer, apply/jump/back, export/restore
  logbuf/                in-memory recent log lines (support bundle)
  logx/                  levelled logging (DEBUG/INFO/WARN/ERROR) over the standard logger
  mdns/                  multicast-DNS responder: announces <name>.local (A records, probing for conflicts, goodbyes)
  netinfo/               host addresses (interfaces that are up with their kind, Ethernet/Wi-Fi/other, main connection first, the `network.interfaces` ticks, one address per network), self-signed certificate, the watched certificate files (reload without a restart)
  report/                the audience report: counts the devices (address + browser) that used the audience page during a run of the show; saves each run as JSON in <data>/reports; CSV
  show/                  show store: shows, show files per language, validated image assets, duplicate
  showfile/              the show file: parser, validator, Markdown-subset renderer, cue model, language structure check (pure, heavily tested)
web/                     embedded browser assets (go:embed)
  admin/ operator/ audience/ helpui/   the four pages
  shared/                the interface shared by Setup and the operator page (ui.css, header.css), speaker colors, the bundled fonts (OpenDyslexic, Atkinson Hyperlegible), and i18n/<code>.js: browser-side translation (the English text is the key)
  help/                  help topics (Markdown); help/<code>/ the translations (es fr de it pt)
  sample/                the sample show (show file, Spanish translation, placeholder images)
docs/                    the specifications (01 to 10), the guides and the man page; docs/embed.go ships the three guides in the program, docs/i18n/<code>/ holds their translations
deploy/                  systemd / launchd / Windows service files
tools/                   make-icons.js (the icons), acme-go.sh (the DNS provider choice for builds)
test/e2e/                browser end-to-end test, accessibility and translation checks (Node + Chrome)
```

`showfile`, `live`, `hub`, `show`, `help`, `check`, `config`, `acme` and `netinfo` must not import `httpapi`; they are testable without a network.

## Authentication

- Two roles with **separate passwords**: `admin` (pages `/admin`) and `operator` (page `/operator`). An admin session also satisfies operator checks.
- **Roles** are ranked `viewer` < `operator` < `admin` (`roleRank`; `require(min, h)` refuses a lower one with 403). The page `GET /follow` (the operator page script, which sees `/follow` in its address and drops every control) signs in with `POST /api/login {role: "viewer"}`, which takes the viewer, operator **or** admin password, each session keeping the role of its password (the passwords are a hierarchy; a viewer password never matches when empty). The operator form no longer takes the viewer password, and a session whose role is `viewer` that opens `/operator` is redirected to `/follow`. A viewer can call just `GET /api/operator/state` and `/stream`, `/api/me` and the help, and never holds control (`follower` mode never claims it). Sessions of a role are dropped when its password changes; the sessions fingerprint is unchanged when there is no viewer password.
- Passwords live in `config.yaml` as a digest written `sha256:<hex>` by default (the file is created with mode 0600; also `sha512:`, `sha1:`; `plain:` forces plain text): `config.CheckPassword` accepts either form, `auth` keeps the stored form and compares in constant time, and a password set in Setup is always stored as a SHA-256 digest (`config.StoredLike`), except that one replacing a value written with `plain:` stays plain. Setup has no way to show a password. `open-caption --hash-password` makes a digest. On first run, if no config exists, the server generates random passwords, writes their digests and prints the passwords once to the console (`Config.FirstRun`, never saved and cleared by `Redacted`).
- Login is an HTML form posting to `/api/login`, producing an HttpOnly, SameSite=Strict cookie (`Secure` over HTTPS). **No HTTP Basic auth** (poor password-manager support).
- Audience routes (the public table in doc 04: `/`, `/events`, `/assets/*`, `/healthz`, `/captive-api`) are open and cannot reach any mutating endpoint. Help content requires a session.
- **Sessions survive a restart** (`sessions.json`, hashed tokens, discarded when a password changes), so an operator is not signed out if the program restarts mid-show. Passwords can be changed from the Setup page (written to `config.yaml`, effective at once). The file is edited in place (`config.SetValues`): comments, key order and settings that were never written stay as the person left them, and an edit that would make the file invalid is refused and undone.
- **Access lists** (`access.admin_allow`, `access.operator_allow`, and `access.viewer_allow` for the Follow page): which computers may open Setup / the operator page / the Follow page (and sign in to them). The pages are a hierarchy (`allowedFor`): a computer allowed for a higher page may open the lower ones too, so the operator page is open to the operator and Setup lists and `/follow` to the Follow list as well; an empty list lets any computer in. The operator and admin passwords, which the Follow form also takes, are accepted only from a computer allowed for the operator page. Entries: `localhost`, `private`, an address, or a network. `admin_allow` defaults to `[localhost]` (Setup only on the computer that runs the program); an empty list lets anyone try the password and logs a warning, and `operator_allow` is empty by default; an unreadable list denies everything. The check is made on pages, JavaScript, APIs and login, so a session cookie does not work from an unlisted computer. An admin address may also use the operator pages. The client address is the TCP peer; proxy headers are not trusted. Setup can change the three lists and `local_auto_login` while the program runs (`PUT /api/admin/settings/access`, `Server.setAccess`): the lists are read through `Server.allowLists` under `accessMu`, the change is written to `config.yaml`, a list for Setup that would exclude the requesting address is refused (409), and turning `local_auto_login` off ends automatic sign-ins at once.
- **Forged requests:** state-changing `/api/*` requests are refused if the browser says they came from another site or another port of this computer (`Origin` not equal to `Host`, or `Sec-Fetch-Site` cross-site/same-site). Requests with no `Origin` (command-line tools) are allowed; the cookie is also SameSite=Strict.
- The generated first-run passwords are printed to the **console only**; they are never sent to a browser, because the login page is reachable from the guest network. (The config file itself is the other place to find them.)
- Intended for a trusted LAN; not hardened for the public internet (documented).

## Configuration (`config.yaml`)

Everything environment-specific is configurable here; command-line flags and
`OPEN_CAPTION_*` env vars override it. **The complete reference, every setting with its default and rules, is the [configuration reference](configuration-reference.md);** the sample below is only an overview.

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
  cert_file: ""                    # supply both, or leave empty with self_signed: true; renewed files are picked up without a restart
  key_file: ""
  self_signed: true
  acme:                            # automatic certificates (docs/10-acme-lets-encrypt.md); off unless enabled: true
    enabled: false
    domain: captions.example.org   # a public name you own; email, agree_to_terms, directory, renew_before_days and dns {provider, credentials,
                                   # credentials_file, propagation_wait} are described in doc 10
mdns:
  enabled: true                    # announce <name>.local on the local network
  name: open-caption
network:
  interfaces: []                   # the network interfaces that are used (names: Ethernet, Wi-Fi, en0); empty = no choice yet, all are used; a new one stays off once a choice is made
  restrict_listening: false        # also refuse connections made to an address of an interface that is not used (localhost is always served)
standby:
  message: ""                      # shown while no show is on air; empty = "Captions will begin shortly" in the viewer's language (the logo is set in Setup)
setup:
  tour_done: false                 # the first-run tour of Setup has been finished or skipped (Setup sets it; Help → Take the tour replays it)
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
    speaker_position: above        # above | beside | column
audience:
  max_clients: 500                 # per language
  scrollback_lines: 200
access:
  admin_allow: [localhost]         # default; [] = any computer; e.g. [192.168.1.0/24, private]
  operator_allow: []
  local_auto_login: true           # a browser on this computer (localhost, 127.0.0.1, [::1]) is signed in as admin with no password; other computers use a password
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
  live.json               active show id and, per language: cursor, scrollback, mode, the last 10 undo steps (debounced ≤ 250 ms; written through a temporary file and renamed, one save at a time; a failed save is logged)
  tls/                    generated self-signed cert/key (if used)
  reports/YYYYMMDD-HHMMSS.nnnnnnnnn-<show>.json   the audience reports (REQ-070), one file for each run of a show; nanoseconds distinguish rapid restarts of the same show. The file is written as the run goes and when it ends; the newest 30 are kept (`report.Keep`), older ones are removed when a report is saved. They list the devices' addresses and nothing depends on them, so they can be deleted by hand
```

Default data dir: OS user-config location (`os.UserConfigDir()/open-caption`).
The parsed cue list is derived from the show file on load, not stored.
Going off air, resetting, or re-activating a show closes its report immediately, even if the same show ID goes on air again before the next two-second report tick. Activating into standby starts no new report until START. CSV download names use seconds for readability; the stored JSON files retain nanoseconds.

## Error handling principles

- Show file parse errors never crash the server; invalid lines become warnings and are skipped or rendered as plain captions (spec in doc 03).
- Operator commands carry the client's last-seen `rev` (not `seq`, which changes on every loop tick); a stale `rev` returns 409 with the current state so two operator tabs can't double-advance, while a loop ticking in the background never causes a spurious 409.
- Beyond `audience.max_clients`, `/events` returns 503 and the page retries with backoff.
- Never block the engine on a slow client.
- The show-level idle message (Look step) overrides `defaults.idle_message`.
