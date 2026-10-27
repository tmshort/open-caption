# Cicli: prima dello spettacolo e intervallo

Un **ciclo** è un insieme di immagini e testi a schermo intero che si ripete finché l'operatore non va avanti. Usane uno prima che lo spettacolo cominci e un altro nell'intervallo.

```showfile
[loop: 10s]
[image: sponsors.png | 15s]
**Welcome!**
Please silence your phones
[/loop]
```

Ogni riga dentro il ciclo è un elemento:

- una semplice riga di testo viene mostrata come testo a schermo intero;
- funzionano anche `[image: file.png]`, `![didascalia](file.png)` e `[fullscreen: testo]`;
- `[blank: 3s]` è una breve pausa su schermo vuoto;
- aggiungi `| 15s` alla fine di un elemento per cambiare quanto resta;
- `[loop: 10s]` imposta il tempo per gli elementi che non lo indicano. Se non dai alcun tempo, si usa il valore predefinito di `config.yaml` (10 secondi).
- un tempo può essere al massimo di 3600 secondi (un'ora); uno più lungo viene ridotto a questo valore, con un avviso.

Sottotitoli, `[clear]`, sezioni, note e altri cicli non possono stare dentro un ciclo.

## Quando un ciclo parte e si ferma

- Se la **prima cosa** nel file dello spettacolo è un ciclo, parte quando l'operatore mette lo spettacolo in onda (**AVVIA**), così il pubblico vede il pre-spettacolo mentre la gente cerca il posto.
- Qualsiasi altro ciclo (per esempio nell'intervallo) parte quando l'operatore preme la barra spaziatrice su di esso.
- La barra spaziatrice termina il ciclo e invia la riga successiva. **Maiusc+Q** ferma il ciclo e lascia lo schermo vuoto. **Indietro** torna al ciclo e lo riavvia dal primo elemento.

L'elenco dell'operatore mostra un ciclo come una sola voce che si espande per elencarne gli elementi. Vedi [I cicli per l'operatore](operator-loops.md).

## Un timer con conto alla rovescia

Dai un **timer** a un ciclo e gli schermi del pubblico possono mostrare quanto tempo manca, contando alla rovescia in diretta («Si riprende tra 15 min»). Scrivi `timer` (o `countdown`) nelle opzioni del ciclo, poi metti `{timer}` in un testo di quel ciclo dove deve comparire il tempo.

```showfile
[loop: timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
![Refreshments are in the lobby](lobby.png)
[/loop]
```

Il valore di un timer è una di due cose:

- Un **numero di minuti**, come `timer 15` o `timer 15 minutes`: per un intervallo. Il conto alla rovescia parte nel momento in cui parte il ciclo (quando l'operatore preme la barra spaziatrice su di esso) e riparte dal tempo intero se l'operatore torna indietro nel ciclo.
- Un **orario**, come `timer 7:30pm`, `timer 7pm` o `timer 19:30`: per il prima dello spettacolo. Gli schermi contano alla rovescia fino a quell'ora di oggi. Si usano l'orologio e il fuso orario del computer che esegue il programma, quindi controllali. Un orario già passato mostra 0:00.

Le opzioni condividono le parentesi, separate da virgole, in qualsiasi ordine.

```showfile
[loop: 10s, timer 7:30pm]
**Welcome!**
The show begins in {timer}
[/loop]
```

- Il tempo rimasto è mostrato in minuti, arrotondato per eccesso e senza secondi (`15 min`), con le ore quando sono più di una (`1 h 5 min`), nella lingua del dispositivo. A zero resta a `0 min`; il ciclo continua finché l'operatore non va avanti.
- Mostrano il timer solo gli elementi il cui testo contiene `{timer}`, quindi puoi alternare un messaggio e il conto alla rovescia.
- Un telefono con l'orologio sbagliato mostra comunque il tempo giusto, perché si usa l'orologio del programma.
- Salvare il file dello spettacolo durante un intervallo non fa ripartire il conto alla rovescia, e nemmeno riavviare il programma.
- In una traduzione, il ciclo deve avere lo stesso timer del file principale dello spettacolo; traduci le parole attorno a `{timer}`.
- Anche l'operatore vede il tempo rimasto nel banner verde. Un `{timer}` in un ciclo senza timer viene mostrato così com'è, con un avviso.

## Dare un nome a un ciclo

Dai un **nome** a un ciclo e il rapporto sul pubblico lo usa per etichettarlo nel suo grafico. Senza nome, il rapporto lo indovina dalla posizione del ciclo (prima della prima riga: *prima dello spettacolo*; tra le righe: *intervallo*; dopo l'ultima riga: *dopo lo spettacolo*).

```showfile
[loop: 12s, name Intermission]
[fullscreen: **INTERMISSION**]
Back soon
[/loop]
```

- Il nome è testo libero fino a 40 caratteri, senza virgole. Il pubblico non lo vede mai.
- Condivide le parentesi con le altre opzioni, separate da virgole, in qualsiasi ordine: `[loop: 10s, name Act 1 break, timer 15 minutes]`.
- Si usa solo il nome del file dello spettacolo della lingua principale; una traduzione può dare altri nomi ai suoi cicli, o nessuno.
