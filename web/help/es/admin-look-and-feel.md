# Aspecto

El paso **Aspecto** define cómo se ven los subtítulos en cada dispositivo del público. Los valores por defecto, texto blanco sobre fondo negro, son los más fáciles de leer en una sala oscura y los mejores para la accesibilidad.

- **Colores:** una lectura del contraste muestra lo legible que es la combinación. Busque 7:1 o más; la página avisa cuando es demasiado bajo.
- **Tamaño del texto:** escala todo el texto. Los dispositivos ajustan el texto a su pantalla, así que es un cambio relativo, no un tamaño en puntos. El control deslizante va del 50 al 300 por ciento; 100 es lo normal.
- **Fuente:** limpia (la fuente habitual del dispositivo), con serifa, monoespaciada, o **apta para dislexia (OpenDyslexic)**, una fuente con la parte inferior de las letras más gruesa y formas de letra distintivas que a algunos lectores con dislexia les resultan más fáciles. Está integrada en el programa, así que no necesita internet, y un dispositivo la descarga (unos 115 KB) solo cuando se usa. Las líneas ganan algo de altura con ella. Elija **alta legibilidad (Atkinson Hyperlegible)** para una fuente limpia diseñada por el Braille Institute para la baja visión, con letras fáciles de distinguir (por ejemplo la I mayúscula, la l minúscula y el número 1). Ambas fuentes están integradas en el programa, y un dispositivo descarga una solo si la usa (unos 17 KB la Atkinson Hyperlegible). Los miembros del público también pueden elegir cualquiera de las dos fuentes para su propio dispositivo con el pequeño control **Aa**, sea cual sea la fuente que usted eligiera aquí.
- **Nombres de personaje:** se muestran en un solo color (el *color de los nombres de personaje*), en un **color distinto para cada personaje**, o se ocultan. Los colores de los personajes se eligen automáticamente, en el orden en que cada personaje habla por primera vez, de una paleta que funciona para espectadores daltónicos, y se ajustan para seguir siendo legibles sobre su fondo.
- **Posición del nombre de personaje:** sobre la línea (lo habitual), **junto al texto** en la misma línea (en negrita), o **en una columna** a la izquierda del texto, como en un guion. Las voces que se muestran a la vez siempre ponen el nombre junto a su propio texto.
- **Mensaje previo al espectáculo:** se muestra hasta que se envía la primera línea, si el archivo del espectáculo no empieza con un bucle.

Los cambios se guardan con **Guardar aspecto** y llegan enseguida a los dispositivos conectados. **Guardar aspecto escribe sus elecciones en el bloque `[settings]` al principio del archivo del espectáculo** (consulte [Escribir un archivo del espectáculo](topic:admin-writing-a-show-file)), de modo que el archivo del espectáculo lleva su propio aspecto dondequiera que se copie; también puede editar ese bloque a mano. Todo lo que el archivo del espectáculo no define usa los valores por defecto de `config.yaml` (su sección `defaults`), que a su vez parten de los valores integrados del programa. Si el editor del archivo del espectáculo tiene cambios sin guardar, guárdelos antes de guardar el aspecto.

## Una pantalla para toda la sala

Si un navegador muestra los subtítulos en un proyector o un monitor grande, abra la dirección del público con `?display=projector` al final. El texto es más grande, el puntero del ratón se oculta y no se muestra nada más. Una pantalla compartida nunca se desplaza, así que si el subtítulo más reciente es más alto que la pantalla (por ejemplo, muchas voces a la vez), el texto se reduce hasta que cabe.

## Releer líneas anteriores

Cualquiera puede desplazar los subtítulos hacia arriba para releer algo que se perdió. Las líneas nuevas esperan entonces tras un botón **↓ Volver al directo (3)** (que cuenta lo que llegó) en lugar de arrastrar la página, y las líneas más antiguas dejan de atenuarse mientras alguien las lee. Pulsar el botón, o desplazarse hasta el final, vuelve a los subtítulos en directo. Una pantalla con `?display=projector` nunca se puede desplazar.

## Controles del público

Los miembros del público nunca necesitan hacer nada. Un pequeño control tenue **Aa** en la esquina superior derecha de la página del público permite a cada persona hacer su propio texto más grande o más pequeño, pasar a oscuro sobre claro, elegir la fuente apta para dislexia o la de alta legibilidad, o desactivar **La pantalla sigue encendida** (la página mantiene encendida la pantalla del teléfono mientras dura el espectáculo, y quien tenga poca batería puede preferir no hacerlo), pero solo en su propio dispositivo.
