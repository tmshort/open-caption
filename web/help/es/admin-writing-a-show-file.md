# Escribir un archivo del espectáculo

Un archivo del espectáculo es un archivo de texto sencillo (`.txt` o `.md`) escrito en una forma simple de Markdown. Si pega un guion de teatro corriente, ya funciona: **cada línea es un subtítulo**, y una pulsación de la barra espaciadora envía una línea. Las líneas en blanco se ignoran, así que úselas para mantener legible el archivo.

Las líneas largas se ajustan solas en la pantalla. Usted decide dónde dividir un parlamento largo en líneas, según el estilo de la representación. El programa nunca divide ni une líneas por usted.

Para comprobar un archivo del espectáculo, o para prepararlo en su propio editor de texto, consulte [Preparar un archivo del espectáculo](admin-preparing-a-show-file.md).

## Formato

| Escriba | Obtiene |
|---------|---------|
| `**fuerte**` | texto en **negrita** |
| `*aparte*` o `_aparte_` | texto en *cursiva* |
| `***very loud***` | texto en negrita y cursiva |
| `~~old line~~` | texto tachado (una corrección, o un chiste tachado) |
| `(thunder rumbles)` | una línea entera entre paréntesis se muestra como un sonido o una acotación |
| `HAMLET: To be, or not to be` | una etiqueta de personaje sobre la línea (el nombre debe ir en MAYÚSCULAS seguido de dos puntos; también vale un grupo: `ALICE & BOB:`, `ALL (EXCEPT BOB):`) |

**Tachar texto.** Ponga dos tildes a cada lado, `~~así~~`, y las palabras se muestran con una línea encima. Debe haber un carácter que no sea un espacio justo dentro de cada pareja (`~~ x ~~` se queda tal cual). Una línea cantada empieza con una sola `~` seguida de un espacio, así que nunca se confunden; para empezar una línea normal con una tilde, escriba `\~`.

Todo lo demás se muestra tal como se escribe. No se admiten enlaces, listas, tablas ni títulos. Tampoco se admite HTML sin procesar: si escribe `<b>`, aparece literalmente, y la página muestra una nota que sugiere la forma Markdown. El subrayado, los superíndices y los subíndices no tienen marca, pero cualquier carácter que pueda escribir funciona tal cual, así que escriba H₂O, x², 1ˢᵗ o ½ directamente.

```showfile
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Revisar los nombres de los personajes

Haga clic en **Caracteres…** sobre el editor para ver todos los nombres usados como personaje, los más usados primero, con el número de líneas de cada uno. Es la forma rápida de encontrar un nombre mal escrito: **ALCE** con una línea junto a **ALICE** con cuarenta se marca con *¿Quiso decir ALICE?*. Haga clic en un nombre para ir a su primera línea. Un grupo como `ALICE & BOB:` cuenta para cada uno de sus nombres.

## Volver a una versión anterior

Cada vez que guarda, se conserva la versión que reemplazó (las últimas 50 de cada idioma). Pulse **Versiones anteriores…** sobre el editor, elija una hora y pulse **Cargar**. Eso solo pone el texto antiguo en el editor; nada cambia hasta que pulse **Guardar**, así que puede mirar primero. **Comparar** muestra, línea a línea, qué cambiaría al cargar una versión. Una versión a la que ponga nombre (en el cuadro junto a ella, o con **Conservar** para dar nombre a lo que está guardado ahora) nunca se olvida, por muchos guardados que sigan. La exportación no incluye las versiones anteriores.

## Ajustes del espectáculo: el título y el aspecto

Todo lo que concierne a un espectáculo salvo sus imágenes se guarda en el archivo del espectáculo, incluidos su **título** y el **aspecto** de los subtítulos. Esto va en un bloque `[settings]`, normalmente al principio, con un `nombre: valor` por línea:

```showfile
[settings]
title: The Garden Party
background: #000000
text color: #ffffff
speaker name color: #ffffff
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
loop seconds: 10
scrollback lines: 200
[/settings]
```

| Ajuste | Qué hace |
|--------|----------|
| `title` | El nombre del espectáculo, tal como aparece en Configuración |
| `background` | El color de la pantalla tras los subtítulos, como `#000000` (también sirve un nombre como `black` o `yellow`) |
| `text color` | El color de los subtítulos, como `#ffffff` |
| `speaker name color` | El color de los nombres de personaje (cuando no se les da un color a cada uno) |
| `font` | `system`, `serif`, `mono`, `dyslexic` o `hyperlegible` |
| `text size` | Un porcentaje de 50 a 300; 100 es normal |
| `speaker names` | `label` (un color), `color` (un color distinto para cada personaje) o `hidden` |
| `speaker position` | `above` (el nombre sobre la línea), `beside` (en la misma línea, antes del texto) o `column` (en una columna a la izquierda del texto) |
| `idle message` | Las palabras que se muestran antes de enviar la primera línea |
| `loop seconds` | Cuánto permanece cada elemento de un `[loop]` cuando no lo indica |
| `scrollback lines` | Cuántos subtítulos puede recorrer hacia atrás un espectador (de 20 a 2000) |
| `archived` | `yes` si el espectáculo está archivado; los botones **Archivar** y **Restaurar** de Configuración lo añaden y lo quitan |

