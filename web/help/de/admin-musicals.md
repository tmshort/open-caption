# Musicals

Untertitel für ein Musical funktionieren am besten, wenn sich Liedtexte leicht von gesprochenen Zeilen unterscheiden lassen und der Operator der Musik folgen kann. Diese Seite sammelt, was sich bewährt.

## Liedtexte markieren

Liedtexte werden **kursiv zwischen Noten** gezeigt, zum Beispiel ♪ Wo ist die Liebe? ♪. Wählen Sie die Schreibweise, die Ihnen passt:

| Schreiben Sie | Ergebnis |
|---------------|----------|
| `~ Where is love? ~` | ♪ Where is love? ♪ (das abschließende `~` ist optional) |
| `OLIVER: ~ Where is love?` | dasselbe, mit dem Namen der Figur darüber |
| `[song: Title]` … `[/song]` | jede Zeile darin ist ein Liedtext; der Titel ist nur für den Operator |
| `♪ getippt oder eingefügt ♪` | genau das, was Sie getippt haben, ohne Kursivschrift |
| `~` allein in einer Zeile | nur eine Note: ♪, für eine Instrumentalpassage |

Verwenden Sie `[song]` für lange Nummern und `~` für eine einzelne gesungene Zeile mitten in einer Szene.

In der Einrichtung fügen die Schaltflächen **🎵 Lied**, **♪** und **♫** über dem Editor der Show-Datei diese für Sie ein. Die Zeichen ♪ und ♫ lassen sich auch von überall kopieren, weil sie gewöhnlicher Text sind.

## Zwei Singende gleichzeitig (Kontrapunkt und Duette)

Wenn verschiedene Singende **gleichzeitig** verschiedene Worte singen, zeigen Sie sie mit `[together]` gemeinsam auf den Publikumsbildschirmen:

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]
```

- Beide Zeilen erscheinen **gleichzeitig**, und der Operator drückt die Leertaste **einmal**.
- Auf einem breiten Bildschirm stehen die Stimmen nebeneinander, jede mit einem Balken an ihrem linken Rand und dem Namen ihrer Figur darüber; auf einem Telefon werden sie gestapelt.
- Fügen Sie so viele Stimmen hinzu, wie Sie brauchen, zum Beispiel ein Terzett oder eine gesprochene Zeile über einer gesungenen.
- Für eine schnelle zweistimmige Zeile können Sie `ANNA: ~ I love you ~ || KRIS: ~ Me too ~` in einer einzigen Zeile schreiben.
- Stimmen, die **nacheinander einsetzen** (ein Kanon oder eine zweite Person, die antwortet), sind gewöhnliche getrennte Zeilen: je ein Druck auf die Leertaste. Schreiben Sie die erste Stimme und dann die zweite, wenn sie einsetzt.
- Halten Sie jede Stimme kurz. Drei oder vier Wörter pro Stimme sind gleichzeitig leichter zu lesen als ein ganzer Satz.
- Verwenden Sie Sprechernamen, damit das Publikum die Stimmen unterscheiden kann.

## Tipps für gute Liedtext-Untertitel

- **Eine Zeile pro gesungener Phrase.** Brechen Sie Zeilen dort um, wo die Singenden atmen, damit der Druck des Operators auf die Leertaste zur Musik passt. Kürzere Zeilen sind leichter zu verfolgen.
- **Verwenden Sie Sprechernamen für Duette** und für Zeilen, die verschiedene Figuren der Reihe nach singen. Lassen Sie sie bei einem Chor weg.
- **Instrumentalpassagen:** Markieren Sie sie mit `~` allein oder mit `(Das Orchester spielt)`. Eine Zeile in Klammern wird als Bühnenanweisung gezeigt.
- **Halten Sie gesprochene und gesungene Zeilen auseinander.** Eine Sprecherzeile gefolgt von Liedtext mit `~` liest sich klar: Das Publikum sieht, wie sich Noten und Kursivschrift ändern.
- **Wiederholungen:** Schreiben Sie die Zeilen jedes Mal aus. Der Operator drückt die Leertaste in beiden Fällen einmal pro Zeile.
- **Leeren zwischen Nummern:** Setzen Sie `[clear]` nach der letzten Zeile eines Lieds, damit die nächste Szene auf einem leeren Bildschirm beginnt.
- **Erinnerungen für den Operator:** `[note: Tempo zieht nach dem Tonartwechsel an]` wird nur dem Operator gezeigt.
- **Proben Sie mit der echten Musik.** Im Schritt Test kann der Operator das Drücken der Leertaste zur Aufnahme der Besetzung üben.

## Ein Beispiel

```showfile
[section: Act 1, Number 4]
[clear]
(The orchestra begins.)
[song: Where Is Love]
OLIVER: Where is love?
Does it fall from skies above?
[note: wait for the held note]
Is it underneath the willow tree
That I've been dreaming of?
[/song]
[clear]
MRS. CORNEY: That's quite enough of that!
```

Siehe auch [Eine Show-Datei schreiben](admin-writing-a-show-file.md).
