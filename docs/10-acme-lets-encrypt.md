# Automatic HTTPS certificates (ACME, DNS-01)

Status: **built** (REQ-068). This document is the design, the build choices, the sizes and the tests; to **use** it, start with the [trusted HTTPS guide](https-certificate-guide.md). The program can get and renew a certificate from Let's Encrypt (or another ACME certificate authority) by itself, using the **DNS-01** challenge, so a venue computer that is only on a private network can still have a certificate that every browser trusts. It is optional and off by default; plain HTTP and the self-signed certificate keep working as before (REQ-006).

## Using it

- **Setting it up, step by step** (what you need, the EasyDNS/Cloudflare/GoDaddy credentials, staging first, checking it, troubleshooting): the [trusted HTTPS guide](https-certificate-guide.md).
- **Every setting** (`https.acme.*`), with defaults, valid values and the rules checked at start-up: the [configuration reference](configuration-reference.md).

Plain HTTP works with no warning and stays the default audience address; HTTPS matters mostly for features browsers restrict to secure pages (such as the Wake Lock API; the audience page has a fallback). Automatic certificates are for a venue that owns a public domain whose DNS provider is supported.

## How it works

Let's Encrypt must see that you control the name before it issues a certificate. With DNS-01 it checks a TXT record in your public DNS, so **the computer never has to be reachable from the Internet**:

1. The program asks the CA for a certificate for `captions.mytheater.org`.
2. It publishes `_acme-challenge.captions.mytheater.org` as a TXT record with a value from the CA. **This is what the DNS provider credentials are for:** they let the program add and delete that one record through the provider's API.
3. It waits for the record to spread, and the CA looks it up in public DNS.
4. The CA issues the certificate; the program deletes the record, keeps the certificate in `<data-dir>/tls/acme/` and starts serving it, with no restart.

Renewal is automatic: the program checks about every 12 hours (with a little random spread) and renews when `renew_before_days` (default 30) or fewer remain. A failed try is retried an hour later (CAs limit failed attempts) and **the previous certificate keeps serving**; before the first one arrives, connections get the self-signed certificate, so HTTPS is never down. The state is in Setup → Settings → Security → Certificate, with **Renew now**.

The venue network needs **outbound** access only: to the CA (`acme-v02.api.letsencrypt.org`), to the DNS provider's API, and to public DNS. The public name must resolve, from the audience's phones, to the computer's address (an `A` record such as `captions.mytheater.org → 192.168.1.20`; a certificate is valid even when the name points at a private address). Some networks refuse public names that answer with private addresses ("DNS rebinding protection"), so the numeric address and QR code stay the main way in.

### Credentials

