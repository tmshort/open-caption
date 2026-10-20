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

- Si la **première chose** du fichier du spectacle est une boucle, elle démarre quand l'opérateur met le spectacle à l'antenne (**DÉMARRER**) : le public voit donc l'avant-spectacle pendant que les gens cherchent leur place.
- Toute autre boucle (par exemple à l'entracte) démarre quand l'opérateur appuie sur la barre d'espace dessus.
- La barre d'espace termine la boucle et envoie la ligne suivante. **Maj+Q** arrête la boucle et laisse l'écran vide. **Retour** revient à la boucle et la redémarre au premier élément.

La liste de l'opérateur montre une boucle comme une seule entrée qui se déplie pour lister ses éléments. Voir [Les boucles pour l'opérateur](topic:operator-loops).

## Un minuteur de compte à rebours

Donnez un **minuteur** à une boucle et les écrans du public peuvent afficher le temps qui reste, décompté en direct (« Retour dans 15 min »). Écrivez `timer` dans les options de la boucle, puis placez `{timer}` dans un texte de cette boucle à l'endroit où le temps doit apparaître.

```showfile
[loop: timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
![Refreshments are in the lobby](lobby.png)
[/loop]
```

La valeur d'un minuteur est l'une de ces deux choses :

- Un **nombre de minutes**, comme `timer 15` ou `timer 15 minutes` : pour un entracte. Le décompte commence au moment où la boucle démarre (quand l'opérateur appuie sur la barre d'espace dessus) et recommence avec la durée complète si l'opérateur revient dans la boucle.
- Une **heure**, comme `timer 7:30pm`, `timer 7pm` ou `timer 19:30` : pour l'avant-spectacle. Les écrans décomptent jusqu'à cette heure aujourd'hui. L'horloge et le fuseau horaire de l'ordinateur qui exécute le programme sont utilisés : vérifiez-les. Une heure déjà passée affiche 0:00.

Les options partagent les crochets, séparées par des virgules, dans n'importe quel ordre.

```showfile
[loop: 10s, timer 7:30pm]
**Welcome!**
The show begins in {timer}
[/loop]
```

- Le temps restant est affiché en minutes, arrondi à la hausse et sans secondes (`15 min`), avec les heures s'il y en a plus d'une (`1 h 5 min`), dans la langue de l'appareil. À zéro, il reste à `0 min` ; la boucle continue jusqu'à ce que l'opérateur passe à la suite.
- Seuls les éléments dont le texte contient `{timer}` l'affichent : vous pouvez donc alterner un message et le décompte.
- Un téléphone dont l'horloge est fausse affiche quand même le bon temps restant, car c'est l'horloge du programme qui est utilisée.
- Enregistrer le fichier du spectacle pendant un entracte ne relance pas le décompte, pas plus que le redémarrage du programme.
- Dans une traduction, la boucle doit avoir le même minuteur que le fichier principal du spectacle ; traduisez les mots autour de `{timer}`.
- L'opérateur voit aussi le temps restant dans le bandeau vert. Un `{timer}` dans une boucle sans minuteur est affiché tel quel, avec un avertissement.

## Donner un nom à une boucle

Donnez un **nom** à une boucle et le rapport sur le public s'en sert pour l'étiqueter sur son graphique. Sans nom, le rapport devine d'après la place de la boucle (avant la première ligne : *avant-spectacle* ; entre des lignes : *entracte* ; après la dernière ligne : *après-spectacle*).

```showfile
[loop: 12s, name Intermission]
[fullscreen: **INTERMISSION**]
Back soon
[/loop]
```

- Le nom est un texte libre d'au plus 40 caractères, sans virgule. Le public ne le voit jamais.
- Il partage les crochets avec les autres options, séparées par des virgules, dans n'importe quel ordre : `[loop: 10s, name Act 1 break, timer 15 minutes]`.
- Seul le nom du fichier de spectacle de la langue principale est utilisé ; une traduction peut donner d'autres noms à ses boucles, ou aucun.
