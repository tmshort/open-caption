# What open-caption can do

The full list. The [README](../README.md) has the short version; how to use each thing is in the [usage guide](usage-guide.md) and in the app's **Help** button. This page is in English only.

## For the audience
- Nothing to install: open one address (or scan a QR code) on any phone, tablet or computer on the venue's WiFi. `http://open-caption.local/` works too where the network allows it.
- Large, high-contrast captions by default; per-device text size, dark-on-light, and dyslexia-friendly (OpenDyslexic) and high-legibility (Atkinson Hyperlegible) fonts.
- Scroll back to re-read a line without losing the live captions; the screen stays awake during the show; reconnects by itself after WiFi drops or sleep.
- Several languages in one show, chosen from the device's own language; works on a projector or shared screen too (`?display=projector`). A Troubleshoot button in Setup checks the network address, port, HTTPS certificate, `.local` name and settings file, lists recent warnings and errors (with a count on the button), and offers the full log and a support bundle.
- Runs on older phones (about iOS 12 and up) and tells people clearly when a browser is too old.

## For the operator
- One key does it all: Space sends the next line, Backspace undoes it, S skips a line without sending it, B blanks, C clears; jump to any line or section, search, and send a one-off announcement.
- A loaded show waits in **standby**: the audience sees your venue's own screen (a logo and a message you set once) until the operator presses START; **Go to standby** returns to it at any time without losing the place.
- A second device (a stage manager's tablet) can follow the show at `/follow` with its own viewer password (or the operator or admin one), watching only; it can send nothing.
- Operator notes, bookmarks and pacing hints (how long a line usually stays up, with a countdown to the next press) the audience never sees; loops for pre-show and intermission; songs, lyrics and two voices shown at once for musicals.
- Survives a restart: the position, sign-ins and last few undo steps are kept.
- An **audience report** when the show ends: how many different devices used the audience page, however long each stayed, with the busiest moment, the languages and a row for each device; printable and downloadable as CSV.

## For whoever sets it up
- A guided Setup page (with a bar at the bottom that holds Back, Next and the state of the step): show file editor with problems listed by line, image upload, look, a private rehearsal, then a Share step with the audience address, a printable QR sign and the way into the operator page, and a Report step for after the show.
- Edit the show file in any program you like: download it as a plain `.md`, edit it offline, and open it again to save over the show (or import any `.md` / `.txt` file as a new show). A guide for the person preparing the show file is [Preparing a show file](../web/help/admin-preparing-a-show-file.md); [Getting started](../web/help/admin-getting-started.md) walks through the six steps of Setup. Both are in the app's Help too. `open-caption --check-show show.md` checks a show file from the command line, like a lint, without starting anything.
- A show is one plain-text Markdown file (plus images): title, colours, fonts and options live in it, layered over `config.yaml` defaults. The editor lists every character name (most used first, misspellings marked) and keeps earlier versions, which can be named and compared; and shows can be exported, imported, duplicated and archived.
- Setup, the operator page, Help and all the built-in documentation are available in English, Spanish, French, German, Italian and Portuguese.
- Passwords are kept in `config.yaml` as SHA-256 digests, not plain text (Setup and the first run write them; `open-caption --hash-password` makes one by hand), and a Markdown heading (`# Act 2`) in the show file is a bookmark for the operator.
- The venue-wide defaults (look, language, pre-show message, loop time, audience limits) can be edited in Setup (Settings → Defaults); they are saved in `config.yaml` and apply at the next start.
- Built-in help, which can be popped out of Setup into a window of its own, a recent-log viewer, a support bundle, and a validated `config.yaml` that reports every mistake at once.
- Optional trusted HTTPS: a renewed certificate file is picked up without a restart, or the program can get and renew a Let's Encrypt certificate itself (DNS-01, with Cloudflare, GoDaddy and EasyDNS built in by default; see below).
- Screen-reader friendly pages, checked automatically with axe-core in both colour schemes.

## Under the hood
- One Go binary for macOS, Windows and Linux; no database, no internet needed, nothing fetched at runtime.
- Tested with 200 simultaneous audience devices (latency well under 250 ms), a real-browser end-to-end test on all three systems, and ready-made files to start it at boot (systemd, launchd, Task Scheduler).
- On the computer that runs it, Setup and the operator page open without a password (turn off with `access.local_auto_login: false`); every other computer signs in. Who may open Setup, the operator page and the Follow page can be changed in Setup (Settings → Access), which refuses a list that would lock out the computer being used. An admin session on the operator page has a **Setup** link back.
- Passwords, allowed computers (`access` lists) and a cross-site request guard protect Setup and the operator page; it is meant for a private network.
