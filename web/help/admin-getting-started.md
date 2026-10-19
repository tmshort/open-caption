# Getting started

Setting up a show takes six steps. The row of buttons across the top of the page follows them in order, and a tick appears when a step looks complete. You can jump to any step at any time.

1. **Shows** — create a show (or the sample show to look around).
2. **Show file** — paste or open your show file and fix anything the page points out.
3. **Images** — upload any pictures the show file uses.
4. **Look** — choose colors and text size.
5. **Test** — rehearse privately before anyone is in the room.
6. **Run** — share the address with the audience and open the operator page, where the operator starts the show.

## Standby and going on air

Opening the operator page loads the show in **standby**: the audience sees the venue's standby screen (a logo and a message, set once under **Settings → Audience screen**) and nothing of the show. The operator presses **Space** (the button says **START**) to put the show on air, and **Go to standby…** takes it off again without losing its place. Setup, rehearsal and editing are never in front of the audience while the show is in standby, and a show in standby can be deleted.

## Where things are stored

A show is a folder containing its **show file** (a plain text file, `show.md`), optional translations of it, and its images. The show file holds everything about the show except the pictures: its title, its look, and the lines. Settings the show file does not mention come from `config.yaml`, then from the program's built-in values.

## Finished with a show?

Rename, Duplicate, Export, Archive and Delete are in the **More ▾** menu on each show. On the Shows step, **Archive** puts a show away: it is kept in full but listed under *Archived shows* and cannot go on air. **Restore** brings it back. **Delete** removes it for good (Export it first if you might want it back). The show that is on air can be neither archived nor deleted.

## Before you start

- This computer must be on the same WiFi network as the audience's phones and tablets. See [Network and joining](topic:admin-network).
- The admin password and the operator password are in `config.yaml`; they were also printed in the console the first time the program started. The operator page has its own password so you can hand it to someone else without giving them setup access. On the computer that runs the program, Setup and the operator page open **without a password** when the address starts with `localhost`, `127.0.0.1` or `[::1]` (the person at that computer is trusted as the admin). Other computers, and any request that has been relayed by a proxy, still sign in with a password. Set `access.local_auto_login: false` in `config.yaml` if other people can log in to this computer.
- Setup, the operator page and this Help are available in English, Spanish, French, German, Italian and Portuguese. Use the language menu at the top of each page; the choice is remembered in that browser.
- Your work is saved on this computer. Use **Export** (in a show's **More ▾** menu) on the Shows step to make a backup you can move to another computer. The export is a `.zip` holding `show.md` (the show file, with its title and look), a `show.<language>.md` for each translation, and an `assets` folder of images; **Import** turns such a zip into a new show, and also takes a plain show file (`.md` or `.txt`, with any name) as a new show.

## New to show files?

A show file is just text. Each line is one caption, and one press of the spacebar sends one line. For the whole job, from script to finished file (including editing in your own program and bringing the file back), read [Preparing a show file](topic:admin-preparing-a-show-file). The markings are listed in [Writing a show file](topic:admin-writing-a-show-file). Or create the sample show and look at how it is written. A starter you can download is at [/help/sample-show.md](../sample/sample-show.md).

## Planning the equipment

For what to buy, where to put the WiFi access point, and a day-of checklist, read the [usage guide](topic:admin-usage-guide). The [venue network guide](topic:admin-venue-network-guide) is for whoever looks after the building's WiFi.

## Where to go next

- [Musicals](topic:admin-musicals)
- [Writing a show file](topic:admin-writing-a-show-file)
- [Loops: pre-show and intermission](topic:admin-loops)
- [Troubleshooting](topic:admin-troubleshooting)
