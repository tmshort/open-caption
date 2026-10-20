# Referencia de `config.yaml`

`config.yaml` contiene todo lo relativo a *este equipo y este local*: las contraseñas, los puertos, HTTPS, quién puede abrir la configuración y los valores por defecto que usa un espectáculo cuando su propio archivo no dice nada. Todo lo relativo a *un espectáculo* (título, colores, duración de los bucles, mensaje previo…) vive en cambio en el archivo de ese espectáculo (véase [Escribir un archivo del espectáculo](../../../web/help/admin-writing-a-show-file.md), «Ajustes del espectáculo»); las capas se explican más abajo en «Dónde se decide un ajuste».

## El archivo

- **Dónde está.** `<carpeta de datos>/config.yaml`; la carpeta de datos es `data_dir`, `--data`, o la predeterminada `<su carpeta de configuración de usuario>/open-caption` (en macOS `~/Library/Application Support/open-caption`, en Windows `%AppData%\open-caption`, en Linux `~/.config/open-caption`). Las primeras líneas que imprime el programa dicen qué archivo usa. `--config` indica otro archivo.
- **Primer arranque.** Si no hay archivo, el programa escribe uno con contraseñas aleatorias de **administrador** y de **operador** y las imprime una sola vez en la consola. El archivo tiene modo 0600 (privado para su cuenta) y las contraseñas están en **texto sin cifrar**, así que manténgalo privado.
- **Solo necesita las claves que cambie.** Una clave ausente conserva su valor por defecto (las tablas de abajo). El archivo generado lista todo.
- **Estricto.** Una clave mal escrita, un valor de tipo equivocado o fuera de rango detiene el programa al arrancar y muestra **todos** los problemas, cada uno con el nombre del ajuste (por ejemplo `audience.max_clients: 0 is not a client count: use 1 to 100000`). No se adivina nada. La única excepción: en un archivo marcado con un `config_version` más nuevo que el que conoce el programa se ignoran las claves desconocidas, con una advertencia.
- **La configuración edita una parte.** Cambiar las contraseñas, el nombre `.local`, el mensaje de espera o el nivel del registro en la configuración reescribe esas claves en el archivo y conserva sus comentarios y todos los demás ajustes. Todo lo demás se edita a mano y después hay que **reiniciar el programa**. (Un archivo de certificado renovado es lo único que se recoge sin reiniciar; véase `https.cert_file`.)
- **Los booleanos** son `yes`/`no`/`true`/`false`. **Las listas** se escriben `[a, b]` o con un `- elemento` por línea.

## Línea de comandos y entorno

Una opción, o su variable de entorno, prevalece sobre el archivo en esa ejecución. (El archivo no cambia.) Una opción puede escribirse con uno o dos guiones (`-data` o `--data`). `open-caption --help` imprime esta lista, y la **página de manual** (`docs/open-caption.1`; léala con `man ./docs/open-caption.1`, o cópiela a una carpeta `man1` como `/usr/local/share/man/man1/`) describe las mismas opciones, archivos, señales y estado de salida.

