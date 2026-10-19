# Configurar un certificado HTTPS de confianza (opcional)

Esta guía es para la persona que cuida del equipo de subtitulado. Es **opcional**: los subtítulos funcionan sin ella. (Los ajustes que usa están en la [referencia de configuración](../../configuration-reference.md).)

## ¿Lo necesita?

Probablemente no. Los teléfonos del público abren `http://…` sin ninguna advertencia, y esa es la dirección que usa el código QR. Una dirección segura `https://` solo hace falta para algunas funciones del navegador limitadas a páginas seguras (como mantener la pantalla del teléfono encendida de la forma correcta; la página del público tiene una alternativa), o si el local lo desea.

Sin esta guía, HTTPS sigue funcionando con un certificado **autofirmado**, pero cada teléfono muestra una advertencia de seguridad. Con ella, el programa obtiene un certificado real y gratuito de Let's Encrypt y **lo renueva solo**, de modo que `https://captions.mytheater.org` se abre sin advertencias en todos los dispositivos.

## Qué necesita antes de empezar

| Necesita | Para qué |
|---|---|
| Un **nombre de dominio propio**, por ejemplo `mytheater.org`, y un nombre dentro de él para los subtítulos, como `captions.mytheater.org` | Un certificado se emite para un nombre público. `open-caption.local` y los números sueltos no pueden obtenerlo |
| Su **DNS alojado en un proveedor compatible con este programa** | El programa demuestra que el nombre es suyo añadiendo un registro temporal en su DNS. El programa por defecto incluye **Cloudflare, GoDaddy y EasyDNS**; ejecute `open-caption --version` para ver qué tiene el suyo. Otro proveedor necesita otra compilación (véase el final) |
| **Credenciales de API** de esa cuenta de DNS | Permiten al programa añadir y quitar ese único registro |
| Que el equipo tenga **acceso a Internet** (solo saliente) | Habla con Let's Encrypt y con su proveedor de DNS. Nada necesita llegar *hasta* el local |
| Una **dirección fija** para el equipo en la red del local (basta una reserva de DHCP) | El nombre público apuntará a ella |

## Pasos

**1. Obtenga las credenciales de API.** En la cuenta de su proveedor de DNS:

