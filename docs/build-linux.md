# Building and running open-caption on Linux (Ubuntu or Fedora)

Start here with a new Ubuntu (or Debian, Mint) or Fedora (or RHEL, Rocky, Alma) computer that has nothing extra installed. When you finish you will have the program built from source and running. About 15 minutes, most of it downloads.

*(English only. The same steps for [macOS](build-macos.md) and [Windows](build-windows.md).)*

If you only want to **run** the program and someone gives you the finished program file, you need none of this: see [the usage guide](usage-guide.md). These steps are for building it yourself.

## What you will install

| What | Why | Needed for |
|---|---|---|
| `git`, `make`, `curl` | gets the source, runs the build, downloads Go | building |
| Go 1.26 or later, from go.dev | the compiler | building |
| Node.js 22.12 or later and Google Chrome | the browser tests | only if you want to run the tests (step 6) |

The Go that comes with the distribution is usually older than 1.26, so install Go from go.dev as below rather than with `apt` or `dnf`.

## 1. Open a terminal

On Ubuntu press `Ctrl`+`Alt`+`T`; on Fedora open *Terminal* from the Activities overview. Every command below is typed there, one line at a time, followed by Enter. Commands that start with `sudo` ask for your password.

## 2. Install git, make and curl

**Ubuntu / Debian / Mint:**

```
sudo apt update
sudo apt install -y git make curl ca-certificates
```

**Fedora / RHEL / Rocky / Alma:**

```
sudo dnf install -y git make curl
```

Check:

```
git --version
make --version
```

## 3. Install Go

Find the newest version at <https://go.dev/dl/> (it is shown as `go1.26.x` or higher) and set it here instead of the example. The architecture is `amd64` for an ordinary 64-bit PC (`uname -m` says `x86_64`) or `arm64` (`aarch64`).

```
GOVER=go1.26.0
GOARCH=amd64
curl -LO https://go.dev/dl/$GOVER.linux-$GOARCH.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf $GOVER.linux-$GOARCH.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin
go version
```

`go version` must say `go1.26` or higher. (The `~/.bashrc` line makes new terminals find Go; if you use another shell such as zsh, add that line to its startup file, for example `~/.zshrc`.)

## 4. Get the source

```
cd ~
git clone https://github.com/tmshort/open-caption.git
cd open-caption
```

The repository may be private. If git asks you to sign in, use your GitHub user name and a *personal access token* (GitHub → Settings → Developer settings → Personal access tokens) as the password. Alternatively, on the repository's page choose **Code → Download ZIP**, unzip it, and `cd` into the folder.

## 5. Build and run

```
make build
./bin/open-caption
```

The first build downloads the libraries it needs (internet required) and takes a minute or two. It creates `bin/open-caption`.

On its first run the program:

- prints the **admin** and **operator** passwords **once**, in the terminal. Write them down. (They are also in `config.yaml`; see [Where your configuration and shows are kept](#where-your-configuration-and-shows-are-kept).)
- tries to listen on port 80. A normal user may not, so on Linux it **uses port 8080** and prints the address (`http://<this computer>:8080/`). To use port 80 instead, allow it once: `sudo setcap 'cap_net_bind_service=+ep' ./bin/open-caption` (repeat after every rebuild).
- tries to open the Setup page (`http://localhost:8080/admin`) in your browser; on a computer with no desktop, open it from another computer, or use `--no-browser`. On this computer it needs no password.

**Let phones reach it: open the port in the firewall.**

- *Fedora* has a firewall on by default:
  ```
  sudo firewall-cmd --add-port=8080/tcp --permanent
  sudo firewall-cmd --reload
  ```
  (use `80/tcp` if you gave the program port 80).
- *Ubuntu*'s firewall (`ufw`) is off by default. If `sudo ufw status` says *active*: `sudo ufw allow 8080/tcp`.

To stop the program, press `Ctrl`+`C`. To start it again later: `cd ~/open-caption && ./bin/open-caption`.

From here, follow [Getting started](../web/help/admin-getting-started.md) (also in the program's **Help** button), then the [usage guide](usage-guide.md). To start the program automatically at boot, see [deploy/README.md](../deploy/README.md) (`open-caption.service`, systemd).

Useful afterwards: `./bin/open-caption --help` lists the options, and `man ./docs/open-caption.1` shows the manual page.

## Where your configuration and shows are kept

By default in your user configuration folder:

```
~/.config/open-caption
```

(or `$XDG_CONFIG_HOME/open-caption` if that variable is set). Open it with `ls ~/.config/open-caption`, or in the file manager with `Ctrl`+`H` to show hidden folders. The program also prints the exact path of `config.yaml` when it starts. A program started as a service (see [deploy/README.md](../deploy/README.md)) uses the `--data` folder named in its service file instead.

| In the folder | What it is |
|---|---|
| `config.yaml` | The settings and the **admin and operator passwords** (readable only by you). Every setting is in the [configuration reference](configuration-reference.md). |
| `shows/` | One folder per show: `show.md` (the show file), `show.<language>.md` for translations, and `assets/` (the images), plus the saved earlier versions. |
| `standby-logo.*` | The logo on the standby screen, if you uploaded one. |
| `live.json`, `sessions.json` | The show's position and who is signed in, so a restart in the middle of a show carries on. Safe to ignore. |
| `tls/` | The HTTPS certificate, if HTTPS is on. |

The folder is **not** inside the `open-caption` source folder, so rebuilding or updating the program never touches your shows or passwords. To back up everything, copy the whole folder (Setup → Shows → **More ▾** → Export saves a single show). To start completely fresh, quit the program and rename or delete the folder: the next run creates new passwords. To keep the data somewhere else, start the program with `--data <folder>` every time (or set `OPEN_CAPTION_DATA`).

## 6. Optional: run the tests

Skip this unless you are changing the program.

1. Install **Node.js 22.12 or later**. The distribution's version is often older; use NodeSource or `nvm`, whose current instructions are at <https://nodejs.org/en/download>. Check `node --version`.
2. Install **Google Chrome** (or Chromium) so that it is at `/usr/bin/google-chrome`:
   - Ubuntu: download the `.deb` from <https://www.google.com/chrome/> and run `sudo apt install ./google-chrome-stable_current_amd64.deb`.
   - Fedora: download the `.rpm` from the same page and run `sudo dnf install ./google-chrome-stable_current_x86_64.rpm`.
3. From the `open-caption` folder:

```
make test                                          # the Go tests
CHROME=/usr/bin/google-chrome make e2e            # the browser tests (the first run downloads their tools)
CHROME=/usr/bin/google-chrome make a11y           # the accessibility audit
make i18n                                          # checks every sentence is translated
```

If Chrome is somewhere else, set `CHROME` to its path. More: [test/e2e/README.md](../test/e2e/README.md) and `make help`.

## Updating later

```
cd ~/open-caption
git pull
make build
```

## If something goes wrong

| What you see | What to do |
|---|---|
| `go: command not found` | Open a new terminal, or run `export PATH=$PATH:/usr/local/go/bin`. |
| `go: go.mod requires go >= 1.26.0` | Your Go is too old. Repeat step 3 with the newest version. |
| `make: command not found` | Repeat step 2. |
| The build stops with a network error | The build needs internet the first time. Check the connection and run `make build` again. |
| The program says it is using `:8080` | Normal without permission for port 80; see step 5. |
| A phone cannot open the address | Open the port in the firewall (step 5). The phone must be on the same WiFi as this computer. |
