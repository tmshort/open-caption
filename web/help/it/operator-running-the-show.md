# Condurre lo spettacolo

La pagina mostra il file dello spettacolo come un elenco. Le righe già inviate sono attenuate, la riga **sullo schermo adesso** ha un punto verde, e la riga **successiva** è grande con una freccia gialla.

## Le tre cose da ricordare

1. **Spazio** invia la riga successiva.
2. **Backspace** annulla l'ultima cosa che hai inviato.
3. **Maiusc+Q** ferma un ciclo di pre-spettacolo o di intervallo e svuota lo schermo.

Segui gli attori e premi **Spazio** a ogni battuta. Puoi usare un telecomando per presentazioni (clicker): i suoi pulsanti avanti e indietro funzionano come Spazio e Backspace.

## Oscura e Cancella

Entrambi lasciano vuoti gli schermi del pubblico, ma non sono la stessa cosa:

| | **Oscura** (`B`) | **Cancella** (`C`) |
|---|---|---|
| Cosa vede il pubblico | Niente: solo lo sfondo | Un'area dei sottotitoli vuota, pronta per nuove righe |
| Le righe precedenti | **Conservate.** Premi di nuovo `B`, o invia la riga successiva, e sono ancora lì sopra | **Cancellate.** La riga successiva inizia su uno schermo vuoto |
| Usalo per | Un buio, una pausa, o nascondere i sottotitoli per un momento | Una nuova scena, o dopo una canzone, perché le vecchie righe non si accumulino |
| Annulla | `B` di nuovo, o **Backspace** | **Backspace** |

Un file dello spettacolo può fare lo stesso con i comandi `[blank]` e `[clear]`; nel tuo elenco appaiono come `OSCURA SCHERMO` e `CANCELLA SCHERMO`.

## Le parole di questa pagina

- **In attesa**: lo spettacolo è caricato, ma il pubblico vede solo la schermata di attesa della sala.
- **In onda**: il pubblico vede lo spettacolo. **AVVIA** lo manda in onda; **Metti in attesa…** lo riporta in attesa mantenendo la posizione.
- **Riavvia lo spettacolo…**: torna alla prima riga e all'attesa.
- **Oscura** nasconde i sottotitoli ma mantiene le righe precedenti; **Cancella** le elimina (vedi sopra).

## Prima dello spettacolo

