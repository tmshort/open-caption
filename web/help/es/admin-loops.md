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

- Si lo **primero** del archivo del espectáculo es un bucle, empieza solo en cuanto se inicia el espectáculo, de modo que el público ve la previa mientras busca su asiento.
- Cualquier otro bucle (por ejemplo en el descanso) empieza cuando el operador pulsa la barra espaciadora sobre él.
- La barra espaciadora termina el bucle y envía la línea siguiente. **Mayús+Q** detiene el bucle y deja la pantalla en blanco. **Atrás** vuelve al bucle y lo reinicia desde el primer elemento.

La lista del operador muestra un bucle como una sola entrada que se despliega para enumerar sus elementos. Consulte [Bucles para el operador](topic:operator-loops).
