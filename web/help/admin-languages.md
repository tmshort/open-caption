# Languages

A show can offer its captions in more than one language. Each language is its **own show file**, and the operator follows only the main one: one press of Space sends the same moment in every language at once.

## Adding a language

1. In the **Show file** step, click **＋ Add a language** next to the language tabs.
2. Choose the language (or pick "Another language…" and type a short code such as `cy`).
3. Leave **Start from a copy of the main show file** ticked. You get a copy with the right shape, so you only have to replace the words.
4. Translate, then **Save**.

Each translation begins with a line naming its language. The editor adds it for you:

```showfile
[language: es | Español]
Hola, Margaret.
```

The name after the `|` is what audience members see in the language menu, so write it in its own language (Español, Français, Deutsch).

## The one rule: keep the same shape

Every language must have **the same cues in the same order**: a caption stays a caption, a `[clear]` stays a `[clear]`, a loop keeps the same number of items, and voices shown together stay the same number of voices. Only the words change. Speaker names, notes, section names, image names and loop timings can be different.

If a translation stops matching, the editor shows an error with the line where it goes wrong, for example "Cue 4 does not match the English show file: that show file has [clear] here, and this one has a caption". A language that does not match is **not offered to the audience** until it is fixed. The English (main) show file is unaffected, and the Go Live step lists the problem.

To tell the audience which words go with which line, keep the lines in step: if the English splits a speech into three lines, so does the Spanish, even where that is not how the language would break it.

## What the audience sees

- A device is given the best language the show has. It looks at the language settings of the phone or browser (for example Spanish for Mexico gets Spanish), and uses the main language when nothing matches.
- A small **🌐** control in the top-left corner lists the languages. Choosing one switches that device, and the address changes to end in `?lang=es`, so you can print a link or QR code for each language.
- The page's own words ("Reconnecting…", text size, and so on) are also translated for Spanish, French, German, Italian and Portuguese. Other languages show those few labels in English; the captions themselves are always in the show file's language.
- Languages written right to left (Arabic, Hebrew) line up on the right.

## Other things to know

- **Pictures:** every language can use the same images. For a picture with words on it, upload a version per language and name each one in that language's show file.
- **Pre-show message:** the Look step has a "message before the show starts" box for each language.
- **Messages from the operator** ("Message…") are sent exactly as typed to every language, so type them in the languages you need.
- **Changing a show file while the show is on:** adding or fixing a translation takes effect straight away and keeps the operator's place. Removing a language is not allowed while the show is live.
- **Counts:** the operator page shows how many devices are reading each language, for example `👥 83 (en 60 · es 23)`.
- **Backups:** Export saves every language; Duplicate copies them too.

See also [Writing a show file](topic:admin-writing-a-show-file).
