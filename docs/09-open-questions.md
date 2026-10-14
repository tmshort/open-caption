# 09 — Decisions and Open Questions

All original questions are resolved. Each row records the decision and the docs
that reflect it. Add new questions at the bottom.

## Resolved

| # | Question | Decision | Docs |
|---|----------|----------|------|
| Q1 | Setup and Operator: one app or separate URLs? | Separate: `/admin` (setup) and `/operator` (live) | 04, 05 |
| Q2 | Caption granularity? | One script line = one cue; long lines wrap; the author splits lines to match the performance style. No auto-splitting | 03 |
| Q3 | HTTP vs HTTPS? | Support both side by side, no redirect. ACME/Let's Encrypt is a possible later feature (needs Internet + public name) | 01 (REQ-006), 06 |
| Q4 | Authentication? | Separate passwords for `/admin` and `/operator`, plaintext in `config.yaml` for now, login form + session cookie, **no HTTP Basic auth**. Config generated on first run with random passwords; admin session also opens `/operator` | 01, 02, 04 |
| Q5 | SVG? | Not supported. PNG/JPG/GIF/WebP; browser scales; server doesn't rescale. Mitigations: prefetch upcoming images, warn on files > 1 MB | 04, 05, 06 |
| Q6 | Audience scrollback? | Rolling, dimmed with age, bounded (200) | 02, 05 |
| Q7 | Edit script during a live show? | Allowed with confirmation; cursor kept by index, warn if cue count changes | 01 (REQ-052), 05 |
| Q8 | Per-character colors? | Optional; default theme is pure black/white, high contrast, no speaker colors | 01 (REQ-045), 03, 05 |
| Q9 | Script formats? | Markdown only (`.md`; `.txt` parsed identically). No embedded HTML. Headings ignored for now (reserved) | 01 (REQ-020/022), 03 |
| Q10 | Pre-show behavior? | A **loop** (mixed images/text, per-item duration, default from config). A loop as the first cue auto-starts on activation; any other loop starts on NEXT. STOP (Shift+Q) → blank screen; NEXT (Space) → cue after the loop | 01 (REQ-025/030), 02, 03, 05 |
| Q11 | Ports; languages? | Try 80/443, fall back to 8080/8443 with a warning; all settings in `config.yaml`. Multi-language later: one identical-structure script per language, `[language: xx]` first | 01 (REQ-005/026), 02, 03, 06 |
| Q12 | Guest network reachability? | A requirement of the venue's guest network (documented as a prerequisite) | 06 |
| Q13 | License? | Apache-2.0 (`LICENSE` added) | 01 (REQ-007) |

## Open

| # | Question | Leaning | Affects |
|---|----------|---------|---------|
| Q14 | Language selection UX: URL only (`/?lang=es`), or also a small picker on the audience page (the page must otherwise need no interaction)? | URL/QR per language, plus an optional picker in the corner | 03, 05 |
| Q15 | Should a loop that is stopped (blank) be restartable by the operator without Back? | Yes: a "Restart loop" button (not a key) later; v1 uses Back | 05 |
| Q16 | Markdown headings as future section markers vs. comments. | Ignored in v1; revisit | 03 |
