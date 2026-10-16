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
3. **Un nome** — il programma si annuncia come **`open-caption.local`**, quindi `http://open-caption.local/` funziona sulla maggior parte di telefoni e computer della stessa rete, senza configurare nulla. Cambia il nome, o disattivalo, in **Impostazioni → Nome di rete**. Alcune reti WiFi ospiti bloccano questo tipo di annuncio; usa allora l'indirizzo numerico o il codice QR. Il locale può anche aggiungere sul proprio router un nome suo (come `captions.theater`), che funziona ovunque.
4. **Pagina automatica** — un locale può mandare gli ospiti alla pagina quando si collegano al WiFi con la pagina di benvenuto del punto di accesso («captive portal»). Il programma risponde anche all'indirizzo standard di rilevamento della rete `/captive-api` per i locali che usano l'opzione 114 del DHCP, ma questo si limita a segnalare che la rete non è captive: non apre la pagina da solo, e richiede un certificato HTTPS fidato. Vedi la [guida di rete del locale](topic:admin-venue-network-guide).

## HTTP e HTTPS

Gli indirizzi `http://` semplici funzionano su ogni dispositivo e sono la scelta predefinita. HTTPS si può attivare in `config.yaml`; con un certificato autofirmato ogni dispositivo mostra un avviso di sicurezza la prima volta. L'HTTP semplice è più semplice per un pubblico.

## Se la porta è occupata

Se la porta 80 è già usata da un altro programma, il server parte su un'altra porta e lo dice nella console. Gli indirizzi del pubblico devono allora includerla, per esempio `http://192.168.1.10:8080/`; il passo Conduci mostra l'indirizzo giusto.
