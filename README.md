# open-caption

A single-binary open-captioning server for community theater, written in Go.

An operator follows the live show and presses a key to release the next caption
line. Every audience device on the guest WiFi (phone, tablet, desktop, or a
browser on a projector) shows the captions live, with no interaction required.

## Features

**For the audience**
- Nothing to install: open one address (or scan a QR code) on any phone, tablet or computer on the venue's WiFi. `http://open-caption.local/` works too where the network allows it.
- Large, high-contrast captions by default; per-device text size, dark-on-light, and dyslexia-friendly (OpenDyslexic) and high-legibility (Atkinson Hyperlegible) fonts.
- Scroll back to re-read a line without losing the live captions; the screen stays awake during the show; reconnects by itself after WiFi drops or sleep.
- Several languages in one show, chosen from the device's own language; works on a projector or shared screen too (`?display=projector`).
- Runs on older phones (about iOS 12 and up) and tells people clearly when a browser is too old.

**For the operator**
- One key does it all: Space sends the next line, Backspace undoes it, B blanks, C clears; jump to any line or section, search, and send a one-off message.
- Operator notes and bookmarks the audience never sees; loops for pre-show and intermission; songs, lyrics and two voices shown at once for musicals.
- Survives a restart: the position, sign-ins and last few undo steps are kept.

**For whoever sets it up**
- A guided Setup page: show file editor with problems listed by line, image upload, look, a private rehearsal, then Go Live with a printable QR sign.
- A show is one plain-text Markdown file (plus images): title, colours, fonts and options live in it, layered over `config.yaml` defaults. Earlier versions are kept, can be named and compared, and shows can be exported, imported, duplicated and archived.
- Setup, the operator page, Help and all the built-in documentation are available in English, Spanish, French, German, Italian and Portuguese.
- Built-in help, a recent-log viewer, a support bundle, and a validated `config.yaml` that reports every mistake at once.
- Screen-reader friendly pages, checked automatically with axe-core in both colour schemes.

**Under the hood**
- One Go binary for macOS, Windows and Linux; no database, no internet needed, nothing fetched at runtime.
- Tested with 200 simultaneous audience devices (latency well under 250 ms), a real-browser end-to-end test on all three systems, and ready-made files to start it at boot (systemd, launchd, Task Scheduler).
- Passwords, allowed computers (`access` lists) and a cross-site request guard protect Setup and the operator page; it is meant for a private network.

## Quick start

```
make build
./bin/open-caption
```

The first run creates `config.yaml` (in your user config folder; the path is printed) with random
**admin** and **operator** passwords, printed once in the console. Then:

- **Setup:** open `http://localhost/admin` (or the `:8080` address printed if port 80 was busy), sign in with the admin password, and follow the six steps. "Create a sample show" is a good first try.
- **Operator:** `/operator` (the operator password; the admin password also works). Space sends the next line.
- **Audience:** the address shown in Setup → Go Live, or its QR code. Devices must be on the same WiFi.
- **Help:** the **Help** button on Setup and Operator. Read [docs/usage-guide.md](docs/usage-guide.md) (deployment, checklist) and [docs/venue-network-guide.md](docs/venue-network-guide.md) (the WiFi side).

`make test` runs the tests; `make e2e` runs the browser end-to-end test (needs Node and Chrome); `make dist` builds binaries for macOS, Windows and Linux.

**Security note:** anyone who can reach the server can open the Setup and operator sign-in pages, and over plain `http://` passwords cross the WiFi unencrypted. Limit them with `access.admin_allow` / `access.operator_allow` in `config.yaml` (see [docs/usage-guide.md](docs/usage-guide.md#security)).

## Spec-driven workflow

The documents in `docs/` are the source of truth. Code follows the specs; if
behavior needs to change, change the spec first (in the same PR).

| Doc | Purpose |
|-----|---------|
| [docs/usage-guide.md](docs/usage-guide.md) | **Start here for using it:** deployment recommendations, setup, running a show, checklist |
| [docs/venue-network-guide.md](docs/venue-network-guide.md) | For whoever manages the venue's WiFi |
| [docs/01-vision-and-requirements.md](docs/01-vision-and-requirements.md) | Goals, users, scope, numbered requirements (REQ-xxx) |
| [docs/02-architecture.md](docs/02-architecture.md) | Components, state model, transport, persistence, package layout |
| [docs/03-show-file-format.md](docs/03-show-file-format.md) | The show file syntax and its parsed model |
| [docs/04-protocol-and-api.md](docs/04-protocol-and-api.md) | HTTP endpoints, event stream, message schemas |
| [docs/05-ui-specs.md](docs/05-ui-specs.md) | Setup, Operator, and Audience interface specs |
| [docs/06-deployment-and-network.md](docs/06-deployment-and-network.md) | Running on macOS/Windows/Linux, WiFi, DNS/DHCP, QR code |
| [docs/07-testing-and-acceptance.md](docs/07-testing-and-acceptance.md) | Test strategy, load target, acceptance checklist |
| [docs/08-roadmap.md](docs/08-roadmap.md) | Milestones and task breakdown |
| [docs/09-open-questions.md](docs/09-open-questions.md) | Resolved decisions and remaining open questions |

Requirement IDs (`REQ-…`) are referenced from the other docs and should be
referenced from tests and commits.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
