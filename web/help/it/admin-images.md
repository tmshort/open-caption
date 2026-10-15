# Immagini

Carica le immagini nel passo **Immagini** e usale nel file dello spettacolo con `[image: nome.png]`.

- Tipi accettati: PNG, JPG, GIF e WebP. Gli altri tipi, compreso SVG, non sono accettati.
- Ogni file può pesare fino a 10 MB, ma **tienili sotto circa 1 MB**. Ogni dispositivo del pubblico scarica l'immagine intera, e 100 dispositivi che richiedono una grande immagine nello stesso momento possono essere lenti. Circa 1920 pixel di larghezza bastano.
- Il programma non ridimensiona le immagini. I browser le riducono per adattarle a ogni schermo.
- Le immagini in arrivo vengono scaricate in background prima di essere mostrate, così compaiono all'istante.

## Nomi

Il file dello spettacolo si riferisce a un'immagine con il suo nome di file, come `sipario.jpg`. I nomi con spazi funzionano. Se il file dello spettacolo usa un nome che non hai caricato, il passo Immagini lo elenca sotto «Immagini che il file dello spettacolo usa ma che non hai caricato», con un pulsante per caricare il file giusto con quel nome. Un errore lieve come `siparo.jpg` viene segnalato con un suggerimento.

## Sostituire ed eliminare

Caricare un file con lo stesso nome lo sostituisce. Eliminare un'immagine che il file dello spettacolo usa ancora è consentito dopo un avviso; il file dello spettacolo mostrerà allora un avviso «non trovata» per quell'immagine.

## In un ciclo

Le immagini di un ciclo si alternano sullo schermo: `[image: sponsor.png | 15s]` mostra quella per 15 secondi. Vedi [Cicli](topic:admin-loops).
