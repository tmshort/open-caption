# Referenz zu `config.yaml`

`config.yaml` enthält alles über *diesen Computer und dieses Haus*: die Passwörter, Ports, HTTPS, wer die Einrichtung öffnen darf, und die Standardwerte, die eine Show verwendet, wenn ihre eigene Show-Datei nichts sagt. Alles über *eine Show* (Titel, Farben, Schleifendauer, Nachricht vor der Show …) steht stattdessen in der Datei dieser Show (siehe [Eine Show-Datei schreiben](../../../web/help/admin-writing-a-show-file.md), „Einstellungen der Show“); die Ebenen werden unten unter „Wo eine Einstellung entschieden wird“ erklärt.

## Die Datei

- **Wo sie liegt.** `<Datenordner>/config.yaml`; der Datenordner ist `data_dir`, `--data` oder standardmäßig `<Ihr Benutzer-Konfigurationsordner>/open-caption` (unter macOS `~/Library/Application Support/open-caption`, unter Windows `%AppData%\open-caption`, unter Linux `~/.config/open-caption`). Die ersten Zeilen, die das Programm ausgibt, nennen die verwendete Datei. `--config` benennt eine andere Datei.
- **Erster Start.** Gibt es keine Datei, schreibt das Programm eine mit zufälligen **Administrator-** und **Bediener**-Passwörtern und gibt sie einmal in der Konsole aus. Die Datei hat den Modus 0600 (privat für Ihr Konto), und die Passwörter stehen im **Klartext**; halten Sie sie also privat.
- **Sie brauchen nur die Schlüssel, die Sie ändern.** Ein fehlender Schlüssel behält seinen Standardwert (die Tabellen unten). Die erzeugte Datei listet alles auf.
- **Streng.** Ein falsch geschriebener Schlüssel, ein Wert der falschen Art oder außerhalb des Bereichs stoppt das Programm beim Start und listet **alle** Probleme auf, jedes mit dem Namen der Einstellung (zum Beispiel `audience.max_clients: 0 is not a client count: use 1 to 100000`). Es wird nichts geraten. Die einzige Ausnahme: In einer Datei mit einer neueren `config_version`, als das Programm kennt, werden unbekannte Schlüssel mit einer Warnung ignoriert.
- **Die Einrichtung ändert einen Teil.** Wenn Sie in der Einrichtung die Passwörter, den `.local`-Namen, die Standby-Nachricht oder die Protokolldetails ändern, werden genau diese Schlüssel in der Datei neu geschrieben; Ihre Kommentare und alle anderen Einstellungen bleiben erhalten. Alles andere bearbeiten Sie von Hand und **starten dann das Programm neu**. (Eine erneuerte Zertifikatsdatei ist das Einzige, was ohne Neustart übernommen wird; siehe `https.cert_file`.)
- **Boolesche Werte** sind `yes`/`no`/`true`/`false`. **Listen** schreibt man `[a, b]` oder mit einem `- Eintrag` pro Zeile.

## Befehlszeile und Umgebung

Eine Option oder ihre Umgebungsvariable hat für diesen Lauf Vorrang vor der Datei. (Die Datei wird nicht verändert.) Eine Option darf mit einem oder zwei Bindestrichen geschrieben werden (`-data` oder `--data`). `open-caption --help` gibt diese Liste aus, und die **Handbuchseite** (`docs/open-caption.1`; lesen Sie sie mit `man ./docs/open-caption.1` oder kopieren Sie sie in einen `man1`-Ordner wie `/usr/local/share/man/man1/`) beschreibt dieselben Optionen sowie Dateien, Signale und den Exit-Status.

