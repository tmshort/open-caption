<!--
  Sample open-caption script: "The Garden Party" (a short made-up scene).
  Everything between the arrows like this is a comment and is never shown to anyone.

  THE BASICS
    - Each line is one caption. One press of the spacebar sends one line.
    - Blank lines are ignored, so use them to keep the script readable.
    - Long lines wrap on the screen. You decide where to split the lines.
    - Lines in [square brackets] are commands.
    - [note: ...] shows a reminder to the operator only. The audience never sees it.
-->

<!-- PRE-SHOW: a loop of pictures and text. Because it is the first thing in the
     script, it starts by itself as soon as the show is loaded. The operator presses
     the spacebar to end it and send the first line. Each item shows for 10 seconds
     unless it says otherwise (like the sponsors picture, which stays 15). -->
[loop: 10s]
[image: welcome.png | Welcome to The Garden Party | 15s]
**Welcome!**
Captions will follow the show
Please silence your phones
[image: sponsors.png | Thank you to our sponsors | 15s]
[/loop]

[section: Act 1, Scene 1 - The garden]
[note: Lights up on the garden. Wait for the first line from Margaret.]
[clear]

(Birdsong. A garden gate creaks open.)

MARGARET: You're late, Edwin.
EDWIN: The train was late, Margaret. The *train*.
MARGARET: The train is never late. **You** are.

<!-- A line in (parentheses) on its own is shown as a sound or stage direction. -->
(A teacup rattles on a saucer.)

EDWIN: Is that tea for me?
MARGARET: It was. Twenty minutes ago.

<!-- Two lines in one caption: end the first with a backslash, or use [[ ]]. -->
[[
EDWIN: Then I'll drink it cold,
and I'll say nothing more about the train.
]]

(Laughter from the next garden.)

<!-- A line ending in a backslash joins with the next line. -->
MARGARET: Nothing more? \
Now that would be a first.

[note: Edwin exits through the gate. Short pause before Scene 2.]
[blank]

[section: Act 1, Scene 2 - The kitchen]
[clear]
(Inside the kitchen. A kettle whistles.)
MARGARET: Some people wait all afternoon for a cup of tea.
MARGARET: Others make it themselves.
<!-- A name must be in CAPITALS followed by a colon to be shown as a speaker.
     "Note: this" is not a speaker, so it is shown as ordinary text. -->
Note: this line is ordinary text, not a speaker.

<!-- A picture on its own line, in Markdown style. It fills the audience screens. -->
![The garden at dusk](welcome.png)

MARGARET: Look at that sky.

<!-- Text that fills the whole screen, for announcements. -->
[fullscreen: **Please remain seated**]

MARGARET: Well. That went better than expected.

<!-- MUSIC: a song. Everything between [song] and [/song] is sung, so each line is
     shown between music notes (a note symbol is added for you). The title appears
     for the operator only. For a single sung line, put a ~ at the start instead. -->
[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
MARGARET: Whistling all day long,
EDWIN: And the tea goes cold, cold, cold,
[note: Both together on the last line]
Cold as a garden gate
[/song]
~ Tea for two, and two for tea ~
<!-- COUNTERPOINT: two singers at the same time. Both lines appear together, side by
     side on wide screens, from ONE press of the spacebar. -->
[together]
MARGARET: ~ Tea for two, and two for tea ~
EDWIN: ~ Biscuits too, if you please ~
[/together]
<!-- The same thing on one line, with || between the voices. -->
MARGARET: ~ Milk and sugar ~ || EDWIN: ~ Lemon, thank you ~

[section: Intermission]
<!-- INTERMISSION: a loop that starts when the operator presses the spacebar on it.
     SHIFT+Q stops the loop and blanks the screen. The spacebar ends it and sends
     the next line. Back (Backspace) returns to the loop and restarts it. -->
[loop: 12s]
[fullscreen: **INTERMISSION**]
15 minutes
![Refreshments are in the lobby](lobby.png)
[fullscreen: Please return to your seats when the lights flash]
[/loop]

[section: Act 2]
[clear]
(The garden again, the next morning. Rain.)
EDWIN: I thought I'd bring my own umbrella this time.
MARGARET: How very organised of you.
EDWIN: I try.
MARGARET: Do try harder.

[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
