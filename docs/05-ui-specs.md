# 05 — UI Specifications

Three interfaces served from one binary. Vanilla JS/CSS, no build step,
responsive. Setup and Operator share one app shell (`/admin`) with tabs; the
audience page is separate and as light as possible.

## 1. Setup interface — `/admin` (tab "Setup")

Goal: a first-time volunteer can finish without reading documentation.

### First run
1. Welcome screen → choose an operator PIN → done.
2. Wizard stepper across the top: **① Show → ② Script → ③ Images → ④ Look → ⑤ Test → ⑥ Go Live**.
   Steps are also reachable directly for experienced users. Each step shows a green check when complete.

### Steps
- **Show**: name; list of existing shows with Duplicate / Archive / Delete (confirm).
- **Script**:
  - Drag-and-drop file, "Choose file…", or paste into an editor.
  - Editor pane (monospace, line numbers) + Diagnostics pane. Clicking a diagnostic jumps to the line.
  - Live preview of the parsed cue list (kinds shown as icons: 💬 caption, 🖼 image, ⬛ blank, 🧹 clear, 📑 section, 📝 note).
  - A "Script help" slide-out showing the syntax cheatsheet with copy-able examples (content of doc 03 "Directives" table).
  - Insert-directive toolbar buttons (Clear, Blank, Image…, Fullscreen text, Section, Slideshow) so authors never type brackets by hand.
  - Re-import warns if cue count differs and the show is live.
- **Images**: grid of uploads with thumbnails; drag-drop multi-upload; shows "used in lines 12, 40" or "unused"; missing-but-referenced images listed at top with an upload-in-place button.
- **Look**: theme (background/text/accent colors, font family from system + bundled web-safe options, base size), speaker display (off / label / colored), idle message. Live preview in a phone-shaped and a projector-shaped frame.
- **Test**: "Rehearsal mode" opens the Operator view against the show with a test audience preview pane side-by-side; does not require a real audience; position resets afterward.
- **Go Live**: shows audience URL (large), QR code (printable page for lobby signs), host IPs, current audience count, and a "Start Show" button that activates the show and opens the Operator tab.

### Network panel (always visible in the header)
`Audience URL: http://192.168.1.10/ ▸ copy ▸ QR` with a warning banner if no non-loopback address is detected or the port isn't 80.

### Guard rails
- While live, Setup tabs show a banner "Show is live" and script-modifying actions require typing the show name to confirm (REQ-052).
- Every destructive action uses an in-app confirm dialog, never silently.

## 2. Operator interface — `/admin` (tab "Live"), also `/operator`

Layout (desktop): three columns of the script as a vertical list.

```
┌ header: Show name • ● LIVE • 83 viewers • Controller: you • [Takeover] ──────┐
│  SENT (dimmed, scrolls up)                                                   │
│   …                                                                           │
│   OSCAR: Felix, you're late.                                                  │
│  ─────────────── on screen now ───────────────                                │
│ ▶ FELIX: I'm so sorry,  (just sent)                                           │
│  ═════════════ NEXT (press Space) ═════════════                               │
│  ► (lights fade)                                                              │
│    [🖼 IMAGE: curtain.jpg]                                                     │
│    📑 Act 2, Scene 1                                                          │
│  UPCOMING (normal)                                                            │
│   …                                                                           │
├──────────────────────────────────────────────────────────────────────────────┤
│  [◀ Back]   [   NEXT ▶  (Space)   ]   [Blank (B)] [Clear (C)] [Message…]      │
│  Search/jump: [____________]   Mini-map of sections                           │
└──────────────────────────────────────────────────────────────────────────────┘
```