| Option | Umgebung | Wirkung |
|---|---|---|
| `--config FILE` | `OPEN_CAPTION_CONFIG` | Diese Einstellungsdatei verwenden (Standard `<Datenordner>/config.yaml`) |
| `--data FOLDER` | `OPEN_CAPTION_DATA` | Der Datenordner (Shows, Sitzungen, `live.json`, Zertifikate). Hat Vorrang vor `data_dir` |
| `--http-addr ADDR` | `OPEN_CAPTION_HTTP_ADDR` | Hat Vorrang vor `http.addr`, zum Beispiel `:80` oder `:8080` |
| `--https-addr ADDR` | `OPEN_CAPTION_HTTPS_ADDR` | Schaltet HTTPS an dieser Adresse **ein** (hat Vorrang vor `https.enabled` und `https.addr`) |
| `--log-level LEVEL` | `OPEN_CAPTION_LOG_LEVEL` | Hat bei jedem Programmstart Vorrang vor `log.level` (die Einrichtung sagt es, wenn es gilt) |
| `--log-file FILE` | `OPEN_CAPTION_LOG` | Hängt das Protokoll zusätzlich an diese Datei an |
| `--network-interfaces LIST` | `OPEN_CAPTION_NETWORK_INTERFACES` | Hat bei jedem Programmstart Vorrang vor `network.interfaces`: eine durch Kommas getrennte Liste von Schnittstellennamen wie `Ethernet` oder `en0,en4` (die Einrichtung sagt es, wenn es gilt) |
| `--network-restrict` | `OPEN_CAPTION_NETWORK_RESTRICT` | Lehnt außerdem Verbindungen zu Adressen ab, die nicht angeboten werden, bei jedem Programmstart (wie `network.restrict_listening: yes`). Die Variable zählt, wenn sie `yes`, `true`, `on` oder `1` ist |
| `--no-browser` | | Öffnet die Einrichtung beim Start nicht in einem Browser (wie `defaults.open_browser: no`) |
| `--version` | | Gibt die Version und die für automatische Zertifikate eingebauten DNS-Anbieter aus und beendet sich |
| `--help` | | Gibt die Optionen und Umgebungsvariablen aus, dazu wo man mehr liest, und beendet sich |
| `--check-show FILE [TRANSLATION ...]` | | **Prüft eine Show-Datei** und gibt einen Bericht aus, wie ein „Lint“, dann Ende: kein Server, keine `config.yaml`. Siehe „Eine Show-Datei prüfen“ unten |
| `--assets FOLDER` | | Der Bilderordner für `--check-show` (Standard: der Ordner `assets` neben der Show-Datei) |
| `--strict` | | Mit `--check-show`: Auch Warnungen ergeben den Exit-Status 1 |
| `--dev`, `--show-file` | | **Nur für Entwickler**: nicht authentifizierte Testzugänge auf diesem Computer, kein Standby; nie in einem Haus verwenden |

## Eine Show-Datei prüfen

`open-caption --check-show show.md` analysiert eine Show-Datei, ohne sie in das Programm zu laden: Es führt denselben Parser und dieselben Prüfungen aus wie die Problemliste im Schritt Show-Datei der Einrichtung und gibt aus, was diese Liste zeigen würde.

```
$ open-caption --check-show show.md show.es.md
show.md: info: 4 captions, 1 images, 0 fullscreen, 1 loops, 1 sections, 0 notes.
show.md:8: warning: Image "curtain.jpg" not found in assets.
show.md:13: error: [song] was opened but never closed with [/song]. The lyrics so far were kept.
show.md: 8 cues; 2 speakers: ALICE, BOB
show.es.md: 8 cues; language es

images checked against assets (1 found)
missing images: curtain.jpg (line 8 of show.md)
1 error, 1 warning, 1 note
```

- Jedes Problem hat die Form `file:line: severity: message` (Schweregrad `error`, `warning` oder `note`), die Editoren und CI-Systeme verstehen. Eine Zeile über die ganze Datei hat keine Zeilennummer.
- Dateien nach der ersten sind **Übersetzungen**: Jede wird als Übersetzung geparst und mit der Hauptdatei verglichen (gleiche Anzahl und Art von Stichwörtern in derselben Reihenfolge), wie es die Einrichtung tut. Die Sprache steht in der Zeile `[language]` der Datei oder in ihrem Namen (`show.es.md`); zwei Dateien für eine Sprache sind ein Fehler.
- Bilder werden im Ordner `assets` neben der Hauptdatei geprüft oder in `--assets FOLDER`. Ohne Ordner werden sie nicht geprüft, und der Bericht sagt das. Ein fehlendes Bild nennt den Ordner, in dem gesucht wurde (der Editor der Einrichtung sagt stattdessen, man solle es im Schritt Bilder hochladen).
- Ein Bild im Ordner, das keine Show-Datei zeigt, wird als `unused images (…): name.png` aufgeführt und als Hinweis gezählt, nicht als Warnung (der Schritt Bilder der Einrichtung schreibt darunter „Nicht verwendet“). Bilder, die nur eine Übersetzung verwendet, gelten als verwendet.
- `-` als Dateiname liest die Standardeingabe.
- **Exit-Status:** `0` keine Fehler, `1` Fehler (oder Warnungen mit `--strict`), `2` eine Datei konnte nicht gelesen werden. Setzen Sie andere Optionen (`--assets`, `--strict`) **vor** `--check-show`: Die Dateien danach sind Übersetzungen.

