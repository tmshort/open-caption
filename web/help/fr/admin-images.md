# Images

Téléversez des images à l'étape **Images** et utilisez-les dans le fichier du spectacle avec `[image: nom.png]`.

- Types acceptés : PNG, JPG, GIF et WebP. Les autres types, y compris SVG, ne sont pas acceptés.
- Chaque fichier peut faire jusqu'à 10 Mo, mais **gardez-les sous 1 Mo environ**. Chaque appareil du public télécharge l'image entière, et 100 appareils demandant une grande image en même temps peuvent être lents. Environ 1920 pixels de large suffisent.
- Le programme ne redimensionne pas les images. Les navigateurs les réduisent pour qu'elles tiennent sur chaque écran.
- Une image **sans légende remplit tout l'écran**, d'un bord à l'autre : elle est mise à l'échelle sans jamais être rognée, et la couleur de fond apparaît dans les bandes si les proportions diffèrent. Ajoutez une légende (`[image: curtain.jpg | Act 1]`) et l'image garde une marge avec les mots en dessous.
- Les images à venir sont téléchargées en arrière-plan avant d'être affichées, elles apparaissent donc instantanément.

## Noms

Le fichier du spectacle désigne une image par son nom de fichier, comme `rideau.jpg`. Les noms avec des espaces fonctionnent. Si le fichier du spectacle utilise un nom que vous n'avez pas téléversé, l'étape Images le liste sous « Images utilisées par le fichier du spectacle mais pas encore téléversées », avec un bouton pour téléverser le bon fichier sous ce nom. Une faute légère comme `rideua.jpg` est signalée avec une suggestion.

## Remplacer et supprimer

Téléverser un fichier du même nom le remplace. Supprimer une image que le fichier du spectacle utilise encore est permis après un avertissement ; le fichier du spectacle affiche alors un avertissement « introuvable » pour elle.

## Dans une boucle

Les images d'une boucle passent à tour de rôle à l'écran : `[image: sponsors.png | 15s]` montre celle-ci pendant 15 secondes. Voir [Boucles](admin-loops.md).
