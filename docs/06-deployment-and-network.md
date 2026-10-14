# 06 — Deployment and Network

## Running

```
open-caption                       # defaults
open-caption --addr :80            # listen address (default: :80, falls back to :8080 with a warning if bind fails)
open-caption --data ~/oc-data      # data directory
open-caption --no-browser          # don't auto-open the setup page
```
Environment variable equivalents: `OPEN_CAPTION_ADDR`, `OPEN_CAPTION_DATA`.

On start, the server logs and shows (in the console and setup UI) all
non-loopback IPv4 addresses and the audience URL, and opens
`http://localhost:<port>/admin` in the default browser (macOS `open`, Windows
`rundll32 url.dll,FileProtocolHandler`, Linux `xdg-open`).

## Platforms

| OS | Notes |
|----|-------|
| macOS | Unsigned binary requires Gatekeeper approval (right-click → Open) unless signed/notarized (stretch goal). Binding port 80 is allowed for non-root on recent macOS. |
| Windows | Windows Firewall prompt on first run — user must allow *Private networks*. Provide a note/screenshot. Port 80 may be held by other software (IIS, Skype); auto-fallback to 8080. |
| Linux | Port 80 needs `setcap 'cap_net_bind_service=+ep'` or run on 8080 + redirect. Optional systemd unit file in `deploy/`. |

Cross-compile with `GOOS/GOARCH`; CI produces release artifacts for all targets.

## Network topology

```
 [Operator laptop running open-caption] ──── ethernet/WiFi ──── [Router / AP]
                                                                     │  guest SSID
                                                      phones / tablets / projector browser
```

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

1. **QR code** (REQ-053): printable page `/admin/qr` for lobby signage and programs. Highest value, zero infrastructure.
2. **Short IP URL**: `http://192.168.1.10` (port 80 so no `:port`).
3. **Local DNS name**: router DNS entry, e.g. `captions.theater` → server IP (dnsmasq/OpenWrt/Pi-hole/most routers allow a host override).
4. **Captive-portal-style redirect**: the router redirects guests to the page after joining WiFi. Options, documented in a `docs/venue-network-guide.md` (to be written):
   - Router/AP's built-in splash page / captive portal pointing at the audience URL (UniFi, Ubiquiti, Omada, OpenWrt nodogsplash).
   - **DHCP option 114 (RFC 8910, Captive-Portal API URI)** set to the audience URL: modern iOS/Android auto-prompt a sign-in popup. Note: the OS expects a captive-portal JSON API at that URI; the server therefore implements a minimal RFC 8908 endpoint `GET /captive-api` returning `{"captive": false, "user-portal-url": "http://<host>/"}` so devices open the page without complaint.
   - Responding to OS connectivity probes (`captive.apple.com/hotspot-detect.html`, `connectivitycheck.gstatic.com/generate_204`) via DNS override so devices open a captive browser sheet to the audience page — requires a router-side DNS rule; documented, not built into the app (no DNS server in v1).
5. **mDNS** (`open-caption.local`): optional advertise via a small pure-Go mDNS responder (stretch goal; unreliable on some guest networks that block multicast).

The app should not run its own DHCP/DNS in v1 (conflicts, privileges, scope).

## Security posture

- LAN-only; HTTP (no TLS) by default. TLS is not required for captions and a self-signed cert would produce scary browser warnings for the audience. Wake Lock API requires a **secure context** on some browsers — see open question Q3 in doc 09; mitigation is the video fallback.
- Setup/operator behind PIN; rate-limit login (5 tries/min).
- Do not expose the port to the internet. Documented.
- Uploaded files: type-sniffed, size limited, served with `X-Content-Type-Options: nosniff`; SVG served with `Content-Security-Policy: sandbox` or rasterized/rejected (decision in doc 09).
- Audience page CSP: `default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'`.

## Backups

Data directory is plain files; "Export show" (zip of script + assets) and "Import show" in setup for moving between machines / sharing with other volunteers.
