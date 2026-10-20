# `config.yaml` reference

`config.yaml` holds everything about *this computer and venue*: the passwords, ports, HTTPS, who may open Setup, and the defaults a show uses when its own show file says nothing. Everything about *one show* (title, colours, loop time, pre-show message…) lives in that show's file instead (see [Writing a show file](../web/help/admin-writing-a-show-file.md), "Show settings"); the layers are explained under "Where a setting is decided" below.

## The file

- **Where it is.** `<data folder>/config.yaml`; the data folder is `data_dir`, `--data`, or the default `<your user config folder>/open-caption` (on macOS `~/Library/Application Support/open-caption`, on Windows `%AppData%\open-caption`, on Linux `~/.config/open-caption`). The first lines the program prints say which file it is using. `--config` names another file.
- **First run.** If there is no file, the program writes one with random **admin** and **operator** passwords and prints them once in the console. The file is mode 0600 (private to your account) and the passwords are **plain text**, so keep it private.
- **You only need the keys you change.** A missing key keeps its default (the table below). The generated file lists everything.
- **Strict.** A misspelled key, a value of the wrong kind, or a value out of range stops the program at start-up with **all** the problems listed, each naming the setting (for example `audience.max_clients: 0 is not a client count: use 1 to 100000`). Nothing is guessed. The one exception: a file marked with a newer `config_version` than the program knows has its unknown keys ignored, with a warning.
- **Setup edits some of it.** Changing the passwords, the `.local` name, the standby message or the log detail in Setup rewrites those keys in the file and keeps your comments and every other setting. Everything else you edit by hand, then **restart the program**. (A renewed certificate file is the one thing picked up without a restart; see `https.cert_file`.)
- **Booleans** are `yes`/`no`/`true`/`false`. **Lists** are written `[a, b]` or one `- item` per line.

## Command line and environment

A flag, or its environment variable, overrides the file for that run. (The file is not changed.) A flag may be written with one or two dashes (`-data` or `--data`). `open-caption --help` prints this list, and the **man page** (`docs/open-caption.1`; read it with `man ./docs/open-caption.1`, or copy it to a `man1` folder such as `/usr/local/share/man/man1/`) describes the same options, files, signals and exit status.

| Flag | Environment | Effect |
|---|---|---|
| `--config FILE` | `OPEN_CAPTION_CONFIG` | Use this settings file (default `<data folder>/config.yaml`) |
| `--data FOLDER` | `OPEN_CAPTION_DATA` | The data folder (shows, sessions, `live.json`, certificates). Overrides `data_dir` |
| `--http-addr ADDR` | `OPEN_CAPTION_HTTP_ADDR` | Overrides `http.addr`, for example `:80` or `:8080` |
| `--https-addr ADDR` | `OPEN_CAPTION_HTTPS_ADDR` | Turns HTTPS **on** at this address (overrides `https.enabled` and `https.addr`) |
| `--log-level LEVEL` | `OPEN_CAPTION_LOG_LEVEL` | Overrides `log.level` every time the program starts (Setup says so when it is in force) |
| `--log-file FILE` | `OPEN_CAPTION_LOG` | Also append the log to this file |
| `--network-interfaces LIST` | `OPEN_CAPTION_NETWORK_INTERFACES` | Overrides `network.interfaces` every time the program starts: a comma-separated list of interface names such as `Ethernet` or `en0,en4` (Setup says so when it is in force) |
| `--network-restrict` | `OPEN_CAPTION_NETWORK_RESTRICT` | Also refuse connections made to addresses that are not offered, every time the program starts (like `network.restrict_listening: yes`). The variable counts when it is `yes`, `true`, `on` or `1` |
| `--no-browser` | | Do not open Setup in a browser at start-up (like `defaults.open_browser: no`) |
| `--version` | | Print the version and the DNS providers built in for automatic certificates, then exit |
| `--help` | | Print the options and the environment variables, and where to read more, then exit |
| `--check-show FILE [TRANSLATION ...]` | | **Check a show file** and print a report, like a lint, then exit: no server, no `config.yaml`. See "Checking a show file" below |
| `--assets FOLDER` | | The images folder for `--check-show` (default: the `assets` folder beside the show file) |
| `--strict` | | With `--check-show`: warnings also make the exit status 1 |
| `--dev`, `--show-file` | | **Developers only**: unauthenticated test endpoints on this computer, no standby; never use them at a venue |

