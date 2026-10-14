# open-caption

A single-binary open-captioning server for community theater, written in Go.

An operator follows the live show and presses a key to release the next caption
line. Every audience device on the guest WiFi (phone, tablet, desktop, or a
browser on a projector) shows the captions live, with no interaction required.

## Spec-driven workflow

The documents in `docs/` are the source of truth. Code follows the specs; if
behavior needs to change, change the spec first (in the same PR).

| Doc | Purpose |
|-----|---------|
| [docs/01-vision-and-requirements.md](docs/01-vision-and-requirements.md) | Goals, users, scope, numbered requirements (REQ-xxx) |
| [docs/02-architecture.md](docs/02-architecture.md) | Components, state model, transport, persistence, package layout |
| [docs/03-script-format.md](docs/03-script-format.md) | The script file syntax and its parsed model |
| [docs/04-protocol-and-api.md](docs/04-protocol-and-api.md) | HTTP endpoints, event stream, message schemas |
| [docs/05-ui-specs.md](docs/05-ui-specs.md) | Setup, Operator, and Audience interface specs |
| [docs/06-deployment-and-network.md](docs/06-deployment-and-network.md) | Running on macOS/Windows/Linux, WiFi, DNS/DHCP, QR code |
| [docs/07-testing-and-acceptance.md](docs/07-testing-and-acceptance.md) | Test strategy, load target, acceptance checklist |
| [docs/08-roadmap.md](docs/08-roadmap.md) | Milestones and task breakdown |
| [docs/09-open-questions.md](docs/09-open-questions.md) | Decisions made by default that need confirmation |

Requirement IDs (`REQ-…`) are referenced from the other docs and should be
referenced from tests and commits.