**No tiene que escribir nada de esto.** El paso **Aspecto** de Configuración rellena el bloque por usted (y **Renombrar** define el título), y todo lo que omita usa simplemente un valor por defecto. Los valores por defecto vienen en dos capas: el programa tiene valores integrados, que el archivo de ajustes `config.yaml` puede cambiar para todos los espectáculos (su sección `defaults`), y luego cada archivo del espectáculo anula lo que defina. Gana el más específico: **archivo del espectáculo, luego config.yaml, luego los valores integrados**.

Algunas reglas: el archivo del espectáculo de una traducción solo puede definir `idle message` (el aspecto se define una vez, para todo el espectáculo, en el archivo principal del espectáculo); un ajuste mal escrito o con un valor que no tiene sentido se señala con su número de línea y se ignora; y los comentarios dentro del bloque no se conservan cuando Configuración lo reescribe, así que ponga los comentarios encima. Varios bloques `[settings]` se combinan, y los valores posteriores ganan.

## Canciones y música

Para un musical, muestre las letras entre notas musicales (♪). Hay tres formas, de la más rápida a la más flexible:

- **Escriba o pegue usted mismo las notas.** ♪ y ♫ funcionan en cualquier parte del archivo del espectáculo. En el editor, los botones **♪** y **♫** las insertan en el cursor.
- **Empiece una línea con `~`.** La línea se trata como una letra cantada: se muestra en cursiva con una ♪ añadida en cada extremo, y la propia `~` no se muestra. Solo hace falta la `~` inicial: la ♪ del final se añade sola, y una `~` de cierre se admite pero no cambia nada. También funciona tras el nombre de un personaje.
- **Envuelva una canción entera en `[song]` y `[/song]`.** Todas las líneas de dentro se cantan, así que no necesita una `~` en cada una. El título tras los dos puntos es un marcador solo para el operador. Las notas, los clear y los blank pueden seguir yendo dentro.

```showfile
OLIVER: ~ Where is love? ~
~ Row, row, row your boat

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[note: hold the last chord]
[/song]
```

### Dos o más voces a la vez (dúos y contrapunto)

Cuando los cantantes cantan líneas distintas **al mismo tiempo**, póngalas en `[together]` … `[/together]`. Todas las líneas de dentro se muestran a la vez, una al lado de otra en una pantalla ancha y apiladas en un teléfono, y el operador las envía con una sola pulsación de Espacio. Cada una conserva su propio nombre de personaje. Para un intercambio breve puede escribir las voces en una línea, separadas por ` || ` (un espacio, dos barras verticales, un espacio). `[duet]` y `[counterpoint]` significan lo mismo que `[together]`.

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

