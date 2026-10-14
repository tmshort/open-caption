# 04 — Protocol and API

Base URL: `http://<host>:<port>/` (default port **80** if permitted, else **8080**;
see doc 06). All JSON is UTF-8.

## Routes

### Audience (public)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Audience page (single HTML shell + JS) |
| GET | `/events` | SSE stream of live state (see below) |
| GET | `/assets/{showId}/{file}` | Image assets used by the active show (cacheable) |
| GET | `/healthz` | `200 ok` |

Query parameters on `/`: `?display=projector` (projector mode), `?size=<pct>` (initial font scale), `?theme=light|dark`.

### Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/login` | `{pin}` → sets session cookie |
| POST | `/api/logout` | |
| POST | `/api/firstrun` | `{pin}` only valid before a PIN exists |

### Setup (auth required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/shows` | List shows |
| POST | `/api/shows` | Create `{name}` |
| GET/PUT/DELETE | `/api/shows/{id}` | Metadata, theme, delete |
| PUT | `/api/shows/{id}/script` | Body: raw script text → returns `ParseResult` diagnostics |
| POST | `/api/shows/{id}/script/validate` | Dry-run parse of text without saving |
| GET | `/api/shows/{id}/script` | Raw text + parsed cues + diagnostics |
| POST | `/api/shows/{id}/assets` | multipart image upload (png/jpg/gif/webp/svg, size limit 10 MB, server downscales > 2560 px optional) |
| GET/DELETE | `/api/shows/{id}/assets/{file}` | |
| POST | `/api/shows/{id}/activate` | Make active (blocked if live and not `force`) |
| GET | `/api/network` | Host IPs, port, audience URL(s), QR PNG at `/api/network/qr.png` |
| GET | `/api/support-bundle` | Zip of logs + config (PIN hash redacted) |

### Live control (auth required)

| Method | Path | Body | Description |
|--------|------|------|-------------|
| GET | `/api/live` | | Full `OperatorState` + `LiveState` + cue list window |
| GET | `/api/live/stream` | | SSE with operator-level events (cursor, audience count, controller) |
| POST | `/api/live/claim` | `{takeover:bool}` | Become the controlling operator |
| POST | `/api/live/advance` | `{seq}` | Send next cue |
| POST | `/api/live/back` | `{seq}` | Undo last advance |
| POST | `/api/live/jump` | `{index, seq}` | Set cursor to cue index; rebuilds state *without* sending intermediate cues individually |
| POST | `/api/live/blank` | `{on:bool}` | Ad-hoc blank toggle (not a script cue) |
| POST | `/api/live/clear` | | Ad-hoc clear |
| POST | `/api/live/message` | `{text}` | Ad-hoc one-off caption/fullscreen text (e.g. "Please wait—technical difficulty") |
| POST | `/api/live/reset` | | Return to start/pre-show (confirm in UI) |

Responses: `200` with new `OperatorState`; `409` with current state if `seq`
is stale or caller isn't the controller; `403` if unauthenticated.
Commands are rejected from non-controller sessions.

## Audience event stream — `GET /events`

`Content-Type: text/event-stream`. Heartbeat comment `: ping` every 15 s
(keeps proxies/WiFi NAT alive and lets clients detect stalls; client
reconnects if no bytes for 40 s).

Each event: `id: <seq>`, `event: <type>`, `data: <json>`.

| Event | When | Data |
|-------|------|------|
| `snapshot` | On connect and on reconnect; also after jumps/resets | full `LiveState` |
| `caption` | Cue appended | `{seq, item: CaptionItem}` |
| `clear` | Scrollback emptied | `{seq}` |
| `blank` | Blank on/off | `{seq, on}` |
| `fullscreen` | Fullscreen content set | `{seq, content: {kind, text|html|src|images[], seconds}}` |
| `fullscreen_end` | Fullscreen dismissed | `{seq}` |
| `theme` | Theme changed | `{seq, theme}` |
| `reload` | Server wants clients to reload page (new version) | `{}` |

If the client sends `Last-Event-ID` and the gap is not contiguous with the
server's retained buffer, the server simply sends a `snapshot`. Clients treat
a `snapshot` as authoritative and replace local state.

### Schemas

```jsonc
// CaptionItem
{ "id": 1042, "speaker": "OSCAR", "html": "Felix, you're <b>late</b>.", "text": "Felix, you're late." }

// LiveState
{
  "seq": 1042,
  "mode": "captions",              // preshow | captions | blank
  "scrollback": [ /* CaptionItem[] */ ],
  "fullscreen": null,              // or {kind:"text"|"html"|"image"|"slideshow", ...}
  "theme": { "bg": "#000", "fg": "#fff", "accent": "#ffd400", "font": "system", "scalePct": 100, "showSpeakers": true }
}

// OperatorState
{
  "seq": 1042,
  "cursor": 57,                    // index of NEXT cue to send
  "total": 312,
  "controller": "operator-session-id|null",
  "audience": 83,
  "cues": [ /* window of cues: kind, text, line, sent:bool */ ]
}
```

Operator display of image cues includes only `{kind:"image", name:"curtain.jpg"}`; never image bytes.

## Concurrency and ordering

- Engine has a single goroutine/mutex owning `LiveState`; every mutation yields exactly one `seq` increment and one event.
- Hub broadcasts events in `seq` order; per-client buffered channel (size 32). Overflow ⇒ close that connection (client reconnects → snapshot).

## Versioning

`GET /api/version` → `{version, commit, protocol: 1}`. Audience page includes
protocol number; if a server upgrade changes it, server sends `reload`.