| Opción | Entorno | Efecto |
|---|---|---|
| `--config FILE` | `OPEN_CAPTION_CONFIG` | Usa este archivo de ajustes (por defecto `<carpeta de datos>/config.yaml`) |
| `--data FOLDER` | `OPEN_CAPTION_DATA` | La carpeta de datos (espectáculos, sesiones, `live.json`, certificados). Prevalece sobre `data_dir` |
| `--http-addr ADDR` | `OPEN_CAPTION_HTTP_ADDR` | Prevalece sobre `http.addr`, por ejemplo `:80` o `:8080` |
| `--https-addr ADDR` | `OPEN_CAPTION_HTTPS_ADDR` | **Activa** HTTPS en esta dirección (prevalece sobre `https.enabled` y `https.addr`) |
| `--log-level LEVEL` | `OPEN_CAPTION_LOG_LEVEL` | Prevalece sobre `log.level` cada vez que arranca el programa (la configuración lo indica cuando está en vigor) |
| `--log-file FILE` | `OPEN_CAPTION_LOG` | Añade además el registro a este archivo |
| `--network-interfaces LIST` | `OPEN_CAPTION_NETWORK_INTERFACES` | Prevalece sobre `network.interfaces` cada vez que arranca el programa: una lista separada por comas de nombres de interfaz, como `Ethernet` o `en0,en4` (la configuración lo indica cuando está en vigor) |
| `--network-restrict` | `OPEN_CAPTION_NETWORK_RESTRICT` | Además rechaza las conexiones hechas a direcciones que no se ofrecen, cada vez que arranca el programa (como `network.restrict_listening: yes`). La variable cuenta cuando vale `yes`, `true`, `on` o `1` |
| `--no-browser` | | No abre la configuración en un navegador al arrancar (como `defaults.open_browser: no`) |
| `--version` | | Imprime la versión y los proveedores de DNS incluidos para los certificados automáticos, y sale |
| `--help` | | Imprime las opciones y las variables de entorno, y dónde leer más, y sale |
| `--check-show FILE [TRANSLATION ...]` | | **Comprueba un archivo del espectáculo** e imprime un informe, como un «lint», y sale: sin servidor, sin `config.yaml`. Véase «Comprobar un archivo del espectáculo» más abajo |
| `--assets FOLDER` | | La carpeta de imágenes para `--check-show` (por defecto: la carpeta `assets` junto al archivo del espectáculo) |
| `--strict` | | Con `--check-show`: las advertencias también hacen que el estado de salida sea 1 |
| `--dev`, `--show-file` | | **Solo para desarrolladores**: puntos de prueba sin autenticación en este equipo, sin espera; no los use nunca en un local |

## Comprobar un archivo del espectáculo

`open-caption --check-show show.md` analiza un archivo del espectáculo sin cargarlo en el programa: ejecuta el mismo analizador y las mismas comprobaciones que la lista de problemas del paso Archivo del espectáculo de la configuración e imprime lo que mostraría esa lista.

```
$ open-caption --check-show show.md show.es.md
show.md: info: 4 captions, 1 images, 0 fullscreen, 1 loops, 1 sections, 0 notes.
show.md:8: warning: Image "curtain.jpg" not found in assets.
show.md:13: error: [song] was opened but never closed with [/song]. The lyrics so far were kept.
show.md: 8 cues; 2 speakers: ALICE, BOB
show.es.md: 8 cues; language es

images checked against assets (1 found)
missing images: curtain.jpg (line 8 of show.md)
1 error, 1 warning, 1 note
```

- Cada problema es `file:line: severity: message` (gravedad `error`, `warning` o `note`), la forma que entienden los editores y los sistemas de integración continua. Una línea sobre todo el archivo no lleva número de línea.
- Los archivos tras el primero son **traducciones**: cada uno se analiza como traducción y se compara con el archivo principal (el mismo número y tipo de señales, en el mismo orden), como hace la configuración. El idioma es la línea `[language]` del archivo o su nombre (`show.es.md`); dos archivos para un mismo idioma son un error.
- Las imágenes se comprueban en la carpeta `assets` junto al archivo principal, o en `--assets FOLDER`. Sin carpeta no se comprueban, y el informe lo dice. Una imagen que falta nombra la carpeta en la que se buscó (el editor de la configuración dice en cambio que la suba en el paso Imágenes).
- `-` como nombre de archivo lee la entrada estándar.
- **Estado de salida:** `0` sin errores, `1` errores (o advertencias con `--strict`), `2` no se pudo leer un archivo. Ponga las demás opciones (`--assets`, `--strict`) **antes** de `--check-show`: los archivos que van después son traducciones.

## Ajustes

Los valores por defecto son los que obtiene sin la clave. «Configuración» indica dónde puede cambiarlo la configuración.

### Nivel superior

