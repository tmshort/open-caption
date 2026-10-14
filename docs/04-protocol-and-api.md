# 04 — Protocol and API

Base URLs: `http://<host>[:port]/` (default port **80**, fallback 8080) and,
when enabled, `https://<host>[:port]/` (default 443, fallback 8443). Both
listeners serve identical routes; there is no HTTP→HTTPS redirect (doc 06).
All JSON is UTF-8.

## Routes

### Audience (public)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Audience page (single HTML shell + JS) |
| GET | `/events` | SSE stream of live state (below) |
| GET | `/assets/{showId}/{file}` | Image assets of the active show (cacheable) |
| GET | `/healthz` | `200 ok` |
| GET | `/captive-api` | RFC 8908 captive-portal JSON (doc 06) |

Query parameters on `/` and `/events`: `?display=projector`, `?size=<pct>`,
`?theme=light|dark`, `?lang=<code>` (planned, doc 03). If `lang` is absent, the server picks one from the
`Accept-Language` request header, falling back to `defaults.language` (`en`).

### Help — admin or operator session

| Method | Path | Description |
|--------|------|-------------|
| GET | `/help` | Help viewer page (topic list + content); `?role=admin\|operator` picks the default topic set |
| GET | `/help/{topic}` | One topic as rendered HTML fragment (`?format=md` returns the Markdown source) |
| GET | `/help/sample-script.md` | Downloadable starter script demonstrating every convention |

Help is embedded in the binary (no network needed). It contains no secrets, but
is served only to logged-in sessions so the audience never sees operator tooling.

### Pages and authentication

| Path | Who | Notes |
|------|-----|-------|
| `/admin` | admin password | Setup UI |
| `/operator` | operator **or** admin password | Live control UI |

Each page shows its own **login form** (HTML form, not HTTP Basic auth, so
password managers work) and sets an HttpOnly, SameSite=Strict session cookie
(`Secure` only when served over HTTPS). The admin and operator passwords are
distinct and are read from `config.yaml`. An admin session is accepted
everywhere an operator session is; the reverse is not true.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/login` | `{role: "admin"\|"operator", password}` → sets cookie; rate-limited (5 attempts/min/IP) |
| POST | `/api/logout` | Clears the session |

### Setup API — admin session required

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/shows` | List shows |
| POST | `/api/admin/shows` | Create `{name}` |
| GET/PUT/DELETE | `/api/admin/shows/{id}` | Metadata, theme, delete |
| GET | `/api/admin/shows/{id}/script` | Raw text + parsed cues + diagnostics (`?lang=`) |
| PUT | `/api/admin/shows/{id}/script` | Body: raw script text (`?lang=`) → diagnostics |
| POST | `/api/admin/shows/{id}/script/validate` | Dry-run parse without saving |
| POST | `/api/admin/shows/{id}/assets` | multipart image upload: png/jpg/gif/webp only (SVG rejected); 10 MB limit; stored as-is (no server-side scaling); response includes a `large: true` flag above 1 MB |
| GET/DELETE | `/api/admin/shows/{id}/assets/{file}` | |
| POST | `/api/admin/shows/{id}/activate` | Make active (blocked if live unless `force`) |
| GET | `/api/admin/settings` | Effective config (passwords redacted) |
| GET | `/api/admin/network` | Host IPs, ports, audience URL(s); QR PNG at `/api/admin/network/qr.png` |
| GET | `/api/admin/support-bundle` | Zip of logs + config (passwords redacted) |

### Live control API — operator or admin session

| Method | Path | Body | Description |
|--------|------|------|-------------|
| GET | `/api/operator/state` | | `OperatorState` + cue list |
| GET | `/api/operator/stream` | | SSE of operator events (cursor, audience count, controller, loop) |
| POST | `/api/operator/claim` | `{takeover:bool}` | Become the controlling operator |
| POST | `/api/operator/next` | `{seq}` | NEXT: send the next cue. If the cue is a loop, starts it. If a loop is running, ends it and sends the next cue |
| POST | `/api/operator/back` | `{seq}` | Undo last NEXT |
| POST | `/api/operator/stop` | `{seq}` | STOP: stop a running loop and leave a blank screen |
| POST | `/api/operator/jump` | `{index, seq}` | Set cursor to a cue; rebuilds the resulting state without replaying cues |
| POST | `/api/operator/blank` | `{on:bool}` | Ad-hoc blank toggle |
| POST | `/api/operator/clear` | | Ad-hoc clear |
| POST | `/api/operator/message` | `{text}` | Ad-hoc one-off fullscreen/caption text |
| POST | `/api/operator/reset` | | Return to start (confirm in UI) |

