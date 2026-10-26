# Resolución de problemas

## El botón Solucionar problemas

El botón **Solucionar problemas**, arriba en Configuración, junto a Configuración, es el lugar por donde empezar cuando la gente no puede conectarse o algo parece ir mal. Muestra lo que el programa sabe: si este ordenador tiene una dirección de red que dar, qué puerto obtuvo realmente el servidor web, si HTTPS y su certificado están en orden, si el nombre `.local` se anuncia, si se puede guardar el archivo de configuración, si cada idioma coincide con el archivo principal del show y cuántos dispositivos están conectados. Cada comprobación se marca **OK**, **Nota**, **Revisar** o **Problema**, y las que se pueden arreglar tienen un botón que lleva al ajuste correcto. Debajo están los últimos avisos y errores del registro, un enlace al registro completo y el paquete de soporte.

Cuando el programa registra avisos o errores nuevos, aparece un número en el botón hasta que lo abra.

## Nadie puede abrir la dirección

- Compruebe que el teléfono está en la WiFi de invitados del local, no en datos móviles.
- Compruebe que la dirección coincide con la del paso **Compartir**. Si lleva número de puerto, inclúyalo.
- Pregunte al local si los dispositivos de la WiFi de invitados pueden llegar a este equipo.
- El cortafuegos de su equipo puede estar bloqueando el programa. En Windows, permítalo en **redes privadas** cuando se lo pida. En macOS, permita las conexiones entrantes para él.
- Abra la dirección en este equipo usando `localhost` para comprobar el propio programa.

## El número de dispositivos conectados es menor de lo esperado

Los dispositivos que se durmieron se desconectan y se reconectan solos al despertar. Pida a la gente que desactive el bloqueo automático durante el espectáculo si su pantalla se atenúa.

## Las imágenes aparecen despacio

Las imágenes grandes tardan en una WiFi ocupada. Busque en el paso Imágenes los archivos marcados como «grande» y sustitúyalos por otros más pequeños.

Para la lista completa de comprobaciones del lado de la red, consulte la [guía de red del local](topic:admin-venue-network-guide).

## Falta un idioma para el público

Abra el paso Archivo del espectáculo y mire la pestaña de ese idioma: un ⚠ significa que su archivo del espectáculo ya no tiene las mismas señales que el principal. Corrija el error que se muestra; el idioma vuelve en cuanto guarde. Consulte [Idiomas](topic:admin-languages).

## Problemas del archivo del espectáculo

Consulte «Qué significan los avisos» en [Escribir un archivo del espectáculo](topic:admin-writing-a-show-file).

## Contraseñas

Para cambiar cualquiera de las dos contraseñas, use **Ajustes** en la parte superior de la página de Configuración. La nueva contraseña surte efecto de inmediato y se guarda en `config.yaml`. Cualquiera que haya iniciado sesión con una contraseña que cambió cierra sesión; usted sigue conectado. La misma ventana puede mostrar la contraseña de operador actual para que se la dé al operador.

Si no puede iniciar sesión de ninguna manera, las dos contraseñas están en `config.yaml`, en la carpeta de datos del programa. Abra el archivo con cualquier editor de texto. Su ubicación se imprime cuando arranca el programa.

## Mirar el registro

**Compartir → Registro reciente…** muestra lo que el programa ha registrado últimamente. Elija *Solo problemas* para ver avisos y errores, como inicios de sesión rechazados o un nombre ya usado en la red. La misma ventana define cuánto detalle se registra (**Nivel de detalle del registro**); *Depuración* añade cada petición, lo que ayuda cuando ocurre algo raro. Las contraseñas nunca se escriben en el registro.

## Pedir ayuda

Descargue el **paquete de soporte** desde la parte inferior del paso Compartir. Contiene los ajustes (sin contraseñas), líneas recientes del registro e información básica sobre este equipo, y es fácil de enviar a quien mantenga el programa.