| Clave | Por defecto | Significado |
|---|---|---|
| `config_version` | `1` | Lo escribe el programa. Un número mayor que el que conoce el programa significa que el archivo viene de una versión más nueva |
| `admin_password` | generada | La contraseña de la configuración (`/admin`). No puede estar vacía (6 o más caracteres al cambiarla en la configuración). *Configuración: Ajustes → Seguridad* |
| `operator_password` | generada | La contraseña de la página del operador (`/operator`); la del administrador también sirve ahí. No puede estar vacía. *Configuración: Ajustes → Seguridad* |
| `data_dir` | vacío | La carpeta de datos. Vacío significa la carpeta por defecto de arriba. No puede ser un archivo |

### `http`

| Clave | Por defecto | Significado |
|---|---|---|
| `http.addr` | `:80` | Dónde escucha el servidor HTTP sin cifrar: `:80` (todas las direcciones, puerto 80), `192.168.1.20:80` o `127.0.0.1:8080` (solo este equipo). No puede estar vacío |
| `http.fallback_addr` | `:8080` | Se usa, con una advertencia visible, si no se puede usar `addr` (el puerto 80 suele estar ocupado o requerir privilegios). Vacío significa sin alternativa |

### `https`

HTTPS está **desactivado** por defecto. Hay tres formas de obtener un certificado; elija una.

| Clave | Por defecto | Significado |
|---|---|---|
| `https.enabled` | `no` | Sirve HTTPS además de HTTP, en puertos distintos (no hay redirección). Necesita una de las tres de abajo |
| `https.addr` | `:443` | El puerto HTTPS. Obligatorio si está activado |
| `https.fallback_addr` | `:8443` | Se usa si no se puede abrir `addr` |
| `https.self_signed` | `yes` | Sin archivos de certificado y sin `acme`: el programa crea su propio certificado (guardado en `<carpeta de datos>/tls/`). Los navegadores avisan de él hasta que se acepta. `no` sin archivos es un error |
| `https.cert_file`, `https.key_file` | vacío | Su propio certificado y clave (PEM). Indique **ambos** o ninguno; los dos archivos deben existir. **Los archivos se vigilan**: cuando se sustituyen (por ejemplo con `certbot`), el nuevo par se usa para las conexiones nuevas en unos 30 segundos, sin reiniciar. Un par que no se puede cargar conserva el certificado anterior y registra una advertencia |

### `https.acme` (certificados automáticos)

Desactivado salvo `enabled: yes`. Necesita `https.enabled: yes`, un nombre de dominio público propio y un proveedor de DNS con el que se haya compilado este programa. No se puede combinar con `cert_file`/`key_file`. Cómo configurarlo, paso a paso: la [guía del certificado HTTPS de confianza](https-certificate-guide.md); las opciones de proveedores de DNS al compilar, y los tamaños, están al final de esa guía.

