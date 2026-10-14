# 09 — Open Questions and Assumed Defaults

Each item lists the default assumed in the specs. Confirm or change; update the
affected docs when decided.

| # | Question | Assumed default | Affects |
|---|----------|-----------------|---------|
| Q1 | Should Setup and Operator be one app (tabs) or separate URLs? | One app at `/admin` with tabs; `/operator` is a direct link to the Live tab | 05 |
| Q2 | Is one caption = one script line the right granularity? Some scripts have long speeches that need to be split to fit the screen. | Yes; authors split lines. Possible later helper: "auto-split long lines at N characters" in setup | 03 |
| Q3 | Is plain HTTP acceptable? Screen Wake Lock needs a secure context (HTTPS) on some browsers; localhost excepted. | HTTP; use video-loop fallback on iOS; revisit optional self-signed/local-CA HTTPS | 06, 05 |
| Q4 | Is a PIN enough for operator/setup security on a guest-reachable port? | Yes, PIN + rate limiting; optionally allow restricting `/admin` to specific IPs or the host only | 04, 06 |
| Q5 | SVG uploads? | Rejected in v1 (XSS surface); PNG/JPG/GIF/WebP only | 04, 06 |
| Q6 | Should the audience see previous captions scroll (a rolling transcript), or only the last N lines? | Rolling scrollback, dimmed with age, bounded at 200 | 02, 05 |
| Q7 | Can the operator edit a script's text during a live show (typo fix)? | Allowed with confirmation; cursor preserved by cue index, warn if cue count changes | 05 |
| Q8 | Do you need per-character colors? | Optional speaker colors in Look step; default off | 03, 05 |
| Q9 | Script source formats beyond `.txt` (e.g., `.docx`, PDF, Final Draft)? | `.txt`/`.md` only in v1 | 03 |
| Q10 | Pre-show behavior: auto-start loop when show loaded, or only after operator action? | Auto-displays as soon as the show is activated | 03 |
| Q11 | Port 80 default OK, given the privilege problems on Linux? | Try 80, fall back to 8080 with prominent notice | 06 |
| Q12 | Does the venue's network allow guest devices to reach a host on the same network? | Unknown; needs venue check before show — see doc 06 | 06 |
| Q13 | License for the repository | Not set; choose before publishing (MIT/Apache-2.0 typical) | — |