## Einstellungen

Standardwerte sind das, was Sie ohne den Schlüssel erhalten. „Einrichtung“ sagt, wo die Einrichtung sie ändern kann.

### Oberste Ebene

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `config_version` | `1` | Vom Programm geschrieben. Eine höhere Nummer, als das Programm kennt, bedeutet, dass die Datei von einer neueren Version stammt |
| `admin_password` | erzeugt | Das Passwort für die Einrichtung (`/admin`). Darf nicht leer sein (mindestens 6 Zeichen, wenn es in der Einrichtung geändert wird). *Einrichtung: Einstellungen → Sicherheit* |
| `operator_password` | erzeugt | Das Passwort für die Bedienerseite (`/operator`); das Administratorpasswort funktioniert dort auch. Darf nicht leer sein. *Einrichtung: Einstellungen → Sicherheit* |
| `data_dir` | leer | Der Datenordner. Leer bedeutet den oben genannten Standardordner. Darf keine Datei sein |

### `http`

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `http.addr` | `:80` | Wo der unverschlüsselte HTTP-Server lauscht: `:80` (alle Adressen, Port 80), `192.168.1.20:80` oder `127.0.0.1:8080` (nur dieser Computer). Darf nicht leer sein |
| `http.fallback_addr` | `:8080` | Wird mit einer sichtbaren Warnung verwendet, wenn `addr` nicht benutzt werden kann (Port 80 ist oft belegt oder braucht Rechte). Leer bedeutet keinen Ersatz |

### `https`

HTTPS ist standardmäßig **aus**. Es gibt drei Wege zu einem Zertifikat; wählen Sie einen.

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `https.enabled` | `no` | Liefert HTTPS zusätzlich zu HTTP, auf getrennten Ports (es gibt keine Umleitung). Braucht einen der drei unten |
| `https.addr` | `:443` | Der HTTPS-Port. Pflicht, wenn aktiviert |
| `https.fallback_addr` | `:8443` | Wird verwendet, wenn `addr` nicht geöffnet werden kann |
| `https.self_signed` | `yes` | Ohne Zertifikatsdateien und ohne `acme`: Das Programm erzeugt ein eigenes Zertifikat (abgelegt in `<Datenordner>/tls/`). Browser warnen davor, bis es akzeptiert wird. `no` ohne Dateien ist ein Fehler |
| `https.cert_file`, `https.key_file` | leer | Ihr eigenes Zertifikat und der Schlüssel (PEM). Setzen Sie **beide** oder keines; beide Dateien müssen existieren. **Die Dateien werden überwacht**: Werden sie ersetzt (zum Beispiel durch `certbot`), gilt das neue Paar für neue Verbindungen nach etwa 30 Sekunden, ohne Neustart. Ein Paar, das sich nicht laden lässt, behält das alte Zertifikat und protokolliert eine Warnung |

### `https.acme` (automatische Zertifikate)

