# Ein vertrauenswürdiges HTTPS-Zertifikat einrichten (optional)

Diese Anleitung richtet sich an die Person, die den Untertitel-Computer betreut. Sie ist **optional**: Die Untertitel funktionieren auch ohne sie. (Die Einstellungen, die sie nutzt, stehen in der [Konfigurationsreferenz](../../configuration-reference.md), Entwurf und Tests in [Dokument 10](../../10-acme-lets-encrypt.md); beide gibt es nur auf Englisch.)

## Brauchen Sie das?

Wahrscheinlich nicht. Die Handys des Publikums öffnen `http://…` ohne jede Warnung, und das ist die Adresse, die der QR-Code verwendet. Eine sichere `https://`-Adresse ist nur für einige Browserfunktionen nötig, die auf sichere Seiten beschränkt sind (etwa den Handybildschirm auf die richtige Art wach zu halten; die Publikumsseite hat eine Ersatzlösung), oder wenn das Haus es wünscht.

Ohne diese Anleitung funktioniert HTTPS trotzdem mit einem **selbstsignierten** Zertifikat, doch jedes Handy zeigt eine Sicherheitswarnung. Mit ihr holt das Programm ein echtes, kostenloses Zertifikat von Let's Encrypt und **erneuert es selbst**, sodass `https://captions.mytheater.org` auf jedem Gerät ohne Warnung öffnet.

## Was Sie vor dem Start brauchen

| Sie brauchen | Wozu |
|---|---|
| Einen **Domainnamen, der Ihnen gehört**, zum Beispiel `mytheater.org`, und darunter einen Namen für die Untertitel, etwa `captions.mytheater.org` | Ein Zertifikat wird für einen öffentlichen Namen ausgestellt. `open-caption.local` und bloße Zahlen bekommen keines |
| Dessen **DNS bei einem Anbieter, den dieses Programm unterstützt** | Das Programm beweist, dass der Name Ihnen gehört, indem es einen vorübergehenden Eintrag in Ihrem DNS anlegt. Das Standardprogramm enthält **Cloudflare, GoDaddy und EasyDNS**; mit `open-caption --version` sehen Sie, was Ihres hat. Ein anderer Anbieter braucht eine andere Programmversion (siehe Ende) |
| **API-Zugangsdaten** für dieses DNS-Konto | Sie erlauben dem Programm, diesen einen Eintrag anzulegen und zu entfernen |
| Dass der Computer **Internetzugang** hat (nur ausgehend) | Er spricht mit Let's Encrypt und Ihrem DNS-Anbieter. Nichts muss *von außen* ins Haus gelangen |
| Eine **feste Adresse** für den Computer im Netz des Hauses (eine DHCP-Reservierung genügt) | Der öffentliche Name wird darauf zeigen |

## Schritte

**1. Besorgen Sie die API-Zugangsdaten.** Im Konto Ihres DNS-Anbieters:

