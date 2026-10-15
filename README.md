# open-caption

A single-binary open-captioning server for community theater, written in Go.

An operator follows the live show and presses a key to release the next caption
line. Every audience device on the guest WiFi (phone, tablet, desktop, or a
browser on a projector) shows the captions live, with no interaction required.

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
