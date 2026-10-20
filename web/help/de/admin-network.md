# Netzwerk und Verbindung

Zuschauer verbinden ihre eigenen Telefone und Tablets mit dem **Gast-WLAN** des Hauses und öffnen eine Adresse. Es muss nichts installiert werden.

## Was das Haus bereitstellen muss

- Dieser Computer und die Geräte des Publikums müssen im selben Netzwerk sein, oder das Netzwerk muss Gastgeräten erlauben, diesen Computer zu erreichen. Das ist eine Einstellung des WLAN-Systems des Hauses, nicht dieses Programms.
- Viele Gastnetzwerke isolieren Geräte voneinander. Das ist in Ordnung. Die Geräte sprechen nur mit diesem Computer, nie untereinander, solange dieser Computer erreichbar ist.
- Für diesen Computer ist eine Kabelverbindung am besten. Bitten Sie um eine reservierte Adresse (eine „DHCP-Reservierung“), damit sich die Adresse zwischen den Vorstellungen nicht ändert.
- Es ist keine Internetverbindung nötig.

Testen Sie es vor der ersten Show: Verbinden Sie ein Telefon mit dem Gast-WLAN und öffnen Sie die im Schritt **Fahren** angezeigte Adresse. Für einen Projektor oder einen großen Bildschirm führt derselbe Schritt einen Projektor-Link auf: Öffnen Sie ihn auf dem Computer, an dem der Projektor angeschlossen ist.

## Die Adresse leicht machen

1. **QR-Code** — drucken Sie das Schild aus dem Schritt Fahren aus und hängen Sie es im Foyer und im Programmheft auf. Das ist die einfachste Möglichkeit.
2. **Eine kurze Adresse** — wenn das Programm Port 80 verwenden kann, hat die Adresse keine Portnummer, etwa `http://192.168.1.10/`.
3. **Ein Name** — das Programm macht sich unter dem Namen **`open-caption.local`** bekannt, sodass `http://open-caption.local/` auf den meisten Telefonen und Computern im selben Netzwerk funktioniert, ohne dass etwas eingerichtet werden muss. Ändern Sie den Namen oder schalten Sie ihn ab unter **Einstellungen** (Schaltfläche oben), Reiter **Netzwerk**, bei **Name im Netzwerk** (das Kästchen **Einen Namen verwenden** und der Name). Manche Gast-WLANs blockieren diese Art von Bekanntmachung; verwenden Sie dann die numerische Adresse oder den QR-Code. Das Haus kann auch auf seinem Router einen eigenen Namen anlegen (etwa `captions.theater`), der überall funktioniert.
4. **Automatische Seite** — ein Haus kann Gäste beim Beitritt zum WLAN mit der Startseite des Zugangspunkts („Captive Portal“) auf die Seite schicken. Das Programm beantwortet auch die Standardadresse zur Netzwerkerkennung `/captive-api` für Häuser, die DHCP-Option 114 verwenden, aber das meldet nur, dass das Netzwerk nicht gefangen ist: Es öffnet die Seite nicht von selbst und braucht ein vertrauenswürdiges HTTPS-Zertifikat. Siehe den [Netzwerkleitfaden für das Haus](topic:admin-venue-network-guide).

## Wenn dieser Computer mehr als eine Adresse hat

Ein Computer mit Kabel und WLAN zugleich, oder mit virtuellen Maschinen oder einem VPN, hat mehrere Netzwerkschnittstellen. Öffnen Sie **Einstellungen** (Schaltfläche oben) und dort den Reiter **Netzwerk**: Er listet jede verbundene Schnittstelle mit einem Kästchen auf. Nur die angekreuzten werden verwendet, und das Publikum bekommt nur deren Adressen (in der Adressliste, im QR-Code und im `.local`-Namen). Kreuzen Sie nur die Netzwerke an, die das Publikum nutzt. Sind zwei angekreuzte Schnittstellen im selben Netzwerk (ein Kabel und WLAN, beide mit einer Adresse 192.168.1.x), wird nur eine Adresse angeboten, damit QR-Code und `.local`-Name übereinstimmen: die kabelgebundene, sonst die der Hauptverbindung. Um die andere zu verwenden, entfernen Sie das Häkchen der ersten. Das Häkchen wird beim Klicken gespeichert, und die Schnittstelle, über die Sie die Einrichtung benutzen, lässt sich nicht abwählen. Eine neue Schnittstelle (ein USB-Adapter, ein VPN) bleibt aus, bis Sie sie ankreuzen. **Auch Verbindungen zu Schnittstellen ablehnen, die nicht angekreuzt sind** geht weiter: Ein Telefon, das eine Adresse einer nicht angekreuzten Schnittstelle benutzt, wird abgewiesen; Verbindungen zu diesem Computer selbst werden immer angenommen, die Einrichtung bleibt also erreichbar.

## HTTP und HTTPS

Einfache `http://`-Adressen funktionieren auf jedem Gerät und sind die Voreinstellung. HTTPS lässt sich in `config.yaml` einschalten; bei einem selbstsignierten Zertifikat zeigt jedes Gerät beim ersten Mal eine Sicherheitswarnung. Einfaches HTTP ist für ein Publikum einfacher.

## Wenn der Port belegt ist

Wird Port 80 schon von einem anderen Programm benutzt, startet der Server auf einem anderen Port und sagt das in der Konsole. Die Publikumsadressen müssen ihn dann enthalten, zum Beispiel `http://192.168.1.10:8080/`; der Schritt Fahren zeigt die richtige Adresse.
