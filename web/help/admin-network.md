# Network and joining

Audience members connect their own phones and tablets to the venue's **guest WiFi** and open one address. Nothing needs to be installed.

## What the venue must provide

- This computer and the audience's devices must be on the same network, or the network must allow guest devices to reach this computer. This is a setting of the venue's WiFi system, not of this program.
- Many guest networks isolate devices from each other. That is fine. Devices only talk to this computer, never to each other, as long as this computer is reachable.
- A wired connection for this computer is best. Ask for a reserved address (a "DHCP reservation") so the address does not change between shows.
- No internet connection is needed.

Test it before the first show: join the guest WiFi on a phone and open the address shown in the **Run** step. For a projector or a large screen, the same step lists a projector link: open it on the computer the projector is plugged into.

## Making the address easy

1. **QR code** — print the sign from the Run step and put it in the lobby and the program. This is the easiest option.
2. **A short address** — if the program can use port 80, the address has no port number, such as `http://192.168.1.10/`.
3. **A name** — the program announces itself as **`open-caption.local`**, so `http://open-caption.local/` works on most phones and computers on the same network, with nothing to set up. Change the name, or switch it off, on the **Run** step, in the row of the `.local` address (**Rename…** and its checkbox). Some guest WiFi networks block this kind of announcement; then use the numeric address or the QR code. A venue can also add a name of its own (such as `captions.theater`) on its router, which works everywhere.
4. **Automatic page** — a venue can send guests to the page when they join the WiFi with the access point's splash page ("captive portal"). The program also answers the standard network-detection address `/captive-api` for venues that use DHCP option 114, but that only reports that the network is not captive: it does not open the page by itself, and it needs a trusted HTTPS certificate. See the [venue network guide](topic:admin-venue-network-guide).

## If this computer has more than one address

A computer with a cable and WiFi at once, or with virtual machines or a VPN, has several addresses. The Run step lists the ones of adapters that are connected, with the main connection first; that is the one the QR code uses at first. To give the audience only some of them (for example only the venue's WiFi network), use **Choose addresses…** next to the QR sign button: **Show only these** (empty means all) and **Never show these**, one per line, each an address (`192.168.1.50`), a network (`192.168.1.0/24`) or the word `private`. The change applies at once. If two adapters are on the same network (a cable and WiFi, both with a 192.168.1.x address), only one is used, so the QR code and the `.local` name agree: the main connection, else the wired one. To choose, press **Use this one** beside the other row, or write the adapter's name (or Ethernet or Wi-Fi) in **Prefer these adapters**. Addresses that are left out are not listed, not in the QR code, not announced as the `.local` name, and not in a new self-signed HTTPS certificate. By default this only chooses what is offered: a phone that already has a hidden address can still connect. Tick **Also refuse connections to the addresses that are not shown** to turn those away too; connections to this computer itself are always accepted, so Setup stays reachable. Setup never lets you turn off the address you are using it through: the window marks that address.

## HTTP and HTTPS

Plain `http://` addresses work on every device and are the default. HTTPS can be turned on in `config.yaml`; with a self-signed certificate each device shows a security warning the first time. Plain HTTP is simpler for an audience.

## If the port is busy

If port 80 is already used by another program, the server starts on a different port and says so in the console. Audience addresses must then include it, for example `http://192.168.1.10:8080/`; the Run step shows the right address.
