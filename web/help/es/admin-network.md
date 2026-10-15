# Red y conexión

Los miembros del público conectan sus propios teléfonos y tabletas a la **WiFi de invitados** del local y abren una dirección. No hay que instalar nada.

## Qué debe proporcionar el local

- Este equipo y los dispositivos del público deben estar en la misma red, o la red debe permitir que los dispositivos invitados lleguen a este equipo. Es un ajuste del sistema WiFi del local, no de este programa.
- Muchas redes de invitados aíslan los dispositivos entre sí. Eso no importa. Los dispositivos solo hablan con este equipo, nunca entre sí, siempre que este equipo sea accesible.
- Lo mejor es una conexión por cable para este equipo. Pida una dirección reservada (una «reserva DHCP») para que la dirección no cambie entre funciones.
- No se necesita conexión a internet.

Pruébelo antes del primer espectáculo: conéctese a la WiFi de invitados con un teléfono y abra la dirección que se muestra en el paso **En vivo**.

## Hacer la dirección fácil

1. **Código QR** — imprima el cartel desde el paso En vivo y póngalo en el vestíbulo y en el programa de mano. Es la opción más sencilla.
2. **Una dirección corta** — si el programa puede usar el puerto 80, la dirección no lleva número de puerto, como `http://192.168.1.10/`.
3. **Un nombre** — el programa se anuncia como **`open-caption.local`**, así que `http://open-caption.local/` funciona en la mayoría de teléfonos y ordenadores de la misma red, sin configurar nada. Cambie el nombre, o desactívelo, en **Ajustes → Nombre de red**. Algunas redes WiFi de invitados bloquean este tipo de anuncio; entonces use la dirección numérica o el código QR. El local también puede añadir en su router un nombre propio (como `captions.theater`), que funciona en todas partes.
4. **Página automática** — un local puede enviar a los invitados a la página cuando se conectan a la WiFi con la página de bienvenida del punto de acceso («portal cautivo»). El programa también responde a la dirección estándar de detección de red `/captive-api` para locales que usan la opción 114 de DHCP, pero eso solo indica que la red no es cautiva: no abre la página por sí solo, y necesita un certificado HTTPS de confianza. Consulte la [guía de red del local](topic:admin-venue-network-guide).

## HTTP y HTTPS

Las direcciones `http://` sencillas funcionan en todos los dispositivos y son la opción por defecto. HTTPS se puede activar en `config.yaml`; con un certificado autofirmado cada dispositivo muestra una advertencia de seguridad la primera vez. HTTP sencillo es más simple para un público.

## Si el puerto está ocupado

Si el puerto 80 ya lo usa otro programa, el servidor arranca en otro puerto y lo indica en la consola. Entonces las direcciones del público deben incluirlo, por ejemplo `http://192.168.1.10:8080/`; el paso En vivo muestra la dirección correcta.
