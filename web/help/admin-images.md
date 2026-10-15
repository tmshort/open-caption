# Images

Upload pictures in the **Images** step and use them in the show file with `[image: name.png]`.

- Accepted types: PNG, JPG, GIF and WebP. Other types, including SVG, are not accepted.
- Each file can be up to 10 MB, but **keep them under about 1 MB**. Every audience device downloads the full picture, and 100 devices fetching a large picture at once can be slow. About 1920 pixels wide is plenty.
- The program does not resize pictures. Browsers shrink them to fit each screen.
- Upcoming pictures are downloaded in the background before they are shown, so they appear instantly.

## Names

The show file refers to a picture by its file name, such as `curtain.jpg`. Names with spaces work. If the show file uses a name you have not uploaded, the Images step lists it under "Images the show file uses but you have not uploaded", with a button to upload the right file under that name. A near miss such as `curtian.jpg` is pointed out with a suggestion.

## Replacing and deleting

Uploading a file with the same name replaces it. Deleting a picture that the show file still uses is allowed after a warning; the show file then shows a "not found" warning for it.

## In a loop

Pictures in a loop take turns on screen: `[image: sponsors.png | 15s]` shows that one for 15 seconds. See [Loops](topic:admin-loops).
