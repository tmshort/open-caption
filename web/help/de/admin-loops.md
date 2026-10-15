# Schleifen: vor der Show und Pause

Eine **Schleife** ist eine Folge von Vollbild-Bildern und -Texten, die sich wiederholt, bis der Operator weitergeht. Verwenden Sie eine vor Beginn der Show und eine weitere in der Pause.

```showfile
[loop: 10s]
[image: sponsors.png | 15s]
**Welcome!**
Please silence your phones
[/loop]
```

Jede Zeile in der Schleife ist ein Element:

- eine einfache Textzeile wird als Vollbildtext gezeigt;
- `[image: datei.png]`, `![Beschriftung](datei.png)` und `[fullscreen: Text]` funktionieren ebenfalls;
- `[blank: 3s]` ist eine kurze Pause auf leerem Bildschirm;
- fügen Sie am Ende eines Elements `| 15s` an, um zu ändern, wie lange es stehen bleibt;
- `[loop: 10s]` legt die Zeit für Elemente fest, die keine angeben. Geben Sie keine Zeit an, wird der Standardwert aus `config.yaml` (10 Sekunden) verwendet.

Untertitel, `[clear]`, Abschnitte, Notizen und andere Schleifen dürfen nicht in einer Schleife stehen.

## Wann eine Schleife startet und stoppt

- Wenn das **Erste** in der Show-Datei eine Schleife ist, startet sie von selbst, sobald die Show gestartet wird; das Publikum sieht die Vorab-Anzeige, während es seine Plätze sucht.
- Jede andere Schleife (zum Beispiel in der Pause) startet, wenn der Operator die Leertaste darauf drückt.
- Die Leertaste beendet die Schleife und sendet die nächste Zeile. **Umschalt+Q** stoppt die Schleife und lässt den Bildschirm leer. **Zurück** kehrt zur Schleife zurück und startet sie beim ersten Element neu.

Die Liste des Operators zeigt eine Schleife als einen Eintrag, der sich aufklappen lässt, um ihre Elemente aufzulisten. Siehe [Schleifen für den Operator](topic:operator-loops).