Responses: `200` with the new `OperatorState`; `409` with current state if
`seq` is stale or the caller isn't the controller; `401` if not logged in.

## Audience event stream — `GET /events`

`Content-Type: text/event-stream`. Heartbeat comment `: ping` every 15 s; the
client reconnects if no bytes arrive for 40 s. Each event has `id: <seq>`,
`event: <type>`, `data: <json>`.

| Event | When | Data |
|-------|------|------|
| `snapshot` | On connect/reconnect; after jump/reset | full `LiveState` |
| `caption` | Caption appended | `{seq, item: CaptionItem}` |
| `clear` | Scrollback emptied | `{seq}` |
| `blank` | Blank on/off | `{seq, on}` |
| `fullscreen` | Fullscreen item set (including each step of a running loop) | `{seq, item: Item}` |
| `fullscreen_end` | Fullscreen dismissed | `{seq}` |
| `preload` | Upcoming image URLs changed | `{seq, urls: [..]}` |
| `theme` | Theme changed | `{seq, theme}` |
| `reload` | Clients should reload the page (server upgraded) | `{}` |

**Loops are server-driven.** The server runs the loop timer and emits one
`fullscreen` event per item, so every device and the projector stay in step
and STOP takes effect everywhere at once. Clients never run their own timers
for loops. If a client drops mid-loop, its reconnect `snapshot` carries the
current item.

**Image preloading.** Because the server doesn't scale images and ~150 devices
may request the same file at once, clients fetch upcoming images ahead of time.
The snapshot and `preload` events list the next few image URLs after the
cursor (including images in the next loop); the page fetches them in the
background at low priority so they're in cache before they're shown.

If the client reconnects with `Last-Event-ID` and the gap isn't covered by
the server's retained buffer, the server sends a `snapshot`. Clients treat a
`snapshot` as authoritative and replace local state.

### Schemas

```jsonc
// CaptionItem
{ "id": 1042, "speaker": "OSCAR", "html": "Felix, you're <strong>late</strong>.", "text": "Felix, you're late." }

// Item (fullscreen)
{ "kind": "text", "html": "<strong>INTERMISSION</strong>" }
{ "kind": "image", "src": "/assets/odd-couple/curtain.jpg", "caption": "Act 1" }
{ "kind": "blank" }

// LiveState
{
  "seq": 1042,
  "lang": "en",
  "languages": [{"code": "en", "name": "English"}, {"code": "es", "name": "Español"}],
  "mode": "captions",            // idle | captions | blank
  "scrollback": [ /* CaptionItem[] */ ],
  "fullscreen": null,            // or Item
  "preload": ["/assets/odd-couple/curtain.jpg"],
  "theme": { "bg": "#000000", "fg": "#ffffff", "accent": "#ffffff", "font": "system",
             "scalePct": 100, "speakers": "label", "idleMessage": "Captions will begin shortly" }
}

// OperatorState
{
  "seq": 1042,
  "cursor": 57,                  // index of the NEXT cue
  "total": 312,
  "controller": "session-id|null",
  "audience": 83,
  "loop": { "running": true, "cue": 3, "item": 1 },   // or null
  "cues": [ /* cue window: kind, text, line, sent, image name only for image cues */ ]
}
```

Operator payloads for image cues contain only `{kind:"image", name:"curtain.jpg"}`, never image bytes or URLs.

## Concurrency and ordering

- The engine owns `LiveState` behind one mutex/goroutine; each mutation (including loop timer ticks) yields exactly one `seq` increment and one event.
- The hub broadcasts in `seq` order; per-client buffered channel (size 32). Overflow ⇒ close that connection (client reconnects → snapshot).

## Versioning

`GET /api/version` → `{version, commit, protocol: 1}`. If a server upgrade
changes the protocol number, the server sends `reload`.
