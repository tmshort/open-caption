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
- One key does it all: Space sends the next line, Backspace undoes it, B blanks, C clears; jump to any line or section, search, and send a one-off announcement.
- A loaded show waits in **standby**: the audience sees your venue's own screen (a logo and a message you set once) until the operator presses START; **Go to standby** returns to it at any time without losing the place.
- Operator notes and bookmarks the audience never sees; loops for pre-show and intermission; songs, lyrics and two voices shown at once for musicals.
- Survives a restart: the position, sign-ins and last few undo steps are kept.

**For whoever sets it up**
- A guided Setup page: show file editor with problems listed by line, image upload, look, a private rehearsal, then a Run step with the audience address, a printable QR sign and the way into the operator page.
- Edit the show file in any program you like: download it as a plain `.md`, edit it offline, and open it again to save over the show (or import any `.md` / `.txt` file as a new show). A guide for the person preparing the show file is [Preparing a show file](web/help/admin-preparing-a-show-file.md); [Getting started](web/help/admin-getting-started.md) walks through the six steps of Setup. Both are in the app's Help too. `open-caption --check-show show.md` checks a show file from the command line, like a lint, without starting anything.
- A show is one plain-text Markdown file (plus images): title, colours, fonts and options live in it, layered over `config.yaml` defaults. Earlier versions are kept, can be named and compared, and shows can be exported, imported, duplicated and archived.
- Setup, the operator page, Help and all the built-in documentation are available in English, Spanish, French, German, Italian and Portuguese.
- Built-in help, a recent-log viewer, a support bundle, and a validated `config.yaml` that reports every mistake at once.
- Optional trusted HTTPS: a renewed certificate file is picked up without a restart, or the program can get and renew a Let's Encrypt certificate itself (DNS-01, with Cloudflare, GoDaddy and EasyDNS built in by default; see below).
- Screen-reader friendly pages, checked automatically with axe-core in both colour schemes.

**Under the hood**
- One Go binary for macOS, Windows and Linux; no database, no internet needed, nothing fetched at runtime.
- Tested with 200 simultaneous audience devices (latency well under 250 ms), a real-browser end-to-end test on all three systems, and ready-made files to start it at boot (systemd, launchd, Task Scheduler).
- On the computer that runs it, Setup and the operator page open without a password (turn off with `access.local_auto_login: false`); every other computer signs in. An admin session on the operator page has a **Setup** link back.
- Passwords, allowed computers (`access` lists) and a cross-site request guard protect Setup and the operator page; it is meant for a private network.

## Quick start

Two steps: get the program, then run it. Pick your computer's row; the guides start from a brand-new computer with nothing installed and list every download.

**1. Get the program.**

- **Download a release (easiest).** On this repository's *Releases* page, download the file for your computer, unzip (Windows) or untar (macOS, Linux) it, and open a terminal in the unpacked folder. The program is `open-caption` (`open-caption.exe` on Windows). **Intel and AMD computers are `amd64`; Apple silicon (M1 and later) and ARM are `arm64`.** The programs are not code-signed, so macOS and Windows warn the first time:
  - **macOS:** if it says the program cannot be opened or is from an unidentified developer, run `xattr -dr com.apple.quarantine .` in the unpacked folder, or try to open it once, then open **System Settings → Privacy & Security** and click **Open Anyway**.
  - **Windows:** if SmartScreen says "Windows protected your PC", click **More info → Run anyway**. If the zip file itself is blocked, right-click it → Properties → **Unblock** before unzipping.
  - **Linux:** nothing to do.
<!-- source-only:start -->
- **Build it from the source in this repository.** It needs Go and `git`, plus `make` except on Windows; a few minutes. The guides start from a brand-new computer with nothing installed and list every download:

  | Your computer | Full step-by-step guide | The build, once the tools are installed |
  |---|---|---|
  | **macOS** | [docs/build-macos.md](docs/build-macos.md) | `make build` |
  | **Windows 10 or 11** | [docs/build-windows.md](docs/build-windows.md) | `.\build.bat` |
  | **Linux** (Ubuntu, Fedora) | [docs/build-linux.md](docs/build-linux.md) | `make build` |

  Each build creates the program in the `bin` folder of the source (`bin/open-caption`, or `bin\open-caption.exe` on Windows).
<!-- source-only:end -->

**2. Run it**, from the folder it is in (a download: the unpacked folder; a build: the `bin` folder is where it is, so use the commands below from the source folder). To start it later, `cd` back there:

| Your computer | Command |
|---|---|
| macOS, Linux | `./open-caption` (a build: `./bin/open-caption`) |
| Windows (PowerShell) | `.\open-caption.exe` (a build: `.\bin\open-caption.exe`) |

The first run creates `config.yaml` (in your user config folder; the path is printed) with random **admin** and **operator** passwords, printed once in the console, and opens Setup in your browser. When the system asks whether to allow network connections, allow them (on Windows, tick *Private networks*), or phones cannot reach the program. Then:

