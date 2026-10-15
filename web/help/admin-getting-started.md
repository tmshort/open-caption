# Getting started

Setting up a show takes six steps. The row of buttons across the top of the page follows them in order, and a tick appears when a step looks complete. You can jump to any step at any time.

1. **Show** — create a show (or the sample show to look around).
2. **Show file** — paste or open your show file and fix anything the page points out.
3. **Images** — upload any pictures the show file uses.
4. **Look** — choose colors and text size.
5. **Test** — rehearse privately before anyone is in the room.
6. **Go Live** — share the address with the audience and start the show.

## Where things are stored

A show is a folder containing its **show file** (a plain text file, `show.md`), optional translations of it, and its images. The show file holds everything about the show except the pictures: its title, its look, and the lines. Settings the show file does not mention come from `config.yaml`, then from the program's built-in values.

## Before you start

- This computer must be on the same WiFi network as the audience's phones and tablets. See [Network and joining](topic:admin-network).
- The admin password and the operator password are in `config.yaml`; they were also printed in the console the first time the program started. The operator page has its own password so you can hand it to someone else without giving them setup access.
- Your work is saved on this computer. Use **Export** on the Show step to make a backup you can move to another computer. The export is a `.zip` holding `show.md` (the show file, with its title and look), a `show.<language>.md` for each translation, and an `assets` folder of images; **Import** turns such a zip into a new show.

## New to show files?

A show file is just text. Each line is one caption, and one press of the spacebar sends one line. Read [Writing a show file](topic:admin-writing-a-show-file), or create the sample show and look at how it is written. A starter you can download is at [/help/sample-show.md](/help/sample-show.md).

## Planning the equipment

For what to buy, where to put the WiFi access point, and a day-of checklist, read the [usage guide](topic:admin-usage-guide). The [venue network guide](topic:admin-venue-network-guide) is for whoever looks after the building's WiFi.

## Where to go next

- [Musicals](topic:admin-musicals)
- [Writing a show file](topic:admin-writing-a-show-file)
- [Loops: pre-show and intermission](topic:admin-loops)
- [Troubleshooting](topic:admin-troubleshooting)
