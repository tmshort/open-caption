# Musicales

Los subtítulos de un musical funcionan mejor cuando las letras se distinguen fácilmente de las líneas habladas y el operador puede seguir la música. Esta página reúne lo que funciona.

## Marcar las letras

Las letras se muestran **en cursiva entre notas musicales**, por ejemplo ♪ ¿Dónde está el amor? ♪. Elija la forma de escribirlas que más le convenga:

| Escriba | Resultado |
|---------|-----------|
| `~ Where is love? ~` | ♪ Where is love? ♪ (el `~` de cierre es opcional) |
| `OLIVER: ~ Where is love?` | lo mismo, con el nombre del personaje encima |
| `[song: Title]` … `[/song]` | cada línea de dentro es una letra; el título es solo para el operador |
| `♪ escrito o pegado ♪` | exactamente lo que escribió, sin cursiva |
| `~` solo en una línea | solo una nota: ♪, para un pasaje instrumental |

Use `[song]` para números largos, y `~` para una sola línea cantada en medio de una escena.

En Configuración, los botones **🎵 Canción**, **♪** y **♫** sobre el editor del archivo del espectáculo los insertan por usted. Los caracteres ♪ y ♫ también se pueden copiar de cualquier sitio, porque son texto normal.

## Dos cantantes a la vez (contrapunto y dúos)

Cuando cantantes distintos cantan palabras distintas **al mismo tiempo**, muéstrelos juntos en las pantallas del público con `[together]`:

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]
```

- Ambas líneas aparecen **a la vez** y el operador pulsa Espacio **una vez**.
- En una pantalla ancha las voces se colocan una al lado de la otra, cada una con una barra en su borde izquierdo y el nombre de su personaje encima; en un teléfono se apilan.
- Añada tantas voces como necesite, por ejemplo un trío o una línea hablada sobre una cantada.
- Para una línea rápida a dos voces puede escribir `ANNA: ~ I love you ~ || KRIS: ~ Me too ~` en una sola línea.
- Las voces que **entran una tras otra** (un canon, o un segundo cantante que responde) son líneas separadas normales: una pulsación de Espacio cada una. Escriba la primera voz y luego la segunda cuando entre.
- Mantenga cada voz corta. Tres o cuatro palabras por voz son más fáciles de leer a la vez que una frase entera.
- Use nombres de personaje para que el público distinga las voces.

## Consejos para buenos subtítulos de letras

- **Una línea por frase cantada.** Corte las líneas donde los cantantes respiran, para que la pulsación de Espacio del operador coincida con la música. Las líneas más cortas son más fáciles de seguir.
- **Use nombres de personaje en los dúos** y en las líneas cantadas por distintos personajes por turnos. Omítalos en un coro.
- **Pasajes instrumentales:** use `~` solo, o `(Toca la orquesta)`, para marcarlos. Una línea entre paréntesis se muestra como una acotación.
- **Mantenga separadas las líneas habladas y las cantadas.** Una línea de personaje seguida de letras con `~` se lee con claridad: el público ve cambiar las notas y la cursiva.
- **Repeticiones:** escriba las líneas cada vez. El operador pulsa Espacio una vez por línea de cualquier forma.
- **Borrar entre números:** ponga `[clear]` tras la última línea de una canción para que la escena siguiente empiece en una pantalla vacía.
- **Recordatorios para el operador:** `[note: el tempo se acelera tras el cambio de tonalidad]` se muestra solo al operador.
- **Ensaye con la música real.** El paso Prueba permite que el operador practique pulsando Espacio con la grabación del reparto.

## Un ejemplo

```showfile
[section: Act 1, Number 4]
[clear]
(The orchestra begins.)
[song: Where Is Love]
OLIVER: Where is love?
Does it fall from skies above?
[note: wait for the held note]
Is it underneath the willow tree
That I've been dreaming of?
[/song]
[clear]
MRS. CORNEY: That's quite enough of that!
```

Consulte también [Escribir un archivo del espectáculo](admin-writing-a-show-file.md).
