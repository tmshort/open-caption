# Set up a trusted HTTPS certificate (optional)

This guide is for the person who looks after the captioning computer. It is **optional**: the captions work without it. (The settings it uses are in the [configuration reference](11-configuration-reference.md), and the design and tests are in [doc 10](10-acme-lets-encrypt.md); both are in English only.)

## Do you need this?

Probably not. Audience phones open `http://…` with no warning, and that is the address the QR code uses. A secure `https://` address is only needed for a few browser features that are limited to secure pages (such as keeping the phone screen awake the proper way; the audience page has a fallback), or if the venue wants it.

Without this guide, HTTPS still works with a **self-signed** certificate, but every phone shows a security warning. With it, the program gets a real, free certificate from Let's Encrypt and **renews it by itself**, so `https://captions.mytheater.org` opens with no warning on every device.

## What you need before you start

| You need | Why |
|---|---|
| A **domain name you own**, for example `mytheater.org`, and a name under it for the captions, such as `captions.mytheater.org` | A certificate is issued for a public name. `open-caption.local` and bare numbers cannot get one |
| Its **DNS hosted by a provider this program supports** | The program proves you own the name by adding a temporary record in your DNS. The default program includes **Cloudflare, GoDaddy and EasyDNS**; run `open-caption --version` to see what yours has. Another provider needs a different build (see the end) |
| **API credentials** for that DNS account | They let the program add and remove that one record |
| The computer to have **Internet access** (outbound only) | It talks to Let's Encrypt and to your DNS provider. Nothing needs to reach *into* the venue |
| A **fixed address** for the computer on the venue network (a DHCP reservation is fine) | The public name will point at it |

## Steps

**1. Get the API credentials.** In your DNS provider's account:

