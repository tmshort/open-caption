# Netzwerk und Verbindung

Zuschauer verbinden ihre eigenen Telefone und Tablets mit dem **Gast-WLAN** des Hauses und öffnen eine Adresse. Es muss nichts installiert werden.

## Was das Haus bereitstellen muss

- Dieser Computer und die Geräte des Publikums müssen im selben Netzwerk sein, oder das Netzwerk muss Gastgeräten erlauben, diesen Computer zu erreichen. Das ist eine Einstellung des WLAN-Systems des Hauses, nicht dieses Programms.
- Viele Gastnetzwerke isolieren Geräte voneinander. Das ist in Ordnung. Die Geräte sprechen nur mit diesem Computer, nie untereinander, solange dieser Computer erreichbar ist.
- Für diesen Computer ist eine Kabelverbindung am besten. Bitten Sie um eine reservierte Adresse (eine „DHCP-Reservierung“), damit sich die Adresse zwischen den Vorstellungen nicht ändert.
- Es ist keine Internetverbindung nötig.

Testen Sie es vor der ersten Show: Verbinden Sie ein Telefon mit dem Gast-WLAN und öffnen Sie die im Schritt **Live** angezeigte Adresse.

## Die Adresse leicht machen

1. **QR-Code** — drucken Sie das Schild aus dem Schritt Live aus und hängen Sie es im Foyer und im Programmheft auf. Das ist die einfachste Möglichkeit.
2. **Eine kurze Adresse** — wenn das Programm Port 80 verwenden kann, hat die Adresse keine Portnummer, etwa `http://192.168.1.10/`.
3. **Ein Name** — das Programm macht sich unter dem Namen **`open-caption.local`** bekannt, sodass `http://open-caption.local/` auf den meisten Telefonen und Computern im selben Netzwerk funktioniert, ohne dass etwas eingerichtet werden muss. Ändern Sie den Namen oder schalten Sie ihn ab unter **Einstellungen → Netzwerkname**. Manche Gast-WLANs blockieren diese Art von Bekanntmachung; verwenden Sie dann die numerische Adresse oder den QR-Code. Das Haus kann auch auf seinem Router einen eigenen Namen anlegen (etwa `captions.theater`), der überall funktioniert.
4. **Automatische Seite** — ein Haus kann Gäste beim Beitritt zum WLAN mit der Startseite des Zugangspunkts („Captive Portal“) auf die Seite schicken. Das Programm beantwortet auch die Standardadresse zur Netzwerkerkennung `/captive-api` für Häuser, die DHCP-Option 114 verwenden, aber das meldet nur, dass das Netzwerk nicht gefangen ist: Es öffnet die Seite nicht von selbst und braucht ein vertrauenswürdiges HTTPS-Zertifikat. Siehe den [Netzwerkleitfaden für das Haus](topic:admin-venue-network-guide).

## HTTP und HTTPS

Einfache `http://`-Adressen funktionieren auf jedem Gerät und sind die Voreinstellung. HTTPS lässt sich in `config.yaml` einschalten; bei einem selbstsignierten Zertifikat zeigt jedes Gerät beim ersten Mal eine Sicherheitswarnung. Einfaches HTTP ist für ein Publikum einfacher.

## Wenn der Port belegt ist

Wird Port 80 schon von einem anderen Programm benutzt, startet der Server auf einem anderen Port und sagt das in der Konsole. Die Publikumsadressen müssen ihn dann enthalten, zum Beispiel `http://192.168.1.10:8080/`; der Schritt Live zeigt die richtige Adresse.
