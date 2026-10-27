# Conduire le spectacle

La page montre le fichier du spectacle sous forme de liste. Les lignes déjà envoyées sont estompées, la ligne **à l'écran maintenant** porte un point vert, et la ligne **suivante** est grande avec une flèche jaune.

## Les trois choses à retenir

1. **Espace** envoie la ligne suivante.
2. **Retour arrière** annule la dernière chose envoyée.
3. **Maj+Q** arrête une boucle d'avant-spectacle ou d'entracte et vide l'écran.

Suivez les comédiens et appuyez sur **Espace** à chaque réplique. Vous pouvez utiliser une télécommande ; ses boutons suivant et précédent fonctionnent comme Espace et Retour arrière.

## Écran vide et Effacer

Les deux laissent les écrans du public vides, mais ce n'est pas la même chose :

| | **Écran vide** (`B`) | **Effacer** (`C`) |
|---|---|---|
| Ce que voit le public | Rien : juste l'arrière-plan | Une zone de sous-titres vide, prête pour de nouvelles lignes |
| Les lignes précédentes | **Conservées.** Appuyez de nouveau sur `B`, ou envoyez la ligne suivante, et elles sont toujours là au-dessus | **Effacées.** La ligne suivante commence sur un écran vide |
| À utiliser pour | Un noir, une pause, ou cacher les sous-titres un instant | Une nouvelle scène, ou après une chanson, pour que les anciennes lignes ne s'accumulent pas |
| Annuler | `B` de nouveau, ou **Retour arrière** | **Retour arrière** |

Un fichier de spectacle peut faire la même chose avec les commandes `[blank]` et `[clear]` ; elles apparaissent dans votre liste sous les noms `ÉCRAN VIDE` et `EFFACER L'ÉCRAN`.

## Les mots de cette page

- **En attente** (**EN ATTENTE**) : le spectacle est chargé, mais le public ne voit que l'écran d'attente du lieu.
- **À l'antenne** (**À L'ANTENNE**) : le public voit le spectacle. **DÉMARRER** (START) le met à l'antenne ; **Mettre en attente…** le remet en attente en gardant sa position.
- **Redémarrer le spectacle…** : retour à la première ligne et à l'attente.
- **Écran vide** masque les sous-titres mais garde les lignes précédentes ; **Effacer** les supprime (voir plus haut).

## Avant le spectacle

- Utilisez la page de l'opérateur sur un ordinateur avec clavier (un portable convient). Elle est pensée pour la touche **Espace**, et un téléphone ou une tablette n'est pas recommandé : un écran tactile se touche trop facilement par erreur pendant le spectacle.
- Connectez-vous à la page de l'opérateur. Vous verrez « Vous avez la main ». Si elle indique qu'un autre opérateur a la main, utilisez **Reprendre la main** seulement si c'est bien vous l'opérateur.
- Vérifiez le nombre de spectateurs (l'icône de personnes). Il indique combien d'appareils sont connectés.
- Un spectacle chargé reste **en attente** : le public voit l'écran d'attente du lieu (un logo et un message) et rien du spectacle, et cette page indique **EN ATTENTE**. Appuyez sur **Espace** (le bouton indique **DÉMARRER**) pour mettre le spectacle à l'antenne. Si le fichier du spectacle commence par une boucle d'avant-spectacle, elle commence alors.

## Pendant le spectacle

- Les lignes marquées comme notes ou sections sont pour vous. Elles ne consomment jamais une pression sur Espace.
- Les images et les messages plein écran n'apparaissent ici que sous forme d'étiquette, comme `IMAGE : rideau.jpg`. Vous ne voyez jamais l'image elle-même.
- **Chansons :** les lignes chantées sont entourées de notes de musique (♪) et s'affichent en italique sur les écrans du public. Le titre d'une chanson apparaît comme un signet bleu pour vous seul. Essayez d'appuyer sur Espace au début de chaque phrase.
- **Langues :** si le spectacle a des traductions, vous suivez la langue principale et tout ce que vous envoyez part vers toutes les langues à la fois. Le nombre de spectateurs dans l'en-tête liste chaque langue, par exemple `83 (en 60 · es 23)`.
- **Plusieurs voix à la fois :** une ligne marquée `ENSEMBLE` montre deux chanteurs ou plus sur les écrans en même temps. Une seule pression sur Espace les envoie tous.
- La bande sous l'en-tête indique qui a le contrôle, **Élément n sur total** et **Dernier envoi il y a …**, pour voir d'un coup d'œil que le spectacle et votre écran sont toujours synchrones. Il devient jaune après deux minutes sans ligne.
- **Aller à la section…** (à côté de la zone de recherche) liste les signets du fichier du spectacle ; en choisir un le sélectionne, et vous confirmez le saut.
- Pour avancer ou reculer, cliquez sur une ligne, puis confirmez « Sauter ici ». Vous pouvez aussi chercher avec `/`.
- Une annonce ponctuelle comme « Le spectacle reprendra bientôt » s'envoie avec **Annonce…**. Appuyez sur **Retour** pour la retirer. Tant qu'elle est à l'écran, un bandeau bleu de cette page la cite.
- **Mettre en attente…** (en haut à droite) retire le spectacle des écrans du public : il revoit l'écran d'attente, et le spectacle garde sa place. **DÉMARRER** le remet à l'antenne exactement où il en était.

Voir [le clavier](operator-keyboard.md) et [En cas de problème](operator-if-something-goes-wrong.md).

## Après le spectacle : le rapport sur le public

Quand vous arrivez à la fin du spectacle, le bandeau indique **Fin du spectacle.** avec un lien **Rapport sur le public** (l'étape **Rapport** de la configuration l'affiche ensuite). Le rapport dit combien d'appareils différents ont utilisé la page du public pendant que le spectacle était à l'antenne, quelle qu'ait été la durée de chacun : un téléphone qui a perdu le WiFi puis est revenu compte une fois. Il montre les totaux (téléphones, tablettes, ordinateurs, écrans de projecteur, le maximum connecté en même temps, combien sont encore connectés), les langues et une ligne par appareil. Il trace aussi un graphique du nombre d'appareils connectés au fil du temps, avec les boucles d'avant-spectacle et d'entracte marquées, pour voir combien de personnes étaient là au début et combien sont restées pendant la pause. Il continue de compter jusqu'à ce que vous reveniez en attente, redémarriez ou chargiez un autre spectacle, et on peut l'imprimer ou le télécharger en fichier CSV pour un tableur.

Un deuxième appareil qui ne fait que regarder le spectacle (la page de suivi) figure dans le même tableau, marqué **Spectateur**, mais n'est compté dans aucun chiffre ni dans le graphique. Les appareils sont distingués par leur adresse réseau et leur type de navigateur ; les chiffres sont donc une estimation : des téléphones qui partagent les deux comptent pour un seul, et un téléphone dont l'adresse change compte deux fois. Chaque rapport est enregistré dans un fichier du dossier `reports` du dossier de données ; les 30 plus récents sont conservés, et vous pouvez en supprimer un avec **Supprimer…** à l'étape **Rapport** de la configuration, ou supprimer les fichiers là.
