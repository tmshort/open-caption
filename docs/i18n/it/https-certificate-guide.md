# Configurare un certificato HTTPS attendibile (facoltativo)

Questa guida è per la persona che si occupa del computer dei sottotitoli. È **facoltativa**: i sottotitoli funzionano anche senza. (Le impostazioni che usa sono nel [riferimento della configurazione](../../configuration-reference.md).)

## Ti serve?

Probabilmente no. I telefoni del pubblico aprono `http://…` senza alcun avviso, ed è l'indirizzo che usa il codice QR. Un indirizzo sicuro `https://` serve solo per alcune funzioni del browser limitate alle pagine sicure (come tenere acceso lo schermo del telefono nel modo corretto; la pagina del pubblico ha un'alternativa), o se la sala lo desidera.

Senza questa guida, HTTPS funziona comunque con un certificato **autofirmato**, ma ogni telefono mostra un avviso di sicurezza. Con essa, il programma ottiene un vero certificato gratuito da Let's Encrypt e **lo rinnova da solo**, così `https://captions.mytheater.org` si apre senza avvisi su ogni dispositivo.

## Cosa ti serve prima di iniziare

| Ti serve | Perché |
|---|---|
| Un **nome di dominio tuo**, per esempio `mytheater.org`, e un nome al suo interno per i sottotitoli, come `captions.mytheater.org` | Un certificato viene emesso per un nome pubblico. `open-caption.local` e i semplici numeri non possono ottenerlo |
| Il suo **DNS ospitato da un provider supportato da questo programma** | Il programma dimostra che il nome è tuo aggiungendo un record temporaneo nel tuo DNS. Il programma predefinito include **Cloudflare, GoDaddy ed EasyDNS**; esegui `open-caption --version` per vedere cosa ha il tuo. Un altro provider richiede un'altra compilazione (vedi la fine) |
| Le **credenziali API** di quell'account DNS | Permettono al programma di aggiungere e rimuovere quel solo record |
| Che il computer abbia **accesso a Internet** (solo in uscita) | Parla con Let's Encrypt e con il tuo provider DNS. Niente deve raggiungere la sala *dall'esterno* |
| Un **indirizzo fisso** per il computer nella rete della sala (va bene una prenotazione DHCP) | Il nome pubblico punterà a quello |

## Passaggi

**1. Ottieni le credenziali API.** Nell'account del tuo provider DNS:

