# Browser end-to-end test

Starts the built program on a free port with a throwaway data folder and drives **real Chrome** through
Setup, the operator page and the audience page. It exits non-zero if any check fails.

What it covers: sign-in, the sample show and its Spanish translation, the edit-conflict dialog, the Look
step, rehearsal isolation, a live show in two languages (Accept-Language, the language menu, translated
labels), Back / Blank / Clear, songs and voices shown together, per-character colours, the dyslexia-friendly
font, help pausing the operator keys, scroll-back with "back to live", Settings (changing a password), and a
server restart (still signed in, position kept, undo still works).

```
make build
cd test/e2e && npm install && node run.js
```

Environment: `CHROME` (path to Chrome or Chromium; default is the macOS location), `OC_BIN` (the program,
default `../../bin/open-caption`), `OC_SHOTS` (a folder to save screenshots in).

Notes: Chrome allows only six connections per site and every open page holds a live stream, so the test
closes pages it has finished with. The test needs no internet.
