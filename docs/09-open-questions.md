# 09 — Decisions and Open Questions

All original questions are resolved. Each row records the decision and the docs
that reflect it. Add new questions at the bottom.

## Resolved

| # | Question | Decision | Docs |
|---|----------|----------|------|
| Q1 | Setup and Operator: one app or separate URLs? | Separate: `/admin` (setup) and `/operator` (live) | 04, 05 |
| Q2 | Caption granularity? | One show file line = one cue; long lines wrap; the author splits lines to match the performance style. No auto-splitting | 03 |
| Q3 | HTTP vs HTTPS? | Support both side by side, no redirect. ACME/Let's Encrypt is a possible later feature (needs Internet + public name) | 01 (REQ-006), 06 |
| Q4 | Authentication? | Separate passwords for `/admin` and `/operator`, plaintext in `config.yaml` for now, login form + session cookie, **no HTTP Basic auth**. Config generated on first run with random passwords; admin session also opens `/operator` | 01, 02, 04 |
| Q5 | SVG? | Not supported. PNG/JPG/GIF/WebP; browser scales; server doesn't rescale. Mitigations: prefetch upcoming images, warn on files > 1 MB | 04, 05, 06 |
| Q6 | Audience scrollback? | Rolling, dimmed with age, bounded (200) | 02, 05 |
| Q7 | Edit show file during a live show? | Allowed with confirmation; cursor kept by index, warn if cue count changes | 01 (REQ-052), 05 |
| Q8 | Per-character colors? | Optional; default theme is pure black/white, high contrast, no speaker colors | 01 (REQ-045), 03, 05 |
| Q9 | Show file formats? | Markdown only (`.md`; `.txt` parsed identically). No embedded HTML. Headings ignored for now (reserved) | 01 (REQ-020/022), 03 |
| Q10 | Pre-show behavior? | A **loop** (mixed images/text, per-item duration, default from config). A loop as the first cue auto-starts on activation; any other loop starts on NEXT. STOP (Shift+Q) → blank screen; NEXT (Space) → cue after the loop | 01 (REQ-025/030), 02, 03, 05 |
| Q11 | Ports; languages? | Try 80/443, fall back to 8080/8443 with a warning; all settings in `config.yaml`. Multi-language: one identical-structure show file per language, `[language: xx]` first (built) | 01 (REQ-005/026), 02, 03, 06 |
| Q12 | Guest network reachability? | A requirement of the venue's guest network (documented as a prerequisite) | 06 |
| Q13 | License? | Apache-2.0 (`LICENSE` added) | 01 (REQ-007) |
| Q14 | Language selection UX? (built) | Auto-detect from `Accept-Language`; otherwise default `en`; a picker (shown only with >1 language) navigates to `/?lang=xx`, so the URL form also works | 01, 02, 03, 04, 05 |
| Q16 | Comment syntax? | `<!-- … -->` for comments (never shown); `[note: text]` for operator-visible, audience-hidden notes; `#` reserved for future headings (ignored with an info diagnostic) | 01, 03 |
| Q15 | Restarting a stopped loop? | No restart button: Back "backs into" the loop and restarts it from item 0 | 02, 03, 05 |
| Q17 | Should Setup/operator be restricted by default? | No: default is open (a volunteer-friendly first run), with `access.admin_allow` / `access.operator_allow` and warnings in Settings and the guides. A default of `admin_allow: [localhost]` was considered and rejected because Setup is normally used from other computers | 01 (REQ-055), 02, 06 |
| Q18 | How do languages stay in step? | One engine and hub per language, driven together by the operator's commands; a drifting language is re-synced; structure is checked cue by cue and a mismatched translation is held back | 02, 03 |
| Q19 | Keep logins across a restart? | Yes: hashed sessions on disk, dropped when a password changes | 02 (REQ-058) |
| Q20 | Archive shows? | Not built; Export and Delete cover it | 05 |
| Q21 | Hidden-video fallback to keep iPhones awake? | Not built; a one-time Auto-Lock tip is shown instead | 01 (REQ-043) |
| Q22 | Log levels? | Not built; plain lines to the console, an optional file, and the support bundle | 01 (REQ-061) |

## Open

- **Real devices:** the program has only been tested in desktop Chrome on one Mac. A rehearsal on the oldest phones, tablets and computers you expect, on the real WiFi, is the main remaining risk (see the acceptance checklist).
- **Browser floor:** the audience page works on current browsers; Setup and the operator page need `<dialog>` (Safari 15.4+). The exact minimum versions are unverified.
- **Passwords are plain text** in `config.yaml` by decision (Q4); hashing them is a later option.
- **Release packaging:** no signed or notarized builds, installers or auto-update.
- **Screen-reader and contrast audit** of all four pages has not been done.
