# Languages

*This feature is planned. The script format and the pages already reserve a place for it, so scripts you write today will keep working.*

To offer captions in more than one language you will keep a **separate script file for each language**, all with exactly the same lines in the same order. Start each translated script with a line naming its language:

```script
[language: es | Español]
Hola, Margaret.
```

When the operator presses the spacebar, every language moves to the same line together. Audience devices pick their language from the browser's settings, can switch with a small menu, and can be sent a link ending in `?lang=es`. When nothing matches, English is used.
