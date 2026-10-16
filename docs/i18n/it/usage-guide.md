# Guida all'uso di open-caption

open-caption mostra sottotitoli in diretta sui telefoni, tablet e schermi del pubblico durante una commedia o un musical. Un **operatore** volontario segue lo spettacolo e preme la barra spaziatrice a ogni riga; ogni dispositivo sul WiFi del locale mostra la riga un istante dopo. Sui dispositivi del pubblico non si installa nulla, e nessuno deve toccare niente.

Questa guida spiega cosa comprare e come disporlo ([Installazione](#raccomandazioni-di-installazione)), come preparare uno spettacolo, come condurlo, e contiene una lista di controllo per il giorno dello spettacolo. Il pulsante **Aiuto** del programma copre gli stessi argomenti in pezzi più piccoli, e [venue-network-guide.md](venue-network-guide.md) è scritta per chi gestisce il WiFi dell'edificio.

Indice: [Le componenti](#le-componenti) · [Raccomandazioni di installazione](#raccomandazioni-di-installazione) · [Installazione e avvio](#installazione-e-avvio) · [Preparare uno spettacolo](#preparare-uno-spettacolo) · [Condurre lo spettacolo](#condurre-lo-spettacolo) · [Lista di controllo del giorno dello spettacolo](#lista-di-controllo-del-giorno-dello-spettacolo) · [Risoluzione dei problemi](#risoluzione-dei-problemi) · [Manutenzione](#manutenzione) · [Scheda rapida dell'operatore](#scheda-rapida-dell-operatore)

## Le componenti

| Componente | Cos'è |
|------------|-------|
| **Il server** | Il programma `open-caption`, in esecuzione su un computer (macOS, Windows o Linux). Conserva il file dello spettacolo e invia le righe a tutti. |
| **Pagina di configurazione** (`/admin`) | Dove i volontari preparano uno spettacolo: file dello spettacolo, immagini, aspetto, prova. Password amministratore. |
| **Pagina dell'operatore** (`/operator`) | Dove l'operatore segue lo spettacolo e invia le righe. Password operatore. |
| **Pagina del pubblico** (`/`) | Ciò che il pubblico apre sui propri dispositivi. Nessuna password. Un browser su un proiettore o un grande monitor può mostrare la stessa pagina per tutta la sala. |

Tutto gira su un solo computer. Il pubblico, l'operatore e i volontari della configurazione usano normali browser web.

## Raccomandazioni di installazione

### La versione breve

1. **Usa un punto di accesso WiFi dedicato ai sottotitoli** a meno che la rete ospiti del locale non sia nota per funzionare (vedi sotto). Elimina il rischio maggiore: una rete che non controlli.
2. **Metti il server su una connessione cablata** verso quel punto di accesso, con un indirizzo riservato.
3. **Usa un punto di accesso fatto per le folle** (adatto a più di 100 dispositivi) collocato nell'area del pubblico.
4. **Tieni il computer server collegato alla corrente, sveglio e per il resto inattivo** durante lo spettacolo.
5. **Stampa il codice QR** e mettilo nel programma di sala e nell'ingresso.
6. **Prova con alcuni telefoni veri** nella sala vera prima della prima.

### Opzioni di rete

| Opzione | Quando usarla | Note |
|---------|---------------|------|
| **A. Punto di accesso dedicato (consigliato)** | Qualsiasi locale, specie se non puoi cambiare la rete dell'edificio | Porti tu un router o un punto di accesso. Il pubblico si collega al suo WiFi (per esempio «Theater Captions»). Non serve connessione a internet. Controlli tu capacità, prenotazioni e firewall. |
| **B. Il WiFi ospiti del locale** | Il personale di rete del locale conferma che i dispositivi ospiti possono raggiungere il tuo computer | Il più economico. Chiedi un indirizzo riservato per il tuo computer e che sia consentito il traffico dagli ospiti al server. Vedi [venue-network-guide.md](venue-network-guide.md). |
| **C. L'hotspot del tuo portatile** | Solo prove e pubblici minuscoli | Gli hotspot dei portatili di solito reggono solo una manciata di dispositivi. Non per una sala piena. |

### L'opzione A in dettaglio (punto di accesso dedicato)

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

- **Attrezzatura:** un punto di accesso o router professionale o semiprofessionale progettato per molti dispositivi simultanei (cerca più di 100 client nelle specifiche e il supporto ai 5 GHz), con una porta Ethernet. I router domestici economici faticano spesso oltre 30–50 dispositivi. Per una sala di circa 150 posti, prevedi un punto di accesso capace, e un secondo per una sala lunga o larga.
- **Impostazioni:**
  - Crea un SSID con un nome breve e una passphrase WPA2 semplice e stampabile, oppure senza passphrase se il locale consente una rete aperta. Stampala sul cartello QR.
  - Attiva i **5 GHz**, e preferiscili. Usa larghezze di canale di 20 o 40 MHz in un edificio affollato.
  - Lascia attivo l'**isolamento dei client** se è offerto. I dispositivi del pubblico parlano solo con il server, mai tra loro.
  - Dai al computer dei sottotitoli una **prenotazione DHCP** (o un indirizzo fisso) perché l'indirizzo non cambi mai, e stampa l'indirizzo sul cartello.
  - Non serve una connessione a internet, e il programma non ne usa mai una.
- **Collocazione:** metti il punto di accesso nell'area del pubblico, in alto e lontano dai muri, non in un retro né nell'ingresso. Non metterlo dietro il palco o il rack delle luci. Un portatile su un tavolo lontano va bene: è collegato via cavo.
- **Alimentazione:** usa una ciabatta con protezione dalle sovratensioni. Un piccolo UPS sul punto di accesso e sul server elimina il rischio di una spina scollegata.

### L'opzione B in dettaglio (il WiFi ospiti del locale)

Al programma non importa su quale rete sia, purché i dispositivi del pubblico possano raggiungerlo. Prima di farci affidamento:

- Chiedi l'**accesso degli ospiti al server**: il tuo computer deve essere raggiungibile dall'SSID ospiti, sulla porta TCP 80 (e 443 se attivi HTTPS). L'«isolamento dei client» tra ospiti va bene; impedire agli ospiti di raggiungere *il tuo computer* no.
- Chiedi un **indirizzo riservato** per la scheda di rete del tuo computer.
- Collega il computer alla rete via **Ethernet** se possibile.
- Prova con un telefono sul WiFi ospiti con largo anticipo (il passo Conduci mostra l'indirizzo).
- Tieni l'opzione A in una borsa come riserva se lo spettacolo è importante.

### Il computer dei sottotitoli

- **Qualsiasi portatile o mini-PC normale** va bene. Il programma usa pochissima memoria (decine di megabyte) e quasi nessun processore. Usa un computer che puoi tenere collegato alla corrente.
- **Sistema operativo:** macOS, Windows o Linux. Usa lo stesso computer per tutti gli spettacoli se puoi, così password, spettacoli e indirizzo restano gli stessi.
- **Rendilo noioso:** imposta il computer perché non vada *mai in stop* quando è alimentato, disattiva i riavvii automatici per gli aggiornamenti intorno alle date degli spettacoli, chiudi gli altri programmi e tieni il coperchio aperto o impostato per non andare in stop.
- **Firewall:** consenti il programma sulle **reti private** (Windows lo chiede la prima volta; su macOS consenti le connessioni in entrata per esso).
- **Indirizzo:** cablato e riservato (vedi sopra). Se l'indirizzo cambia, i codici QR stampati smettono di funzionare.
- **Ora:** l'orologio non conta.

### Chi sta dove

- **L'operatore** ha bisogno di una vista libera sul palco e di un dispositivo affidabile. La disposizione più semplice è il computer dei sottotitoli stesso al banco tecnico. La pagina dell'operatore funziona anche da qualsiasi altro portatile, tablet o telefono sulla stessa rete (`http://<indirizzo del computer>/operator`), così l'operatore può sedersi dove vede il palco. Un telecomando Bluetooth o una tastiera è più comodo di un touch screen.
- **Solo un operatore alla volta controlla lo spettacolo.** Chiunque altro apra la pagina dell'operatore può guardare, e può usare **Subentra** se l'operatore gli passa la mano.
- **I volontari della configurazione** usano la pagina di configurazione da qualsiasi computer della rete prima dello spettacolo, o dal computer dei sottotitoli.
- **Uno schermo condiviso** per tutta la sala: apri l'indirizzo del pubblico su un computer collegato a un proiettore o a un grande monitor, aggiungi `?display=projector` all'indirizzo per un testo grande e nessun ingombro sullo schermo, e metti il browser a schermo intero (`F11`, oppure avvia Chrome con `--kiosk`).

### Far funzionare il server

Per la maggior parte dei teatri, il modo più semplice è il migliore: avvia il programma a mano prima di ogni spettacolo, lascia aperta la sua finestra e premi `Ctrl+C` (o chiudila) dopo. La posizione dello spettacolo viene salvata di continuo, quindi riavviare a metà spettacolo è sicuro (vedi [Risoluzione dei problemi](#risoluzione-dei-problemi)).

Se il computer deve avviare il server da solo dopo un riavvio, la cartella [`deploy/`](../deploy/) contiene file pronti per Linux (systemd), macOS (launchd) e Windows (Utilità di pianificazione). Ognuno comincia con i propri passi di installazione.

### Portare la gente alla pagina

In ordine di affidabilità:

1. **Il codice QR e il cartello** (Configurazione → Conduci → *Stampa un cartello con codice QR per l'ingresso*). Mettilo al muro, nel programma di sala e sul cartoncino dietro le poltrone. È su questo che bisogna contare.
2. **Un indirizzo breve.** Sulla porta 80 non c'è un `:porta` da scrivere. Il programma si annuncia anche come **`open-caption.local`** (cambialo in Configurazione → Impostazioni → *Nome di rete*): dove la rete ospiti lo permette, `http://open-caption.local/` funziona sulla maggior parte dei telefoni senza configurare nulla. Alcune reti ospiti bloccano questo tipo di annuncio, quindi non stampare solo il nome; il cartello QR usa l'indirizzo numerico. Un nome nel DNS del router (come `captions.theater`) funziona ovunque ma richiede che il locale lo aggiunga.
3. **La pagina di benvenuto del punto di accesso** (un «portale ospiti» o «captive portal» che reindirizza al tuo indirizzo dopo il collegamento). Si configura sul router o punto di accesso, non in questo programma. Vedi [venue-network-guide.md](venue-network-guide.md).
4. **Opzione 114 del DHCP e `/captive-api` (avanzato, sperimentale).** L'opzione 114 è un'impostazione DHCP che indirizza i dispositivi a una «API del captive portal». Il programma risponde a quell'API su `/captive-api` (per esempio `https://<indirizzo>/captive-api`) con «questa rete non è captive» e l'indirizzo della pagina. Due avvertenze: **non fa aprire ai telefoni la pagina dei sottotitoli da soli**, e lo standard (RFC 8910/8908) richiede **HTTPS con un certificato di cui i dispositivi si fidano**, quindi un'impostazione con `http://` semplice o il certificato autofirmato del programma viene ignorata. Configurala solo se il locale usa già l'opzione 114 e può fornire un certificato fidato. Non è stata provata su dispositivi reali.

### Porte e indirizzi

- Per impostazione predefinita il server ascolta sulla **porta 80**, così l'indirizzo del pubblico non ha bisogno di un numero di porta, per esempio `http://192.168.1.10/`. Se qualcos'altro usa la porta 80, o il sistema operativo non la consente, il server te lo dice e usa la **porta 8080**; l'indirizzo del pubblico finisce allora con `:8080`. Su Linux la porta 80 richiede un permesso (`sudo setcap 'cap_net_bind_service=+ep' ./open-caption`) a meno che tu non usi il file systemd, che lo concede.
- Puoi scegliere tu le porte con `--http-addr :8081`, o in `config.yaml`.
- **HTTPS è facoltativo.** L'`http://` semplice funziona su ogni dispositivo senza avvisi ed è l'indirizzo consigliato per il pubblico. Un indirizzo HTTPS autofirmato fa mostrare a ogni dispositivo un avviso di sicurezza. Attiva HTTPS solo se puoi installare un certificato di cui i tuoi dispositivi si fidano.

### Sicurezza

- La pagina del pubblico può solo leggere. Non può cambiare lo spettacolo né raggiungere le pagine di configurazione e dell'operatore.
- Configurazione e operatore richiedono ciascuno la propria password (memorizzata in chiaro in `config.yaml`, che viene creato leggibile solo da te). Dai la password operatore all'operatore, non quella amministratore. Le password si possono cambiare da **Impostazioni** nella configurazione.
- Sul computer che esegue il programma, Configurazione e la pagina dell'operatore si aprono **senza password** quando l'indirizzo inizia con `localhost`, `127.0.0.1` o `[::1]` (la persona a quel computer è considerata amministratore). Gli altri computer, e qualsiasi richiesta inoltrata da un proxy, accedono ancora con una password. Imposta `access.local_auto_login: false` in `config.yaml` se altre persone possono accedere a questo computer.
- **Chiunque riesca a raggiungere il server può aprire le pagine di accesso della configurazione e dell'operatore, e con `http://` semplice le password viaggiano non cifrate sul WiFi.** Su un WiFi ospiti condiviso questo significa che uno spettatore curioso potrebbe catturare una password e prendere il controllo dei sottotitoli. Riduci il rischio, in quest'ordine:
  1. **Limita quali computer possono aprirle.** In `config.yaml`:
     ```yaml
     access:
       admin_allow: [localhost]            # Setup only on the captioning computer itself
       operator_allow: [192.168.1.50]      # the operator's computer (or a network such as 192.168.1.0/24)
     ```
     Le voci possono essere `localhost`, `private` (intervalli di indirizzi domestici/da ufficio), un singolo indirizzo, o una rete come `192.168.1.0/24`. Un elenco vuoto (il valore predefinito) significa qualsiasi computer. Un indirizzo in `admin_allow` può usare anche la pagina dell'operatore. Riavvia il programma dopo la modifica. Una voce errata ferma il programma con un messaggio invece di consentire a tutti. Impostazioni nella configurazione mostra cosa è attualmente consentito.
  2. **Metti l'operatore su una rete cablata o separata** a cui i dispositivi del pubblico non possano unirsi, oppure dai una password WPA2 alla rete dei sottotitoli.
  3. **Usa HTTPS** (`https.enabled` in `config.yaml`) se puoi dare ai dispositivi un certificato di cui si fidano; uno autofirmato fa mostrare un avviso a ogni dispositivo, e un indirizzo del pubblico non cifrato è più semplice.
- Un sito web, o un altro servizio sullo stesso computer, non può far cambiare lo spettacolo a un browser alle tue spalle: le richieste che non provengono dalle pagine del programma stesso vengono rifiutate.
- Il programma è pensato per una rete privata. **Non esporlo a internet.**

## Installazione e avvio

1. **Procurati il programma.** Scarica la versione per il tuo computer (oppure esegui `make dist` per costruirle tutte). Mettila in una cartella che conserverai, per esempio `Documenti/open-caption`.
2. **Primo avvio.** Eseguilo: doppio clic su Windows, oppure in un terminale `./open-caption` (macOS/Linux; su macOS il primo avvio può richiedere *clic destro → Apri* perché il programma non è firmato).
3. **Annota le password.** La prima volta, il programma stampa le password **amministratore** e **operatore** e dove si trova `config.yaml`. Restano in quel file.
4. La pagina di configurazione si apre nel tuo browser. Se non succede, apri `http://localhost/admin` (o `http://localhost:8080/admin` se la console dice di aver usato la porta 8080).
5. **Scegli dove stanno i dati (facoltativo).** Per impostazione predefinita, spettacoli e impostazioni sono nella cartella di configurazione del tuo utente (macOS `~/Library/Application Support/open-caption`, Linux `~/.config/open-caption`, Windows `%AppData%\open-caption`). Per tenere tutto in una cartella a tua scelta, avvia con `--data <cartella>`. Usa sempre la stessa cartella.

Opzioni utili: `--data <cartella>`, `--http-addr :8080`, `--https-addr :443`, `--no-browser`, `--log-file <file>`, `--log-level debug|info|warn|error`, `--version`.

Se `config.yaml` ha un errore (un'impostazione scritta male, un numero fuori intervallo, un indirizzo errato), il programma elenca ogni problema con la sua riga o il nome dell'impostazione e si ferma; correggi il file e riavvia. (Un file contrassegnato con un `config_version` più recente di quello che il programma comprende è l'eccezione: le impostazioni sconosciute vengono ignorate con un avviso, così puoi tornare a un programma più vecchio.)

## Preparare uno spettacolo

Apri la pagina di configurazione e segui i sei passi in alto. Il pulsante **Aiuto** spiega ciascuno.

1. **Spettacoli.** Crea uno spettacolo (o lo spettacolo di esempio, per provare).
2. **File dello spettacolo.** Incolla il file dello spettacolo o trascina un file `.txt` o `.md`. Ogni riga diventa un sottotitolo. Correggi ciò che è elencato sotto *Problemi e note*; fai clic su uno per saltare alla sua riga. Salva.
3. **Immagini.** Carica le immagini che il file dello spettacolo usa (PNG, JPG, GIF o WebP, idealmente sotto 1 MB ciascuna).
4. **Aspetto.** Scegli colori, carattere e dimensione del testo. I valori predefiniti (bianco su nero, grande) vanno meglio per la maggior parte delle persone. Sono incorporati due caratteri per la leggibilità: uno adatto alla dislessia (OpenDyslexic) e uno ad alta leggibilità (Atkinson Hyperlegible, progettato per l'ipovisione). Anche i singoli spettatori possono scegliere l'uno o l'altro per il proprio dispositivo dal piccolo controllo **Aa**.
5. **Prova.** Fai una prova privata: premi Spazio come farebbe l'operatore, e vedi accanto lo schermo del pubblico. Nulla raggiunge il pubblico.
6. **Conduci.** Controlla l'elenco di preparazione, stampa il cartello QR e premi **Apri la pagina dell'operatore**.

Aprire la pagina dell'operatore carica lo spettacolo **in attesa**: il pubblico vede la schermata di attesa della sala (un logo e un messaggio, impostati una sola volta in **Impostazioni → Schermata di attesa**) e nulla dello spettacolo, così puoi preparare, provare e controllare la pagina dell'operatore con calma. Quando sei pronto, premi **Spazio** (il pulsante dice **AVVIA**) nella pagina dell'operatore per mettere lo spettacolo in onda. **Togli dall'onda…** lo ritira di nuovo senza perdere il punto.

### Scrivere il file dello spettacolo

Chi prepara il file dello spettacolo non ha bisogno di leggere il resto di questa guida: [Preparare un file dello spettacolo](../web/help/admin-preparing-a-show-file.md) copre il lavoro dal copione al file finito, compresa la modifica nel tuo programma, e [Primi passi](../web/help/admin-getting-started.md) percorre i sei passaggi della Configurazione.

Un file dello spettacolo è testo semplice. La versione breve:

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

- **Una riga = un sottotitolo** = una pressione di Spazio. Decidi tu dove spezzare le battute lunghe.
- I **nomi dei personaggi** in maiuscolo seguiti da due punti vengono mostrati sopra la riga.
- Le **indicazioni di scena** tra `(parentesi)` vengono mostrate in corsivo.
- I **testi delle canzoni** tra note `♪`, in corsivo: inizia una riga con `~`, oppure racchiudi la canzone in `[song]`…`[/song]`. Per due cantanti insieme (un duetto o contrappunto), metti le loro righe tra `[together]` e `[/together]`, oppure uniscile con ` || ` su una riga; entrambe compaiono insieme con una pressione di Spazio. I pulsanti **♪**, **♫** e **🎵 Canzone** dell'editor li inseriscono. Vedi l'argomento di aiuto *Musical*.
- I **cicli** (pre-spettacolo e intervallo) ripetono immagini e testo finché l'operatore non va avanti. Un ciclo proprio all'inizio del file dello spettacolo parte quando l'operatore mette lo spettacolo in onda (AVVIA).
- Le **note** `[note: …]` e i segnalibri `[section: …]` sono solo per l'operatore.
- L'argomento di aiuto **Scrivere un file dello spettacolo** elenca tutto, con esempi, e il *promemoria del file dello spettacolo* dell'editor lo riassume.

### Impostazioni dello spettacolo (titolo, aspetto e altro)

Tutto ciò che riguarda uno spettacolo tranne le immagini è conservato nel suo file dello spettacolo, compresi il titolo e l'aspetto dei sottotitoli, in un blocco `[settings]` vicino all'inizio:

```text
[settings]
title: The Garden Party
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
[/settings]
```

Il passo **Aspetto** della configurazione scrive questo blocco per te, e puoi modificarlo a mano. Tutto ciò che un file dello spettacolo omette usa un valore predefinito: il programma ha valori incorporati, `config.yaml` (la sua sezione `defaults`) può cambiarli per ogni spettacolo su questo computer, e ogni file dello spettacolo sostituisce entrambi. Porte, password, HTTPS e liste di accesso non sono impostazioni dello spettacolo e restano in `config.yaml`. Poiché il file dello spettacolo porta con sé il suo aspetto, copiare o esportare lo spettacolo porta con sé anche l'aspetto. Vedi l'argomento di aiuto *Scrivere un file dello spettacolo* per tutte le impostazioni.

### Tornare indietro

Ogni salvataggio conserva la versione sostituita (le ultime 50). **Versioni precedenti…** sopra l'editor le elenca; **Carica** ne mette una nell'editor, e **Salva** la rende attuale. **Confronta** mostra riga per riga cosa cambierebbe caricandone una, e una versione a cui dai un nome (o **Conserva** per ciò che è salvato adesso) non viene mai dimenticata. Esporta comunque prima dello spettacolo: le versioni precedenti vivono nella cartella dati, non nell'esportazione.

### Due persone che modificano

Se qualcun altro salva un file dello spettacolo mentre tu lo hai aperto, il programma ti avvisa al salvataggio invece di sovrascrivere il suo lavoro, e ti lascia prendere la sua versione o tenere la tua.

### Più di una lingua

Uno spettacolo può offrire sottotitoli in più lingue. In Configurazione → File dello spettacolo, fai clic su **＋ Aggiungi una lingua**, lascia spuntata «Parti da una copia del file principale dello spettacolo» e traduci le parole. Ogni lingua è un proprio file dello spettacolo con le stesse righe, comandi e schermate vuote nello stesso ordine, così una pressione di Spazio mostra lo stesso momento in ogni lingua. L'operatore segue solo la lingua principale. I dispositivi del pubblico ricevono la lingua impostata sul loro telefono (o la principale) e possono cambiare dal piccolo menu 🌐, oppure ricevere un link che finisce con `?lang=es`. Una traduzione che smette di coincidere con il file principale dello spettacolo viene segnalata e trattenuta finché non è corretta. Lo spettacolo di esempio include una traduzione in spagnolo da guardare. Vedi l'argomento di aiuto *Lingue*.

### Cambiare cose durante le prove e gli spettacoli

Puoi modificare e risalvare un file dello spettacolo mentre uno spettacolo è in diretta (per esempio per correggere un refuso). Il programma ti chiede di scrivere il nome dello spettacolo per sicurezza, mantiene il punto dell'operatore per numero di riga, e avvisa se il numero di righe è cambiato.

## Condurre lo spettacolo

### Prima che arrivi il pubblico

1. Avvia il server e apri la **pagina dell'operatore**; accedi. Dovrebbe dire **Hai il controllo**.
2. In Configurazione → Conduci premi **Apri la pagina dell'operatore**. Lo spettacolo viene caricato in attesa: il pubblico vede la schermata di attesa della sala. Quando sei pronto, premi **Spazio** (AVVIA) nella pagina dell'operatore per metterlo in onda. Un ciclo di pre-spettacolo, se il file dello spettacolo ne ha uno, parte allora su ogni schermo.
3. Apri l'indirizzo del pubblico sul tuo telefono e, se ne hai uno, sul proiettore. Controlla che il conteggio degli spettatori sulla pagina dell'operatore salga.
4. Metti il cartello QR dove la gente lo veda.

### Durante lo spettacolo

- **Spazio** invia la riga successiva (anche Freccia destra, Pag giù, Invio e la maggior parte dei telecomandi).
- **Backspace** annulla l'ultima cosa che hai inviato.
- **B** svuota gli schermi del pubblico (nasconde tutto; gli stessi sottotitoli tornano quando premi di nuovo **B**). **C** li cancella (cancella i sottotitoli, così la riga successiva inizia su uno schermo vuoto). **Backspace** annulla entrambe le cose.
- **Maiusc+Q** ferma un ciclo di intervallo o di pre-spettacolo (schermo vuoto). **Spazio** continua.
- **Messaggio…** mette un avviso singolo su ogni schermo («Lo spettacolo riprenderà a breve»). **Indietro** lo toglie.
- Fai clic su una riga e conferma per saltare ovunque. **/** cerca nel file dello spettacolo.
- La riga con la freccia gialla è la *successiva*. Il punto verde segna ciò che è sullo schermo adesso.

### Quando lo spettacolo finisce

Premi `Ctrl+C` nella finestra del programma (o ferma il servizio). La posizione viene salvata. Per rifare lo spettacolo dall'inizio, usa **Riavvia lo spettacolo…** nella pagina dell'operatore, oppure **Riavvia lo spettacolo** in Configurazione → Conduci. Entrambi lo riportano in attesa.

## Lista di controllo del giorno dello spettacolo

**Alcuni giorni prima**
- [ ] Prova con il vero file dello spettacolo e alcuni telefoni veri sulla rete vera nella sala vera.
- [ ] Esporta lo spettacolo (Configurazione → Spettacoli → **Esporta**) e conserva lo `.zip` in un posto sicuro, come una chiavetta USB.
- [ ] Conferma che l'indirizzo del computer dei sottotitoli sia riservato e che il cartello QR lo mostri.

**Al locale, un'ora prima**
- [ ] Alimentazione: computer e punto di accesso collegati; computer impostato per non andare mai in stop.
- [ ] Punto di accesso acceso, al suo posto, con il nome giusto; il computer collegato via cavo.
- [ ] Avvia il programma; annota l'indirizzo stampato (e quello in Conduci).
- [ ] Un telefono sul WiFi del pubblico può aprire l'indirizzo e mostra «I sottotitoli inizieranno a breve».
- [ ] Configurazione → Prova: fai una prova rapida; i tasti dell'operatore funzionano.
- [ ] Browser del proiettore (se usato) aperto su `…/?display=projector`, a schermo intero.
- [ ] Premi **AVVIA** nella pagina dell'operatore. Il ciclo di pre-spettacolo (se c'è) è su ogni schermo.
- [ ] L'operatore seduto dove può vedere il palco; telecomando o tastiera provati.
- [ ] Cartello QR sistemato; un volontario conosce il nome del WiFi e l'indirizzo per chi chiede.

**Dopo lo spettacolo**
- [ ] `Ctrl+C` nella finestra del programma.
- [ ] Esporta di nuovo lo spettacolo se hai cambiato qualcosa.

## Risoluzione dei problemi

| Problema | Cosa provare |
|----------|--------------|
| Nessuno riesce ad aprire l'indirizzo | Il telefono è sul WiFi giusto (non sui dati mobili)? L'indirizzo include `:8080` se il programma ha usato quella porta? Il traffico dagli ospiti al server è consentito? Il firewall del computer consente il programma sulle reti private? Vedi [venue-network-guide.md](venue-network-guide.md). |
| Alcuni dispositivi si connettono e altri no | Il punto di accesso potrebbe aver raggiunto il suo limite di dispositivi. Aggiungi un secondo punto di accesso, oppure sposta il primo nella sala. |
| I sottotitoli arrivano in ritardo o si bloccano su alcuni telefoni | Segnale debole o canale occupato. Quei telefoni si ricollegano da soli. Migliora la collocazione del punto di accesso; usa i 5 GHz. |
| Qualcuno ha perso una riga | Può far scorrere i sottotitoli verso l'alto per rileggerla; le nuove righe aspettano dietro un pulsante **Torna in diretta** finché non lo tocca. |
| Una pagina dice «Questo browser è troppo vecchio» | Il browser del dispositivo è sotto ciò che la pagina richiede: circa iOS 12 per la pagina del pubblico, iOS 14.1 / Chrome 87 per la configurazione e la pagina dell'operatore. Aggiorna il dispositivo, oppure usane un altro. |
| Gli schermi dei telefoni si spengono | I telefoni bloccano lo schermo. La pagina tiene acceso lo schermo con il blocco di attivazione del browser dove può e, dove non può (HTTP semplice), con un minuscolo video invisibile in ciclo. Chi preferisce lasciar dormire il proprio telefono (per la batteria scarica, per esempio) può disattivare *Lo schermo resta acceso* nel pannello «Aa». Alcuni telefoni in modalità risparmio energetico rifiutano entrambi: la pagina mostra allora un suggerimento per disattivare il blocco automatico per la serata. |
| È stata inviata la riga sbagliata | **Backspace** la annulla. |
| La pagina dell'operatore dice che un altro operatore ha il controllo | Usa **Subentra**, se sei tu l'operatore. |
| Il computer si è bloccato o è stato riavviato a metà spettacolo | Riavvia il programma. La pagina dell'operatore si ricollega, resta connessa e riprende il controllo da sola, lo spettacolo è dove l'avevi lasciato, e i dispositivi del pubblico si ricollegano da soli. Un ciclo riparte dal primo elemento. **Backspace** può ancora annullare le ultime cose inviate prima del riavvio. |
| Ho dimenticato una password | Accedi con l'altra e usa **Impostazioni** nella configurazione, oppure apri `config.yaml` nella cartella dati con qualsiasi editor di testo. |
| Altro | Configurazione → Conduci → **Scarica un pacchetto di supporto**; contiene le impostazioni (senza password) e le righe recenti del registro. |

## Manutenzione

- **Copie di sicurezza:** *Esporta* in Configurazione → Spettacoli salva uno spettacolo (file dello spettacolo, immagini, aspetto) come `.zip`. *Importa* lo ripristina, su questo o su un altro computer. Puoi anche copiare l'intera cartella dati.
- **Passare a un altro computer:** installa il programma, importa gli spettacoli esportati, imposta l'indirizzo del nuovo computer e, se vuoi mantenere le vecchie password, copia `config.yaml`.
- **Aggiornare il programma:** fermalo (`Ctrl+C`), sostituisci il file del programma con quello più recente e riavvialo con la stessa cartella `--data`. Non aggiornare subito prima di uno spettacolo.
- **Riutilizzare uno spettacolo:** Configurazione → Spettacoli ha **Rinomina…** e **Duplica**. Duplicare crea una copia del file dello spettacolo, delle immagini e dell'aspetto da adattare alla prossima replica, e lascia intatto l'originale.
- **Fare pulizia:** elimina gli spettacoli che non ti servono più da Configurazione → Spettacoli. I `.zip` esportati sono il tuo archivio.
- **Registri:** Configurazione → Conduci → **Registro recente…** mostra cosa il programma ha registrato di recente, filtrato per livello (scegli *Solo problemi* per vedere avvisi ed errori); le righe recenti del registro sono anche nel pacchetto di supporto; aggiungi `--log-file` per tenere un file. Ogni riga inizia con il suo livello (`INFO`, `WARN`, `ERROR` o `DEBUG`). Il livello predefinito `info` registra accessi, chi ha preso il controllo della pagina dell'operatore, spettacoli attivati ed eventi simili; `warn` tiene solo i problemi; `debug` registra anche ogni richiesta, il che aiuta quando succede qualcosa di strano. Impostalo con `log: {level: debug}` in `config.yaml`, con `--log-level debug`, oppure mentre il programma è in esecuzione da Configurazione → Impostazioni → *Livello di dettaglio del registro*. Se il programma è stato avviato con `--log-level`, quell'opzione prevale a ogni avvio: la configurazione lo dice, una modifica fatta lì dura solo fino al prossimo riavvio, e il registro di avvio avvisa quando l'opzione e il file non concordano. Un file di registro tenuto con `--log-file` ne inizia uno nuovo a 5 MB e ne conserva tre più vecchi (`open-caption.log.1` fino a `.3`), così non riempie mai il disco.
- **Archiviare:** in Configurazione → Spettacoli, **Archivia** mette da parte uno spettacolo concluso (viene conservato, elencato sotto «Spettacoli archiviati», e non può andare in onda) e **Ripristina** lo riporta indietro. **Elimina** rimuove uno spettacolo definitivamente.

## Scheda rapida dell'operatore

Stampala e tienila accanto all'operatore.

| Tasto | Fa |
|-------|----|
| **Spazio** | Invia la riga successiva |
| **Backspace** | Annulla l'ultima cosa inviata |
| **Maiusc+Q** | Ferma un ciclo (schermo vuoto) |
| **B** | Schermo vuoto: nasconde tutto, i sottotitoli tornano quando riattivi |
| **C** | Cancella: cancella i sottotitoli per ripartire da zero |
| **/** | Cerca nel file dello spettacolo |
| **?** | Aiuto (i tasti sono in pausa mentre è aperto) |

- Freccia gialla = riga successiva. Punto verde = sullo schermo adesso.
- Riga sbagliata? **Backspace.** Hai perso il segno? Fai clic sulla riga che vuoi e conferma.
- Devi dire qualcosa al pubblico? **Messaggio…**
- Spettacolo interrotto? **B** per svuotare gli schermi, **Messaggio…** per spiegare.
