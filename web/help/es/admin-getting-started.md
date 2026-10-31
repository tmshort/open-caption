# Primeros pasos

Preparar un espectáculo lleva seis pasos (un séptimo, **Informe**, es para después del espectáculo). La fila de botones de la parte superior de la página los sigue en orden, y aparece una marca cuando un paso parece completo. Puede saltar a cualquier paso en cualquier momento.

1. **Espectáculos** — cree un espectáculo (o el espectáculo de ejemplo para echar un vistazo).
2. **Archivo del espectáculo** — pegue o abra su archivo del espectáculo y corrija lo que la página señale.
3. **Imágenes** — suba las imágenes que use el archivo del espectáculo.
4. **Aspecto** — elija los colores y el tamaño del texto.
5. **Prueba** — ensaye en privado antes de que haya nadie en la sala.
6. **Compartir** — comparta la dirección con el público y abra la página del operador, donde se inicia el espectáculo. Opcionalmente, la página de seguimiento (`/follow`) permite que un segundo dispositivo mire sin poder enviar nada.
7. **Informe** — después del espectáculo: cuántos dispositivos usaron la página del público, con un gráfico a lo largo del tiempo.

La primera vez que se abre Configuración sin ningún espectáculo, una breve **visita** señala cada paso y los botones de la cabecera; pulse Esc para omitirla. Para verla otra vez, abra la Ayuda y elija **Recorrer este paso** en el paso Espectáculos. En cualquier otro paso, el mismo botón da una breve visita de lo que hay en él.

## En espera y en el aire

Al abrir la página del operador, el espectáculo se carga **en espera**: el público ve la pantalla de espera del local (un logotipo y un mensaje, que se fijan una sola vez en **Ajustes → Pantalla del público**) y nada del espectáculo. El operador pulsa **Espacio** (el botón dice **INICIAR**) para poner el espectáculo en el aire, y **Poner en espera…** lo retira de nuevo sin perder su lugar. La preparación, el ensayo y la edición nunca están delante del público mientras el espectáculo está en espera, y un espectáculo en espera se puede eliminar.

## Dónde se guarda todo

Un espectáculo es una carpeta que contiene su **archivo del espectáculo** (un archivo de texto sencillo, `show.md`), sus traducciones opcionales y sus imágenes. El archivo del espectáculo contiene todo sobre el espectáculo excepto las imágenes: su título, su aspecto y las líneas. Los ajustes que el archivo del espectáculo no menciona vienen de `config.yaml` y, después, de los valores integrados del programa.

## ¿Ha terminado con un espectáculo?

Renombrar, Duplicar, Exportar, Archivar y Eliminar están en el menú **Más ▾** de cada espectáculo. En el paso Espectáculos, **Archivar** guarda un espectáculo aparte: se conserva completo pero aparece en *Espectáculos archivados* y no puede salir en directo. **Restaurar** lo devuelve. **Eliminar** lo borra definitivamente (Expórtelo antes si puede querer recuperarlo). El espectáculo que está en emisión no se puede archivar ni eliminar.

## Antes de empezar

- Este equipo debe estar en la misma red WiFi que los teléfonos y tabletas del público. Consulte [Red y conexión](admin-network.md).
- La contraseña de administrador y la de operador se guardan en `config.yaml` solo como resúmenes; se mostraron en la consola, una sola vez, la primera vez que arrancó el programa. Para poner otras nuevas, use **Ajustes → Contraseñas**. La página del operador tiene su propia contraseña para que pueda dársela a otra persona sin darle acceso a la configuración. En el equipo que ejecuta el programa, Configuración y la página del operador se abren **sin contraseña** cuando la dirección empieza por `localhost`, `127.0.0.1` o `[::1]` (se confía en la persona que está en ese equipo como administrador). Los demás equipos, y cualquier petición retransmitida por un proxy, siguen entrando con contraseña. Ponga `access.local_auto_login: false` en `config.yaml` si otras personas pueden iniciar sesión en este equipo.
- Configuración, la página del operador y esta Ayuda están disponibles en inglés, español, francés, alemán, italiano y portugués. Use el menú de idioma de la parte superior de cada página; la elección se recuerda en ese navegador.
- Su trabajo se guarda en este equipo. Use **Exportar** (menú **Más ▾** de un espectáculo) en el paso Espectáculos para hacer una copia de seguridad que pueda llevar a otro equipo. La exportación es un `.zip` con `show.md` (el archivo del espectáculo, con su título y aspecto), un `show.<idioma>.md` por cada traducción y una carpeta `assets` con las imágenes; **Importar** convierte un zip así en un espectáculo nuevo, y también acepta un archivo de espectáculo de texto sin formato (`.md` o `.txt`, con cualquier nombre) como espectáculo nuevo.

## ¿Es nuevo con los archivos del espectáculo?

Un archivo del espectáculo es solo texto. Cada línea es un subtítulo, y una pulsación de la barra espaciadora envía una línea. Para todo el trabajo, del guion al archivo terminado (incluido editar en su propio programa y traer el archivo de vuelta), lea [Preparar un archivo del espectáculo](admin-preparing-a-show-file.md). Las marcas están enumeradas en [Escribir un archivo del espectáculo](admin-writing-a-show-file.md). O cree el espectáculo de ejemplo y vea cómo está escrito. Hay un archivo de partida que puede descargar en [sample-show.md](../../sample/sample-show.md).

## Planificar el equipo

Para saber qué comprar, dónde poner el punto de acceso WiFi y una lista de comprobación para el día de la función, lea la [guía de uso](../../../docs/i18n/es/usage-guide.md). La [guía de red del local](../../../docs/i18n/es/venue-network-guide.md) es para quien se ocupa de la WiFi del edificio.

## Adónde ir después

- [Musicales](admin-musicals.md)
- [Escribir un archivo del espectáculo](admin-writing-a-show-file.md)
- [Bucles: antes del espectáculo y descanso](admin-loops.md)
- [Resolución de problemas](admin-troubleshooting.md)
