# Making a release

Push a version tag and GitHub builds the program for every platform, packages it with the documentation, and attaches everything to a **draft** release for you to check and publish. (English only; for whoever maintains the project.)

## The short version

```
git tag v1.2.3
git push origin v1.2.3
```

Then open the repository's **Releases** page, check the draft (the notes, the files), and click **Publish release**. A tag with a dash, such as `v1.2.3-rc1`, is marked a pre-release automatically.

Before tagging: the work is committed and pushed, `main` is what you want to ship, and `go test -race ./...` and `make e2e` pass. (The release also runs `go test ./...` itself and stops if it fails.)

## What a release contains

One download per platform, named `open-caption_<version>_<system>_<arch>`:

| System | Files | Format |
|---|---|---|
| macOS | `macOS_arm64` (Apple silicon), `macOS_amd64` (Intel) | `.tar.gz` |
| Windows 10/11 | `Windows_amd64` | `.zip` |
| Linux | `Linux_amd64`, `Linux_arm64` | `.tar.gz` |

plus `checksums.txt` (SHA-256 of every file). Each download unpacks into one folder holding:

- the program (`open-caption`, or `open-caption.exe`), with the default DNS providers for automatic certificates (Cloudflare, GoDaddy, EasyDNS; [doc 10](10-acme-lets-encrypt.md));
- `README.md` (the repository README without the parts that only concern the source code), `LICENSE`, `NOTICE`;
- `docs/`: the usage guide, venue network guide, HTTPS certificate guide, configuration reference and the certificate design/how-to (`10-acme-lets-encrypt.md`), the man page, and the translations in `docs/i18n/`;
- `web/help/` (the in-app help topics, which the README links to) and `web/sample/` (the sample show);
- `deploy/`: the start-at-boot files for systemd, launchd and Windows Task Scheduler.

**Not included:** the numbered design specifications `01`–`09`, the build guides, the tests and the source. A test (`deploy/release_test.go`) checks that no design document slips in and that no link inside the download is dead.

The version reported by `open-caption --version` is the tag without the `v`.

## The programs are not signed

There is no paid code signing or notarization. The release notes (written by `.goreleaser.yaml`) and the README tell users what to do: on macOS `xattr -dr com.apple.quarantine .` or **Open Anyway** in System Settings; on Windows **More info → Run anyway** in the SmartScreen window (and **Unblock** on the zip file if it is blocked); nothing on Linux. If you ever buy certificates, signing is added to `.goreleaser.yaml` (`signs:` and the notarize options).

## How it works

- `.github/workflows/release.yml` starts on a `v*` tag and runs `make release`.
- `make release` installs the pinned GoReleaser (`GORELEASER_VERSION` in the Makefile, into `.tools/`), runs `go test ./...`, then `goreleaser release --clean`.
- `.goreleaser.yaml` says what to build, what goes in each archive, the checksums, the release notes (the commit messages since the previous tag, plus the installing notes) and that the release is a draft.
- `tools/package-readme.sh` makes the download's README: everything between `<!-- source-only:start -->` and `<!-- source-only:end -->` in `README.md` is left out (invisible on GitHub), so wrap anything in the README that needs the source code (build instructions, the specifications) in those markers.

To change what is in the download, edit the `files:` list in `.goreleaser.yaml`; `go test ./deploy` then checks the result.

## Releasing from your own computer

Useful when GitHub Actions is not available (for example the billing limit on a private repository):

```
git tag v1.2.3
git push origin v1.2.3
GITHUB_TOKEN=$(gh auth token) make release
```

It builds on your computer and uploads a draft release for the tag at `HEAD`. (The tag must exist and point at the commit you are on.) If the workflow also runs for the same tag, one of the two will report that the release already exists; use one or the other.

To see exactly what users will get without publishing anything:

```
make release-snapshot      # builds every platform into dist/ (about a minute)
```

## Fixing a mistake

- **A draft is wrong:** delete the draft on GitHub, fix the problem, move the tag (`git tag -d v1.2.3`, `git push origin :refs/tags/v1.2.3`, tag again) and push it.
- **Published a bad release:** edit it to a draft or delete it on GitHub, and publish a new, higher version. Do not reuse a version people may have downloaded.
- **Changing the GoReleaser version:** edit `GORELEASER_VERSION` in the Makefile, run `make release-snapshot` and look at `dist/`.