Las voces que se unen **una tras otra** son líneas separadas: pulse Espacio para cada una. Use `[together]` solo para líneas que aparecen en el mismo momento. Un bloque `[together]` contiene solo líneas de texto, necesita al menos dos, y dentro de un `[song]` todas las voces se cantan.

Una tilde en medio de una línea, como `about ~5 minutes`, es solo texto. Para empezar una línea normal con una, escriba `\~`. Consulte [Musicales](admin-musicals.md) para más consejos.

## Varias líneas en un subtítulo

Termine una línea con una barra invertida para unirla con la siguiente, o ponga las líneas entre `[[` y `]]`.

```showfile
GHOST: Remember me. \
Remember me.

[[
Row, row, row your boat,
Gently down the stream
]]
```

## Comentarios y notas

- `<!-- así -->` es un **comentario**. Nadie lo ve, y puede ocupar varias líneas.
- `[note: texto]` es un recordatorio que ve **solo el operador**, por ejemplo `[note: hold for applause]`.
- `[section: Act 2, Scene 1]` añade un marcador en la lista del operador. Las notas y las secciones nunca gastan una pulsación de la barra espaciadora.
- `[pace: 8s]` dice cuánto suele permanecer la línea siguiente en pantalla, como recordatorio **solo para el operador**. La página de control lo muestra junto a esa línea (`⏱ about 8 s`) y, una vez enviada, cuenta atrás hasta que toca la siguiente pulsación, y se pone amarilla cuando toca. Nunca envía nada por sí solo. El tiempo es de más de 0 segundos. Pertenece a la siguiente línea que envía el operador (las secciones y las notas se saltan) y no puede ir dentro de un bucle ni de un bloque `[together]`. Si el tiempo es mucho mayor que lo que tarda en leerse la línea, la lista de problemas añade una nota, por si es un error. El botón **Ritmo** del editor inserta uno.
- Una línea que empieza con `#` y un espacio es un **título** y funciona como `[section: …]`: `# Act 2` (o `## Scene 1`, con hasta seis `#`) añade un marcador en la lista del operador y nunca se muestra al público. Un `#` sin espacio detrás (un hashtag, por ejemplo) se ignora, con una nota; para empezar un subtítulo con `#`, escriba `\#`. Los títulos no pueden ir dentro de un bucle.

```showfile
<!-- Act 1 captions, version 3 -->
[section: Act 1, Scene 1]
[note: lights come up slowly, wait for the first line]
HAMLET: Who's there?
```

## Comandos

Una línea entre corchetes es un comando. Los nombres no distinguen mayúsculas de minúsculas.

| Comando | Qué hace |
|---------|----------|
| `[clear]` | Borra los subtítulos de la pantalla |
| `[blank]` | Deja la pantalla en blanco (los subtítulos se conservan) |
| `[image: file.png]` | Muestra una imagen a pantalla completa. Añada una leyenda con `[image: file.png \| text]` |
| `![caption](file.png)` | Lo mismo, escrito a la manera de Markdown |
| `[fullscreen: text]` | Muestra un texto grande a pantalla completa. Para varias líneas use `[fullscreen]` … `[/fullscreen]` |
| `[loop]` … `[/loop]` | Un conjunto repetido de elementos a pantalla completa para la previa o el descanso. Consulte [Bucles](admin-loops.md) |
| `[together]` … `[/together]` | Dos o más voces mostradas al mismo tiempo. Consulte [Canciones y música](admin-musicals.md) |
| `[song: title]` … `[/song]` | Una canción: todas las líneas de dentro se muestran como letra cantada. Consulte [Canciones y música](admin-musicals.md) |
| `[section: name]` | Un marcador para el operador |
| `[note: text]` | Una nota para el operador |
| `[pace: 8s]` | Un recordatorio de ritmo para el operador (véase arriba) |
| `[language: es]` | Marca un archivo del espectáculo traducido. Consulte [Idiomas](admin-languages.md) |
| `[settings]` … `[/settings]` | El título y el aspecto del espectáculo. Consulte *Ajustes del espectáculo* más arriba |