- **Setup:** open `http://localhost/admin` (or the `:8080` address printed if port 80 was busy; on Linux a normal user cannot use port 80, so it uses 8080 unless the program is given permission, see the Linux guide). On the computer that runs the program it needs no password; elsewhere sign in with the admin password. Follow the six steps; "Create a sample show" is a good first try.
- **Operator:** the Run page, opened from Setup's Run step (or `/operator`, with the operator password; the admin password also works). Space sends the next line.
- **Audience:** the address shown in Setup → Run, or its QR code. Devices must be on the same WiFi.
- **Help:** the **Help** button on Setup and Operator. Read [Getting started](web/help/admin-getting-started.md) and [Preparing a show file](web/help/admin-preparing-a-show-file.md) (for whoever writes the show file), [docs/usage-guide.md](docs/usage-guide.md) (deployment, checklist) and [docs/venue-network-guide.md](docs/venue-network-guide.md) (the WiFi side).

To stop the program, press `Ctrl`+`C` in its window. Where the settings and shows are kept, and how to back them up: see [docs/usage-guide.md](docs/usage-guide.md) (*Choose where data lives*).

`open-caption --help` lists the command-line options; the manual page is `docs/open-caption.1` (`man ./docs/open-caption.1` on macOS and Linux).
<!-- source-only:start -->
`make help` lists every build and test target (Windows: the same jobs are described in [docs/build-windows.md](docs/build-windows.md)).
<!-- source-only:end -->

## Documentation

For people who use and run the program (the specifications are further down):

| Guide | For |
|-------|-----|
| [docs/usage-guide.md](docs/usage-guide.md) | **Start here:** deployment recommendations, setting up a show, running it, the day-of checklist, troubleshooting |
| [docs/venue-network-guide.md](docs/venue-network-guide.md) | Whoever manages the venue's WiFi: how audience phones reach the computer |
| [docs/https-certificate-guide.md](docs/https-certificate-guide.md) | Step by step: a trusted HTTPS certificate with Cloudflare, GoDaddy or EasyDNS |
| [docs/configuration-reference.md](docs/configuration-reference.md) | Every `config.yaml` setting, the command-line options and environment variables, checking a show file |
<!-- source-only:start -->
| [docs/build-macos.md](docs/build-macos.md), [docs/build-windows.md](docs/build-windows.md), [docs/build-linux.md](docs/build-linux.md) | Starting from a brand-new computer: exactly what to install, then how to build and run the program (English) |
<!-- source-only:end -->
| [docs/open-caption.1](docs/open-caption.1) | The manual page (`man ./docs/open-caption.1`): options, environment, files, signals, exit status (English) |

The usage guide, the venue network guide, the HTTPS certificate guide and the configuration reference are also available in Spanish, French, German,
Italian and Portuguese (`docs/i18n/`), and everything in the app's **Help** button (written for the person writing the show file and for the
operator) is translated the same way; the topics are in [web/help/](web/help/).

<!-- source-only:start -->
## Building and testing

**To run the program you need nothing but the binary.** To build and test it (starting from a bare computer? Follow the step-by-step guide for [macOS](docs/build-macos.md), [Windows](docs/build-windows.md) or [Linux](docs/build-linux.md)):

| You need | For | Notes |
|---|---|---|
| **Go 1.26 or later** | `make build` (on Windows, which has no make: `build.bat`), `make test`, `make race`, `make vet`, `make dist` | No CGO and no C compiler; `make dist` cross-compiles for macOS, Windows and Linux. |
| **Node.js 22.12 or later** (CI uses 22) | `make e2e`, `make a11y`, `make i18n` | The first run downloads the test tools with `npm` (listed in `test/e2e/package.json`); after that the tests need no internet. They are `puppeteer-core` (drives Chrome), `axe-core` (the accessibility checker) and `acorn` (a JavaScript parser, used to check that the pages stay inside the old-browser floor). |
| **Google Chrome or Chromium** | `make e2e`, `make a11y` | The tests drive a real browser. They look for Chrome in its macOS location; on Linux or Windows set `CHROME` to the browser's path, for example `CHROME=/usr/bin/google-chrome make e2e`. |

