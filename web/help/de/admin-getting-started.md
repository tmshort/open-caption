# Erste Schritte

Eine Show einzurichten dauert sechs Schritte. Die Schaltflächenreihe oben auf der Seite folgt ihnen der Reihe nach, und ein Häkchen erscheint, sobald ein Schritt vollständig aussieht. Sie können jederzeit zu jedem Schritt springen.

1. **Shows** — legen Sie eine Show an (oder die Beispiel-Show, um sich umzusehen).
2. **Show-Datei** — fügen Sie Ihre Show-Datei ein oder öffnen Sie sie und beheben Sie, worauf die Seite hinweist.
3. **Bilder** — laden Sie die Bilder hoch, die die Show-Datei verwendet.
4. **Erscheinungsbild** — wählen Sie Farben und Textgröße.
5. **Test** — proben Sie privat, bevor jemand im Saal ist.
6. **Live** — teilen Sie die Adresse mit dem Publikum und starten Sie die Show.

## Wo alles gespeichert wird

Eine Show ist ein Ordner mit ihrer **Show-Datei** (einer einfachen Textdatei, `show.md`), optionalen Übersetzungen davon und ihren Bildern. Die Show-Datei enthält alles zur Show außer den Bildern: Titel, Erscheinungsbild und die Zeilen. Einstellungen, die die Show-Datei nicht erwähnt, stammen aus `config.yaml` und danach aus den eingebauten Werten des Programms.

## Mit einer Show fertig?

Im Schritt Shows legt **Archivieren** eine Show beiseite: Sie bleibt vollständig erhalten, wird aber unter *Archivierte Shows* aufgeführt und kann nicht auf Sendung gehen. **Wiederherstellen** holt sie zurück. **Löschen** entfernt sie endgültig (exportieren Sie sie vorher, falls Sie sie vielleicht noch brauchen). Die Show, die gerade läuft, kann weder archiviert noch gelöscht werden.

## Bevor Sie anfangen

- Dieser Computer muss im selben WLAN sein wie die Telefone und Tablets des Publikums. Siehe [Netzwerk und Verbindung](topic:admin-network).
- Das Administrator- und das Operator-Passwort stehen in `config.yaml`; sie wurden auch beim ersten Start des Programms in der Konsole ausgegeben. Die Operator-Seite hat ein eigenes Passwort, damit Sie es jemand anderem geben können, ohne Zugang zur Einrichtung zu gewähren.
- Die Einrichtung, die Operator-Seite und diese Hilfe gibt es auf Englisch, Spanisch, Französisch, Deutsch, Italienisch und Portugiesisch. Verwenden Sie das Sprachmenü oben auf jeder Seite; die Wahl wird in diesem Browser gemerkt.
- Ihre Arbeit wird auf diesem Computer gespeichert. Mit **Exportieren** im Schritt Shows machen Sie eine Sicherung, die Sie auf einen anderen Computer mitnehmen können. Der Export ist eine `.zip`-Datei mit `show.md` (der Show-Datei samt Titel und Erscheinungsbild), je einer `show.<Sprache>.md` pro Übersetzung und einem Ordner `assets` mit den Bildern; **Importieren** macht aus einem solchen Zip eine neue Show und nimmt auch eine einfache Show-Textdatei (`.md` oder `.txt`, mit beliebigem Namen) als neue Show.

## Neu bei Show-Dateien?

Eine Show-Datei ist einfach Text. Jede Zeile ist ein Untertitel, und ein Druck auf die Leertaste sendet eine Zeile. Für die ganze Arbeit, vom Stück bis zur fertigen Datei (auch das Bearbeiten im eigenen Programm und das Zurückholen der Datei), lesen Sie [Eine Show-Datei vorbereiten](topic:admin-preparing-a-show-file). Die Auszeichnungen stehen in [Eine Show-Datei schreiben](topic:admin-writing-a-show-file). Oder legen Sie die Beispiel-Show an und sehen Sie sich an, wie sie geschrieben ist. Eine Vorlage zum Herunterladen finden Sie unter [/help/sample-show.md](/help/sample-show.md).

## Die Ausstattung planen

Was Sie kaufen sollten, wo der WLAN-Zugangspunkt stehen soll und eine Checkliste für den Tag der Vorstellung finden Sie in der [Bedienungsanleitung](topic:admin-usage-guide). Der [Netzwerkleitfaden für das Haus](topic:admin-venue-network-guide) richtet sich an die Person, die das WLAN des Gebäudes betreut.

## Wie es weitergeht

- [Musicals](topic:admin-musicals)
- [Eine Show-Datei schreiben](topic:admin-writing-a-show-file)
- [Schleifen: vor der Show und Pause](topic:admin-loops)
- [Fehlerbehebung](topic:admin-troubleshooting)
