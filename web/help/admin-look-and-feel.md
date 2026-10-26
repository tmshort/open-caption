# Look and feel

The **Look** step sets how captions appear on every audience device. The defaults, white text on a black background, are the easiest to read in a dark room and are best for accessibility.

- **Colors:** a contrast reading shows how readable the combination is. Aim for 7:1 or better; the page warns when it is too low.
- **Text size:** scales all text. Devices fit the text to their screen, so this is a relative change, not a size in points. The slider goes from 50 to 300 percent; 100 is normal.
- **Font:** clean (the device's usual font), serif, monospace, or **dyslexia-friendly (OpenDyslexic)**, a font with heavier letter bottoms and distinctive letter shapes that some readers with dyslexia find easier. It is built into the program, so it needs no internet, and a device downloads it (about 115 KB) only when it is used. Lines get a little extra height with it. Choose **high legibility (Atkinson Hyperlegible)** for a clean font designed by the Braille Institute for low vision, with letters made easy to tell apart (for example capital I, lowercase l and the number 1). Both fonts are built into the program, and a device downloads one only if it uses it (about 17 KB for Atkinson Hyperlegible). Audience members can also pick either font for their own device from the small **Aa** control, whatever font you chose here.
- **Speaker names:** show in one color (the *speaker name color*), show them in a **different color for each character**, or hide them. Character colors are chosen automatically, in the order the characters first speak, from a palette that works for color-blind viewers, and are adjusted to stay readable on your background.
- **Speaker name position:** above the line (the default), **beside the text** on the same line (in bold), or **in a column** to the left of the text, like a script. Voices shown together always put the name beside their own text.
- **Pre-show message:** shown until the first line is sent, if the show file does not begin with a loop.

Changes are saved with **Save look** and reach connected devices straight away. **Save look writes your choices into the `[settings]` block at the top of the show file** (see [Writing a show file](topic:admin-writing-a-show-file)), so the show file carries its own look wherever it is copied; you can also edit that block by hand. Anything the show file does not set uses the defaults from `config.yaml` (its `defaults` section), which in turn start from the program's built-in values. If the Show file editor has unsaved changes, save them before saving the look.

## A screen for the whole room

If a browser is showing captions on a projector or large monitor, open the audience address with `?display=projector` on the end. The text is larger, the mouse pointer is hidden and nothing else is shown. A shared screen never scrolls, so if the newest caption is taller than the screen (many voices at once, say) the text is shrunk until it fits.

## Re-reading earlier lines

Anyone can scroll the captions up to re-read something they missed. New lines then wait behind a **↓ Back to live (3)** button (it counts what arrived) instead of pulling the page away, and the older lines stop fading while someone is reading them. Tapping the button, or scrolling to the bottom, returns to the live captions. A screen showing `?display=projector` is never scrollable.

## Audience controls

Audience members never need to do anything. A small, faint **Aa** control in the top-right corner of the audience page lets a person make their own text bigger or smaller, switch to dark-on-light, choose the dyslexia-friendly or high-legibility font, or turn **Screen stays on** off (the page keeps a phone's screen awake while the show runs, and someone short of battery may prefer not to), but only on their own device.
