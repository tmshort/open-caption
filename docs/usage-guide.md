# open-caption usage guide

open-caption shows live captions on audience phones, tablets and screens during a play or musical. A volunteer **operator** follows the show and presses the spacebar for each line; every device on the venue's WiFi shows the line a moment later. Nothing is installed on the audience's devices, and they never have to touch anything.

This guide covers what to buy and how to arrange it ([Deployment](#deployment-recommendations)), how to set up a show, how to run it, and a day-of checklist. The in-app **Help** button covers the same ground in smaller pieces, and [venue-network-guide.md](venue-network-guide.md) is written for whoever manages the building's WiFi.

Contents: [The pieces](#the-pieces) · [Deployment recommendations](#deployment-recommendations) · [Installing and starting](#installing-and-starting) · [Setting up a show](#setting-up-a-show) · [Running the show](#running-the-show) · [Day-of checklist](#day-of-checklist) · [Troubleshooting](#troubleshooting) · [Maintenance](#maintenance) · [Operator quick card](#operator-quick-card)

## The pieces

| Piece | What it is |
|-------|------------|
| **The server** | The `open-caption` program, running on one computer (macOS, Windows or Linux). It holds the script and sends lines to everyone. |
| **Setup page** (`/admin`) | Where volunteers prepare a show: script, images, look, rehearsal. Admin password. |
| **Operator page** (`/operator`) | Where the operator follows the show and sends lines. Operator password. |
| **Audience page** (`/`) | What audiences open on their devices. No password. A browser on a projector or large monitor can show the same page for the whole room. |

Everything runs on one computer. The audience, the operator and the setup volunteers all use ordinary web browsers.

## Deployment recommendations

### The short version

1. **Use a dedicated WiFi access point for captions** unless the venue's guest network is known to work (see below). It removes the biggest risk: a network you do not control.
2. **Put the server on a wired connection** to that access point, with a reserved address.
3. **Use an access point built for crowds** (rated for 100+ devices) placed in the audience area.
4. **Keep the server computer plugged in, awake, and otherwise idle** during the show.
5. **Print the QR code** and put it in the program and in the lobby.
6. **Rehearse with a few real phones** in the real room before opening night.

### Network options

| Option | When to use it | Notes |
|--------|----------------|-------|
| **A. Dedicated access point (recommended)** | Any venue, especially if you cannot change the building's network | You bring a router or access point. Audiences join its WiFi (for example "Theater Captions"). No internet connection is needed. You control capacity, reservations and firewalls. |
| **B. The venue's guest WiFi** | The venue's network staff will confirm guest devices can reach your computer | Cheapest. Ask for a reserved address for your computer and for guest-to-server traffic to be allowed. See [venue-network-guide.md](venue-network-guide.md). |
| **C. Your laptop's own hotspot** | Rehearsals and tiny audiences only | Laptop hotspots usually handle only a handful of devices. Not for a full house. |

### Option A in detail (dedicated access point)

```
                 ┌───────────────────────┐
                 │  Access point/router  │  WiFi "Theater Captions"
                 │  (in the audience     │ ─ ─ ─ ─ ─ ─ ─ ─  phones / tablets
                 │   area, 5 GHz on)     │                  projector browser
                 └───────────┬───────────┘
                    Ethernet │
                 ┌───────────┴───────────┐
                 │ Captioning computer   │   ← operator uses this screen,
                 │  (open-caption)       │     or a second device on the same WiFi
                 └───────────────────────┘
```

- **Equipment:** a business-grade or prosumer access point or router designed for many simultaneous devices (look for 100+ clients in its specifications and for 5 GHz support), with an Ethernet port. Cheap home routers often struggle past 30–50 devices. For a ~150-seat house, plan for one capable access point, and a second one for a deep or wide room.
- **Settings:**
  - Create one SSID with a short name and a simple, printable WPA2 passphrase, or no passphrase if the venue allows an open network. Print it on the QR sign.
  - Turn **5 GHz** on, and prefer it. Use 20 or 40 MHz channel widths in a crowded building.
  - Leave **client isolation** on if it is offered. Audience devices only talk to the server, never to each other.
  - Give the captioning computer a **DHCP reservation** (or a fixed address) so the address never changes, and print the address on the sign.
  - You do not need an internet connection, and the program never uses one.
- **Placement:** put the access point in the audience area, high and clear of walls, not in a back room or the lobby. Do not put it behind the stage or lighting rack. A laptop on a distant table is fine: it is wired.
- **Power:** use a surge protector. A small UPS on the access point and server removes the risk of a tripped plug.

### Option B in detail (the venue's guest WiFi)

The program does not care which network it is on, as long as audience devices can reach it. Before relying on it:

- Ask for **guest-to-server access**: your computer must be reachable from the guest SSID, on TCP port 80 (and 443 if you enable HTTPS). "Client isolation" between guests is fine; blocking guests from reaching *your computer* is not.
- Ask for a **reserved address** for your computer's network adapter.
- Plug the computer into the network by **Ethernet** if at all possible.
- Test with a phone on the guest WiFi well before the show (the Go Live step shows the address).
- Have Option A in a bag as a backup if the show matters.

### The captioning computer

- **Any ordinary laptop or mini-PC** will do. The program uses very little memory (tens of megabytes) and almost no processor. Use a computer you can keep plugged in.
- **Operating system:** macOS, Windows or Linux. Use the same computer for all shows if you can, so the passwords, shows and address stay the same.
- **Make it boring:** set the computer to *never sleep* while plugged in, disable automatic restarts for updates around show dates, close other programs, and keep the lid open or set the lid not to sleep.
- **Firewall:** allow the program on **private networks** (Windows asks the first time; on macOS allow incoming connections for it).
- **Address:** wired and reserved (see above). If the address changes, printed QR codes stop working.
- **Time:** the clock does not matter.

### Who sits where

- **The operator** needs a clear view of the stage and a reliable device. The simplest arrangement is the captioning computer itself at the tech desk. The operator page also works from any other laptop, tablet or phone on the same network (`http://<computer's address>/operator`), which lets the operator sit where they can see the stage. A Bluetooth clicker or keyboard is easier than a touch screen.
- **Only one operator controls the show at a time.** Anyone else who opens the operator page can watch, and can use **Take over** if the operator hands over.
- **The setup volunteers** use the Setup page from any computer on the network before the show, or from the captioning computer.
- **A shared screen** for the whole room: open the audience address on a computer connected to a projector or a large monitor, add `?display=projector` to the address for large text and no on-screen clutter, and make the browser full screen (`F11`, or run Chrome with `--kiosk`).

### Running the server

For most theaters, the simplest way is best: start the program by hand before each show, leave its window open, and press `Ctrl+C` (or close it) afterwards. The show position is saved continuously, so restarting in the middle of a show is safe (see [Troubleshooting](#troubleshooting)).

If the computer should start the server by itself after a reboot, the [`deploy/`](../deploy/) folder has ready-made files for Linux (systemd), macOS (launchd) and Windows (Task Scheduler). Each begins with its own install steps.

### Getting people to the page

In order of reliability:

1. **The QR code and sign** (Setup → Go Live → *Print a QR sign for the lobby*). Put it on the wall, in the program and on the seat-back card. This is what to rely on.
2. **A short address.** On port 80 there is no `:port` to type. A name such as `captions.theater` (a local DNS entry on the router) is easier to remember than numbers.
3. **The access point's splash page** (a "guest portal" or "captive portal" that redirects to your address after joining). This is configured on the router or access point, not in this program. See [venue-network-guide.md](venue-network-guide.md).
4. **DHCP option 114 and `/captive-api` (advanced, experimental).** Option 114 is a DHCP setting that points devices at a "captive portal API". The program answers that API at `/captive-api` (for example `https://<address>/captive-api`) with "this network is not captive" and the page's address. Two cautions: it **does not make phones open the captions page by themselves**, and the standard (RFC 8910/8908) requires **HTTPS with a certificate the devices trust**, so a plain `http://` setting or the program's self-signed certificate is ignored. Only set it up if the venue already uses option 114 and can provide a trusted certificate. It has not been tested on real devices.

### Ports and addresses

- By default the server listens on **port 80**, so the audience address needs no port number, for example `http://192.168.1.10/`. If something else is using port 80, or the operating system does not allow it, the server tells you and uses **port 8080** instead; the audience address then ends in `:8080`. On Linux, port 80 needs permission (`sudo setcap 'cap_net_bind_service=+ep' ./open-caption`) unless you use the systemd file, which grants it.
- You can choose ports yourself with `--http-addr :8081`, or in `config.yaml`.
- **HTTPS is optional.** Plain `http://` works on every device without warnings and is the recommended audience address. A self-signed HTTPS address makes every device show a security warning. Turn HTTPS on only if you can install a certificate your devices trust.

### Security

- The audience page can only read. It cannot change the show or reach the setup and operator pages.
- Setup and operator each require their own password (stored in plain text in `config.yaml`, which is created readable only by you). Give the operator password to the operator, not the admin password.
- The program is meant for a private network. **Do not expose it to the internet.**
- Passwords can be changed by editing `config.yaml` and restarting the program. Changing a password signs everyone out. Otherwise logins last 12 hours and survive a restart, so an operator does not have to sign in again if the program is restarted mid-show.

## Installing and starting

1. **Get the program.** Download the build for your computer (or run `make dist` to build them all). Put it in a folder you will keep, for example `Documents/open-caption`.
2. **First start.** Run it: double-click it on Windows, or in a terminal run `./open-caption` (macOS/Linux; on macOS the first run may need *right-click → Open* because the program is not signed).
3. **Note the passwords.** The first time, the program prints the **admin** and **operator** passwords and where `config.yaml` is. They stay in that file.
4. The setup page opens in your browser. If it does not, open `http://localhost/admin` (or `http://localhost:8080/admin` if the console says it used port 8080).
5. **Choose where data lives (optional).** By default, shows and settings are in your user configuration folder (macOS `~/Library/Application Support/open-caption`, Linux `~/.config/open-caption`, Windows `%AppData%\open-caption`). To keep everything in a folder you choose, start with `--data <folder>`. Always use the same folder.

Useful options: `--data <folder>`, `--http-addr :8080`, `--https-addr :443`, `--no-browser`, `--log-file <file>`, `--version`.

## Setting up a show

Open the setup page and follow the six steps along the top. The **Help** button explains each one.

1. **Show.** Create a show (or the sample show, to try things out).
2. **Script.** Paste the script or drop in a `.txt` or `.md` file. Each line becomes one caption. Fix anything listed under *Problems and notes*; click one to jump to its line. Save.
3. **Images.** Upload the pictures the script uses (PNG, JPG, GIF or WebP, ideally under 1 MB each).
4. **Look.** Choose colors and text size. The defaults (white on black, large) are best for most people.
5. **Test.** Run a private rehearsal: you press Space as the operator would, and see the audience screen next to it. Nothing reaches the audience.
6. **Go Live.** Check the readiness list, print the QR sign, and press **Start show**.

### Writing the script

A script is plain text. The short version:

```text
<!-- Comments like this are never shown. -->
[loop: 10s]
[image: welcome.png | 15s]
Welcome! Please silence your phones.
[/loop]

[section: Act 1, Scene 1]
[note: lights come up slowly]
[clear]
(A garden gate creaks open.)
MARGARET: You're late, Edwin.
EDWIN: The *train* was late.

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[/song]
~ Tea for two, and two for tea ~

[fullscreen: **INTERMISSION**]
```

- **One line = one caption** = one press of Space. You decide where to break long speeches.
- **Speaker names** in capitals followed by a colon are shown above the line.
- **Stage directions** in `(parentheses)` are shown in italics.
- **Lyrics** between `♪` notes, in italics: start a line with `~`, or wrap the song in `[song]`…`[/song]`. For two singers at the same time (a duet or counterpoint), put their lines between `[together]` and `[/together]`, or join them with ` || ` on one line; both show at once on one press of Space. The editor's **♪**, **♫** and **🎵 Song** buttons insert these. See the *Musicals* help topic.
- **Loops** (pre-show and intermission) repeat pictures and text until the operator moves on. A loop at the very start of the script starts by itself.
- **Notes** `[note: …]` and bookmarks `[section: …]` are for the operator only.
- The **Writing a script** help topic lists everything, with examples, and the editor's *Script cheat sheet* summarizes it.

### Changing things during rehearsals and shows

You can edit and re-save a script while a show is live (for example to fix a typo). The program asks you to type the show's name to be sure, keeps the operator's place by line number, and warns if the number of lines changed.

## Running the show

### Before the audience arrives

1. Start the server and open the **Operator page**; sign in. It should say **You are in control**.
2. In Setup → Go Live press **Start show** (once; this resets the show to the beginning). A pre-show loop, if the script has one, begins on every screen.
3. Open the audience address on your own phone and, if you have one, on the projector. Check the viewer count on the operator page goes up.
4. Put the QR sign where people can see it.

### During the show

- **Space** sends the next line (also Right arrow, Page Down, Enter, and most clickers).
- **Backspace** undoes the last thing you sent.
- **B** blanks the audience screens (hides everything; the same captions return when you press **B** again). **C** clears them (erases the captions, so the next line starts on an empty screen). **Backspace** undoes either.
- **Shift+Q** stops an intermission or pre-show loop (blank screen). **Space** continues.
- **Message…** puts a one-off notice on every screen ("The show will resume shortly"). **Back** removes it.
- Click a line, then confirm, to jump anywhere. **/** searches the script.
- The line with the yellow arrow is *next*. The green dot marks what is on screen now.

### When the show ends

Press `Ctrl+C` in the program's window (or stop the service). The position is saved. To run the show again from the beginning, use **Restart show…** on the operator page, or press **Start show** in Setup.

## Day-of checklist

**A few days before**
- [ ] Rehearse with the real script and a few real phones on the real network in the real room.
- [ ] Export the show (Setup → Show → **Export**) and keep the `.zip` somewhere safe, such as a USB stick.
- [ ] Confirm the captioning computer's address is reserved and the QR sign shows it.

**At the venue, an hour before**
- [ ] Power: computer and access point plugged in; computer set to never sleep.
- [ ] Access point on, in place, with the right name; the computer wired to it.
- [ ] Start the program; note the address printed (and in Go Live).
- [ ] A phone on the audience WiFi can open the address and shows "Captions will begin shortly".
- [ ] Setup → Test: do a quick rehearsal; operator keys work.
- [ ] Projector browser (if used) open on `…/?display=projector`, full screen.
- [ ] Press **Start show**. Pre-show loop (if any) is on every screen.
- [ ] Operator sitting where they can see the stage; clicker or keyboard tested.
- [ ] QR sign placed; a volunteer knows the WiFi name and address for people who ask.

**After the show**
- [ ] `Ctrl+C` in the program window.
- [ ] Export the show again if you changed anything.

## Troubleshooting

| Problem | What to try |
|---------|-------------|
| Nobody can open the address | Is the phone on the right WiFi (not mobile data)? Does the address include `:8080` if the program used that port? Is guest-to-server traffic allowed? Is the computer's firewall allowing the program on private networks? See [venue-network-guide.md](venue-network-guide.md). |
| Some devices connect and others do not | The access point may be at its device limit. Add a second access point, or move the first into the room. |
| Captions are late or stall on a few phones | Weak signal or a busy channel. Those phones reconnect on their own. Improve access-point placement; use 5 GHz. |
| Screens go dark on phones | Phones lock their screens. The page asks the browser to stay awake where it can, but plain-HTTP pages are often refused. Tell people to switch off Auto-Lock for the evening, or use a projector. |
| The wrong line was sent | **Backspace** undoes it. |
| The operator page says another operator is in control | Use **Take over**, if you are the operator. |
| The computer froze or was restarted mid-show | Start the program again. The operator page reconnects, stays signed in and takes control back by itself, the show is where you left it, and audience devices reconnect on their own. A loop restarts from its first item, and **Backspace** cannot undo anything sent before the restart. |
| I forgot a password | Open `config.yaml` in the data folder in any text editor. |
| Something else | Setup → Go Live → **Download a support bundle**; it holds the settings (without passwords) and recent log lines. |

## Maintenance

- **Backups:** *Export* in Setup → Show saves a show (script, images, look) as a `.zip`. *Import* restores it, on this or another computer. You can also copy the whole data folder.
- **Moving to another computer:** install the program, import the exported shows, set the new computer's address and, if you want to keep the old passwords, copy `config.yaml`.
- **Updating the program:** stop it (`Ctrl+C`), replace the program file with the newer one, and start it again with the same `--data` folder. Do not update right before a show.
- **Reusing a show:** Setup → Show has **Rename…** and **Duplicate**. Duplicating makes a copy of the script, images and look to adapt for the next run, and leaves the original alone.
- **Cleaning up:** delete shows you no longer need from Setup → Show. Exported `.zip` files are your archive.
- **Logs:** recent log lines are in the support bundle; add `--log-file` to keep a file.

## Operator quick card

Print this and keep it by the operator.

| Key | Does |
|-----|------|
| **Space** | Send the next line |
| **Backspace** | Undo the last thing sent |
| **Shift+Q** | Stop a loop (blank screen) |
| **B** | Blank: hide everything, captions come back on unblank |
| **C** | Clear: erase the captions for a fresh start |
| **/** | Search the script |
| **?** | Help (keys are paused while it is open) |

- Yellow arrow = next line. Green dot = on screen now.
- Wrong line? **Backspace.** Lost your place? Click the line you want, then confirm.
- Need to say something to the audience? **Message…**
- Show interrupted? **B** to blank the screens, **Message…** to explain.
