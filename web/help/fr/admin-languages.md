# Langues

Un spectacle peut proposer ses sous-titres dans plusieurs langues. Chaque langue est son **propre fichier de spectacle**, et l'opérateur ne suit que le principal : une pression sur Espace envoie le même moment dans toutes les langues à la fois.

## Ajouter une langue

1. À l'étape **Fichier du spectacle**, ouvrez le menu de langue à gauche de la barre d'outils au-dessus de l'éditeur et choisissez **＋ Ajouter une langue**.
2. Choisissez la langue (ou « Autre langue… » et saisissez un code court comme `cy`).
3. Laissez cochée la case **Partir d'une copie du fichier principal du spectacle**. Vous obtenez une copie de la bonne forme, il ne reste qu'à remplacer les mots.
4. Traduisez, puis **Enregistrez**.

Chaque traduction commence par une ligne qui nomme sa langue. L'éditeur l'ajoute pour vous :

```showfile
[language: es | Español]
Hola, Margaret.
```

Le nom après le `|` est ce que les spectateurs voient dans le menu des langues ; écrivez-le donc dans sa propre langue (Español, Français, Deutsch).

## La seule règle : garder la même forme

Chaque langue doit avoir **les mêmes tops dans le même ordre** : un sous-titre reste un sous-titre, un `[clear]` reste un `[clear]`, une boucle garde le même nombre d'éléments, et des voix montrées ensemble restent le même nombre de voix. Seuls les mots changent. Les noms de personnage, les notes, les noms de section, les noms d'image et les durées des boucles peuvent être différents.

Si une traduction cesse de correspondre, l'éditeur affiche une erreur avec la ligne où cela cloche, par exemple « Le top 4 ne correspond pas au fichier du spectacle en anglais : ce fichier a [clear] ici, et celui-ci a un sous-titre ». Une langue qui ne correspond pas **n'est pas proposée au public** tant qu'elle n'est pas corrigée. Le fichier du spectacle en anglais (principal) n'est pas touché, et l'étape Partager signale le problème.

Pour indiquer au public quels mots vont avec quelle ligne, gardez les lignes en phase : si l'anglais découpe une réplique en trois lignes, l'espagnol aussi, même si ce n'est pas ainsi que cette langue la découperait.

## Ce que voit le public

- Chaque appareil reçoit la meilleure langue que propose le spectacle. On regarde les réglages de langue du téléphone ou du navigateur (par exemple l'espagnol du Mexique reçoit l'espagnol), et on utilise la langue principale quand rien ne correspond.
- Un petit contrôle **🌐** dans le coin supérieur gauche liste les langues. En choisir une change cet appareil, et l'adresse se termine alors par `?lang=es`, ce qui permet d'imprimer un lien ou un code QR pour chaque langue.
- Les mots de la page elle-même (« Reconnexion… », taille du texte, etc.) sont aussi traduits en espagnol, français, allemand, italien et portugais. Les autres langues affichent ces quelques libellés en anglais ; les sous-titres eux-mêmes sont toujours dans la langue du fichier du spectacle.
- Les langues écrites de droite à gauche (arabe, hébreu) s'alignent à droite.

## Autres choses à savoir

- **Images :** toutes les langues peuvent utiliser les mêmes images. Pour une image contenant du texte, téléversez une version par langue et nommez chacune dans le fichier du spectacle de cette langue.
- **Message d'avant-spectacle :** chaque langue a le sien, conservé dans le fichier du spectacle de cette langue (`idle message:` dans son bloc `[settings]`). L'étape Apparence a une zone pour chaque langue et les écrit pour vous. Le fichier du spectacle d'une traduction ne peut rien définir d'autre : l'apparence se définit une seule fois, dans le fichier principal du spectacle.
- **Les annonces de l'opérateur** (« Annonce… ») sont envoyés tels quels à toutes les langues ; saisissez-les donc dans les langues voulues.
- **Modifier un fichier de spectacle pendant le spectacle :** ajouter ou corriger une traduction prend effet tout de suite et conserve la position de l'opérateur. Retirer une langue n'est pas permis tant que le spectacle est en direct.
- **Décomptes :** la page de l'opérateur montre combien d'appareils lisent chaque langue, par exemple `👥 83 (en 60 · es 23)`.
- **Sauvegardes :** Exporter enregistre toutes les langues ; Dupliquer les copie aussi.

Voir aussi [Écrire un fichier du spectacle](admin-writing-a-show-file.md).
