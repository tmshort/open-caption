# Sprachen

Eine Show kann ihre Untertitel in mehr als einer Sprache anbieten. Jede Sprache ist eine **eigene Show-Datei**, und der Operator folgt nur der Hauptdatei: Ein Druck auf die Leertaste sendet denselben Moment gleichzeitig in jeder Sprache.

## Eine Sprache hinzufügen

1. Klicken Sie im Schritt **Show-Datei** neben den Sprach-Tabs auf **＋ Sprache hinzufügen**.
2. Wählen Sie die Sprache (oder „Andere Sprache …“ und geben Sie einen kurzen Code wie `cy` ein).
3. Lassen Sie **Mit einer Kopie der Haupt-Show-Datei beginnen** angekreuzt. Sie erhalten eine Kopie in der richtigen Form und müssen nur die Wörter ersetzen.
4. Übersetzen Sie und **speichern** Sie dann.

Jede Übersetzung beginnt mit einer Zeile, die ihre Sprache nennt. Der Editor fügt sie für Sie ein:

```showfile
[language: es | Español]
Hola, Margaret.
```

Der Name nach dem `|` ist das, was Zuschauer im Sprachmenü sehen; schreiben Sie ihn daher in seiner eigenen Sprache (Español, Français, Deutsch).

## Die eine Regel: dieselbe Form behalten

Jede Sprache muss **dieselben Einsätze in derselben Reihenfolge** haben: Ein Untertitel bleibt ein Untertitel, ein `[clear]` bleibt ein `[clear]`, eine Schleife behält dieselbe Zahl von Elementen, und gleichzeitig gezeigte Stimmen bleiben dieselbe Zahl von Stimmen. Nur die Wörter ändern sich. Sprechernamen, Notizen, Abschnittsnamen, Bildnamen und Schleifenzeiten dürfen abweichen.

Passt eine Übersetzung nicht mehr, zeigt der Editor einen Fehler mit der Zeile, in der es schiefgeht, zum Beispiel „Einsatz 4 passt nicht zur Show-Datei auf Englisch: Diese Datei hat hier [clear], und diese hat einen Untertitel“. Eine Sprache, die nicht passt, wird **dem Publikum nicht angeboten**, bis sie korrigiert ist. Die englische (Haupt-)Show-Datei bleibt unberührt, und der Schritt Fahren führt das Problem auf.

Damit das Publikum erkennt, welche Wörter zu welcher Zeile gehören, halten Sie die Zeilen im Gleichschritt: Teilt das Englische eine Rede in drei Zeilen, tut das Spanische es auch, selbst wenn die Sprache sie anders umbrechen würde.

## Was das Publikum sieht

- Jedes Gerät bekommt die beste Sprache, die die Show hat. Es werden die Spracheinstellungen des Telefons oder Browsers angesehen (zum Beispiel erhält Spanisch für Mexiko Spanisch), und wenn nichts passt, wird die Hauptsprache verwendet.
- Ein kleines Steuerelement **🌐** in der oberen linken Ecke listet die Sprachen auf. Wählt man eine aus, wechselt dieses Gerät, und die Adresse endet dann auf `?lang=es`, sodass Sie für jede Sprache einen Link oder QR-Code drucken können.
- Auch die eigenen Wörter der Seite („Verbindung wird wiederhergestellt …“, Textgröße usw.) sind ins Spanische, Französische, Deutsche, Italienische und Portugiesische übersetzt. Andere Sprachen zeigen diese wenigen Beschriftungen auf Englisch; die Untertitel selbst sind immer in der Sprache der Show-Datei.
- Von rechts nach links geschriebene Sprachen (Arabisch, Hebräisch) werden rechts ausgerichtet.

## Weitere Hinweise

- **Bilder:** Alle Sprachen können dieselben Bilder verwenden. Laden Sie für ein Bild mit Text eine Fassung je Sprache hoch und nennen Sie jede in der Show-Datei dieser Sprache.
- **Nachricht vor der Show:** Jede Sprache hat ihre eigene, aufbewahrt in der eigenen Show-Datei dieser Sprache (`idle message:` in deren `[settings]`-Block). Der Schritt Erscheinungsbild hat ein Feld für jede Sprache und schreibt sie für Sie. Die Show-Datei einer Übersetzung kann nichts anderes festlegen: Das Erscheinungsbild wird einmal in der Haupt-Show-Datei festgelegt.
- **Durchsagen des Operators** („Durchsage …“) werden genau wie getippt an alle Sprachen gesendet; tippen Sie sie also in den Sprachen, die Sie brauchen.
- **Eine Show-Datei ändern, während die Show läuft:** Das Hinzufügen oder Korrigieren einer Übersetzung wirkt sofort und behält die Position des Operators. Das Entfernen einer Sprache ist nicht erlaubt, solange die Show live ist.
- **Zählungen:** Die Operator-Seite zeigt, wie viele Geräte jede Sprache lesen, zum Beispiel `👥 83 (en 60 · es 23)`.
- **Sicherungen:** Exportieren speichert jede Sprache; Duplizieren kopiert sie ebenfalls.

Siehe auch [Eine Show-Datei schreiben](topic:admin-writing-a-show-file).
