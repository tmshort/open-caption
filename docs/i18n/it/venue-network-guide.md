# Guida di rete del locale

Per chi si occupa del WiFi del teatro. Il computer dei sottotitoli non ha bisogno di nulla da internet. Deve solo essere **raggiungibile dai dispositivi del pubblico**.

## Cosa fa il pubblico

1. Si collega al WiFi ospiti.
2. Apre un indirizzo, per esempio `http://192.168.1.10/` (oppure inquadra il codice QR).

## Requisiti

| Requisito | Perché |
|-----------|--------|
| La rete ospiti può raggiungere il computer dei sottotitoli | I dispositivi scaricano la pagina e tengono una connessione aperta con esso |
| Il computer ha un indirizzo fisso (prenotazione DHCP o statico) | Perché i codici QR e i cartelli stampati restino validi da uno spettacolo all'altro |
| Il punto di accesso regge più di 100 dispositivi contemporanei (con 5 GHz attivo) | Una sala piena sono circa 150 dispositivi |
| Non serve alcun collegamento a internet | Tutto è servito dal computer dei sottotitoli |

**L'isolamento dei client tra gli ospiti va bene ed è consigliato.** I dispositivi del pubblico parlano solo con il computer dei sottotitoli, mai tra loro. Ciò che *non* deve essere bloccato è il traffico dagli ospiti al server. Se la rete ospiti è isolata dalla LAN principale, collega il computer dei sottotitoli alla rete ospiti (una porta libera della VLAN ospiti o l'SSID ospiti) oppure aggiungi una regola di firewall che permetta agli ospiti di raggiungere l'indirizzo del computer sulla porta TCP 80 (e 443 se HTTPS è attivo).

## Controllare prima dello spettacolo

Da un telefono sul WiFi ospiti, apri l'indirizzo mostrato in Configurazione → **Condividi**. Se la pagina si carica e dice «I sottotitoli inizieranno a breve» (o mostra il pre-spettacolo), hai finito. Altrimenti vedi «Non si carica nulla» più sotto.

## Rendere l'indirizzo più facile

In ordine di impegno:

1. **Codice QR** — Configurazione → Condividi → «Stampa un cartello con codice QR». Mettilo nell'ingresso e nel programma di sala. Non servono modifiche alla rete.
2. **Porta 80** — se il programma gira sulla porta 80, l'indirizzo non ha `:porta`. Su Linux, consentilo con `sudo setcap 'cap_net_bind_service=+ep' ./open-caption`.
3. **Un nome locale** — aggiungi sul router una voce DNS per il computer dei sottotitoli, per esempio `captions.theater` → `192.168.1.10`. La maggior parte dei router (OpenWrt, Pi-hole, UniFi, Omada, basati su dnsmasq) permette un «record DNS locale» o una «sostituzione host».
4. **Mandare gli ospiti lì automaticamente.** Opzioni:
   - **Pagina di benvenuto / captive portal** sul punto di accesso che punti all'indirizzo del pubblico. La funzione «portale ospiti» incorporata nel punto di accesso funziona con un URL esterno (UniFi: Hotspot → portale esterno; OpenWrt: nodogsplash `RedirectURL`). Assicurati che lasci caricare la pagina senza login.
   - **Opzione 114 del DHCP (RFC 8910) e `/captive-api`.** Il programma risponde a `GET /captive-api` come descrive la RFC 8908: `{"captive": false, "user-portal-url": "http://<server-address>/"}`. Se la tua rete distribuisce già l'opzione 114 («Captive-Portal API URI»), puoi farla puntare a quell'indirizzo. Tieni presente cosa fa: dice ai dispositivi che la rete *non* è un captive portal e pubblica l'indirizzo della pagina. **Non** fa aprire ai telefoni la pagina dei sottotitoli da soli (servirebbe `"captive": true`, che il programma non invia mai). La RFC richiede inoltre che l'indirizzo dell'API sia **https con un certificato di cui i dispositivi si fidano**, quindi un indirizzo dell'opzione 114 in semplice `http://` o il certificato autofirmato del programma verranno ignorati. È sperimentale e non provato su dispositivi reali: usa il cartello QR, o la pagina di benvenuto del tuo punto di accesso, per portare le persone alla pagina.
