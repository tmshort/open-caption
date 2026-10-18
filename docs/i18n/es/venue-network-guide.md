# Guía de red del local

Para quien se ocupa de la WiFi del teatro. El equipo de subtitulado no necesita nada de internet. Solo necesita ser **accesible desde los dispositivos del público**.

## Lo que hace el público

1. Se conecta a la WiFi de invitados.
2. Abre una dirección, por ejemplo `http://192.168.1.10/` (o escanea el código QR).

## Requisitos

| Requisito | Por qué |
|-----------|---------|
| La red de invitados puede llegar al equipo de subtitulado | Los dispositivos piden la página y mantienen abierta una conexión con él |
| El equipo tiene una dirección fija (reserva DHCP o estática) | Para que los códigos QR y carteles impresos sigan valiendo entre funciones |
| El punto de acceso aguanta más de 100 dispositivos a la vez (con 5 GHz activado) | Una sala llena son unos 150 dispositivos |
| No hace falta salida a internet | Todo lo sirve el equipo de subtitulado |

**El aislamiento de clientes entre invitados está bien y es recomendable.** Los dispositivos del público hablan solo con el equipo de subtitulado, nunca entre sí. Lo que *no* debe bloquearse es el tráfico de los invitados hacia el servidor. Si la red de invitados está aislada de la red local principal, conecte el equipo de subtitulado a la red de invitados (un puerto libre de la VLAN de invitados o la SSID de invitados) o añada una regla de cortafuegos que permita a los invitados llegar a la dirección del equipo por el puerto TCP 80 (y el 443 si se activa HTTPS).

## Comprobar antes de la función

Desde un teléfono en la WiFi de invitados, abra la dirección que se muestra en Configuración → **Dirigir**. Si la página carga y dice «Los subtítulos comenzarán en breve» (o muestra la previa), ya está. Si no, consulte «No carga nada» más abajo.

## Hacer la dirección más fácil

En orden de esfuerzo:

1. **Código QR** — Configuración → Dirigir → «Imprimir un cartel con código QR». Póngalo en el vestíbulo y en el programa de mano. No hace falta cambiar nada en la red.
2. **Puerto 80** — si el programa funciona en el puerto 80, la dirección no lleva `:puerto`. En Linux, permítalo con `sudo setcap 'cap_net_bind_service=+ep' ./open-caption`.
3. **Un nombre local** — añada en el router una entrada DNS para el equipo de subtitulado, por ejemplo `captions.theater` → `192.168.1.10`. La mayoría de routers (OpenWrt, Pi-hole, UniFi, Omada, basados en dnsmasq) permiten un «registro DNS local» o «anulación de host».
4. **Enviar a los invitados allí automáticamente.** Opciones:
   - **Página de bienvenida / portal cautivo** en el punto de acceso que apunte a la dirección del público. La función de «portal de invitados» integrada del punto de acceso funciona con una URL externa (UniFi: Hotspot → portal externo; OpenWrt: nodogsplash `RedirectURL`). Asegúrese de que deje cargar la página sin iniciar sesión.
   - **Opción 114 de DHCP (RFC 8910) y `/captive-api`.** El programa responde a `GET /captive-api` como describe la RFC 8908: `{"captive": false, "user-portal-url": "http://<server-address>/"}`. Si su red ya reparte la opción 114 («Captive-Portal API URI»), puede apuntarla a esa dirección. Tenga presente lo que hace: indica a los dispositivos que la red *no* es un portal cautivo y publica la dirección de la página. **No** hace que los teléfonos abran solos la página de subtítulos (eso requiere `"captive": true`, que el programa nunca envía). La RFC también exige que la dirección de la API sea **https con un certificado en el que confíen los dispositivos**, así que una dirección de opción 114 con `http://` sin más o el certificado autofirmado del programa será ignorado. Esto es experimental y no se ha probado en dispositivos reales: use el cartel QR, o la página de bienvenida de su punto de acceso, para que la gente llegue a la página.
