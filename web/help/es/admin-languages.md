# Idiomas

Un espectáculo puede ofrecer sus subtítulos en más de un idioma. Cada idioma es su **propio archivo del espectáculo**, y el operador sigue solo el principal: una pulsación de Espacio envía el mismo momento en todos los idiomas a la vez.

## Añadir un idioma

1. En el paso **Archivo del espectáculo**, pulse **＋ Añadir un idioma** junto a las pestañas de idioma.
2. Elija el idioma (o elija «Otro idioma…» y escriba un código corto como `cy`).
3. Deje marcada la casilla **Empezar con una copia del archivo principal del espectáculo**. Obtiene una copia con la forma correcta, así que solo tiene que reemplazar las palabras.
4. Traduzca y luego **Guarde**.

Cada traducción empieza con una línea que nombra su idioma. El editor la añade por usted:

```showfile
[language: es | Español]
Hola, Margaret.
```

El nombre que va después de la `|` es lo que ven los miembros del público en el menú de idiomas, así que escríbalo en su propio idioma (Español, Français, Deutsch).

## La única regla: conservar la misma forma

Cada idioma debe tener **las mismas señales en el mismo orden**: un subtítulo sigue siendo un subtítulo, un `[clear]` sigue siendo un `[clear]`, un bucle conserva el mismo número de elementos y las voces mostradas a la vez siguen siendo el mismo número de voces. Solo cambian las palabras. Los nombres de personaje, las notas, los nombres de sección, los nombres de imagen y los tiempos de los bucles pueden ser distintos.

Si una traducción deja de coincidir, el editor muestra un error con la línea donde falla, por ejemplo «La señal 4 no coincide con el archivo del espectáculo en inglés: ese archivo tiene [clear] aquí, y este tiene un subtítulo». Un idioma que no coincide **no se ofrece al público** hasta que se corrija. El archivo del espectáculo en inglés (principal) no se ve afectado, y el paso Dirigir enumera el problema.

Para indicar al público qué palabras van con qué línea, mantenga las líneas sincronizadas: si el inglés divide un parlamento en tres líneas, el español también, aunque no sea así como lo dividiría ese idioma.

## Lo que ve el público

- A cada dispositivo se le da el mejor idioma que tenga el espectáculo. Se mira la configuración de idioma del teléfono o del navegador (por ejemplo el español de México recibe español), y se usa el idioma principal cuando nada coincide.
- Un pequeño control **🌐** en la esquina superior izquierda enumera los idiomas. Al elegir uno se cambia ese dispositivo, y la dirección pasa a terminar en `?lang=es`, de modo que puede imprimir un enlace o un código QR para cada idioma.
- Las propias palabras de la página («Reconectando…», tamaño del texto, etc.) también están traducidas al español, francés, alemán, italiano y portugués. Los demás idiomas muestran esas pocas etiquetas en inglés; los subtítulos en sí están siempre en el idioma del archivo del espectáculo.
- Los idiomas que se escriben de derecha a izquierda (árabe, hebreo) se alinean a la derecha.

## Otras cosas que conviene saber

- **Imágenes:** todos los idiomas pueden usar las mismas imágenes. Para una imagen con palabras, suba una versión por idioma e indique cada una en el archivo del espectáculo de ese idioma.
- **Mensaje previo:** cada idioma tiene el suyo, guardado en el archivo del espectáculo de ese idioma (`idle message:` en su bloque `[settings]`). El paso Aspecto tiene un cuadro para cada idioma y los escribe por usted. El archivo del espectáculo de una traducción no puede definir nada más: el aspecto se define una sola vez, en el archivo principal del espectáculo.
- **Los mensajes del operador** («Mensaje…») se envían tal como se escriben a todos los idiomas, así que escríbalos en los idiomas que necesite.
- **Cambiar un archivo del espectáculo con el espectáculo en marcha:** añadir o corregir una traducción surte efecto al instante y conserva el sitio del operador. No se permite quitar un idioma mientras el espectáculo está en directo.
- **Recuentos:** la página del operador muestra cuántos dispositivos leen cada idioma, por ejemplo `👥 83 (en 60 · es 23)`.
- **Copias de seguridad:** Exportar guarda todos los idiomas; Duplicar también los copia.

Consulte también [Escribir un archivo del espectáculo](topic:admin-writing-a-show-file).
