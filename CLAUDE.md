# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A single-binary open-captioning server for community theater (Go 1.22+, no CGO). One host serves three web UIs: `/admin` (Setup wizard), `/operator` (spacebar advances the caption), and the audience page. Module path: `github.com/tmshort/open-caption`.

## Commands

- `make build` → `bin/open-caption`; run `./bin/open-caption` (first run writes `config.yaml` with random passwords; `--show-file` flag exists).
- `make test` / `make race` (`go test -race ./...`) / `make vet`
- Single test: `go test ./internal/httpapi -run TestName -count=1`
- `make fuzz` fuzzes the show-file parser (`FuzzParse`)
- `make e2e` — browser end-to-end suite (`test/e2e/run.js`, puppeteer-core; needs Node and Chrome, set `CHROME` if not at the default path). Chrome allows 6 connections per host, so e2e pages must be closed when finished (SSE streams stay open).
- Log with `internal/logx` (`logx.Infof/Warnf/Errorf/Debugf`), not `log.Printf`; level comes from `log.level` / `--log-level`.
- `config.Load` is strict (unknown YAML keys are errors) and reports all problems at once via `*config.ProblemsError`; add new settings to `normalize()`.
- `tools/make-icons.js` regenerates `favicon.ico` / `apple-touch-icon.png` from `web/shared/favicon.svg`.
- `make dist` cross-compiles release binaries (CGO disabled).

Run `gofmt`, `go vet ./...` and `go test -race ./...` before committing.

## Spec-driven

`docs/` is the source of truth (see README table); change the spec in the same commit as behavior. Requirements are numbered `REQ-xxx`; reference them from tests and commits. Terminology: the text file a show is written in is the **show file** — never "script".

## Architecture (needs several files to see)

- **`internal/showfile`**: parser for the Markdown-subset show file (one line = one caption; directives like `[clear] [image] [loop] [song] [together] [language] [settings]`). Produces `Cue`s plus diagnostics with line numbers; `CompareStructure` checks a translation matches the main language cue-for-cue. `settings.go` parses and rewrites the `[settings]` block (`UpdateSettings` / `ReplaceSettings` keep everything else in the file byte-for-byte).
- **`internal/show`**: on-disk store, `<data-dir>/shows/<id>/` holding `show.md` (main language), `show.<lang>.md` (translations) and `assets/`. There is no separate metadata file: the title, look, idle message(s), loop time and scrollback size are all in the show file's `[settings]`.
- **Configuration layering**: built-in defaults (`config.Default`) < `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) < show file `[settings]`. `show.Defaults.Resolve` does the layering; `httpapi.showDefaults(cfg)` builds the defaults. Environment-only settings (ports, HTTPS, passwords, access lists, data dir) live in `config.yaml` only. Look/Rename in Setup write the show file; a translation's file may set only `idle message`.
- **`internal/live`**: the per-language caption `Engine` (cursor, scrollback, loops, undo, blank/clear). Pure state machine; emits events via a callback. `Engine.Configure` changes loop time/scrollback on a running engine.
- **`internal/hub`**: fan-out of SSE events to audience/operator connections for one language. Clients subscribe first, then receive a snapshot (ordering matters for not missing events); `seq` orders audience events, `rev` orders operator commands.
- **`internal/httpapi`**: `Server` (routes, auth, Origin/Sec-Fetch-Site guard, access allow-lists), `Runtime` (the active show: one engine + hub per language, driven in lockstep by `Runtime.Do`; persists to `live.json`), and `admin.go`/`operator.go`/`audience.go` handlers. `Server.applyActive(id)` is how any change to the show on air (look, title, loop time, every language's lines) reaches the runtime — call it after saving show files or settings.
- **`internal/config`**: `config.yaml` load/validation and access lists. **`internal/help`**: embedded in-app help (a test checks every show-file convention and setting name is documented in `web/help/*.md`).
- **`web/`** is embedded with `go:embed`: `admin/`, `operator/`, `audience/` (plain JS, no build step), `help/`, `sample/` (sample show, English + Spanish), fonts. Nothing is fetched from the internet at runtime.

## Things to remember

- Show-file edits while live require confirmation (typing the show name); the cursor is kept by cue index. Saving the look is refused when the editor has unsaved changes (it rewrites the saved file).
- Archiving a show is the `archived: yes` setting in its show file (`Store.SetArchived`); archived shows cannot be activated.
- The audience page keeps the screen on with the Wake Lock API, or a tiny looping video (`web/audience/nosleep.*`) where that API is missing (plain HTTP).
- Browser floor: `web/shared/compat.js` (plain ES5, loaded first on every page) gates old browsers and fills in `<dialog>`/`replaceChildren`; `test/e2e/check-compat.js` keeps JS at ES2018 and CSS free of newer features (the audience page also avoids `inset` and flex `gap`). Page CSPs forbid `eval`/`new Function`.
- `internal/mdns` announces `<name>.local` (own minimal responder over `golang.org/x/net`, IPv4 only, probes for name conflicts); `httpapi/mdns.go` runs it and Setup's Settings dialog changes the name. The real network is only touched from `main`, never in tests (`fakeResponder`, the in-memory `bus`).
- UI language (Setup, operator, Help): wrap every on-screen English literal in `__('…')` (`{name}` placeholders: `__('Hi {name}', {name})`); `web/shared/i18n/<es|fr|de|it|pt>.js` map English → translation. `node test/e2e/check-i18n.js --list` shows what needs translating (it also reads the Go error/diagnostic strings and `i18n-tables.json`/`i18n-extra.json` for strings reached through variables); the e2e run fails if any language lacks one.
- Sessions persist hashed in `sessions.json`; per-language live state in `live.json`.
- Multi-language: one operator keypress advances every language, so structure mismatches are errors and the mismatching translation is not served.
