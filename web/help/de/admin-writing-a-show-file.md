# Eine Show-Datei schreiben

Eine Show-Datei ist eine einfache Textdatei (`.txt` oder `.md`), geschrieben in einer schlichten Form von Markdown. Wenn Sie ein gewöhnliches Theater-Textbuch einfügen, funktioniert es bereits: **Jede Zeile ist ein Untertitel**, und ein Druck auf die Leertaste sendet eine Zeile. Leere Zeilen werden ignoriert; nutzen Sie sie, um die Datei lesbar zu halten.

Lange Zeilen werden auf dem Bildschirm automatisch umgebrochen. Sie entscheiden, wo Sie eine lange Rede in Zeilen teilen, passend zum Stil der Aufführung. Das Programm teilt oder verbindet nie von sich aus Zeilen.

## Formatierung

| Schreiben Sie | Sie erhalten |
|---------------|--------------|
| `**laut**` | **fetten** Text |
| `*Beiseite*` oder `_Beiseite_` | *kursiven* Text |
| `***very loud***` | fetter und kursiver Text |
| `~~old line~~` | durchgestrichener Text (eine Korrektur oder ein durchgestrichener Scherz) |
| `(thunder rumbles)` | eine ganze Zeile in Klammern wird als Geräusch oder Bühnenanweisung gezeigt |
| `HAMLET: To be, or not to be` | eine Sprecher-Kennzeichnung über der Zeile (der Name muss in GROSSBUCHSTABEN stehen, gefolgt von einem Doppelpunkt; eine Gruppe geht auch: `DEE DEE & BARRY:`, `ALL (EXCEPT BARRY):`) |

**Text durchstreichen.** Setzen Sie auf jede Seite zwei Tilden, `~~so~~`, und die Wörter werden durchgestrichen gezeigt. Direkt innerhalb jedes Paars muss ein Zeichen stehen, das kein Leerzeichen ist (`~~ x ~~` bleibt unverändert). Eine gesungene Zeile beginnt mit einer einzelnen `~` und einem Leerzeichen, die beiden werden also nie verwechselt; um eine normale Zeile mit einer Tilde zu beginnen, schreiben Sie `\~`.

Alles andere wird so gezeigt, wie es getippt wurde. Links, Listen, Tabellen und Überschriften werden nicht unterstützt. Rohes HTML wird ebenfalls nicht unterstützt: Wenn Sie `<b>` tippen, erscheint es wörtlich, und die Seite zeigt einen Hinweis, der die Markdown-Form vorschlägt. Unterstreichung, Hoch- und Tiefstellung haben keine Auszeichnung, aber jedes Zeichen, das Sie tippen können, funktioniert so, wie es ist: Schreiben Sie H₂O, x², 1ˢᵗ oder ½ einfach direkt.

