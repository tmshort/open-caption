# ACME / Let's Encrypt: design notes

Status: **not built.** This records what automatic certificates would involve, so the decision can be made later without redoing the research. It is not a requirement (see REQ-006 and doc 06, "Security posture").

## Today

HTTPS uses a certificate file the venue supplies (`https.cert_file` / `https.key_file`) or a self-signed certificate the program generates in `<data-dir>/tls/`. Plain HTTP stays the default audience address because a self-signed certificate shows a browser warning. A venue that has a real certificate already can use it; renewal is then up to them. The program has no renewal or reload of its own: a new certificate file is read at the next restart.

## What a public CA needs

A certificate authority such as Let's Encrypt must see that you control the name before it issues a certificate. For this program that rules out the easy cases:

- The name must be a **public DNS name** the venue owns (for example `captions.mytheater.org`). `open-caption.local` can never get a certificate, and a bare IP address cannot get a normal one.
- The program (or something beside it) needs **Internet access** when the certificate is issued and again at each renewal, about every 60 to 90 days.

## Challenge types

| Challenge | How control is proved | Fit |
|---|---|---|
| HTTP-01 | The CA connects to port 80 of the server from the Internet | Poor: the server is on a private venue network |
| TLS-ALPN-01 | The same, on port 443 | Poor, for the same reason |
| **DNS-01** | A TXT record is published in the public DNS zone | **The realistic choice**: the server never has to be reachable from outside |

With DNS-01 the public name can point at a private address (`captions.mytheater.org` → `192.168.1.20`) and the certificate is still valid.

### What the DNS provider token is for

During DNS-01 the program must publish, then remove, one TXT record:

1. It asks the CA for a challenge for `captions.mytheater.org`.
2. It creates `_acme-challenge.captions.mytheater.org` as a TXT record holding a value the CA gave it. This is the only thing the DNS credentials are used for, and it is done through the provider's API.
3. It waits for the record to appear, then tells the CA, which looks the record up in public DNS.
4. The CA issues the certificate and the program deletes the TXT record.

So the **API token (and key) is the password the program uses to add and delete DNS records in the venue's zone.** Without it the program cannot answer the challenge on its own. Consequences:

- Treat it like the other secrets: keep it in `config.yaml` (mode 0600) or an environment variable, never in a show file, never in the support bundle or the log.
- A DNS token can usually change **every** record in the zone, not just `_acme-challenge`. Give it as little scope as the provider allows, or use the delegation trick below.
- **Delegation (limits the damage).** Add a permanent CNAME `_acme-challenge.captions.mytheater.org` → `_acme-challenge.some-small-zone.example`, in a different zone that holds nothing else. The program's token then only needs access to that small zone. Most ACME clients support this ("DNS alias" mode).

### EasyDNS (this venue's provider)

- EasyDNS has a REST API, enabled per account, authenticated with an **API token and key** pair created in the EasyDNS account; there is a sandbox for testing. (Check EasyDNS's current API documentation before building: account settings and token scopes change.)
- The ACME client library `lego` has an `easydns` DNS provider, configured with `EASYDNS_TOKEN` and `EASYDNS_KEY` (and an optional endpoint). Verify the provider is still current in the library's list when the work starts. Whether `certmagic` (through `libdns`) has an EasyDNS provider is not known and would have to be checked; if not, a small provider would have to be written against EasyDNS's REST API.
- Propagation matters: EasyDNS must publish the TXT record to its public name servers before the CA checks, so the client should poll and wait rather than assume a fixed delay.

## What it would look like

1. **Configuration:** an `https.acme` block in `config.yaml`: domain, contact email, provider (`easydns`), credentials (or `*_FILE` references to files), and an optional challenge alias. Setup's Settings would get a read-only "Certificate" panel: the name, who issued it, when it expires, the last renewal result, and a **Renew now** button. The credentials are never sent back to the browser.
2. **Library.** `golang.org/x/crypto/acme/autocert` is the standard choice but **does not do DNS-01**, only HTTP-01 and TLS-ALPN-01. DNS-01 needs `lego` or `certmagic`.
3. **Storage and renewal.** Keep the account key and certificates in `<data-dir>/tls/`, renew about 30 days before expiry in the background, and swap the certificate into running listeners through `tls.Config.GetCertificate`, so no restart is needed.
4. **Failure behaviour.** If issuance or renewal fails (for example the venue computer has no Internet that week), keep serving the last good certificate, or the self-signed one, and show a warning in Setup. A show must never fail because the Internet was down.
5. **Name resolution for audiences.** Phones open `https://captions.mytheater.org`. They resolve it through public DNS, so the name must point at the computer's private address. Some networks refuse public names that answer with private addresses ("DNS rebinding protection"); guest WiFi with filtering can break it. The numeric address and QR code must remain the primary way in.
6. **Testing.** Use the CA's staging environment and EasyDNS's sandbox. Unit tests must never touch the Internet (the rule already used for mDNS: the real network is only touched from `main`), so the DNS provider and the CA sit behind small interfaces with fakes.

## What "compatible but not free" means

The project is one binary with no CGO and very few dependencies (today only `yaml.v3` and `golang.org/x/net`). Any ACME library is compatible with that (pure Go, cross-compiles with `make dist`), but each one has a cost:

- `autocert` is small, but it is in `golang.org/x/crypto`, a **new module** the project does not use today, and it cannot do DNS-01.
- `lego` and `certmagic` can do DNS-01, but each brings **many dependencies** (HTTP clients, a DNS library, and code for dozens of DNS providers). That makes the binary larger (not measured here; expect a noticeable number of megabytes), lengthens builds, and widens the supply-chain surface: more code whose security updates must be followed.
- Either way the project takes on code to maintain, a test suite for failure cases, and a dependency whose provider list can break when a DNS company changes its API.

So "compatible" means it fits the single-binary, no-CGO rules; "not free" means it adds size, dependencies and maintenance.

## Failure cases to design for

An expired certificate, a CA rate limit, a DNS provider API change or outage, a computer that was off or offline when renewal was due, and a clock that is wrong (certificate validity is time-based).

## Why it may not be worth it

- Plain HTTP works with no warning. HTTPS mostly matters for features browsers restrict to secure pages, such as the Wake Lock API; the audience page already falls back to a tiny looping video where Wake Lock is missing.
- It needs a domain and DNS credentials, which many community venues don't have or want to manage.
- The failure modes above are awkward for volunteers.

## Cheaper steps first

1. **Reload without restart:** pick up a renewed `cert_file` / `key_file` while running (a file watcher or a **Reload certificate** button).
2. **A documented recipe** for obtaining the certificate outside the program with `certbot` or `lego` (DNS-01 with EasyDNS), then pointing `cert_file` and `key_file` at it, with a scheduled renewal.
3. Only then, if there is demand, build the DNS-01 integration above.
