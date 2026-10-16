# Venue network guide

For whoever looks after the theater's WiFi. The captioning computer needs nothing from the internet. It only needs to be **reachable from the audience's devices**.

## What the audience does

1. Join the guest WiFi.
2. Open one address, for example `http://192.168.1.10/` (or scan the QR code).

## Requirements

| Requirement | Why |
|-------------|-----|
| The guest network can reach the captioning computer | Devices fetch the page and keep one connection open to it |
| The computer has a fixed address (DHCP reservation or static) | So printed QR codes and signs stay valid between shows |
| The access point handles 100+ simultaneous devices (5 GHz enabled) | A full house is about 150 devices |
| No internet uplink is needed | Everything is served by the captioning computer |

**Client isolation between guests is fine and recommended.** Audience devices talk only to the captioning computer, never to each other. What must *not* be blocked is guest-to-server traffic. If the guest network is isolated from the main LAN, either plug the captioning computer into the guest network (a spare guest-VLAN port or the guest SSID) or add a firewall rule allowing guests to reach the computer's address on TCP port 80 (and 443 if HTTPS is enabled).

## Check before the show

From a phone on the guest WiFi, open the address shown in Setup → **Run**. If the page loads and says "Captions will begin shortly" (or shows the pre-show), you are done. If not, see "Nothing loads" below.

## Making the address easier

In order of effort:

1. **QR code** — Setup → Run → "Print a QR sign". Put it in the lobby and the program. No network changes needed.
2. **Port 80** — if the program runs on port 80, the address has no `:port`. On Linux, allow it with `sudo setcap 'cap_net_bind_service=+ep' ./open-caption`.
3. **A local name** — add a DNS host entry on the router for the captioning computer, for example `captions.theater` → `192.168.1.10`. Most routers (OpenWrt, Pi-hole, UniFi, Omada, dnsmasq-based) allow a "local DNS record" or "host override".
4. **Send guests there automatically.** Options:
   - **Splash page / captive portal** on the access point pointing at the audience address. The AP's built-in "guest portal" feature works with an external URL (UniFi: Hotspot → external portal; OpenWrt: nodogsplash `RedirectURL`). Make sure it lets the page load without a login.
   - **DHCP option 114 (RFC 8910) and `/captive-api`.** The program answers `GET /captive-api` the way RFC 8908 describes: `{"captive": false, "user-portal-url": "http://<server-address>/"}`. If your network already hands out option 114 ("Captive-Portal API URI"), you can point it at that address. Be aware of what it does: it tells devices the network is *not* a captive portal and publishes the page's address. It does **not** make phones open the captions page by themselves (that needs `"captive": true`, which the program never sends). The RFC also requires the API address to be **https with a certificate the devices trust**, so a plain `http://` option 114 address or the program's self-signed certificate will be ignored. This is experimental and untested on real devices: use the QR sign, or your access point's splash page, to get people to the page.
5. **mDNS** (`open-caption.local`) is provided and on by default: devices that can use it open `http://open-caption.local/` with nothing set up on the router. It needs the guest network to allow **multicast** (UDP 5353 to 224.0.0.251) between the captioning computer and the audience's devices. Many guest networks block that ("multicast filtering" or "mDNS/Bonjour blocking" on the access point), and then the name simply does not resolve; the numeric address and the QR code still work. If you want the name to work, enable mDNS/Bonjour forwarding or turn off multicast filtering for the guest network. The name can be changed, or switched off, in Setup → Settings → *Network name*.

The program does not run its own DHCP or DNS server, on purpose: it would clash with the venue's router.

## HTTP and HTTPS

Use plain `http://` for the audience. A self-signed HTTPS certificate makes every device show a security warning. HTTPS is available (`https.enabled: true` in `config.yaml`) for sites that can install a certificate their devices trust.

One small cost of plain HTTP: browsers only allow "keep the screen awake" on secure pages. On HTTP the page keeps a phone's screen on with a tiny, silent, invisible looping video instead. A phone in low-power mode may refuse it; the page then shows a one-time tip asking people to turn off auto-lock.

## Security

- Setup (`/admin`) and the operator page (`/operator`) each need a password from `config.yaml`. The audience page needs none and cannot change anything.
- Do not expose the port to the internet.
- The operator/admin passwords are stored in plain text in `config.yaml` in the program's data folder. Keep that folder private to the people who run the show.

## Firewalls on the captioning computer

- **Windows:** when the program first starts, allow it on **Private networks**. If the prompt was dismissed: Windows Security → Firewall → Allow an app.
- **macOS:** allow incoming connections for the program when asked, or add it in System Settings → Network → Firewall.
- **Linux:** open the port, for example `sudo ufw allow 80/tcp`.

## Nothing loads

1. Is the phone on the guest WiFi (not mobile data)?
2. Does the address include the port if the program is not on port 80? The Run step shows the correct address.
3. Can another computer on the guest WiFi reach it (`curl http://192.168.1.10/healthz` should print `ok`)?
4. Is guest-to-server traffic blocked by the router or access point (client isolation set to "block all" instead of "block between clients")?
5. Is the computer's firewall blocking the program?
6. Is the computer still awake, and still on the same address? Check for a changed address in Run.

Setup → Run → **Download a support bundle** collects the settings (without passwords), recent log lines and basic facts about the computer.

## Capacity

The program has been load-tested with 200 and 500 simultaneous audience connections on one computer (latency under 10 ms on the same machine). In a real venue the limit is the WiFi, not the program: use an access point rated for the crowd, enable 5 GHz, and keep pictures small (under 1 MB) so many devices can fetch them at once.
