# Preparing a show file

This page is for the person who turns the script into captions. You do not need to run the show or know anything about the network. A show file is an ordinary text file, so you can write it in any program you like, on any computer, and bring it into Setup when it is ready. [Writing a show file](topic:admin-writing-a-show-file) lists every marking you can use; this page is about how to go about the job.

## The short version

1. Get the script into a plain text file. Start from the sample show at [/help/sample-show.md](/help/sample-show.md) if you like.
2. Make **one line of the file equal one caption**: one press of the spacebar.
3. Add the extras that help the audience and the operator (speaker names, sound descriptions, sections, notes).
4. Bring the file into Setup, read the list of problems, fix them and save.
5. Rehearse it on the Test step, and fix what you notice.

## What kind of file

- Any plain text file works: a name ending in `.txt` or `.md`, and any name at all, such as `Our Town - captions v3.txt`.
- Save it as **UTF-8** if your program asks. Accented letters, curly quotes and music notes then survive. A file saved as Windows "ANSI" (Windows-1252) is also understood. Word processor files (`.doc`, `.docx`, `.pdf`) are not text files: copy the text out and paste it into a plain text editor first, or use Save As and choose plain text.
- Windows (CRLF) and Mac or Linux (LF) line endings both work, and so does the invisible mark some editors put at the start of a file.
- A `.zip` is only needed for a whole backup, which includes the images. The show file is just the text.

## Deciding where each caption begins and ends

The program never splits or joins your lines. A caption is exactly what you typed on one line, so you control how much the audience reads at once.

- Aim for lines that read comfortably in one glance, a sentence or part of a sentence. Very long speeches should be cut at natural pauses, where the actor breathes.
- Blank lines mean nothing, so use them freely between speeches to keep the file easy to read.
- If two or more lines must appear together, join them with a backslash at the end of the line, or put them between `[[` and `]]`.
- Use `HAMLET: line` to show who is speaking. Put the name in capital letters, then a colon.
- Describe sounds and stage directions in parentheses, on a line of their own, for example `(thunder rumbles)`. They are shown to the audience in a different style.

```showfile
[section: Act 1]
[note: house lights down, then start]
GEORGE: Morning, Emily.
EMILY: Oh, hello, George.
(a train whistle, far away)
```

## Things that only the operator sees

Mark your place and leave reminders without disturbing the audience: `[section: Act 2]` adds a bookmark in the operator's list, and `[note: wait for applause]` shows a reminder. Neither uses up a press of the spacebar. A comment between `<!--` and `-->` is for you and is seen by nobody.

## Working in your own editor

You never have to type in the editor on the Show file step. Instead:

1. On the **Show file** step, press **Download the show file**. You get a plain `.md` file.
2. Open it in your own program, edit it and save it. Keep the file as plain text.
3. Back on the Show file step, use **Upload a show file…** and choose your file. Its text replaces what is in the editor, and the list of problems is updated at once. Nothing is changed for good until you press **Save**.
4. Press **Save**. The version it replaces is kept under **Earlier versions**, where you can compare it with the new one, so a mistake can always be undone.

The same works for a translation: choose the language tab first, then download and open the file for that language. You can also drop a file onto the editor.

To start a **new show** from a file, use **Import a show** on the Show step and choose the file. It becomes a new show, named by the `title:` line in its `[settings]` block, or by the file's name if there is none. Pictures are not part of a plain text file; upload them on the Images step, using the same names as in your `[image: ...]` lines.

## Checking your work

- The Show file step lists every problem with its line number. Click one to jump to it. Errors never stop you from saving, but an item with an error is skipped, so fix them before the show.
- You can check a show file without opening Setup: on the captioning computer run `open-caption --check-show show.md` (add translations after it, for example `show.es.md`). It prints the same problems with their line numbers and exits with a non-zero status when there are errors, so you can use it while you edit or in a script. `open-caption --help` lists the options.
- The panel beside the editor shows what the operator will step through, one row per press of the spacebar. Check that the count of rows is what you expect.
- Run the **Test** step to see the captions as the audience will, on a phone as well as on a large screen.
- If other people also work on the show file, see *If two people edit the same show file* in [Writing a show file](topic:admin-writing-a-show-file).

## Before you hand over

- Read the file once from top to bottom as the audience would, not as a proofreader.
- Check that every `[image: ...]` name matches an uploaded picture, and that every `[loop]`, `[song]`, `[together]` and `[fullscreen]` has its closing line.
- Give the operator a copy of the file and the section names, so they know where to find their place. [Songs and music](topic:admin-musicals), [Loops](topic:admin-loops) and [Languages](topic:admin-languages) have the details for those parts.
