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
| Q14 | Language selection UX? | Auto-detect from `Accept-Language`; otherwise default `en`; a picker (shown only with >1 language) navigates to `/?lang=xx`, so the URL form also works | 01, 02, 03, 04, 05 |
| Q16 | Comment syntax? | `<!-- … -->` for comments (never shown); `[note: text]` for operator-visible, audience-hidden notes; `#` reserved for future headings (ignored with an info diagnostic) | 01, 03 |
| Q15 | Restarting a stopped loop? | No restart button: Back "backs into" the loop and restarts it from item 0 | 02, 03, 05 |

## Open

None. A consistency review on 2026-10-14 resolved the following gaps (no decision needed, recorded for traceability):
operator commands use a separate `rev` counter so loop ticks can't cause 409s; "live" is defined; rehearsal is an isolated engine;
first-run passwords are console-only; Back undoes any operator action; any cue ends a running loop; scope text updated for loops and planned multi-language.
