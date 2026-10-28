# Running the show

The page shows the show file as a list. Lines already sent are dim, the line **on screen now** has a green dot, and the line **next** is large with a yellow arrow.

## The three things to remember

1. **Space** sends the next line.
2. **Backspace** undoes the last thing you sent.
3. **Shift+Q** stops a pre-show or intermission loop and blanks the screen.

Follow the actors, and press **Space** as each line is spoken. You can use a clicker; a clicker's next and previous buttons work like Space and Backspace.

## Blank and Clear

Both leave the audience screens empty, but they are not the same:

| | **Blank** (`B`) | **Clear** (`C`) |
|---|---|---|
| What the audience sees | Nothing: just the background | An empty caption area, ready for new lines |
| The earlier lines | **Kept.** Press `B` again, or send the next line, and they are still there above it | **Erased.** The next line starts on an empty screen |
| Use it for | A blackout, a pause, or hiding the captions for a moment | A new scene, or after a song, so old lines do not pile up |
| Undo | `B` again, or **Backspace** | **Backspace** |

A show file can do the same with the `[blank]` and `[clear]` commands; they appear in your list as `BLANK SCREEN` and `CLEAR SCREEN`.

## Words on this page

- **Standby**: the show is loaded, but the audience sees only the venue's standby screen.
- **On air**: the audience sees the show. **START** puts a show on air; **Go to standby…** puts it back in standby, keeping its place.
- **Restart show…**: back to the first line, and to standby.
- **Blank** hides the captions but keeps the earlier lines; **Clear** erases them (see above).

## Before the show

- Run the operator page on a computer with a keyboard (a laptop is fine). It is built around the **Space** key, and a phone or tablet is not recommended: a touch screen is easy to mis-tap during a show.
- Sign in to the operator page. You will see "You are in control". If it says another operator is in control, use **Take over** only if you are meant to be the operator.
- A second device (a stage manager's tablet) can watch from the Follow page, `/follow`, with the viewer password (or the operator or admin password). It shows what you see, notes included, but cannot send anything; Setup's **Share** step lists its address.
- Check the viewer count (the people icon). It shows how many devices are connected.
- A loaded show waits in **standby**: the audience sees the venue's standby screen (a logo and a message) and nothing of the show, and this page says **STANDBY**. Press **Space** (the button says **START**) to put the show on air. If the show file starts with a pre-show loop, it begins then.

## During the show

- **Pace:** when a line has a `[pace]` in the show file, its time is shown beside the next line, and a thin bar along the bottom of the status strip fills as the line is on screen. When the next press is due the next line gets a yellow frame and the bar turns yellow. Nothing is sent by itself and there is no sound.
- Lines marked as notes or sections are for you. They never use up a press of Space.
- Pictures and full-screen messages show only as a label here, such as `IMAGE: curtain.jpg`. You never see the picture itself.
- **Songs:** sung lines have music notes (♪) around them and are shown in italics on the audience's screens. A song's title appears as a blue bookmark for you only. Aim to press Space as each phrase begins.
- **Languages:** if the show has translations, you follow the main language and everything you send goes to every language at once. The audience count in the header lists each language, for example `83 (en 60 · es 23)`.
- **Several voices at once:** a line marked `TOGETHER` shows two or more singers on the screens at the same time. One press of Space sends all of them.
- The strip under the header shows who is in control, **Item n of total**, and **Last sent … ago**, so you can see at a glance that the show and your screen are still in step. It turns yellow after two minutes without a line.
- **Go to section…** (beside the search box) lists the show file's bookmarks; picking one selects it, and you confirm the jump.
- To skip ahead or go back, click a line, then confirm "Jump here". You can also search with `/`.
- A one-off announcement such as "The show will resume shortly" can be sent with **Announcement…**. Press **Back** to remove it. While it is on the screens, a blue banner on this page quotes it.
- **Go to standby…** (top right) takes the show off the audience's screens: they see the standby screen again, and the show keeps its place. **START** puts it back on air exactly where it was.

See [the keyboard](operator-keyboard.md) and [If something goes wrong](operator-if-something-goes-wrong.md).

## After the show: the audience report

When you reach the end of the show the banner says **End of the show.** with an **Audience report** link (Setup's **Report** step shows it afterwards). The report says how many different devices used the audience page while the show was on air, however long each stayed: a phone that dropped off the WiFi and came back counts once. It shows the totals (phones, tablets, computers, projector screens, the most connected at once, how many are still connected), the languages, and a row for each device. It also draws a graph of how many devices were connected over time, with the pre-show and intermission loops marked, so you can see how many people were there when the show began and how many stayed through the break. It keeps counting until you go back to standby, restart or load another show, and you can print it or download it as a CSV file for a spreadsheet.

A second device that only watches the show (the Follow page) is listed in the same table, marked **Viewer**, but is not counted in any of the numbers or in the graph. Devices are told apart by their network address and browser type, so the numbers are an estimate: phones that share both count as one, and a phone whose address changes counts twice. Each report is saved as a file in the `reports` folder of the data folder; the newest 30 are kept, and you can delete one with **Delete…** in Setup's **Report** step, or delete the files there.
