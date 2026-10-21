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
- eine Zeit darf höchstens 3600 Sekunden (eine Stunde) betragen; eine längere wird darauf gekürzt, mit einer Warnung.

Untertitel, `[clear]`, Abschnitte, Notizen und andere Schleifen dürfen nicht in einer Schleife stehen.

## Wann eine Schleife startet und stoppt

- Wenn das **Erste** in der Show-Datei eine Schleife ist, startet sie, wenn der Bediener die Show auf Sendung bringt (**START**), sodass das Publikum die Zeit vor der Show sieht, während die Leute ihre Plätze suchen.
- Jede andere Schleife (zum Beispiel in der Pause) startet, wenn der Operator die Leertaste darauf drückt.
- Die Leertaste beendet die Schleife und sendet die nächste Zeile. **Umschalt+Q** stoppt die Schleife und lässt den Bildschirm leer. **Zurück** kehrt zur Schleife zurück und startet sie beim ersten Element neu.

Die Liste des Operators zeigt eine Schleife als einen Eintrag, der sich aufklappen lässt, um ihre Elemente aufzulisten. Siehe [Schleifen für den Operator](topic:operator-loops).

## Ein Countdown-Timer

Geben Sie einer Schleife einen **Timer**, dann können die Bildschirme des Publikums anzeigen, wie viel Zeit noch bleibt, live heruntergezählt („Weiter in 15 min“). Schreiben Sie `timer` in die Optionen der Schleife und setzen Sie `{timer}` in einen beliebigen Text dieser Schleife, wo die Zeit erscheinen soll.

```showfile
[loop: timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
![Refreshments are in the lobby](lobby.png)
[/loop]
```

Der Wert eines Timers ist eines von zwei Dingen:

- Eine **Anzahl Minuten**, etwa `timer 15` oder `timer 15 minutes`: für eine Pause. Der Countdown beginnt in dem Moment, in dem die Schleife startet (wenn der Bediener darauf die Leertaste drückt), und beginnt wieder mit der vollen Zeit, wenn der Bediener in die Schleife zurückgeht.
- Eine **Uhrzeit**, etwa `timer 7:30pm`, `timer 7pm` oder `timer 19:30`: für die Zeit vor der Show. Die Bildschirme zählen bis zu dieser Uhrzeit heute herunter. Verwendet werden Uhr und Zeitzone des Computers, auf dem das Programm läuft, prüfen Sie sie also. Eine bereits vergangene Uhrzeit zeigt 0:00.

Optionen teilen sich die Klammern, durch Kommas getrennt, in beliebiger Reihenfolge.

```showfile
[loop: 10s, timer 7:30pm]
**Welcome!**
The show begins in {timer}
[/loop]
```

- Die verbleibende Zeit wird in Minuten angezeigt, aufgerundet und ohne Sekunden (`15 Min.`), mit Stunden, wenn es mehr als eine gibt (`1 Std. 5 Min.`), in der Sprache des Geräts. Bei null bleibt sie bei `0 Min.`; die Schleife läuft weiter, bis der Bediener fortfährt.
- Nur Elemente, deren Text `{timer}` enthält, zeigen ihn, Sie können also eine Mitteilung und den Countdown abwechseln.
- Ein Telefon mit falsch gehender Uhr zeigt trotzdem die richtige Restzeit, weil die Uhr des Programms verwendet wird.
- Das Speichern der Show-Datei während einer Pause startet den Countdown nicht neu, ebenso wenig ein Neustart des Programms.
- In einer Übersetzung muss die Schleife denselben Timer haben wie die Haupt-Show-Datei; übersetzen Sie die Wörter um `{timer}`.
- Der Bediener sieht die Restzeit auch im grünen Banner. Ein `{timer}` in einer Schleife ohne Timer wird wie geschrieben angezeigt, mit einer Warnung.

## Eine Schleife benennen

Geben Sie einer Schleife einen **Namen**, und der Publikumsbericht beschriftet sie damit in seinem Diagramm. Ohne Namen rät der Bericht nach der Stelle der Schleife (vor der ersten Zeile: *Vor der Show*; zwischen Zeilen: *Pause*; nach der letzten Zeile: *Nach der Show*).

```showfile
[loop: 12s, name Intermission]
[fullscreen: **INTERMISSION**]
Back soon
[/loop]
```

- Der Name ist freier Text mit höchstens 40 Zeichen, ohne Kommas. Das Publikum sieht ihn nie.
- Er teilt sich die Klammern mit den anderen Optionen, durch Kommas getrennt, in beliebiger Reihenfolge: `[loop: 10s, name Act 1 break, timer 15 minutes]`.
- Es wird nur der Name aus der Show-Datei der Hauptsprache verwendet; eine Übersetzung kann ihren Schleifen andere Namen geben oder keine.
