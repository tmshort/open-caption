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
| GET | `/assets/{showId}/{file}` | Image assets of the active show (cacheable; audience URLs carry `?v=<version>` and change when an image is replaced) |
| GET | `/healthz` | `200 ok` |
| GET | `/standby-logo` | The venue's standby logo image (PNG, JPEG, GIF or WebP) if one was uploaded in Setup (`404` otherwise; the audience page then shows the built-in `[oc]` mark, `/static/shared/favicon.svg`). The address carries `?v=<modified time>` so a new logo is fetched at once |
| GET | `/favicon.ico` | The browser tab icon (browsers ask for it at the root) |
| GET | `/captive-api` | RFC 8908 captive-portal API: `application/captive+json` `{"captive": false, "user-portal-url": "<scheme>://<host>/"}`. Always reports not captive; clients require HTTPS with a trusted certificate (doc 06) |

Query parameters on `/` and `/events`: `?display=projector`, `?size=<pct>`,
`?theme=light|dark`, `?lang=<code>`. If `lang` is absent or not offered, the server picks one from the
`Accept-Language` request header, falling back to the show's main language (doc 02). The page's language menu just navigates to `/?lang=<code>`.

### Help — admin or operator session

| Method | Path | Description |
|--------|------|-------------|
| GET | `/follow` | The operator page for a second device that only watches (REQ-071): same page, no controls; same access list as `/operator` |
| GET | `/help` | Help viewer page (topic list + content); `?role=admin\|operator` picks the default topic set |
| GET | `/help/{topic}` | One topic as rendered HTML fragment (`?format=md` returns the Markdown source) |
| GET | `/help/sample-show.md` | Downloadable starter show file demonstrating every convention |
| GET | `/api/help/topics`, `/api/help/search?q=` | Topic list (with role) and full-text search |

Help is embedded in the binary (no network needed). It contains no secrets, but
is served only to logged-in sessions so the audience never sees operator tooling.

### Pages and authentication

| Path | Who | Notes |
|------|-----|-------|
| `/admin` | admin password | Setup UI |
| `/operator` | operator **or** admin password | Live control UI |

**Access and request checks.** Before authentication, every request to the Setup/operator pages, their JavaScript, APIs and login is checked against `access.admin_allow` / `access.operator_allow` (doc 02): `403` from an unlisted computer. A viewer-role session (viewer password only) is also refused with `403`, at sign-in and on every call, from a computer outside `access.viewer_allow` when that list is set. State-changing `/api/*` requests (POST/PUT/DELETE) are refused with `403` if `Origin` does not match `Host` or `Sec-Fetch-Site` is `cross-site`/`same-site`. `/static/shared/` (speaker colors, fonts) is public.

**Automatic sign-in on this computer (REQ-067).** With `access.local_auto_login` (default true), an admin/operator API call that has no valid session gets an **admin session and its cookie at once** when the request is `localTrusted`: the TCP peer is a loopback address, `Host` is `localhost`, `127.0.0.1` or `[::1]` (a name that merely resolves to 127.0.0.1, as in DNS rebinding, is refused), none of `X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Proto`, `Forwarded`, `X-Real-Ip` is present, and `Sec-Fetch-Site` is absent, `none` or `same-origin`. The access lists are checked first. `GET /api/admin/ping` and `GET /api/operator/state` report `role` and `auto` (true for such a session; pages hide Sign out). The login form and `/api/login` are unchanged for every other computer.

