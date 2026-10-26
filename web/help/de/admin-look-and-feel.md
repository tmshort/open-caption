# Erscheinungsbild

Der Schritt **Erscheinungsbild** legt fest, wie Untertitel auf jedem Publikumsgerät aussehen. Die Voreinstellungen, weißer Text auf schwarzem Hintergrund, sind in einem dunklen Saal am leichtesten zu lesen und für die Barrierefreiheit am besten.

- **Farben:** Ein Kontrastwert zeigt, wie gut lesbar die Kombination ist. Streben Sie 7:1 oder besser an; die Seite warnt, wenn er zu niedrig ist.
- **Textgröße:** skaliert den gesamten Text. Die Geräte passen den Text ihrem Bildschirm an, es ist also eine relative Änderung, keine Größe in Punkt. Der Regler reicht von 50 bis 300 Prozent; 100 ist normal.
- **Schrift:** schlicht (die übliche Schrift des Geräts), Serifenschrift, Festbreitenschrift oder **legasthenie-freundlich (OpenDyslexic)**, eine Schrift mit schwereren Buchstabenunterseiten und markanten Buchstabenformen, die manche Leser mit Legasthenie leichter finden. Sie ist im Programm eingebaut, braucht also kein Internet, und ein Gerät lädt sie (etwa 115 KB) nur, wenn sie verwendet wird. Die Zeilen werden damit etwas höher. Wählen Sie **hohe Lesbarkeit (Atkinson Hyperlegible)** für eine schlichte, vom Braille Institute für Sehbehinderung entworfene Schrift mit leicht unterscheidbaren Buchstaben (zum Beispiel großes I, kleines l und die Ziffer 1). Beide Schriften sind im Programm eingebaut, und ein Gerät lädt eine nur, wenn es sie verwendet (etwa 17 KB bei Atkinson Hyperlegible). Zuschauer können über das kleine Steuerelement **Aa** auch für ihr eigenes Gerät eine der beiden Schriften wählen, egal welche Schrift Sie hier gewählt haben.
- **Sprechernamen:** in einer Farbe (der *Farbe der Sprechernamen*), in einer **anderen Farbe für jede Figur** oder ausgeblendet. Die Figurenfarben werden automatisch gewählt, in der Reihenfolge, in der die Figuren zum ersten Mal sprechen, aus einer Palette, die für farbenblinde Zuschauer funktioniert, und so angepasst, dass sie auf Ihrem Hintergrund lesbar bleiben.
- **Position des Sprechernamens:** über der Zeile (Standard), **neben dem Text** in derselben Zeile (fett) oder **in einer Spalte** links vom Text, wie in einem Drehbuch. Gleichzeitig gezeigte Stimmen setzen den Namen immer neben ihren eigenen Text.
- **Nachricht vor der Show:** wird gezeigt, bis die erste Zeile gesendet wird, falls die Show-Datei nicht mit einer Schleife beginnt.

Änderungen werden mit **Erscheinungsbild speichern** gesichert und erreichen verbundene Geräte sofort. **Erscheinungsbild speichern schreibt Ihre Auswahl in den `[settings]`-Block am Anfang der Show-Datei** (siehe [Eine Show-Datei schreiben](topic:admin-writing-a-show-file)), sodass die Show-Datei ihr Erscheinungsbild überallhin mitnimmt, wohin sie kopiert wird; Sie können diesen Block auch von Hand bearbeiten. Alles, was die Show-Datei nicht festlegt, verwendet die Standardwerte aus `config.yaml` (deren Abschnitt `defaults`), die ihrerseits von den eingebauten Werten des Programms ausgehen. Hat der Editor der Show-Datei nicht gespeicherte Änderungen, speichern Sie diese, bevor Sie das Erscheinungsbild speichern.

## Ein Bildschirm für den ganzen Saal

Wenn ein Browser Untertitel auf einem Projektor oder großen Monitor zeigt, öffnen Sie die Publikumsadresse mit `?display=projector` am Ende. Der Text ist größer, der Mauszeiger ist verborgen, und sonst wird nichts angezeigt. Ein gemeinsamer Bildschirm scrollt nie. Ist der neueste Untertitel höher als der Bildschirm (etwa bei vielen Stimmen gleichzeitig), wird der Text verkleinert, bis er passt.

## Frühere Zeilen noch einmal lesen

Jeder kann die Untertitel nach oben scrollen, um etwas noch einmal zu lesen, das er verpasst hat. Neue Zeilen warten dann hinter einer Schaltfläche **↓ Zurück zu Live (3)** (sie zählt, was hereingekommen ist), statt die Seite wegzuziehen, und die älteren Zeilen blenden nicht mehr aus, solange jemand sie liest. Ein Tipp auf die Schaltfläche oder das Scrollen nach unten führt zu den Live-Untertiteln zurück. Ein Bildschirm mit `?display=projector` ist nie scrollbar.

## Bedienelemente für das Publikum

Zuschauer müssen nie etwas tun. Ein kleines, blasses Steuerelement **Aa** in der oberen rechten Ecke der Publikumsseite lässt jede Person ihren eigenen Text größer oder kleiner machen, auf dunkel auf hell umschalten, die legasthenie-freundliche oder die gut lesbare Schrift wählen oder **Bildschirm bleibt an** ausschalten (die Seite hält den Bildschirm eines Telefons während der Show wach, und wer wenig Akku hat, möchte das vielleicht nicht), aber nur auf dem eigenen Gerät.
