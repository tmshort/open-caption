# Apparence

L'étape **Apparence** règle l'aspect des sous-titres sur chaque appareil du public. Les valeurs par défaut, texte blanc sur fond noir, sont les plus faciles à lire dans une salle sombre et les meilleures pour l'accessibilité.

- **Couleurs :** une mesure du contraste indique la lisibilité de la combinaison. Visez 7:1 ou mieux ; la page prévient quand il est trop faible.
- **Taille du texte :** met tout le texte à l'échelle. Les appareils adaptent le texte à leur écran, c'est donc un changement relatif, pas une taille en points. Le curseur va de 50 à 300 pour cent ; 100 est la valeur normale.
- **Police :** sobre (la police habituelle de l'appareil), avec empattements, à chasse fixe, ou **adaptée à la dyslexie (OpenDyslexic)**, une police aux bas de lettres plus lourds et aux formes distinctives que certains lecteurs dyslexiques trouvent plus faciles. Elle est intégrée au programme, ne demande donc pas d'internet, et un appareil ne la télécharge (environ 115 Ko) que lorsqu'elle est utilisée. Les lignes gagnent un peu de hauteur avec elle. Choisissez **haute lisibilité (Atkinson Hyperlegible)** pour une police sobre conçue par le Braille Institute pour la basse vision, aux lettres faciles à distinguer (par exemple le I majuscule, le l minuscule et le chiffre 1). Les deux polices sont intégrées au programme, et un appareil n'en télécharge une que s'il l'utilise (environ 17 Ko pour Atkinson Hyperlegible). Les spectateurs peuvent aussi choisir l'une ou l'autre pour leur propre appareil avec le petit contrôle **Aa**, quelle que soit la police choisie ici.
- **Noms de personnage :** affichés dans une seule couleur (la *couleur des noms de personnage*), dans une **couleur différente pour chaque personnage**, ou masqués. Les couleurs des personnages sont choisies automatiquement, dans l'ordre où chaque personnage parle pour la première fois, dans une palette adaptée aux spectateurs daltoniens, et ajustées pour rester lisibles sur votre arrière-plan.
- **Position du nom de personnage :** au-dessus de la ligne (par défaut), **à côté du texte** sur la même ligne (en gras), ou **dans une colonne** à gauche du texte, comme dans un scénario. Les voix affichées ensemble placent toujours le nom à côté de leur propre texte.
- **Message d'avant-spectacle :** affiché jusqu'à l'envoi de la première ligne, si le fichier du spectacle ne commence pas par une boucle.

Les changements s'enregistrent avec **Enregistrer l'apparence** et parviennent tout de suite aux appareils connectés. **Enregistrer l'apparence écrit vos choix dans le bloc `[settings]` en haut du fichier du spectacle** (voir [Écrire un fichier du spectacle](topic:admin-writing-a-show-file)), de sorte que le fichier du spectacle emporte son apparence partout où il est copié ; vous pouvez aussi modifier ce bloc à la main. Tout ce que le fichier du spectacle ne définit pas utilise les valeurs par défaut de `config.yaml` (sa section `defaults`), qui partent elles-mêmes des valeurs intégrées du programme. Si l'éditeur du fichier du spectacle contient des modifications non enregistrées, enregistrez-les avant d'enregistrer l'apparence.

## Un écran pour toute la salle

Si un navigateur affiche les sous-titres sur un projecteur ou un grand moniteur, ouvrez l'adresse du public en ajoutant `?display=projector` à la fin. Le texte est plus grand, le pointeur de la souris est masqué et rien d'autre n'est affiché. Un écran partagé ne défile jamais : si le dernier sous-titre est plus haut que l'écran (beaucoup de voix à la fois, par exemple), le texte est réduit jusqu'à ce qu'il tienne.

## Relire des lignes précédentes

Chacun peut faire défiler les sous-titres vers le haut pour relire ce qu'il a manqué. Les nouvelles lignes attendent alors derrière un bouton **↓ Retour au direct (3)** (il compte ce qui est arrivé) au lieu d'entraîner la page, et les anciennes lignes cessent de s'estomper pendant qu'on les lit. Toucher le bouton, ou faire défiler jusqu'en bas, ramène aux sous-titres en direct. Un écran affichant `?display=projector` n'est jamais défilable.

## Commandes du public

Les spectateurs n'ont jamais rien à faire. Un petit contrôle discret **Aa** dans le coin supérieur droit de la page du public permet à chacun d'agrandir ou de réduire son propre texte, de passer en sombre sur clair, de choisir la police adaptée à la dyslexie ou celle de haute lisibilité, ou de désactiver **L'écran reste allumé** (la page garde allumé l'écran du téléphone pendant le spectacle, et quelqu'un qui manque de batterie peut préférer ne pas le faire), mais uniquement sur son propre appareil.