```showfile
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Die Namen der Figuren prüfen

Klicken Sie über dem Editor auf **Figuren…**, um alle als Sprecher verwendeten Namen zu sehen, die am häufigsten verwendeten zuerst, mit der Zahl ihrer Zeilen. So finden Sie schnell einen falsch geschriebenen Namen: **ALCE** mit einer Zeile neben **ALICE** mit vierzig wird mit *Meinten Sie ALICE?* markiert. Klicken Sie auf einen Namen, um zu seiner ersten Zeile zu gehen. Eine Gruppe wie `ALICE & BOB:` zählt für jeden ihrer Namen.

## Zu einer früheren Version zurückgehen

Bei jedem Speichern wird die ersetzte Version aufbewahrt (die letzten 50 je Sprache). Klicken Sie über dem Editor auf **Frühere Versionen …**, wählen Sie eine Zeit und klicken Sie auf **Laden**. Das legt nur den alten Text in den Editor; erst mit **Speichern** ändert sich etwas, Sie können also erst einmal nachsehen. **Vergleichen** zeigt Zeile für Zeile, was das Laden einer Version ändern würde. Eine Version, der Sie einen Namen geben (im Feld daneben, oder mit **Behalten**, um zu benennen, was jetzt gespeichert ist), wird nie vergessen, egal wie viele Speicherungen folgen. Der Export enthält die früheren Versionen nicht.

## Show-Einstellungen: der Titel und das Erscheinungsbild

Alles an einer Show außer ihren Bildern wird in der Show-Datei gespeichert, auch ihr **Titel** und das **Erscheinungsbild** der Untertitel. Das steht in einem `[settings]`-Block, meist am Anfang, mit je einem `Name: Wert` pro Zeile:

```showfile
[settings]
title: The Garden Party
background: #000000
text color: #ffffff
speaker name color: #ffffff
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
loop seconds: 10
scrollback lines: 200
[/settings]
```

| Einstellung | Was sie bewirkt |
|-------------|-----------------|
| `title` | Der Name der Show, wie er in der Einrichtung aufgeführt wird |
| `background` | Die Bildschirmfarbe hinter den Untertiteln, wie `#000000` (auch ein Name wie `black` oder `yellow` geht) |
| `text color` | Die Farbe der Untertitel, wie `#ffffff` |
| `speaker name color` | Die Farbe der Sprechernamen (wenn nicht jede Figur eine eigene bekommt) |
| `font` | `system`, `serif`, `mono`, `dyslexic` oder `hyperlegible` |
| `text size` | Eine Prozentzahl von 50 bis 300; 100 ist normal |
| `speaker names` | `label` (eine Farbe), `color` (eine andere Farbe für jede Figur) oder `hidden` |
| `speaker position` | `above` (der Name über der Zeile), `beside` (in derselben Zeile, vor dem Text) oder `column` (in einer Spalte links vom Text) |
| `idle message` | Die Worte, die vor dem Senden der ersten Zeile gezeigt werden |
| `loop seconds` | Wie lange jedes Element einer `[loop]` stehen bleibt, wenn es nichts angibt |
| `scrollback lines` | Wie viele Untertitel ein Zuschauer zurückblättern kann (20 bis 2000) |
| `archived` | `yes`, wenn die Show archiviert ist; die Schaltflächen **Archivieren** und **Wiederherstellen** der Einrichtung fügen die Zeile hinzu und entfernen sie |

**Sie müssen nichts davon schreiben.** Der Schritt **Erscheinungsbild** der Einrichtung füllt den Block für Sie aus (und **Umbenennen** setzt den Titel), und alles, was Sie weglassen, verwendet einfach einen Standardwert. Die Standardwerte kommen in zwei Schichten: Das Programm hat eingebaute Werte, die die Einstellungsdatei `config.yaml` für alle Shows ändern kann (ihr Abschnitt `defaults`), und dann überschreibt jede Show-Datei, was sie festlegt. Das Spezifischste gewinnt: **Show-Datei, dann config.yaml, dann eingebaut**.

Ein paar Regeln: Die Show-Datei einer Übersetzung darf nur `idle message` festlegen (das Erscheinungsbild wird einmal für die ganze Show in der Haupt-Show-Datei festgelegt); eine falsch geschriebene Einstellung oder ein sinnloser Wert wird mit Zeilennummer gemeldet und ignoriert; und Kommentare im Block bleiben nicht erhalten, wenn die Einrichtung ihn neu schreibt, setzen Sie Kommentare also darüber. Mehrere `[settings]`-Blöcke werden zusammengeführt, wobei spätere Werte gewinnen.

## Lieder und Musik

Zeigen Sie bei einem Musical die Liedtexte zwischen Noten (♪). Es gibt drei Wege, vom schnellsten zum flexibelsten:

- **Tippen oder fügen Sie die Noten selbst ein.** ♪ und ♫ funktionieren überall in der Show-Datei. Im Editor fügen die Schaltflächen **♪** und **♫** sie an der Cursorposition ein.
- **Beginnen Sie eine Zeile mit `~`.** Die Zeile wird als gesungener Text behandelt: Sie wird kursiv gezeigt, mit einem ♪ an jedem Ende, und das `~` selbst wird nicht gezeigt. Nur das einleitende `~` ist nötig: das ♪ am Ende wird automatisch ergänzt, und ein abschließendes `~` ist erlaubt, ändert aber nichts. Das funktioniert auch hinter einem Sprechernamen.
- **Fassen Sie ein ganzes Lied in `[song]` und `[/song]`.** Jede Zeile darin wird gesungen, Sie brauchen also kein `~` an jeder. Der Titel nach dem Doppelpunkt ist ein Lesezeichen nur für den Operator. Notizen, clear und blank dürfen weiterhin darin stehen.

