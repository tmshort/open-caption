# 06 — Deployment and Network

## Running

```
open-caption                       # defaults
open-caption --config ~/oc/config.yaml   # config file (default: <data dir>/config.yaml)
open-caption --data ~/oc-data            # data directory
open-caption --http-addr :80 --https-addr :443   # override listen addresses
open-caption --no-browser                # don't auto-open the setup page
```
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
| Windows | Windows Firewall prompt on first run — user must allow *Private networks*. Provide a note/screenshot. Port 80/443 may be held by other software (IIS, Skype); auto-fallback to 8080 (HTTP) / 8443 (HTTPS). |
| Linux | Port 80 needs `setcap 'cap_net_bind_service=+ep'` or run on 8080 + redirect. Optional systemd unit file in `deploy/`. |

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

1. **QR code** (REQ-053): printable page `/admin` → Go Live → QR for lobby signage and programs. Highest value, zero infrastructure.
2. **Short IP URL**: `http://192.168.1.10` (port 80 so no `:port`).
3. **Local DNS name**: router DNS entry, e.g. `captions.theater` → server IP (dnsmasq/OpenWrt/Pi-hole/most routers allow a host override).
4. **Send guests to the page when they join the WiFi** (a captive-portal-style redirect). Documented for venue staff in [venue-network-guide.md](venue-network-guide.md). The options, most reliable first:
   - **The access point's own splash page / guest portal** pointing at the audience URL (UniFi hotspot with an external portal, Omada portal, OpenWrt nodogsplash `RedirectURL`, etc.). This is router configuration, not part of this program, and it is the option most likely to work today.
   - **DHCP option 114 (RFC 8910) and `GET /captive-api` (RFC 8908).** Option 114 ("Captive-Portal API URI") tells a device where to ask whether it is behind a captive portal. The server implements the API endpoint `GET /captive-api`, which answers `{"captive": false, "user-portal-url": "<this server>/"}` (`Content-Type: application/captive+json`). What this does and does not do:
     - It tells compliant devices the network is **not** captive, and publishes the audience page as the network's user-portal URL. It does **not** make a phone open the captions page by itself: devices only pop up a sign-in sheet when the API says `"captive": true`, which this server never does.
     - RFC 8908 requires the API to be reached over **HTTPS with a certificate the device trusts**. A plain `http://` option 114 URI, or this server's self-signed certificate, is ignored by compliant clients. The endpoint is therefore only useful when the venue serves it with a trusted certificate (for example HTTPS enabled with a real certificate, or a reverse proxy in front).
     - Support is **experimental and untested on real devices.** Treat it as a bonus for networks that already use option 114, not as the way to get people to the page.
   - Responding to OS connectivity probes (`captive.apple.com/hotspot-detect.html`, `connectivitycheck.gstatic.com/generate_204`) via DNS override so devices open a captive browser sheet to the audience page — requires a router-side DNS rule; documented, not built into the app (no DNS server in v1).
5. **mDNS** (`open-caption.local`): optional advertise via a small pure-Go mDNS responder (stretch goal; unreliable on some guest networks that block multicast).

The app should not run its own DHCP/DNS in v1 (conflicts, privileges, scope).

## Security posture

- LAN-only. **HTTP and HTTPS are both supported**, on separate ports, serving the same routes, and there is **no redirect** from HTTP to HTTPS. HTTP is the default audience URL because a self-signed certificate shows a browser warning. HTTPS is enabled in `config.yaml` with either `cert_file`/`key_file` or `self_signed: true` (generated into `<data>/tls/`). Screen Wake Lock needs a secure context, so audiences on HTTPS keep their screens awake more reliably; on HTTP the video-loop fallback applies.
- ACME/Let's Encrypt is a possible later feature. It needs Internet access and a publicly resolvable name for validation, which a venue-only network doesn't have (DNS-01 with a real domain is the realistic route).
- `/admin` and `/operator` use separate passwords from `config.yaml` (plaintext, file mode 0600; changeable from Setup), a login form and session cookie (not HTTP Basic auth); login is rate-limited (5 tries/min per address).
- **These pages are reachable by everyone who can reach the server, and over plain HTTP passwords and the session cookie cross the WiFi unencrypted.** Mitigate with `access.admin_allow` / `access.operator_allow` (for example `admin_allow: [localhost]` so Setup can only be opened on the captioning computer, and `operator_allow` set to the operator's address or a wired network), by putting the operator on a wired or separate network, and/or by enabling HTTPS with a certificate the devices trust. Setup's Settings window shows the current rules and warns when Setup is open to the whole network.
- Forged requests (another website, or another port of this computer, making a browser send a request) are refused by an `Origin` / `Sec-Fetch-Site` check on all state-changing API calls.
- Do not expose the port to the internet. Documented.
- Uploaded files: PNG/JPG/GIF/WebP only (content-sniffed, 10 MB limit), served with `X-Content-Type-Options: nosniff`. SVG is rejected.
- Audience page CSP: `default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self'` (fonts and JavaScript are all served by the program itself).

## Backups

Data directory is plain files; "Export show" (zip of script + assets) and "Import show" in setup for moving between machines / sharing with other volunteers.
