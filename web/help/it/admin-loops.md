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

Sottotitoli, `[clear]`, sezioni, note e altri cicli non possono stare dentro un ciclo.

## Quando un ciclo parte e si ferma

- Se la **prima cosa** nel file dello spettacolo è un ciclo, parte da solo non appena lo spettacolo viene avviato, così il pubblico vede il pre-spettacolo mentre cerca il posto.
- Qualsiasi altro ciclo (per esempio nell'intervallo) parte quando l'operatore preme la barra spaziatrice su di esso.
- La barra spaziatrice termina il ciclo e invia la riga successiva. **Maiusc+Q** ferma il ciclo e lascia lo schermo vuoto. **Indietro** torna al ciclo e lo riavvia dal primo elemento.

L'elenco dell'operatore mostra un ciclo come una sola voce che si espande per elencarne gli elementi. Vedi [I cicli per l'operatore](topic:operator-loops).
