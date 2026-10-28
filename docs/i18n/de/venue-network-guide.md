# Netzwerkleitfaden für das Haus

Für die Person, die das WLAN des Theaters betreut. Der Untertitel-Computer braucht nichts aus dem Internet. Er muss nur **von den Geräten des Publikums erreichbar** sein.

## Was das Publikum tut

1. Es verbindet sich mit dem Gast-WLAN.
2. Es öffnet eine Adresse, zum Beispiel `http://192.168.1.10/` (oder scannt den QR-Code).

## Anforderungen

| Anforderung | Warum |
|-------------|-------|
| Das Gastnetz kann den Untertitel-Computer erreichen | Die Geräte holen die Seite ab und halten eine Verbindung zu ihm offen |
| Der Computer hat eine feste Adresse (DHCP-Reservierung oder statisch) | Damit gedruckte QR-Codes und Schilder zwischen Vorstellungen gültig bleiben |
| Der Zugangspunkt verkraftet über 100 gleichzeitige Geräte (5 GHz aktiviert) | Ein volles Haus sind etwa 150 Geräte |
| Es wird keine Internetanbindung gebraucht | Alles wird vom Untertitel-Computer ausgeliefert |

**Client-Isolation zwischen Gästen ist in Ordnung und empfehlenswert.** Publikumsgeräte sprechen nur mit dem Untertitel-Computer, nie untereinander. Was *nicht* blockiert sein darf, ist der Verkehr von Gästen zum Server. Ist das Gastnetz vom Haupt-LAN isoliert, stecken Sie den Untertitel-Computer entweder ins Gastnetz (ein freier Port des Gast-VLANs oder die Gast-SSID) oder fügen eine Firewall-Regel hinzu, die Gästen den Zugriff auf die Adresse des Computers über TCP-Port 80 erlaubt (und 443, wenn HTTPS aktiviert ist).

## Vor der Vorstellung prüfen

Öffnen Sie von einem Telefon im Gast-WLAN die in der Einrichtung unter **Teilen** angezeigte Adresse. Wenn die Seite lädt und „Die Untertitel beginnen in Kürze“ sagt (oder die Anzeige vor der Show zeigt), sind Sie fertig. Falls nicht, siehe „Nichts lädt“ unten.

## Die Adresse leichter machen

In der Reihenfolge des Aufwands:

1. **QR-Code** — Einrichtung → Teilen → „QR-Schild für das Foyer drucken“. Hängen Sie es im Foyer auf und legen Sie es ins Programmheft. Es sind keine Netzwerkänderungen nötig.
2. **Port 80** — läuft das Programm auf Port 80, hat die Adresse kein `:Port`. Unter Linux erlauben Sie das mit `sudo setcap 'cap_net_bind_service=+ep' ./open-caption`.
3. **Ein lokaler Name** — legen Sie auf dem Router einen DNS-Eintrag für den Untertitel-Computer an, zum Beispiel `captions.theater` → `192.168.1.10`. Die meisten Router (OpenWrt, Pi-hole, UniFi, Omada, auf dnsmasq basierende) erlauben einen „lokalen DNS-Eintrag“ oder eine „Host-Überschreibung“.
4. **Gäste automatisch dorthin schicken.** Möglichkeiten:
   - **Startseite / Captive Portal** auf dem Zugangspunkt, das auf die Publikumsadresse zeigt. Die eingebaute „Gästeportal“-Funktion des Zugangspunkts arbeitet mit einer externen URL (UniFi: Hotspot → externes Portal; OpenWrt: nodogsplash `RedirectURL`). Achten Sie darauf, dass die Seite ohne Anmeldung laden darf.
   - **DHCP-Option 114 (RFC 8910) und `/captive-api`.** Das Programm beantwortet `GET /captive-api`, wie RFC 8908 es beschreibt: `{"captive": false, "user-portal-url": "http://<server-address>/"}`. Wenn Ihr Netzwerk bereits Option 114 („Captive-Portal API URI“) verteilt, können Sie sie auf diese Adresse zeigen lassen. Beachten Sie, was das bewirkt: Es teilt den Geräten mit, dass das Netzwerk *kein* Captive Portal ist, und veröffentlicht die Adresse der Seite. Es lässt Telefone die Untertitelseite **nicht** von selbst öffnen (dafür bräuchte es `"captive": true`, das das Programm nie sendet). Die RFC verlangt außerdem, dass die API-Adresse **https mit einem Zertifikat** ist, dem die Geräte vertrauen; eine schlichte `http://`-Adresse für Option 114 oder das selbstsignierte Zertifikat des Programms wird also ignoriert. Das ist experimentell und auf echten Geräten nicht getestet: Nutzen Sie das QR-Schild oder die Startseite Ihres Zugangspunkts, um die Leute zur Seite zu bringen.
