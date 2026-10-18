# Automatic HTTPS certificates (ACME, DNS-01)

Status: **built** (REQ-068). The program can get and renew a certificate from Let's Encrypt (or another ACME certificate authority) by itself, using the **DNS-01** challenge, so a venue computer that is only on a private network can still have a certificate that every browser trusts. It is optional and off by default; plain HTTP and the self-signed certificate keep working as before (REQ-006).

## When to use it

Plain HTTP works with no warning, and it is the default audience address. HTTPS matters mostly for features browsers restrict to secure pages, such as the Wake Lock API that keeps phone screens on (the audience page falls back to a tiny looping video without it). A self-signed certificate shows a browser warning on every device. Use automatic certificates if you **own a public domain name** (for example `mytheater.org`), **host its DNS with a provider whose API this program supports** (EasyDNS, Cloudflare and GoDaddy by default), and want audiences to open `https://captions.mytheater.org` without a warning.

You do not need it if the venue has no domain, or HTTP is fine.

## How it works

Let's Encrypt must see that you control the name before it issues a certificate. With DNS-01 it checks a TXT record in your public DNS, so **the computer never has to be reachable from the Internet**:

1. The program asks the CA for a certificate for `captions.mytheater.org`.
2. It publishes `_acme-challenge.captions.mytheater.org` as a TXT record with a value from the CA. **This is what the DNS provider credentials are for:** they let the program add and delete that one record through the provider's API.
3. It waits for the record to spread, and the CA looks it up in public DNS.
4. The CA issues the certificate; the program deletes the record, keeps the certificate in `<data-dir>/tls/acme/` and starts serving it, with no restart.

Renewal is automatic: the program checks about every 12 hours and renews when 30 days (`renew_before_days`) or fewer remain. If a try fails it tries again an hour later (CAs limit failed attempts), and **keeps serving the previous certificate**; before the first certificate arrives, connections get the self-signed one, so HTTPS is never down. The state is shown in **Setup → Settings → Security → Certificate**, with a **Renew now** button.

What the venue's network needs: **outbound** access from the computer to the CA (`acme-v02.api.letsencrypt.org`), to the DNS provider's API, and to public DNS. Nothing inbound. The public name must resolve, from the audience's phones, to the computer's address on the venue network (an `A` record such as `captions.mytheater.org → 192.168.1.20`; a certificate is valid even when the name points at a private address). Some networks refuse public names that answer with private addresses ("DNS rebinding protection"); the numeric address and QR code must stay the main way in.

## Configuration (`config.yaml`)

```yaml
https:
  enabled: true                           # automatic certificates need HTTPS on
  acme:
    enabled: true
    domain: captions.mytheater.org        # the public name; one name only (no wildcard, not .local, not an IP address)
    email: webmaster@mytheater.org        # the CA sends expiry notices here
    agree_to_terms: true                  # required: you accept the CA's terms of service
    directory: letsencrypt                # letsencrypt (default) | letsencrypt-staging | an https:// ACME directory address
    renew_before_days: 30                 # optional, 1 to 60
    dns:
      provider: easydns                   # a provider built into this program (see "Choosing the DNS providers")
      credentials:                        # optional: the provider's own settings, named as the lego library documents them
        EASYDNS_TOKEN: "…"
        EASYDNS_KEY: "…"
      credentials_file: /etc/open-caption/dns.env   # optional: NAME=value lines, read at every attempt
      propagation_wait: 2m                # optional: wait this long for the TXT record instead of checking for it
```

Rules: `acme` cannot be used together with `cert_file`/`key_file` (choose one way to get the certificate). Unknown keys, a provider that is not built in, and every other mistake are reported together at start-up, naming the setting.

### Credentials

