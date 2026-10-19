# Riferimento di `config.yaml`

`config.yaml` contiene tutto ciò che riguarda *questo computer e questo locale*: le password, le porte, HTTPS, chi può aprire la configurazione e i valori predefiniti che uno spettacolo usa quando il suo file non dice nulla. Tutto ciò che riguarda *un singolo spettacolo* (titolo, colori, durata dei cicli, messaggio pre-spettacolo…) sta invece nel file di quello spettacolo (vedi [Scrivere un file dello spettacolo](../../../web/help/admin-writing-a-show-file.md), «Impostazioni dello spettacolo»); i livelli sono spiegati più sotto in «Dove si decide un'impostazione».

## Il file

- **Dove si trova.** `<cartella dati>/config.yaml`; la cartella dati è `data_dir`, `--data`, oppure quella predefinita `<la tua cartella di configurazione utente>/open-caption` (su macOS `~/Library/Application Support/open-caption`, su Windows `%AppData%\open-caption`, su Linux `~/.config/open-caption`). Le prime righe che il programma stampa dicono quale file sta usando. `--config` indica un altro file.
- **Primo avvio.** Se non c'è il file, il programma ne scrive uno con password casuali di **amministratore** e di **operatore** e le stampa una sola volta nella console. Il file ha modalità 0600 (privato per il tuo account) e le password sono in **chiaro**, quindi tienilo privato.
- **Ti servono solo le chiavi che cambi.** Una chiave mancante mantiene il suo valore predefinito (le tabelle qui sotto). Il file generato elenca tutto.
- **Rigoroso.** Una chiave scritta male, un valore del tipo sbagliato o fuori intervallo ferma il programma all'avvio elencando **tutti** i problemi, ciascuno con il nome dell'impostazione (per esempio `audience.max_clients: 0 is not a client count: use 1 to 100000`). Non si indovina nulla. L'unica eccezione: in un file marcato con un `config_version` più recente di quello che il programma conosce, le chiavi sconosciute vengono ignorate, con un avviso.
- **La configurazione ne modifica una parte.** Cambiare le password, il nome `.local`, il messaggio di attesa o il livello del registro nella configurazione riscrive quelle chiavi nel file e conserva i tuoi commenti e tutte le altre impostazioni. Tutto il resto si modifica a mano, poi **riavvia il programma**. (Un file di certificato rinnovato è l'unica cosa recepita senza riavvio; vedi `https.cert_file`.)
- **I booleani** sono `yes`/`no`/`true`/`false`. **Gli elenchi** si scrivono `[a, b]` o con un `- elemento` per riga.

## Riga di comando e ambiente

Un'opzione, o la sua variabile d'ambiente, ha la precedenza sul file per quell'esecuzione. (Il file non viene modificato.) Un'opzione si può scrivere con uno o due trattini (`-data` o `--data`). `open-caption --help` stampa questo elenco, e la **pagina di manuale** (`docs/open-caption.1`; leggila con `man ./docs/open-caption.1`, oppure copiala in una cartella `man1` come `/usr/local/share/man/man1/`) descrive le stesse opzioni, oltre a file, segnali e stato di uscita.

| Opzione | Ambiente | Effetto |
|---|---|---|
| `--config FILE` | `OPEN_CAPTION_CONFIG` | Usa questo file di impostazioni (predefinito `<cartella dati>/config.yaml`) |
| `--data FOLDER` | `OPEN_CAPTION_DATA` | La cartella dati (spettacoli, sessioni, `live.json`, certificati). Ha la precedenza su `data_dir` |
| `--http-addr ADDR` | `OPEN_CAPTION_HTTP_ADDR` | Ha la precedenza su `http.addr`, per esempio `:80` o `:8080` |
| `--https-addr ADDR` | `OPEN_CAPTION_HTTPS_ADDR` | **Attiva** HTTPS a questo indirizzo (ha la precedenza su `https.enabled` e `https.addr`) |
| `--log-level LEVEL` | `OPEN_CAPTION_LOG_LEVEL` | Ha la precedenza su `log.level` a ogni avvio del programma (la configurazione lo dice quando è in vigore) |
| `--log-file FILE` | `OPEN_CAPTION_LOG` | Aggiunge anche il registro a questo file |
| `--network-include LIST` | `OPEN_CAPTION_NETWORK_INCLUDE` | Ha la precedenza su `network.include` a ogni avvio del programma: un elenco separato da virgole, come `192.168.1.0/24,10.0.0.0/8` (la configurazione lo dice quando è in vigore) |
| `--network-exclude LIST` | `OPEN_CAPTION_NETWORK_EXCLUDE` | Lo stesso per `network.exclude` |
| `--network-restrict` | `OPEN_CAPTION_NETWORK_RESTRICT` | Rifiuta anche le connessioni fatte a indirizzi non offerti, a ogni avvio del programma (come `network.restrict_listening: yes`). La variabile conta quando vale `yes`, `true`, `on` o `1` |
| `--no-browser` | | Non apre la configurazione in un browser all'avvio (come `defaults.open_browser: no`) |
| `--version` | | Stampa la versione e i provider DNS inclusi per i certificati automatici, poi esce |
| `--help` | | Stampa le opzioni e le variabili d'ambiente, e dove leggere di più, poi esce |
| `--check-show FILE [TRANSLATION ...]` | | **Controlla un file dello spettacolo** e stampa un rapporto, come un «lint», poi esce: nessun server, nessun `config.yaml`. Vedi «Controllare un file dello spettacolo» più sotto |
| `--assets FOLDER` | | La cartella delle immagini per `--check-show` (predefinita: la cartella `assets` accanto al file dello spettacolo) |
| `--strict` | | Con `--check-show`: anche gli avvisi danno lo stato di uscita 1 |
| `--dev`, `--show-file` | | **Solo per sviluppatori**: punti di accesso di prova senza autenticazione su questo computer, senza attesa; non usarli mai in un locale |

## Controllare un file dello spettacolo

`open-caption --check-show show.md` analizza un file dello spettacolo senza caricarlo nel programma: esegue lo stesso analizzatore e gli stessi controlli dell'elenco dei problemi del passo File dello spettacolo della configurazione e stampa ciò che quell'elenco mostrerebbe.

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

- Ogni problema ha la forma `file:line: severity: message` (gravità `error`, `warning` o `note`), quella che capiscono gli editor e i sistemi di integrazione continua. Una riga che riguarda tutto il file non ha numero di riga.
- I file dopo il primo sono **traduzioni**: ciascuno viene analizzato come traduzione e confrontato con il file principale (lo stesso numero e tipo di indicazioni, nello stesso ordine), come fa la configurazione. La lingua è la riga `[language]` del file o il suo nome (`show.es.md`); due file per la stessa lingua sono un errore.
- Le immagini vengono controllate nella cartella `assets` accanto al file principale, oppure in `--assets FOLDER`. Senza cartella non vengono controllate, e il rapporto lo dice. Un'immagine mancante indica la cartella in cui è stata cercata (l'editor della configurazione dice invece di caricarla nel passo Immagini).
- `-` come nome del file legge lo standard input.
- **Stato di uscita:** `0` nessun errore, `1` errori (o avvisi con `--strict`), `2` un file non si è potuto leggere. Metti le altre opzioni (`--assets`, `--strict`) **prima** di `--check-show`: i file che lo seguono sono traduzioni.

