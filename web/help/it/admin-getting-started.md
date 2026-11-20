# Primi passi

Preparare uno spettacolo richiede sei passi (un settimo, **Rapporto**, è per dopo lo spettacolo). La fila di pulsanti in alto nella pagina li segue in ordine, e compare una spunta quando un passo sembra completo. Puoi passare a qualsiasi passo in qualsiasi momento.

1. **Spettacoli** — crea uno spettacolo (o lo spettacolo di esempio per dare un'occhiata).
2. **File dello spettacolo** — incolla o apri il file dello spettacolo e correggi ciò che la pagina segnala.
3. **Immagini** — carica le immagini usate dal file dello spettacolo.
4. **Aspetto** — scegli i colori e la dimensione del testo.
5. **Verifica** — fai una prova in privato prima che ci sia qualcuno in sala.
6. **Condividi** — condividi l'indirizzo con il pubblico e apri la pagina dell'operatore, da cui si avvia lo spettacolo. Facoltativo: la pagina Segui (`/follow`) permette a un secondo dispositivo di guardare senza poter inviare nulla.
7. **Rapporto** — dopo lo spettacolo: quanti dispositivi hanno usato la pagina del pubblico, con un grafico nel tempo.

La prima volta che si apre la configurazione senza spettacoli, una breve **visita guidata** mostra ogni passo e i pulsanti nell'intestazione; premi Esc per saltarla. Per rivederla, vai al passo Spettacoli, apri l'Aiuto e scegli **Visita guidata di questo passo**. In ogni altro passo lo stesso pulsante mostra brevemente ciò che contiene quel passo.

## Attesa e messa in onda

Quando si apre la pagina dell'operatore, lo spettacolo viene caricato **in attesa**: il pubblico vede la schermata di attesa della sala (un logo e un messaggio, impostati una sola volta in **Impostazioni → Schermo del pubblico**) e nulla dello spettacolo. L'operatore preme **Spazio** (il pulsante dice **AVVIA**) per mettere lo spettacolo in onda, e **Metti in attesa…** lo toglie dall'onda senza perdere il punto in cui era. Finché lo spettacolo è in attesa, il pubblico non vede mai la preparazione, le prove o le modifiche, e uno spettacolo in attesa può essere eliminato.

## Dove viene conservato tutto

Uno spettacolo è una cartella che contiene il suo **file dello spettacolo** (un semplice file di testo, `show.md`), le eventuali traduzioni e le sue immagini. Il file dello spettacolo contiene tutto ciò che riguarda lo spettacolo, tranne le immagini: il titolo, l'aspetto e le righe. Le impostazioni che il file dello spettacolo non menziona provengono da `config.yaml` e poi dai valori incorporati del programma.

## Hai finito con uno spettacolo?

Rinomina, Duplica, Esporta, Archivia ed Elimina sono nel menu **Altro ▾** di ogni spettacolo. Nel passo Spettacoli, **Archivia** mette da parte uno spettacolo: viene conservato per intero ma elencato sotto *Spettacoli archiviati* e non può andare in onda. **Ripristina** lo riporta indietro. **Elimina** lo rimuove definitivamente (se pensi che possa servirti ancora, esportalo prima). Lo spettacolo in onda non può essere né archiviato né eliminato.

## Prima di iniziare

- Questo computer deve essere sulla stessa rete Wi-Fi dei telefoni e dei tablet del pubblico. Vedi [Rete e connessione](admin-network.md).
- La password amministratore e la password operatore sono in `config.yaml` solo come hash; sono state mostrate una sola volta nella console al primo avvio del programma. Per impostarne di nuove, usa **Impostazioni → Password**. La pagina dell'operatore ha una password propria, così puoi darla a un'altra persona senza concederle l'accesso alla configurazione. Sul computer che esegue il programma, Configurazione e la pagina dell'operatore si aprono **senza password** quando l'indirizzo inizia con `localhost`, `127.0.0.1` o `[::1]` (chi usa quel computer è considerato l'amministratore). Gli altri computer, e qualsiasi richiesta inoltrata da un proxy, accedono ancora con una password. Imposta `access.local_auto_login: false` in `config.yaml` se altre persone possono accedere a questo computer.
- La configurazione, la pagina dell'operatore e questo aiuto sono disponibili in inglese, spagnolo, francese, tedesco, italiano e portoghese. Usa il menu della lingua in alto in ogni pagina; la scelta viene ricordata in quel browser.
- Il tuo lavoro viene salvato su questo computer. Usa **Esporta** (menu **Altro ▾** di uno spettacolo) nel passo Spettacoli per fare una copia di sicurezza da portare su un altro computer. L'esportazione è uno `.zip` che contiene `show.md` (il file dello spettacolo, con titolo e aspetto), un `show.<lingua>.md` per ogni traduzione e una cartella `assets` con le immagini; **Importa** trasforma uno zip così in un nuovo spettacolo, e accetta anche un file dello spettacolo di testo semplice (`.md` o `.txt`, con qualsiasi nome) come nuovo spettacolo.

## Non hai mai scritto un file dello spettacolo?

Un file dello spettacolo è solo testo. Ogni riga è un sottotitolo, e ogni pressione della barra spaziatrice invia una riga. Per tutto il lavoro, dal copione al file finito (compresa la modifica con un tuo programma e il ricaricamento del file), leggi [Preparare un file dello spettacolo](admin-preparing-a-show-file.md). I marcatori sono elencati in [Scrivere un file dello spettacolo](admin-writing-a-show-file.md). Oppure crea lo spettacolo di esempio e guarda come è scritto. Un modello scaricabile si trova su [sample-show.md](../../sample/sample-show.md).

## Pianificare l'attrezzatura

Per sapere cosa acquistare, dove mettere il punto di accesso Wi-Fi e avere una lista di controllo per il giorno dello spettacolo, leggi la [guida all'uso](../../../docs/i18n/it/usage-guide.md). La [guida alla rete della sala](../../../docs/i18n/it/venue-network-guide.md) è per chi si occupa della rete Wi-Fi dell'edificio.

## Dove andare dopo

- [Musical](admin-musicals.md)
- [Scrivere un file dello spettacolo](admin-writing-a-show-file.md)
- [Cicli: prima dello spettacolo e intervallo](admin-loops.md)
- [Risoluzione dei problemi](admin-troubleshooting.md)