- `make vulncheck` runs `govulncheck` (downloaded on first use, so it needs internet) to look for known vulnerabilities in the dependencies and the Go standard library (and in the pinned tools: Pebble, which fails the check, and GoReleaser, whose findings are only shown, since they are in that build tool's own libraries); run it now and then, since new advisories appear over time. It also says when a newer release of a pinned tool exists. `govulncheck` itself is deliberately not pinned, since an old copy would miss new advisories.
- `make test` / `make race` run the Go tests (the race detector is the one to run before committing); `make fuzz` fuzzes the show-file parser.
- `make e2e` is the browser end-to-end test (about 200 checks). It also runs the browser-compatibility check (the pages must stay inside the old-browser floor) and the translation check.
- `make a11y` is the accessibility audit (axe-core, reflow and keyboard focus in both colour schemes); keep it at zero violations.
- `make i18n` checks that every on-screen sentence is translated into all six languages (it needs Node only).
- `make pebble-test` runs the real automatic-certificate path against Pebble, Let's Encrypt's test CA; it installs the **pinned** version (`PEBBLE_VERSION` in the Makefile) into `.tools/`, so it needs Go and internet the first time.
- `make acme-variants` checks that every automatic-certificate build choice still compiles (below); it needs internet.
- More detail is in [test/e2e/README.md](test/e2e/README.md).

**Automatic HTTPS certificates (optional).** If you own a public domain name, the program can get and renew a Let's Encrypt certificate itself with the DNS-01 challenge, without being reachable from the Internet. It needs the DNS provider's API credentials, and the provider must be built in. The default build includes **Cloudflare, GoDaddy and EasyDNS** (about 1.7 MB more than a build without ACME). Choose others when you build:

```
make build ACME_PROVIDERS="easydns route53"   # exactly these (names are lego's provider names)
make build ACME_PROVIDERS=all                 # every provider lego has (a much larger binary, about 61 MB)
make build ACME_PROVIDERS=none                # no automatic certificates (the smallest binary)
```

`./bin/open-caption --version` lists what a binary contains. Building with a provider outside the default list needs internet. To set it up, follow [docs/https-certificate-guide.md](docs/https-certificate-guide.md); every setting is in [docs/configuration-reference.md](docs/configuration-reference.md); the design, measured sizes and tests are in [docs/10-acme-lets-encrypt.md](docs/10-acme-lets-encrypt.md).

<!-- source-only:end -->

## Security

**Security note:** anyone who can reach the server can open the Setup and operator sign-in pages, and over plain `http://` passwords cross the WiFi unencrypted. Limit them with `access.admin_allow` / `access.operator_allow` in `config.yaml` (see [docs/usage-guide.md](docs/usage-guide.md#security)).

**HTTPS** (off by default) encrypts the sign-in. Turn it on with `https.enabled` in `config.yaml`, and choose where the certificate comes from: one the program makes itself (every browser warns about it until it is accepted), your own certificate files (a renewed file is picked up without a restart), or, if you own a public domain name, a trusted Let's Encrypt certificate that the program gets and renews by itself (ACME, DNS-01). Step by step: [docs/https-certificate-guide.md](docs/https-certificate-guide.md); every setting: [docs/configuration-reference.md](docs/configuration-reference.md). Plain `http://` stays the simplest address for the audience, and the passwords are only encrypted when the sign-in page is opened through the `https://` address.

<!-- source-only:start -->
## Spec-driven workflow

The specifications in `docs/` are the source of truth. Code follows the specs; if
behavior needs to change, change the spec first (in the same PR). (The guides for people who *use* the program are
under [Documentation](#documentation) above.)

| Spec | Purpose |
|-----|---------|
| [docs/01-vision-and-requirements.md](docs/01-vision-and-requirements.md) | Goals, users, scope, numbered requirements (REQ-xxx) |
| [docs/02-architecture.md](docs/02-architecture.md) | Components, state model, transport, persistence, package layout |
| [docs/03-show-file-format.md](docs/03-show-file-format.md) | The show file syntax and its parsed model |
| [docs/04-protocol-and-api.md](docs/04-protocol-and-api.md) | HTTP endpoints, event stream, message schemas |
| [docs/05-ui-specs.md](docs/05-ui-specs.md) | Setup, Operator, and Audience interface specs |
| [docs/06-deployment-and-network.md](docs/06-deployment-and-network.md) | Running on macOS/Windows/Linux, WiFi, DNS/DHCP, QR code |
| [docs/07-testing-and-acceptance.md](docs/07-testing-and-acceptance.md) | Test strategy, load target, acceptance checklist |
| [docs/08-roadmap.md](docs/08-roadmap.md) | Milestones and task breakdown |
| [docs/09-open-questions.md](docs/09-open-questions.md) | Resolved decisions and remaining open questions |
| [docs/10-acme-lets-encrypt.md](docs/10-acme-lets-encrypt.md) | Automatic HTTPS certificates (ACME DNS-01): design, build choices, sizes, tests |
| [docs/releasing.md](docs/releasing.md) | Making a release: push a `v*` tag; what the downloads contain |

Requirement IDs (`REQ-…`) are referenced from the other docs and should be
referenced from tests and commits.
<!-- source-only:end -->

<!-- source-only:start -->
## Contributing

Bug reports, ideas, translation reviews from native speakers and private security reports are welcome; code pull requests are not being accepted yet. See [CONTRIBUTING.md](CONTRIBUTING.md) for how, and for the rules changes follow.
<!-- source-only:end -->

## License

Apache License 2.0 — see [LICENSE](LICENSE).

## If this is useful to you

open-caption is free to use, and a credit is **not required**, but it would be appreciated. If it helped your production, please consider saying so, in the captions themselves (a line at the start or end), in the printed program, or on your website, for example:

> Open captions by open-caption (github.com/tmshort/open-caption)

or "This production uses open-caption for its open captions". A star on the [repository](https://github.com/tmshort/open-caption), or a note in an issue about how it went (what worked, what did not), helps other theaters find it and helps it get better.