Aus, außer bei `enabled: yes`. Braucht `https.enabled: yes`, einen öffentlichen Domainnamen, der Ihnen gehört, und einen DNS-Anbieter, mit dem dieses Programm gebaut wurde. Nicht kombinierbar mit `cert_file`/`key_file`. Wie man es einrichtet, Schritt für Schritt: die [Anleitung für ein vertrauenswürdiges HTTPS-Zertifikat](https-certificate-guide.md); die Auswahl der DNS-Anbieter beim Bauen und die Größen stehen am Ende dieser Anleitung.

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `https.acme.enabled` | `no` | Holt und erneuert ein Zertifikat automatisch (ACME, DNS-01) |
| `https.acme.domain` | | Der öffentliche Name, zum Beispiel `captions.example.org`. Ein Name: kein Platzhalter, kein `.local`/`.lan`/`.internal`, keine IP-Adresse. Er muss für die Handys des Publikums auf die Adresse dieses Computers auflösen |
| `https.acme.email` | | Wohin die Zertifizierungsstelle Ablaufhinweise schickt |
| `https.acme.agree_to_terms` | `no` | **Muss `yes` sein**: Es hält fest, dass Sie die Nutzungsbedingungen der Zertifizierungsstelle akzeptieren |
| `https.acme.directory` | `letsencrypt` | `letsencrypt`, `letsencrypt-staging` (zum Ausprobieren: Browser vertrauen ihm nicht) oder die `https://`-Adresse eines anderen ACME-Verzeichnisses |
| `https.acme.renew_before_days` | `30` | Erneuert, wenn noch so viele Tage oder weniger bleiben (1 bis 60) |
| `https.acme.dns.provider` | | Der DNS-Anbieter mit seinem [lego-Namen](https://go-acme.github.io/lego/dns/) (`cloudflare`, `godaddy`, `easydns`, …). Muss im Programm enthalten sein (`--version` listet sie auf) |
| `https.acme.dns.credentials` | leer | Die eigenen Einstellungen des Anbieters als Zuordnung: für EasyDNS `EASYDNS_TOKEN: …` und `EASYDNS_KEY: …`. Namen bestehen aus Großbuchstaben, Ziffern und Unterstrichen. Die Umgebung und `credentials_file` haben Vorrang vor diesen |
| `https.acme.dns.credentials_file` | leer | Eine Datei mit Zeilen `NAME=value` mit denselben Einstellungen, bei jedem Versuch neu gelesen. Muss existieren; mit `chmod 600` schützen |
| `https.acme.dns.propagation_wait` | leer | Eine feste Wartezeit (`90s`, `2m`, höchstens `30m`), bis sich der neue DNS-Eintrag verbreitet hat, statt danach zu suchen |

### `defaults`

Was eine Show verwendet, wenn ihre eigene Show-Datei es nicht sagt. Der `[settings]`-Block der Show-Datei hat Vorrang vor jedem Wert hier (siehe [Eine Show-Datei schreiben](../../../web/help/admin-writing-a-show-file.md)). Der Schritt Erscheinungsbild der Einrichtung schreibt die Show-Datei, nicht diese.

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `defaults.language` | `en` | Die Hauptsprache einer Show, deren Datei keine nennt: zwei oder drei Buchstaben, optional mit Region (`en`, `es`, `pt-br`) |
| `defaults.loop_item_seconds` | `10` | Wie lange jedes Element einer Schleife vor der Show oder in der Pause stehen bleibt, in Sekunden (1 bis 3600), sofern das Element nichts anderes sagt |
| `defaults.idle_message` | `Captions will begin shortly` | Was das Publikum liest, bevor die erste Zeile gesendet wird, wenn eine Show keine Nachricht vor der Show hat (bis zu 200 Zeichen) |
| `defaults.open_browser` | `yes` | Öffnet die Einrichtung beim Programmstart in einem Browser (keine Show-Einstellung; `--no-browser` schaltet es für einen Lauf aus) |
| `defaults.look.background` | `#000000` | Der Untertitelhintergrund: `#rgb`, `#rrggbb` oder `black`, `white`, `yellow`, `red`, `green`, `blue`, `cyan`, `magenta`, `orange`, `gray` |
| `defaults.look.text_color` | `#ffffff` | Die Textfarbe der Untertitel (gleiche Formen) |
| `defaults.look.speaker_name_color` | `#ffffff` | Die Farbe der Sprechernamen (gleiche Formen) |
| `defaults.look.font` | `system` | `system`, `serif`, `mono`, `dyslexic` oder `hyperlegible` |
| `defaults.look.text_size_percent` | `100` | Textgröße der Untertitel, 50 bis 300 |
| `defaults.look.speaker_names` | `label` | `label` (Namen in einer Farbe), `color` (eine andere Farbe für jede Figur) oder `hidden` |

### `audience`

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `audience.max_clients` | `500` | Die meisten verbundenen Publikumsgeräte **pro Sprache** (1 bis 100000); weitere werden abgewiesen, bis einige gehen |
| `audience.scrollback_lines` | `200` | Wie viele frühere Zeilen ein Gerät zurückscrollen kann (20 bis 2000) |

### `access`

Wer die Einrichtung und die Bedienerseite öffnen darf. Die Publikumsseite ist immer für alle offen.

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `access.admin_allow` | `[]` | Computer, die die Einrichtung öffnen dürfen. Leer bedeutet: **Jeder Computer im Netz darf das Passwort versuchen**. Einträge: `localhost`, `private` (die üblichen Heim- und Büronetze), eine Adresse (`192.168.1.50`) oder ein Netz (`192.168.1.0/24`). Beispiel: `[localhost]` |
| `access.operator_allow` | `[]` | Dasselbe für die Bedienerseite |
| `access.local_auto_login` | `yes` | Ein Browser auf dem Computer, auf dem das Programm läuft (geöffnet unter `localhost`, `127.0.0.1` oder `[::1]`), wird ohne Passwort als Administrator angemeldet. Andere Computer verwenden immer ein Passwort. Schalten Sie es auf einem Computer aus, an dem sich andere Personen anmelden können |

### `log`, `mdns`, `standby`

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `log.level` | `info` | `debug` (jede Anfrage), `info` (Ereignisse und Anmeldungen), `warn` oder `error`. *Einrichtung: Einstellungen → Info* |
| `mdns.enabled` | `yes` | Macht diesen Computer im lokalen Netz als `<Name>.local` bekannt |
| `mdns.name` | `open-caption` | Der Name: nur Buchstaben, Ziffern und Bindestriche, bis zu 63 Zeichen (`.local` wird angehängt). *Einrichtung: Einstellungen → Netzwerk* |
| `standby.message` | leer | Die Zeile, die das Publikum auf dem Standby-Bildschirm sieht. Leer bedeutet „Captions will begin shortly“ in der Sprache jedes Zuschauers. Eine Zeile, bis zu 200 Zeichen. *Einrichtung: Einstellungen → Publikumsbildschirm* (das Logo ist eine dort gewählte Datei, keine Einstellung) |

### `network`

| Schlüssel | Standard | Bedeutung |
|---|---|---|
| `network.interfaces` | `[]` | Die verwendeten Netzwerkschnittstellen, nach Namen (`Ethernet`, `Wi-Fi`, `en0`, `eth0`): nur ihre Adressen werden dem Publikum gegeben, in der Einrichtung gelistet, in ein neues selbstsigniertes HTTPS-Zertifikat aufgenommen und als `<Name>.local` angekündigt. Leer bedeutet, dass noch nichts gewählt wurde: alle Schnittstellen werden verwendet. Ist einmal gewählt, bleibt eine später auftauchende Schnittstelle (ein USB-Adapter, ein VPN) aus, bis sie hinzugefügt wird. Sind zwei verwendete Schnittstellen im selben Netzwerk (ein Kabel und WLAN zugleich), wird nur eine Adresse angeboten: die kabelgebundene, sonst die der Hauptverbindung; lassen Sie die andere weg, um zu wählen. *Einrichtung: Einstellungen → Netzwerk* |
| `network.restrict_listening` | `no` | **Lehnt außerdem Verbindungen ab**, die zu einer Adresse einer nicht verwendeten Schnittstelle gemacht werden (ein Telefon, das sie benutzt, wird abgewiesen), statt die Adresse nur aus den Listen zu lassen. Verbindungen zu diesem Computer selbst (`localhost`) werden immer bedient, die Einrichtung lässt sich also zum Ändern öffnen. Die Einrichtung schaltet nie die Schnittstelle aus, über die Sie sie benutzen. *Einrichtung: Einstellungen → Netzwerk* |

Die Adressen einer nicht verwendeten Schnittstelle werden im Schritt Teilen und beim Start nicht aufgelistet, dem Publikum nicht gegeben, nicht als `<Name>.local` angekündigt und nicht in ein neues selbstsigniertes HTTPS-Zertifikat aufgenommen. Aufgelistet werden nur verbundene Schnittstellen, und je Netzwerk wird nur eine Adresse angeboten. Standardmäßig wählt das nur, was *angeboten* wird: Der Server nimmt weiterhin auf allen Adressen Verbindungen an, bis `network.restrict_listening` eingeschaltet wird.

## Wo eine Einstellung entschieden wird

Von schwächer zu stärker: der eingebaute Wert des Programms → `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) → das `[settings]` der Show-Datei. Eine Befehlszeilenoption oder Umgebungsvariable hat Vorrang vor der Datei. Die Einstellungen des Hauses (Ports, HTTPS, Passwörter, `access`, `data_dir`, `audience.max_clients`, `mdns`, `network`, `standby`, `log`) gibt es nur hier.

## Ein vollständiges Beispiel

```yaml
config_version: 1
admin_password: "set-a-long-one"
operator_password: "another-one"
http:
  addr: ":80"
  fallback_addr: ":8080"
https:
  enabled: yes
  addr: ":443"
  fallback_addr: ":8443"
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org
    agree_to_terms: yes
    directory: letsencrypt
    dns:
      provider: easydns
      credentials_file: /etc/open-caption/dns.env
mdns:
  enabled: yes
  name: open-caption
network:
  interfaces: []
  restrict_listening: no
standby:
  message: "Welcome to the Little Theatre"
defaults:
  language: en
  loop_item_seconds: 10
  idle_message: "Captions will begin shortly"
  look:
    font: hyperlegible
    text_size_percent: 120
audience:
  max_clients: 500
  scrollback_lines: 200
access:
  admin_allow: [localhost]
  operator_allow: [private]
  local_auto_login: yes
log:
  level: info
```
