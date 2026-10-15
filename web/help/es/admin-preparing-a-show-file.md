# Preparar un archivo del espectáculo

Esta página es para quien convierte el guion en subtítulos. No necesita dirigir la función ni saber nada de redes. Un archivo del espectáculo es un archivo de texto corriente, así que puede escribirlo con el programa que prefiera, en cualquier ordenador, y traerlo a Configuración cuando esté listo. [Escribir un archivo del espectáculo](topic:admin-writing-a-show-file) enumera todas las marcas que puede usar; esta página trata de cómo abordar el trabajo.

## La versión corta

1. Ponga el guion en un archivo de texto sin formato. Si quiere, parta del espectáculo de ejemplo en [/help/sample-show.md](/help/sample-show.md).
2. Haga que **una línea del archivo sea un subtítulo**: una pulsación de la barra espaciadora.
3. Añada lo que ayuda al público y al operador (nombres de personajes, descripciones de sonidos, secciones, notas).
4. Traiga el archivo a Configuración, lea la lista de problemas, corríjalos y guarde.
5. Ensaye en el paso Prueba y corrija lo que note.

## Qué tipo de archivo

- Sirve cualquier archivo de texto sin formato: un nombre que termine en `.txt` o `.md`, y cualquier nombre, como `Nuestra ciudad - subtitulos v3.txt`.
- Guárdelo como **UTF-8** si su programa se lo pregunta. Así se conservan las letras acentuadas, las comillas curvas y las notas musicales. También se entiende un archivo guardado como «ANSI» de Windows (Windows-1252). Los archivos de procesador de textos (`.doc`, `.docx`, `.pdf`) no son archivos de texto: copie el texto y péguelo antes en un editor de texto sin formato, o use Guardar como y elija texto sin formato.
- Valen tanto los saltos de línea de Windows (CRLF) como los de Mac o Linux (LF), y también la marca invisible que algunos editores ponen al comienzo del archivo.
- Un `.zip` solo hace falta para una copia de seguridad completa, que incluye las imágenes. El archivo del espectáculo es solo el texto.

## Decidir dónde empieza y termina cada subtítulo

El programa nunca divide ni une sus líneas. Un subtítulo es exactamente lo que escribió en una línea, así que usted controla cuánto lee el público de una vez.

- Procure líneas que se lean cómodamente de un vistazo: una frase o parte de una frase. Los discursos muy largos deben cortarse en pausas naturales, donde el actor respira.
- Las líneas en blanco no significan nada, así que úselas con libertad entre intervenciones para que el archivo sea fácil de leer.
- Si dos o más líneas deben aparecer juntas, únalas con una barra invertida al final de la línea, o póngalas entre `[[` y `]]`.
- Use `HAMLET: línea` para mostrar quién habla. Ponga el nombre en mayúsculas y luego dos puntos.
- Describa los sonidos y las acotaciones entre paréntesis, en una línea aparte, por ejemplo `(retumban truenos)`. Se muestran al público con otro estilo.

```showfile
[section: Act 1]
[note: house lights down, then start]
GEORGE: Morning, Emily.
EMILY: Oh, hello, George.
(a train whistle, far away)
```

## Lo que solo ve el operador

Marque su lugar y deje recordatorios sin molestar al público: `[section: Acto 2]` añade un marcador en la lista del operador, y `[note: esperar los aplausos]` muestra un recordatorio. Ninguno gasta una pulsación de la barra espaciadora. Un comentario entre `<!--` y `-->` es para usted y no lo ve nadie.

## Trabajar en su propio editor

Nunca está obligado a escribir en el editor del paso Archivo del espectáculo. En cambio:

1. En el paso **Archivo del espectáculo**, pulse **Descargar el archivo del espectáculo**. Obtiene un archivo `.md` de texto sin formato.
2. Ábralo en su propio programa, edítelo y guárdelo. Conserve el archivo como texto sin formato.
3. De vuelta en el paso Archivo del espectáculo, use **Abrir un archivo** y elija el suyo. Su texto reemplaza lo que hay en el editor y la lista de problemas se actualiza al instante. Nada cambia definitivamente hasta que pulse **Guardar**.
4. Pulse **Guardar**. La versión que sustituye se conserva en **Versiones anteriores**, donde puede compararla con la nueva, así que un error siempre se puede deshacer.

Lo mismo vale para una traducción: elija primero la pestaña del idioma y luego descargue y abra el archivo de ese idioma. También puede soltar un archivo sobre el editor.

Para empezar un **espectáculo nuevo** a partir de un archivo, use **Importar un espectáculo** en el paso Espectáculo y elija el archivo. Se convierte en un espectáculo nuevo, con el nombre de la línea `title:` de su bloque `[settings]`, o con el nombre del archivo si no la hay. Las imágenes no forman parte de un archivo de texto; súbalas en el paso Imágenes, con los mismos nombres que en sus líneas `[image: ...]`.

## Comprobar su trabajo

- El paso Archivo del espectáculo enumera cada problema con su número de línea. Pulse uno para ir a él. Los errores nunca le impiden guardar, pero se omite el elemento que tiene un error, así que corríjalos antes de la función.
- El panel junto al editor muestra lo que recorrerá el operador, una fila por cada pulsación de la barra espaciadora. Compruebe que el número de filas es el que espera.
- Ejecute el paso **Prueba** para ver los subtítulos como los verá el público, en un teléfono además de en una pantalla grande.
- Si otras personas también trabajan en el archivo del espectáculo, vea *Si dos personas editan el mismo archivo* en [Escribir un archivo del espectáculo](topic:admin-writing-a-show-file).

## Antes de entregarlo

- Lea el archivo una vez de principio a fin como lo haría el público, no como un corrector.
- Compruebe que cada nombre de `[image: ...]` coincide con una imagen subida, y que cada `[loop]`, `[song]`, `[together]` y `[fullscreen]` tiene su línea de cierre.
- Entregue al operador una copia del archivo y los nombres de las secciones, para que sepa dónde encontrar su sitio. [Canciones y música](topic:admin-musicals), [Bucles](topic:admin-loops) e [Idiomas](topic:admin-languages) detallan esas partes.