```showfile
OLIVER: ~ Where is love? ~
~ Row, row, row your boat

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[note: hold the last chord]
[/song]
```

### Zwei oder mehr Stimmen gleichzeitig (Duette und Kontrapunkt)

Wenn Singende **gleichzeitig** verschiedene Zeilen singen, setzen Sie sie in `[together]` … `[/together]`. Jede Zeile darin wird auf einmal gezeigt, nebeneinander auf einem breiten Bildschirm und gestapelt auf einem Telefon, und der Operator sendet sie mit einem einzigen Druck auf die Leertaste. Jede behält ihren eigenen Sprechernamen. Für einen kurzen Wortwechsel können Sie die Stimmen in eine Zeile schreiben, getrennt durch ` || ` (ein Leerzeichen, zwei senkrechte Striche, ein Leerzeichen). `[duet]` und `[counterpoint]` bedeuten dasselbe wie `[together]`.

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

Stimmen, die **nacheinander** einsetzen, sind getrennte Zeilen: Drücken Sie für jede die Leertaste. Verwenden Sie `[together]` nur für Zeilen, die im selben Moment erscheinen. Ein `[together]`-Block enthält nur Textzeilen, braucht mindestens zwei, und innerhalb eines `[song]` wird jede Stimme gesungen.

Eine Tilde mitten in einer Zeile, etwa `about ~5 minutes`, ist nur Text. Um eine gewöhnliche Zeile mit einer zu beginnen, schreiben Sie `\~`. Weitere Hinweise finden Sie unter [Musicals](topic:admin-musicals).

## Mehrere Zeilen in einem Untertitel

Beenden Sie eine Zeile mit einem Backslash, um sie mit der nächsten zu verbinden, oder setzen Sie die Zeilen zwischen `[[` und `]]`.

```showfile
GHOST: Remember me. \
Remember me.

[[
Row, row, row your boat,
Gently down the stream
]]
```

## Kommentare und Notizen

- `<!-- so -->` ist ein **Kommentar**. Niemand sieht ihn, und er darf sich über mehrere Zeilen erstrecken.
- `[note: Text]` ist eine Erinnerung, die **nur der Operator** sieht, zum Beispiel `[note: hold for applause]`.
- `[section: Act 2, Scene 1]` fügt in der Liste des Operators ein Lesezeichen hinzu. Notizen und Abschnitte verbrauchen nie einen Druck auf die Leertaste.
- Zeilen, die mit `#` beginnen, sind für Überschriften in einer künftigen Version reserviert und werden ignoriert. Verwenden Sie stattdessen einen Kommentar.

```showfile
<!-- Act 1 captions, version 3 -->
[section: Act 1, Scene 1]
[note: lights come up slowly, wait for the first line]
HAMLET: Who's there?
```

## Befehle

Eine Zeile in eckigen Klammern ist ein Befehl. Bei Namen wird nicht zwischen Groß- und Kleinschreibung unterschieden.

| Befehl | Was er bewirkt |
|--------|----------------|
| `[clear]` | Löscht die Untertitel vom Bildschirm |
| `[blank]` | Leert den Bildschirm (die Untertitel bleiben erhalten) |
| `[image: file.png]` | Zeigt ein Bild im Vollbild. Fügen Sie mit `[image: file.png \| text]` eine Beschriftung hinzu |
| `![caption](file.png)` | Dasselbe, in Markdown-Schreibweise |
| `[fullscreen: text]` | Zeigt großen Text im Vollbild. Für mehrere Zeilen verwenden Sie `[fullscreen]` … `[/fullscreen]` |
| `[loop]` … `[/loop]` | Eine sich wiederholende Folge von Vollbild-Elementen für die Zeit vor der Show oder die Pause. Siehe [Schleifen](topic:admin-loops) |
| `[together]` … `[/together]` | Zwei oder mehr gleichzeitig gezeigte Stimmen. Siehe [Lieder und Musik](topic:admin-musicals) |
| `[song: title]` … `[/song]` | Ein Lied: Jede Zeile darin wird als gesungener Text gezeigt. Siehe [Lieder und Musik](topic:admin-musicals) |
| `[section: name]` | Ein Lesezeichen für den Operator |
| `[note: text]` | Eine Notiz für den Operator |
| `[language: es]` | Kennzeichnet eine übersetzte Show-Datei. Siehe [Sprachen](topic:admin-languages) |
| `[settings]` … `[/settings]` | Titel und Erscheinungsbild der Show. Siehe *Show-Einstellungen* oben |