5. **mDNS** (`open-caption.local`) se incluye y viene activado por defecto: los dispositivos que pueden usarlo abren `http://open-caption.local/` sin configurar nada en el router. Necesita que la red de invitados permita **multidifusión** (UDP 5353 hacia 224.0.0.251) entre el equipo de subtitulado y los dispositivos del público. Muchas redes de invitados la bloquean («filtrado de multidifusión» o «bloqueo de mDNS/Bonjour» en el punto de acceso), y entonces el nombre simplemente no se resuelve; la dirección numérica y el código QR siguen funcionando. Si quiere que el nombre funcione, active el reenvío mDNS/Bonjour o desactive el filtrado de multidifusión en la red de invitados. El nombre se puede cambiar, o desactivar, en Configuración → Ajustes → *Nombre de red*.

El programa no ejecuta su propio servidor DHCP ni DNS, a propósito: chocaría con el router del local.

## HTTP y HTTPS

Use `http://` sencillo para el público. Un certificado HTTPS autofirmado hace que cada dispositivo muestre una advertencia de seguridad. HTTPS está disponible (`https.enabled: true` en `config.yaml`) para sitios que puedan instalar un certificado en el que confíen sus dispositivos. Si el local tiene un nombre de dominio propio, el programa puede obtener y renovar solo un certificado de confianza; véase la [guía del certificado HTTPS de confianza](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

Un pequeño inconveniente del HTTP sencillo: los navegadores solo permiten «mantener la pantalla encendida» en páginas seguras. Con HTTP la página mantiene encendida la pantalla del teléfono con un pequeño vídeo en bucle, silencioso e invisible. Un teléfono en modo de bajo consumo puede rechazarlo; entonces la página muestra un consejo, una sola vez, que pide a la gente desactivar el bloqueo automático.

## Seguridad

- Configuración (`/admin`) y la página del operador (`/operator`) necesitan cada una una contraseña de `config.yaml`. La página del público no necesita ninguna y no puede cambiar nada.
- No exponga el puerto a internet.
- Las contraseñas de operador y de administrador se guardan en texto sin cifrar en `config.yaml`, en la carpeta de datos del programa. Mantenga esa carpeta privada para quienes llevan el espectáculo.

## Cortafuegos en el equipo de subtitulado

- **Windows:** cuando el programa arranque por primera vez, permítalo en **Redes privadas**. Si se descartó el aviso: Seguridad de Windows → Firewall → Permitir una aplicación.
- **macOS:** permita las conexiones entrantes para el programa cuando se pregunte, o añádalo en Ajustes del Sistema → Red → Firewall.
- **Linux:** abra el puerto, por ejemplo `sudo ufw allow 80/tcp`.

## No carga nada

1. ¿Está el teléfono en la WiFi de invitados (no en datos móviles)?
2. ¿La dirección incluye el puerto si el programa no está en el puerto 80? El paso Dirigir muestra la dirección correcta.
3. ¿Puede otro equipo de la WiFi de invitados llegar a él (`curl http://192.168.1.10/healthz` debería imprimir `ok`)?
4. ¿Está el tráfico de invitados al servidor bloqueado por el router o el punto de acceso (aislamiento de clientes en «bloquear todo» en vez de «bloquear entre clientes»)?
5. ¿Está el cortafuegos del equipo bloqueando el programa?
6. ¿Sigue el equipo despierto, y con la misma dirección? Busque un cambio de dirección en Dirigir.

Configuración → Dirigir → **Descargar un paquete de soporte** recoge los ajustes (sin contraseñas), líneas recientes del registro y datos básicos del equipo.

## Capacidad

El programa se ha probado con 200 y 500 conexiones de público simultáneas en un solo equipo (latencia inferior a 10 ms en la misma máquina). En un local real el límite es la WiFi, no el programa: use un punto de acceso preparado para el público, active 5 GHz y mantenga las imágenes pequeñas (menos de 1 MB) para que muchos dispositivos puedan pedirlas a la vez.