## Checking a show file

`open-caption --check-show show.md` analyses a show file without loading it into the program: it runs the same parser and checks as the problems list on Setup's Show file step and prints what that list would show.

```
$ open-caption --check-show show.md show.es.md
show.md: info: 4 captions, 1 images, 0 fullscreen, 1 loops, 1 sections, 0 notes.
show.md:8: warning: Image "curtain.jpg" not found in assets.
show.md:13: error: [song] was opened but never closed with [/song]. The lyrics so far were kept.
show.md: 8 cues; 2 speakers: ALICE, BOB
show.es.md: 8 cues; language es

images checked against assets (1 found)
missing images: curtain.jpg (line 8 of show.md)
1 error, 1 warning, 1 note
```

- Each problem is `file:line: severity: message` (severity `error`, `warning` or `note`), the form editors and CI systems understand. A line about the whole file has no line number.
- Files after the first are **translations**: each is parsed as a translation and compared with the main file (the same number and kinds of cues, in the same order), as Setup does. The language is the file's `[language]` line or its name (`show.es.md`); two files for one language are an error.
- Images are checked against the `assets` folder beside the main file, or `--assets FOLDER`. Without a folder they are not checked, and the report says so. A missing image names the folder it was looked for in (Setup's editor says instead to upload it in the Images step).
- `-` as the file name reads standard input.
- **Exit status:** `0` no errors, `1` errors (or warnings with `--strict`), `2` a file could not be read. Put other options (`--assets`, `--strict`) **before** `--check-show`: the files that follow it are translations.

## Settings

Defaults are what you get without the key. "Setup" says where Setup can change it.

### Top level

| Key | Default | Meaning |
|---|---|---|
| `config_version` | `1` | Written by the program. A newer number than the program knows means the file came from a newer version |
| `admin_password` | generated | The password for Setup (`/admin`). Must not be empty (6 or more characters when changed in Setup). *Setup: Settings → Security* |
| `operator_password` | generated | The password for the operator page (`/operator`); the admin password works there too. Must not be empty. *Setup: Settings → Security* |
| `data_dir` | empty | The data folder. Empty means the default folder above. Must not be a file |

### `http`

| Key | Default | Meaning |
|---|---|---|
| `http.addr` | `:80` | Where the plain-HTTP server listens: `:80` (all addresses, port 80), `192.168.1.20:80`, or `127.0.0.1:8080` (this computer only). Must not be empty |
| `http.fallback_addr` | `:8080` | Used, with a visible warning, if `addr` cannot be used (port 80 is often busy or needs privileges). Empty means no fallback |

### `https`

HTTPS is **off** by default. There are three ways to get a certificate; choose one.

| Key | Default | Meaning |
|---|---|---|
| `https.enabled` | `no` | Serve HTTPS as well as HTTP, on separate ports (there is no redirect). Needs one of the three below |
| `https.addr` | `:443` | The HTTPS port. Required when enabled |
| `https.fallback_addr` | `:8443` | Used if `addr` cannot be bound |
| `https.self_signed` | `yes` | With no certificate files and no `acme`: the program makes its own certificate (kept in `<data folder>/tls/`). Browsers warn about it until it is accepted. `no` together with no files is an error |
| `https.cert_file`, `https.key_file` | empty | Your own certificate and key (PEM). Set **both** or neither; both files must exist. **The files are watched**: when they are replaced (for example by `certbot`), the new pair is used for new connections within about 30 seconds, no restart. A pair that cannot be loaded keeps the old certificate and logs one warning |

### `https.acme` (automatic certificates)

Off unless `enabled: yes`. Needs `https.enabled: yes`, a public domain name you own, and a DNS provider this program was built with. It cannot be combined with `cert_file`/`key_file`. How to set it up, step by step: the [trusted HTTPS guide](https-certificate-guide.md); the choices of DNS providers when building, and the sizes, are at the end of that guide.

| Key | Default | Meaning |
|---|---|---|
| `https.acme.enabled` | `no` | Get and renew a certificate automatically (ACME, DNS-01) |
| `https.acme.domain` | | The public name, for example `captions.example.org`. One name: no wildcard, no `.local`/`.lan`/`.internal`, no IP address. It must resolve, for the audience's phones, to this computer's address |
| `https.acme.email` | | Where the certificate authority sends expiry notices |
| `https.acme.agree_to_terms` | `no` | **Must be `yes`**: it records that you accept the certificate authority's terms of service |
| `https.acme.directory` | `letsencrypt` | `letsencrypt`, `letsencrypt-staging` (for trying it out: not trusted by browsers), or the `https://` address of another ACME directory |
| `https.acme.renew_before_days` | `30` | Renew when this many days or fewer remain (1 to 60) |
| `https.acme.dns.provider` | | The DNS provider, by its [lego name](https://go-acme.github.io/lego/dns/) (`cloudflare`, `godaddy`, `easydns`, …). Must be built into the program (`--version` lists them) |
| `https.acme.dns.credentials` | empty | The provider's own settings, as a map: for EasyDNS `EASYDNS_TOKEN: …` and `EASYDNS_KEY: …`. Names are capital letters, digits and underscores. The environment and `credentials_file` win over these |
| `https.acme.dns.credentials_file` | empty | A file of `NAME=value` lines with the same settings, read again at every attempt. Must exist; `chmod 600` it |
| `https.acme.dns.propagation_wait` | empty | A fixed wait (`90s`, `2m`, at most `30m`) for the new DNS record to spread, instead of checking for it |

### `defaults`

What a show uses when its own show file does not say. A show file's `[settings]` block overrides every value here (see [Writing a show file](../web/help/admin-writing-a-show-file.md)). Setup's Look step writes the show file, not these.

| Key | Default | Meaning |
|---|---|---|
| `defaults.language` | `en` | The main language of a show whose file does not name one: two or three letters, optionally with a region (`en`, `es`, `pt-br`) |
| `defaults.loop_item_seconds` | `10` | How long each item of a pre-show or intermission loop stays, in seconds (1 to 3600), unless the item says otherwise |
| `defaults.idle_message` | `Captions will begin shortly` | What the audience reads before the first line is sent when a show has no pre-show message (up to 200 characters) |
| `defaults.open_browser` | `yes` | Open Setup in a browser when the program starts (not a show setting; `--no-browser` turns it off for one run) |
| `defaults.look.background` | `#000000` | The caption background: `#rgb`, `#rrggbb`, or `black`, `white`, `yellow`, `red`, `green`, `blue`, `cyan`, `magenta`, `orange`, `gray` |
| `defaults.look.text_color` | `#ffffff` | The caption text colour (same forms) |
| `defaults.look.speaker_name_color` | `#ffffff` | The colour of speaker names (same forms) |
| `defaults.look.font` | `system` | `system`, `serif`, `mono`, `dyslexic` or `hyperlegible` |
| `defaults.look.text_size_percent` | `100` | Caption text size, 50 to 300 |
| `defaults.look.speaker_names` | `label` | `label` (names in one colour), `color` (a different colour for each character) or `hidden` |

### `audience`

| Key | Default | Meaning |
|---|---|---|
| `audience.max_clients` | `500` | The most audience devices connected **per language** (1 to 100000); more are turned away until some leave |
| `audience.scrollback_lines` | `200` | How many earlier lines a device can scroll back through (20 to 2000) |

### `access`

Who may open Setup and the operator page. The audience page is always open to everyone.

| Key | Default | Meaning |
|---|---|---|
| `access.admin_allow` | `[]` | Computers allowed to open Setup. Empty means **any computer on the network may try the password**. Entries: `localhost`, `private` (the usual home and office networks), an address (`192.168.1.50`) or a network (`192.168.1.0/24`). Example: `[localhost]` |
| `access.operator_allow` | `[]` | The same, for the operator page |
| `access.local_auto_login` | `yes` | A browser on the computer that runs the program (opened at `localhost`, `127.0.0.1` or `[::1]`) is signed in as admin with no password. Other computers always use a password. Turn it off on a computer other people can log in to |

### `log`, `mdns`, `standby`

| Key | Default | Meaning |
|---|---|---|
| `log.level` | `info` | `debug` (every request), `info` (events and sign-ins), `warn` or `error`. *Setup: Settings → About* |
| `mdns.enabled` | `yes` | Announce this computer as `<name>.local` on the local network |
| `mdns.name` | `open-caption` | The name: letters, digits and hyphens only, up to 63 characters (`.local` is added). *Setup: Settings → Network* |
| `standby.message` | empty | The line the audience sees on the standby screen. Empty means "Captions will begin shortly" in each viewer's language. One line, up to 200 characters. *Setup: Settings → Audience screen* (the logo is a file chosen there, not a setting) |

### `network`

| Key | Default | Meaning |
|---|---|---|
| `network.interfaces` | `[]` | The network interfaces that are used, by name (`Ethernet`, `Wi-Fi`, `en0`, `eth0`): only their addresses are given to the audience, listed in Setup, put in a new self-signed HTTPS certificate and announced as `<name>.local`. Empty means no choice has been made: every interface is used. Once a choice is made, an interface that appears later (a USB adapter, a VPN) stays off until it is added. When two used interfaces are on the same network (a cable and Wi-Fi at once) only one address is offered: the wired one, else the main connection's; leave the other out to choose. *Setup: Settings → Network* |
| `network.restrict_listening` | `no` | Also **refuse connections** made to an address of an interface that is not used (a phone that uses it is turned away), instead of only leaving the address out of the lists. Connections to this computer itself (`localhost`) are always served, so Setup can still be opened to change it. Setup never turns off the interface you are using it through. *Setup: Settings → Network* |

The addresses of an interface that is not used are not listed in the Share step or at start-up, are not given to the audience, are not announced as `<name>.local` and are not put in a new self-signed HTTPS certificate. Only interfaces that are connected are listed, and only one address is offered for each network. By default this only chooses what is *offered*: the server still accepts connections on every address, until `network.restrict_listening` is turned on.

## Where a setting is decided

From weakest to strongest: the program's built-in value → `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) → the show file's `[settings]`. A command-line flag or environment variable beats the file. The venue-level settings (ports, HTTPS, passwords, `access`, `data_dir`, `audience.max_clients`, `mdns`, `network`, `standby`, `log`) exist only here.

## A complete example

```yaml
config_version: 1
admin_password: "set-a-long-one"
operator_password: "another-one"
http:
  addr: ":80"
  fallback_addr: ":8080"
https:
  enabled: yes
  addr: ":443"
  fallback_addr: ":8443"
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org
    agree_to_terms: yes
    directory: letsencrypt
    dns:
      provider: easydns
      credentials_file: /etc/open-caption/dns.env
mdns:
  enabled: yes
  name: open-caption
network:
  interfaces: []
  restrict_listening: no
standby:
  message: "Welcome to the Little Theatre"
defaults:
  language: en
  loop_item_seconds: 10
  idle_message: "Captions will begin shortly"
  look:
    font: hyperlegible
    text_size_percent: 120
audience:
  max_clients: 500
  scrollback_lines: 200
access:
  admin_allow: [localhost]
  operator_allow: [private]
  local_auto_login: yes
log:
  level: info
```