5. **mDNS** (`open-caption.local`) ist vorhanden und standardmäßig aktiviert: Geräte, die es nutzen können, öffnen `http://open-caption.local/`, ohne dass am Router etwas eingerichtet wird. Dafür muss das Gastnetz **Multicast** (UDP 5353 an 224.0.0.251) zwischen dem Untertitel-Computer und den Geräten des Publikums zulassen. Viele Gastnetze blockieren das („Multicast-Filterung“ oder „mDNS/Bonjour-Blockierung“ am Zugangspunkt), und dann wird der Name schlicht nicht aufgelöst; die numerische Adresse und der QR-Code funktionieren weiterhin. Soll der Name funktionieren, aktivieren Sie die mDNS/Bonjour-Weiterleitung oder schalten Sie die Multicast-Filterung für das Gastnetz aus. Der Name lässt sich ändern oder abschalten in der Einrichtung unter Einstellungen → Netzwerk (*Name im Netzwerk*).

Das Programm betreibt absichtlich keinen eigenen DHCP- oder DNS-Server: Er würde mit dem Router des Hauses kollidieren.

## HTTP und HTTPS

Verwenden Sie für das Publikum einfaches `http://`. Ein selbstsigniertes HTTPS-Zertifikat lässt jedes Gerät eine Sicherheitswarnung zeigen. HTTPS ist verfügbar (`https.enabled: true` in `config.yaml`) für Standorte, die ein Zertifikat installieren können, dem ihre Geräte vertrauen. Besitzt das Haus einen eigenen Domainnamen, kann das Programm ein vertrauenswürdiges Zertifikat selbst holen und erneuern; siehe die [Anleitung für ein vertrauenswürdiges HTTPS-Zertifikat](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

Ein kleiner Nachteil von einfachem HTTP: Browser erlauben „Bildschirm anlassen“ nur auf sicheren Seiten. Bei HTTP hält die Seite den Telefonbildschirm stattdessen mit einem winzigen, stillen, unsichtbaren Endlosvideo an. Ein Telefon im Energiesparmodus lehnt es womöglich ab; die Seite zeigt dann einmalig einen Hinweis, die automatische Sperre auszuschalten.

## Sicherheit

- Die Einrichtung (`/admin`) und die Operator-Seite (`/operator`) brauchen jeweils ein Passwort aus `config.yaml`. Die Publikumsseite braucht keines und kann nichts ändern. Ein optionales Zuschauer-Passwort öffnet die Zuschauer-Seite (`/follow`), mit der man nur zusehen kann.
- Setzen Sie den Port nicht dem Internet aus.
- Die Operator- und Administrator-Passwörter stehen im Klartext in `config.yaml` (oder als Hash, siehe die Konfigurationsreferenz) im Datenordner des Programms. Halten Sie diesen Ordner für die Leute, die die Show fahren, privat.

## Firewalls auf dem Untertitel-Computer

- **Windows:** Erlauben Sie das Programm beim ersten Start in **privaten Netzwerken**. Wurde die Abfrage weggeklickt: Windows-Sicherheit → Firewall → App zulassen.
- **macOS:** Erlauben Sie eingehende Verbindungen für das Programm, wenn Sie gefragt werden, oder fügen Sie es unter Systemeinstellungen → Netzwerk → Firewall hinzu.
- **Linux:** Öffnen Sie den Port, zum Beispiel `sudo ufw allow 80/tcp`.

## Nichts lädt

1. Ist das Telefon im Gast-WLAN (nicht im Mobilfunknetz)?
2. Enthält die Adresse den Port, wenn das Programm nicht auf Port 80 läuft? Der Schritt Teilen zeigt die richtige Adresse.
3. Kann ein anderer Computer im Gast-WLAN ihn erreichen (`curl http://192.168.1.10/healthz` sollte `ok` ausgeben)?
4. Wird der Verkehr von Gästen zum Server vom Router oder Zugangspunkt blockiert (Client-Isolation auf „alles blockieren“ statt „zwischen Clients blockieren“)?
5. Blockiert die Firewall des Computers das Programm?
6. Ist der Computer noch wach und noch unter derselben Adresse? Achten Sie in Teilen auf eine geänderte Adresse.

Einrichtung → Fehlersuche → **Supportpaket herunterladen** sammelt die Einstellungen (ohne Passwörter), aktuelle Protokollzeilen und grundlegende Angaben zum Computer.

## Kapazität

Das Programm wurde unter Last mit 200 und 500 gleichzeitigen Publikumsverbindungen auf einem Computer getestet (Latenz unter 10 ms auf derselben Maschine). In einem echten Haus ist das WLAN die Grenze, nicht das Programm: Verwenden Sie einen für die Menge ausgelegten Zugangspunkt, aktivieren Sie 5 GHz und halten Sie Bilder klein (unter 1 MB), damit viele Geräte sie gleichzeitig abrufen können.
