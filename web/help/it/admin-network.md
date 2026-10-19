# Rete e connessione

Gli spettatori collegano i propri telefoni e tablet al **WiFi ospiti** del locale e aprono un indirizzo. Non c'è nulla da installare.

## Cosa deve fornire il locale

- Questo computer e i dispositivi del pubblico devono essere sulla stessa rete, oppure la rete deve permettere ai dispositivi ospiti di raggiungere questo computer. È un'impostazione del sistema WiFi del locale, non di questo programma.
- Molte reti ospiti isolano i dispositivi l'uno dall'altro. Va bene. I dispositivi parlano solo con questo computer, mai tra loro, finché questo computer è raggiungibile.
- Per questo computer è meglio una connessione via cavo. Chiedi un indirizzo riservato (una «prenotazione DHCP») perché l'indirizzo non cambi da uno spettacolo all'altro.
- Non serve alcuna connessione a internet.

Provalo prima del primo spettacolo: collega un telefono al WiFi ospiti e apri l'indirizzo mostrato nel passo **Conduci**. Per un proiettore o uno schermo grande, lo stesso passo elenca un link per il proiettore: aprilo sul computer a cui è collegato il proiettore.

## Rendere facile l'indirizzo

1. **Codice QR** — stampa il cartello dal passo Conduci e mettilo nell'ingresso e nel programma di sala. È l'opzione più semplice.
2. **Un indirizzo breve** — se il programma può usare la porta 80, l'indirizzo non ha numero di porta, come `http://192.168.1.10/`.
3. **Un nome** — il programma si annuncia come **`open-caption.local`**, quindi `http://open-caption.local/` funziona sulla maggior parte di telefoni e computer della stessa rete, senza configurare nulla. Cambia il nome, o disattivalo, nel passo **Conduci**, sulla riga dell'indirizzo `.local` (**Rinomina…** e la sua casella). Alcune reti WiFi ospiti bloccano questo tipo di annuncio; usa allora l'indirizzo numerico o il codice QR. Il locale può anche aggiungere sul proprio router un nome suo (come `captions.theater`), che funziona ovunque.
4. **Pagina automatica** — un locale può mandare gli ospiti alla pagina quando si collegano al WiFi con la pagina di benvenuto del punto di accesso («captive portal»). Il programma risponde anche all'indirizzo standard di rilevamento della rete `/captive-api` per i locali che usano l'opzione 114 del DHCP, ma questo si limita a segnalare che la rete non è captive: non apre la pagina da solo, e richiede un certificato HTTPS fidato. Vedi la [guida di rete del locale](topic:admin-venue-network-guide).

## Se questo computer ha più di un indirizzo

Un computer con cavo e WiFi insieme, o con macchine virtuali o una VPN, ha più indirizzi. Il passo Conduci elenca quelli degli adattatori connessi, con la connessione principale per prima; è quella che il codice QR usa all'inizio. Per dare al pubblico solo alcuni indirizzi (per esempio solo la rete WiFi del locale), apri **Impostazioni** (pulsante in alto) e la sua scheda **Rete**: **Mostra solo questi** (vuoto significa tutti) e **Non mostrare mai questi**, uno per riga, ciascuno un indirizzo (`192.168.1.50`), una rete (`192.168.1.0/24`) o la parola `private`. La modifica ha effetto subito. Se due adattatori sono sulla stessa rete (un cavo e il WiFi, entrambi con un indirizzo 192.168.1.x), se ne usa uno solo, così il codice QR e il nome `.local` coincidono: la connessione principale, altrimenti quella cablata. Per scegliere, premi **Usa questo** accanto all'altra riga, o scrivi il nome dell'adattatore (o Ethernet o Wi-Fi) in **Preferisci questi adattatori**. Gli indirizzi esclusi non compaiono nell'elenco, né nel codice QR, non sono annunciati come nome `.local` e non entrano in un nuovo certificato HTTPS autofirmato. Per impostazione predefinita questo sceglie solo ciò che viene offerto: un telefono che ha già un indirizzo nascosto può ancora collegarsi. Seleziona **Rifiuta anche le connessioni agli indirizzi che non sono mostrati** per respingere anche quelli; le connessioni al computer stesso sono sempre accettate, quindi la configurazione resta raggiungibile. La configurazione non ti lascia mai disattivare l'indirizzo attraverso cui la usi: la finestra lo segnala.

## HTTP e HTTPS

Gli indirizzi `http://` semplici funzionano su ogni dispositivo e sono la scelta predefinita. HTTPS si può attivare in `config.yaml`; con un certificato autofirmato ogni dispositivo mostra un avviso di sicurezza la prima volta. L'HTTP semplice è più semplice per un pubblico.

## Se la porta è occupata

Se la porta 80 è già usata da un altro programma, il server parte su un'altra porta e lo dice nella console. Gli indirizzi del pubblico devono allora includerla, per esempio `http://192.168.1.10:8080/`; il passo Conduci mostra l'indirizzo giusto.
