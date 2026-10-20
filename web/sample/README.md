# Sample show

`sample-show.md` is a short made-up scene that demonstrates every show file
feature (captions, speakers, stage directions, songs and sung lyrics (music notes), bold/italic, multi-line
captions, comments, operator notes, sections, images (with a caption, and
without one so that the picture fills the whole screen), fullscreen text, and
named loops for pre-show and intermission). The pictures in `assets/` are plain
colored placeholders.

Try it:

```
make build
./bin/open-caption --dev --show-file web/sample/sample-show.md --assets web/sample/assets --http-addr :8080
```

Then open `http://localhost:8080/operator` (operator password is in
`config.yaml`; the server prints its location at startup) and
`http://localhost:8080/` (audience view).
