# Écrire un fichier du spectacle

Un fichier du spectacle est un fichier texte simple (`.txt` ou `.md`) écrit dans une forme simple de Markdown. Si vous collez un texte de théâtre ordinaire, cela fonctionne déjà : **chaque ligne est un sous-titre**, et une pression sur la barre d'espace envoie une ligne. Les lignes vides sont ignorées ; servez-vous-en pour garder le fichier lisible.

Les longues lignes passent automatiquement à la ligne à l'écran. C'est vous qui décidez où couper une longue réplique en lignes, selon le style de la représentation. Le programme ne coupe ni ne joint jamais de lignes à votre place.

Pour vérifier un fichier du spectacle, ou le préparer dans votre propre éditeur de texte, voir [Préparer un fichier de spectacle](admin-preparing-a-show-file.md).

## Mise en forme

| Écrivez | Vous obtenez |
|---------|--------------|
| `**fort**` | du texte en **gras** |
| `*aparté*` ou `_aparté_` | du texte en *italique* |
| `***very loud***` | texte en gras et en italique |
| `~~old line~~` | texte barré (une correction, ou une blague biffée) |
| `(thunder rumbles)` | une ligne entière entre parenthèses est affichée comme un bruit ou une indication de scène |
| `HAMLET: To be, or not to be` | une étiquette de personnage au-dessus de la ligne (le nom doit être en MAJUSCULES suivi de deux-points ; un groupe convient aussi : `ALICE & BOB:`, `ALL (EXCEPT BOB):`) |

**Barrer du texte.** Mettez deux tildes de chaque côté, `~~comme ceci~~`, et les mots sont affichés barrés. Il faut un caractère autre qu'une espace juste à l'intérieur de chaque paire (`~~ x ~~` reste tel quel). Une ligne chantée commence par un seul `~` suivi d'une espace, les deux ne sont donc jamais confondus ; pour commencer une ligne ordinaire par un tilde, écrivez `\~`.

Tout le reste est affiché tel quel. Les liens, listes, tableaux et titres ne sont pas pris en charge. Le HTML brut non plus : si vous tapez `<b>`, il apparaît tel quel, et la page affiche une remarque suggérant la forme Markdown. Le soulignement, les exposants et les indices n'ont pas de balisage, mais tout caractère que vous pouvez saisir fonctionne tel quel : écrivez directement H₂O, x², 1ˢᵗ ou ½.

