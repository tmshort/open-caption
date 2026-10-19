# Building and running open-caption on macOS

Start here with a new Mac that has nothing installed. When you finish you will have the program built from source and running. About 20 minutes, most of it downloads.

*(English only. The same steps for [Windows](build-windows.md) and [Linux](build-linux.md).)*

If you only want to **run** the program and someone gives you the finished program file, you need none of this: see [the usage guide](usage-guide.md). These steps are for building it yourself.

## What you will install

| What | Why | Needed for |
|---|---|---|
| Apple's Command Line Tools | `git` (gets the source) and `make` (runs the build) | building |
| Go 1.26 or later | the compiler | building |
| Node.js 22.12 or later and Google Chrome | the browser tests | only if you want to run the tests (step 6) |

You do **not** need Xcode (the large App Store program) or Homebrew.

## 1. Open Terminal

Press `Cmd`+`Space`, type `Terminal`, press `Return`. Every command below is typed there, one line at a time, followed by `Return`.

## 2. Install the Command Line Tools (git and make)

```
xcode-select --install
```

A window asks you to install the tools; click **Install**, agree to the license, and wait for it to finish (several minutes). If Terminal says the tools are already installed, that is fine. Check:

```
git --version
make --version
```

Both should print a version number.

## 3. Install Go

1. Open <https://go.dev/dl/> in a browser.
2. Under *Featured downloads* choose the macOS installer (`.pkg`) for your Mac: **Apple** (`darwin-arm64`) if it has an M1, M2, M3 or later chip, **Intel** (`darwin-amd64`) otherwise. If unsure, choose  → *About This Mac*: it says *Chip* or *Processor*.
3. Open the downloaded file and click through the installer.
4. **Close Terminal and open it again** (so it finds Go), then check:

```
go version
```

It must say `go1.26` or higher. If it says a lower number, install the newest version from the same page.

## 4. Get the source

```
cd ~
git clone https://github.com/tmshort/open-caption.git
cd open-caption
```

The repository may be private. If git asks you to sign in, use your GitHub user name and a *personal access token* (GitHub → Settings → Developer settings → Personal access tokens) as the password. Alternatively, on the repository's page choose **Code → Download ZIP**, unzip it in your home folder, and `cd` into the unzipped folder.

## 5. Build and run

```
make build
./bin/open-caption
```

The build downloads the libraries it needs the first time (internet required) and takes a minute or two. It creates `bin/open-caption`.

On its first run the program:

- prints the **admin** and **operator** passwords **once**, in the Terminal window. Write them down. (They are also in `config.yaml`; the program prints where.)
- asks *Do you want the application "open-caption" to accept incoming network connections?* Click **Allow**, or phones on the WiFi cannot reach it.
- opens the Setup page (`http://localhost/admin`) in your browser. On this Mac it needs no password. If port 80 was busy it uses `:8080` and prints the address.

To stop the program, press `Ctrl`+`C` in the Terminal window. To start it again later: `cd ~/open-caption` and `./bin/open-caption`.

From here, follow [Getting started](../web/help/admin-getting-started.md) (also in the program's **Help** button), then the [usage guide](usage-guide.md).

Useful afterwards: `./bin/open-caption --help` lists the options, and `man ./docs/open-caption.1` shows the manual page.

## 6. Optional: run the tests

Skip this unless you are changing the program.

1. Install **Google Chrome** from <https://www.google.com/chrome/> (the tests expect it in `/Applications`).
2. Install **Node.js** (the LTS release, 22.12 or later) with the macOS installer from <https://nodejs.org/en/download>. Reopen Terminal and check `node --version`.
3. From the `open-caption` folder:

```
make test       # the Go tests
make e2e        # the browser tests (the first run downloads their tools)
make a11y       # the accessibility audit
make i18n       # checks every sentence is translated
```

More: [test/e2e/README.md](../test/e2e/README.md) and, for the full list of targets, `make help`.

## Updating later

```
cd ~/open-caption
git pull
make build
```

## If something goes wrong

| What you see | What to do |
|---|---|
| `xcode-select: error: command line tools are already installed` | Nothing; continue. |
| `go: command not found` | Close and reopen Terminal after installing Go. Check `ls /usr/local/go/bin/go` exists; if not, run the installer again. |
| `go: go.mod requires go >= 1.26.0` | Your Go is too old. Install the newest from <https://go.dev/dl/>. |
| `make: *** No rule to make target` | You are not in the `open-caption` folder: `cd ~/open-caption`. |
| The build stops with a network error | The build needs internet the first time. Check the connection and run `make build` again. |
| A phone cannot open the address | Click **Allow** on the firewall question (System Settings → Network → Firewall → Options shows the list). The phone must be on the same WiFi as the Mac. |
