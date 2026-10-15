# Loops: pre-show and intermission

A **loop** is a set of full-screen pictures and text that repeats until the operator moves on. Use one before the show starts and another at intermission.

```showfile
[loop: 10s]
[image: sponsors.png | 15s]
**Welcome!**
Please silence your phones
[/loop]
```

Each line inside the loop is one item:

- a plain line of text is shown as full-screen text;
- `[image: file.png]`, `![caption](file.png)` and `[fullscreen: text]` also work;
- `[blank: 3s]` is a short pause on a blank screen;
- add `| 15s` at the end of an item to change how long it stays;
- `[loop: 10s]` sets the time for items that do not say. If you give no time, the default from `config.yaml` (10 seconds) is used.

Captions, `[clear]`, sections, notes and other loops cannot go inside a loop.

## When a loop starts and stops

- If the **first thing** in the show file is a loop, it starts by itself as soon as the show is started, so the audience sees the pre-show while people find their seats.
- Any other loop (for example at intermission) starts when the operator presses the spacebar on it.
- The spacebar ends the loop and sends the next line. **Shift+Q** stops the loop and leaves the screen blank. **Back** returns to the loop and restarts it from the first item.

The operator's list shows a loop as one entry that expands to list its items. See [Loops for the operator](topic:operator-loops).