```showfile
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Vérifier les noms des personnages

Cliquez sur **Fichier ▾ → Personnages…** au-dessus de l'éditeur pour voir tous les noms utilisés comme personnage, les plus utilisés d'abord, avec le nombre de lignes de chacun. C'est le moyen rapide de repérer un nom mal orthographié : **ALCE** avec une ligne à côté de **ALICE** avec quarante est signalé par *Vouliez-vous dire ALICE ?*. Cliquez sur un nom pour aller à sa première ligne. Un groupe comme `ALICE & BOB:` compte pour chacun de ses noms.

## Revenir à une version antérieure

À chaque enregistrement, la version remplacée est conservée (les 50 dernières pour chaque langue). Cliquez sur **Fichier ▾ → Versions antérieures…** au-dessus de l'éditeur, choisissez une heure et cliquez sur **Charger**. Cela ne fait que placer l'ancien texte dans l'éditeur ; rien ne change tant que vous ne cliquez pas sur **Enregistrer**, vous pouvez donc regarder d'abord. **Comparer** montre, ligne à ligne, ce que changerait le chargement d'une version. Une version à laquelle vous donnez un nom (dans la zone à côté, ou avec **Conserver** pour nommer ce qui est enregistré maintenant) n'est jamais oubliée, quel que soit le nombre d'enregistrements qui suivent. L'exportation n'inclut pas les versions antérieures.

## Réglages du spectacle : le titre et l'apparence

Tout ce qui concerne un spectacle hormis ses images est rangé dans le fichier du spectacle, y compris son **titre** et l'**apparence** des sous-titres. Cela va dans un bloc `[settings]`, généralement en haut, avec un `nom: valeur` par ligne :

```showfile
[settings]
title: The Garden Party
background: #000000
text color: #ffffff
speaker name color: #ffffff
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
loop seconds: 10
scrollback lines: 200
[/settings]
```

| Réglage | Ce qu'il fait |
|---------|---------------|
| `title` | Le nom du spectacle, tel qu'il est listé dans la configuration |
| `background` | La couleur de l'écran derrière les sous-titres, comme `#000000` (un nom tel que `black` ou `yellow` fonctionne aussi) |
| `text color` | La couleur des sous-titres, comme `#ffffff` |
| `speaker name color` | La couleur des noms de personnage (quand ils n'ont pas chacun une couleur) |
| `font` | `system`, `serif`, `mono`, `dyslexic` ou `hyperlegible` |
| `text size` | Un pourcentage de 50 à 300 ; 100 est normal |
| `speaker names` | `label` (une couleur), `color` (une couleur différente pour chaque personnage) ou `hidden` |
| `speaker position` | `above` (le nom au-dessus de la ligne), `beside` (sur la même ligne, avant le texte) ou `column` (dans une colonne à gauche du texte) |
| `idle message` | Les mots affichés avant l'envoi de la première ligne |
| `loop seconds` | Combien de temps chaque élément d'un `[loop]` reste quand il ne l'indique pas |
| `scrollback lines` | Combien de sous-titres un spectateur peut faire défiler vers l'arrière (de 20 à 2000) |
| `archived` | `yes` si le spectacle est archivé ; les boutons **Archiver** et **Restaurer** de la configuration l'ajoutent et le retirent |

**Vous n'avez rien de tout cela à écrire.** L'étape **Apparence** de la configuration remplit le bloc pour vous (et **Renommer** définit le titre), et tout ce que vous omettez utilise simplement une valeur par défaut. Les valeurs par défaut viennent en deux couches : le programme a des valeurs intégrées, que le fichier de réglages `config.yaml` peut modifier pour tous les spectacles (sa section `defaults`), puis chaque fichier de spectacle remplace ce qu'il définit. Le plus spécifique l'emporte : **fichier du spectacle, puis config.yaml, puis valeurs intégrées**.

Quelques règles : le fichier du spectacle d'une traduction ne peut définir que `idle message` (l'apparence se définit une fois, pour tout le spectacle, dans le fichier principal du spectacle) ; un réglage mal orthographié ou dont la valeur n'a pas de sens est signalé avec son numéro de ligne et ignoré ; et les commentaires à l'intérieur du bloc ne sont pas conservés quand la configuration le réécrit, placez donc vos commentaires au-dessus. Plusieurs blocs `[settings]` sont combinés, les valeurs postérieures l'emportant.

## Chansons et musique

Pour une comédie musicale, affichez les paroles entre des notes de musique (♪). Il y a trois façons, de la plus rapide à la plus souple :

- **Saisissez ou collez vous-même les notes.** ♪ et ♫ fonctionnent n'importe où dans le fichier du spectacle. Dans l'éditeur, les boutons **♪** et **♫** les insèrent au curseur.
- **Commencez une ligne par `~`.** La ligne est traitée comme une parole chantée : elle est affichée en italique avec un ♪ ajouté à chaque bout, et le `~` lui-même n'est pas affiché. Seul le `~` d'ouverture est nécessaire : le ♪ de fin est ajouté pour vous, et un `~` de fermeture est permis mais ne change rien. Cela fonctionne aussi après le nom d'un personnage.
- **Entourez toute une chanson de `[song]` et `[/song]`.** Chaque ligne à l'intérieur est chantée, vous n'avez donc pas besoin d'un `~` sur chacune. Le titre après les deux-points est un signet réservé à l'opérateur. Les notes, clear et blank peuvent toujours figurer à l'intérieur.

```showfile
OLIVER: ~ Where is love? ~
~ Row, row, row your boat

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[note: hold the last chord]
[/song]
```

### Deux voix ou plus à la fois (duos et contrepoint)

Quand des chanteurs chantent des lignes différentes **en même temps**, placez-les dans `[together]` … `[/together]`. Toutes les lignes à l'intérieur sont affichées à la fois, côte à côte sur un grand écran et empilées sur un téléphone, et l'opérateur les envoie d'une seule pression sur Espace. Chacune garde son propre nom de personnage. Pour un court échange, vous pouvez écrire les voix sur une seule ligne, séparées par ` || ` (une espace, deux barres verticales, une espace). `[duet]` et `[counterpoint]` ont le même sens que `[together]`.

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