Each page shows its own **login form** (HTML form, not HTTP Basic auth, so
password managers work) and sets an HttpOnly, SameSite=Strict session cookie
(`Secure` only when served over HTTPS). The admin and operator passwords are
distinct and are read from `config.yaml`. An admin session is accepted
everywhere an operator session is; the reverse is not true.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/login` | `{role: "admin"\|"operator"\|"viewer", password}` → sets cookie; role `operator` accepts the operator or admin password, **role `viewer` (the `/follow` page, REQ-071) accepts the viewer, operator or admin password and the answer says which role the password gave**, role `admin` only the admin password; rate-limited (5 attempts/min/IP) |
| POST | `/api/logout` | Clears the session |
| GET | `/api/me` | `{role, sid}` of the signed-in session (operator or admin), where `sid` is the session's public identifier (a hash of the token, safe to send to browsers) that the operator page compares with the controller's to know whether it is in control (REQ-034). `401` without a session |

### Setup API — admin session required

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/shows` | List shows, sorted by name (case does not matter, then by id), so the list keeps its order whichever show is edited or on air |
| POST | `/api/admin/shows` | Create `{name}` (the title is written into the new show file's `[settings]`) or `{sample: true}` |
| GET/PUT/DELETE | `/api/admin/shows/{id}` | `GET`: `name` (the show file's title), `language`, `updated`, `lookSaved`, and `theme`, the **resolved** look (show file over config defaults; includes `loopSeconds`, `scrollbackLines` and per-language `idleMessages`). `PUT {name?, theme?}` rewrites the show file(s)' `[settings]` (a saved `theme` writes every look setting; each translation's `idle message` goes in its own file) and applies the change to the show on air. `DELETE` removes the show |
| POST | `/api/admin/shows/{id}/duplicate` | `{name?}` copy show files (every language), images and look |
| GET | `/api/admin/shows/{id}/showfile` | Raw text + parsed cues + diagnostics + `characters` (`?lang=`): every speaker name with the lines it speaks on, `[{name, lines, label?, similar?}]`, most lines first; a group label counts for each name, `similar` names a more used name one letter away (a likely misspelling) |
| PUT | `/api/admin/shows/{id}/showfile` | Body: raw show file text (`?lang=`) → diagnostics |
| GET | `/api/admin/shows/{id}/showfile/download` | The show file as a plain `<id>.md` (or `<id>.<lang>.md`) attachment, UTF-8 with a BOM and CRLF line ends so any Windows editor opens it correctly; for editing offline. It goes back through the editor's **Upload a show file** and the normal save |
| POST | `/api/admin/shows/{id}/archive` | Archive the show (`archived: yes` in its show file); 409 if it is the show on air. Returns the show summary (`archived: true`) |
| POST | `/api/admin/shows/{id}/unarchive` | Restore an archived show |
| GET | `/api/admin/shows/{id}/export` | Download the show as a `.zip` (format: doc 06, *Show archive*); built completely before it is sent, so a failure is a 500 `{error}` and never a truncated zip |
| POST | `/api/admin/shows/import` | multipart field `file`: a show `.zip` (creates a show titled `<title> (imported)`, with its images), **or a plain text show file** of any name (anything that does not start with the zip signature; UTF-8, UTF-16 with a mark, or Windows-1252; titled from `[settings]` `title:`, else from the file name). Returns `201 {show, skippedFiles}`; 400 if a zip has no `show.md`, or the file is empty or binary; 413 if the file is over 64 MB or a zip would unpack to more than 200 MB (or hold more than 500 entries) |
| GET | `/api/admin/shows/{id}/showfile/history` | Earlier versions of a show file (`?lang=`), newest first: `{versions: [{id, time, lines, bytes}]}` |
| GET | `/api/admin/shows/{id}/showfile/history/{version}` | `{text}` of one earlier version |
| PUT | `/api/admin/shows/{id}/showfile/history/{version}/label` | `{label}`: name an earlier version (1–80 characters, one line; empty removes the name). A named version is never pruned |
| POST | `/api/admin/shows/{id}/showfile/history` | `{label}`: keep a copy of the show file as saved now, under a name. Returns the version (`201`) |
| POST | `/api/admin/shows/{id}/showfile/validate` | Dry-run parse without saving |
| GET | `/api/admin/shows/{id}/languages` | The show's languages (main first) with cue/error/warning counts, `available` (false when a translation does not match the main show file) and the `structure` problems; plus the `known` language names offered in Setup |
| POST | `/api/admin/shows/{id}/languages` | `{code, name?, copyFromMain}`: start a translation (409 if it exists) |
| DELETE | `/api/admin/shows/{id}/languages/{lang}` | Remove a translation (not the main language; not while live) |
| POST | `/api/admin/shows/{id}/assets` | multipart image upload: png/jpg/gif/webp only (SVG rejected); 10 MB limit; stored as-is (no server-side scaling); response includes a `large: true` flag above 1 MB; 413 if the upload is over the limit, 400 if the body cannot be read as a form |
| GET/DELETE | `/api/admin/shows/{id}/assets/{file}` | |
| POST | `/api/admin/shows/{id}/activate` | `{force?, start?}`: load the show **in standby** (blocked if a show is on air unless `force`; 409 if the show is archived). `start: true` also puts it on air at once (automation and tests; Setup never does) |
| POST | `/api/admin/rehearsal/start` | Create an isolated rehearsal engine for this session (doc 02) |
| POST | `/api/admin/rehearsal/{next,back,stop,jump,blank,clear,reset}` | Same bodies/semantics as `/api/operator/*`, applied to the rehearsal engine only |
| GET | `/api/admin/rehearsal/events` | SSE of the rehearsal's audience-view events (drives the preview pane). `?lang=<code>` selects the rehearsal language, and `?display=projector` marks a projector client in the rehearsal clients list |
| GET | `/api/admin/settings` | Effective config (passwords redacted), settings file and data folder locations, listening ports, `adminAllow`/`operatorAllow` and whether they restrict anyone |
| PUT | `/api/admin/settings/passwords` | `{admin?, operator?, viewer?, removeViewer?}` (min 6 characters, different from each other; `removeViewer` turns viewer sign-in off; `GET /api/admin/settings` says whether one is set as `viewerPasswordSet`) written to `config.yaml`, effective at once; signs out others using a changed password |
| GET | `/api/admin/settings/operator-password` | The current operator password (for handing over); 409 when `config.yaml` holds only its digest (`GET /api/admin/settings` says so as `operatorPasswordHashed`) |
| GET | `/api/admin/network` | Host IPs, ports, audience URL(s); QR PNG at `/api/admin/network/qr.png?url=` (400 unless the address is one of this computer's: a numeric address, `localhost`, the host name, the `.local` name or a name in the HTTPS certificate). The warning that HTTPS uses a self-signed certificate is given only when neither `https.cert_file` nor `https.acme` supplies the certificate |
| PUT | `/api/admin/settings/mdns` | `{enabled, name}`: announce this computer as `<name>.local` (name: letters, digits, hyphens; a trailing `.local` is accepted). Applies at once (waits up to 6 s for the name to be claimed), is saved in `config.yaml` (`mdns.enabled`, `mdns.name`), and returns `{mdns: {enabled, requested, name, host, active, renamed, error}, saved}`. `GET /api/admin/settings` includes `mdns`; `GET /api/admin/network` lists the name first (`name: true`) |
| PUT | `/api/admin/settings/network` | `{interfaces: [...], restrictListening}`: the names of the network interfaces that are used (blank and repeated names are dropped; at most 100; 400 if none is given). `restrictListening: true` also makes the server close connections made to an address of an interface that is not used, except from this computer itself. 409 whenever the choice would turn off the interface this request arrived on (nothing is protected when Setup is used at `localhost`). Applies at once, is saved in `config.yaml` (`network.interfaces`, `network.restrict_listening`) and returns `{saved, interfaces, note?}` (`note` when the program was started with `--network-interfaces`, which wins at the next start). Each interface is `{name, kind, addresses, selected, used, instead, insteadIp, current}`: `kind` is `ethernet`, `wifi` or `other`; `used` says an address of it is offered; a ticked interface that is not used names the interface offered for the same network in `instead`/`insteadIp`; `current` marks the interface the request arrived on. `GET /api/admin/settings` includes `network: {interfaces, restrictListening}` (an empty list means no choice has been made: all are used), `networkPinned` and `interfaces` (every connected interface, the one with the main connection first); `GET /api/admin/network` lists only the offered addresses, each numeric one with `adapter` (the kind); `localhost` is listed for http only; with HTTPS on, the certificate's DNS names come first as `https` entries with `cert: true` (the ACME domain, or the non-wildcard names in `https.cert_file`; none for a self-signed certificate) |
| PUT | `/api/admin/settings/log-level` | `{level}`: debug, info, warn or error; applies at once and is written to `log.level` in config.yaml (edited in place). `GET /api/admin/settings` includes `logLevel` and `logLevelPinned` (the `--log-level` / `OPEN_CAPTION_LOG_LEVEL` value, if given, which overrides the file at each start; the PUT response then carries a `note` saying so) |
| PUT | `/api/admin/settings/standby` | `{message}`: the venue's standby message (one line, up to 200 characters; `""` for the built-in words in each viewer's language); written to `standby.message` in `config.yaml` and applied to the show on air and every rehearsal at once |
| POST | `/api/admin/settings/standby-logo` | `multipart/form-data` with a `file` (PNG, JPEG, GIF or WebP, up to 2 MB; SVG is refused) replaces the standby logo (a file `standby-logo.<ext>` in the data folder) |
| DELETE | `/api/admin/settings/standby-logo` | Removes the logo; the built-in `[oc]` mark is shown again |
| GET | `/api/admin/settings/certificate` | Where HTTPS gets its certificate: `source` is `none`, `self-signed`, `file` or `acme`; for `acme` also `status` (domain, state `waiting`/`ok`/`renewing`/`failed`, expiry, issuer, last attempt and last problem). Never contains credentials (doc 10) |
| POST | `/api/admin/settings/certificate/renew` | Get a new certificate now (it runs in the background); 409 if a fetch is already running or automatic certificates are off |
| GET | `/api/admin/log` | The recent in-memory log for the viewer in Setup: `?level=` (lowest severity: debug, info, warn, error; a line with no level counts as info) and `?lines=` (newest N, default 300, at most 1000). Returns `{lines, level}`. Admin only |
| GET | `/api/admin/troubleshoot` | What the Troubleshoot panel shows: `{checks: [{id, status: ok\|info\|warn\|bad, params}], problems, level}`. The checks are facts, not words (Setup holds the text, so it is translated): `address` (a network address to give out), `port` (the HTTP port really bound, and the one wanted) and `port-https` (the same for HTTPS, only while HTTPS is on), `https` (off, self-signed, certificate files, or the automatic certificate's state and days left), `name` (the `.local` name), `settings` (the settings file can be opened for writing), `show` (loaded, or languages held back), `devices`, `problems` (warnings and errors in the recent log) and `debug`. `problems` is how many warnings and errors have been logged since the program started (`logbuf.Buffer.Problems`), which Setup compares with the count it saw when the panel was last opened to badge the button; `?brief=1` returns only that count |
| GET | `/api/admin/support-bundle` | Zip: `info.txt`, `config.redacted.yaml`, `log.txt` (last 1000 lines), `shows.txt`. Passwords and DNS credentials are never included |

### Live control API — operator or admin session

| Method | Path | Body | Description |
|--------|------|------|-------------|
| GET | `/api/operator/state` | | `OperatorState` + cue list (also for a viewer session, whose `role` is `"viewer"`: every other `/api/operator/*` call and everything under `/api/admin/*` answers 403) |
| GET | `/api/operator/stream` | | SSE of operator events (cursor, audience count, controller, loop) |
| POST | `/api/operator/claim` | `{takeover:bool}` | Become the controlling operator |
| POST | `/api/operator/next` | `{rev}` | NEXT: send the next cue. If the cue is a loop, starts it. If a loop is running, ends it and sends the next cue |
| POST | `/api/operator/back` | `{rev}` | Undo the last operator action (NEXT, STOP, blank, clear, message). Backing into a loop restarts it from its first item |
| POST | `/api/operator/skip` | `{rev}` | SKIP: pass over the next line without sending it; the audience's screen does not change (no event, "last sent" unchanged) and Back undoes it. The operator state lists the cues passed over in `skipped` (kept across a restart). A running loop keeps running. `notice` "End of the show" when no line is left; 409 in standby |
| POST | `/api/operator/stop` | `{rev}` | STOP: stop a running loop and leave a blank screen |
| POST | `/api/operator/jump` | `{index, rev}` | Set cursor to a cue; rebuilds the resulting state without replaying cues |
| POST | `/api/operator/blank` | `{on:bool}` | Ad-hoc blank toggle |
| POST | `/api/operator/clear` | | Ad-hoc clear |
| POST | `/api/operator/message` | `{text}` | Ad-hoc one-off fullscreen/caption text |
| POST | `/api/operator/reset` | | Return to the start **and to standby** (confirm in UI) |
| GET | `/api/operator/clients` | | Who is connected to the audience page: `{clients: [{ip, kind, os, browser, lang, langName, since, role?}], total}` (a page that only watches, `/follow`, is a row with `role: "viewer"` after the audience's; `total` leaves it out. The page opens its stream as `/api/operator/stream?follow=1` to be listed; a viewer session is always listed). `kind` is `phone`, `tablet`, `computer`, `projector` (opened with `?display=projector`) or `other`, guessed from the browser's `User-Agent`; `os` and `browser` are brand names (`iPhone`, `Safari`). Setup uses it too; a rehearsal has its own at `/api/admin/rehearsal/clients` |
| GET | `/operator/report` | | The audience report page (REQ-070; operator or admin). |
| GET | `/api/operator/report` | `?id=` | `{id, report, languageNames}`: the report of the run in progress (`id: "current"`, `report.running: true`) or, when there is none, the newest saved one; `?id=<file>` (a name from `/api/operator/reports`) opens a saved one. `report` is `{showId, show, started, ended, running, totals, languages, devices}`; `totals` has `devices, phones, tablets, computers, projectors, peak, peakAt, now, connections, reconnects, avgConnected, longest` (the durations in whole seconds); `series` is `[{t, n}]`: how many devices were connected from `t` until the next point (the last point is the end of the run, or now) and `loops` is `[{start, end, kind, name, cue}]`: when a loop was running (`name` is what the main show file calls it, when it does; `kind` is `preshow` before the first line the audience sees, `postshow` after the last, `intermission` between, `loop` in a show of nothing else; `cue` its place in the show file; the running loop is read every 2 seconds, so a loop's edges are accurate to about that); `languages` maps a language code to the number of devices that used it; each device is `{id, address, kind, projector, languages, first, last, connections, connected, longest, now}` (`id` is a short code made from the address and browser type). 404 when no report exists yet or the `id` is not a saved report |
| GET | `/api/operator/reports` | | `{reports: [{id, show, started, ended, running}]}`: the run in progress (`id: "current"`) first, then the saved ones, newest first (the newest 30 are kept in `<data>/reports/`). A reset or re-activation ends the previous run even when the same show ID immediately goes on air again; stored IDs include nanoseconds so rapid runs remain distinct |
| GET | `/api/operator/report.csv` | `?id=` | The same report as a CSV file, one row for each device (UTF-8 with a byte-order mark so a spreadsheet reads accents; the Screen column says `device`, `projector` or `viewer`: a viewer row is a page that only watches, which is in no total). The download name uses a seconds-resolution start time for readability |
| DELETE | `/api/admin/reports` | `?id=` | Admin only. Deletes a saved report (`id` is a name from `/api/operator/reports`): `{ok: true}`. 404 when there is no such saved report (a path is never accepted); 409 for the run in progress (`current`, or the file it is being saved to), which is saved as it goes |
| POST | `/api/operator/onair` | `{rev}` | START: put the show on air (from standby: the pre-show loop begins, or the show's idle message appears). A no-op when already on air |
| POST | `/api/operator/offair` | `{rev}` | Go to standby: the audience sees the standby screen again; the position, lines and undo history are kept; START resumes. A no-op in standby |

Request bodies are JSON; `rev` is optional (omitted = no staleness check). `jump` takes `index`, `blank` takes `on`, `message` takes `text` (1–500 chars).

Operator state is pushed on `/api/operator/stream` as `opstate` events carrying `{op: OperatorState, audience, controller (public session id), cuesVer, show}`; when `cuesVer` changes the page re-fetches `/api/operator/state` for the cue list. Operators see image **names only** (`CueView`), never image URLs.

`claim` returns `409 {ok:false}` if another live operator holds control and `takeover` was not set. A controller whose stream is closed keeps control for 30 s so a reload doesn't lose it.

Responses: `200` with the new `OperatorState` and, for harmless no-ops, a `notice` ("End of the show", "Nothing to undo", "No loop is running"); `409` with current state if
`rev` is stale or the caller isn't the controller; `401` if not logged in.

### Development endpoints (`--dev` only)

With `open-caption --dev` (never at a venue: the program then listens on this computer only, has no standby and no passwords) there are unauthenticated `POST /dev/next`, `/dev/back`, `/dev/stop` and `/dev/reset`, answering `{state, error}`. They exist for tests and for the load generator (`cmd/loadtest`).

## Audience event stream — `GET /events`

`Content-Type: text/event-stream`. A `ping` event (`event: ping`, `data: {}`) every 15 s; it is a real
event rather than a comment because `EventSource` can't observe comments. The client reconnects if it sees no event for 40 s. Each event has `id: <seq>`,
`event: <type>`, `data: <json>`.

| Event | When | Data |
|-------|------|------|
| `snapshot` | On connect/reconnect; after jump/reset | full `LiveState` |
| `caption` | Caption appended. Also dismisses any fullscreen item/loop and un-blanks | `{seq, item: CaptionItem}` |
| `clear` | Scrollback emptied | `{seq}` |
| `blank` | Blank on/off. `on` also dismisses any fullscreen item/loop | `{seq, on}` |
| `fullscreen` | Fullscreen item set (including each step of a running loop) | `{seq, item: Item}` |
| `preload` | Upcoming image URLs changed. A hint with **no `id`/`seq`**; always delivered, may arrive between sequenced events | `{urls: [..]}` |
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
When an image referenced by the active show is uploaded or replaced, its versioned URL changes. The server sends a new `preload` hint if the image is upcoming, and a `snapshot` to viewers of a language showing it now. Uploading an unrelated image does not rebuild audience screens. Setup refuses to delete an image still referenced by the show on air (`409`); in standby, deletion is allowed. A rehearsal keeps the image URLs assigned when it starts, so an image replaced during a rehearsal may remain cached for up to 60 seconds; starting a new rehearsal assigns fresh URLs.

If the client reconnects with `Last-Event-ID` and the gap isn't covered by
the server's retained buffer, the server sends a `snapshot`. Clients treat a
`snapshot` as authoritative and replace local state.

### Schemas

```jsonc
// CaptionItem
{ "id": 1042, "speaker": "OSCAR", "html": "Felix, you're <strong>late</strong>.", "text": "Felix, you're late." }

// Item (fullscreen)
{ "kind": "text", "html": "<strong>INTERMISSION</strong>" }
{ "kind": "image", "src": "/assets/odd-couple/curtain.jpg?v=abc-1", "caption": "Act 1" }
{ "kind": "blank" }
// a text item of a loop that has a timer: HTML holds <span data-timer> where {timer} was written, and
// deadline is when the countdown ends (unix ms, the server's clock); the page fills the span in every second
{ "kind": "text", "html": "Back in <span class=\"oc-timer\" data-timer></span>", "deadline": 1760650800000 }

// LiveState
{
  "seq": 1042,
  "ts": 1760649900000,           // the server's clock when this snapshot was made; a page uses ts - Date.now() to correct its own clock for deadlines
  "lang": "en",
  "languages": [{"code": "en", "name": "English"}, {"code": "es", "name": "Español"}],
  "mode": "captions",            // standby | idle | captions | blank; "standby": nothing of the show is sent (no scrollback, no fullscreen) and `standby` below is set
  "standby": { "message": "Welcome", "logo": "/standby-logo?v=1760649900" },   // only while mode is "standby"; both optional
  "scrollback": [ /* CaptionItem[] */ ],
  "scrollbackMax": 200,          // the show's scrollback lines setting: a page keeps no more lines than this
  "fullscreen": null,            // or Item
  "preload": ["/assets/odd-couple/curtain.jpg?v=abc-1"],
  "theme": { "bg": "#000000", "fg": "#ffffff", "accent": "#ffffff", "font": "system",  /* system | serif | mono | dyslexic | hyperlegible */
             "scalePct": 100, "speakers": "label", "idleMessage": "Captions will begin shortly" }
}

// OperatorState
{
  "seq": 1042,
  "rev": 311,                  // operator-command revision; send this with commands
  "cursor": 57,                  // index of the NEXT cue
  "total": 312,
  "controller": "session-id|null",
  "audience": 83,
  "loop": { "running": true, "cue": 3, "item": 1, "deadline": 1760650800000 },   // or null; deadline only when the loop has a timer
  "cues": [ /* cue window: kind, text, line, sent, hint (seconds, from [pace]), image name only for image cues */ ]
}
```

Operator payloads for image cues contain only `{kind:"image", name:"curtain.jpg"}`, never image bytes or URLs.

## Concurrency and ordering

- The engine owns `LiveState` behind one mutex; every emitted event (including each loop timer tick) increments `seq` by exactly one. A single operator action may emit more than one event.
- The hub broadcasts in `seq` order; per-client buffered channel (size 32). Overflow ⇒ close that connection (client reconnects → snapshot). Beyond `audience.max_clients`, `/events` returns `503`.

## Versioning

`GET /api/version` → `{version, commit, protocol: 1}`. If a server upgrade
changes the protocol number, the server sends `reload`.
