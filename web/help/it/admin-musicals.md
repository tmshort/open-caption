# Musical

I sottotitoli per un musical funzionano meglio quando i testi delle canzoni si distinguono facilmente dalle battute parlate e l'operatore riesce a seguire la musica. Questa pagina raccoglie ciò che funziona.

## Segnare i testi

I testi vengono mostrati **in corsivo tra note musicali**, per esempio ♪ Dov'è l'amore? ♪. Scegli il modo di scriverli che ti è più comodo:

| Scrivi | Risultato |
|--------|-----------|
| `~ Where is love? ~` | ♪ Where is love? ♪ (la `~` di chiusura è facoltativa) |
| `OLIVER: ~ Where is love?` | lo stesso, con il nome del personaggio sopra |
| `[song: Title]` … `[/song]` | ogni riga all'interno è un testo cantato; il titolo è solo per l'operatore |
| `♪ scritto o incollato ♪` | esattamente ciò che hai scritto, senza corsivo |
| `~` da sola su una riga | solo una nota: ♪, per un passaggio strumentale |

Usa `[song]` per i numeri lunghi, e `~` per una singola riga cantata nel mezzo di una scena.

Nella configurazione, i pulsanti **🎵 Canzone**, **♪** e **♫** sopra l'editor del file dello spettacolo li inseriscono per te. I caratteri ♪ e ♫ si possono anche copiare da qualsiasi parte, perché sono testo normale.

## Due cantanti insieme (contrappunto e duetti)

Quando cantanti diversi cantano parole diverse **nello stesso momento**, mostrali insieme sugli schermi del pubblico con `[together]`:

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]
```

- Entrambe le righe appaiono **insieme** e l'operatore preme Spazio **una volta**.
- Su uno schermo largo le voci stanno affiancate, ciascuna con una barra sul bordo sinistro e il nome del personaggio sopra; su un telefono si impilano.
- Aggiungi tutte le voci che servono, per esempio un trio o una riga parlata sopra una cantata.
- Per una rapida riga a due voci puoi scrivere `ANNA: ~ I love you ~ || KRIS: ~ Me too ~` su una sola riga.
- Le voci che **entrano una dopo l'altra** (un canone, o un secondo cantante che risponde) sono righe separate normali: una pressione di Spazio ciascuna. Scrivi la prima voce, poi la seconda quando entra.
- Tieni ogni voce breve. Tre o quattro parole per voce si leggono più facilmente insieme di un'intera frase.
- Usa i nomi dei personaggi perché il pubblico distingua le voci.

## Consigli per buoni sottotitoli di testi

- **Una riga per frase cantata.** Spezza le righe dove i cantanti respirano, così la pressione di Spazio dell'operatore corrisponde alla musica. Le righe più brevi sono più facili da seguire.
- **Usa i nomi dei personaggi per i duetti** e per le righe cantate a turno da personaggi diversi. Omettili per un coro.
- **Pause strumentali:** usa `~` da sola, oppure `(L'orchestra suona)`, per segnarle. Una riga tra parentesi viene mostrata come indicazione di scena.
- **Tieni separate le righe parlate e cantate.** Una riga di personaggio seguita da testi con `~` si legge chiaramente: il pubblico vede cambiare note e corsivo.
- **Ripetizioni:** scrivi le righe ogni volta. L'operatore preme Spazio una volta per riga in ogni caso.
- **Cancella tra i numeri:** metti `[clear]` dopo l'ultima riga di una canzone perché la scena successiva inizi su uno schermo vuoto.
- **Promemoria per l'operatore:** `[note: il tempo accelera dopo il cambio di tonalità]` è mostrato solo all'operatore.
- **Prova con la musica vera.** Il passo Prova permette all'operatore di esercitarsi a premere Spazio sulla registrazione del cast.

## Un esempio

```showfile
[section: Act 1, Number 4]
[clear]
(The orchestra begins.)
[song: Where Is Love]
OLIVER: Where is love?
Does it fall from skies above?
[note: wait for the held note]
Is it underneath the willow tree
That I've been dreaming of?
[/song]
[clear]
MRS. CORNEY: That's quite enough of that!
```

Vedi anche [Scrivere un file dello spettacolo](admin-writing-a-show-file.md).