Les voix qui s'ajoutent **l'une après l'autre** sont des lignes séparées : appuyez sur Espace pour chacune. N'utilisez `[together]` que pour des lignes qui apparaissent au même instant. Un bloc `[together]` ne contient que des lignes de texte, en exige au moins deux, et dans un `[song]` toutes les voix sont chantées.

Un tilde au milieu d'une ligne, comme `about ~5 minutes`, n'est que du texte. Pour commencer une ligne ordinaire par un tilde, écrivez `\~`. Voir [Comédies musicales](admin-musicals.md) pour d'autres conseils.

## Plusieurs lignes dans un sous-titre

Terminez une ligne par une barre oblique inverse pour la joindre à la suivante, ou placez les lignes entre `[[` et `]]`.

```showfile
GHOST: Remember me. \
Remember me.

[[
Row, row, row your boat,
Gently down the stream
]]
```

## Commentaires et notes

- `<!-- comme ceci -->` est un **commentaire**. Personne ne le voit, et il peut s'étendre sur plusieurs lignes.
- `[note: texte]` est un rappel montré **à l'opérateur seul**, par exemple `[note: hold for applause]`.
- `[section: Act 2, Scene 1]` ajoute un signet dans la liste de l'opérateur. Les notes et les sections ne consomment jamais une pression sur la barre d'espace.
- `[pace: 8s]` indique combien de temps la ligne suivante reste d'habitude à l'écran, comme rappel **pour l'opérateur seul**. La page de conduite l'affiche à côté de cette ligne (`⏱ about 8 s`) et, une fois la ligne envoyée, décompte jusqu'au moment où la pression suivante est due, en devenant jaune quand elle l'est. Rien n'est jamais envoyé automatiquement. Le temps est supérieur à 0 seconde. Il appartient à la prochaine ligne que l'opérateur envoie (les sections et les notes sont sautées) et ne peut pas aller dans une boucle ni dans un bloc `[together]`. Si le temps est bien plus long que la lecture de la ligne, la liste des problèmes ajoute une remarque, au cas où ce serait une faute de frappe. Le bouton **Rythme** de l'éditeur en insère un.
- Une ligne qui commence par `#` et une espace est un **titre** et fonctionne comme `[section: …]` : `# Act 2` (ou `## Scene 1`, avec jusqu'à six `#`) ajoute un signet dans la liste de l'opérateur et n'est jamais montré au public. Un `#` sans espace derrière (un hashtag, par exemple) est ignoré, avec une remarque ; pour commencer un sous-titre par `#`, écrivez `\#`. Les titres ne peuvent pas aller dans une boucle.

```showfile
<!-- Act 1 captions, version 3 -->
[section: Act 1, Scene 1]
[note: lights come up slowly, wait for the first line]
HAMLET: Who's there?
```

## Commandes

Une ligne entre crochets est une commande. Les noms ne distinguent pas les majuscules des minuscules.

| Commande | Ce qu'elle fait |
|----------|-----------------|
| `[clear]` | Efface les sous-titres de l'écran |
| `[blank]` | Vide l'écran (les sous-titres sont conservés) |
| `[image: file.png]` | Affiche une image en plein écran. Ajoutez une légende avec `[image: file.png \| text]` |
| `![caption](file.png)` | Pareil, écrit à la manière de Markdown |
| `[fullscreen: text]` | Affiche un grand texte en plein écran. Pour plusieurs lignes, utilisez `[fullscreen]` … `[/fullscreen]` |
| `[loop]` … `[/loop]` | Un ensemble répété d'éléments plein écran pour l'avant-spectacle ou l'entracte. Voir [Boucles](admin-loops.md) |
| `[together]` … `[/together]` | Deux voix ou plus montrées en même temps. Voir [Chansons et musique](admin-musicals.md) |
| `[song: title]` … `[/song]` | Une chanson : chaque ligne à l'intérieur est affichée comme une parole chantée. Voir [Chansons et musique](admin-musicals.md) |
| `[section: name]` | Un signet pour l'opérateur |
| `[note: text]` | Une note pour l'opérateur |
| `[pace: 8s]` | Un rappel de rythme pour l'opérateur (voir plus haut) |
| `[language: es]` | Marque un fichier de spectacle traduit. Voir [Langues](admin-languages.md) |
| `[settings]` … `[/settings]` | Le titre et l'apparence du spectacle. Voir *Réglages du spectacle* plus haut |

Les options de durée n'ont de sens que dans une [boucle](admin-loops.md) : `| 15s` après une image ou un texte plein écran, et `[blank: 3s]`. En dehors d'une boucle, une image ou un texte plein écran reste affiché jusqu'à ce que l'opérateur passe à la suite, la durée est ignorée et signalée par un avertissement, tout comme une durée sur `[blank]`.

```showfile
[clear]
(Lights up on the garden.)
[image: garden.png | The garden at dusk]
[fullscreen: **Please remain seated**]
[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
[blank]
```

Une image, un texte plein écran ou un écran vide reste sur les écrans du public jusqu'à ce que l'opérateur envoie la ligne suivante.

## Caractères spéciaux

Placez une barre oblique inverse avant un caractère pour l'afficher tel quel au lieu de le traiter comme de la mise en forme : `\*`, `\_`, `\#`, `\[`, `\]`, `\:`, `\(`, `\)`, `\!`, `\~` et `\\` (une barre oblique inverse elle-même). Par exemple, un sous-titre qui doit commencer par un dièse s'écrit `\# 1 hit`, et `HAMLET\: not a speaker` affiche les deux-points sans créer d'étiquette de personnage.

```showfile
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## Ce que signifient les avertissements

L'étape Fichier du spectacle liste les problèmes avec leurs numéros de ligne. Cliquez sur l'un d'eux pour aller à la ligne.

- **« Commande inconnue [x] »** — une ligne entre crochets qui n'est pas une commande. Si elle devait être un sous-titre tel qu'un bruitage, écrivez-la avec des parenthèses, comme `(laughter)`.
- **« Image … introuvable »** — l'image n'est pas téléversée, ou le nom est orthographié autrement. Téléversez-la à l'étape Images.
- **« [loop] a été ouvert mais jamais fermé »** — ajoutez `[/loop]` après le dernier élément. Même chose pour les blocs `[fullscreen]`, les blocs `[[` et les commentaires.
- **« Le HTML n'est pas pris en charge »** — utilisez `**gras**` et `*italique*` à la place des balises.
- **« ne peut pas aller dans une boucle »** — les boucles ne contiennent que des images et du texte.
- **« [together] a besoin d'au moins deux lignes »** — un bloc `[together]` d'une seule ligne est montré comme un sous-titre ordinaire. **« [together] a été ouvert mais jamais fermé »** — ajoutez `[/together]` ; les lignes que vous avez écrites sont conservées.
- **« Ce sous-titre affiche 5 voix à la fois »** — cinq voix ou plus dans un bloc `[together]` (ou une ligne avec ` || `) font beaucoup pour un téléphone, qui les empile, et pour un écran partagé, qui a de la place pour quatre environ et ne défile jamais. Divisez-le en deux sous-titres.
- **« Réglage inconnu » / « n'est pas une police » / « n'est pas une couleur »** — une ligne du bloc `[settings]` que la configuration ne peut pas utiliser ; elle est ignorée et la valeur par défaut s'applique. Le message dit quoi écrire à la place.
- **« [settings] a été ouvert mais jamais fermé »** — ajoutez `[/settings]` ; tout ce qui suit un bloc non fermé est lu comme des réglages.
- **« [song] a été ouvert mais jamais fermé »** — ajoutez `[/song]` après la dernière parole. Les paroles que vous avez écrites sont tout de même conservées.

Les erreurs ne vous empêchent jamais d'enregistrer.

## Si deux personnes modifient le même fichier du spectacle

Si quelqu'un d'autre enregistre le fichier du spectacle pendant que vous l'avez ouvert, le programme n'écrasera pas discrètement son travail. Quand vous enregistrez, il vous dit que le fichier du spectacle a été modifié par quelqu'un d'autre et vous laisse **utiliser sa version** (en remplaçant ce qui est dans votre éditeur) ou **enregistrer la vôtre quand même**. Copiez ce que vous voulez garder avant de choisir. La partie qui contient l'erreur est ignorée, et le reste du fichier du spectacle fonctionne toujours.