## Impostazioni

I valori predefiniti sono quelli che ottieni senza la chiave. «Configurazione» dice dove la configurazione può cambiarla.

### Livello principale

| Chiave | Predefinito | Significato |
|---|---|---|
| `config_version` | `1` | Scritta dal programma. Un numero maggiore di quello che il programma conosce significa che il file viene da una versione più recente |
| `admin_password` | generata | La password della configurazione (`/admin`). Non può essere vuota (6 o più caratteri se cambiata nella configurazione). *Configurazione: Impostazioni → Sicurezza* |
| `operator_password` | generata | La password della pagina dell'operatore (`/operator`); anche quella dell'amministratore vale lì. Non può essere vuota. *Configurazione: Impostazioni → Sicurezza* |
| `data_dir` | vuoto | La cartella dati. Vuoto significa la cartella predefinita qui sopra. Non può essere un file |

### `http`

| Chiave | Predefinito | Significato |
|---|---|---|
| `http.addr` | `:80` | Dove ascolta il server HTTP non cifrato: `:80` (tutti gli indirizzi, porta 80), `192.168.1.20:80` o `127.0.0.1:8080` (solo questo computer). Non può essere vuoto |
| `http.fallback_addr` | `:8080` | Usato, con un avviso ben visibile, se `addr` non si può usare (la porta 80 è spesso occupata o richiede privilegi). Vuoto significa nessuna alternativa |

### `https`

HTTPS è **disattivato** per impostazione predefinita. Ci sono tre modi per ottenere un certificato; scegline uno.

