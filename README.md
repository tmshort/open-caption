# open-caption

A single-binary open-captioning server for community theater, written in Go.

An operator follows the live show and presses a key to release the next caption
line. It is **bring your own device (BYOD)**: the audience uses their own
phones and tablets on the venue's WiFi to see the captions live, with no app to
install and nothing to hand out. The venue can also put the captions on a
**projector or TV** for everyone: open the same address in a browser on a
computer connected to it.

- **Audience:** open one address (or scan a QR code) on a phone, tablet, projector or TV; large high-contrast captions, text size and dyslexia-friendly fonts, scroll back, the screen stays awake, several languages in one show.
- **Operator:** one key does it all (Space sends the next line, Backspace undoes it); notes, loops for pre-show and intermission, songs and two voices at once; pacing hints with a countdown; a standby screen until the show starts; a second device can watch at `/follow`.
- **Setup:** a guided page for the show file, images, look, a private rehearsal and a printable QR sign; a report of how many devices used the captions, once the show ends.
- **A show is one plain-text Markdown file** you can edit in any program; earlier versions are kept.
- **Six languages:** Setup, the operator page, Help and the guides are in English, Spanish, French, German, Italian and Portuguese.
- **One Go binary** for macOS, Windows and Linux: no database, no internet needed, tested with 200 simultaneous devices.

The whole list is in [docs/features.md](docs/features.md).

## Quick start

**1. Get the program.** On this repository's *Releases* page, download the file for your computer, unzip (Windows) or untar (macOS, Linux) it, and open a terminal in the unpacked folder. Intel and AMD computers are `amd64`; Apple silicon (M1 and later) and ARM are `arm64`. The programs are not code-signed, so the first time:

- **macOS:** if it cannot be opened or is from an unidentified developer, run `xattr -dr com.apple.quarantine .` in the unpacked folder, or try to open it once, then **System Settings → Privacy & Security → Open Anyway**.
- **Windows:** if SmartScreen says "Windows protected your PC", click **More info → Run anyway**. If the zip is blocked, right-click it → Properties → **Unblock** before unzipping.
- **Linux:** nothing to do.

<!-- source-only:start -->
Or build it from this source (a few minutes; it needs Go and `git`, plus `make` except on Windows). The guides start from a brand-new computer: [macOS](docs/build-macos.md) (`make build`), [Windows](docs/build-windows.md) (`.\build.bat`) and [Linux](docs/build-linux.md) (`make build`). The program ends up in `bin/`. More in [docs/developing.md](docs/developing.md).

<!-- source-only:end -->
**2. Run it** from the folder it is in (a build: `bin/`):

| Your computer | Command |
|---|---|
| macOS, Linux | `./open-caption` |
| Windows (PowerShell) | `.\open-caption.exe` |

The first run creates `config.yaml` (in your user config folder; the path is printed) with random **admin** and **operator** passwords, printed once in the console, and opens Setup in your browser. When the system asks whether to allow network connections, allow them (on Windows, tick *Private networks*), or phones cannot reach the program. Press `Ctrl`+`C` to stop it.

**3. Use it.**

- **Setup:** `http://localhost/admin` (or the `:8080` address printed if port 80 was busy; on Linux a normal user cannot use port 80, see the Linux guide). It needs no password on the computer that runs the program; elsewhere, the admin password. Follow the steps; "Create a sample show" is a good first try.
- **Operator:** opened from Setup's Share step (or `/operator`, with the operator password). Space sends the next line.
- **Follow (optional):** `/follow`, with the viewer password (or the operator or admin one), lets a second device such as a stage manager's tablet watch the show; it can send nothing.
- **Audience:** the address or QR code in Setup → Share, on the same WiFi.
- **Help:** the **Help** button in Setup and the operator page; [Getting started](web/help/admin-getting-started.md) and [Preparing a show file](web/help/admin-preparing-a-show-file.md) cover the rest.

`open-caption --help` lists the command-line options; the manual page is `docs/open-caption.1`.

## Documentation

| Guide | For |
|-------|-----|
| [docs/usage-guide.md](docs/usage-guide.md) | **Start here:** deployment recommendations, setting up a show, running it, the day-of checklist, troubleshooting |
| [docs/venue-network-guide.md](docs/venue-network-guide.md) | Whoever manages the venue's WiFi: how audience phones reach the computer |
| [docs/https-certificate-guide.md](docs/https-certificate-guide.md) | A trusted HTTPS certificate with Cloudflare, GoDaddy or EasyDNS, step by step |
| [docs/configuration-reference.md](docs/configuration-reference.md) | Every `config.yaml` setting, the command-line options and environment variables |
| [docs/features.md](docs/features.md) | Everything the program can do (English) |
| [docs/open-caption.1](docs/open-caption.1) | The manual page (`man ./docs/open-caption.1`) (English) |
| [docs/developing.md](docs/developing.md) | Building, testing, and the design specifications (English) <!-- source-only-line --> |

The first four guides are also in Spanish, French, German, Italian and Portuguese (`docs/i18n/`), as is everything behind the app's **Help** button ([web/help/](web/help/)).

## Security

Anyone who can reach the server can open the Setup and operator sign-in pages, and over plain `http://` passwords cross the WiFi unencrypted; it is meant for a private network. Limit who can sign in with `access.admin_allow` / `access.operator_allow` in `config.yaml` ([docs/usage-guide.md](docs/usage-guide.md#security)). **HTTPS** (off by default, `https.enabled`) encrypts the sign-in, with a certificate the program makes itself, your own files, or a trusted Let's Encrypt certificate it gets and renews by itself: [docs/https-certificate-guide.md](docs/https-certificate-guide.md). Plain `http://` stays the simplest address for the audience.

<!-- source-only:start -->
## Contributing

Bug reports, ideas, translation reviews from native speakers and private security reports are welcome; code pull requests are not being accepted yet. See [CONTRIBUTING.md](CONTRIBUTING.md) for how, and for the rules changes follow. The design specifications are the source of truth: [docs/developing.md](docs/developing.md).

<!-- source-only:end -->
## License

Apache License 2.0 — see [LICENSE](LICENSE).

## If this is useful to you

open-caption is free to use, and a credit is **not required**, but it would be appreciated. If it helped your production, please consider saying so, in the captions themselves (a line at the start or end), in the printed program, or on your website, for example:

> Open captions by open-caption (github.com/tmshort/open-caption)

or "This production uses open-caption for its open captions". A star on the [repository](https://github.com/tmshort/open-caption), or a note in an issue about how it went (what worked, what did not), helps other theaters find it and helps it get better.
