# Preparare un file dello spettacolo

Questa pagina è per chi trasforma il copione in sottotitoli. Non serve condurre lo spettacolo né sapere nulla della rete. Un file dello spettacolo è un normale file di testo: puoi scriverlo con il programma che preferisci, su qualsiasi computer, e portarlo nella Configurazione quando è pronto. [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file) elenca tutti i marcatori che puoi usare; questa pagina spiega come affrontare il lavoro.

## In breve

1. Metti il copione in un file di testo semplice. Se vuoi, parti dallo spettacolo di esempio in [sample-show.md](../../sample/sample-show.md).
2. Fai in modo che **una riga del file sia un sottotitolo**: una pressione della barra spaziatrice.
3. Aggiungi ciò che aiuta il pubblico e l'operatore (nomi dei personaggi, descrizioni di suoni, sezioni, note).
4. Porta il file nella Configurazione, leggi l'elenco dei problemi, correggili e salva.
5. Prova nel passaggio Prova e correggi ciò che noti.

## Che tipo di file

- Va bene qualsiasi file di testo semplice: un nome che finisce in `.txt` o `.md`, e qualsiasi nome, come `La nostra città - sottotitoli v3.txt`.
- Salvalo come **UTF-8** se il programma lo chiede. Così lettere accentate, virgolette curve e note musicali sopravvivono. Viene capito anche un file salvato come «ANSI» di Windows (Windows-1252). I file dei programmi di videoscrittura (`.doc`, `.docx`, `.pdf`) non sono file di testo: copia prima il testo in un editor di testo semplice, oppure usa Salva con nome e scegli testo semplice.
- Vanno bene sia le interruzioni di riga di Windows (CRLF) sia quelle di Mac o Linux (LF), e anche il segno invisibile che alcuni editor mettono all'inizio del file.
- Un `.zip` serve solo per una copia di sicurezza completa, che comprende le immagini. Il file dello spettacolo è solo il testo.

## Decidere dove comincia e finisce ogni sottotitolo

Il programma non divide né unisce mai le tue righe. Un sottotitolo è esattamente ciò che hai scritto su una riga, quindi sei tu a controllare quanto legge il pubblico in una volta.

- Punta a righe che si leggano comodamente a colpo d'occhio: una frase o parte di una frase. I discorsi molto lunghi vanno spezzati nelle pause naturali, dove l'attore respira.
- Le righe vuote non significano nulla: usale liberamente tra una battuta e l'altra per tenere il file leggibile.
- Se due o più righe devono comparire insieme, uniscile con una barra rovesciata a fine riga, oppure mettile tra `[[` e `]]`.
- Usa `HAMLET: battuta` per mostrare chi parla. Scrivi il nome in maiuscole, poi i due punti.
- Descrivi suoni e indicazioni di scena tra parentesi, su una riga a parte, per esempio `(rimbomba il tuono)`. Vengono mostrati al pubblico con uno stile diverso.

```showfile
[section: Act 1]
[note: house lights down, then start]
GEORGE: Morning, Emily.
EMILY: Oh, hello, George.
(a train whistle, far away)
```

## Ciò che vede solo l'operatore

Segna il tuo punto e lascia promemoria senza disturbare il pubblico: `[section: Atto 2]` aggiunge un segnalibro nell'elenco dell'operatore, e `[note: aspettare l'applauso]` mostra un promemoria. Nessuno dei due consuma una pressione della barra spaziatrice. Un commento tra `<!--` e `-->` è per te e non lo vede nessuno.

## Lavorare nel tuo editor

Non sei mai obbligato a scrivere nell'editor del passaggio File dello spettacolo. Invece:

1. Nel passaggio **File dello spettacolo**, premi **Scarica il file dello spettacolo**. Ottieni un file `.md` di testo semplice.
2. Aprilo nel tuo programma, modificalo e salvalo. Tienilo come testo semplice.
3. Tornato nel passaggio File dello spettacolo, usa **Carica un file dello spettacolo…** e scegli il tuo. Il suo testo sostituisce quello dell'editor e l'elenco dei problemi si aggiorna subito. Nulla cambia in via definitiva finché non premi **Salva**.
4. Premi **Salva**. La versione sostituita viene conservata in **Versioni precedenti**, dove puoi confrontarla con quella nuova, quindi un errore si può sempre annullare.

Lo stesso vale per una traduzione: scegli prima la scheda della lingua, poi scarica e apri il file di quella lingua. Puoi anche trascinare un file sull'editor.

Per iniziare uno **spettacolo nuovo** da un file, usa **Importa uno spettacolo** nel passaggio Spettacolo e scegli il file. Diventa uno spettacolo nuovo, con il nome della riga `title:` del suo blocco `[settings]`, oppure con il nome del file se non c'è. Le immagini non fanno parte di un file di testo: caricale nel passaggio Immagini, con gli stessi nomi delle tue righe `[image: ...]`.

## Controllare il lavoro

- Il passaggio File dello spettacolo elenca ogni problema con il numero di riga. Fai clic su uno per andarci. Gli errori non ti impediscono mai di salvare, ma un elemento con un errore viene saltato, quindi correggili prima dello spettacolo.
- Puoi controllare un file dello spettacolo senza aprire la configurazione: sul computer dei sottotitoli esegui `open-caption --check-show show.md` (aggiungi dopo le traduzioni, per esempio `show.es.md`). Mostra gli stessi problemi con i numeri di riga e termina con uno stato diverso da zero se ci sono errori, così puoi usarlo mentre modifichi o in uno script. `open-caption --help` elenca le opzioni.
- Il pannello accanto all'editor mostra ciò che l'operatore percorrerà, una riga per ogni pressione della barra spaziatrice. Controlla che il numero di righe sia quello che ti aspetti.
- Esegui il passaggio **Prova** per vedere i sottotitoli come li vedrà il pubblico, su un telefono oltre che su uno schermo grande.
- Se anche altre persone lavorano al file dello spettacolo, vedi *Se due persone modificano lo stesso file* in [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file).

## Prima di consegnare

- Leggi il file una volta dall'inizio alla fine come farebbe il pubblico, non come un correttore di bozze.
- Controlla che ogni nome in `[image: ...]` corrisponda a un'immagine caricata, e che ogni `[loop]`, `[song]`, `[together]` e `[fullscreen]` abbia la riga di chiusura.
- Dai all'operatore una copia del file e i nomi delle sezioni, così sa dove ritrovare il punto. [Canzoni e musica](topic:admin-musicals), [Cicli](topic:admin-loops) e [Lingue](topic:admin-languages) spiegano quelle parti nel dettaglio.