- Usa la pagina dell'operatore su un computer con tastiera (va bene un portatile). È pensata per il tasto **Spazio**, e un telefono o un tablet non è consigliato: su un touch screen è facile toccare per sbaglio durante lo spettacolo.
- Accedi alla pagina dell'operatore. Vedrai «Hai il controllo». Se compare «Un altro operatore ha il controllo», usa **Subentra** solo se l'operatore designato sei tu.
- Un secondo dispositivo (il tablet del direttore di scena) può guardare dalla pagina Segui, `/follow`, con la password spettatore (o quella di operatore o amministratore). Mostra ciò che vedi tu, note comprese, ma non può inviare nulla; il passo **Condividi** della configurazione ne indica l'indirizzo.
- Controlla il conteggio degli spettatori (l'icona con le persone). Mostra quanti dispositivi sono connessi.
- Uno spettacolo caricato resta **in attesa**: il pubblico vede la schermata di attesa della sala (un logo e un messaggio) e nulla dello spettacolo, e questa pagina dice **IN ATTESA**. Premi **Spazio** (il pulsante dice **AVVIA**) per mettere lo spettacolo in onda. Se il file dello spettacolo inizia con un ciclo di pre-spettacolo, il ciclo parte in quel momento.

## Durante lo spettacolo

- **Ritmo:** quando una riga ha un `[pace]` nel file dello spettacolo, il suo tempo compare accanto alla riga successiva e una barra sottile, lungo il bordo inferiore della striscia di stato, si riempie finché la riga è sullo schermo. Quando è il momento di premere di nuovo, la riga successiva riceve una cornice gialla e la barra diventa gialla. Nulla viene inviato da solo e non c'è alcun suono.
- Le righe segnate come note o sezioni sono per te. Non richiedono mai di premere Spazio.
- Le immagini e i messaggi a schermo intero compaiono qui solo come etichetta, come `IMMAGINE: sipario.jpg`. Non vedi mai l'immagine vera e propria.
- **Canzoni:** le righe cantate hanno note musicali (♪) intorno e sono mostrate in corsivo sugli schermi del pubblico. Il titolo di una canzone appare come segnalibro blu solo per te. Cerca di premere Spazio all'inizio di ogni frase.
- **Lingue:** se lo spettacolo ha traduzioni, segui la lingua principale e tutto ciò che invii va a tutte le lingue insieme. Il conteggio del pubblico nell'intestazione elenca ogni lingua, per esempio `83 (en 60 · es 23)`.
- **Più voci insieme:** una riga segnata `INSIEME` mostra due o più cantanti sugli schermi nello stesso momento. Basta premere Spazio una volta per inviarle tutte.
- La striscia sotto l'intestazione mostra chi ha il controllo, **Elemento n di totale** e **Ultimo invio … fa**, così vedi a colpo d'occhio che lo spettacolo e il tuo schermo sono ancora allineati. Diventa gialla dopo due minuti senza una riga.
- **Vai alla sezione…** (accanto alla casella di ricerca) elenca i segnalibri del file dello spettacolo; scegline uno e conferma il salto.
- Per andare avanti o indietro, fai clic su una riga e conferma «Vai qui». Puoi anche cercare con `/`.
- Un annuncio occasionale, come «Lo spettacolo riprenderà a breve», si può inviare con **Annuncio…**. Premi **Indietro** per toglierlo. Finché è sugli schermi, un banner blu in questa pagina ne riporta il testo.
- **Metti in attesa…** (in alto a destra) toglie lo spettacolo dagli schermi del pubblico: il pubblico rivede la schermata di attesa e lo spettacolo resta al punto in cui era. **AVVIA** lo rimette in onda esattamente dov'era.

Vedi [la tastiera](operator-keyboard.md) e [Se qualcosa va storto](operator-if-something-goes-wrong.md).

## Dopo lo spettacolo: il rapporto sul pubblico

Quando arrivi alla fine dello spettacolo, il banner dice **Fine dello spettacolo.** con un link **Rapporto sul pubblico** (in seguito lo mostra anche il passo **Rapporto** della configurazione). Il rapporto dice quanti dispositivi diversi hanno usato la pagina del pubblico mentre lo spettacolo era in onda, indipendentemente da quanto tempo sia rimasto collegato ciascuno: un telefono che ha perso il Wi-Fi e si è ricollegato conta una sola volta. Mostra i totali (telefoni, tablet, computer, schermi del proiettore, il numero massimo di dispositivi connessi contemporaneamente, quanti sono ancora connessi), le lingue e una riga per dispositivo. Disegna anche un grafico di quanti dispositivi erano connessi nel tempo, con evidenziati i cicli del pre-spettacolo e dell'intervallo, per vedere quante persone c'erano all'inizio e quante sono rimaste durante la pausa. Continua a contare finché non torni in attesa, riavvii o carichi un altro spettacolo, e si può stampare o scaricare come file CSV per un foglio di calcolo.

Un secondo dispositivo che si limita a guardare lo spettacolo (la pagina Segui) compare nella stessa tabella, contrassegnato come **Spettatore**, ma non viene contato in nessun numero né nel grafico. I dispositivi si distinguono per indirizzo di rete e tipo di browser, quindi i numeri sono una stima: i telefoni che condividono entrambi contano come uno, e un telefono il cui indirizzo cambia conta due volte. Ogni rapporto è salvato come file nella cartella `reports` della cartella dei dati; si conservano i 30 più recenti, e puoi eliminarne uno con **Elimina…** nel passo **Rapporto** della configurazione, o cancellare i file lì.
