# Bilder

Laden Sie Bilder im Schritt **Bilder** hoch und verwenden Sie sie in der Show-Datei mit `[image: name.png]`.

- Akzeptierte Typen: PNG, JPG, GIF und WebP. Andere Typen, auch SVG, werden nicht akzeptiert.
- Jede Datei darf bis zu 10 MB groß sein, aber **halten Sie sie unter etwa 1 MB**. Jedes Publikumsgerät lädt das ganze Bild herunter, und 100 Geräte, die gleichzeitig ein großes Bild abrufen, können langsam sein. Etwa 1920 Pixel Breite genügen.
- Das Programm verkleinert Bilder nicht. Browser passen sie an jeden Bildschirm an.
- Kommende Bilder werden im Hintergrund heruntergeladen, bevor sie gezeigt werden, und erscheinen deshalb sofort.

## Namen

Die Show-Datei bezieht sich auf ein Bild über seinen Dateinamen, etwa `vorhang.jpg`. Namen mit Leerzeichen funktionieren. Wenn die Show-Datei einen Namen verwendet, den Sie nicht hochgeladen haben, führt der Schritt Bilder ihn unter „Bilder, die die Show-Datei verwendet, die Sie aber nicht hochgeladen haben“ auf, mit einer Schaltfläche, um die richtige Datei unter diesem Namen hochzuladen. Ein knapper Tippfehler wie `vorhamg.jpg` wird mit einem Vorschlag angezeigt.

## Ersetzen und löschen

Wenn Sie eine Datei mit demselben Namen hochladen, wird sie ersetzt. Ein Bild zu löschen, das die Show-Datei noch verwendet, ist nach einer Warnung erlaubt; die Show-Datei zeigt dann dafür eine Warnung „nicht gefunden“.

## In einer Schleife

Bilder in einer Schleife wechseln sich auf dem Bildschirm ab: `[image: sponsoren.png | 15s]` zeigt dieses 15 Sekunden lang. Siehe [Schleifen](topic:admin-loops).
