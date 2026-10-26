# Aspetto

Il passo **Aspetto** imposta come appaiono i sottotitoli su ogni dispositivo del pubblico. I valori predefiniti, testo bianco su sfondo nero, sono i più facili da leggere in una sala buia e i migliori per l'accessibilità.

- **Colori:** una lettura del contrasto mostra quanto è leggibile la combinazione. Punta a 7:1 o meglio; la pagina avvisa quando è troppo basso.
- **Dimensione del testo:** scala tutto il testo. I dispositivi adattano il testo al loro schermo, quindi è una variazione relativa, non una misura in punti.
- **Carattere:** pulito (il carattere abituale del dispositivo), con grazie, a spaziatura fissa, oppure **adatto alla dislessia (OpenDyslexic)**, un carattere con la parte bassa delle lettere più pesante e forme distintive che alcuni lettori dislessici trovano più facili. È incorporato nel programma, quindi non serve internet, e un dispositivo lo scarica (circa 115 KB) solo quando viene usato. Le righe guadagnano un po' di altezza. Scegli **alta leggibilità (Atkinson Hyperlegible)** per un carattere pulito progettato dal Braille Institute per l'ipovisione, con lettere facili da distinguere (per esempio la I maiuscola, la l minuscola e il numero 1). Entrambi i caratteri sono incorporati nel programma, e un dispositivo ne scarica uno solo se lo usa (circa 17 KB per Atkinson Hyperlegible). Gli spettatori possono anche scegliere l'uno o l'altro per il proprio dispositivo con il piccolo controllo **Aa**, qualunque sia il carattere scelto qui.
- **Nomi dei personaggi:** mostrati sopra la riga in un solo colore (il *colore dei nomi dei personaggi*), in un **colore diverso per ogni personaggio**, oppure nascosti. I colori dei personaggi sono scelti automaticamente, nell'ordine in cui i personaggi parlano per la prima volta, da una tavolozza adatta agli spettatori daltonici, e regolati per restare leggibili sul tuo sfondo.
- **Messaggio pre-spettacolo:** mostrato finché non viene inviata la prima riga, se il file dello spettacolo non inizia con un ciclo.

Le modifiche si salvano con **Salva l'aspetto** e raggiungono subito i dispositivi connessi. **Salva l'aspetto scrive le tue scelte nel blocco `[settings]` in cima al file dello spettacolo** (vedi [Scrivere un file dello spettacolo](topic:admin-writing-a-show-file)), così il file dello spettacolo porta con sé il proprio aspetto ovunque venga copiato; puoi anche modificare quel blocco a mano. Tutto ciò che il file dello spettacolo non imposta usa i valori predefiniti di `config.yaml` (la sua sezione `defaults`), che a loro volta partono dai valori incorporati del programma. Se l'editor del file dello spettacolo ha modifiche non salvate, salvale prima di salvare l'aspetto.

## Uno schermo per tutta la sala

Se un browser mostra i sottotitoli su un proiettore o un grande monitor, apri l'indirizzo del pubblico con `?display=projector` in fondo. Il testo è più grande, il puntatore del mouse è nascosto e non viene mostrato altro. Uno schermo condiviso non scorre mai: se l'ultimo sottotitolo è più alto dello schermo (ad esempio con molte voci insieme), il testo viene rimpicciolito finché non ci sta.

## Rileggere righe precedenti

Chiunque può far scorrere i sottotitoli verso l'alto per rileggere qualcosa che si è perso. Le nuove righe aspettano allora dietro un pulsante **↓ Torna in diretta (3)** (che conta ciò che è arrivato) invece di trascinare via la pagina, e le righe più vecchie smettono di sfumare mentre qualcuno le legge. Toccare il pulsante, o scorrere fino in fondo, riporta ai sottotitoli in diretta. Uno schermo con `?display=projector` non è mai scorrevole.

## Controlli per il pubblico

Gli spettatori non devono mai fare nulla. Un piccolo controllo tenue **Aa** nell'angolo in alto a destra della pagina del pubblico permette a ognuno di ingrandire o rimpicciolire il proprio testo, passare a scuro su chiaro, scegliere il carattere adatto alla dislessia o quello ad alta leggibilità, oppure disattivare **Lo schermo resta acceso** (la pagina tiene acceso lo schermo del telefono durante lo spettacolo, e chi ha poca batteria può preferire di no), ma solo sul proprio dispositivo.
