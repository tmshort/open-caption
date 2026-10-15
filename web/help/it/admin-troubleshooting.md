# Risoluzione dei problemi

## Nessuno riesce ad aprire l'indirizzo

- Controlla che il telefono sia sul WiFi ospiti del locale, non sui dati mobili.
- Controlla che l'indirizzo coincida con quello del passo **In diretta**. Se ha un numero di porta, includilo.
- Chiedi al locale se i dispositivi del WiFi ospiti possono raggiungere questo computer.
- Il firewall del computer potrebbe bloccare il programma. Su Windows, consentilo sulle **reti private** quando richiesto. Su macOS, consenti le connessioni in entrata per esso.
- Apri l'indirizzo su questo computer usando `localhost` per verificare il programma stesso.

## Il numero di dispositivi connessi è inferiore al previsto

I dispositivi che si sono addormentati si disconnettono e si riconnettono da soli al risveglio. Chiedi alle persone di disattivare il blocco automatico durante lo spettacolo se il loro schermo si oscura.

## Le immagini appaiono lentamente

Le immagini grandi richiedono tempo su un WiFi affollato. Cerca nel passo Immagini i file segnati come «grande» e sostituiscili con altri più piccoli.

Per l'elenco completo dei controlli lato rete, vedi la [guida di rete del locale](topic:admin-venue-network-guide).

## Manca una lingua per il pubblico

Apri il passo File dello spettacolo e guarda la scheda di quella lingua: un ⚠ significa che il suo file dello spettacolo non ha più le stesse battute di quello principale. Correggi l'errore mostrato; la lingua torna non appena salvi. Vedi [Lingue](topic:admin-languages).

## Problemi del file dello spettacolo

Vedi «Cosa significano gli avvisi» in [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file).

## Password

Per cambiare una delle due password, usa **Impostazioni** in cima alla pagina di Configurazione. La nuova password ha effetto subito e viene salvata in `config.yaml`. Chiunque sia connesso con una password che è cambiata viene disconnesso; tu resti connesso. La stessa finestra può mostrare la password operatore attuale così puoi darla all'operatore.

Se non riesci proprio ad accedere, entrambe le password sono in `config.yaml`, nella cartella dati del programma. Apri il file con un qualsiasi editor di testo. La sua posizione viene stampata all'avvio del programma.

## Guardare il registro

**Impostazioni → Registro recente…** mostra cosa il programma ha registrato di recente. Scegli *Solo problemi* per vedere avvisi ed errori, come accessi rifiutati o un nome già usato sulla rete. La stessa finestra imposta quanto dettaglio viene registrato (**Livello di dettaglio del registro**); *Debug* aggiunge ogni richiesta, il che aiuta quando succede qualcosa di strano. Le password non vengono mai scritte nel registro.

## Chiedere aiuto

Scarica il **pacchetto di supporto** dal fondo del passo In diretta. Contiene le impostazioni (senza password), le righe recenti del registro e informazioni di base su questo computer, ed è facile da inviare a chi mantiene il programma.
