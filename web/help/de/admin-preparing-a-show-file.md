# Eine Show-Datei vorbereiten

Diese Seite ist für die Person, die das Stück in Untertitel verwandelt. Sie müssen die Vorstellung nicht fahren und nichts über das Netzwerk wissen. Eine Show-Datei ist eine gewöhnliche Textdatei: Sie können sie mit jedem Programm auf jedem Computer schreiben und in die Einrichtung holen, wenn sie fertig ist. [Eine Show-Datei schreiben](topic:admin-writing-a-show-file) listet alle Auszeichnungen auf, die Sie verwenden können; diese Seite beschreibt, wie man die Arbeit angeht.

## Die Kurzfassung

1. Bringen Sie das Stück in eine reine Textdatei. Beginnen Sie auf Wunsch mit der Beispiel-Show unter [/help/sample-show.md](/help/sample-show.md).
2. Machen Sie **eine Zeile der Datei zu einem Untertitel**: einem Druck auf die Leertaste.
3. Fügen Sie hinzu, was Publikum und Bediener hilft (Sprechernamen, Geräuschbeschreibungen, Abschnitte, Notizen).
4. Holen Sie die Datei in die Einrichtung, lesen Sie die Problemliste, beheben Sie die Probleme und speichern Sie.
5. Proben Sie im Schritt Test und beheben Sie, was Ihnen auffällt.

## Welche Art von Datei

- Jede reine Textdatei funktioniert: ein Name, der auf `.txt` oder `.md` endet, und überhaupt jeder Name, etwa `Unsere Stadt - Untertitel v3.txt`.
- Speichern Sie sie als **UTF-8**, wenn Ihr Programm danach fragt. Dann bleiben Umlaute, geschwungene Anführungszeichen und Musiknoten erhalten. Eine als Windows-„ANSI“ (Windows-1252) gespeicherte Datei wird ebenfalls verstanden. Textverarbeitungsdateien (`.doc`, `.docx`, `.pdf`) sind keine Textdateien: Kopieren Sie den Text zuerst in einen einfachen Texteditor, oder wählen Sie Speichern unter und dann reinen Text.
- Windows-Zeilenenden (CRLF) und Mac- oder Linux-Zeilenenden (LF) funktionieren beide, ebenso das unsichtbare Zeichen, das manche Editoren an den Dateianfang setzen.
- Eine `.zip` wird nur für eine vollständige Sicherung gebraucht, die die Bilder einschließt. Die Show-Datei ist nur der Text.

## Entscheiden, wo jeder Untertitel beginnt und endet

Das Programm trennt oder verbindet Ihre Zeilen nie. Ein Untertitel ist genau das, was Sie in eine Zeile getippt haben, Sie bestimmen also, wie viel das Publikum auf einmal liest.

- Streben Sie Zeilen an, die sich mit einem Blick bequem lesen lassen: ein Satz oder ein Satzteil. Sehr lange Reden sollten an natürlichen Pausen getrennt werden, wo der Schauspieler Luft holt.
- Leere Zeilen haben keine Bedeutung. Nutzen Sie sie großzügig zwischen den Reden, damit die Datei übersichtlich bleibt.
- Sollen zwei oder mehr Zeilen zusammen erscheinen, verbinden Sie sie mit einem Backslash am Zeilenende oder setzen Sie sie zwischen `[[` und `]]`.
- Mit `HAMLET: Zeile` zeigen Sie, wer spricht. Schreiben Sie den Namen in Großbuchstaben, dann einen Doppelpunkt.
- Beschreiben Sie Geräusche und Regieanweisungen in Klammern, in einer eigenen Zeile, zum Beispiel `(Donnergrollen)`. Sie werden dem Publikum in einem anderen Stil gezeigt.

```showfile
[section: Act 1]
[note: house lights down, then start]
GEORGE: Morning, Emily.
EMILY: Oh, hello, George.
(a train whistle, far away)
```

## Was nur der Bediener sieht

Merken Sie sich Ihre Stelle und hinterlassen Sie Erinnerungen, ohne das Publikum zu stören: `[section: Akt 2]` fügt ein Lesezeichen in die Liste des Bedieners ein, und `[note: Applaus abwarten]` zeigt eine Erinnerung. Keines von beiden verbraucht einen Druck auf die Leertaste. Ein Kommentar zwischen `<!--` und `-->` ist für Sie und wird von niemandem gesehen.

## In Ihrem eigenen Editor arbeiten

Sie müssen nie im Editor des Schritts Show-Datei tippen. Stattdessen:

1. Drücken Sie im Schritt **Show-Datei** auf **Show-Datei herunterladen**. Sie erhalten eine reine `.md`-Datei.
2. Öffnen Sie sie in Ihrem eigenen Programm, bearbeiten und speichern Sie sie. Behalten Sie sie als reinen Text.
3. Verwenden Sie zurück im Schritt Show-Datei **Datei öffnen** und wählen Sie Ihre Datei. Ihr Text ersetzt den im Editor, und die Problemliste wird sofort aktualisiert. Erst mit **Speichern** wird etwas endgültig geändert.
4. Drücken Sie **Speichern**. Die ersetzte Version bleibt unter **Frühere Versionen** erhalten, wo Sie sie mit der neuen vergleichen können, ein Fehler lässt sich also immer rückgängig machen.

Dasselbe gilt für eine Übersetzung: Wählen Sie zuerst den Sprachreiter, laden Sie dann die Datei dieser Sprache herunter und öffnen Sie sie. Sie können eine Datei auch auf den Editor ziehen.

Um aus einer Datei eine **neue Show** zu beginnen, verwenden Sie im Schritt Show **Show importieren** und wählen die Datei. Sie wird zu einer neuen Show, benannt nach der Zeile `title:` in ihrem Block `[settings]`, oder nach dem Dateinamen, falls es keine gibt. Bilder sind nicht Teil einer reinen Textdatei; laden Sie sie im Schritt Bilder hoch, mit denselben Namen wie in Ihren `[image: ...]`-Zeilen.

## Ihre Arbeit prüfen

- Der Schritt Show-Datei listet jedes Problem mit seiner Zeilennummer auf. Klicken Sie eines an, um dorthin zu springen. Fehler hindern Sie nie am Speichern, aber ein fehlerhaftes Element wird übersprungen, beheben Sie sie also vor der Vorstellung.
- Das Feld neben dem Editor zeigt, was der Bediener durchläuft, eine Zeile pro Druck auf die Leertaste. Prüfen Sie, ob die Anzahl der Zeilen Ihren Erwartungen entspricht.
- Führen Sie den Schritt **Test** aus, um die Untertitel so zu sehen, wie das Publikum sie sieht, auf einem Telefon ebenso wie auf einem großen Bildschirm.
- Wenn auch andere an der Show-Datei arbeiten, lesen Sie *Wenn zwei Personen dieselbe Show-Datei bearbeiten* in [Eine Show-Datei schreiben](topic:admin-writing-a-show-file).

## Bevor Sie übergeben

- Lesen Sie die Datei einmal von oben nach unten so, wie es das Publikum täte, nicht wie ein Korrekturleser.
- Prüfen Sie, dass jeder `[image: ...]`-Name zu einem hochgeladenen Bild passt und dass jedes `[loop]`, `[song]`, `[together]` und `[fullscreen]` seine abschließende Zeile hat.
- Geben Sie dem Bediener eine Kopie der Datei und die Abschnittsnamen, damit er seine Stelle findet. [Lieder und Musik](topic:admin-musicals), [Schleifen](topic:admin-loops) und [Sprachen](topic:admin-languages) erklären diese Teile im Einzelnen.
