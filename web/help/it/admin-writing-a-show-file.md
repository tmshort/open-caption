# Scrivere un file dello spettacolo

Un file dello spettacolo è un semplice file di testo (`.txt` o `.md`) scritto in una forma semplice di Markdown. Se incolli un normale copione teatrale, funziona già: **ogni riga è un sottotitolo**, e una pressione della barra spaziatrice invia una riga. Le righe vuote vengono ignorate, quindi usale per tenere leggibile il file.

Le righe lunghe vanno a capo da sole sullo schermo. Decidi tu dove spezzare una battuta lunga in righe, secondo lo stile della rappresentazione. Il programma non divide né unisce mai le righe al posto tuo.

## Formattazione

| Scrivi | Ottieni |
|--------|---------|
| `**forte**` | testo in **grassetto** |
| `*a parte*` o `_a parte_` | testo in *corsivo* |
| `***very loud***` | testo in grassetto e corsivo |
| `~~old line~~` | testo barrato (una correzione, o una battuta cancellata) |
| `(thunder rumbles)` | un'intera riga tra parentesi viene mostrata come suono o indicazione di scena |
| `HAMLET: To be, or not to be` | un'etichetta del personaggio sopra la riga (il nome deve essere in MAIUSCOLO seguito da due punti; va bene anche un gruppo: `DEE DEE & BARRY:`, `ALL (EXCEPT BARRY):`) |

