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
3. **A name** — the program announces itself as **`open-caption.local`**, so `http://open-caption.local/` works on most phones and computers on the same network, with nothing to set up. Change the name, or switch it off, in **Settings** (button at the top), **Network** tab, under **Name on the network** (the **Use a name** box and the name). Some guest WiFi networks block this kind of announcement; then use the numeric address or the QR code. A venue can also add a name of its own (such as `captions.theater`) on its router, which works everywhere.
4. **Automatic page** — a venue can send guests to the page when they join the WiFi with the access point's splash page ("captive portal"). The program also answers the standard network-detection address `/captive-api` for venues that use DHCP option 114, but that only reports that the network is not captive: it does not open the page by itself, and it needs a trusted HTTPS certificate. See the [venue network guide](topic:admin-venue-network-guide).

## If this computer has more than one address

A computer with a cable and WiFi at once, or with virtual machines or a VPN, has several network interfaces. Open **Settings** (button at the top) and its **Network** tab: it lists each connected interface with a tick. Only the ticked ones are used, and the audience is given only their addresses (in the address list, the QR code and the `.local` name). Tick just the networks the audience uses. If two ticked interfaces are on the same network (a cable and WiFi, both with a 192.168.1.x address), only one address is offered, so the QR code and the `.local` name agree: the wired one, else the main connection. To use the other, untick the first. A tick is saved when you click it, and the interface you are using Setup through cannot be unticked. A new interface (a USB adapter, a VPN) stays off until you tick it. **Also refuse connections to interfaces that are not ticked** goes further: a phone that uses an address of an unticked interface is turned away; connections to this computer itself are always accepted, so Setup stays reachable.

## HTTP and HTTPS

Plain `http://` addresses work on every device and are the default. HTTPS can be turned on in `config.yaml`; with a self-signed certificate each device shows a security warning the first time. Plain HTTP is simpler for an audience.

## If the port is busy

If port 80 is already used by another program, the server starts on a different port and says so in the console. Audience addresses must then include it, for example `http://192.168.1.10:8080/`; the Run step shows the right address.
