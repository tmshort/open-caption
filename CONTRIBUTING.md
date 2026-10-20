# Contributing to open-caption

Thank you for your interest. This page says what helps, how to report things, and the rules that code and documentation changes follow.

## Where things stand

**The project is not accepting code contributions (pull requests) yet.** What is welcome now:

- **Bug reports** and things that were confusing to use ([how](#reporting-a-bug-or-an-idea)).
- **Ideas and requests** ([how](#reporting-a-bug-or-an-idea)).
- **Translation reviews** from native speakers ([how](#reviewing-a-translation)). This is the most useful help available.
- **Private security reports** ([how](#security)).

Pull requests that arrive anyway may be closed without review, with thanks. The rules further down describe how changes are made in this repository, and are what a contributor will be asked to follow when code contributions open.

## Reporting a bug or an idea

Open an [issue](https://github.com/tmshort/open-caption/issues). Please include:

- **What you did, what you expected, and what happened.** Screenshots help.
- **The version:** run `open-caption --version`.
- **Your system:** macOS, Windows or Linux and its version, and which browser and device show the problem (the audience page, Setup, or the operator page).
- **A support bundle** if the problem is in the running program: Setup → Share → *Download a support bundle*. It holds the recent log lines and the settings with passwords and certificate credentials removed. Look through it before attaching.
- For a **show file** problem, the output of `open-caption --check-show your-show.md` and, if you can share it, a small show file that shows the problem.

Do not put passwords, tokens or DNS credentials in an issue.

For an idea, say what you are trying to do rather than only the feature you imagine; there may be a simpler way already.

## Reviewing a translation

Setup, the operator page, Help and the guides are available in Spanish, French, German, Italian and Portuguese. **All of those translations were written without review by native speakers**, so wrong, awkward or unnatural wording is likely, especially in the screens used during a live show. A correction from someone who speaks the language is valuable.

Open an issue titled "Translation: <language>" and list each correction as:

- **Where:** the screen or document, or the file (`web/shared/i18n/<code>.js`, `web/help/<code>/…`, `docs/i18n/<code>/…`);
- **Now:** the current text;
- **Better:** your wording, and a few words on why if it is not obvious.

Say in the issue whether you are happy to be credited. Sentences with `{name}`-style placeholders keep the placeholder as it is.

## Security

If you find a vulnerability, please **report it privately** on the repository's *Security* tab, *Report a vulnerability*, rather than in a public issue. Say what is affected, how to reproduce it and what the effect is. The program is meant for a private network (see the security notes in the [usage guide](docs/usage-guide.md)), but problems with sign-in, access lists, the forged-request check, uploads or certificates matter all the same.

## Code of conduct

Be kind and be useful. Assume good faith, keep to the subject, and remember that the people who run this software are often volunteers working on a show. Harassment, insults and personal attacks are not acceptable, and the maintainer may remove comments, close or lock issues, and block people who persist. Concerns about behaviour can be raised privately with the maintainer through the repository's *Security* tab (*Report a vulnerability*) until a better channel exists.

## Rules for changes

These are how every change in this repository is made, by the maintainer and by anyone else.

### Get it building

[docs/build-macos.md](docs/build-macos.md), [docs/build-windows.md](docs/build-windows.md) and [docs/build-linux.md](docs/build-linux.md) start from a brand-new computer. [CLAUDE.md](CLAUDE.md) describes the architecture and the commands (`make help` lists every target). The specifications in `docs/` (`01`–`10`) are the source of truth: change the specification in the same commit as the behaviour, and reference requirement numbers (`REQ-xxx`) from tests and commit messages. The text file a show is written in is the **show file**, never "script".

### One change per commit

- **One fix or feature per commit.** If a request has several parts, make several commits, each complete on its own (code, tests, documentation, translations). Do not mix refactors or unrelated tidying into a feature or a fix.
- Commit messages have a short headline of **at most 80 characters**; **what changed and why** goes in the description after a blank line, which also names the `REQ-xxx` where there is one.

### The documentation changes in the same commit

Never "in a follow-up". Before committing, go through the places a change can touch:

- the specification and guides in `docs/`: requirements `01`, architecture `02`, show file `03`, API `04`, UI `05`, deployment `06`, testing `07`, roadmap `08`, open questions `09`, certificates `10`, `configuration-reference.md`, `usage-guide.md`, `venue-network-guide.md`, `https-certificate-guide.md` and the build guides;
- `README.md` and `CLAUDE.md`;
- the man page `docs/open-caption.1` and the `--help` text for any command-line option, and `docs/configuration-reference.md` for any new `config.yaml` setting (tests check that these agree);
- the in-app help (`web/help/*.md`);
- **the translations**: help topics in `web/help/<lang>/`, the four guides in `docs/i18n/<lang>/`, and on-screen text wrapped in `__('…')` with its translation in `web/shared/i18n/<lang>.js`, for es, fr, de, it and pt. The man page, `docs/10-acme-lets-encrypt.md` and the build and release guides are English only by design;
- tests, including the browser checks in `test/e2e/run.js` and the accessibility audit when a page changes.

### Check before you commit

```
gofmt -l .
go vet ./...
go test -race ./...
```

and for anything that touches the web pages:

```
make e2e
node test/e2e/a11y.js
node test/e2e/check-i18n.js
```

`make vulncheck` after changing a dependency. The browser tests need Node and Chrome (see the build guides). GitHub Actions runs the whole suite on Linux, macOS and Windows for every push.

### Sign off your commits

Every commit carries a `Signed-off-by` line, made with `git commit -s`:

```
Signed-off-by: Your Name <you@example.com>
```

It certifies the [Developer Certificate of Origin](https://developercertificate.org/): in short, that you wrote the change or have the right to submit it under the project's licence. Use your real name. Contributions are licensed under the project's [Apache License 2.0](LICENSE) (Section 5 of the licence), and you keep your copyright.

### AI-assisted work

Using AI tools to help write code, documentation or translations is allowed. You are responsible for it: understand every line you submit, run the checks above yourself, and say in the commit message or issue when a substantial part was generated. A sign-off is a statement by a person; an AI tool cannot give one (a `Co-Authored-By` line naming the tool is welcome in addition). Do not submit text or code you cannot vouch for, and do not paste in output that may copy someone else's work.

### Releases

Releases are made by the maintainer by pushing a `v*` tag; see [docs/releasing.md](docs/releasing.md).