Each provider reads its own settings (EasyDNS: `EASYDNS_TOKEN` and `EASYDNS_KEY`; every provider's names are in the [lego documentation](https://go-acme.github.io/lego/dns/)). Sources, highest first: the **environment** (including `NAME_FILE` variables), **`dns.credentials_file`** (read again at every attempt, so a rotated token needs no restart; a warning if other users can read it), **`dns.credentials`** in `config.yaml`. A DNS token can usually change every record in the zone, so give it as little scope as the provider allows, or use **CNAME delegation**: a permanent `_acme-challenge.<name>  CNAME  _acme-challenge.<small zone>` makes the library write its record in a small, otherwise empty zone. Credentials are never logged, never in the support bundle (`Config.Redacted`) and never sent to Setup.

## Choosing the DNS providers

Every provider the program can use is compiled in. The default build includes **Cloudflare, GoDaddy and EasyDNS**. A builder can choose another list when building:

```
make build                                          # the default: cloudflare godaddy easydns
make build ACME_PROVIDERS="easydns"                 # exactly one
make build ACME_PROVIDERS="easydns route53 ovh"     # a short list
make build ACME_PROVIDERS=all                       # every provider the lego library has
make build ACME_PROVIDERS=none                      # no automatic certificates at all
ACME_PROVIDERS="cloudflare porkbun" make dist      # the same choice for the release binaries
```

Provider names are the directory names under [`providers/dns`](https://github.com/go-acme/lego/tree/master/providers/dns) in lego (for example `route53`, `gcloud`, `azuredns`, `digitalocean`, `namecheap`, `ovh`, `hetzner`, `porkbun`, `duckdns`; `exec` runs a script of yours and `httpreq` calls a URL, which covers a provider nobody supports). `./bin/open-caption --version` lists what a binary was built with. A plain `go build ./cmd/open-caption` gives the default list, and `go test ./...` needs nothing special.

How it works: `tools/acme-go.sh` (which the Makefile calls) writes `internal/acme/zz_providers_gen.go`, a file that imports exactly the chosen providers, builds, and removes the file. Providers outside the default list need Go modules that `go.mod` does not list, so for them the build uses a copy of `go.mod`/`go.sum` (`go.acme.mod`) that Go fills in: the repository's `go.mod` stays small. **A build with such a provider, or `all`, needs internet.** `make acme-variants` builds none, the default, one provider, a custom list and `all`, to check they all still compile. A configuration that names a provider the binary does not have stops at start-up with a message that lists what is built in.

### What it costs in size

Measured on macOS/arm64 with the release flags (`-s -w`) on 2026-10-18, for the program as built:

| Build | Binary | Change |
|---|---|---|
| `ACME_PROVIDERS=none` | 10,931,122 bytes (10.4 MiB) | n/a |
| `easydns` only | 12,541,842 bytes | +1.6 MB |
| **default** (cloudflare, godaddy, easydns) | **12,610,098 bytes** | **+1.7 MB (+15%)** |
| `easydns route53` | 14,991,170 bytes | +4.1 MB |
| `all` | 60,916,482 bytes | **+50 MB** |

Almost all of the default's cost is the ACME library itself (about 1.6 MB); each small provider adds tens of kilobytes. Route 53 is expensive because it brings the AWS SDK (about 2.4 MB); Google Cloud DNS and Azure are similar or worse. `all` links 196 providers and their SDKs and is for builders who need it, not a default. (An earlier draft of these notes estimated 0.2 to 0.3 MB; that was wrong, because the throwaway programs measured then let the linker drop most of the library.)

## Design

- `internal/acme`: `Manager` keeps one certificate current (check, renew inside the window, retry after a failure, the new key and certificate written completely under temporary names and then moved in, an install interrupted between the two moves finished at the next start, the key at 0600, `GetCertificate` with a fallback, `RenewNow`, `Status`). It knows nothing of the protocol: an `Issuer` does that. The real `Issuer` (`issuer.go`) uses the lego library: an account per CA (an EC P-256 key and the registration, kept in `tls/acme/`), an EC P-256 certificate key, the DNS-01 solver with the chosen provider, and `propagation_wait` as a fixed wait. A CNAME at `_acme-challenge` is followed by the library automatically.
- Build tags choose the providers (`internal/acme/providers.go`): no tags is the default list (`prov_*.go`), `acme_off` leaves lego out, `acme_custom` uses the generated file.
- `config.HTTPS.ACME` with validation (`checkACME`); `cmd/open-caption` wires the `Manager` into the HTTPS listener and `httpapi.Server.SetCertificates` for Setup; `GET /api/admin/settings/certificate` and `POST …/certificate/renew` (admin only; doc 04).
- A certificate you supply yourself (`cert_file`/`key_file`) is a separate path and is watched for replacement (`netinfo.Reloader`): useful if you prefer `certbot` or `lego` outside the program.

## Testing

- Unit tests with a fake `Issuer` (no network): first issue, the files and their permissions, restart reuse, the renew window, a failed renewal keeping the old certificate, wrong-name and unusable certificates, `RenewNow` once at a time; the credentials precedence and the credentials file; the configuration rules; the Setup endpoints (and that credentials are never sent).
- An **end-to-end test against Pebble**, Let's Encrypt's test CA, with its test DNS server standing in for the provider (so lego, the DNS-01 challenge and our `Manager` run for real, on this computer only). It is off unless asked for, because it needs two programs. `make pebble-test` installs a **pinned** Pebble version (`PEBBLE_VERSION` in the Makefile, into `.tools/`, which `make clean` removes) and runs the test, so every run uses the same one; CI does the same. `make vulncheck` scans the pinned Pebble programs for known vulnerabilities too, and says when a newer Pebble release exists.
- **Not tested here:** a real provider's API (EasyDNS, Cloudflare, GoDaddy) and the real Let's Encrypt. Those are the library's code and the providers' services; the first run on a real venue should use `letsencrypt-staging`, as in the example above.

## Limits and ideas not built

- One name per certificate: no wildcard, no extra names, no HTTP-01 or TLS-ALPN-01 challenge (they need the computer reachable from the Internet).
- Setup's Share step still lists the numeric addresses and the `.local` name; it does not yet list `https://<your domain>/`. Give people that address yourself or put it on the QR sign.
- No OCSP stapling, no certificate transparency check, no revocation.
- The credentials can not be edited from Setup (by design: the settings file or the environment); Setup only shows the state.

## Failure cases

An expired certificate (the program warns in the log and Setup; the self-signed one still serves), a CA rate limit (use staging while testing), a DNS provider outage or API change (the next hourly try may succeed; a newer lego release fixes API changes), a computer that was off or offline when renewal was due (it renews as soon as it is running again, as long as the certificate has not yet expired), a wrong clock (certificate validity is time-based), a record that does not spread in time (set `propagation_wait`), and a CNAME for `_acme-challenge` pointing at a zone the credentials cannot change.