| Clave | Por defecto | Significado |
|---|---|---|
| `https.acme.enabled` | `no` | Obtiene y renueva un certificado automáticamente (ACME, DNS-01) |
| `https.acme.domain` | | El nombre público, por ejemplo `captions.example.org`. Un solo nombre: sin comodines, sin `.local`/`.lan`/`.internal`, sin dirección IP. Debe resolverse, para los teléfonos del público, a la dirección de este equipo |
| `https.acme.email` | | Adónde envía la autoridad de certificación los avisos de caducidad |
| `https.acme.agree_to_terms` | `no` | **Debe ser `yes`**: registra que acepta las condiciones de servicio de la autoridad de certificación |
| `https.acme.directory` | `letsencrypt` | `letsencrypt`, `letsencrypt-staging` (para probar: los navegadores no confían en él) o la dirección `https://` de otro directorio ACME |
| `https.acme.renew_before_days` | `30` | Renueva cuando faltan estos días o menos (1 a 60) |
| `https.acme.dns.provider` | | El proveedor de DNS, por su [nombre en lego](https://go-acme.github.io/lego/dns/) (`cloudflare`, `godaddy`, `easydns`, …). Debe estar incluido en el programa (`--version` los lista) |
| `https.acme.dns.credentials` | vacío | Los ajustes propios del proveedor, como un mapa: para EasyDNS `EASYDNS_TOKEN: …` y `EASYDNS_KEY: …`. Los nombres van en mayúsculas, dígitos y guiones bajos. El entorno y `credentials_file` prevalecen sobre estos |
| `https.acme.dns.credentials_file` | vacío | Un archivo de líneas `NAME=value` con los mismos ajustes, que se vuelve a leer en cada intento. Debe existir; hágalo privado con `chmod 600` |
| `https.acme.dns.propagation_wait` | vacío | Una espera fija (`90s`, `2m`, como máximo `30m`) para que se propague el nuevo registro DNS, en lugar de comprobarlo |

### `defaults`

Lo que usa un espectáculo cuando su propio archivo no lo dice. El bloque `[settings]` del archivo del espectáculo prevalece sobre todos los valores de aquí (véase [Escribir un archivo del espectáculo](../../../web/help/admin-writing-a-show-file.md)). El paso Aspecto de la configuración escribe el archivo del espectáculo, no estos.

| Clave | Por defecto | Significado |
|---|---|---|
| `defaults.language` | `en` | El idioma principal de un espectáculo cuyo archivo no lo indica: dos o tres letras, con región opcional (`en`, `es`, `pt-br`) |
| `defaults.loop_item_seconds` | `10` | Cuánto se queda cada elemento de un bucle previo o de descanso, en segundos (1 a 3600), salvo que el elemento diga otra cosa |
| `defaults.idle_message` | `Captions will begin shortly` | Lo que lee el público antes de enviar la primera línea cuando un espectáculo no tiene mensaje previo (hasta 200 caracteres) |
| `defaults.open_browser` | `yes` | Abre la configuración en un navegador al arrancar el programa (no es un ajuste del espectáculo; `--no-browser` lo desactiva en una ejecución) |
| `defaults.look.background` | `#000000` | El fondo de los subtítulos: `#rgb`, `#rrggbb`, o `black`, `white`, `yellow`, `red`, `green`, `blue`, `cyan`, `magenta`, `orange`, `gray` |
| `defaults.look.text_color` | `#ffffff` | El color del texto de los subtítulos (mismas formas) |
| `defaults.look.speaker_name_color` | `#ffffff` | El color de los nombres de los personajes (mismas formas) |
| `defaults.look.font` | `system` | `system`, `serif`, `mono`, `dyslexic` o `hyperlegible` |
| `defaults.look.text_size_percent` | `100` | Tamaño del texto de los subtítulos, de 50 a 300 |
| `defaults.look.speaker_names` | `label` | `label` (nombres en un solo color), `color` (un color distinto para cada personaje) o `hidden` |

### `audience`

| Clave | Por defecto | Significado |
|---|---|---|
| `audience.max_clients` | `500` | El máximo de dispositivos del público conectados **por idioma** (1 a 100000); los demás se rechazan hasta que se vayan algunos |
| `audience.scrollback_lines` | `200` | Cuántas líneas anteriores puede recorrer hacia atrás un dispositivo (20 a 2000) |

### `access`

Quién puede abrir la configuración y la página del operador. La página del público está siempre abierta a todos.

| Clave | Por defecto | Significado |
|---|---|---|
| `access.admin_allow` | `[]` | Equipos que pueden abrir la configuración. Vacío significa que **cualquier equipo de la red puede probar la contraseña**. Entradas: `localhost`, `private` (las redes domésticas y de oficina habituales), una dirección (`192.168.1.50`) o una red (`192.168.1.0/24`). Ejemplo: `[localhost]` |
| `access.operator_allow` | `[]` | Lo mismo, para la página del operador |
| `access.local_auto_login` | `yes` | Un navegador en el equipo que ejecuta el programa (abierto en `localhost`, `127.0.0.1` o `[::1]`) entra como administrador sin contraseña. Los demás equipos usan siempre contraseña. Desactívelo en un equipo en el que puedan iniciar sesión otras personas |

### `log`, `mdns`, `standby`

| Clave | Por defecto | Significado |
|---|---|---|
| `log.level` | `info` | `debug` (cada petición), `info` (eventos e inicios de sesión), `warn` o `error`. *Configuración: Ajustes → Acerca de* |
| `mdns.enabled` | `yes` | Anuncia este equipo como `<nombre>.local` en la red local |
| `mdns.name` | `open-caption` | El nombre: solo letras, dígitos y guiones, hasta 63 caracteres (se añade `.local`). *Configuración: Ajustes → Red* |
| `standby.message` | vacío | La línea que ve el público en la pantalla de espera. Vacío significa «Captions will begin shortly» en el idioma de cada espectador. Una línea, hasta 200 caracteres. *Configuración: Ajustes → Pantalla del público* (el logotipo es un archivo que se elige allí, no un ajuste) |

### `network`

| Clave | Valor por defecto | Significado |
|---|---|---|
| `network.interfaces` | `[]` | Las interfaces de red que se usan, por nombre (`Ethernet`, `Wi-Fi`, `en0`, `eth0`): solo sus direcciones se dan al público, se muestran en la configuración, entran en un certificado HTTPS autofirmado nuevo y se anuncian como `<nombre>.local`. Vacío significa que aún no se ha elegido: se usan todas las interfaces. Una vez hecha la elección, una interfaz que aparezca después (un adaptador USB, una VPN) queda desactivada hasta que se añada. Cuando dos interfaces en uso están en la misma red (un cable y Wi-Fi a la vez) se ofrece una sola dirección: la de la conexión principal y, si no, la cableada; deje fuera la otra para elegir. *Configuración: Ajustes → Red* |
| `network.restrict_listening` | `no` | También **rechaza las conexiones** hechas a una dirección de una interfaz que no se usa (un teléfono que la use es rechazado), en lugar de solo dejar la dirección fuera de las listas. Las conexiones al propio equipo (`localhost`) siempre se atienden, así que se puede abrir la configuración para cambiarlo. La configuración nunca desactiva la interfaz a través de la que la está usando. *Configuración: Ajustes → Red* |

Las direcciones de una interfaz que no se usa no aparecen en el paso Dirigir ni al arrancar, no se dan al público, no se anuncian como `<nombre>.local` y no se incluyen en un certificado HTTPS autofirmado nuevo. Solo se muestran las interfaces conectadas, y se ofrece una sola dirección por red. Por defecto esto solo elige lo que se *ofrece*: el servidor sigue aceptando conexiones en todas las direcciones, hasta que se active `network.restrict_listening`.

## Dónde se decide un ajuste

De menor a mayor prioridad: el valor integrado del programa → `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) → el `[settings]` del archivo del espectáculo. Una opción de línea de comandos o una variable de entorno prevalece sobre el archivo. Los ajustes propios del local (puertos, HTTPS, contraseñas, `access`, `data_dir`, `audience.max_clients`, `mdns`, `network`, `standby`, `log`) existen solo aquí.

## Un ejemplo completo

```yaml
config_version: 1
admin_password: "set-a-long-one"
operator_password: "another-one"
http:
  addr: ":80"
  fallback_addr: ":8080"
https:
  enabled: yes
  addr: ":443"
  fallback_addr: ":8443"
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org
    agree_to_terms: yes
    directory: letsencrypt
    dns:
      provider: easydns
      credentials_file: /etc/open-caption/dns.env
mdns:
  enabled: yes
  name: open-caption
network:
  interfaces: []
  restrict_listening: no
standby:
  message: "Welcome to the Little Theatre"
defaults:
  language: en
  loop_item_seconds: 10
  idle_message: "Captions will begin shortly"
  look:
    font: hyperlegible
    text_size_percent: 120
audience:
  max_clients: 500
  scrollback_lines: 200
access:
  admin_allow: [localhost]
  operator_allow: [private]
  local_auto_login: yes
log:
  level: info
```
