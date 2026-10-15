# Primi passi

Preparare uno spettacolo richiede sei passi. La fila di pulsanti in alto nella pagina li segue in ordine, e compare una spunta quando un passo sembra completo. Puoi passare a qualsiasi passo in qualsiasi momento.

1. **Spettacoli** — crea uno spettacolo (o lo spettacolo di esempio per dare un'occhiata).
2. **File dello spettacolo** — incolla o apri il file dello spettacolo e correggi ciò che la pagina segnala.
3. **Immagini** — carica le immagini usate dal file dello spettacolo.
4. **Aspetto** — scegli i colori e la dimensione del testo.
5. **Prova** — fai una prova in privato prima che ci sia qualcuno in sala.
6. **In diretta** — condividi l'indirizzo con il pubblico e avvia lo spettacolo.

## Dove viene conservato tutto

Uno spettacolo è una cartella che contiene il suo **file dello spettacolo** (un semplice file di testo, `show.md`), le eventuali traduzioni e le sue immagini. Il file dello spettacolo contiene tutto sullo spettacolo tranne le immagini: il titolo, l'aspetto e le righe. Le impostazioni che il file dello spettacolo non menziona provengono da `config.yaml` e poi dai valori incorporati del programma.

## Hai finito con uno spettacolo?

Nel passo Spettacoli, **Archivia** mette da parte uno spettacolo: viene conservato per intero ma elencato sotto *Spettacoli archiviati* e non può andare in onda. **Ripristina** lo riporta indietro. **Elimina** lo rimuove definitivamente (esportalo prima se potresti averne ancora bisogno). Lo spettacolo in onda non può essere né archiviato né eliminato.

## Prima di iniziare

- Questo computer deve essere nella stessa rete WiFi dei telefoni e tablet del pubblico. Vedi [Rete e connessione](topic:admin-network).
- La password amministratore e la password operatore sono in `config.yaml`; sono state anche mostrate nella console al primo avvio del programma. La pagina dell'operatore ha una password propria, così puoi darla a un'altra persona senza concederle l'accesso alla configurazione.
- La configurazione, la pagina dell'operatore e questo aiuto sono disponibili in inglese, spagnolo, francese, tedesco, italiano e portoghese. Usa il menu della lingua in alto in ogni pagina; la scelta viene ricordata in quel browser.
- Il tuo lavoro viene salvato su questo computer. Usa **Esporta** nel passo Spettacoli per fare una copia di sicurezza da portare su un altro computer. L'esportazione è uno `.zip` che contiene `show.md` (il file dello spettacolo, con titolo e aspetto), un `show.<lingua>.md` per ogni traduzione e una cartella `assets` con le immagini; **Importa** trasforma uno zip così in un nuovo spettacolo, e accetta anche un file dello spettacolo di testo semplice (`.md` o `.txt`, con qualsiasi nome) come nuovo spettacolo.

## Sei nuovo ai file dello spettacolo?

Un file dello spettacolo è solo testo. Ogni riga è un sottotitolo, e una pressione della barra spaziatrice invia una riga. Per tutto il lavoro, dal copione al file finito (compresa la modifica nel tuo programma e il ritorno del file), leggi [Preparare un file dello spettacolo](topic:admin-preparing-a-show-file). I marcatori sono elencati in [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file). Oppure crea lo spettacolo di esempio e guarda come è scritto. Un modello scaricabile si trova su [/help/sample-show.md](/help/sample-show.md).

## Pianificare l'attrezzatura

Per sapere cosa acquistare, dove mettere il punto di accesso WiFi e avere una lista di controllo per il giorno dello spettacolo, leggi la [guida all'uso](topic:admin-usage-guide). La [guida di rete del locale](topic:admin-venue-network-guide) è per chi si occupa del WiFi dell'edificio.

## Dove andare dopo

- [Musical](topic:admin-musicals)
- [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file)
- [Cicli: prima dello spettacolo e intervallo](topic:admin-loops)
- [Risoluzione dei problemi](topic:admin-troubleshooting)