Las opciones de tiempo solo tienen sentido dentro de un [bucle](admin-loops.md): `| 15s` tras una imagen o un texto a pantalla completa, y `[blank: 3s]`. Fuera de un bucle, una imagen o un texto a pantalla completa permanece hasta que el operador pasa a la siguiente línea, el tiempo se ignora y se señala como advertencia, igual que un tiempo en `[blank]`.

```showfile
[clear]
(Lights up on the garden.)
[image: garden.png | The garden at dusk]
[fullscreen: **Please remain seated**]
[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
[blank]
```

Una imagen, un texto a pantalla completa o un blanco permanece en las pantallas del público hasta que el operador envía la línea siguiente.

## Caracteres especiales

Ponga una barra invertida antes de un carácter para mostrarlo literalmente en lugar de tratarlo como formato: `\*`, `\_`, `\#`, `\[`, `\]`, `\:`, `\(`, `\)`, `\!`, `\~` y `\\` (una barra invertida en sí). Por ejemplo, un subtítulo que debe empezar con un signo de número se escribe `\# 1 hit`, y `HAMLET\: not a speaker` muestra los dos puntos sin crear una etiqueta de personaje.

```showfile
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## Qué significan los avisos

El paso Archivo del espectáculo enumera los problemas con números de línea. Pulse uno para saltar a la línea.

- **«Comando desconocido [x]»** — una línea entre corchetes que no es un comando. Si pretendía ser un subtítulo, como un efecto de sonido, escríbala con paréntesis, como `(laughter)`.
- **«No se encontró la imagen …»** — la imagen no está subida, o el nombre está escrito de otra forma. Súbala en el paso Imágenes.
- **«[loop] se abrió pero nunca se cerró»** — añada `[/loop]` tras el último elemento. Lo mismo vale para los bloques `[fullscreen]`, los bloques `[[` y los comentarios.
- **«No se admite HTML»** — use `**negrita**` y `*cursiva*` en lugar de etiquetas.
- **«no puede ir dentro de un bucle»** — los bucles contienen solo imágenes y texto.
- **«[together] necesita al menos dos líneas»** — un bloque `[together]` con una línea se muestra como un subtítulo normal. **«[together] se abrió pero nunca se cerró»** — añada `[/together]`; las líneas que escribió se conservan.
- **«Este subtítulo muestra 5 voces a la vez»** — cinco o más voces en un bloque `[together]` (o una línea con ` || `) son muchas para un teléfono, que las apila, y para una pantalla compartida, que tiene espacio para unas cuatro y nunca se desplaza. Divídalo en dos subtítulos.
- **«Ajuste desconocido» / «no es una fuente» / «no es un color»** — una línea del bloque `[settings]` que Configuración no puede usar; se ignora y se aplica el valor por defecto. El mensaje dice qué escribir en su lugar.
- **«[settings] se abrió pero nunca se cerró»** — añada `[/settings]`; todo lo que sigue a un bloque sin cerrar se lee como ajustes.
- **«[song] se abrió pero nunca se cerró»** — añada `[/song]` tras la última letra. Las letras que escribió se conservan igualmente.

Los errores nunca le impiden guardar.

## Si dos personas editan el mismo archivo del espectáculo

Si otra persona guarda el archivo del espectáculo mientras usted lo tiene abierto, el programa no sobrescribirá en silencio su trabajo. Cuando guarda, le dice que el archivo del espectáculo fue cambiado por otra persona y le permite **usar su versión** (reemplazando lo que hay en su editor) o **guardar la suya de todos modos**. Copie lo que quiera conservar antes de elegir. La parte con el error se omite, y el resto del archivo del espectáculo sigue funcionando.
