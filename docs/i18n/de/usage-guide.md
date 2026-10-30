# Bedienungsanleitung für open-caption

open-caption zeigt bei einem Theaterstück oder Musical Live-Untertitel auf den Telefonen, Tablets und Bildschirmen des Publikums. Eine freiwillige Person, der **Operator**, folgt der Show und drückt bei jeder Zeile die Leertaste; jedes Gerät im WLAN des Hauses zeigt die Zeile einen Moment später. Auf den Geräten des Publikums wird nichts installiert, und niemand muss etwas anfassen.

Diese Anleitung behandelt, was man kaufen und wie man es aufstellen soll ([Aufbau](#aufbau-empfehlungen)), wie man eine Show einrichtet, wie man sie fährt, und enthält eine Checkliste für den Vorstellungstag. Die **Hilfe**-Schaltfläche im Programm behandelt dasselbe in kleineren Stücken, und [venue-network-guide.md](venue-network-guide.md) ist für die Person geschrieben, die das WLAN des Gebäudes verwaltet.

Inhalt: [Die Bestandteile](#die-bestandteile) · [Aufbau-Empfehlungen](#aufbau-empfehlungen) · [Installation und Start](#installation-und-start) · [Eine Show einrichten](#eine-show-einrichten) · [Die Show fahren](#die-show-fahren) · [Checkliste für den Vorstellungstag](#checkliste-fur-den-vorstellungstag) · [Fehlerbehebung](#fehlerbehebung) · [Wartung](#wartung) · [Kurzkarte für den Operator](#kurzkarte-fur-den-operator)

## Die Bestandteile

| Bestandteil | Was es ist |
|-------------|------------|
| **Der Server** | Das Programm `open-caption`, das auf einem Computer läuft (macOS, Windows oder Linux). Es verwaltet die Show-Datei und sendet die Zeilen an alle. |
| **Einrichtungsseite** (`/admin`) | Wo Freiwillige eine Show vorbereiten: Show-Datei, Bilder, Erscheinungsbild, Probe. Administrator-Passwort. |
| **Operator-Seite** (`/operator`) | Wo der Operator der Show folgt und Zeilen sendet. Operator-Passwort. Der Schritt **Teilen** der Einrichtung und sein Menülink öffnen sie. |
| **Zuschauer-Seite** (`/follow`) | Ein zweites Gerät, etwa das Tablet der Inspizienz, das die Show nur ansieht: Es sieht die Zeilen, Notizen und die Position und kann nichts senden. Optionales Zuschauer-Passwort; das Operator- und das Admin-Passwort öffnen sie ebenfalls. Der Schritt **Teilen** der Einrichtung führt sie auf. |
| **Publikumsseite** (`/`) | Was das Publikum auf seinen Geräten öffnet. Kein Passwort. Ein Browser an einem Projektor oder großen Monitor kann dieselbe Seite für den ganzen Saal zeigen. |

Alles läuft auf einem Computer. Publikum, Operator und die Freiwilligen der Einrichtung verwenden gewöhnliche Webbrowser.

## Aufbau-Empfehlungen

### Die Kurzfassung

1. **Verwenden Sie einen eigenen WLAN-Zugangspunkt für die Untertitel**, es sei denn, das Gastnetz des Hauses funktioniert nachweislich (siehe unten). Das beseitigt das größte Risiko: ein Netzwerk, das Sie nicht kontrollieren.
2. **Schließen Sie den Server per Kabel** an diesen Zugangspunkt an, mit reservierter Adresse.
3. **Verwenden Sie einen für Menschenmengen gebauten Zugangspunkt** (für über 100 Geräte ausgelegt), aufgestellt im Zuschauerbereich.
4. **Lassen Sie den Server-Computer während der Show angesteckt, wach und sonst unbeschäftigt.**
5. **Drucken Sie den QR-Code** und legen Sie ihn ins Programmheft und hängen Sie ihn im Foyer auf.
6. **Proben Sie mit ein paar echten Telefonen** im echten Saal vor der Premiere.

### Netzwerkoptionen

| Option | Wann sie sich eignet | Hinweise |
|--------|----------------------|----------|
| **A. Eigener Zugangspunkt (empfohlen)** | Jedes Haus, besonders wenn Sie das Netzwerk des Gebäudes nicht ändern können | Sie bringen einen Router oder Zugangspunkt mit. Das Publikum verbindet sich mit dessen WLAN (zum Beispiel „Theater Captions“). Eine Internetverbindung ist nicht nötig. Sie kontrollieren Kapazität, Reservierungen und Firewalls. |
| **B. Das Gast-WLAN des Hauses** | Das Netzwerkpersonal des Hauses bestätigt, dass Gastgeräte Ihren Computer erreichen können | Am günstigsten. Bitten Sie um eine reservierte Adresse für Ihren Computer und darum, dass der Verkehr von Gästen zum Server erlaubt wird. Siehe [venue-network-guide.md](venue-network-guide.md). |
| **C. Der Hotspot Ihres eigenen Laptops** | Nur Proben und winzige Zuschauerzahlen | Laptop-Hotspots schaffen meist nur eine Handvoll Geräte. Nichts für ein volles Haus. |

### Option A im Detail (eigener Zugangspunkt)

```
                 ┌───────────────────────┐
                 │  Access point/router  │  WiFi "Theater Captions"
                 │  (in the audience     │ ─ ─ ─ ─ ─ ─ ─ ─  phones / tablets
                 │   area, 5 GHz on)     │                  projector browser
                 └───────────┬───────────┘
                    Ethernet │
                 ┌───────────┴───────────┐
                 │ Captioning computer   │   ← operator uses this screen,
                 │  (open-caption)       │     or a second device on the same WiFi
                 └───────────────────────┘
```

- **Ausstattung:** ein professioneller oder semiprofessioneller Zugangspunkt oder Router, der für viele gleichzeitige Geräte ausgelegt ist (achten Sie in den Daten auf über 100 Clients und 5-GHz-Unterstützung), mit Ethernet-Anschluss. Billige Heimrouter haben oft ab 30–50 Geräten Mühe. Planen Sie für ein Haus mit etwa 150 Plätzen einen leistungsfähigen Zugangspunkt ein, und einen zweiten für einen tiefen oder breiten Saal.
- **Einstellungen:**
  - Legen Sie eine SSID mit kurzem Namen und einer einfachen, druckbaren WPA2-Passphrase an, oder ohne Passphrase, wenn das Haus ein offenes Netz erlaubt. Drucken Sie sie auf das QR-Schild.
  - Schalten Sie **5 GHz** ein und bevorzugen Sie es. Verwenden Sie in einem überfüllten Gebäude Kanalbreiten von 20 oder 40 MHz.
  - Lassen Sie die **Client-Isolation** an, wenn sie angeboten wird. Publikumsgeräte sprechen nur mit dem Server, nie untereinander.
  - Geben Sie dem Untertitel-Computer eine **DHCP-Reservierung** (oder eine feste Adresse), damit sich die Adresse nie ändert, und drucken Sie die Adresse auf das Schild.
  - Sie brauchen keine Internetverbindung, und das Programm verwendet nie eine.
- **Aufstellung:** Stellen Sie den Zugangspunkt im Zuschauerbereich auf, hoch und frei von Wänden, nicht in einem Hinterzimmer oder im Foyer. Stellen Sie ihn nicht hinter die Bühne oder das Lichtrack. Ein Laptop auf einem entfernten Tisch ist in Ordnung: Er hängt am Kabel.
- **Strom:** Verwenden Sie einen Überspannungsschutz. Eine kleine USV an Zugangspunkt und Server beseitigt das Risiko eines herausgezogenen Steckers.

### Option B im Detail (das Gast-WLAN des Hauses)

Dem Programm ist egal, in welchem Netzwerk es ist, solange Publikumsgeräte es erreichen können. Bevor Sie sich darauf verlassen:

- Bitten Sie um **Gast-Zugriff auf den Server**: Ihr Computer muss von der Gast-SSID aus erreichbar sein, über TCP-Port 80 (und 443, wenn Sie HTTPS einschalten). „Client-Isolation“ zwischen Gästen ist in Ordnung; Gäste daran zu hindern, *Ihren Computer* zu erreichen, nicht.
- Bitten Sie um eine **reservierte Adresse** für den Netzwerkadapter Ihres Computers.
- Schließen Sie den Computer wenn irgend möglich per **Ethernet** ans Netzwerk an.
- Testen Sie rechtzeitig vor der Show mit einem Telefon im Gast-WLAN (der Schritt Teilen zeigt die Adresse).
- Haben Sie Option A als Reserve in einer Tasche, wenn die Show wichtig ist.

### Der Untertitel-Computer

- **Jeder gewöhnliche Laptop oder Mini-PC** genügt. Das Programm braucht sehr wenig Arbeitsspeicher (einige Dutzend Megabyte) und fast keinen Prozessor. Nehmen Sie einen Computer, den Sie angesteckt lassen können.
- **Betriebssystem:** macOS, Windows oder Linux. Verwenden Sie wenn möglich für alle Shows denselben Computer, damit Passwörter, Shows und Adresse gleich bleiben.
- **Machen Sie ihn langweilig:** Stellen Sie den Computer so ein, dass er im Netzbetrieb *nie in den Ruhezustand* geht, schalten Sie automatische Neustarts für Updates rund um die Vorstellungstermine ab, schließen Sie andere Programme und lassen Sie den Deckel offen oder so eingestellt, dass er nicht den Ruhezustand auslöst.
- **Firewall:** Erlauben Sie das Programm in **privaten Netzwerken** (Windows fragt beim ersten Mal; erlauben Sie unter macOS eingehende Verbindungen dafür).
- **Adresse:** per Kabel und reserviert (siehe oben). Ändert sich die Adresse, funktionieren gedruckte QR-Codes nicht mehr.
- **Uhrzeit:** Die Uhr spielt keine Rolle.

### Wer sitzt wo

- **Der Operator** braucht freie Sicht auf die Bühne und ein zuverlässiges Gerät. Am einfachsten ist der Untertitel-Computer selbst am Tontechnik-Tisch. Die Operator-Seite ist für einen Computer mit Tastatur gedacht (die Leertaste sendet die nächste Zeile). Nehmen Sie den Untertitel-Computer oder einen anderen Laptop im selben Netzwerk (`http://<Adresse des Computers>/operator`), damit der Operator dort sitzen kann, wo er die Bühne sieht. Ein Telefon oder Tablet zeigt die Seite an, wird aber **nicht empfohlen**: Auf einem Touchscreen tippt man während der Vorstellung leicht daneben. Ein Bluetooth-Klicker oder eine Tastatur ist eine gute Alternative.
- **Nur ein Operator steuert die Show zur selben Zeit.** Wer sonst die Operator-Seite öffnet, kann zusehen und **Übernehmen** verwenden, wenn der Operator übergibt.
- **Die Freiwilligen der Einrichtung** nutzen die Einrichtungsseite vor der Show von jedem Computer im Netzwerk oder vom Untertitel-Computer aus.
- **Ein gemeinsamer Bildschirm** für den ganzen Saal: Öffnen Sie die Publikumsadresse auf einem Computer, der an einen Projektor oder großen Monitor angeschlossen ist, hängen Sie `?display=projector` an die Adresse an für großen Text und keinen Bildschirmkram, und schalten Sie den Browser auf Vollbild (`F11`, oder starten Sie Chrome mit `--kiosk`).

### Den Server betreiben

Für die meisten Theater ist der einfachste Weg der beste: Starten Sie das Programm vor jeder Show von Hand, lassen Sie sein Fenster offen und drücken Sie danach `Strg+C` (oder schließen Sie es). Die Position der Show wird laufend gespeichert, ein Neustart mitten in einer Show ist daher sicher (siehe [Fehlerbehebung](#fehlerbehebung)).

Wenn der Computer den Server nach einem Neustart von selbst starten soll, enthält der Ordner [`deploy/`](../../../deploy/) fertige Dateien für Linux (systemd), macOS (launchd) und Windows (Aufgabenplanung). Jede beginnt mit ihren eigenen Installationsschritten.

### Die Leute zur Seite bringen

In der Reihenfolge der Zuverlässigkeit:

1. **Der QR-Code und das Schild** (Einrichtung → Teilen → *QR-Schild für das Foyer drucken*). Hängen Sie es an die Wand, legen Sie es ins Programmheft und auf die Karte an der Sitzlehne. Darauf sollten Sie sich verlassen.
2. **Eine kurze Adresse.** Auf Port 80 gibt es kein `:Port` zu tippen. Das Programm macht sich auch unter **`open-caption.local`** bekannt (ändern unter Einrichtung → Einstellungen → Netzwerk → *Einen Namen verwenden*): Wo das Gastnetz es erlaubt, funktioniert `http://open-caption.local/` auf den meisten Telefonen, ohne dass etwas eingerichtet wird. Manche Gastnetze blockieren diese Art von Bekanntmachung, drucken Sie also nicht nur den Namen; das QR-Schild verwendet die numerische Adresse. Ein Name im eigenen DNS des Routers (etwa `captions.theater`) funktioniert überall, muss aber vom Haus angelegt werden.
3. **Die Startseite des Zugangspunkts** (ein „Gästeportal“ oder „Captive Portal“, das nach dem Beitritt auf Ihre Adresse umleitet). Das wird am Router oder Zugangspunkt eingerichtet, nicht in diesem Programm. Siehe [venue-network-guide.md](venue-network-guide.md).
4. **DHCP-Option 114 und `/captive-api` (fortgeschritten, experimentell).** Option 114 ist eine DHCP-Einstellung, die Geräte auf eine „Captive-Portal-API“ verweist. Das Programm beantwortet diese API unter `/captive-api` (zum Beispiel `https://<Adresse>/captive-api`) mit „dieses Netzwerk ist nicht gefangen“ und der Adresse der Seite. Zwei Vorbehalte: Es **lässt Telefone die Untertitelseite nicht von selbst öffnen**, und der Standard (RFC 8910/8908) verlangt **HTTPS mit einem Zertifikat, dem die Geräte vertrauen**, sodass eine einfache `http://`-Einstellung oder das selbstsignierte Zertifikat des Programms ignoriert wird. Richten Sie sie nur ein, wenn das Haus Option 114 bereits verwendet und ein vertrauenswürdiges Zertifikat bereitstellen kann. Sie wurde auf echten Geräten nicht getestet.

### Ports und Adressen

- Standardmäßig lauscht der Server auf **Port 80**, sodass die Publikumsadresse keine Portnummer braucht, zum Beispiel `http://192.168.1.10/`. Benutzt etwas anderes Port 80 oder erlaubt das Betriebssystem ihn nicht, sagt der Server es Ihnen und verwendet stattdessen **Port 8080**; die Publikumsadresse endet dann auf `:8080`. Unter Linux braucht Port 80 eine Berechtigung (`sudo setcap 'cap_net_bind_service=+ep' ./open-caption`), es sei denn, Sie verwenden die systemd-Datei, die sie gewährt.
- Sie können Ports selbst wählen, mit `--http-addr :8081` oder in `config.yaml`.
- **HTTPS ist optional.** Einfaches `http://` funktioniert auf jedem Gerät ohne Warnungen und ist die empfohlene Publikumsadresse. Eine selbstsignierte HTTPS-Adresse lässt jedes Gerät eine Sicherheitswarnung zeigen. Schalten Sie HTTPS nur ein, wenn Sie ein Zertifikat installieren können, dem Ihre Geräte vertrauen. Besitzt das Haus einen eigenen Domainnamen, kann das Programm ein vertrauenswürdiges Zertifikat selbst holen und erneuern; siehe die [Anleitung für ein vertrauenswürdiges HTTPS-Zertifikat](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

### Sicherheit

- Die Publikumsseite kann nur lesen. Sie kann die Show nicht ändern und die Einrichtungs- und Operator-Seiten nicht erreichen.
- Einrichtung und Operator verlangen je ein eigenes Passwort (in `config.yaml` nur als Hash gespeichert, eine Einwegprüfsumme, die sich nicht zurücklesen lässt; die Datei wird außerdem so angelegt, dass nur Sie sie lesen können). Ein optionales **Zuschauer**-Passwort (unter **Einstellungen**) lässt ein zweites Gerät, etwa das Tablet der Inspizienz, die Mitlese-Seite (`/follow`; das Operator- und das Administrator-Passwort öffnen sie ebenfalls) öffnen und die Show ansehen, ohne etwas senden zu können. Geben Sie dem Operator das Operator-Passwort, nicht das Administrator-Passwort. Passwörter lassen sich unter **Einstellungen** in der Einrichtung ändern.
- Auf dem Computer, auf dem das Programm läuft, öffnen sich Einrichtung und Bedienerseite **ohne Passwort**, wenn die Adresse mit `localhost`, `127.0.0.1` oder `[::1]` beginnt (der Person an diesem Computer wird als Admin vertraut). Andere Computer und jede über einen Proxy weitergeleitete Anfrage melden sich weiterhin mit Passwort an. Setzen Sie `access.local_auto_login: false` in `config.yaml`, wenn sich andere Personen an diesem Computer anmelden können.
- **Jeder, der den Server erreichen kann, kann die Anmeldeseiten von Einrichtung und Operator öffnen, und bei einfachem `http://` laufen die Passwörter unverschlüsselt über das WLAN.** In einem gemeinsamen Gast-WLAN heißt das, dass ein neugieriger Zuschauer ein Passwort mitschneiden und die Untertitel übernehmen könnte. Verringern Sie das Risiko in dieser Reihenfolge:
  1. **Begrenzen Sie, welche Computer sie öffnen dürfen.** In `config.yaml`:
     ```yaml
     access:
       admin_allow: [localhost]            # Setup only on the captioning computer itself
       operator_allow: [192.168.1.50]      # the operator's computer (or a network such as 192.168.1.0/24)
     ```
     Einträge können `localhost`, `private` (Adressbereiche für Zuhause/Büro), eine einzelne Adresse oder ein Netzwerk wie `192.168.1.0/24` sein. `admin_allow` ist standardmäßig `[localhost]` (die Einrichtung öffnet sich nur auf dem Untertitel-Computer); eine leere Liste bedeutet jeder Computer, mit einer Warnung. `operator_allow` ist standardmäßig leer. Eine Adresse in `admin_allow` darf auch die Operator-Seite verwenden. Statt die Datei zu bearbeiten, können Sie all das in der Einrichtung unter **Einstellungen → Zugriff → Wer die Seiten öffnen darf** ändern; die Änderung gilt sofort, und die Einrichtung lehnt eine Liste ab, die den Computer ausschließen würde, an dem Sie arbeiten. Wenn Sie die Datei von Hand bearbeiten, starten Sie das Programm neu; ein fehlerhafter Eintrag hält das Programm mit einer Meldung an, statt allen den Zugang zu erlauben.
  2. **Setzen Sie den Operator in ein kabelgebundenes oder getrenntes Netzwerk**, dem Publikumsgeräte nicht beitreten können, oder geben Sie dem Untertitel-Netz ein WPA2-Passwort.
  3. **Verwenden Sie HTTPS** (`https.enabled` in `config.yaml`), wenn Sie den Geräten ein Zertifikat geben können, dem sie vertrauen; ein selbstsigniertes lässt jedes Gerät eine Warnung zeigen, und eine unverschlüsselte Publikumsadresse ist einfacher.
- Eine Website oder ein anderer Dienst auf demselben Computer kann einen Browser nicht dazu bringen, die Show hinter Ihrem Rücken zu ändern: Anfragen, die nicht von den eigenen Seiten des Programms stammen, werden abgelehnt.
- Das Programm ist für ein privates Netzwerk gedacht. **Setzen Sie es nicht dem Internet aus.**

## Installation und Start

1. **Besorgen Sie das Programm.** Laden Sie die Fassung für Ihren Computer von der Seite *Releases* des Repositorys herunter und entpacken Sie sie (oder bauen Sie sie selbst: siehe die Bauanleitungen im Ordner `docs` des Quell-Repositorys). Legen Sie es in einen Ordner, den Sie behalten, zum Beispiel `Dokumente/open-caption`.
2. **Erster Start.** Starten Sie es: unter Windows per Doppelklick, oder in einem Terminal `./open-caption` (macOS/Linux). Das Programm ist nicht signiert, daher kann beim ersten Start ein zusätzlicher Klick nötig sein: unter macOS *Rechtsklick → Öffnen* (oder **Trotzdem öffnen** in Systemeinstellungen → Datenschutz & Sicherheit), unter Windows *Weitere Informationen → Trotzdem ausführen* im SmartScreen-Fenster.
3. **Notieren Sie die Passwörter.** Beim ersten Mal gibt das Programm die **Administrator-** und **Operator-**Passwörter aus und sagt, wo `config.yaml` liegt. Sie bleiben in dieser Datei.
4. Die Einrichtungsseite öffnet sich in Ihrem Browser. Falls nicht, öffnen Sie `http://localhost/admin` (oder `http://localhost:8080/admin`, wenn die Konsole sagt, dass sie Port 8080 verwendet hat).
5. **Wählen Sie, wo die Daten liegen (optional).** Standardmäßig liegen Shows und Einstellungen im Konfigurationsordner Ihres Benutzers (macOS `~/Library/Application Support/open-caption`, Linux `~/.config/open-caption`, Windows `%AppData%\open-caption`). Um alles in einem Ordner Ihrer Wahl zu halten, starten Sie mit `--data <Ordner>`. Verwenden Sie immer denselben Ordner.

Nützliche Optionen: `--data <Ordner>`, `--http-addr :8080`, `--https-addr :443`, `--no-browser`, `--log-file <Datei>`, `--log-level debug|info|warn|error`, `--version`. Alle Optionen: `open-caption --help`, die Handbuchseite `docs/open-caption.1` (nur auf Englisch) oder `docs/configuration-reference.md`.

Hat `config.yaml` einen Fehler (eine falsch geschriebene Einstellung, eine Zahl außerhalb des Bereichs, eine fehlerhafte Adresse), listet das Programm jedes Problem mit Zeile oder Einstellungsnamen auf und hält an; beheben Sie die Datei und starten Sie neu. (Eine Datei, die mit einer neueren `config_version` gekennzeichnet ist, als das Programm versteht, ist die Ausnahme: Unbekannte Einstellungen werden mit einer Warnung ignoriert, sodass Sie zu einem älteren Programm zurückgehen können.)

## Eine Show einrichten

Öffnen Sie die Einrichtungsseite und folgen Sie den sechs Schritten oben (ein siebter, **Bericht**, ist für nach der Show). Die **Hilfe**-Schaltfläche erklärt jeden.

1. **Shows.** Legen Sie eine Show an (oder die Beispiel-Show zum Ausprobieren).
2. **Show-Datei.** Fügen Sie die Show-Datei ein oder ziehen Sie eine `.txt`- oder `.md`-Datei hinein. Jede Zeile wird zu einem Untertitel. Beheben Sie, was unter *Probleme und Hinweise* aufgeführt ist; klicken Sie eines an, um zu seiner Zeile zu springen. Speichern.
3. **Bilder.** Laden Sie die Bilder hoch, die die Show-Datei verwendet (PNG, JPG, GIF oder WebP, idealerweise unter 1 MB je Bild).
4. **Erscheinungsbild.** Wählen Sie Farben, Schrift und Textgröße. Die Voreinstellungen (weiß auf schwarz, groß) sind für die meisten am besten. Zwei Lesbarkeitsschriften sind eingebaut: eine legasthenie-freundliche (OpenDyslexic) und eine mit hoher Lesbarkeit (Atkinson Hyperlegible, für Sehbehinderung entworfen). Einzelne Zuschauer können über das kleine Steuerelement **Aa** auch für ihr eigenes Gerät eine der beiden wählen.
5. **Test.** Machen Sie eine private Probe: Sie drücken die Leertaste wie der Operator und sehen daneben den Bildschirm des Publikums. Nichts erreicht das Publikum.
6. **Teilen.** Prüfen Sie die Bereitschaftsliste, drucken Sie das QR-Schild und drücken Sie **Die Seite „Fahren“ öffnen**.

**An einer Show zu arbeiten heißt nicht, sie zur aktiven Show zu machen.** *An dieser Show arbeiten* wählt nur, welche Show die Einrichtung bearbeitet (ihr Name steht in der Kopfzeile). Die Show, die das Publikum bekommt, die **aktive** (im Schritt Shows mit AKTIV markiert und in der Kopfzeile durch die Markierung Aktiv gezeigt, bernsteinfarben im Standby und rot auf Sendung), wechselt nur, wenn Sie im Schritt Teilen dieser Show auf **Die Seite „Fahren“ öffnen** drücken. Eine Show auf Sendung wird erst nach einer Rückfrage ersetzt.

Das Öffnen der Bedienerseite lädt die Show im **Standby**: Das Publikum sieht den Standby-Bildschirm des Hauses (ein Logo und eine Meldung, einmal unter **Einstellungen → Publikumsbildschirm** festgelegt) und nichts von der Show, sodass Sie in Ruhe vorbereiten, proben und die Bedienerseite prüfen können. Wenn Sie bereit sind, drücken Sie auf der Bedienerseite die **Leertaste** (die Schaltfläche heißt **START**), um die Show auf Sendung zu bringen. **In den Standby …** nimmt sie wieder heraus, ohne ihren Platz zu verlieren.

### Die Show-Datei schreiben

Wer die Show-Datei vorbereitet, muss den Rest dieser Anleitung nicht lesen: [Eine Show-Datei vorbereiten](../../../web/help/admin-preparing-a-show-file.md) behandelt die Arbeit vom Stück bis zur fertigen Datei, auch das Bearbeiten im eigenen Programm, und [Erste Schritte](../../../web/help/admin-getting-started.md) führt durch die sechs Schritte der Einrichtung.

Eine Show-Datei ist einfacher Text. Die Kurzfassung:

```text
<!-- Comments like this are never shown. -->
[loop: 10s]
[image: welcome.png | 15s]
Welcome! Please silence your phones.
[/loop]

[section: Act 1, Scene 1]
[note: lights come up slowly]
[clear]
(A garden gate creaks open.)
MARGARET: You're late, Edwin.
EDWIN: The *train* was late.

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[/song]
~ Tea for two, and two for tea ~

[fullscreen: **INTERMISSION**]
```

- **Eine Zeile = ein Untertitel** = ein Druck auf die Leertaste. Sie entscheiden, wo lange Reden umbrochen werden.
- **Sprechernamen** in Großbuchstaben mit Doppelpunkt werden über der Zeile gezeigt.
- **Bühnenanweisungen** in `(Klammern)` werden kursiv gezeigt.
- **Liedtexte** zwischen `♪`-Noten, kursiv: Beginnen Sie eine Zeile mit `~` oder fassen Sie das Lied in `[song]`…`[/song]`. Für zwei Singende gleichzeitig (ein Duett oder Kontrapunkt) setzen Sie ihre Zeilen zwischen `[together]` und `[/together]` oder verbinden Sie sie mit ` || ` in einer Zeile; beide erscheinen mit einem Druck auf die Leertaste zugleich. Die Schaltflächen **♪**, **♫** und **🎵 Lied** des Editors fügen diese ein. Siehe das Hilfethema *Musicals*.
- **Schleifen** (vor der Show und Pause) wiederholen Bilder und Text, bis der Bediener weitermacht. Eine Schleife ganz am Anfang der Show-Datei startet, wenn der Bediener die Show auf Sendung bringt (START).
- **Notizen** `[note: …]` und Lesezeichen `[section: …]` sind nur für den Operator. `[pace: 8s]` sagt dem Operator, wie lange die nächste Zeile üblicherweise stehen bleibt, und eine Zeile, die mit `# ` beginnt, ist eine Überschrift, die wie ein Lesezeichen funktioniert.
- Das Hilfethema **Eine Show-Datei schreiben** listet alles mit Beispielen auf, und der *Spickzettel zur Show-Datei* des Editors fasst es zusammen.

### Show-Einstellungen (Titel, Erscheinungsbild und mehr)

Alles an einer Show außer ihren Bildern wird in ihrer Show-Datei gespeichert, auch ihr Titel und das Erscheinungsbild der Untertitel, in einem `[settings]`-Block nahe am Anfang:

```text
[settings]
title: The Garden Party
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
[/settings]
```

Der Schritt **Erscheinungsbild** der Einrichtung schreibt diesen Block für Sie, und Sie können ihn von Hand bearbeiten. Alles, was eine Show-Datei auslässt, verwendet einen Standardwert: Das Programm hat eingebaute Werte, `config.yaml` (ihr Abschnitt `defaults`) kann sie für jede Show auf diesem Computer ändern, und jede Show-Datei überschreibt beides. Ports, Passwörter, HTTPS und die Zugriffslisten sind keine Show-Einstellungen und bleiben in `config.yaml`. Weil die Show-Datei ihr Erscheinungsbild trägt, nimmt Kopieren oder Exportieren der Show das Erscheinungsbild mit. Alle Einstellungen finden Sie im Hilfethema *Eine Show-Datei schreiben*.

### Die Namen der Figuren prüfen

Klicken Sie über dem Editor auf **Figuren…**, um alle als Sprecher verwendeten Namen zu sehen, die am häufigsten verwendeten zuerst, mit der Zahl ihrer Zeilen. So finden Sie schnell einen falsch geschriebenen Namen: **ALCE** mit einer Zeile neben **ALICE** mit vierzig wird mit *Meinten Sie ALICE?* markiert. Klicken Sie auf einen Namen, um zu seiner ersten Zeile zu gehen. Eine Gruppe wie `ALICE & BOB:` zählt für jeden ihrer Namen.

### Zurückgehen

Jedes Speichern bewahrt die ersetzte Version auf (die letzten 50). **Frühere Versionen …** über dem Editor listet sie auf; **Laden** legt eine in den Editor, und **Speichern** macht sie aktuell. **Vergleichen** zeigt Zeile für Zeile, was das Laden einer Version ändern würde, und eine Version, der Sie einen Namen geben (oder **Behalten** für das, was jetzt gespeichert ist), wird nie vergessen. Exportieren Sie vor der Show trotzdem: Die früheren Versionen liegen im Datenordner, nicht im Export.

### Zwei Personen bearbeiten gleichzeitig

Speichert jemand anderes eine Show-Datei, während Sie sie geöffnet haben, warnt Sie das Programm beim Speichern, statt deren Arbeit zu überschreiben, und lässt Sie deren Version nehmen oder Ihre behalten.

### Mehr als eine Sprache

Eine Show kann Untertitel in mehreren Sprachen anbieten. Klicken Sie in Einrichtung → Show-Datei auf **＋ Sprache hinzufügen**, lassen Sie „Mit einer Kopie der Haupt-Show-Datei beginnen“ angekreuzt und übersetzen Sie die Wörter. Jede Sprache ist eine eigene Show-Datei mit denselben Zeilen, Befehlen und leeren Bildschirmen in derselben Reihenfolge, sodass ein Druck auf die Leertaste in jeder Sprache denselben Moment zeigt. Der Operator folgt nur der Hauptsprache. Publikumsgeräte erhalten die Sprache, auf die ihr Telefon eingestellt ist (oder die Hauptsprache) und können über das kleine 🌐-Menü wechseln oder einen Link erhalten, der auf `?lang=es` endet. Eine Übersetzung, die nicht mehr zur Haupt-Show-Datei passt, wird gemeldet und zurückgehalten, bis sie korrigiert ist. Die Beispiel-Show enthält eine spanische Übersetzung zum Ansehen. Siehe das Hilfethema *Sprachen*.

### Dinge während Proben und Shows ändern

Sie können eine Show-Datei bearbeiten und erneut speichern, während eine Show live läuft (zum Beispiel um einen Tippfehler zu beheben). Das Programm bittet Sie zur Sicherheit, den Namen der Show einzugeben, behält die Position des Operators über die Zeilennummer bei und warnt, wenn sich die Zeilenzahl geändert hat.

## Die Show fahren

### Bevor das Publikum kommt

1. Starten Sie den Server und öffnen Sie die **Operator-Seite**; melden Sie sich an. Dort sollte **Sie haben die Steuerung** stehen.
2. Drücken Sie in Einrichtung → Teilen auf **Die Seite „Fahren“ öffnen**. Die Show wird im Standby geladen: Das Publikum sieht den Standby-Bildschirm des Hauses. Wenn Sie bereit sind, drücken Sie auf der Bedienerseite die **Leertaste** (START), um sie auf Sendung zu bringen. Eine Schleife vor der Show, falls die Show-Datei eine hat, beginnt dann auf jedem Bildschirm.
3. Öffnen Sie die Publikumsadresse auf Ihrem eigenen Telefon und, falls vorhanden, am Projektor. Prüfen Sie, dass die Zuschauerzahl auf der Operator-Seite steigt.
4. Stellen Sie das QR-Schild dort auf, wo die Leute es sehen.

### Während der Show

- Die **Leertaste** sendet die nächste Zeile (auch Pfeil rechts, Bild ab, Eingabe und die meisten Klicker).
- Die **Rücktaste** macht das Letzte rückgängig, was Sie gesendet haben.
- **B** blendet die Publikumsbildschirme aus (verbirgt alles; dieselben Untertitel kommen zurück, wenn Sie noch einmal **B** drücken). **C** leert sie (löscht die Untertitel, damit die nächste Zeile auf einem leeren Bildschirm beginnt). Die **Rücktaste** macht beides rückgängig.
- **Umschalt+Q** stoppt eine Schleife in der Pause oder vor der Show (leerer Bildschirm). Die **Leertaste** macht weiter.
- **Durchsage …** bringt eine einmalige Mitteilung auf jeden Bildschirm („Die Show geht gleich weiter“). **Zurück** entfernt sie.
- Klicken Sie eine Zeile an und bestätigen Sie, um irgendwohin zu springen. **/** durchsucht die Show-Datei.
- Die Zeile mit dem gelben Pfeil ist die *nächste*. Der grüne Punkt markiert, was jetzt auf dem Bildschirm ist.

### Wenn die Show endet

Drücken Sie `Strg+C` im Fenster des Programms (oder beenden Sie den Dienst). Die Position wird gespeichert. Um die Show wieder von vorn laufen zu lassen, verwenden Sie **Show neu starten …** auf der Bedienerseite oder **Show neu starten …** in Einrichtung → Teilen. Beides bringt die Show zurück in den Standby. **Wie viele Leute haben es genutzt?** Erreicht die Show ihr Ende, steht im Banner der Bedienerseite *Ende der Show.* mit einem Link **Publikumsbericht** (der Schritt **Bericht** der Einrichtung zeigt den letzten). Er zählt die verschiedenen Geräte, die die Publikumsseite genutzt haben, während die Show auf Sendung war, egal wie lange jedes blieb (ein Telefon, das aus dem WLAN fiel und zurückkam, zählt einmal), mit einem Diagramm der Geräte im Zeitverlauf (das zeigt, wann die Schleifen vor der Show und in der Pause liefen), dem Moment des größten Andrangs, den Sprachen und einer Zeile je Gerät; er lässt sich drucken oder als CSV-Datei herunterladen. Geräte werden an Adresse und Browsertyp unterschieden, daher ist es eine Schätzung. Die Berichte liegen im Ordner `reports` des Datenordners.

## Checkliste für den Vorstellungstag

**Ein paar Tage vorher**
- [ ] Proben Sie mit der echten Show-Datei und ein paar echten Telefonen im echten Netzwerk im echten Saal.
- [ ] Exportieren Sie die Show (Einrichtung → Shows → **Mehr ▾** → **Exportieren**) und bewahren Sie das `.zip` an einem sicheren Ort auf, etwa auf einem USB-Stick.
- [ ] Bestätigen Sie, dass die Adresse des Untertitel-Computers reserviert ist und das QR-Schild sie zeigt.

**Im Haus, eine Stunde vorher**
- [ ] Strom: Computer und Zugangspunkt angesteckt; Computer so eingestellt, dass er nie in den Ruhezustand geht.
- [ ] Zugangspunkt an, an Ort und Stelle, mit dem richtigen Namen; der Computer per Kabel daran.
- [ ] Starten Sie das Programm; notieren Sie die ausgegebene Adresse (und die in Teilen).
- [ ] Ein Telefon im Publikums-WLAN kann die Adresse öffnen und zeigt „Die Untertitel beginnen in Kürze“.
- [ ] Einrichtung → Test: Machen Sie eine kurze Probe; die Tasten des Operators funktionieren.
- [ ] Projektor-Browser (falls verwendet) auf `…/?display=projector` geöffnet, im Vollbild.
- [ ] Drücken Sie auf der Bedienerseite **START**. Die Schleife vor der Show (falls vorhanden) läuft auf jedem Bildschirm.
- [ ] Der Operator sitzt dort, wo er die Bühne sieht; Klicker oder Tastatur getestet.
- [ ] QR-Schild aufgestellt; eine freiwillige Person kennt WLAN-Namen und Adresse für Leute, die fragen.

**Nach der Show**
- [ ] `Strg+C` im Programmfenster.
- [ ] Exportieren Sie die Show erneut, falls Sie etwas geändert haben.

## Fehlerbehebung

| Problem | Was zu versuchen ist |
|---------|----------------------|
| Niemand kann die Adresse öffnen | Ist das Telefon im richtigen WLAN (nicht im Mobilfunknetz)? Enthält die Adresse `:8080`, falls das Programm diesen Port verwendet hat? Ist der Verkehr von Gästen zum Server erlaubt? Erlaubt die Firewall des Computers das Programm in privaten Netzwerken? Siehe [venue-network-guide.md](venue-network-guide.md). |
| Manche Geräte verbinden sich, andere nicht | Der Zugangspunkt hat vielleicht sein Gerätelimit erreicht. Fügen Sie einen zweiten Zugangspunkt hinzu oder rücken Sie den ersten in den Saal. |
| Untertitel kommen spät oder stocken auf einigen Telefonen | Schwaches Signal oder ein ausgelasteter Kanal. Diese Telefone verbinden sich von selbst neu. Verbessern Sie die Aufstellung des Zugangspunkts; verwenden Sie 5 GHz. |
| Jemand hat eine Zeile verpasst | Er kann die Untertitel nach oben scrollen, um sie noch einmal zu lesen; neue Zeilen warten hinter einer Schaltfläche **Zurück zu Live**, bis er sie antippt. |
| Eine Seite sagt „Dieser Browser ist zu alt“ | Der Browser des Geräts liegt unter dem, was die Seite braucht: etwa iOS 12 für die Publikumsseite, iOS 14.1 / Chrome 87 für die Einrichtung und die Operator-Seite. Aktualisieren Sie das Gerät oder verwenden Sie ein anderes. |
| Bildschirme von Telefonen werden dunkel | Telefone sperren ihren Bildschirm. Die Seite hält den Bildschirm mit der Wake-Lock-Funktion des Browsers an, wo sie kann, und, wo nicht (einfaches HTTP), mit einem winzigen unsichtbaren Endlosvideo. Wer sein Telefon lieber schlafen lassen möchte (etwa bei schwachem Akku), kann **Bildschirm bleibt an** im Bereich „Aa“ ausschalten. Manche Telefone im Energiesparmodus lehnen beides ab: Die Seite zeigt dann einen Hinweis, die automatische Sperre für den Abend auszuschalten. |
| Es wurde die falsche Zeile gesendet | Die **Rücktaste** macht es rückgängig. |
| Die Operator-Seite sagt, ein anderer Operator hat die Steuerung | Verwenden Sie **Übernehmen**, wenn Sie der Operator sind. |
| Der Computer hängte sich auf oder wurde mitten in der Show neu gestartet | Starten Sie das Programm erneut. Die Operator-Seite verbindet sich neu, bleibt angemeldet und übernimmt die Steuerung von selbst, die Show steht, wo Sie sie verlassen haben, und die Publikumsgeräte verbinden sich von selbst neu. Eine Schleife beginnt wieder beim ersten Element. Die **Rücktaste** kann die letzten Dinge, die vor dem Neustart gesendet wurden, weiterhin rückgängig machen. |
| Ich habe ein Passwort vergessen | Melden Sie sich mit dem anderen an und verwenden Sie **Einstellungen** in der Einrichtung, oder öffnen Sie `config.yaml` im Datenordner mit einem beliebigen Texteditor. |
| Etwas anderes | Einrichtung → Fehlersuche → **Supportpaket herunterladen**; es enthält die Einstellungen (ohne Passwörter) und aktuelle Protokollzeilen. |

## Wartung

- **Sicherungen:** *Exportieren* (Menü **Mehr ▾**) unter Einrichtung → Shows speichert eine Show (Show-Datei, Bilder, Erscheinungsbild) als `.zip`. *Importieren* stellt sie wieder her, auf diesem oder einem anderen Computer. Sie können auch den ganzen Datenordner kopieren.
- **Auf einen anderen Computer umziehen:** Installieren Sie das Programm, importieren Sie die exportierten Shows, stellen Sie die Adresse des neuen Computers ein und kopieren Sie, wenn Sie die alten Passwörter behalten wollen, `config.yaml`.
- **Das Programm aktualisieren:** Beenden Sie es (`Strg+C`), ersetzen Sie die Programmdatei durch die neuere und starten Sie es mit demselben `--data`-Ordner erneut. Aktualisieren Sie nicht kurz vor einer Show.
- **Eine Show wiederverwenden:** Einrichtung → Shows hat **Umbenennen …** und **Duplizieren**. Duplizieren erstellt eine Kopie der Show-Datei, der Bilder und des Erscheinungsbilds zum Anpassen für die nächste Serie und lässt das Original unberührt. Umbenennen, Duplizieren, Exportieren, Archivieren und Löschen finden Sie im Menü **Mehr ▾** jeder Show.
- **Aufräumen:** Löschen Sie nicht mehr benötigte Shows unter Einrichtung → Shows. Exportierte `.zip`-Dateien sind Ihr Archiv.
- **Publikumsberichte:** Jede Durchführung einer Show hinterlässt ihren Publikumsbericht als Datei im Ordner `reports` im Datenordner (Einrichtung → Einstellungen → Info zeigt, wo der Datenordner ist). Die neuesten 30 bleiben erhalten, ältere werden automatisch entfernt. Ein Bericht listet die Netzwerkadressen der Geräte auf; löschen Sie daher die Berichte, die Sie nicht behalten möchten (**Löschen …** im Schritt **Bericht** der Einrichtung oder die Dateien selbst), einzeln oder den ganzen Ordner `reports` (er wird neu angelegt, wenn die nächste Show auf Sendung geht); nichts anderes hängt davon ab, und das Löschen berührt Ihre Shows nicht.
- **Protokolle:** Einrichtung → Fehlersuche → **Das ganze Protokoll öffnen…** zeigt, was das Programm zuletzt protokolliert hat, nach Stufe gefiltert (wählen Sie *Nur Probleme*, um Warnungen und Fehler zu sehen); aktuelle Protokollzeilen stehen auch im Supportpaket; fügen Sie `--log-file` hinzu, um eine Datei zu führen. Jede Zeile beginnt mit ihrer Stufe (`INFO`, `WARN`, `ERROR` oder `DEBUG`). Das voreingestellte `info` zeichnet Anmeldungen auf, wer die Steuerung der Operator-Seite übernommen hat, aktivierte Shows und ähnliche Ereignisse; `warn` behält nur Probleme; `debug` zeichnet zusätzlich jede Anfrage auf, was hilft, wenn etwas Merkwürdiges passiert. Legen Sie es mit `log: {level: debug}` in `config.yaml`, mit `--log-level debug` oder bei laufendem Programm unter Einrichtung → Einstellungen → *Protokoll-Detailgrad* fest. Wurde das Programm mit `--log-level` gestartet, gilt bei jedem Start diese Option: Die Einrichtung sagt das, eine dort vorgenommene Änderung gilt nur bis zum nächsten Neustart, und das Startprotokoll warnt, wenn Option und Datei nicht übereinstimmen. Eine mit `--log-file` geführte Protokolldatei beginnt bei 5 MB eine neue und behält drei ältere (`open-caption.log.1` bis `.3`), sodass sie nie den Datenträger füllt.
- **Archivieren:** Unter Einrichtung → Shows legt **Archivieren** eine abgeschlossene Show beiseite (sie bleibt erhalten, wird unter „Archivierte Shows“ aufgeführt und kann nicht auf Sendung gehen), und **Wiederherstellen** holt sie zurück. **Löschen** entfernt eine Show endgültig.

## Kurzkarte für den Operator

Drucken Sie sie aus und halten Sie sie beim Operator bereit.

| Taste | Bewirkt |
|-------|---------|
| **Leertaste** | Die nächste Zeile senden |
| **Rücktaste** | Das zuletzt Gesendete rückgängig machen |
| **S** | Die nächste Zeile überspringen, ohne sie zu senden (die Rücktaste nimmt es zurück) |
| **Umschalt+Q** | Eine Schleife stoppen (leerer Bildschirm) |
| **B** | Ausblenden: verbirgt alles, die Untertitel kommen beim Einblenden zurück |
| **C** | Leeren: löscht die Untertitel für einen frischen Start |
| **/** | In der Show-Datei suchen |
| **?** | Hilfe (die Tasten pausieren, solange sie geöffnet ist) |

- Gelber Pfeil = nächste Zeile. Grüner Punkt = jetzt auf dem Bildschirm.
- Falsche Zeile? **Rücktaste.** Stelle verloren? Klicken Sie die gewünschte Zeile an und bestätigen Sie.
- Müssen Sie dem Publikum etwas sagen? **Durchsage …**
- Show unterbrochen? **B**, um die Bildschirme leer zu machen, **Durchsage …**, um es zu erklären.