Behavior:
- **Next cue is visually unmistakable** (highlighted block with large text). The list auto-scrolls to keep the next cue ~⅓ from the top.
- **Image and fullscreen cues** render as a labeled chip: `🖼 IMAGE: curtain.jpg`, `⛶ FULLSCREEN: INTERMISSION`, `🎞 SLIDESHOW (3 images, 10 s)`; no image rendering. Operator-only `note` and `section` cues are styled distinctly and are skipped over by Space (they don't consume a keypress, they auto-pass when reached).
- **Keyboard** (ignored when typing in an input):

  | Key | Action |
  |-----|--------|
  | `Space`, `→`, `Enter`, `PageDown` | Next (clickers typically send PageDown/Right) |
  | `Backspace`, `←`, `PageUp` | Back (undo last) |
  | `B` | Toggle blank |
  | `C` | Clear screen |
  | `/` or `Ctrl+F` | Focus search/jump |
  | `Esc` | Cancel search / close dialog |
  | `Home` | Reset (confirm) |

  Space must call `preventDefault` to avoid page scroll. Key repeat is ignored; a command is debounced (~120 ms) to prevent double-fire, while the UI optimistically shows pending state and reconciles with the server response.
- **Safety**: clicking a cue in the list opens a "Jump here?" confirmation (single click select, Enter/second-click confirm) so accidental taps don't change the screen. Jumps never replay intermediate lines onto the audience view as a flood — they set the resulting state.
- **Status indicators**: connection state (green/yellow/red); audience count; warning if another session claims control; "Last sent n seconds ago".
- **Touch layout** (phone/tablet): large bottom Next/Back buttons; swipe not used (accidental). Wake lock requested.
- **Dark theme default** (theater is dark; low glare). Optional "stage-dark" red-tinted mode.
- **Ad-hoc message**: dialog to send one-off caption or fullscreen text (e.g., "Show paused — please remain seated") without editing the script.

## 3. Audience interface — `/`

Principle: **no interaction needed.** Controls, if any, are hidden/minimal.

- **Captions surface**: full-viewport, dark background by default, large text. New captions appear at the bottom and the surface scrolls smoothly (instant if `prefers-reduced-motion`). Latest caption full brightness; the previous ~2 slightly dimmed; older progressively dimmer and eventually clipped off the top. Speaker label small-caps accent color above/before the text.
- **Fullscreen overlay**: image (contain-fit, letterboxed on theme background) or centered large text/HTML, covering everything, with fade transition (≤ 250 ms). On the next caption it is removed and the scrolling surface returns with its prior content, unless a `[clear]` occurred.
- **Slideshow**: client-side timer over the image list, cross-fade; preloads images; restarts on snapshot.
- **Blank**: pure theme background.
- **Idle**: before the show starts, shows the configured message and a subtle "connected" indicator.
- **Connection**: a small unobtrusive status dot; if offline > 5 s shows "Reconnecting…" overlay while retaining the last text visible underneath, and auto-recovers on reconnect (re-snapshot). Reconnect backoff 0.5 s → 5 s with jitter.
- **Keep awake**: request Screen Wake Lock on load and on first user visibility change; on iOS Safari where unavailable, silent hidden `<video>` loop fallback (documented, off by default). Page shows a one-time tiny hint "Tip: turn off Auto-Lock to keep captions on" only when wake lock fails.
- **Optional local controls** (collapsed, appear on tap/click in a corner): text size − / +, light/dark toggle. Stored in `localStorage`; never required.
- **Projector mode** (`?display=projector`): font scale ×1.5 by default, hidden cursor, no controls, no connection dot unless offline.
- **Responsive**: font size uses `clamp()` with `vw`/`vh` units relative to the theme `scalePct`; works portrait and landscape; safe-area insets for iPhone.
- **Accessibility**: captions container `role="log" aria-live="polite"`; contrast ≥ 7:1 by default; no flashing.
- **Size budget**: audience page + JS + CSS ≤ 50 KB gzipped, no external requests.

## Visual design

Clean, minimal, theater-friendly: near-black backgrounds, high-contrast
off-white text, amber accent. Setup UI is lighter and friendlier (light theme
default) and uses plain-language labels, not jargon (say "Captions", not "cues", in the UI).
