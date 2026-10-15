# Imágenes

Suba imágenes en el paso **Imágenes** y úselas en el archivo del espectáculo con `[image: nombre.png]`.

- Tipos admitidos: PNG, JPG, GIF y WebP. No se aceptan otros tipos, incluido SVG.
- Cada archivo puede pesar hasta 10 MB, pero **manténgalos por debajo de 1 MB aproximadamente**. Cada dispositivo del público descarga la imagen completa, y 100 dispositivos pidiendo una imagen grande a la vez pueden ir lentos. Con unos 1920 píxeles de ancho basta.
- El programa no redimensiona las imágenes. Los navegadores las reducen para ajustarlas a cada pantalla.
- Las imágenes próximas se descargan en segundo plano antes de mostrarse, así que aparecen al instante.

## Nombres

El archivo del espectáculo se refiere a una imagen por su nombre de archivo, como `telon.jpg`. Los nombres con espacios funcionan. Si el archivo del espectáculo usa un nombre que no ha subido, el paso Imágenes lo enumera en «Imágenes que usa el archivo del espectáculo pero que no ha subido», con un botón para subir el archivo correcto con ese nombre. Un error leve como `telom.jpg` se señala con una sugerencia.

## Reemplazar y eliminar

Subir un archivo con el mismo nombre lo reemplaza. Eliminar una imagen que el archivo del espectáculo aún usa está permitido tras un aviso; el archivo del espectáculo mostrará entonces un aviso de «no encontrada» para ella.

## En un bucle

Las imágenes de un bucle se turnan en pantalla: `[image: patrocinadores.png | 15s]` muestra esa durante 15 segundos. Consulte [Bucles](topic:admin-loops).