- **EasyDNS:** API-Zugriff aktivieren und ein API-**Token samt Schlüssel** erstellen. Die Einstellungen heißen `EASYDNS_TOKEN` und `EASYDNS_KEY`. (EasyDNS hat eine Sandbox zum Ausprobieren der API; die aktuellen Schritte stehen in deren Dokumentation.)
- **Cloudflare:** ein **API-Token** erstellen, das auf Ihre Zone beschränkt ist und die Berechtigung *Zone → DNS → Edit* hat. Die Einstellung heißt `CLOUDFLARE_DNS_API_TOKEN`.
- **GoDaddy:** einen produktiven API-**Schlüssel samt Secret** erstellen. Die Einstellungen heißen `GODADDY_API_KEY` und `GODADDY_API_SECRET`. (GoDaddy beschränkt seine DNS-API möglicherweise auf bestimmte Kontoarten; prüfen Sie die aktuellen Regeln.)
- Jeder andere Anbieter: Die genauen Namen der Einstellungen stehen in der [lego-Dokumentation](https://go-acme.github.io/lego/dns/).

Ein DNS-Token kann meist **alle** Einträge der Zone ändern, nicht nur den vorübergehenden; geben Sie ihm daher so wenig Macht, wie der Anbieter zulässt. Für den besten Schutz verwenden Sie einen CNAME (siehe „Sicherere Zugangsdaten“ weiter unten).

**2. Lassen Sie den Namen auf den Computer zeigen.** Legen Sie bei Ihrem DNS-Anbieter einen `A`-Eintrag an: `captions.mytheater.org` → die Adresse des Computers im Netz des Hauses (zum Beispiel `192.168.1.20`). Ein öffentlicher Name darf ruhig auf eine private Adresse zeigen; dem Zertifikat ist das gleichgültig.

**3. Bewahren Sie die Zugangsdaten in einer privaten Datei auf.** Legen Sie `/etc/open-caption/dns.env` an (jeder private Ort geht; unter Windows ein Ordner, den nur Ihr Konto lesen kann) mit einer Zeile je Einstellung, zum Beispiel für EasyDNS:

```
EASYDNS_TOKEN=paste-the-token-here
EASYDNS_KEY=paste-the-key-here
```

Machen Sie sie unter macOS und Linux privat: `chmod 600 /etc/open-caption/dns.env`. Das Programm warnt, wenn andere Benutzer sie lesen können.

**4. Bearbeiten Sie `config.yaml`** (die ersten Zeilen, die das Programm ausgibt, nennen den Ort; jede Einstellung ist in der [Referenz](../../configuration-reference.md) erklärt):

```yaml
https:
  enabled: yes
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org      # Let's Encrypt writes here before a certificate expires
    agree_to_terms: yes                 # you accept Let's Encrypt's terms of service
    directory: letsencrypt-staging      # practice first (see step 5)
    dns:
      provider: easydns                 # or cloudflare, godaddy
      credentials_file: /etc/open-caption/dns.env
```

**5. Üben Sie zuerst mit dem Testdienst.** `letsencrypt-staging` ist der Testdienst von Let's Encrypt: Er funktioniert genau wie der echte, aber seine Zertifikate werden von Browsern **nicht als vertrauenswürdig anerkannt** (eine Warnung ist zu erwarten), und die Limits sind großzügig. Starten Sie das Programm.

**6. Prüfen Sie, dass es geklappt hat.** Öffnen Sie **Einrichtung → Einstellungen → Sicherheit → Zertifikat**. Der Status wechselt innerhalb weniger Minuten von *Warten auf das erste Zertifikat* zu **Gültig** und zeigt den Namen, den Aussteller und das Ablaufdatum. Das Protokoll des Programms nennt `HTTPS: got a certificate for … valid until …`. Steht dort *Der letzte Versuch ist fehlgeschlagen*, lesen Sie **Letztes Problem** und die Tabelle unten.

**7. Wechseln Sie zum echten Zertifikat.** Ändern Sie `directory: letsencrypt-staging` in `directory: letsencrypt` (oder löschen Sie die Zeile) und starten Sie das Programm neu. Das echte Zertifikat kommt auf dieselbe Weise. Öffnen Sie `https://captions.mytheater.org/` von einem Handy im WLAN des Hauses: keine Warnung. (War Port 443 belegt, nutzte das Programm den Ersatz, `:8443`; die Adresse braucht dann den Port, `https://captions.mytheater.org:8443/`, und das Protokoll nennt die genaue Adresse.)

**8. Geben Sie den Leuten die Adresse.** Der Schritt „Fahren“ der Einrichtung listet die numerischen Adressen und den `.local`-Namen, aber diese noch nicht. Um dafür ein QR-Schild zu drucken, öffnen Sie bei angemeldeter Einrichtung
`http://localhost/admin?print=qr&url=https%3A%2F%2Fcaptions.mytheater.org%2F` (ändern Sie den Namen; fügen Sie bei Bedarf den Port hinzu) und drucken Sie diese Seite.

## Danach

- **Die Erneuerung ist automatisch.** Das Programm prüft etwa zweimal am Tag und erneuert, wenn noch 30 Tage oder weniger bleiben. Sie müssen nichts tun.
- **Fällt das Internet im Haus aus,** geht nichts kaputt: Das Programm liefert weiter das Zertifikat, das es hat, und versucht es eine Stunde später erneut. Bevor das allererste Zertifikat eintrifft (oder wenn keines kommt), erhalten die Handys das selbstsignierte.
- **Jetzt erneuern** (Einrichtung → Einstellungen → Sicherheit → Zertifikat) holt sofort ein neues Zertifikat. Nutzen Sie es, nachdem Sie ein Problem behoben haben.
- **Das DNS-Token ändern:** Bearbeiten Sie `/etc/open-caption/dns.env`. Die Datei wird beim nächsten Versuch neu gelesen, ein Neustart ist also nicht nötig.
- **Ausschalten:** Setzen Sie `acme: enabled: no` und starten Sie neu. HTTPS verwendet dann das selbstsignierte Zertifikat (oder Ihre eigenen Dateien).

## Wenn etwas schiefgeht

Die Einrichtung zeigt **Letztes Problem**; das Protokoll enthält denselben Text. Die häufigsten Fälle:

| Sie sehen | Wahrscheinliche Ursache und Lösung |
|---|---|
| `the DNS provider "…" is not built into this program` (das Programm bricht beim Start ab) | Diese Programmversion wurde ohne diesen Anbieter gebaut. Nehmen Sie `cloudflare`, `godaddy` oder `easydns`, oder besorgen Sie sich eine Version mit Ihrem (unten) |
| `… credentials information are missing: EASYDNS_TOKEN,EASYDNS_KEY` | Die Namen in der Datei sind falsch oder die Datei ist nicht lesbar. Namen bestehen aus Großbuchstaben, `NAME=Wert`, einer pro Zeile |
| Ein Fehler des DNS-Anbieters wie *unauthorized*, *forbidden* oder eine 4xx-Nummer | Token oder Schlüssel sind falsch, abgelaufen oder dürfen diese Zone nicht ändern. Erstellen Sie neue |
| Lange *Warten auf das erste Zertifikat* oder eine Meldung, der Eintrag sei nicht gefunden | Der neue DNS-Eintrag verbreitet sich langsam. Fügen Sie unter `dns:` `propagation_wait: 3m` hinzu und drücken Sie **Jetzt erneuern** |
| `too many certificates already issued` oder `rate limit` | Let's Encrypt begrenzt Wiederholungen. Warten Sie (die Meldung sagt, wie lange) und nutzen Sie beim Testen `letsencrypt-staging` |
| Der Browser warnt weiterhin | Sie nutzen den Testdienst (wechseln Sie zu `letsencrypt`), oder Sie haben die numerische Adresse oder einen anderen Namen statt `https://captions…` geöffnet, oder die Uhr des Computers oder Handys geht falsch |
| Ein Handy kann den Namen nicht öffnen | Manche Gast-WLANs und Router blockieren öffentliche Namen, die mit privaten Adressen antworten („DNS-Rebinding-Schutz“). Bitten Sie den Betreiber des Netzes, `captions.mytheater.org` zuzulassen, oder nutzen Sie wie zuvor die numerische Adresse und den QR-Code |
| `https.acme: cannot be used together with cert_file/key_file` | Sie haben beides gesetzt. Entfernen Sie eines: automatische Zertifikate **oder** Ihre eigenen Dateien |
| Das Programm startet nicht und listet Probleme in `config.yaml` auf | Jede Zeile nennt die Einstellung und was nicht stimmt. Beheben Sie alle und starten Sie erneut |

## Sicherere Zugangsdaten (optional)

Ein DNS-Token, das Ihre ganze Zone ändern kann, ist ein Risiko, wenn der Computer verloren geht oder geteilt wird. Zwei Wege, es zu verringern:

- Halten Sie die Geheimnisse **aus `config.yaml` heraus**: in der Zugangsdaten-Datei (Schritt 3) oder als Umgebungsvariablen des Dienstes, der das Programm startet (`EASYDNS_TOKEN_FILE=/path/to/file` funktioniert ebenfalls: Die Einstellung wird aus dieser Datei gelesen).
- Nutzen Sie die **CNAME-Delegation**: Legen Sie in Ihrem DNS einen dauerhaften Eintrag `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.other-zone.example` an, wobei `other-zone.example` eine kleine, sonst leere Zone ist. Das Programm folgt ihm automatisch, sodass das Token nur Zugriff auf diese kleine Zone braucht, nicht auf Ihre Hauptzone.

Die Zugangsdaten werden nie ins Protokoll, ins Support-Paket oder in die Einrichtung geschrieben.

## Ein anderer DNS-Anbieter

Das Programm enthält nur die DNS-Anbieter, mit denen es gebaut wurde. Wer es baut, kann mit `make build ACME_PROVIDERS="…"` andere oder alle wählen (die Namen stammen aus der [lego-Liste](https://go-acme.github.io/lego/dns/)); Größen und Schritte stehen in [Dokument 10](../../10-acme-lets-encrypt.md). Wird Ihr Anbieter gar nicht unterstützt, können `exec` (führt ein Skript von Ihnen aus) und `httpreq` (ruft eine Webadresse auf) fast jeden DNS-Dienst ansteuern.