Each DNS provider reads its own settings (for EasyDNS: `EASYDNS_TOKEN` and `EASYDNS_KEY`; the names for every provider are in the [lego documentation](https://go-acme.github.io/lego/dns/)). They can come from **three places**; highest first:

1. **The environment of the program**, including `NAME_FILE` variables: `EASYDNS_TOKEN_FILE=/etc/open-caption/easydns-token` makes the library read the value from that file. Nothing about credentials then appears in `config.yaml`.
2. **`dns.credentials_file`**: a file of `NAME=value` lines (blank lines and `#` comments allowed, an optional `export`, and quotes around the value). It is read again at every attempt, so a rotated token needs no restart. The program warns if other users can read the file; `chmod 600` it.
3. **`dns.credentials`** in `config.yaml` (the file is mode 0600, like the passwords).

Prefer 1 or 2: a DNS token can usually change **every** record in the zone. Give it as little scope as the provider allows, or use CNAME delegation: add a permanent record `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.some-small-zone.example` and the library follows it, so the token only needs access to that small, otherwise empty zone. Credentials are never logged, never put in the support bundle and never sent to Setup.

### Example: EasyDNS

1. In your EasyDNS account, enable API access and create an API **token and key** (see EasyDNS's API documentation for the current steps and for a sandbox to try it in).
2. Add the `A` record `captions.mytheater.org → <the computer's address>` in EasyDNS (give the computer a fixed address or a DHCP reservation).
3. Put the secrets in a file only the program's account can read, `/etc/open-caption/dns.env`:
   ```
   EASYDNS_TOKEN=…
   EASYDNS_KEY=…
   ```
4. In `config.yaml` set `https.enabled: yes` and the `acme` block above with `provider: easydns` and `credentials_file`. **Start with `directory: letsencrypt-staging`**: the certificate is not trusted by browsers, but it proves the whole chain works without using the real CA's limits.
5. Start the program, open Setup → Settings → Security → Certificate, and wait for the state **Valid** (a few minutes at most). The log says `HTTPS: got a certificate for … valid until …`.
6. Change `directory` to `letsencrypt` (or delete the line) and restart; the real certificate arrives the same way. Open `https://captions.mytheater.org:<port>/`.

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

- `internal/acme`: `Manager` keeps one certificate current (check, renew inside the window, retry after a failure, atomic files with the key at 0600, `GetCertificate` with a fallback, `RenewNow`, `Status`). It knows nothing of the protocol: an `Issuer` does that. The real `Issuer` (`issuer.go`) uses the lego library: an account per CA (an EC P-256 key and the registration, kept in `tls/acme/`), an EC P-256 certificate key, the DNS-01 solver with the chosen provider, and `propagation_wait` as a fixed wait. A CNAME at `_acme-challenge` is followed by the library automatically.
- Build tags choose the providers (`internal/acme/providers.go`): no tags is the default list (`prov_*.go`), `acme_off` leaves lego out, `acme_custom` uses the generated file.
- `config.HTTPS.ACME` with validation (`checkACME`); `cmd/open-caption` wires the `Manager` into the HTTPS listener and `httpapi.Server.SetCertificates` for Setup; `GET /api/admin/settings/certificate` and `POST …/certificate/renew` (admin only; doc 04).
- A certificate you supply yourself (`cert_file`/`key_file`) is a separate path and is watched for replacement (`netinfo.Reloader`): useful if you prefer `certbot` or `lego` outside the program.

## Testing

- Unit tests with a fake `Issuer` (no network): first issue, the files and their permissions, restart reuse, the renew window, a failed renewal keeping the old certificate, wrong-name and unusable certificates, `RenewNow` once at a time; the credentials precedence and the credentials file; the configuration rules; the Setup endpoints (and that credentials are never sent).
- An **end-to-end test against Pebble**, Let's Encrypt's test CA, with its test DNS server standing in for the provider (so lego, the DNS-01 challenge and our `Manager` run for real, on this computer only). It is off unless asked for, because it needs two programs:
  ```
  go install github.com/letsencrypt/pebble/v2/cmd/pebble@latest github.com/letsencrypt/pebble/v2/cmd/pebble-challtestsrv@latest
  PEBBLE_BIN=$(go env GOPATH)/bin PEBBLE_DIR=$(ls -d $(go env GOMODCACHE)/github.com/letsencrypt/pebble/v2@* | tail -1) go test -tags pebble -run Pebble ./internal/acme
  ```
- **Not tested here:** a real provider's API (EasyDNS, Cloudflare, GoDaddy) and the real Let's Encrypt. Those are the library's code and the providers' services; the first run on a real venue should use `letsencrypt-staging`, as in the example above.

## Limits and ideas not built

- One name per certificate: no wildcard, no extra names, no HTTP-01 or TLS-ALPN-01 challenge (they need the computer reachable from the Internet).
- Setup's Run step still lists the numeric addresses and the `.local` name; it does not yet list `https://<your domain>/`. Give people that address yourself or put it on the QR sign.
- No OCSP stapling, no certificate transparency check, no revocation.
- The credentials can not be edited from Setup (by design: the settings file or the environment); Setup only shows the state.

## Failure cases

An expired certificate (the program warns in the log and Setup; the self-signed one still serves), a CA rate limit (use staging while testing), a DNS provider outage or API change (the next hourly try may succeed; a newer lego release fixes API changes), a computer that was off or offline when renewal was due (it renews as soon as it is running again, as long as the certificate has not yet expired), a wrong clock (certificate validity is time-based), a record that does not spread in time (set `propagation_wait`), and a CNAME for `_acme-challenge` pointing at a zone the credentials cannot change.
