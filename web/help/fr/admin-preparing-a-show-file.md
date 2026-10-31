# Préparer un fichier de spectacle

Cette page s'adresse à la personne qui transforme le texte de la pièce en sous-titres. Vous n'avez pas à conduire le spectacle ni à connaître quoi que ce soit du réseau. Un fichier de spectacle est un fichier texte ordinaire : vous pouvez l'écrire avec le programme de votre choix, sur n'importe quel ordinateur, et l'apporter dans la Configuration quand il est prêt. [Écrire un fichier de spectacle](admin-writing-a-show-file.md) énumère tous les marquages possibles ; cette page explique comment s'y prendre.

## En bref

1. Mettez le texte dans un fichier texte brut. Partez du spectacle d'exemple à [sample-show.md](../../sample/sample-show.md) si vous voulez.
2. Faites en sorte qu'**une ligne du fichier soit un sous-titre** : un appui sur la barre d'espace.
3. Ajoutez ce qui aide le public et l'opérateur (noms des personnages, descriptions de sons, sections, notes).
4. Apportez le fichier dans la Configuration, lisez la liste des problèmes, corrigez-les et enregistrez.
5. Répétez à l'étape Test et corrigez ce que vous remarquez.

## Quel type de fichier

- Tout fichier texte brut convient : un nom qui se termine par `.txt` ou `.md`, et n'importe quel nom, comme `Notre ville - sous-titres v3.txt`.
- Enregistrez-le en **UTF-8** si votre programme le demande. Les lettres accentuées, les guillemets courbes et les notes de musique sont alors conservés. Un fichier enregistré en « ANSI » Windows (Windows-1252) est aussi compris. Les fichiers de traitement de texte (`.doc`, `.docx`, `.pdf`) ne sont pas des fichiers texte : copiez le texte et collez-le d'abord dans un éditeur de texte brut, ou utilisez Enregistrer sous et choisissez le texte brut.
- Les fins de ligne Windows (CRLF) comme Mac ou Linux (LF) fonctionnent, de même que la marque invisible que certains éditeurs placent au début d'un fichier.
- Un `.zip` n'est nécessaire que pour une sauvegarde complète, qui inclut les images. Le fichier de spectacle, lui, n'est que du texte.

## Décider où chaque sous-titre commence et finit

Le programme ne coupe ni ne joint jamais vos lignes. Un sous-titre est exactement ce que vous avez tapé sur une ligne : c'est vous qui décidez de la quantité de texte que le public lit d'un coup.

- Visez des lignes qui se lisent confortablement d'un coup d'œil : une phrase ou une partie de phrase. Les très longues tirades se coupent aux pauses naturelles, là où l'acteur respire.
- Les lignes vides ne signifient rien : utilisez-les librement entre les répliques pour garder le fichier lisible.
- Si deux lignes ou plus doivent apparaître ensemble, joignez-les par une barre oblique inverse en fin de ligne, ou placez-les entre `[[` et `]]`.
- Utilisez `HAMLET: réplique` pour indiquer qui parle. Écrivez le nom en majuscules, puis deux-points.
- Décrivez les sons et les indications de jeu entre parenthèses, sur une ligne à part, par exemple `(le tonnerre gronde)`. Ils sont montrés au public dans un style différent.

```showfile
[section: Act 1]
[note: house lights down, then start]
GEORGE: Morning, Emily.
EMILY: Oh, hello, George.
(a train whistle, far away)
```

## Ce que seul l'opérateur voit

Repérez votre place et laissez des rappels sans déranger le public : `[section: Acte 2]` ajoute un signet dans la liste de l'opérateur, et `[note: attendre les applaudissements]` affiche un rappel. Ni l'un ni l'autre ne consomme un appui sur la barre d'espace. Un commentaire entre `<!--` et `-->` est pour vous et personne ne le voit.

## Travailler dans votre propre éditeur

Vous n'êtes jamais obligé de taper dans l'éditeur de l'étape Fichier du spectacle. À la place :

1. À l'étape **Fichier du spectacle**, appuyez sur **Fichier ▾ → Télécharger le fichier du spectacle**. Vous obtenez un fichier `.md` en texte brut.
2. Ouvrez-le dans votre propre programme, modifiez-le et enregistrez-le. Gardez-le en texte brut.
3. De retour à l'étape Fichier du spectacle, utilisez **Fichier ▾ → Téléverser un fichier de spectacle…** et choisissez le vôtre. Son texte remplace celui de l'éditeur et la liste des problèmes est mise à jour aussitôt. Rien n'est changé pour de bon avant que vous appuyiez sur **Enregistrer**.
4. Appuyez sur **Enregistrer**. La version remplacée est conservée sous **Versions antérieures**, où vous pouvez la comparer à la nouvelle : une erreur peut donc toujours être annulée.

Il en va de même pour une traduction : choisissez d'abord l'onglet de la langue, puis téléchargez et ouvrez le fichier de cette langue. Vous pouvez aussi déposer un fichier sur l'éditeur.

Pour créer un **nouveau spectacle** à partir d'un fichier, utilisez **Importer un spectacle** à l'étape Spectacle et choisissez le fichier. Il devient un nouveau spectacle, nommé d'après la ligne `title:` de son bloc `[settings]`, ou d'après le nom du fichier s'il n'y en a pas. Les images ne font pas partie d'un fichier texte : envoyez-les à l'étape Images, avec les mêmes noms que dans vos lignes `[image: ...]`.

## Vérifier votre travail

- L'étape Fichier du spectacle énumère chaque problème avec son numéro de ligne. Cliquez sur l'un d'eux pour y aller. Les erreurs ne vous empêchent jamais d'enregistrer, mais un élément en erreur est ignoré : corrigez-les avant le spectacle.
- Vous pouvez vérifier un fichier de spectacle sans ouvrir la configuration : sur l'ordinateur de sous-titrage, lancez `open-caption --check-show show.md` (ajoutez ensuite les traductions, par exemple `show.es.md`). Il affiche les mêmes problèmes avec leurs numéros de ligne et se termine avec un statut non nul s'il y a des erreurs, ce qui permet de l'utiliser pendant l'édition ou dans un script. `open-caption --help` liste les options.
- Le panneau à côté de l'éditeur montre ce que l'opérateur va parcourir, une ligne par appui sur la barre d'espace. Vérifiez que le nombre de lignes est celui que vous attendez.
- Lancez l'étape **Test** pour voir les sous-titres comme le public les verra, sur un téléphone comme sur un grand écran.
- Si d'autres personnes travaillent aussi sur le fichier du spectacle, voyez *Si deux personnes modifient le même fichier* dans [Écrire un fichier de spectacle](admin-writing-a-show-file.md).

## Avant de passer la main

- Relisez le fichier une fois de haut en bas comme le public, et non comme un correcteur.
- Vérifiez que chaque nom d'`[image: ...]` correspond à une image envoyée, et que chaque `[loop]`, `[song]`, `[together]` et `[fullscreen]` a sa ligne de fermeture.
- Donnez à l'opérateur une copie du fichier et les noms des sections, pour qu'il sache où retrouver sa place. [Chansons et musique](admin-musicals.md), [Boucles](admin-loops.md) et [Langues](admin-languages.md) détaillent ces parties.
