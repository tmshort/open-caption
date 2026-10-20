# 06 — Deployment and Network

## Running

```
open-caption                       # defaults
open-caption --config ~/oc/config.yaml   # config file (default: <data dir>/config.yaml)
open-caption --data ~/oc-data            # data directory
open-caption --http-addr :80 --https-addr :443   # override listen addresses
open-caption --no-browser                # don't auto-open the setup page
open-caption --log-level debug --log-file ~/oc.log   # more detail, also to a file
open-caption --version                   # the version and the ACME DNS providers built in
open-caption --check-show show.md        # check a show file (a lint); no server starts
open-caption --help                      # every option
```
Every option, its environment variable (`OPEN_CAPTION_*`) and what it overrides is in the
[configuration reference](configuration-reference.md#command-line-and-environment); the same is in the
manual page, `docs/open-caption.1` (`man ./docs/open-caption.1`, or install it in a `man1` folder).
Every setting (ports, addresses, TLS, passwords, defaults) lives in
`config.yaml` (doc 02). Flags and `OPEN_CAPTION_*` environment variables override it.
On first run the config is generated with random admin/operator passwords.

On start, the server logs and shows (in the console and setup UI) all
non-loopback IPv4 addresses and the audience URL, and opens
`http://localhost:<port>/admin` in the default browser (macOS `open`, Windows
`rundll32 url.dll,FileProtocolHandler`, Linux `xdg-open`).

## Platforms

| OS | Notes |
|----|-------|
| macOS | Unsigned binary requires Gatekeeper approval (right-click → Open) unless signed/notarized (stretch goal). Binding port 80 is allowed for non-root on recent macOS. |
| Windows | Windows Firewall prompt on first run — user must allow *Private networks*. Provide a note/screenshot. Port 80/443 may be held by other software (IIS, Skype); auto-fallback to 8080 (HTTP) / 8443 (HTTPS). The address list (startup, Setup, the self-signed certificate) comes from the network interfaces that are *up*: Windows also reports the last address of a switched-off or disconnected adapter, which must not be offered to the audience (`netinfo.LocalIPs`). |
| Linux | Port 80 needs `setcap 'cap_net_bind_service=+ep'` or run on 8080 + redirect. Optional systemd unit file in `deploy/`. |

When a computer has more than one network interface (a cable and Wi-Fi at once, virtual machines, a VPN), **Settings → Network** has a tick for each connected interface; only the ticked ones are used (`network.interfaces` in `config.yaml`, or `--network-interfaces`). The addresses of the used interfaces are what the startup log, Setup's address list and QR code, a new self-signed certificate and the `<name>.local` announcement offer, the main connection's (the address that reaches the default gateway) first. Until a choice is made every interface is used; once made, an interface that appears later stays off until it is ticked. **Two interfaces on one network** (a cable and Wi-Fi at once, both 192.168.1.x): only one address is offered for the network, so the audience, the QR code, the certificate and the `<name>.local` announcement agree (two announcements for one name on one network would replace each other in the devices' caches). The main connection is used, then a wired interface, then any other, then Wi-Fi; to choose the other, untick the first. The kind of interface comes from the operating system (Windows adapter type, `networksetup` on macOS, `/sys/class/net` on Linux), else from its name. By default this only chooses what is *offered*: the server still accepts connections on every address, so someone who types the other address still reaches it. `network.restrict_listening: yes` (also `--network-restrict`, and a box in the same tab) makes the server close connections made to an address of an interface that is not ticked: the server stays bound to every address (so a change applies without a restart, and an interface that appears later is covered) and looks at the address each connection arrived on; loopback (this computer, `localhost`) is always served so Setup stays reachable, and Setup never turns off the interface it is being used through (the tab marks it).

Ready-made service files for systemd, launchd and Windows Task Scheduler are in `deploy/`; day-to-day deployment advice is in [usage-guide.md](usage-guide.md).

Cross-compile with `GOOS/GOARCH`; CI produces release artifacts for all targets.

## Network topology

```
 [Operator laptop running open-caption] ──── ethernet/WiFi ──── [Router / AP]
                                                                     │  guest SSID
                                                      phones / tablets / projector browser
```

- **Venue prerequisite:** the guest network must allow guest devices to reach the server host (same subnet/VLAN, or an explicit rule for the server's ports). This is a requirement of the venue's guest network, not something the app can fix; verify before the first show by loading the audience URL from a phone on the guest SSID.
- The host must be reachable from the guest WiFi subnet. Many "guest network"
  features isolate guests from the *LAN*; the venue must place the server on
  the guest network itself, or allow guest→server on the server's port.
  **Client isolation between guests is fine** (and recommended) — audience
  devices only talk to the server, never to each other.
- Give the server a **static IP or DHCP reservation**.
- Wired connection for the server host is recommended; WiFi congestion with
  150 devices is the real-world risk. Recommend an AP/router rated for ≥ 100
  concurrent clients, 5 GHz enabled.
- No internet uplink is required.

## Making the URL easy to reach

Audiences should type as little as possible. In increasing effort:

1. **QR code** (REQ-053): printable page `/admin` → Run → QR for lobby signage and programs. Highest value, zero infrastructure.
2. **Short IP URL**: `http://192.168.1.10` (port 80 so no `:port`).
3. **Local DNS name**: router DNS entry, e.g. `captions.theater` → server IP (dnsmasq/OpenWrt/Pi-hole/most routers allow a host override).
4. **Send guests to the page when they join the WiFi** (a captive-portal-style redirect). Documented for venue staff in [venue-network-guide.md](venue-network-guide.md). The options, most reliable first:
   - **The access point's own splash page / guest portal** pointing at the audience URL (UniFi hotspot with an external portal, Omada portal, OpenWrt nodogsplash `RedirectURL`, etc.). This is router configuration, not part of this program, and it is the option most likely to work today.
   - **DHCP option 114 (RFC 8910) and `GET /captive-api` (RFC 8908).** Option 114 ("Captive-Portal API URI") tells a device where to ask whether it is behind a captive portal. The server implements the API endpoint `GET /captive-api`, which answers `{"captive": false, "user-portal-url": "<this server>/"}` (`Content-Type: application/captive+json`). What this does and does not do:
     - It tells compliant devices the network is **not** captive, and publishes the audience page as the network's user-portal URL. It does **not** make a phone open the captions page by itself: devices only pop up a sign-in sheet when the API says `"captive": true`, which this server never does.
     - RFC 8908 requires the API to be reached over **HTTPS with a certificate the device trusts**. A plain `http://` option 114 URI, or this server's self-signed certificate, is ignored by compliant clients. The endpoint is therefore only useful when the venue serves it with a trusted certificate (for example HTTPS enabled with a real certificate, or a reverse proxy in front).
     - Support is **experimental and untested on real devices.** Treat it as a bonus for networks that already use option 114, not as the way to get people to the page.
   - Responding to OS connectivity probes (`captive.apple.com/hotspot-detect.html`, `connectivitycheck.gstatic.com/generate_204`) via DNS override so devices open a captive browser sheet to the audience page — requires a router-side DNS rule; documented, not built into the app (no DNS server in v1).
5. **mDNS** (`open-caption.local`): the server announces itself by name using multicast DNS (RFC 6762), so devices on the same network can open `http://open-caption.local/` with no router setting and no internet. It is **on by default** under the name `open-caption`; the `.local` row of the address list on Setup's Run step turns it off or changes the name (letters, digits, hyphens), and `mdns.enabled` / `mdns.name` in `config.yaml` do the same.
   - Before claiming the name it asks the network three times whether anyone else has it; if so it uses `open-caption-2` and so on, and Setup says which name it got. It says goodbye when the program stops and announces again if the computer's address changes.
   - It is **IPv4 only**, uses UDP port 5353 (shared with the operating system's own mDNS service) and the multicast group 224.0.0.251. If multicast is not allowed, the program logs a warning and carries on without it.
   - Reliability: Apple devices, Windows 10 and later, and Linux with Avahi resolve `.local` names; recent Android versions mostly do. Many guest WiFi networks block multicast between devices, and then the name will not resolve. For that reason the QR code and numeric address stay the primary way in (the QR sign defaults to the numeric address).

The app should not run its own DHCP/DNS in v1 (conflicts, privileges, scope).

## Security posture

- LAN-only. **HTTP and HTTPS are both supported**, on separate ports, serving the same routes, and there is **no redirect** from HTTP to HTTPS. HTTP is the default audience URL because a self-signed certificate shows a browser warning. A supplied certificate is watched and a renewed one is picked up within about 30 seconds without a restart (see [doc 10](10-acme-lets-encrypt.md)). HTTPS is enabled in `config.yaml` with either `cert_file`/`key_file` or `self_signed: true` (generated into `<data>/tls/`). Screen Wake Lock needs a secure context, so audiences on HTTPS keep their screens awake more reliably; on HTTP the video-loop fallback applies.
- Automatic certificates (ACME/Let's Encrypt) are optional: with a public domain name whose DNS provider this program supports, the DNS-01 challenge gets a trusted certificate without the computer being reachable from the Internet. It needs outbound access to the CA and the DNS provider's API. See [doc 10](10-acme-lets-encrypt.md).
- `/admin` and `/operator` use separate passwords from `config.yaml` (plaintext, file mode 0600; changeable from Setup), a login form and session cookie (not HTTP Basic auth); login is rate-limited (5 tries/min per address).
- **These pages are reachable by everyone who can reach the server, and over plain HTTP passwords and the session cookie cross the WiFi unencrypted.** Mitigate with `access.admin_allow` / `access.operator_allow` (for example `admin_allow: [localhost]` so Setup can only be opened on the captioning computer, and `operator_allow` set to the operator's address or a wired network), by putting the operator on a wired or separate network, and/or by enabling HTTPS with a certificate the devices trust. Setup's Settings window shows the current rules and warns when Setup is open to the whole network.
- Forged requests (another website, or another port of this computer, making a browser send a request) are refused by an `Origin` / `Sec-Fetch-Site` check on all state-changing API calls.
- Do not expose the port to the internet. Documented.
- Uploaded files: PNG/JPG/GIF/WebP only (content-sniffed, 10 MB limit), served with `X-Content-Type-Options: nosniff`. SVG is rejected.
- Audience page CSP: `default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self'` (fonts and JavaScript are all served by the program itself).

## Backups

Data directory is plain files; "Export" and "Import" on the Shows step of Setup move a show between machines or share it with other volunteers.

### Show archive (.zip)

An export is a `.zip` with these entries, all at the top level except the images:

| Entry | Contents |
|-------|----------|
| `show.md` | The main-language show file, including its `[settings]` block (title, look, idle message, loop time, scrollback size). **Required.** |
| `show.<lang>.md` | One show file per translation, for example `show.es.md`. Optional. |
| `assets/<filename>` | The show's images (PNG, JPG, GIF or WebP). Optional; no sub-folders. |

Nothing else is stored: the show file is the whole configuration of a show, so there is no separate metadata file.

Import creates a **new** show (it never overwrites one), titled with the archive's title plus " (imported)", and does not activate it. Rules:

- Only top-level `.md` files and `assets/*` are read; any other entry is ignored. A `.md` file that is not `show.md` or `show.<valid-language-code>.md` is ignored.
- Each image goes through the normal upload checks (type sniffed from its content, 10 MB limit, SVG refused). Images that fail are skipped and counted; Setup reports how many.
- Limits: 64 MB for the upload, at most 500 entries, and each file is size-limited. A file that is not a zip, or has no `show.md`, is refused with an explanation and nothing is created.
- You can write an archive by hand: zip `show.md` (and any translations and an `assets/` folder) with any zip tool.