- **EasyDNS:** active el acceso a la API y cree un **token y una clave** de API. Los ajustes son `EASYDNS_TOKEN` y `EASYDNS_KEY`. (EasyDNS tiene un entorno de pruebas para la API; consulte su documentación para los pasos actuales.)
- **Cloudflare:** cree un **token de API** limitado a su zona con el permiso *Zone → DNS → Edit*. El ajuste es `CLOUDFLARE_DNS_API_TOKEN`.
- **GoDaddy:** cree una **clave y un secreto** de API de producción. Los ajustes son `GODADDY_API_KEY` y `GODADDY_API_SECRET`. (GoDaddy puede limitar su API de DNS a algunos tipos de cuenta; consulte sus reglas actuales.)
- Cualquier otro proveedor: los nombres exactos de los ajustes están en la [documentación de lego](https://go-acme.github.io/lego/dns/).

Un token de DNS suele poder cambiar **todos** los registros de la zona, no solo el temporal, así que dele el menor poder que permita el proveedor. Para la máxima protección use un CNAME (véase «Credenciales más seguras» más abajo).

**2. Haga que el nombre apunte al equipo.** En su proveedor de DNS, añada un registro `A`: `captions.mytheater.org` → la dirección del equipo en la red del local (por ejemplo `192.168.1.20`). Está bien que un nombre público apunte a una dirección privada; al certificado le da igual.

**3. Guarde las credenciales en un archivo privado.** Cree `/etc/open-caption/dns.env` (sirve cualquier lugar privado; en Windows use una carpeta que solo su cuenta pueda leer) con una línea por ajuste, por ejemplo para EasyDNS:

```
EASYDNS_TOKEN=paste-the-token-here
EASYDNS_KEY=paste-the-key-here
```

En macOS y Linux hágalo privado: `chmod 600 /etc/open-caption/dns.env`. El programa avisa si otros usuarios pueden leerlo.

**4. Edite `config.yaml`** (las primeras líneas que imprime el programa dicen dónde está; cada ajuste se explica en la [referencia](../../configuration-reference.md)):

```yaml
https:
  enabled: yes
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org      # Let's Encrypt writes here before a certificate expires
    agree_to_terms: yes                 # you accept Let's Encrypt's terms of service
    directory: letsencrypt-staging      # practice first (see step 5)
    dns:
      provider: easydns                 # or cloudflare, godaddy
      credentials_file: /etc/open-caption/dns.env
```

**5. Practique primero con el servicio de pruebas.** `letsencrypt-staging` es el servicio de pruebas de Let's Encrypt: funciona exactamente como el real, pero sus certificados **no son de confianza para los navegadores** (se espera una advertencia) y tiene límites generosos. Inicie el programa.

**6. Compruebe que ha funcionado.** Abra **Configuración → Ajustes → Seguridad → Certificado**. El estado pasa de *Esperando el primer certificado* a **Válido** en pocos minutos, y muestra el nombre, quién lo emitió y la fecha de caducidad. El registro del programa dice `HTTPS: got a certificate for … valid until …`. Si dice *El último intento falló*, lea **Último problema** y la tabla de más abajo.

**7. Pase al certificado real.** Cambie `directory: letsencrypt-staging` por `directory: letsencrypt` (o borre la línea) y reinicie el programa. El certificado real llega de la misma manera. Abra `https://captions.mytheater.org/` desde un teléfono en el WiFi del local: sin advertencias. (Si el puerto 443 estaba ocupado, el programa usó el alternativo, `:8443`; entonces la dirección necesita el puerto, `https://captions.mytheater.org:8443/`, y el registro muestra la dirección exacta.)

**8. Dé la dirección a la gente.** El paso Dirigir de la configuración lista las direcciones numéricas y el nombre `.local`, pero todavía no esta. Para imprimir un cartel con código QR de esta, con la sesión iniciada en la configuración abra
`http://localhost/admin?print=qr&url=https%3A%2F%2Fcaptions.mytheater.org%2F` (cambie el nombre; añada el puerto si hace falta) e imprima esa página.

## Después

- **La renovación es automática.** El programa comprueba unas dos veces al día y renueva cuando faltan 30 días o menos. No tiene que hacer nada.
- **Si se cae Internet en el local,** no se rompe nada: el programa sigue sirviendo el certificado que tiene y lo intenta de nuevo una hora después. Antes de que llegue el primer certificado (o si nunca llega) los teléfonos reciben el autofirmado.
- **Renovar ahora** (Configuración → Ajustes → Seguridad → Certificado) obtiene un certificado nuevo de inmediato. Úselo después de resolver un problema.
- **Cambiar el token de DNS:** edite `/etc/open-caption/dns.env`. Se vuelve a leer en el siguiente intento, así que no hace falta reiniciar.
- **Desactivarlo:** ponga `acme: enabled: no` y reinicie. HTTPS usará entonces el certificado autofirmado (o sus propios archivos).

## Si algo sale mal

La configuración muestra **Último problema**; el registro tiene el mismo texto. Los más habituales:

| Ve | Causa probable y solución |
|---|---|
| `the DNS provider "…" is not built into this program` (el programa se detiene al arrancar) | Esta copia del programa se compiló sin ese proveedor. Use `cloudflare`, `godaddy` o `easydns`, o consiga una compilación con el suyo (más abajo) |
| `… credentials information are missing: EASYDNS_TOKEN,EASYDNS_KEY` | Los nombres del archivo son incorrectos o no se puede leer. Los nombres van en mayúsculas, `NOMBRE=valor`, uno por línea |
| Un error del proveedor de DNS como *unauthorized*, *forbidden* o un número 4xx | El token o la clave son incorrectos, han caducado o no tienen permiso para editar esa zona. Cree otros nuevos |
| *Esperando el primer certificado* durante mucho tiempo, o un mensaje de que no se encuentra el registro | El nuevo registro DNS tarda en propagarse. Añada `propagation_wait: 3m` bajo `dns:` y pulse **Renovar ahora** |
| `too many certificates already issued` o `rate limit` | Let's Encrypt limita las repeticiones. Espere (el mensaje dice cuánto) y use `letsencrypt-staging` mientras prueba |
| El navegador sigue avisando | Está usando el servicio de pruebas (cambie a `letsencrypt`), o abrió la dirección numérica u otro nombre en lugar de `https://captions…`, o el reloj del equipo o del teléfono es incorrecto |
| Un teléfono no puede abrir el nombre | Algunas redes WiFi de invitados y routers bloquean los nombres públicos que responden con direcciones privadas («protección contra DNS rebinding»). Pida al responsable de la red que permita `captions.mytheater.org`, o use la dirección numérica y el código QR como antes |
| `https.acme: cannot be used together with cert_file/key_file` | Ha puesto ambos. Quite uno: certificados automáticos **o** sus propios archivos |
| El programa se niega a arrancar y lista problemas en `config.yaml` | Cada línea nombra el ajuste y lo que está mal. Corríjalos todos y vuelva a arrancar |

## Credenciales más seguras (opcional)

Un token de DNS que puede cambiar toda su zona es un riesgo si el equipo se pierde o se comparte. Dos formas de reducirlo:

- Mantenga los secretos **fuera de `config.yaml`**: en el archivo de credenciales (paso 3), o como variables de entorno del servicio que inicia el programa (`EASYDNS_TOKEN_FILE=/path/to/file` también funciona: el ajuste se lee de ese archivo).
- Use la **delegación por CNAME**: en su DNS añada un registro permanente `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.other-zone.example`, donde `other-zone.example` es una zona pequeña y por lo demás vacía. El programa lo sigue automáticamente, así que el token solo necesita acceso a esa zona pequeña, no a la principal.

Las credenciales nunca se escriben en el registro, en el paquete de soporte ni en la configuración.

## Otro proveedor de DNS

El programa solo contiene los proveedores de DNS con los que se compiló. Quien lo compila puede elegir otros, o todos, con `make build ACME_PROVIDERS="…"` (los nombres son los de la [lista de lego](https://go-acme.github.io/lego/dns/)). Compilar necesita el código fuente y conexión a Internet (las guías de compilación para macOS, Windows y Linux están en el repositorio del código; en Windows use `build.bat -AcmeProviders "…"`). El programa crece con cada proveedor: unos 10,9 MB sin certificados automáticos, 12,6 MB con los tres predeterminados (Cloudflare, GoDaddy, EasyDNS), 15,0 MB con EasyDNS y Route 53, y 60,9 MB con todos los proveedores. Si su proveedor no es compatible en absoluto, `exec` (ejecuta un script suyo) y `httpreq` (llama a una dirección web) pueden manejar casi cualquier servicio de DNS.
