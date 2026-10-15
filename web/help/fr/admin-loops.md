# Boucles : avant le spectacle et entracte

Une **boucle** est un ensemble d'images et de textes en plein écran qui se répète jusqu'à ce que l'opérateur passe à la suite. Utilisez-en une avant le début du spectacle et une autre à l'entracte.

```showfile
[loop: 10s]
[image: sponsors.png | 15s]
**Welcome!**
Please silence your phones
[/loop]
```

Chaque ligne à l'intérieur de la boucle est un élément :

- une simple ligne de texte s'affiche comme texte plein écran ;
- `[image: fichier.png]`, `![légende](fichier.png)` et `[fullscreen: texte]` fonctionnent aussi ;
- `[blank: 3s]` est une courte pause sur un écran vide ;
- ajoutez `| 15s` à la fin d'un élément pour changer sa durée d'affichage ;
- `[loop: 10s]` fixe la durée des éléments qui n'en indiquent pas. Si vous n'en donnez aucune, la valeur par défaut de `config.yaml` (10 secondes) est utilisée.

Les sous-titres, `[clear]`, les sections, les notes et les autres boucles ne peuvent pas figurer dans une boucle.

## Quand une boucle démarre et s'arrête

- Si la **première chose** du fichier du spectacle est une boucle, elle démarre toute seule dès que le spectacle est lancé : le public voit l'avant-spectacle pendant qu'il s'installe.
- Toute autre boucle (par exemple à l'entracte) démarre quand l'opérateur appuie sur la barre d'espace dessus.
- La barre d'espace termine la boucle et envoie la ligne suivante. **Maj+Q** arrête la boucle et laisse l'écran vide. **Retour** revient à la boucle et la redémarre au premier élément.

La liste de l'opérateur montre une boucle comme une seule entrée qui se déplie pour lister ses éléments. Voir [Les boucles pour l'opérateur](topic:operator-loops).