**Barrare il testo.** Metti due tilde per lato, `~~così~~`, e le parole vengono mostrate barrate. Subito dentro ogni coppia ci deve essere un carattere che non sia uno spazio (`~~ x ~~` resta com'è). Una riga cantata inizia con una sola `~` seguita da uno spazio, quindi le due cose non si confondono mai; per iniziare una riga normale con una tilde, scrivi `\~`.

Tutto il resto viene mostrato com'è scritto. Link, elenchi, tabelle e titoli non sono supportati. Neanche l'HTML puro è supportato: se scrivi `<b>`, appare letteralmente, e la pagina mostra una nota che suggerisce la forma Markdown. Sottolineatura, apici e pedici non hanno un marcatore, ma qualsiasi carattere che puoi digitare funziona così com'è: scrivi direttamente H₂O, x², 1ˢᵗ o ½.

```showfile
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Tornare a una versione precedente

A ogni salvataggio viene conservata la versione sostituita (le ultime 50 per ogni lingua). Fai clic su **Versioni precedenti…** sopra l'editor, scegli un'ora e premi **Carica**. Questo mette solo il vecchio testo nell'editor; nulla cambia finché non premi **Salva**, quindi puoi prima guardare. **Confronta** mostra, riga per riga, cosa cambierebbe caricando una versione. Una versione a cui dai un nome (nella casella accanto, oppure con **Conserva** per dare un nome a ciò che è salvato adesso) non viene mai dimenticata, per quanti salvataggi seguano. L'esportazione non include le versioni precedenti.

## Impostazioni dello spettacolo: il titolo e l'aspetto

Tutto ciò che riguarda uno spettacolo tranne le immagini è conservato nel file dello spettacolo, compresi il suo **titolo** e l'**aspetto** dei sottotitoli. Va in un blocco `[settings]`, di solito in cima, con un `nome: valore` per riga:

```showfile
[settings]
title: The Garden Party
background: #000000
text color: #ffffff
speaker name color: #ffffff
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
loop seconds: 10
scrollback lines: 200
[/settings]
```

| Impostazione | Cosa fa |
|--------------|---------|
| `title` | Il nome dello spettacolo, come elencato nella configurazione |
| `background` | Il colore dello schermo dietro i sottotitoli, come `#000000` (funziona anche un nome come `black` o `yellow`) |
| `text color` | Il colore dei sottotitoli, come `#ffffff` |
| `speaker name color` | Il colore dei nomi dei personaggi (quando non hanno un colore ciascuno) |
| `font` | `system`, `serif`, `mono`, `dyslexic` o `hyperlegible` |
| `text size` | Una percentuale da 50 a 300; 100 è normale |
| `speaker names` | `label` (un colore), `color` (un colore diverso per ogni personaggio) o `hidden` |
| `idle message` | Le parole mostrate prima che venga inviata la prima riga |
| `loop seconds` | Quanto resta ogni elemento di un `[loop]` quando non lo indica |
| `scrollback lines` | Quanti sottotitoli uno spettatore può scorrere all'indietro (da 20 a 2000) |
| `archived` | `yes` se lo spettacolo è archiviato; i pulsanti **Archivia** e **Ripristina** della configurazione lo aggiungono e lo tolgono |

**Non devi scrivere nulla di tutto questo.** Il passo **Aspetto** della configurazione compila il blocco per te (e **Rinomina** imposta il titolo), e tutto ciò che ometti usa semplicemente un valore predefinito. I valori predefiniti sono su due livelli: il programma ha valori incorporati, che il file delle impostazioni `config.yaml` può cambiare per tutti gli spettacoli (la sua sezione `defaults`), e poi ogni file dello spettacolo sostituisce ciò che imposta. Vince il più specifico: **file dello spettacolo, poi config.yaml, poi valori incorporati**.

Alcune regole: il file dello spettacolo di una traduzione può impostare solo `idle message` (l'aspetto si imposta una volta, per tutto lo spettacolo, nel file principale dello spettacolo); un'impostazione scritta male o con un valore senza senso viene segnalata con il numero di riga e ignorata; e i commenti dentro il blocco non vengono conservati quando la configurazione lo riscrive, quindi metti i commenti sopra. Più blocchi `[settings]` vengono combinati, e i valori successivi vincono.

## Canzoni e musica

Per un musical, mostra i testi tra note musicali (♪). Ci sono tre modi, dal più rapido al più flessibile:

- **Scrivi o incolla tu le note.** ♪ e ♫ funzionano ovunque nel file dello spettacolo. Nell'editor, i pulsanti **♪** e **♫** le inseriscono al cursore.
- **Inizia una riga con `~`.** La riga è trattata come un testo cantato: viene mostrata in corsivo con un ♪ aggiunto a ogni estremità, e la `~` stessa non viene mostrata. Serve solo la `~` iniziale: il ♪ finale viene aggiunto per te, e una `~` di chiusura è ammessa ma non cambia nulla. Funziona anche dopo il nome di un personaggio.
- **Racchiudi un'intera canzone in `[song]` e `[/song]`.** Ogni riga all'interno è cantata, quindi non serve una `~` su ognuna. Il titolo dopo i due punti è un segnalibro solo per l'operatore. Note, clear e blank possono comunque stare all'interno.

```showfile
OLIVER: ~ Where is love? ~
~ Row, row, row your boat

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[note: hold the last chord]
[/song]
```

### Due o più voci insieme (duetti e contrappunto)

Quando i cantanti cantano righe diverse **nello stesso momento**, mettili in `[together]` … `[/together]`. Ogni riga all'interno viene mostrata in una volta, affiancata su uno schermo largo e impilata su un telefono, e l'operatore le invia con una sola pressione di Spazio. Ognuna mantiene il proprio nome di personaggio. Per un breve scambio puoi scrivere le voci su una riga, separate da ` || ` (uno spazio, due barre verticali, uno spazio). `[duet]` e `[counterpoint]` significano come `[together]`.

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

Le voci che entrano **una dopo l'altra** sono righe separate: premi Spazio per ognuna. Usa `[together]` solo per righe che compaiono nello stesso momento. Un blocco `[together]` contiene solo righe di testo, ne richiede almeno due, e dentro un `[song]` ogni voce è cantata.

Una tilde in mezzo a una riga, come `about ~5 minutes`, è solo testo. Per iniziare una riga normale con una tilde, scrivi `\~`. Vedi [Musical](topic:admin-musicals) per altri consigli.

## Più righe in un sottotitolo

Termina una riga con una barra rovesciata per unirla alla successiva, oppure metti le righe tra `[[` e `]]`.

```showfile
GHOST: Remember me. \
Remember me.

[[
Row, row, row your boat,
Gently down the stream
]]
```

## Commenti e note

- `<!-- così -->` è un **commento**. Nessuno lo vede, e può occupare più righe.
- `[note: testo]` è un promemoria mostrato **solo all'operatore**, per esempio `[note: hold for applause]`.
- `[section: Act 2, Scene 1]` aggiunge un segnalibro nell'elenco dell'operatore. Note e sezioni non consumano mai una pressione della barra spaziatrice.
- Le righe che iniziano con `#` sono riservate ai titoli in una versione futura e vengono ignorate. Usa invece un commento.

```showfile
<!-- Act 1 captions, version 3 -->
[section: Act 1, Scene 1]
[note: lights come up slowly, wait for the first line]
HAMLET: Who's there?
```

## Comandi

Una riga tra parentesi quadre è un comando. I nomi non distinguono maiuscole e minuscole.

| Comando | Cosa fa |
|---------|---------|
| `[clear]` | Cancella i sottotitoli dallo schermo |
| `[blank]` | Svuota lo schermo (i sottotitoli vengono conservati) |
| `[image: file.png]` | Mostra un'immagine a schermo intero. Aggiungi una didascalia con `[image: file.png \| text]` |
| `![caption](file.png)` | Lo stesso, scritto alla maniera di Markdown |
| `[fullscreen: text]` | Mostra un testo grande a schermo intero. Per più righe usa `[fullscreen]` … `[/fullscreen]` |
| `[loop]` … `[/loop]` | Un insieme ripetuto di elementi a schermo intero per il pre-spettacolo o l'intervallo. Vedi [Cicli](topic:admin-loops) |
| `[together]` … `[/together]` | Due o più voci mostrate nello stesso momento. Vedi [Canzoni e musica](topic:admin-musicals) |
| `[song: title]` … `[/song]` | Una canzone: ogni riga all'interno è mostrata come testo cantato. Vedi [Canzoni e musica](topic:admin-musicals) |
| `[section: name]` | Un segnalibro per l'operatore |
| `[note: text]` | Una nota per l'operatore |
| `[language: es]` | Segna un file dello spettacolo tradotto. Vedi [Lingue](topic:admin-languages) |
| `[settings]` … `[/settings]` | Il titolo e l'aspetto dello spettacolo. Vedi *Impostazioni dello spettacolo* sopra |

```showfile
[clear]
(Lights up on the garden.)
[image: garden.png | The garden at dusk]
[fullscreen: **Please remain seated**]
[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
[blank]
```

Un'immagine, un testo a schermo intero o uno schermo vuoto resta sugli schermi del pubblico finché l'operatore non invia la riga successiva.

## Caratteri speciali

Metti una barra rovesciata prima di un carattere per mostrarlo letteralmente invece di trattarlo come formattazione: `\*`, `\_`, `\#`, `\[`, `\]`, `\:`, `\(`, `\)`, `\!`, `\~` e `\\` (una barra rovesciata stessa). Per esempio, un sottotitolo che deve iniziare con un cancelletto si scrive `\# 1 hit`, e `HAMLET\: not a speaker` mostra i due punti senza creare un'etichetta di personaggio.

```showfile
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## Cosa significano gli avvisi

Il passo File dello spettacolo elenca i problemi con i numeri di riga. Fai clic su uno per saltare alla riga.

- **«Comando sconosciuto [x]»** — una riga tra parentesi quadre che non è un comando. Se doveva essere un sottotitolo come un effetto sonoro, scrivila con le parentesi tonde, come `(laughter)`.
- **«Immagine … non trovata»** — l'immagine non è stata caricata, o il nome è scritto in modo diverso. Caricala nel passo Immagini.
- **«[loop] è stato aperto ma mai chiuso»** — aggiungi `[/loop]` dopo l'ultimo elemento. Lo stesso vale per i blocchi `[fullscreen]`, i blocchi `[[` e i commenti.
- **«L'HTML non è supportato»** — usa `**grassetto**` e `*corsivo*` invece dei tag.
- **«non può stare dentro un ciclo»** — i cicli contengono solo immagini e testo.
- **«[together] richiede almeno due righe»** — un blocco `[together]` con una riga viene mostrato come un normale sottotitolo. **«[together] è stato aperto ma mai chiuso»** — aggiungi `[/together]`; le righe che hai scritto vengono conservate.
- **«Questo sottotitolo mostra 5 voci contemporaneamente»** — cinque o più voci in un blocco `[together]` (o in una riga con ` || `) sono molte per un telefono, che le impila, e per uno schermo condiviso, che ha spazio per circa quattro e non scorre mai. Dividilo in due sottotitoli.
- **«Impostazione sconosciuta» / «non è un carattere» / «non è un colore»** — una riga del blocco `[settings]` che la configurazione non può usare; viene ignorata e si applica il valore predefinito. Il messaggio dice cosa scrivere invece.
- **«[settings] è stato aperto ma mai chiuso»** — aggiungi `[/settings]`; tutto ciò che segue un blocco non chiuso viene letto come impostazioni.
- **«[song] è stato aperto ma mai chiuso»** — aggiungi `[/song]` dopo l'ultimo testo. I testi che hai scritto vengono comunque conservati.

Gli errori non ti impediscono mai di salvare.

## Se due persone modificano lo stesso file dello spettacolo

Se qualcun altro salva il file dello spettacolo mentre tu lo hai aperto, il programma non sovrascrive in silenzio il suo lavoro. Quando salvi, ti dice che il file dello spettacolo è stato modificato da qualcun altro e ti lascia **usare la sua versione** (sostituendo ciò che c'è nel tuo editor) oppure **salvare comunque la tua**. Copia ciò che vuoi tenere prima di scegliere. La parte con l'errore viene saltata, e il resto del file dello spettacolo continua a funzionare.
