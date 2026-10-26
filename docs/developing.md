# Building, testing and the specifications

**To run the program you need nothing but the binary.** To build and test it (starting from a bare computer? Follow the step-by-step guide for [macOS](build-macos.md), [Windows](build-windows.md) or [Linux](build-linux.md)):

| You need | For | Notes |
|---|---|---|
| **Go 1.26 or later** | `make build` (on Windows, which has no make: `build.bat`), `make test`, `make race`, `make vet`, `make lint`, `make dist` | No CGO and no C compiler; `make dist` cross-compiles for macOS, Windows and Linux. `go.mod` names a `toolchain` (the Go release that has the standard-library security fixes `make vulncheck` looks for); a Go 1.21 or later downloads it by itself when yours is older, and `GOTOOLCHAIN=local` turns that off. Raise it when `make vulncheck` reports a fix in a newer Go. |
| **Node.js 22.12 or later** (CI uses 22) | `make e2e`, `make a11y`, `make i18n`, `make lint-js` | The first run downloads the test tools with `npm` (listed in `test/e2e/package.json`); after that the tests need no internet. They are `puppeteer-core` (drives Chrome), `axe-core` (the accessibility checker) and `acorn` (a JavaScript parser, used to check that the pages stay inside the old-browser floor). |
| **Google Chrome or Chromium** | `make e2e`, `make a11y` | The tests drive a real browser. They look for Chrome in its macOS location; on Linux or Windows set `CHROME` to the browser's path, for example `CHROME=/usr/bin/google-chrome make e2e`. |

- `make vulncheck` runs `govulncheck` (downloaded on first use, so it needs internet) to look for known vulnerabilities in the dependencies and the Go standard library (and in the pinned tools: Pebble, which fails the check, and GoReleaser, whose findings are only shown, since they are in that build tool's own libraries); run it now and then, since new advisories appear over time. It also says when a newer release of a pinned tool exists. `govulncheck` itself is deliberately not pinned, since an old copy would miss new advisories.
- `make test` / `make race` run the Go tests (the race detector is the one to run before committing); `make fuzz` fuzzes the show-file parser. `make cover` runs the Go tests with every package counted and prints the share of statements the tests ran, per package and in total; `coverage/go.html` shows the lines none ran (the `coverage/` folder is not committed). `make cover-js` runs the browser tests with Chrome's own coverage switched on and reports, per page script and style sheet, the lines that some code ran on (`coverage/js.txt`, and `coverage/js.html` with the lines no test ran in red); it is Chrome's block coverage, a guide to what the tests never touch, not a pass mark.
- `make e2e` is the browser end-to-end test (over 200 checks over plain http, then `test/e2e/https.js` with HTTPS on, which needs `openssl` and is skipped without it). It also runs the browser-compatibility check (the pages must stay inside the old-browser floor) and the translation check.
- `make lint-js` runs ESLint over the page scripts, the browser tests and tools (`eslint.config.mjs`); keep it at zero findings.
- `make a11y` is the accessibility audit (axe-core, reflow and keyboard focus in both colour schemes); keep it at zero violations.
- `make i18n` checks that every on-screen sentence is translated into all six languages (it needs Node only).
- `make pebble-test` runs the real automatic-certificate path against Pebble, Let's Encrypt's test CA; it installs the **pinned** version (`PEBBLE_VERSION` in the Makefile) into `.tools/`, so it needs Go and internet the first time.
- `make acme-variants` checks that every automatic-certificate build choice still compiles (below); it needs internet.
- More detail is in [test/e2e/README.md](../test/e2e/README.md).

**Automatic HTTPS certificates (optional).** If you own a public domain name, the program can get and renew a Let's Encrypt certificate itself with the DNS-01 challenge, without being reachable from the Internet. It needs the DNS provider's API credentials, and the provider must be built in. The default build includes **Cloudflare, GoDaddy and EasyDNS** (about 1.7 MB more than a build without ACME). Choose others when you build:

```
make build ACME_PROVIDERS="easydns route53"   # exactly these (names are lego's provider names)
make build ACME_PROVIDERS=all                 # every provider lego has (a much larger binary, about 61 MB)
make build ACME_PROVIDERS=none                # no automatic certificates (the smallest binary)
```

`./bin/open-caption --version` lists what a binary contains. Building with a provider outside the default list needs internet. To set it up, follow [https-certificate-guide.md](https-certificate-guide.md); every setting is in [configuration-reference.md](configuration-reference.md); the design, measured sizes and tests are in [10-acme-lets-encrypt.md](10-acme-lets-encrypt.md).

## Spec-driven workflow

The specifications in `docs/` are the source of truth. Code follows the specs; if
behavior needs to change, change the spec first (in the same PR). (The guides for people who *use* the program are
in the [README](../README.md).)

| Spec | Purpose |
|-----|---------|
| [01-vision-and-requirements.md](01-vision-and-requirements.md) | Goals, users, scope, numbered requirements (REQ-xxx) |
| [02-architecture.md](02-architecture.md) | Components, state model, transport, persistence, package layout |
| [03-show-file-format.md](03-show-file-format.md) | The show file syntax and its parsed model |
| [04-protocol-and-api.md](04-protocol-and-api.md) | HTTP endpoints, event stream, message schemas |
| [05-ui-specs.md](05-ui-specs.md) | Setup, Operator, and Audience interface specs |
| [06-deployment-and-network.md](06-deployment-and-network.md) | Running on macOS/Windows/Linux, WiFi, DNS/DHCP, QR code |
| [07-testing-and-acceptance.md](07-testing-and-acceptance.md) | Test strategy, load target, acceptance checklist |
| [08-roadmap.md](08-roadmap.md) | Milestones and task breakdown |
| [09-open-questions.md](09-open-questions.md) | Resolved decisions and remaining open questions |
| [10-acme-lets-encrypt.md](10-acme-lets-encrypt.md) | Automatic HTTPS certificates (ACME DNS-01): design, build choices, sizes, tests |
| [releasing.md](releasing.md) | Making a release: push a `v*` tag; what the downloads contain |

Requirement IDs (`REQ-…`) are referenced from the other docs and should be
referenced from tests and commits.
