# Network and joining

Audience members connect their own phones and tablets to the venue's **guest WiFi** and open one address. Nothing needs to be installed.

## What the venue must provide

- This computer and the audience's devices must be on the same network, or the network must allow guest devices to reach this computer. This is a setting of the venue's WiFi system, not of this program.
- Many guest networks isolate devices from each other. That is fine. Devices only talk to this computer, never to each other, as long as this computer is reachable.
- A wired connection for this computer is best. Ask for a reserved address (a "DHCP reservation") so the address does not change between shows.
- No internet connection is needed.

Test it before the first show: join the guest WiFi on a phone and open the address shown in the **Go Live** step.

## Making the address easy

1. **QR code** — print the sign from the Go Live step and put it in the lobby and the program. This is the easiest option.
2. **A short address** — if the program can use port 80, the address has no port number, such as `http://192.168.1.10/`.
3. **A name** — ask the venue to add a local name such as `captions.theater` for this computer's address.
4. **Automatic page** — a venue can send guests to the page when they join the WiFi (a "captive portal" or DHCP option 114). See the venue network guide in the project's `docs` folder.

## HTTP and HTTPS

Plain `http://` addresses work on every device and are the default. HTTPS can be turned on in `config.yaml`; with a self-signed certificate each device shows a security warning the first time. Plain HTTP is simpler for an audience.

## If the port is busy

If port 80 is already used by another program, the server starts on a different port and says so in the console. Audience addresses must then include it, for example `http://192.168.1.10:8080/`; the Go Live step shows the right address.
