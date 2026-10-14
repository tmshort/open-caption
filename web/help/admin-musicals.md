# Musicals

Open captions for a musical work best when lyrics are easy to tell apart from spoken lines, and when the operator can keep up with the music. This page collects what works.

## Marking lyrics

Lyrics are shown **in italics between music notes**, for example ♪ Where is love? ♪. Choose whichever way of writing them suits you:

| Write | Result |
|-------|--------|
| `~ Where is love? ~` | ♪ Where is love? ♪ (the closing `~` is optional) |
| `OLIVER: ~ Where is love?` | the same, with the speaker's name above |
| `[song: Title]` … `[/song]` | every line inside is a lyric; the title is for the operator only |
| `♪ typed or pasted ♪` | exactly what you typed, with no italics |
| `~` on a line by itself | just a note: ♪, for an instrumental passage |

Use `[song]` for long numbers, and `~` for a single sung line in the middle of a scene.

In Setup, the **🎵 Song**, **♪** and **♫** buttons above the script editor put these in for you. The ♪ and ♫ characters can also be copied from anywhere, because they are ordinary text.

## Tips for good lyric captions

- **One line per sung phrase.** Break lines where the singers breathe, so the operator's press of Space matches the music. Shorter lines are easier to keep up with.
- **Use speaker names for duets** and for lines sung by different characters in turn. Leave them off for a chorus.
- **Instrumental breaks:** use `~` on its own, or `(Orchestra plays)`, to mark them. A line in parentheses is shown as a stage direction.
- **Keep spoken and sung lines apart.** A speaker line followed by `~` lyrics reads clearly: the audience sees the notes and italics change.
- **Repeats:** write the lines out each time. The operator presses Space once per line either way.
- **Clear between numbers:** put `[clear]` after the last line of a song so the next scene starts on an empty screen.
- **Operator reminders:** `[note: tempo picks up after the key change]` is shown to the operator only.
- **Rehearse with the real music.** The Test step lets the operator practice pressing Space against the cast recording.

## An example

```script
[section: Act 1, Number 4]
[clear]
(The orchestra begins.)
[song: Where Is Love]
OLIVER: Where is love?
Does it fall from skies above?
[note: wait for the held note]
Is it underneath the willow tree
That I've been dreaming of?
[/song]
[clear]
MRS. CORNEY: That's quite enough of that!
```

See also [Writing a script](topic:admin-writing-a-script).
