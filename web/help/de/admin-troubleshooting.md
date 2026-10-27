# Fehlerbehebung

## Die Schaltfläche Fehlersuche

Die Schaltfläche **Fehlersuche** oben in der Einrichtung, neben Einstellungen, ist die erste Anlaufstelle, wenn sich Leute nicht verbinden können oder etwas nicht stimmt. Sie zeigt, was das Programm weiß: ob dieser Computer eine Netzwerkadresse zum Weitergeben hat, welche Ports (HTTP und HTTPS) der Webserver tatsächlich bekommen hat, ob HTTPS und sein Zertifikat in Ordnung sind, ob der `.local`-Name bekannt gegeben wird, ob die Einstellungsdatei gespeichert werden kann, ob jede Sprache zur Hauptdatei der Show passt und wie viele Geräte verbunden sind. Jede Prüfung ist mit **OK**, **Hinweis**, **Prüfen** oder **Problem** gekennzeichnet; was sich beheben lässt, hat eine Schaltfläche zur richtigen Einstellung. Darunter stehen die neuesten Warnungen und Fehler aus dem Protokoll, ein Link zum ganzen Protokoll und das Support-Paket.

Wenn das Programm neue Warnungen oder Fehler protokolliert, erscheint eine Zahl auf der Schaltfläche, bis Sie sie öffnen.

## Niemand kann die Adresse öffnen

- Prüfen Sie, dass das Telefon im Gast-WLAN des Hauses ist, nicht im Mobilfunknetz.
- Prüfen Sie, dass die Adresse mit der aus dem Schritt **Teilen** übereinstimmt. Hat sie eine Portnummer, geben Sie sie mit an.
- Fragen Sie das Haus, ob Geräte im Gast-WLAN diesen Computer erreichen dürfen.
- Die Firewall Ihres Computers blockiert das Programm vielleicht. Erlauben Sie es unter Windows in **privaten Netzwerken**, wenn Sie gefragt werden. Erlauben Sie unter macOS eingehende Verbindungen dafür.
- Öffnen Sie die Adresse auf diesem Computer mit `localhost`, um das Programm selbst zu prüfen.

## Die Zahl der verbundenen Geräte ist niedriger als erwartet

Geräte, die eingeschlafen sind, trennen sich und verbinden sich beim Aufwachen von selbst wieder. Bitten Sie die Leute, die automatische Sperre für die Show auszuschalten, falls ihr Bildschirm dunkel wird.

## Bilder erscheinen langsam

Große Bilder brauchen in einem ausgelasteten WLAN Zeit. Suchen Sie im Schritt Bilder nach als „groß“ gekennzeichneten Dateien und ersetzen Sie sie durch kleinere.

Die vollständige Liste der Prüfungen auf der Netzwerkseite finden Sie im [Netzwerkleitfaden für das Haus](../../../docs/i18n/de/venue-network-guide.md).

## Für das Publikum fehlt eine Sprache

Öffnen Sie den Schritt Show-Datei und sehen Sie sich den Tab dieser Sprache an: Ein ⚠ bedeutet, dass ihre Show-Datei nicht mehr dieselben Einsätze hat wie die Hauptdatei. Beheben Sie den angezeigten Fehler; die Sprache kommt zurück, sobald Sie speichern. Siehe [Sprachen](admin-languages.md).

## Probleme in der Show-Datei

Siehe „Was die Warnungen bedeuten“ unter [Eine Show-Datei schreiben](admin-writing-a-show-file.md).

## Passwörter

Um eines der beiden Passwörter zu ändern, verwenden Sie **Einstellungen** oben auf der Einrichtungsseite. Das neue Passwort gilt sofort und wird in `config.yaml` gespeichert. Wer mit einem geänderten Passwort angemeldet ist, wird abgemeldet; Sie bleiben angemeldet. Dasselbe Fenster kann das aktuelle Operator-Passwort anzeigen, damit Sie es dem Operator geben können. Ein optionales **Zuschauer-Passwort** lässt ein zweites Gerät, etwa das Tablet der Inspizienz, die Mitlese-Seite (`/follow`) nur zum Ansehen öffnen: Sie zeigt die Zeilen, die Notizen und die Position, und es lässt sich nichts senden. Das Operator- und das Administrator-Passwort öffnen diese Seite ebenfalls.

Wenn Sie sich gar nicht anmelden können, stehen beide Passwörter in `config.yaml` im Datenordner des Programms. Öffnen Sie die Datei mit einem beliebigen Texteditor. Ihr Speicherort wird beim Start des Programms ausgegeben.

## Das Protokoll ansehen

**Fehlersuche → Das ganze Protokoll öffnen…** zeigt, was das Programm zuletzt protokolliert hat. Wählen Sie *Nur Probleme*, um Warnungen und Fehler zu sehen, etwa abgelehnte Anmeldungen oder einen im Netzwerk bereits verwendeten Namen. Dasselbe Fenster legt fest, wie viele Details protokolliert werden (**Protokoll-Detailgrad**); *Debug* fügt jede Anfrage hinzu, was hilft, wenn etwas Merkwürdiges passiert. Passwörter werden nie ins Protokoll geschrieben.

## Hilfe holen

Laden Sie das **Supportpaket** über die Schaltfläche **Fehlersuche** oben in der Einrichtung herunter. Es enthält die Einstellungen (ohne Passwörter), aktuelle Protokollzeilen und grundlegende Angaben zu diesem Computer und lässt sich leicht an die Person schicken, die das Programm betreut.
