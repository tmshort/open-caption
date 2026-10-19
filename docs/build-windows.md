# Building and running open-caption on Windows 10 or 11

Start here with a new PC that has nothing installed. When you finish you will have the program built from source and running. About 20 minutes, most of it downloads.

*(English only. The same steps for [macOS](build-macos.md) and [Linux](build-linux.md).)*

If you only want to **run** the program and someone gives you the finished `open-caption.exe`, you need none of this: see [the usage guide](usage-guide.md). These steps are for building it yourself.

## What you will install

| What | Why | Needed for |
|---|---|---|
| Git for Windows | gets the source (and gives you *Git Bash*, used by the tests) | building |
| Go 1.26 or later | the compiler | building |
| Node.js 22.12 or later and Google Chrome | the browser tests | only if you want to run the tests (step 6) |

You do **not** need Visual Studio, a C compiler or `make`: Windows builds with `build.bat`.

## 1. Open PowerShell

Click Start, type `PowerShell`, press Enter. Every command below is typed there, one line at a time, followed by Enter. (You do not need "Run as administrator".)

## 2. Install Git and Go

**With winget** (included in Windows 11 and in current Windows 10; check with `winget --version`):

```
winget install --id Git.Git -e
winget install --id GoLang.Go -e
```

Accept the license prompts. **Close PowerShell and open it again** so it finds the new programs, then check:

```
git --version
go version
```

`go version` must say `go1.26` or higher.

**Without winget**, download and run the installers, accepting the defaults:

- Git: <https://git-scm.com/download/win>
- Go: <https://go.dev/dl/> (the Windows `.msi`; choose `amd64`, or `arm64` on an ARM PC)

Then reopen PowerShell and check as above.

## 3. Get the source

```
cd $HOME
git clone https://github.com/tmshort/open-caption.git
cd open-caption
```

The repository may be private. If a sign-in window appears, sign in to GitHub with the account that was given access. Alternatively, on the repository's page choose **Code → Download ZIP**, extract it, and open PowerShell in the extracted folder.

## 4. Build

```
.\build.bat
```

(`build.bat` runs `build.ps1`, which does the work. Use `build.bat` rather than `build.ps1` directly: Windows normally refuses to run `.ps1` scripts typed at the prompt, and `build.bat` is not affected. The command stays the same if building ever needs more steps.)

The first build downloads the libraries it needs (internet required) and takes a minute or two. It creates `bin\open-caption.exe`.

**Choosing the DNS providers for automatic HTTPS certificates (optional).** The default build includes Cloudflare, GoDaddy and EasyDNS, which is right for most people. To choose others, as `ACME_PROVIDERS` does in the Makefile: `.\build.bat -AcmeProviders "easydns route53"`, `-AcmeProviders all` or `-AcmeProviders none`. Anything but the default needs Git for Windows (installed above) and internet. See the [HTTPS certificate guide](https-certificate-guide.md) and [docs/10](10-acme-lets-encrypt.md).

`go build -o bin\open-caption.exe .\cmd\open-caption` also works and builds the default.

## 5. Run

```
.\bin\open-caption.exe
```

On its first run the program:

- prints the **admin** and **operator** passwords **once**, in the PowerShell window. Write them down. (They are also in `config.yaml`; see [Where your configuration and shows are kept](#where-your-configuration-and-shows-are-kept).)
- makes **Windows Defender Firewall** ask whether to allow it. Tick **Private networks** and click **Allow access**; otherwise phones on the WiFi cannot reach it. (If your WiFi is set to *Public*, change it to Private in Settings → Network & internet → the network's properties, or also tick *Public networks*.)
- opens the Setup page (`http://localhost/admin`) in your browser. On this PC it needs no password. If port 80 is used by other software it uses `:8080` and prints the address.

To stop the program, press `Ctrl`+`C` in the PowerShell window. To start it again later, open PowerShell, `cd $HOME\open-caption` and run `.\bin\open-caption.exe`.

From here, follow [Getting started](../web/help/admin-getting-started.md) (also in the program's **Help** button), then the [usage guide](usage-guide.md). To start the program automatically at boot, see [deploy/README.md](../deploy/README.md) (`windows-install.ps1`).

`.\bin\open-caption.exe --help` lists the options.

## Where your configuration and shows are kept

By default in your user folder:

```
%AppData%\open-caption
```

which is `C:\Users\<you>\AppData\Roaming\open-caption`. Open it by pressing `Win`+`R`, typing `%AppData%\open-caption` and pressing Enter, or in PowerShell: `explorer $env:AppData\open-caption`. The program also prints the exact path of `config.yaml` when it starts.

| In the folder | What it is |
|---|---|
| `config.yaml` | The settings and the **admin and operator passwords** (readable only by you). Every setting is in the [configuration reference](configuration-reference.md). |
| `shows/` | One folder per show: `show.md` (the show file), `show.<language>.md` for translations, and `assets/` (the images), plus the saved earlier versions. |
| `standby-logo.*` | The logo on the standby screen, if you uploaded one. |
| `live.json`, `sessions.json` | The show's position and who is signed in, so a restart in the middle of a show carries on. Safe to ignore. |
| `tls/` | The HTTPS certificate, if HTTPS is on. |

The folder is **not** inside the `open-caption` source folder, so rebuilding or updating the program never touches your shows or passwords. To back up everything, copy the whole folder (Setup → Shows → **More ▾** → Export saves a single show). To start completely fresh, quit the program and rename or delete the folder: the next run creates new passwords. To keep the data somewhere else, start the program with `--data <folder>` every time (or set `OPEN_CAPTION_DATA`).

## 6. Optional: run the tests

Skip this unless you are changing the program. The browser tests are written for a Unix-style shell, so run them in **Git Bash** (installed with Git: Start → *Git Bash*).

1. Install **Google Chrome** from <https://www.google.com/chrome/>.
2. Install **Node.js** (LTS, 22.12 or later): `winget install --id OpenJS.NodeJS.LTS -e`, or the installer from <https://nodejs.org/en/download>. Reopen Git Bash and check `node --version`.
3. In Git Bash, from the `open-caption` folder:

```
go test ./...
cd test/e2e
npm ci --no-audit --no-fund
CHROME='C:\Program Files\Google\Chrome\Application\chrome.exe' node run.js
```

(`make` is not needed. The `Makefile` shortcuts such as `make e2e` need GNU make, which Windows does not have; the commands above are what they run.) More: [test/e2e/README.md](../test/e2e/README.md).

## Updating later

```
cd $HOME\open-caption
git pull
.\build.bat
```

## If something goes wrong

| What you see | What to do |
|---|---|
| `'go' is not recognized` or `'git' is not recognized` | Close PowerShell and open it again after installing. |
| `go: go.mod requires go >= 1.26.0` | Your Go is too old. Install the newest from <https://go.dev/dl/> (`winget upgrade GoLang.Go`). |
| The build stops with a network error | The build needs internet the first time. Check the connection and run `.\build.bat` again. |
| `bind: Only one usage of each socket address` or the program says it moved to `:8080` | Another program holds port 80 (IIS, Skype, another web server). The program already falls back to 8080; use the address it prints. |
| A phone cannot open the address | Allow the program on **Private networks** (Windows Security → Firewall & network protection → *Allow an app through firewall*). The phone must be on the same WiFi as the PC. |
| Antivirus quarantines `open-caption.exe` | A program you just built yourself is unsigned and unknown to the antivirus; add an exception for the `bin` folder. |