5. **mDNS** (`open-caption.local`) è fornito e attivo per impostazione predefinita: i dispositivi che possono usarlo aprono `http://open-caption.local/` senza configurare nulla sul router. Richiede che la rete ospiti permetta il **multicast** (UDP 5353 verso 224.0.0.251) tra il computer dei sottotitoli e i dispositivi del pubblico. Molte reti ospiti lo bloccano («filtro multicast» o «blocco mDNS/Bonjour» sul punto di accesso), e allora il nome semplicemente non si risolve; l'indirizzo numerico e il codice QR funzionano comunque. Se vuoi che il nome funzioni, attiva l'inoltro mDNS/Bonjour o disattiva il filtro multicast per la rete ospiti. Il nome si può cambiare, o disattivare, nella configurazione, in Impostazioni → Rete (*Nome sulla rete*).

Il programma non esegue un proprio server DHCP o DNS, di proposito: andrebbe in conflitto con il router del locale.

## HTTP e HTTPS

Usa `http://` semplice per il pubblico. Un certificato HTTPS autofirmato fa mostrare a ogni dispositivo un avviso di sicurezza. HTTPS è disponibile (`https.enabled: true` in `config.yaml`) per i luoghi che possono installare un certificato di cui i loro dispositivi si fidano. Se la sala possiede un nome di dominio, il programma può ottenere e rinnovare da solo un certificato attendibile; vedi la [guida al certificato HTTPS attendibile](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

Un piccolo costo dell'HTTP semplice: i browser permettono «tieni lo schermo acceso» solo sulle pagine sicure. Con HTTP la pagina tiene acceso lo schermo del telefono con un minuscolo video in ciclo, silenzioso e invisibile. Un telefono in modalità risparmio energetico può rifiutarlo; la pagina mostra allora, una sola volta, un suggerimento che chiede di disattivare il blocco automatico.

## Sicurezza

- La configurazione (`/admin`) e la pagina dell'operatore (`/operator`) richiedono ciascuna una password da `config.yaml`. La pagina del pubblico non ne richiede e non può cambiare nulla.
- Non esporre la porta a internet.
- Le password operatore e amministratore sono memorizzate in chiaro in `config.yaml`, nella cartella dati del programma. Tieni quella cartella privata per le persone che conducono lo spettacolo.

## Firewall sul computer dei sottotitoli

- **Windows:** al primo avvio del programma, consentilo sulle **reti private**. Se la richiesta è stata chiusa: Sicurezza di Windows → Firewall → Consenti un'app.
- **macOS:** consenti le connessioni in entrata per il programma quando richiesto, oppure aggiungilo in Impostazioni di Sistema → Rete → Firewall.
- **Linux:** apri la porta, per esempio `sudo ufw allow 80/tcp`.

## Non si carica nulla

1. Il telefono è sul WiFi ospiti (non sui dati mobili)?
2. L'indirizzo include la porta se il programma non è sulla porta 80? Il passo Condividi mostra l'indirizzo corretto.
3. Un altro computer sul WiFi ospiti riesce a raggiungerlo (`curl http://192.168.1.10/healthz` dovrebbe stampare `ok`)?
4. Il traffico dagli ospiti al server è bloccato dal router o dal punto di accesso (isolamento dei client impostato su «blocca tutto» invece di «blocca tra client»)?
5. Il firewall del computer sta bloccando il programma?
6. Il computer è ancora sveglio, e allo stesso indirizzo? Cerca un cambio di indirizzo in Condividi.

Configurazione → Risoluzione problemi → **Scarica un pacchetto di supporto** raccoglie le impostazioni (senza password), le righe recenti del registro e informazioni di base sul computer.

## Capacità

Il programma è stato provato sotto carico con 200 e 500 connessioni di pubblico simultanee su un solo computer (latenza sotto i 10 ms sulla stessa macchina). In un locale reale il limite è il WiFi, non il programma: usa un punto di accesso adatto alla folla, attiva i 5 GHz e tieni le immagini piccole (sotto 1 MB) perché molti dispositivi possano scaricarle insieme.
