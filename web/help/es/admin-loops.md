# Bucles: antes del espectáculo y descanso

Un **bucle** es un conjunto de imágenes y textos a pantalla completa que se repite hasta que el operador continúa. Use uno antes de que empiece el espectáculo y otro en el descanso.

```showfile
[loop: 10s]
[image: sponsors.png | 15s]
**Welcome!**
Please silence your phones
[/loop]
```

Cada línea dentro del bucle es un elemento:

- una línea de texto sencilla se muestra como texto a pantalla completa;
- también sirven `[image: archivo.png]`, `![leyenda](archivo.png)` y `[fullscreen: texto]`;
- `[blank: 3s]` es una breve pausa con la pantalla en blanco;
- añada `| 15s` al final de un elemento para cambiar cuánto tiempo permanece;
- `[loop: 10s]` fija el tiempo de los elementos que no lo indican. Si no da ningún tiempo, se usa el valor por defecto de `config.yaml` (10 segundos).

Los subtítulos, `[clear]`, las secciones, las notas y otros bucles no pueden ir dentro de un bucle.

## Cuándo empieza y termina un bucle

- Si lo **primero** del archivo del espectáculo es un bucle, empieza cuando el operador pone el espectáculo en el aire (**INICIAR**), así que el público ve la previa mientras la gente busca su asiento.
- Cualquier otro bucle (por ejemplo en el descanso) empieza cuando el operador pulsa la barra espaciadora sobre él.
- La barra espaciadora termina el bucle y envía la línea siguiente. **Mayús+Q** detiene el bucle y deja la pantalla en blanco. **Atrás** vuelve al bucle y lo reinicia desde el primer elemento.

La lista del operador muestra un bucle como una sola entrada que se despliega para enumerar sus elementos. Consulte [Bucles para el operador](topic:operator-loops).

## Un temporizador de cuenta atrás

Dé un **temporizador** a un bucle y las pantallas del público pueden mostrar cuánto tiempo queda, contando hacia atrás en directo («Volvemos en 15 min»). Escriba `timer` en las opciones del bucle y ponga `{timer}` en cualquier texto de ese bucle donde deba aparecer el tiempo.

```showfile
[loop: timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
![Refreshments are in the lobby](lobby.png)
[/loop]
```

El valor de un temporizador es una de dos cosas:

- Un **número de minutos**, como `timer 15` o `timer 15 minutes`: para un descanso. La cuenta atrás empieza en el momento en que empieza el bucle (cuando el operador pulsa la barra espaciadora sobre él) y vuelve a empezar con todo el tiempo si el operador retrocede hasta el bucle.
- Una **hora del día**, como `timer 7:30pm`, `timer 7pm` o `timer 19:30`: para el antes del espectáculo. Las pantallas cuentan hacia atrás hasta esa hora de hoy. Se usa el reloj y la zona horaria del equipo que ejecuta el programa, así que compruébelos. Una hora que ya pasó muestra 0:00.

Las opciones comparten los corchetes, separadas por comas, en cualquier orden.

```showfile
[loop: 10s, timer 7:30pm]
**Welcome!**
The show begins in {timer}
[/loop]
```

- El tiempo que queda se muestra en minutos, redondeado hacia arriba y sin segundos (`15 min`), con horas cuando hay más de una (`1 h 5 min`), en el idioma del dispositivo. En cero se queda en `0 min`; el bucle sigue hasta que el operador continúe.
- Solo muestran el temporizador los elementos cuyo texto contiene `{timer}`, así que puede alternar un mensaje con la cuenta atrás.
- Un teléfono con el reloj mal puesto muestra igualmente el tiempo correcto, porque se usa el reloj del programa.
- Guardar el archivo del espectáculo durante un descanso no reinicia la cuenta atrás, ni tampoco reiniciar el programa.
- En una traducción, el bucle debe tener el mismo temporizador que el archivo principal del espectáculo; traduzca las palabras que rodean a `{timer}`.
- El operador también ve el tiempo que queda en el cartel verde. Un `{timer}` en un bucle sin temporizador se muestra tal como está escrito, con una advertencia.

## Poner nombre a un bucle

Dé un **nombre** a un bucle y el informe del público lo usa para rotular ese bucle en su gráfico. Sin nombre, el informe lo adivina por el lugar del bucle (antes de la primera línea: *antes del espectáculo*; entre líneas: *intermedio*; después de la última línea: *después del espectáculo*).

```showfile
[loop: 12s, name Intermission]
[fullscreen: **INTERMISSION**]
Back soon
[/loop]
```

- El nombre es texto libre de hasta 40 caracteres, sin comas. El público nunca lo ve.
- Comparte los corchetes con las demás opciones, separadas por comas y en cualquier orden: `[loop: 10s, name Act 1 break, timer 15 minutes]`.
- Solo se usa el nombre del archivo del espectáculo del idioma principal; una traducción puede dar otros nombres a sus bucles o ninguno.
