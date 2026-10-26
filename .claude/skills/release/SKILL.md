---
name: release
description: Make an open-caption release. Creates and pushes a signed version tag, writes Enhancements and Fixes release notes for the maintainer to approve, waits for CI and the release build, then publishes the draft release (as latest only when it is the highest version).
argument-hint: <version, e.g. v0.5.0>
disable-model-invocation: true
---

# /release <version>

Release `$ARGUMENTS` of open-caption. The project's release mechanics are in `docs/releasing.md`: pushing a `v*` tag starts two workflows (`ci` on the tag, and `release`, which builds every platform and creates a **draft** GitHub release). This skill drives that from the tag to the published release.

Typing `/release <version>` is the maintainer's go-ahead to create and push the tag and, once the notes are approved and everything passes, to publish. Everything else in this file is a stop condition: **if a step fails or something is unexpected, stop and report; never delete, move or force-push a tag, and never delete a release, without asking.**

## 1. Check the version

- `$ARGUMENTS` is empty: ask for it and stop.
- Accept `1.2.3` or `v1.2.3` (optionally `-rc1`-style suffix); the tag is always `v1.2.3`. Anything else: say what is wrong and stop.
- Look at the existing tags (`git tag -l 'v*' --sort=-v:refname`, `git ls-remote --tags origin`). If the tag exists locally or on `origin`: stop.
- The version is the **highest** if it has no suffix and sorts above every other suffix-free `v*` tag (`sort -V`). Otherwise it is a **patch** (or a pre-release): it will be published but **not** as latest. Tell the maintainer which it is. If it is a patch and `HEAD` is not on the line it patches, ask which commit to tag before going on.

## 2. Check the commit to tag

Tag the current `HEAD` of the branch you are on (normally `main`).

- `git status --porcelain` must be empty; `git fetch origin` and `HEAD` must equal `origin/<branch>` (nothing unpushed, nothing to pull). Otherwise stop.
- CI must have passed for this exact commit: `gh run list --workflow ci --commit $(git rev-parse HEAD) --json status,conclusion,url`. If it is still running, wait for it (Monitor with an `until` loop on `gh run view`); if it failed or there is no run, stop and say so. (An earlier release was tagged on a failing commit; that must not happen again.)
- Show the maintainer the tag, the commit (`git log -1 --oneline`), and the previous tag, in one short message, and carry on (the invocation is the approval).

## 3. Create and push the signed tag

The existing tags are GPG-signed annotated tags whose message is just the version:

```
git tag -s -m "v1.2.3" v1.2.3
git push origin v1.2.3
```

If signing fails (a passphrase prompt cannot appear here), tell the maintainer to run the `git tag -s -m …` line themselves with `! ` in front, then continue. Verify with `git tag -v v1.2.3` (or `git cat-file -p`) that it carries a signature.

Pushing starts the `ci` and `release` workflows for the tag.

## 4. Write the release notes while the build runs

- The previous version is the highest `v*` tag **below** the new one (`git tag -l 'v*' --sort=-v:refname`, skipping suffixes unless the new version has one). Read the commits `PREVIOUS..v1.2.3` (`git log --format='%h %s%n%b' PREVIOUS..v1.2.3`), and look at the diff where a message is unclear.
- Write the notes for the people who run the program (venue operators and show preparers), not for developers: **Enhancements** (new things, improvements, security updates) and **Fixes** (things that were wrong). At most **4 bullets in each**; group related commits into one bullet; leave out refactors, tests, CI and documentation-only changes unless a user would notice them. Start each bullet with a short bold-free label and a colon (as in earlier releases: `gh release view v0.4.0`), name an issue or pull request number when the commit has one, and use plain sentences with no commit hashes. A section with nothing in it is left out.
- Format (this goes at the very top of the release body):

```
## Enhancements
- Label: what it does for the user.

## Fixes
- Label: what was wrong and what happens now.
```

- **Show the bullets to the maintainer and ask for approval** (AskUserQuestion: *Approve* / *Change it*). If they want changes (or give free text), redo the bullets and ask again. Do not touch the release until they approve.

## 5. Wait for the builds

Find the two runs for the tag: `gh run list --branch v1.2.3 --json databaseId,workflowName,status,conclusion,url` (they start a few seconds after the push; retry). Wait for both with Monitor (`gh run view <id> --json status,conclusion` in an `until` loop; the `release` run takes about 3 minutes and the `ci` run on a tag, which includes macOS and Windows, about 10).

- Both must end `success`. If one fails, show which job and the failing lines (`gh run view <id> --log-failed | tail`), do **not** publish, and stop. The maintainer decides whether to fix and re-tag.
- The draft release must exist with its files: `gh release view v1.2.3 --json isDraft,assets,body`. Expect `isDraft: true` and the archives for macOS (arm64, amd64), Windows (amd64) and Linux (amd64, arm64) plus `checksums.txt`, six assets. Anything else: stop and report.

## 6. Put the notes in the draft

The body GoReleaser wrote holds the install instructions and a Changelog of commits. Keep all of it and put the approved notes **above** it (if the body already starts with `## Enhancements` or `## Fixes`, replace that part instead of adding another):

```
gh release view v1.2.3 --json body -q .body > "$SCRATCH/body.md"
# write "$SCRATCH/notes.md": approved bullets, a blank line, then the old body
gh release edit v1.2.3 --notes-file "$SCRATCH/notes.md"
```

(Use the session's scratchpad directory for the files.) Read the body back to check it.

## 7. Publish

Only now, with both runs green, the assets present and the notes in place:

- Highest version (no suffix, sorts above every other suffix-free tag): `gh release edit v1.2.3 --draft=false --latest`
- Otherwise (a patch for an older line, or a pre-release): `gh release edit v1.2.3 --draft=false --latest=false`

Then confirm with `gh release list --limit 5` and `gh release view v1.2.3 --json url,isDraft,isPrerelease`, and report: the release URL, whether it is marked Latest, and anything the maintainer should know. Do not do anything further (no further tags, no cleanup).
