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

- If the **first thing** in the show file is a loop, it starts when the operator puts the show on air (**START**), so the audience sees the pre-show while people find their seats.
- Any other loop (for example at intermission) starts when the operator presses the spacebar on it.
- The spacebar ends the loop and sends the next line. **Shift+Q** stops the loop and leaves the screen blank. **Back** returns to the loop and restarts it from the first item.

The operator's list shows a loop as one entry that expands to list its items. See [Loops for the operator](topic:operator-loops).

## A countdown timer

Give a loop a **timer** and the audience screens can show how much time is left, counting down live ("Back in 15 min"). Write `timer` in the loop's options, then put `{timer}` in any text of that loop where the time should appear.

```showfile
[loop: timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
![Refreshments are in the lobby](lobby.png)
[/loop]
```

The value of a timer is one of two things:

- A **number of minutes**, such as `timer 15` or `timer 15 minutes`: for an intermission. The countdown starts at the moment the loop starts (when the operator presses the spacebar on it), and starts again from the full time if the operator backs into the loop.
- A **time of day**, such as `timer 7:30pm`, `timer 7pm` or `timer 19:30`: for a pre-show. The screens count down to that time today. It uses the clock and time zone of the computer running the program, so check them. A time that has already passed shows 0:00.

Options share the brackets, separated by commas, in any order.

```showfile
[loop: 10s, timer 7:30pm]
**Welcome!**
The show begins in {timer}
[/loop]
```

- The time left is shown in minutes, rounded up and without seconds (`15 min`), with hours when there are more than one (`1 hr 5 min`), in the language of the device. At zero it stays at `0 min`; the loop carries on until the operator moves on.
- Only the items whose text contains `{timer}` show it, so you can alternate a message with the countdown.
- A phone whose own clock is wrong still shows the right time left, because the program's clock is used.
- Saving the show file during an intermission does not restart the countdown, and neither does restarting the program.
- In a translation, the loop must have the same timer as the main show file; translate the words around `{timer}`.
- The operator sees the time left in the green banner too. A `{timer}` in a loop without a timer is shown as typed, with a warning.

## Naming a loop

Give a loop a **name** and the audience report uses it to label that loop on its graph. Without a name the report guesses from where the loop stands (before the first line: *pre-show*; between lines: *intermission*; after the last line: *post-show*).

```showfile
[loop: 12s, name Intermission]
[fullscreen: **INTERMISSION**]
Back soon
[/loop]
```

- The name is free text of up to 40 characters, without commas. The audience never sees it.
- It shares the brackets with the other options, separated by commas, in any order: `[loop: 10s, name Act 1 break, timer 15 minutes]`.
- Only the name in the main language's show file is used; a translation may give its loops other names or none.
