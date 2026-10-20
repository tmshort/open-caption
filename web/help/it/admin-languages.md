# Lingue

Uno spettacolo può offrire i propri sottotitoli in più di una lingua. Ogni lingua è un **proprio file dello spettacolo**, e l'operatore segue solo quello principale: una pressione di Spazio invia lo stesso momento in ogni lingua contemporaneamente.

## Aggiungere una lingua

1. Nel passo **File dello spettacolo**, fai clic su **＋ Aggiungi una lingua** accanto alle schede delle lingue.
2. Scegli la lingua (oppure «Un'altra lingua…» e scrivi un codice breve come `cy`).
3. Lascia spuntata la casella **Parti da una copia del file principale dello spettacolo**. Ottieni una copia della forma giusta, così devi solo sostituire le parole.
4. Traduci, poi **Salva**.

Ogni traduzione inizia con una riga che ne indica la lingua. L'editor la aggiunge per te:

```showfile
[language: es | Español]
Hola, Margaret.
```

Il nome dopo il `|` è ciò che gli spettatori vedono nel menu delle lingue, quindi scrivilo nella sua lingua (Español, Français, Deutsch).

## L'unica regola: mantenere la stessa forma

Ogni lingua deve avere **le stesse battute nello stesso ordine**: un sottotitolo resta un sottotitolo, un `[clear]` resta un `[clear]`, un ciclo mantiene lo stesso numero di elementi, e le voci mostrate insieme restano lo stesso numero di voci. Cambiano solo le parole. I nomi dei personaggi, le note, i nomi delle sezioni, i nomi delle immagini e i tempi dei cicli possono essere diversi.

Se una traduzione smette di coincidere, l'editor mostra un errore con la riga in cui qualcosa non torna, per esempio «La battuta 4 non corrisponde al file dello spettacolo in inglese: quel file ha [clear] qui, e questo ha un sottotitolo». Una lingua che non coincide **non viene offerta al pubblico** finché non è corretta. Il file dello spettacolo in inglese (principale) non è toccato, e il passo Condividi elenca il problema.

Per dire al pubblico quali parole vanno con quale riga, tieni le righe in passo: se l'inglese divide una battuta in tre righe, lo fa anche lo spagnolo, anche dove non sarebbe il modo in cui la lingua la spezzerebbe.

## Cosa vede il pubblico

- A ogni dispositivo viene data la lingua migliore che lo spettacolo ha. Si guardano le impostazioni di lingua del telefono o del browser (per esempio lo spagnolo del Messico riceve lo spagnolo), e si usa la lingua principale quando nulla corrisponde.
- Un piccolo controllo **🌐** nell'angolo in alto a sinistra elenca le lingue. Sceglierne una cambia quel dispositivo, e l'indirizzo finisce con `?lang=es`, così puoi stampare un link o un codice QR per ogni lingua.
- Anche le parole della pagina stessa («Riconnessione…», dimensione del testo, ecc.) sono tradotte in spagnolo, francese, tedesco, italiano e portoghese. Le altre lingue mostrano quelle poche etichette in inglese; i sottotitoli stessi sono sempre nella lingua del file dello spettacolo.
- Le lingue scritte da destra a sinistra (arabo, ebraico) si allineano a destra.

## Altre cose da sapere

- **Immagini:** ogni lingua può usare le stesse immagini. Per un'immagine con delle parole, carica una versione per lingua e indica ciascuna nel file dello spettacolo di quella lingua.
- **Messaggio pre-spettacolo:** ogni lingua ha il proprio, conservato nel file dello spettacolo di quella lingua (`idle message:` nel suo blocco `[settings]`). Il passo Aspetto ha una casella per ogni lingua e li scrive per te. Il file dello spettacolo di una traduzione non può impostare altro: l'aspetto si imposta una volta sola, nel file principale dello spettacolo.
- **Gli annunci dell'operatore** («Annuncio…») vengono inviati esattamente come scritti a ogni lingua, quindi scrivili nelle lingue che servono.
- **Modificare un file dello spettacolo mentre lo spettacolo è in corso:** aggiungere o correggere una traduzione ha effetto subito e mantiene il punto dell'operatore. Rimuovere una lingua non è permesso mentre lo spettacolo è in diretta.
- **Conteggi:** la pagina dell'operatore mostra quanti dispositivi leggono ciascuna lingua, per esempio `👥 83 (en 60 · es 23)`.
- **Copie di sicurezza:** Esporta salva ogni lingua; Duplica le copia tutte.

Vedi anche [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file).
