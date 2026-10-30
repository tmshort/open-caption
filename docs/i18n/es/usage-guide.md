# Guía de uso de open-caption

open-caption muestra subtítulos en directo en los teléfonos, tabletas y pantallas del público durante una obra de teatro o un musical. Una persona voluntaria, el **operador**, sigue el espectáculo y pulsa la barra espaciadora en cada línea; todos los dispositivos de la WiFi del local muestran la línea un instante después. No se instala nada en los dispositivos del público, y nadie tiene que tocar nada.

Esta guía explica qué comprar y cómo organizarlo ([Despliegue](#recomendaciones-de-despliegue)), cómo preparar un espectáculo, cómo dirigirlo y una lista de comprobación para el día de la función. El botón **Ayuda** del programa cubre lo mismo en piezas más pequeñas, y [venue-network-guide.md](venue-network-guide.md) está escrita para quien gestiona la WiFi del edificio.

Contenido: [Las piezas](#las-piezas) · [Recomendaciones de despliegue](#recomendaciones-de-despliegue) · [Instalación y arranque](#instalacion-y-arranque) · [Preparar un espectáculo](#preparar-un-espectaculo) · [Dirigir el espectáculo](#dirigir-el-espectaculo) · [Lista de comprobación del día](#lista-de-comprobacion-del-dia) · [Resolución de problemas](#resolucion-de-problemas) · [Mantenimiento](#mantenimiento) · [Tarjeta rápida del operador](#tarjeta-rapida-del-operador)

## Las piezas

| Pieza | Qué es |
|-------|--------|
| **El servidor** | El programa `open-caption`, funcionando en un equipo (macOS, Windows o Linux). Guarda el archivo del espectáculo y envía las líneas a todos. |
| **Página de Configuración** (`/admin`) | Donde las personas voluntarias preparan un espectáculo: archivo del espectáculo, imágenes, aspecto, ensayo. Contraseña de administrador. |
| **Página del operador** (`/operator`) | Donde el operador sigue el espectáculo y envía las líneas. Contraseña de operador. El paso **Compartir** de la configuración y su enlace del menú la abren. |
| **Página de seguimiento** (`/follow`) | Un segundo dispositivo, como la tableta del director de escena, que solo mira el espectáculo: ve las líneas, las notas y la posición y no puede enviar nada. Contraseña de espectador opcional; las contraseñas de operador y de administrador también la abren. El paso **Compartir** de la configuración la muestra. |
| **Página del público** (`/`) | Lo que el público abre en sus dispositivos. Sin contraseña. Un navegador en un proyector o un monitor grande puede mostrar la misma página para toda la sala. |

Todo funciona en un solo equipo. El público, el operador y las personas de configuración usan navegadores web corrientes.

## Recomendaciones de despliegue

### La versión corta

1. **Use un punto de acceso WiFi dedicado a los subtítulos** a menos que se sepa que la red de invitados del local funciona (véase más abajo). Elimina el mayor riesgo: una red que usted no controla.
2. **Ponga el servidor en una conexión por cable** a ese punto de acceso, con una dirección reservada.
3. **Use un punto de acceso hecho para multitudes** (preparado para más de 100 dispositivos) colocado en la zona del público.
4. **Mantenga el equipo servidor enchufado, despierto y sin hacer nada más** durante el espectáculo.
5. **Imprima el código QR** y póngalo en el programa de mano y en el vestíbulo.
6. **Ensaye con unos cuantos teléfonos reales** en la sala real antes del estreno.

### Opciones de red

| Opción | Cuándo usarla | Notas |
|--------|---------------|-------|
| **A. Punto de acceso dedicado (recomendado)** | Cualquier local, sobre todo si no puede cambiar la red del edificio | Usted trae un router o punto de acceso. El público se conecta a su WiFi (por ejemplo «Theater Captions»). No hace falta conexión a internet. Usted controla la capacidad, las reservas y los cortafuegos. |
| **B. La WiFi de invitados del local** | El personal de red del local confirma que los dispositivos de invitados pueden llegar a su equipo | La más barata. Pida una dirección reservada para su equipo y que se permita el tráfico de invitados al servidor. Consulte [venue-network-guide.md](venue-network-guide.md). |
| **C. El punto de acceso del propio portátil** | Solo ensayos y públicos minúsculos | Los puntos de acceso de los portátiles suelen aguantar solo un puñado de dispositivos. No es para una sala llena. |

### La opción A en detalle (punto de acceso dedicado)

```
                 ┌───────────────────────┐
                 │  Access point/router  │  WiFi "Theater Captions"
                 │  (in the audience     │ ─ ─ ─ ─ ─ ─ ─ ─  phones / tablets
                 │   area, 5 GHz on)     │                  projector browser
                 └───────────┬───────────┘
                    Ethernet │
                 ┌───────────┴───────────┐
                 │ Captioning computer   │   ← operator uses this screen,
                 │  (open-caption)       │     or a second device on the same WiFi
                 └───────────────────────┘
```

- **Equipo:** un punto de acceso o router profesional o semiprofesional diseñado para muchos dispositivos simultáneos (busque más de 100 clientes en sus especificaciones y compatibilidad con 5 GHz), con un puerto Ethernet. Los routers domésticos baratos suelen sufrir a partir de 30–50 dispositivos. Para una sala de unas 150 butacas, cuente con un punto de acceso capaz, y un segundo para una sala profunda o ancha.
- **Ajustes:**
  - Cree un SSID con un nombre corto y una clave WPA2 sencilla e imprimible, o sin clave si el local permite una red abierta. Imprímalo en el cartel QR.
  - Active los **5 GHz**, y prefiéralos. Use anchos de canal de 20 o 40 MHz en un edificio saturado.
  - Deje activado el **aislamiento de clientes** si se ofrece. Los dispositivos del público solo hablan con el servidor, nunca entre sí.
  - Dé al equipo de subtitulado una **reserva DHCP** (o una dirección fija) para que la dirección nunca cambie, e imprima la dirección en el cartel.
  - No necesita conexión a internet, y el programa nunca usa una.
- **Colocación:** ponga el punto de acceso en la zona del público, en alto y despejado de paredes, no en un cuarto trasero ni en el vestíbulo. No lo ponga detrás del escenario ni del rack de iluminación. Un portátil en una mesa lejana está bien: va por cable.
- **Alimentación:** use un protector contra sobretensiones. Un pequeño SAI en el punto de acceso y el servidor elimina el riesgo de un enchufe desconectado.

### La opción B en detalle (la WiFi de invitados del local)

Al programa no le importa en qué red esté, mientras los dispositivos del público puedan llegar a él. Antes de confiar en ella:

- Pida **acceso de invitados al servidor**: su equipo debe ser accesible desde el SSID de invitados, por el puerto TCP 80 (y 443 si activa HTTPS). El «aislamiento de clientes» entre invitados está bien; impedir que los invitados lleguen a *su equipo* no.
- Pida una **dirección reservada** para el adaptador de red de su equipo.
- Conecte el equipo a la red por **Ethernet** si es en absoluto posible.
- Pruebe con un teléfono en la WiFi de invitados mucho antes del espectáculo (el paso Compartir muestra la dirección).
- Lleve la opción A en una bolsa como respaldo si el espectáculo importa.

### El equipo de subtitulado

- **Cualquier portátil o minipc corriente** sirve. El programa usa muy poca memoria (decenas de megabytes) y casi nada de procesador. Use un equipo que pueda mantener enchufado.
- **Sistema operativo:** macOS, Windows o Linux. Use el mismo equipo para todos los espectáculos si puede, para que las contraseñas, los espectáculos y la dirección sigan siendo los mismos.
- **Hágalo aburrido:** configure el equipo para que *nunca se suspenda* mientras está enchufado, desactive los reinicios automáticos por actualizaciones en torno a las fechas de función, cierre otros programas y mantenga la tapa abierta o configurada para no suspender.
- **Cortafuegos:** permita el programa en **redes privadas** (Windows lo pregunta la primera vez; en macOS permita las conexiones entrantes para él).
- **Dirección:** por cable y reservada (véase arriba). Si la dirección cambia, los códigos QR impresos dejan de funcionar.
- **Hora:** el reloj no importa.

### Quién se sienta dónde

- **El operador** necesita una buena vista del escenario y un dispositivo fiable. Lo más sencillo es el propio equipo de subtitulado en la mesa técnica. La página del operador está pensada para un equipo con teclado (Espacio envía la línea siguiente). Use el propio equipo de subtitulado u otro portátil de la misma red (`http://<dirección del equipo>/operator`) para que el operador se siente donde vea el escenario. Un teléfono o una tableta mostrarán la página, pero **no se recomiendan**: en una pantalla táctil es fácil pulsar mal durante la función. Un pulsador Bluetooth o un teclado es una buena alternativa.
- **Solo un operador controla el espectáculo a la vez.** Cualquier otra persona que abra la página del operador puede mirar, y puede usar **Asumir el control** si el operador se lo cede.
- **Las personas de configuración** usan la página de Configuración desde cualquier equipo de la red antes del espectáculo, o desde el propio equipo de subtitulado.
- **Una pantalla compartida** para toda la sala: abra la dirección del público en un equipo conectado a un proyector o a un monitor grande, añada `?display=projector` a la dirección para tener texto grande y nada de ruido en pantalla, y ponga el navegador a pantalla completa (`F11`, o ejecute Chrome con `--kiosk`).

### Ejecutar el servidor

Para la mayoría de teatros, lo más sencillo es lo mejor: arranque el programa a mano antes de cada espectáculo, deje abierta su ventana y pulse `Ctrl+C` (o ciérrela) al terminar. La posición del espectáculo se guarda continuamente, así que reiniciar en mitad de un espectáculo es seguro (véase [Resolución de problemas](#resolucion-de-problemas)).

Si el equipo debe arrancar el servidor solo tras un reinicio, la carpeta [`deploy/`](../../../deploy/) tiene archivos listos para Linux (systemd), macOS (launchd) y Windows (Programador de tareas). Cada uno empieza con sus propios pasos de instalación.

### Llevar a la gente a la página

En orden de fiabilidad:

1. **El código QR y el cartel** (Configuración → Compartir → *Imprimir un cartel con código QR para el vestíbulo*). Póngalo en la pared, en el programa de mano y en la tarjeta del respaldo de las butacas. Esto es en lo que hay que confiar.
2. **Una dirección corta.** En el puerto 80 no hay que teclear `:puerto`. El programa también se anuncia como **`open-caption.local`** (cámbielo en Configuración → Ajustes → Red → *Usar un nombre*): donde la red de invitados lo permite, `http://open-caption.local/` funciona en la mayoría de teléfonos sin configurar nada. Algunas redes de invitados bloquean ese tipo de anuncio, así que no imprima solo el nombre; el cartel QR usa la dirección numérica. Un nombre en el DNS propio del router (como `captions.theater`) funciona en todas partes pero exige que el local lo añada.
3. **La página de bienvenida del punto de acceso** (un «portal de invitados» o «portal cautivo» que redirige a su dirección tras conectarse). Se configura en el router o punto de acceso, no en este programa. Consulte [venue-network-guide.md](venue-network-guide.md).
4. **Opción 114 de DHCP y `/captive-api` (avanzado, experimental).** La opción 114 es un ajuste de DHCP que apunta a los dispositivos a una «API de portal cautivo». El programa responde a esa API en `/captive-api` (por ejemplo `https://<dirección>/captive-api`) con «esta red no es cautiva» y la dirección de la página. Dos advertencias: **no hace que los teléfonos abran solos la página de subtítulos**, y el estándar (RFC 8910/8908) exige **HTTPS con un certificado en el que confíen los dispositivos**, así que un ajuste con `http://` sencillo o el certificado autofirmado del programa se ignora. Configúrelo solo si el local ya usa la opción 114 y puede proporcionar un certificado de confianza. No se ha probado en dispositivos reales.

### Puertos y direcciones

- Por defecto el servidor escucha en el **puerto 80**, así que la dirección del público no necesita número de puerto, por ejemplo `http://192.168.1.10/`. Si algo más usa el puerto 80, o el sistema operativo no lo permite, el servidor se lo dice y usa el **puerto 8080**; entonces la dirección del público termina en `:8080`. En Linux, el puerto 80 necesita permiso (`sudo setcap 'cap_net_bind_service=+ep' ./open-caption`) salvo que use el archivo de systemd, que lo concede.
- Puede elegir los puertos usted mismo con `--http-addr :8081`, o en `config.yaml`.
- **HTTPS es opcional.** `http://` sencillo funciona en todos los dispositivos sin advertencias y es la dirección recomendada para el público. Una dirección HTTPS autofirmada hace que cada dispositivo muestre una advertencia de seguridad. Active HTTPS solo si puede instalar un certificado en el que confíen sus dispositivos. Si el local tiene un nombre de dominio propio, el programa puede obtener y renovar solo un certificado de confianza; véase la [guía del certificado HTTPS de confianza](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

### Seguridad

- La página del público solo puede leer. No puede cambiar el espectáculo ni llegar a las páginas de configuración y del operador.
- Configuración y operador exigen cada una su propia contraseña (guardada en `config.yaml` solo como resumen, un hash de un solo sentido, de modo que no se puede leer en él; el archivo además se crea legible solo para usted). Una contraseña de **espectador** opcional (en **Ajustes**) deja que un segundo dispositivo, como la tableta de un regidor, abra la página de seguimiento (`/follow`; las contraseñas de operador y de administrador también la abren) y mire el espectáculo sin poder enviar nada. Dé la contraseña de operador al operador, no la de administrador. Las contraseñas se pueden cambiar desde **Ajustes** en Configuración.
- En el equipo que ejecuta el programa, Configuración y la página del operador se abren **sin contraseña** cuando la dirección empieza por `localhost`, `127.0.0.1` o `[::1]` (se confía en la persona que está en ese equipo como administrador). Los demás equipos, y cualquier petición retransmitida por un proxy, siguen entrando con contraseña. Ponga `access.local_auto_login: false` en `config.yaml` si otras personas pueden iniciar sesión en este equipo.
- **Cualquiera que pueda llegar al servidor puede abrir las páginas de inicio de sesión de Configuración y del operador, y con `http://` sencillo las contraseñas viajan sin cifrar por la WiFi.** En una WiFi de invitados compartida eso significa que un espectador curioso podría capturar una contraseña y tomar el control de los subtítulos. Reduzca el riesgo, por este orden:
  1. **Limite qué equipos pueden abrirlas.** En `config.yaml`:
     ```yaml
     access:
       admin_allow: [localhost]            # Setup only on the captioning computer itself
       operator_allow: [192.168.1.50]      # the operator's computer (or a network such as 192.168.1.0/24)
     ```
     Las entradas pueden ser `localhost`, `private` (rangos de direcciones domésticas o de oficina), una sola dirección, o una red como `192.168.1.0/24`. Por defecto, `admin_allow` es `[localhost]` (la configuración solo se abre en el equipo de subtitulado); una lista vacía significa cualquier equipo, con una advertencia. `operator_allow` está vacío por defecto. Una dirección de `admin_allow` también puede usar la página del operador. En lugar de editar el archivo, puede cambiar todo esto en la Configuración, en **Ajustes → Acceso → Quién puede abrir las páginas**; el cambio se aplica al instante y la Configuración rechaza una lista que dejaría fuera al equipo que está usando. Si edita el archivo a mano, reinicie el programa; una entrada incorrecta detiene el programa con un mensaje en lugar de permitir a todos.
  2. **Ponga al operador en una red por cable o aparte** a la que los dispositivos del público no puedan unirse, o ponga una clave WPA2 a la red de subtítulos.
  3. **Use HTTPS** (`https.enabled` en `config.yaml`) si puede dar a los dispositivos un certificado en el que confíen; uno autofirmado hace que cada dispositivo muestre una advertencia, y una dirección de público sin cifrar es más simple.
- Un sitio web, u otro servicio del mismo equipo, no puede hacer que un navegador cambie el espectáculo a sus espaldas: se rechazan las peticiones que no vengan de las propias páginas del programa.
- El programa está pensado para una red privada. **No lo exponga a internet.**

## Instalación y arranque

1. **Consiga el programa.** Descargue la versión para su equipo de la página *Releases* del repositorio y descomprímala (o constrúyala usted: vea las guías de compilación en la carpeta `docs` del repositorio de código fuente). Póngala en una carpeta que vaya a conservar, por ejemplo `Documentos/open-caption`.
2. **Primer arranque.** Ejecútelo: haga doble clic en Windows, o en una terminal ejecute `./open-caption` (macOS/Linux). El programa no está firmado, así que la primera vez puede hacer falta un clic: en macOS *clic derecho → Abrir* (o **Abrir de todos modos** en Ajustes del Sistema → Privacidad y seguridad), en Windows *Más información → Ejecutar de todas formas* en la ventana de SmartScreen.
3. **Anote las contraseñas.** La primera vez, el programa imprime las contraseñas de **administrador** y de **operador** y dónde está `config.yaml`. Se quedan en ese archivo.
4. La página de configuración se abre en su navegador. Si no, abra `http://localhost/admin` (o `http://localhost:8080/admin` si la consola dice que usó el puerto 8080).
5. **Elija dónde viven los datos (opcional).** Por defecto, los espectáculos y los ajustes están en la carpeta de configuración de su usuario (macOS `~/Library/Application Support/open-caption`, Linux `~/.config/open-caption`, Windows `%AppData%\open-caption`). Para guardar todo en una carpeta que usted elija, arranque con `--data <carpeta>`. Use siempre la misma carpeta.

Opciones útiles: `--data <carpeta>`, `--http-addr :8080`, `--https-addr :443`, `--no-browser`, `--log-file <archivo>`, `--log-level debug|info|warn|error`, `--version`. Todas las opciones: `open-caption --help`, la página de manual `docs/open-caption.1` (solo en inglés) o `docs/configuration-reference.md`.

Si `config.yaml` tiene un error (un ajuste mal escrito, un número fuera de rango, una dirección incorrecta), el programa enumera todos los problemas con su línea o nombre de ajuste y se detiene; corrija el archivo y arranque de nuevo. (Un archivo marcado con un `config_version` más nuevo de lo que entiende el programa es la excepción: los ajustes desconocidos se ignoran con un aviso, de modo que puede volver a un programa más antiguo.)

## Preparar un espectáculo

Abra la página de configuración y siga los seis pasos de la parte superior (un séptimo, **Informe**, es para después del espectáculo). El botón **Ayuda** explica cada uno.

1. **Espectáculos.** Cree un espectáculo (o el de ejemplo, para probar).
2. **Archivo del espectáculo.** Pegue el archivo del espectáculo o suelte un archivo `.txt` o `.md`. Cada línea se convierte en un subtítulo. Corrija lo que aparezca en *Problemas y avisos*; pulse uno para saltar a su línea. Guarde.
3. **Imágenes.** Suba las imágenes que usa el archivo del espectáculo (PNG, JPG, GIF o WebP, idealmente de menos de 1 MB cada una).
4. **Aspecto.** Elija colores, fuente y tamaño del texto. Los valores por defecto (blanco sobre negro, grande) son los mejores para la mayoría. Hay dos fuentes de legibilidad integradas: una apta para dislexia (OpenDyslexic) y una de alta legibilidad (Atkinson Hyperlegible, diseñada para baja visión). Cada miembro del público también puede elegir cualquiera para su propio dispositivo con el pequeño control **Aa**.
5. **Prueba.** Haga un ensayo privado: usted pulsa Espacio como lo haría el operador, y ve la pantalla del público al lado. Nada llega al público.
6. **Compartir.** Revise la lista de preparación, imprima el cartel QR y pulse **Abrir la página Dirigir**.

**Trabajar en un espectáculo no es lo mismo que hacerlo el espectáculo activo.** *Trabajar en este espectáculo* solo elige qué espectáculo edita Configuración (su nombre está en la cabecera). El espectáculo que recibe el público, el **activo** (marcado ACTIVO en el paso Espectáculos y mostrado en la cabecera por la etiqueta Activo, ámbar en espera y roja en el aire), cambia solo cuando pulsa **Abrir la página Dirigir** en el paso Compartir de ese espectáculo. Un espectáculo que está en el aire solo se reemplaza después de que Configuración se lo pregunte.

Al abrir la página del operador, el espectáculo se carga **en espera**: el público ve la pantalla de espera del local (un logotipo y un mensaje, que se fijan una sola vez en **Ajustes → Pantalla del público**) y nada del espectáculo, así que puede preparar, ensayar y revisar la página del operador con tranquilidad. Cuando esté listo, pulse **Espacio** (el botón dice **INICIAR**) en la página del operador para poner el espectáculo en el aire. **Poner en espera…** lo retira de nuevo sin perder su lugar.

### Escribir el archivo del espectáculo

Quien prepara el archivo del espectáculo no necesita leer el resto de esta guía: [Preparar un archivo del espectáculo](../../../web/help/admin-preparing-a-show-file.md) cubre el trabajo desde el guion hasta el archivo terminado, incluido editarlo en su propio programa, y [Primeros pasos](../../../web/help/admin-getting-started.md) recorre los seis pasos de Configuración.

Un archivo del espectáculo es texto sencillo. La versión corta:

```text
<!-- Comments like this are never shown. -->
[loop: 10s]
[image: welcome.png | 15s]
Welcome! Please silence your phones.
[/loop]

[section: Act 1, Scene 1]
[note: lights come up slowly]
[clear]
(A garden gate creaks open.)
MARGARET: You're late, Edwin.
EDWIN: The *train* was late.

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[/song]
~ Tea for two, and two for tea ~

[fullscreen: **INTERMISSION**]
```

- **Una línea = un subtítulo** = una pulsación de Espacio. Usted decide dónde cortar los parlamentos largos.
- Los **nombres de personaje** en mayúsculas seguidos de dos puntos se muestran sobre la línea.
- Las **acotaciones** entre `(paréntesis)` se muestran en cursiva.
- Las **letras** entre notas `♪`, en cursiva: empiece una línea con `~`, o envuelva la canción en `[song]`…`[/song]`. Para dos cantantes a la vez (un dúo o contrapunto), ponga sus líneas entre `[together]` y `[/together]`, o únalas con ` || ` en una línea; ambas se muestran a la vez con una pulsación de Espacio. Los botones **♪**, **♫** y **🎵 Canción** del editor los insertan. Consulte el tema de ayuda *Musicales*.
- Los **bucles** (previa y descanso) repiten imágenes y texto hasta que el operador continúa. Un bucle justo al principio del archivo del espectáculo arranca cuando el operador pone el espectáculo en el aire (INICIAR).
- Las **notas** `[note: …]` y los marcadores `[section: …]` son solo para el operador. `[pace: 8s]` indica al operador cuánto suele quedarse la línea siguiente, y una línea que empieza por `# ` es un encabezado que funciona como un marcador.
- El tema de ayuda **Escribir un archivo del espectáculo** lo enumera todo, con ejemplos, y la *chuleta del archivo del espectáculo* del editor lo resume.

### Ajustes del espectáculo (título, aspecto y más)

Todo lo que concierne a un espectáculo salvo sus imágenes se guarda en su archivo del espectáculo, incluidos su título y el aspecto de los subtítulos, en un bloque `[settings]` cerca del principio:

```text
[settings]
title: The Garden Party
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
[/settings]
```

El paso **Aspecto** de Configuración escribe este bloque por usted, y también puede editarlo a mano. Todo lo que un archivo del espectáculo omite usa un valor por defecto: el programa tiene valores integrados, `config.yaml` (su sección `defaults`) puede cambiarlos para todos los espectáculos de este equipo, y cada archivo del espectáculo anula ambos. Los puertos, las contraseñas, HTTPS y las listas de acceso no son ajustes del espectáculo y se quedan en `config.yaml`. Como el archivo del espectáculo lleva su aspecto, copiar o exportar el espectáculo también lleva el aspecto. Consulte el tema de ayuda *Escribir un archivo del espectáculo* para ver todos los ajustes.

### Revisar los nombres de los personajes

Haga clic en **Caracteres…** sobre el editor para ver todos los nombres usados como personaje, los más usados primero, con el número de líneas de cada uno. Es la forma rápida de encontrar un nombre mal escrito: **ALCE** con una línea junto a **ALICE** con cuarenta se marca con *¿Quiso decir ALICE?*. Haga clic en un nombre para ir a su primera línea. Un grupo como `ALICE & BOB:` cuenta para cada uno de sus nombres.

### Volver atrás

Cada guardado conserva la versión que reemplazó (las últimas 50). **Versiones anteriores…** sobre el editor las enumera; **Cargar** pone una en el editor, y **Guardar** la hace actual. **Comparar** muestra línea a línea qué cambiaría al cargar una, y una versión a la que ponga nombre (o **Conservar** para lo que está guardado ahora) nunca se olvida. Exporte antes del espectáculo de todos modos: las versiones anteriores viven en la carpeta de datos, no en la exportación.

### Dos personas editando

Si otra persona guarda un archivo del espectáculo mientras usted lo tiene abierto, el programa le avisa al guardar en lugar de sobrescribir su trabajo, y le permite quedarse con su versión o conservar la suya.

### Más de un idioma

Un espectáculo puede ofrecer subtítulos en varios idiomas. En Configuración → Archivo del espectáculo, pulse **＋ Añadir un idioma**, deje marcada la casilla «Empezar con una copia del archivo principal del espectáculo» y traduzca las palabras. Cada idioma es su propio archivo del espectáculo con las mismas líneas, comandos y pantallas en blanco en el mismo orden, así que una pulsación de Espacio muestra el mismo momento en todos los idiomas. El operador sigue solo el idioma principal. Los dispositivos del público reciben el idioma en que está configurado su teléfono (o el principal) y pueden cambiar desde el pequeño menú 🌐, o recibir un enlace que termine en `?lang=es`. Una traducción que deja de coincidir con el archivo principal del espectáculo se señala y se retiene hasta que se corrija. El espectáculo de ejemplo incluye una traducción al español para verla. Consulte el tema de ayuda *Idiomas*.

### Cambiar cosas durante los ensayos y los espectáculos

Puede editar y volver a guardar un archivo del espectáculo con el espectáculo en directo (por ejemplo para corregir una errata). El programa le pide que escriba el nombre del espectáculo para asegurarse, conserva el sitio del operador por número de línea y avisa si cambió el número de líneas.

## Dirigir el espectáculo

### Antes de que llegue el público

1. Arranque el servidor y abra la **página del operador**; inicie sesión. Debería decir **Usted tiene el control**.
2. En Configuración → Compartir pulse **Abrir la página Dirigir**. El espectáculo se carga en espera: el público ve la pantalla de espera del local. Cuando esté listo, pulse **Espacio** (INICIAR) en la página del operador para ponerlo en el aire. Un bucle de previa, si el archivo del espectáculo lo tiene, empieza entonces en todas las pantallas.
3. Abra la dirección del público en su propio teléfono y, si tiene uno, en el proyector. Compruebe que el recuento de espectadores sube en la página del operador.
4. Ponga el cartel QR donde la gente lo vea.

### Durante el espectáculo

- **Espacio** envía la línea siguiente (también Flecha derecha, Av Pág, Intro y la mayoría de pulsadores).
- **Retroceso** deshace lo último que envió.
- **B** deja en blanco las pantallas del público (oculta todo; los mismos subtítulos vuelven al pulsar **B** otra vez). **C** las borra (borra los subtítulos, para que la línea siguiente empiece en una pantalla vacía). **Retroceso** deshace cualquiera de las dos.
- **Mayús+Q** detiene un bucle de descanso o de previa (pantalla en blanco). **Espacio** continúa.
- **Anuncio…** pone un aviso puntual en todas las pantallas («El espectáculo se reanudará en breve»). **Atrás** lo quita.
- Haga clic en una línea y confirme para saltar a cualquier sitio. **/** busca en el archivo del espectáculo.
- La línea con la flecha amarilla es la *siguiente*. El punto verde marca lo que está ahora en pantalla.

### Cuando termina el espectáculo

Pulse `Ctrl+C` en la ventana del programa (o detenga el servicio). La posición se guarda. Para volver a dirigir el espectáculo desde el principio, use **Reiniciar espectáculo…** en la página del operador, o **Reiniciar espectáculo…** en Configuración → Compartir. Ambos lo devuelven a la espera. **¿Cuánta gente lo usó?** Cuando el espectáculo llega a su final, el cartel de la página del operador dice *Fin del espectáculo.* con un enlace **Informe del público** (el paso **Informe** de la configuración muestra el último). Cuenta los dispositivos distintos que usaron la página del público mientras el espectáculo estaba en el aire, sin importar cuánto se quedó cada uno (un teléfono que se cayó de la WiFi y volvió cuenta una vez), con un gráfico de los dispositivos a lo largo del tiempo (que muestra cuándo funcionaron los bucles de antes del espectáculo y del intermedio), el momento de más afluencia, los idiomas y una fila por dispositivo; se puede imprimir o descargar como archivo CSV. Los dispositivos se distinguen por su dirección y el tipo de navegador, así que es una estimación. Los informes se guardan en la carpeta `reports` de la carpeta de datos.

## Lista de comprobación del día

**Unos días antes**
- [ ] Ensaye con el archivo real del espectáculo y unos cuantos teléfonos reales en la red real y en la sala real.
- [ ] Exporte el espectáculo (Configuración → Espectáculos → **Más ▾** → **Exportar**) y guarde el `.zip` en un lugar seguro, como una memoria USB.
- [ ] Confirme que la dirección del equipo de subtitulado está reservada y que el cartel QR la muestra.

**En el local, una hora antes**
- [ ] Alimentación: equipo y punto de acceso enchufados; equipo configurado para no suspenderse nunca.
- [ ] Punto de acceso encendido, en su sitio, con el nombre correcto; el equipo conectado a él por cable.
- [ ] Arranque el programa; anote la dirección que imprime (y la de Compartir).
- [ ] Un teléfono en la WiFi del público puede abrir la dirección y muestra «Los subtítulos comenzarán en breve».
- [ ] Configuración → Prueba: haga un ensayo rápido; las teclas del operador funcionan.
- [ ] Navegador del proyector (si se usa) abierto en `…/?display=projector`, a pantalla completa.
- [ ] Pulse **INICIAR** en la página del operador. El bucle de previa (si lo hay) está en todas las pantallas.
- [ ] El operador sentado donde pueda ver el escenario; pulsador o teclado probado.
- [ ] Cartel QR colocado; una persona voluntaria sabe el nombre de la WiFi y la dirección para quien pregunte.

**Después del espectáculo**
- [ ] `Ctrl+C` en la ventana del programa.
- [ ] Vuelva a exportar el espectáculo si cambió algo.

## Resolución de problemas

| Problema | Qué probar |
|----------|------------|
| Nadie puede abrir la dirección | ¿Está el teléfono en la WiFi correcta (no en datos móviles)? ¿Incluye la dirección `:8080` si el programa usó ese puerto? ¿Se permite el tráfico de invitados al servidor? ¿Permite el cortafuegos del equipo el programa en redes privadas? Consulte [venue-network-guide.md](venue-network-guide.md). |
| Algunos dispositivos se conectan y otros no | El punto de acceso puede estar en su límite de dispositivos. Añada un segundo punto de acceso, o acerque el primero a la sala. |
| Los subtítulos llegan tarde o se detienen en unos pocos teléfonos | Señal débil o canal saturado. Esos teléfonos se reconectan solos. Mejore la colocación del punto de acceso; use 5 GHz. |
| Alguien se perdió una línea | Puede desplazar los subtítulos hacia arriba para releerla; las líneas nuevas esperan tras un botón **Volver al directo** hasta que lo toque. |
| Una página dice «Este navegador es demasiado antiguo» | El navegador del dispositivo está por debajo de lo que la página necesita: unos iOS 12 para la página del público, iOS 14.1 / Chrome 87 para Configuración y la página del operador. Actualice el dispositivo, o use otro. |
| Las pantallas de los teléfonos se oscurecen | Los teléfonos bloquean su pantalla. La página mantiene la pantalla encendida con el bloqueo de activación del navegador cuando puede y, cuando no puede (HTTP sencillo), con un pequeño vídeo invisible en bucle. Quien prefiera dejar dormir su teléfono (por poca batería, por ejemplo) puede desactivar *La pantalla sigue encendida* en el panel «Aa». Algunos teléfonos en modo de bajo consumo rechazan ambas cosas: la página muestra entonces un consejo para desactivar el bloqueo automático durante la noche. |
| Se envió la línea equivocada | **Retroceso** la deshace. |
| La página del operador dice que otro operador tiene el control | Use **Asumir el control**, si usted es el operador. |
| El equipo se congeló o se reinició a mitad del espectáculo | Arranque el programa otra vez. La página del operador se reconecta, sigue con la sesión iniciada y recupera el control sola, el espectáculo está donde lo dejó y los dispositivos del público se reconectan solos. Un bucle empieza de nuevo desde su primer elemento. **Retroceso** aún puede deshacer las últimas cosas enviadas antes del reinicio. |
| Olvidé una contraseña | Inicie sesión con la otra y use **Ajustes** en Configuración, o abra `config.yaml` en la carpeta de datos con cualquier editor de texto. |
| Otra cosa | Configuración → Solucionar problemas → **Descargar un paquete de soporte**; contiene los ajustes (sin contraseñas) y líneas recientes del registro. |

## Mantenimiento

- **Copias de seguridad:** *Exportar* (menú **Más ▾**) en Configuración → Espectáculos guarda un espectáculo (archivo del espectáculo, imágenes, aspecto) como un `.zip`. *Importar* lo restaura, en este equipo o en otro. También puede copiar toda la carpeta de datos.
- **Cambiar a otro equipo:** instale el programa, importe los espectáculos exportados, ajuste la dirección del equipo nuevo y, si quiere conservar las contraseñas antiguas, copie `config.yaml`.
- **Actualizar el programa:** deténgalo (`Ctrl+C`), sustituya el archivo del programa por el más nuevo y arránquelo de nuevo con la misma carpeta `--data`. No actualice justo antes de un espectáculo.
- **Reutilizar un espectáculo:** Configuración → Espectáculos tiene **Renombrar…** y **Duplicar**. Duplicar hace una copia del archivo del espectáculo, las imágenes y el aspecto para adaptarla a la próxima temporada, y deja el original intacto. Renombrar, Duplicar, Exportar, Archivar y Eliminar están en el menú **Más ▾** de cada espectáculo.
- **Limpiar:** elimine los espectáculos que ya no necesite en Configuración → Espectáculos. Los `.zip` exportados son su archivo.
- **Informes del público:** cada función deja su informe del público como un archivo en la carpeta `reports` dentro de la carpeta de datos (Configuración → Ajustes → Acerca de muestra dónde está la carpeta de datos). Se conservan los 30 más recientes y los más antiguos se eliminan solos. Un informe enumera las direcciones de red de los dispositivos, así que elimine los informes que no quiera conservar (**Eliminar…** en el paso **Informe** de Configuración, o los propios archivos), cualquiera de ellos o toda la carpeta `reports` (se vuelve a crear cuando el siguiente espectáculo sale al aire); nada más depende de ellos, y borrarlos no afecta a sus espectáculos.
- **Registros:** Configuración → Solucionar problemas → la pestaña **Avisos y errores** muestra lo que el programa ha registrado últimamente, filtrado por nivel (elija *Solo problemas* para ver avisos y errores); las líneas recientes del registro también están en el paquete de soporte; añada `--log-file` para conservar un archivo. Cada línea empieza con su nivel (`INFO`, `WARN`, `ERROR` o `DEBUG`). El `info` por defecto registra inicios de sesión, quién tomó el control de la página del operador, espectáculos activados y eventos similares; `warn` conserva solo los problemas; `debug` registra además cada petición, lo que ayuda cuando ocurre algo raro. Defínalo con `log: {level: debug}` en `config.yaml`, con `--log-level debug`, o mientras el programa funciona desde Configuración → Ajustes → *Nivel de detalle del registro*. Si el programa se inició con `--log-level`, esa opción manda en cada arranque: Configuración lo dice, un cambio hecho ahí dura solo hasta el siguiente reinicio, y el registro de arranque avisa cuando la opción y el archivo discrepan. Un archivo de registro conservado con `--log-file` empieza uno nuevo a los 5 MB y guarda tres anteriores (`open-caption.log.1` a `.3`), de modo que nunca llena el disco.
- **Archivar:** en Configuración → Espectáculos, **Archivar** guarda aparte un espectáculo terminado (se conserva, aparece en «Espectáculos archivados» y no puede salir en directo) y **Restaurar** lo devuelve. **Eliminar** borra un espectáculo definitivamente.

## Tarjeta rápida del operador

Imprímala y téngala junto al operador.

| Tecla | Hace |
|-------|------|
| **Espacio** | Enviar la línea siguiente |
| **Retroceso** | Deshacer lo último enviado |
| **S** | Saltar la siguiente línea sin enviarla (Retroceso lo deshace) |
| **Mayús+Q** | Detener un bucle (pantalla en blanco) |
| **B** | En blanco: oculta todo, los subtítulos vuelven al quitar el blanco |
| **C** | Borrar: borra los subtítulos para empezar de cero |
| **/** | Buscar en el archivo del espectáculo |
| **?** | Ayuda (las teclas se pausan mientras está abierta) |

- Flecha amarilla = línea siguiente. Punto verde = en pantalla ahora.
- ¿Línea equivocada? **Retroceso.** ¿Ha perdido el sitio? Haga clic en la línea que quiere y confirme.
- ¿Necesita decir algo al público? **Anuncio…**
- ¿Espectáculo interrumpido? **B** para dejar en blanco las pantallas, **Anuncio…** para explicarlo.