- **EasyDNS:** enable API access and create an API **token and key**. The settings are `EASYDNS_TOKEN` and `EASYDNS_KEY`. (EasyDNS has a sandbox for trying the API; see their documentation for the current steps.)
- **Cloudflare:** create an **API token** limited to your zone with the permission *Zone → DNS → Edit*. The setting is `CLOUDFLARE_DNS_API_TOKEN`.
- **GoDaddy:** create a production API **key and secret**. The settings are `GODADDY_API_KEY` and `GODADDY_API_SECRET`. (GoDaddy may limit its DNS API to some account types; check their current rules.)
- Any other provider: the exact setting names are in the [lego documentation](https://go-acme.github.io/lego/dns/).

A DNS token can usually change **every** record in the zone, not just the temporary one, so give it as little power as the provider allows. For the strongest protection use a CNAME (see "Safer credentials" below).

**2. Point the name at the computer.** In your DNS provider, add an `A` record: `captions.mytheater.org` → the computer's address on the venue network (for example `192.168.1.20`). It is fine for a public name to point at a private address; the certificate does not care.

**3. Keep the credentials in a private file.** Create `/etc/open-caption/dns.env` (any private place works; on Windows use a folder only your account can read) with one line per setting, for example for EasyDNS:

```
EASYDNS_TOKEN=paste-the-token-here
EASYDNS_KEY=paste-the-key-here
```

On macOS and Linux make it private: `chmod 600 /etc/open-caption/dns.env`. The program warns if other users can read it.

**4. Edit `config.yaml`** (the first lines the program prints say where it is; every setting is explained in the [reference](11-configuration-reference.md)):

```yaml
https:
  enabled: yes
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org      # Let's Encrypt writes here before a certificate expires
    agree_to_terms: yes                 # you accept Let's Encrypt's terms of service
    directory: letsencrypt-staging      # practice first (see step 5)
    dns:
      provider: easydns                 # or cloudflare, godaddy
      credentials_file: /etc/open-caption/dns.env
```

**5. Practice with the staging service first.** `letsencrypt-staging` is Let's Encrypt's test service: it works exactly like the real one, but its certificates are **not trusted by browsers** (a warning is expected) and it has generous limits. Start the program.

**6. Check that it worked.** Open **Setup → Settings → Security → Certificate**. The state goes from *Waiting for the first certificate* to **Valid** within a few minutes, and shows the name, who issued it and the expiry date. The program's log says `HTTPS: got a certificate for … valid until …`. If it says *The last attempt failed*, read **Last problem** and the table below.

**7. Switch to the real certificate.** Change `directory: letsencrypt-staging` to `directory: letsencrypt` (or delete the line) and restart the program. The real certificate arrives the same way. Open `https://captions.mytheater.org/` from a phone on the venue WiFi: no warning. (If port 443 was busy the program used the fallback, `:8443`; the address then needs the port, `https://captions.mytheater.org:8443/`, and the log shows the exact address.)

**8. Give people the address.** Setup's Run step lists the numeric addresses and the `.local` name but does not yet list this one. To print a QR sign for it, while signed in to Setup open
`http://localhost/admin?print=qr&url=https%3A%2F%2Fcaptions.mytheater.org%2F` (change the name; add the port if needed) and print that page.

## After that

- **Renewal is automatic.** The program checks about twice a day and renews when 30 days or fewer remain. You do not need to do anything.
- **If the Internet is down at the venue,** nothing breaks: the program keeps serving the certificate it has and tries again an hour later. Before the very first certificate arrives (or if one never does) phones get the self-signed one.
- **Renew now** (Setup → Settings → Security → Certificate) fetches a new certificate immediately. Use it after fixing a problem.
- **Changing the DNS token:** edit `/etc/open-caption/dns.env`. It is read again at the next attempt, so no restart is needed.
- **Turning it off:** set `acme: enabled: no` and restart. HTTPS then uses the self-signed certificate (or your own files).

## If something goes wrong

Setup shows **Last problem**; the log has the same text. The common ones:

| You see | Likely cause and fix |
|---|---|
| `the DNS provider "…" is not built into this program` (the program stops at start-up) | This copy of the program was built without that provider. Use `cloudflare`, `godaddy` or `easydns`, or get a build with yours (below) |
| `… credentials information are missing: EASYDNS_TOKEN,EASYDNS_KEY` | The names in the file are wrong or the file cannot be read. Names are capital letters, `NAME=value`, one per line |
| An error from the DNS provider such as *unauthorized*, *forbidden* or a 4xx number | The token or key is wrong, expired, or not allowed to edit that zone. Create a new one |
| *Waiting for the first certificate* for a long time, or a message about the record not being found | The new DNS record is slow to spread. Add `propagation_wait: 3m` under `dns:` and press **Renew now** |
| `too many certificates already issued` or `rate limit` | Let's Encrypt limits repeats. Wait (the message says how long) and use `letsencrypt-staging` while testing |
| The browser still warns | You are using the staging service (switch to `letsencrypt`), or you opened the numeric address or another name instead of `https://captions…`, or the clock on the computer or phone is wrong |
| A phone cannot open the name | Some guest WiFi networks and routers block public names that answer with private addresses ("DNS rebinding protection"). Ask the network's owner to allow `captions.mytheater.org`, or use the numeric address and QR code as before |
| `https.acme: cannot be used together with cert_file/key_file` | You set both. Remove one: automatic certificates **or** your own files |
| The program refuses to start and lists problems in `config.yaml` | Each line names the setting and what is wrong. Fix them all and start again |

## Safer credentials (optional)

A DNS token that can change your whole zone is a risk if the computer is lost or shared. Two ways to reduce it:

- Keep the secrets **out of `config.yaml`**: in the credentials file (step 3), or as environment variables of the service that starts the program (`EASYDNS_TOKEN_FILE=/path/to/file` also works: the setting is read from that file).
- Use **CNAME delegation**: in your DNS add a permanent record `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.other-zone.example`, where `other-zone.example` is a small, otherwise empty zone. The program follows it automatically, so the token only needs access to that small zone, not your main one.

The credentials are never written to the log, the support bundle or Setup.

## A different DNS provider

The program only contains the DNS providers it was built with. People who build it can choose others, or all of them, with `make build ACME_PROVIDERS="…"` (names are those in the [lego list](https://go-acme.github.io/lego/dns/)); the sizes and the steps are in [doc 10](10-acme-lets-encrypt.md). If your provider is not supported at all, `exec` (runs a script of yours) and `httpreq` (calls a web address) can drive almost any DNS service.