- **EasyDNS:** attiva l'accesso all'API e crea un **token e una chiave** API. Le impostazioni sono `EASYDNS_TOKEN` e `EASYDNS_KEY`. (EasyDNS ha una sandbox per provare l'API; vedi la loro documentazione per i passaggi attuali.)
- **Cloudflare:** crea un **token API** limitato alla tua zona con il permesso *Zone → DNS → Edit*. L'impostazione è `CLOUDFLARE_DNS_API_TOKEN`.
- **GoDaddy:** crea una **chiave e un segreto** API di produzione. Le impostazioni sono `GODADDY_API_KEY` e `GODADDY_API_SECRET`. (GoDaddy può limitare la sua API DNS ad alcuni tipi di account; controlla le regole attuali.)
- Qualsiasi altro provider: i nomi esatti delle impostazioni sono nella [documentazione di lego](https://go-acme.github.io/lego/dns/).

Un token DNS di solito può modificare **tutti** i record della zona, non solo quello temporaneo, quindi dagli il minor potere possibile che il provider consente. Per la massima protezione usa un CNAME (vedi «Credenziali più sicure» più sotto).

**2. Fai puntare il nome al computer.** Presso il tuo provider DNS, aggiungi un record `A`: `captions.mytheater.org` → l'indirizzo del computer nella rete della sala (per esempio `192.168.1.20`). Va bene che un nome pubblico punti a un indirizzo privato; al certificato non importa.

**3. Tieni le credenziali in un file privato.** Crea `/etc/open-caption/dns.env` (va bene qualsiasi posto privato; su Windows una cartella che solo il tuo account può leggere) con una riga per impostazione, per esempio per EasyDNS:

```
EASYDNS_TOKEN=paste-the-token-here
EASYDNS_KEY=paste-the-key-here
```

Su macOS e Linux rendilo privato: `chmod 600 /etc/open-caption/dns.env`. Il programma avvisa se altri utenti possono leggerlo.

**4. Modifica `config.yaml`** (le prime righe che il programma stampa dicono dov'è; ogni impostazione è spiegata nel [riferimento](../../configuration-reference.md)):

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

**5. Esercitati prima con il servizio di prova.** `letsencrypt-staging` è il servizio di prova di Let's Encrypt: funziona esattamente come quello vero, ma i suoi certificati **non sono considerati attendibili dai browser** (è previsto un avviso) e ha limiti generosi. Avvia il programma.

**6. Verifica che abbia funzionato.** Apri **Configurazione → Impostazioni → Certificato**. Lo stato passa da *In attesa del primo certificato* a **Valido** entro pochi minuti, e mostra il nome, chi l'ha emesso e la data di scadenza. Il registro del programma dice `HTTPS: got a certificate for … valid until …`. Se dice *L'ultimo tentativo non è riuscito*, leggi **Ultimo problema** e la tabella qui sotto.

**7. Passa al certificato vero.** Cambia `directory: letsencrypt-staging` in `directory: letsencrypt` (o cancella la riga) e riavvia il programma. Il certificato vero arriva allo stesso modo. Apri `https://captions.mytheater.org/` da un telefono sul WiFi della sala: nessun avviso. (Se la porta 443 era occupata, il programma ha usato quella di riserva, `:8443`; l'indirizzo richiede allora la porta, `https://captions.mytheater.org:8443/`, e il registro mostra l'indirizzo esatto.)

**8. Dai l'indirizzo alle persone.** Il passo Condividi della configurazione elenca gli indirizzi numerici e il nome `.local`, ma non ancora questo. Per stampare un cartello con codice QR per questo indirizzo, con la sessione aperta nella configurazione apri
`http://localhost/admin?print=qr&url=https%3A%2F%2Fcaptions.mytheater.org%2F` (cambia il nome; aggiungi la porta se serve) e stampa quella pagina.

## Dopo

- **Il rinnovo è automatico.** Il programma controlla circa due volte al giorno e rinnova quando mancano 30 giorni o meno. Non devi fare nulla.
- **Se Internet cade nella sala,** non si rompe nulla: il programma continua a servire il certificato che ha e riprova un'ora dopo. Prima che arrivi il primissimo certificato (o se non arriva mai) i telefoni ricevono quello autofirmato.
- **Rinnova ora** (Configurazione → Impostazioni → Certificato) ottiene subito un nuovo certificato. Usalo dopo aver risolto un problema.
- **Cambiare il token DNS:** modifica `/etc/open-caption/dns.env`. Viene riletto al tentativo successivo, quindi non serve riavviare.
- **Disattivarlo:** imposta `acme: enabled: no` e riavvia. HTTPS usa allora il certificato autofirmato (o i tuoi file).

## Se qualcosa va storto

La configurazione mostra **Ultimo problema**; il registro ha lo stesso testo. I casi più comuni:

| Vedi | Causa probabile e rimedio |
|---|---|
| `the DNS provider "…" is not built into this program` (il programma si ferma all'avvio) | Questa copia del programma è stata compilata senza quel provider. Usa `cloudflare`, `godaddy` o `easydns`, oppure procurati una compilazione con il tuo (più sotto) |
| `… credentials information are missing: EASYDNS_TOKEN,EASYDNS_KEY` | I nomi nel file sono sbagliati o il file non è leggibile. I nomi sono in maiuscolo, `NOME=valore`, uno per riga |
| Un errore del provider DNS come *unauthorized*, *forbidden* o un numero 4xx | Il token o la chiave sono sbagliati, scaduti o non autorizzati a modificare quella zona. Creane di nuovi |
| *In attesa del primo certificato* per molto tempo, o un messaggio che il record non è stato trovato | Il nuovo record DNS è lento a propagarsi. Aggiungi `propagation_wait: 3m` sotto `dns:` e premi **Rinnova ora** |
| `too many certificates already issued` o `rate limit` | Let's Encrypt limita le ripetizioni. Aspetta (il messaggio dice quanto) e usa `letsencrypt-staging` mentre provi |
| Il browser avvisa ancora | Stai usando il servizio di prova (passa a `letsencrypt`), oppure hai aperto l'indirizzo numerico o un altro nome invece di `https://captions…`, oppure l'orologio del computer o del telefono è sbagliato |
| Un telefono non riesce ad aprire il nome | Alcune reti WiFi per ospiti e router bloccano i nomi pubblici che rispondono con indirizzi privati («protezione DNS rebinding»). Chiedi al responsabile della rete di consentire `captions.mytheater.org`, oppure usa l'indirizzo numerico e il codice QR come prima |
| `https.acme: cannot be used together with cert_file/key_file` | Hai impostato entrambi. Togline uno: certificati automatici **oppure** i tuoi file |
| Il programma rifiuta di partire ed elenca problemi in `config.yaml` | Ogni riga nomina l'impostazione e cosa non va. Correggili tutti e riavvia |

## Credenziali più sicure (facoltativo)

Un token DNS che può modificare tutta la tua zona è un rischio se il computer viene perso o condiviso. Due modi per ridurlo:

- Tieni i segreti **fuori da `config.yaml`**: nel file delle credenziali (passo 3), o come variabili d'ambiente del servizio che avvia il programma (funziona anche `EASYDNS_TOKEN_FILE=/path/to/file`: l'impostazione viene letta da quel file).
- Usa la **delega CNAME**: nel tuo DNS aggiungi un record permanente `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.other-zone.example`, dove `other-zone.example` è una piccola zona altrimenti vuota. Il programma lo segue automaticamente, così il token ha bisogno di accedere solo a quella piccola zona, non alla principale.

Le credenziali non vengono mai scritte nel registro, nel pacchetto di supporto né nella configurazione.

## Un altro provider DNS

Il programma contiene solo i provider DNS con cui è stato compilato. Chi lo compila può sceglierne altri, o tutti, con `make build ACME_PROVIDERS="…"` (i nomi sono quelli dell'[elenco di lego](https://go-acme.github.io/lego/dns/)). Compilare richiede il codice sorgente e una connessione a Internet (le guide alla compilazione per macOS, Windows e Linux sono nel repository del codice; su Windows usa `build.bat -AcmeProviders "…"`). Il programma cresce con ogni provider: circa 10,9 MB senza certificati automatici, 12,6 MB con i tre predefiniti (Cloudflare, GoDaddy, EasyDNS), 15,0 MB con EasyDNS e Route 53 e 60,9 MB con tutti i provider. Se il tuo provider non è supportato affatto, `exec` (esegue uno script tuo) e `httpreq` (chiama un indirizzo web) possono pilotare quasi qualsiasi servizio DNS.