```showfile
[clear]
(Lights up on the garden.)
[image: garden.png | The garden at dusk]
[fullscreen: **Please remain seated**]
[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
[blank]
```

Ein Bild, ein Vollbildtext oder ein leerer Bildschirm bleibt auf den Publikumsbildschirmen, bis der Operator die nächste Zeile sendet.

## Sonderzeichen

Setzen Sie einen Backslash vor ein Zeichen, um es wörtlich anzuzeigen, statt es als Formatierung zu behandeln: `\*`, `\_`, `\#`, `\[`, `\]`, `\:`, `\(`, `\)`, `\!`, `\~` und `\\` (ein Backslash selbst). Ein Untertitel, der mit einem Rautezeichen beginnen muss, wird zum Beispiel als `\# 1 hit` geschrieben, und `HAMLET\: not a speaker` zeigt den Doppelpunkt, ohne eine Sprecher-Kennzeichnung zu erzeugen.

```showfile
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## Was die Warnungen bedeuten

Der Schritt Show-Datei listet Probleme mit Zeilennummern auf. Klicken Sie eines an, um zur Zeile zu springen.

- **„Unbekannter Befehl [x]“** — eine Zeile in eckigen Klammern, die kein Befehl ist. Wenn sie ein Untertitel wie ein Geräuscheffekt sein sollte, schreiben Sie sie mit runden Klammern, etwa `(laughter)`.
- **„Bild … nicht gefunden“** — das Bild ist nicht hochgeladen, oder der Name ist anders geschrieben. Laden Sie es im Schritt Bilder hoch.
- **„[loop] wurde geöffnet, aber nie geschlossen“** — fügen Sie `[/loop]` nach dem letzten Element hinzu. Dasselbe gilt für `[fullscreen]`-Blöcke, `[[`-Blöcke und Kommentare.
- **„HTML wird nicht unterstützt“** — verwenden Sie `**fett**` und `*kursiv*` statt Tags.
- **„darf nicht in einer Schleife stehen“** — Schleifen enthalten nur Bilder und Text.
- **„[together] braucht mindestens zwei Zeilen“** — ein `[together]`-Block mit einer Zeile wird als gewöhnlicher Untertitel gezeigt. **„[together] wurde geöffnet, aber nie geschlossen“** — fügen Sie `[/together]` hinzu; die geschriebenen Zeilen bleiben erhalten.
- **„Dieser Untertitel zeigt 5 Stimmen gleichzeitig“** — fünf oder mehr Stimmen in einem `[together]`-Block (oder einer Zeile mit ` || `) sind viel für ein Telefon, das sie stapelt, und für einen gemeinsamen Bildschirm, der für etwa vier Platz hat und nie scrollt. Teilen Sie ihn in zwei Untertitel.
- **„Unbekannte Einstellung“ / „ist keine Schrift“ / „ist keine Farbe“** — eine Zeile im `[settings]`-Block, die die Einrichtung nicht verwenden kann; sie wird ignoriert und der Standardwert gilt. Die Meldung sagt, was stattdessen zu schreiben ist.
- **„[settings] wurde geöffnet, aber nie geschlossen“** — fügen Sie `[/settings]` hinzu; alles nach einem nicht geschlossenen Block wird als Einstellungen gelesen.
- **„[song] wurde geöffnet, aber nie geschlossen“** — fügen Sie nach dem letzten Liedtext `[/song]` hinzu. Die geschriebenen Texte bleiben trotzdem erhalten.

Fehler hindern Sie nie am Speichern.

## Wenn zwei Personen dieselbe Show-Datei bearbeiten

Speichert jemand anderes die Show-Datei, während Sie sie geöffnet haben, überschreibt das Programm seine Arbeit nicht stillschweigend. Beim Speichern sagt es Ihnen, dass die Show-Datei von jemand anderem geändert wurde, und lässt Sie **deren Version verwenden** (ersetzt, was in Ihrem Editor steht) oder **Ihre trotzdem speichern**. Kopieren Sie vor der Wahl, was Sie behalten möchten. Der Teil mit dem Fehler wird übersprungen, und der Rest der Show-Datei funktioniert weiter.
