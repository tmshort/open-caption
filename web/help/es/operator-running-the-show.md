# Dirigir el espectáculo

La página muestra el archivo del espectáculo como una lista. Las líneas ya enviadas están atenuadas, la línea **en pantalla ahora** tiene un punto verde y la línea **siguiente** es grande y lleva una flecha amarilla.

## Las tres cosas que hay que recordar

1. **Espacio** envía la línea siguiente.
2. **Retroceso** deshace lo último que envió.
3. **Mayús+Q** detiene un bucle de previa o de descanso y deja la pantalla en blanco.

Siga a los actores y pulse **Espacio** cuando se diga cada línea. Puede usar un pulsador; los botones de siguiente y anterior de un pulsador funcionan como Espacio y Retroceso.

## En blanco y Borrar

Ambos dejan vacías las pantallas del público, pero no son lo mismo:

| | **En blanco** (`B`) | **Borrar** (`C`) |
|---|---|---|
| Lo que ve el público | Nada: solo el fondo | Una zona de subtítulos vacía, lista para líneas nuevas |
| Las líneas anteriores | **Se conservan.** Pulse `B` otra vez, o envíe la línea siguiente, y siguen ahí encima | **Se borran.** La línea siguiente empieza en una pantalla vacía |
| Úselo para | Un apagón, una pausa, u ocultar los subtítulos un momento | Una escena nueva, o tras una canción, para que no se acumulen líneas viejas |
| Deshacer | `B` otra vez, o **Retroceso** | **Retroceso** |

Un archivo del espectáculo puede hacer lo mismo con los comandos `[blank]` y `[clear]`; aparecen en su lista como `PANTALLA EN BLANCO` y `BORRAR PANTALLA`.

## Palabras de esta página

- **En espera**: el espectáculo está cargado, pero el público solo ve la pantalla de espera del local.
- **En el aire**: el público ve el espectáculo. **INICIAR** (START) lo pone en el aire; **Poner en espera…** lo devuelve a la espera conservando su posición.
- **Reiniciar espectáculo…**: vuelve a la primera línea y a la espera.
- **En blanco** oculta los subtítulos pero conserva las líneas anteriores; **Borrar** las elimina (véase arriba).

## Antes del espectáculo

- Use la página del operador en un equipo con teclado (un portátil sirve). Está pensada para la tecla **Espacio**, y no se recomienda un teléfono ni una tableta: en una pantalla táctil es fácil pulsar mal durante la función.
- Inicie sesión en la página del operador. Verá «Usted tiene el control». Si dice que otro operador tiene el control, use **Asumir el control** solo si le corresponde ser el operador.
- Un segundo dispositivo (la tableta del director de escena) puede mirar desde la página de seguimiento, `/follow`, con la contraseña de espectador (o la de operador o administrador). Muestra lo mismo que usted ve, notas incluidas, pero no puede enviar nada; el paso **Compartir** de la configuración muestra su dirección.
- Compruebe el recuento de espectadores (el icono de personas). Muestra cuántos dispositivos están conectados.
- Un espectáculo cargado queda en **espera**: el público ve la pantalla de espera del local (un logotipo y un mensaje) y nada del espectáculo, y esta página dice **EN ESPERA**. Pulse **Espacio** (el botón dice **INICIAR**) para poner el espectáculo en el aire. Si el archivo del espectáculo empieza con un bucle de previa, empieza entonces.

## Durante el espectáculo

- **Ritmo:** cuando una línea tiene un `[pace]` en el archivo del espectáculo, su tiempo se muestra junto a la línea siguiente y una barra fina en el borde inferior de la franja de estado se llena mientras la línea está en pantalla. Cuando toca la siguiente pulsación, la línea siguiente recibe un marco amarillo y la barra se vuelve amarilla. Nada se envía solo y no hay sonido.
- Las líneas marcadas como notas o secciones son para usted. Nunca gastan una pulsación de Espacio.
- Las imágenes y los mensajes a pantalla completa aparecen aquí solo como una etiqueta, como `IMAGEN: telon.jpg`. Nunca ve la imagen en sí.
- **Canciones:** las líneas cantadas llevan notas musicales (♪) alrededor y se muestran en cursiva en las pantallas del público. El título de una canción aparece como marcador azul solo para usted. Intente pulsar Espacio al empezar cada frase.
- **Idiomas:** si el espectáculo tiene traducciones, usted sigue el idioma principal y todo lo que envía va a todos los idiomas a la vez. El recuento de público en la cabecera enumera cada idioma, por ejemplo `83 (en 60 · es 23)`.
- **Varias voces a la vez:** una línea marcada `A LA VEZ` muestra dos o más cantantes en las pantallas al mismo tiempo. Una pulsación de Espacio los envía todos.
- La franja bajo la cabecera muestra quién tiene el control, **Elemento n de total** y **Último envío hace …**, para ver de un vistazo que el espectáculo y su pantalla siguen sincronizados. Se pone amarillo tras dos minutos sin una línea.
- **Ir a la sección…** (junto al cuadro de búsqueda) enumera los marcadores del archivo del espectáculo; al elegir uno se selecciona, y usted confirma el salto.
- Para avanzar o retroceder, haga clic en una línea y luego confirme «Saltar aquí». También puede buscar con `/`.
- Un aviso puntual como «El espectáculo se reanudará en breve» se puede enviar con **Anuncio…**. Pulse **Atrás** para quitarlo. Mientras está en las pantallas, un cartel azul de esta página lo cita.
- **Poner en espera…** (arriba a la derecha) quita el espectáculo de las pantallas del público: vuelven a ver la pantalla de espera y el espectáculo conserva su lugar. **INICIAR** lo vuelve a poner en el aire exactamente donde estaba.

Consulte [el teclado](operator-keyboard.md) y [Si algo sale mal](operator-if-something-goes-wrong.md).

## Después del espectáculo: el informe del público

Cuando llega al final del espectáculo, el cartel dice **Fin del espectáculo.** con un enlace **Informe del público** (el paso **Informe** de la configuración lo muestra después). El informe dice cuántos dispositivos distintos usaron la página del público mientras el espectáculo estaba en el aire, sin importar cuánto se quedó cada uno: un teléfono que se cayó de la WiFi y volvió cuenta una vez. Muestra las cifras (teléfonos, tabletas, ordenadores, pantallas de proyector, el máximo conectado a la vez, cuántos siguen conectados), los idiomas y una fila por dispositivo. También dibuja un gráfico de cuántos dispositivos estaban conectados a lo largo del tiempo, con los bucles de antes del espectáculo y del intermedio marcados, para ver cuánta gente había al empezar y cuánta se quedó durante el descanso. Sigue contando hasta que vuelva a la espera, reinicie o cargue otro espectáculo, y se puede imprimir o descargar como archivo CSV para una hoja de cálculo.

Un segundo dispositivo que solo mira el espectáculo (la página de seguimiento) aparece en la misma tabla, marcado como **Espectador**, pero no se cuenta en ninguna cifra ni en el gráfico. Los dispositivos se distinguen por su dirección de red y el tipo de navegador, así que las cifras son una estimación: los teléfonos que comparten ambos cuentan como uno, y un teléfono cuya dirección cambia cuenta dos veces. Cada informe se guarda como un archivo en la carpeta `reports` de la carpeta de datos; se conservan los 30 más recientes, y puede eliminar uno con **Eliminar…** en el paso **Informe** de la configuración, o borrar los archivos allí.
