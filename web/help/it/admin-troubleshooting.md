# Risoluzione dei problemi

## Il pulsante Risoluzione problemi

Il pulsante **Risoluzione problemi**, in alto nella configurazione, accanto a Impostazioni, è il punto da cui partire quando le persone non riescono a collegarsi o qualcosa non va. Elenca ciò che il programma sa: se questo computer ha un indirizzo di rete da dare, quali porte (HTTP e HTTPS) ha ottenuto davvero il server web, se HTTPS e il suo certificato sono a posto, se il nome `.local` viene annunciato, se il file delle impostazioni può essere salvato, se ogni lingua corrisponde al file principale dello spettacolo se un elenco di accesso respinge dei computer (o lascia aperta la pagina Segui), se lo spettacolo è in attesa o in onda (ciò che il pubblico vede adesso) e quanti dispositivi sono collegati. Ogni controllo è segnato **OK**, **Nota**, **Da controllare** o **Problema**, e quelli che si possono correggere hanno un pulsante che porta all'impostazione giusta. Sotto ci sono gli ultimi avvisi ed errori del registro, un collegamento al registro completo e il pacchetto di assistenza. **Prova la connessione** interroga il server web a ogni indirizzo fornito dal programma e indica quali rispondono e con che rapidità; un firewall su questo computer o sulla rete può comunque bloccare altri dispositivi.

Quando il programma registra nuovi avvisi o errori, sul pulsante compare un numero finché non lo apri.

## Nessuno riesce ad aprire l'indirizzo

- Controlla che il telefono sia sul WiFi ospiti del locale, non sui dati mobili.
- Controlla che l'indirizzo coincida con quello del passo **Condividi**. Se ha un numero di porta, includilo.
- Chiedi al locale se i dispositivi del WiFi ospiti possono raggiungere questo computer.
- Il firewall del computer potrebbe bloccare il programma. Su Windows, consentilo sulle **reti private** quando richiesto. Su macOS, consenti le connessioni in entrata per esso.
- Apri l'indirizzo su questo computer usando `localhost` per verificare il programma stesso.

## Il numero di dispositivi connessi è inferiore al previsto

I dispositivi che si sono addormentati si disconnettono e si riconnettono da soli al risveglio. Chiedi alle persone di disattivare il blocco automatico durante lo spettacolo se il loro schermo si oscura.

## Le immagini appaiono lentamente

Le immagini grandi richiedono tempo su un WiFi affollato. Cerca nel passo Immagini i file segnati come «grande» e sostituiscili con altri più piccoli.

Per l'elenco completo dei controlli lato rete, vedi la [guida di rete del locale](../../../docs/i18n/it/venue-network-guide.md).

## Manca una lingua per il pubblico

Apri il passo File dello spettacolo e guarda la scheda di quella lingua: un ⚠ significa che il suo file dello spettacolo non ha più le stesse battute di quello principale. Correggi l'errore mostrato; la lingua torna non appena salvi. Vedi [Lingue](admin-languages.md).

## Problemi del file dello spettacolo

Vedi «Cosa significano gli avvisi» in [Scrivere un file dello spettacolo](admin-writing-a-show-file.md).

## Password

Per cambiare una delle due password, usa **Impostazioni** in cima alla pagina di Configurazione. La nuova password ha effetto subito e viene salvata in `config.yaml`. Chiunque sia connesso con una password che è cambiata viene disconnesso; tu resti connesso. Le password sono conservate come hash, quindi la configurazione non può mostrarne nessuna: annota la password quando la imposti. Una **password spettatore** facoltativa permette a un secondo dispositivo, come il tablet di un direttore di scena, di aprire la pagina di visione (`/follow`) solo per guardare: mostra le righe, le note e la posizione, e da lì non si può inviare nulla. Anche le password di operatore e di amministratore aprono quella pagina.

Se non riesci proprio ad accedere, entrambe le password sono in `config.yaml`, nella cartella dati del programma. Apri il file con un qualsiasi editor di testo. La sua posizione viene stampata all'avvio del programma.

## Guardare il registro

**Risoluzione problemi → Apri il registro completo…** mostra cosa il programma ha registrato di recente. Scegli *Solo problemi* per vedere avvisi ed errori, come accessi rifiutati o un nome già usato sulla rete. La stessa finestra imposta quanto dettaglio viene registrato (**Livello di dettaglio del registro**); *Debug* aggiunge ogni richiesta, il che aiuta quando succede qualcosa di strano. Le password non vengono mai scritte nel registro.

## Chiedere aiuto

Scarica il **pacchetto di supporto** con il pulsante **Risoluzione problemi** in alto nella configurazione. Contiene le impostazioni (senza password), le righe recenti del registro e informazioni di base su questo computer, ed è facile da inviare a chi mantiene il programma.
