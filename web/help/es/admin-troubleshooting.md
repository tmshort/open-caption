# Resolución de problemas

## El botón Solucionar problemas

El botón **Solucionar problemas**, arriba en la pantalla de Configuración, junto a Ajustes, es el lugar por donde empezar cuando la gente no puede conectarse o algo parece ir mal. Se abre en tres pestañas: **Comprobaciones**, **Prueba de conexión** y **Avisos y errores**. Las comprobaciones muestran lo que el programa sabe: si este ordenador tiene una dirección de red que dar, si se ofrecen direcciones de más de una red, qué puertos (HTTP y HTTPS) obtuvo realmente el servidor web, si HTTPS y su certificado están en orden, si el nombre `.local` se anuncia, si se puede guardar el archivo de configuración, si cada idioma coincide con el archivo principal del show si una lista de acceso rechaza ordenadores (o deja abierta la página Seguir), si el reloj del ordenador coincide con el del navegador, si el archivo del espectáculo cargado tiene errores o falta una imagen, si se puede escribir en la carpeta de datos y tiene espacio, si las contraseñas son iguales o están en texto sin cifrar, si el espectáculo está en espera o en el aire (lo que ve el público ahora) y cuántos dispositivos están conectados. Cada comprobación se marca **OK**, **Nota**, **Revisar** o **Problema**, y las que se pueden arreglar tienen un botón que lleva al ajuste correcto. Debajo están los últimos avisos y errores del registro, un enlace al registro completo, **Copiar resumen** (las comprobaciones como texto, para pegar en un correo) y el paquete de soporte. **Probar la conexión** pide el servidor web en cada dirección que ofrece el programa y dice cuáles responden y con qué rapidez; un cortafuegos de este ordenador o de la red aún puede bloquear a otros dispositivos.

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

Para la lista completa de comprobaciones del lado de la red, consulte la [guía de red del local](../../../docs/i18n/es/venue-network-guide.md).

## Falta un idioma para el público

Abra el paso Archivo del espectáculo y mire la pestaña de ese idioma: un ⚠ significa que su archivo del espectáculo ya no tiene las mismas señales que el principal. Corrija el error que se muestra; el idioma vuelve en cuanto guarde. Consulte [Idiomas](admin-languages.md).

## Problemas del archivo del espectáculo

Consulte «Qué significan los avisos» en [Escribir un archivo del espectáculo](admin-writing-a-show-file.md).

## Contraseñas

Para cambiar cualquiera de las dos contraseñas, use **Ajustes** en la parte superior de la página de Configuración. La nueva contraseña surte efecto de inmediato y se guarda en `config.yaml`. Cualquiera que haya iniciado sesión con una contraseña que cambió cierra sesión; usted sigue conectado. Las contraseñas se guardan como resúmenes, así que la configuración no puede mostrar ninguna: anote la contraseña al ponerla. Una **contraseña de espectador** opcional deja que un segundo dispositivo, como la tableta de un regidor, abra la página de seguimiento (`/follow`) solo para mirar: muestra las líneas, las notas y la posición, y desde ella no se puede enviar nada. Las contraseñas de operador y de administrador también abren esa página.

Si no puede iniciar sesión de ninguna manera, las dos contraseñas están en `config.yaml`, en la carpeta de datos del programa. Abra el archivo con cualquier editor de texto. Su ubicación se imprime cuando arranca el programa.

## Mirar el registro

**Solucionar problemas → Avisos y errores** muestra lo que el programa ha registrado últimamente; se abre en *Solo problemas* (avisos y errores, como inicios de sesión rechazados o un nombre ya usado en la red), y el menú de nivel lo amplía a *Todo*. **Ajustes → Acerca de** define cuánto detalle se registra (**Nivel de detalle del registro**); *Depuración* añade cada petición, lo que ayuda cuando ocurre algo raro. Las contraseñas nunca se escriben en el registro.

## Pedir ayuda

Descargue el **paquete de soporte** con el botón **Solucionar problemas** de la parte superior de Configuración. Contiene los ajustes (sin contraseñas), líneas recientes del registro e información básica sobre este equipo, y es fácil de enviar a quien mantenga el programa.