| Chiave | Predefinito | Significato |
|---|---|---|
| `https.enabled` | `no` | Serve HTTPS oltre a HTTP, su porte separate (non c'è reindirizzamento). Richiede uno dei tre qui sotto |
| `https.addr` | `:443` | La porta HTTPS. Obbligatoria se attivato |
| `https.fallback_addr` | `:8443` | Usata se `addr` non si può aprire |
| `https.self_signed` | `yes` | Senza file di certificato e senza `acme`: il programma crea un proprio certificato (conservato in `<cartella dati>/tls/`). I browser lo segnalano finché non viene accettato. `no` senza file è un errore |
| `https.cert_file`, `https.key_file` | vuoto | Il tuo certificato e la chiave (PEM). Indica **entrambi** o nessuno; entrambi i file devono esistere. **I file sono sorvegliati**: quando vengono sostituiti (per esempio da `certbot`), la nuova coppia viene usata per le nuove connessioni entro circa 30 secondi, senza riavvio. Una coppia che non si può caricare mantiene il vecchio certificato e registra un solo avviso |

### `https.acme` (certificati automatici)

Disattivato salvo `enabled: yes`. Richiede `https.enabled: yes`, un nome di dominio pubblico tuo e un provider DNS con cui questo programma è stato compilato. Non si può combinare con `cert_file`/`key_file`. Come configurarlo, passo dopo passo: la [guida al certificato HTTPS attendibile](https-certificate-guide.md); le scelte dei provider DNS in fase di compilazione e le dimensioni sono alla fine di quella guida.

| Chiave | Predefinito | Significato |
|---|---|---|
| `https.acme.enabled` | `no` | Ottiene e rinnova un certificato automaticamente (ACME, DNS-01) |
| `https.acme.domain` | | Il nome pubblico, per esempio `captions.example.org`. Un solo nome: niente caratteri jolly, niente `.local`/`.lan`/`.internal`, niente indirizzo IP. Deve risolvere, per i telefoni del pubblico, all'indirizzo di questo computer |
| `https.acme.email` | | Dove l'autorità di certificazione invia gli avvisi di scadenza |
| `https.acme.agree_to_terms` | `no` | **Deve essere `yes`**: registra che accetti i termini di servizio dell'autorità di certificazione |
| `https.acme.directory` | `letsencrypt` | `letsencrypt`, `letsencrypt-staging` (per provare: i browser non si fidano) o l'indirizzo `https://` di un'altra directory ACME |
| `https.acme.renew_before_days` | `30` | Rinnova quando mancano questi giorni o meno (da 1 a 60) |
| `https.acme.dns.provider` | | Il provider DNS, con il suo [nome in lego](https://go-acme.github.io/lego/dns/) (`cloudflare`, `godaddy`, `easydns`, …). Deve essere incluso nel programma (`--version` li elenca) |
| `https.acme.dns.credentials` | vuoto | Le impostazioni proprie del provider, come mappa: per EasyDNS `EASYDNS_TOKEN: …` e `EASYDNS_KEY: …`. I nomi sono lettere maiuscole, cifre e trattini bassi. L'ambiente e `credentials_file` prevalgono su queste |
| `https.acme.dns.credentials_file` | vuoto | Un file di righe `NAME=value` con le stesse impostazioni, riletto a ogni tentativo. Deve esistere; proteggilo con `chmod 600` |
| `https.acme.dns.propagation_wait` | vuoto | Un'attesa fissa (`90s`, `2m`, al massimo `30m`) perché il nuovo record DNS si propaghi, invece di controllarlo |

### `defaults`

Ciò che uno spettacolo usa quando il suo file non lo dice. Il blocco `[settings]` del file dello spettacolo prevale su ogni valore qui (vedi [Scrivere un file dello spettacolo](../../../web/help/admin-writing-a-show-file.md)). Il passo Aspetto della configurazione scrive il file dello spettacolo, non questi.

| Chiave | Predefinito | Significato |
|---|---|---|
| `defaults.language` | `en` | La lingua principale di uno spettacolo il cui file non ne indica una: due o tre lettere, con regione facoltativa (`en`, `es`, `pt-br`) |
| `defaults.loop_item_seconds` | `10` | Quanto resta ogni elemento di un ciclo pre-spettacolo o d'intervallo, in secondi (da 1 a 3600), salvo che l'elemento dica altro |
| `defaults.idle_message` | `Captions will begin shortly` | Ciò che il pubblico legge prima che venga inviata la prima riga quando uno spettacolo non ha un messaggio pre-spettacolo (fino a 200 caratteri) |
| `defaults.open_browser` | `yes` | Apre la configurazione in un browser all'avvio del programma (non è un'impostazione dello spettacolo; `--no-browser` la disattiva per un'esecuzione) |
| `defaults.look.background` | `#000000` | Lo sfondo dei sottotitoli: `#rgb`, `#rrggbb` oppure `black`, `white`, `yellow`, `red`, `green`, `blue`, `cyan`, `magenta`, `orange`, `gray` |
| `defaults.look.text_color` | `#ffffff` | Il colore del testo dei sottotitoli (stesse forme) |
| `defaults.look.speaker_name_color` | `#ffffff` | Il colore dei nomi dei personaggi (stesse forme) |
| `defaults.look.font` | `system` | `system`, `serif`, `mono`, `dyslexic` o `hyperlegible` |
| `defaults.look.text_size_percent` | `100` | Dimensione del testo dei sottotitoli, da 50 a 300 |
| `defaults.look.speaker_names` | `label` | `label` (nomi di un solo colore), `color` (un colore diverso per ogni personaggio) o `hidden` |

### `audience`

| Chiave | Predefinito | Significato |
|---|---|---|
| `audience.max_clients` | `500` | Il massimo di dispositivi del pubblico connessi **per lingua** (da 1 a 100000); gli altri vengono respinti finché qualcuno non se ne va |
| `audience.scrollback_lines` | `200` | Quante righe precedenti un dispositivo può scorrere all'indietro (da 20 a 2000) |

### `access`

Chi può aprire la configurazione e la pagina dell'operatore. La pagina del pubblico è sempre aperta a tutti.

| Chiave | Predefinito | Significato |
|---|---|---|
| `access.admin_allow` | `[]` | Computer autorizzati ad aprire la configurazione. Vuoto significa che **qualsiasi computer della rete può provare la password**. Voci: `localhost`, `private` (le usuali reti domestiche e d'ufficio), un indirizzo (`192.168.1.50`) o una rete (`192.168.1.0/24`). Esempio: `[localhost]` |
| `access.operator_allow` | `[]` | Lo stesso, per la pagina dell'operatore |
| `access.local_auto_login` | `yes` | Un browser sul computer che esegue il programma (aperto a `localhost`, `127.0.0.1` o `[::1]`) entra come amministratore senza password. Gli altri computer usano sempre una password. Disattivalo su un computer in cui possono accedere altre persone |

### `log`, `mdns`, `standby`

| Chiave | Predefinito | Significato |
|---|---|---|
| `log.level` | `info` | `debug` (ogni richiesta), `info` (eventi e accessi), `warn` o `error`. *Configurazione: Impostazioni → Informazioni* |
| `mdns.enabled` | `yes` | Annuncia questo computer come `<nome>.local` sulla rete locale |
| `mdns.name` | `open-caption` | Il nome: solo lettere, cifre e trattini, fino a 63 caratteri (viene aggiunto `.local`). *Configurazione: passo Conduci* |
| `standby.message` | vuoto | La riga che il pubblico vede sulla schermata di attesa. Vuoto significa «Captions will begin shortly» nella lingua di ogni spettatore. Una riga, fino a 200 caratteri. *Configurazione: Impostazioni → Schermo del pubblico* (il logo è un file scelto lì, non un'impostazione) |

### `network`

| Chiave | Predefinito | Significato |
|---|---|---|
| `network.include` | `[]` | Offre al pubblico solo questi indirizzi di questo computer. Vuoto significa tutti gli indirizzi. Voci: un indirizzo (`192.168.1.50`), una rete (`192.168.1.0/24`) o `private` (le solite reti domestiche e d'ufficio). Esempio: `[192.168.1.0/24]` |
| `network.exclude` | `[]` | Non offrire mai questi (gli stessi tipi di voci); ha la precedenza su `include`. Serve per gli adattatori di macchine virtuali, VPN o Docker. Esempio: `[192.168.56.0/24]` |
| `network.restrict_listening` | `no` | **Rifiuta anche le connessioni** fatte a un indirizzo che `include` / `exclude` non offrono (un telefono che usa un indirizzo nascosto viene respinto), invece di lasciare l'indirizzo solo fuori dagli elenchi. Le connessioni al computer stesso (`localhost`) sono sempre servite, quindi la configurazione si può aprire per cambiarlo. La configurazione si rifiuta di attivarlo mentre la usi attraverso un indirizzo che respingerebbe. *Configurazione: passo Conduci → Scegli gli indirizzi…* |

Un indirizzo escluso non compare nel passo Conduci né all'avvio, non viene dato al pubblico, non è annunciato come `<nome>.local` e non entra in un nuovo certificato HTTPS autofirmato. Vengono offerti solo gli indirizzi degli adattatori di rete connessi. Per impostazione predefinita questo sceglie solo ciò che viene *offerto*: il server accetta comunque connessioni su tutti gli indirizzi, finché non si attiva `network.restrict_listening`. *Configurazione: passo Conduci → Scegli gli indirizzi…*

## Dove si decide un'impostazione

Dal più debole al più forte: il valore incorporato nel programma → `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) → il `[settings]` del file dello spettacolo. Un'opzione della riga di comando o una variabile d'ambiente prevale sul file. Le impostazioni del locale (porte, HTTPS, password, `access`, `data_dir`, `audience.max_clients`, `mdns`, `network`, `standby`, `log`) esistono solo qui.

## Un esempio completo

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
  include: []
  exclude: []
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
