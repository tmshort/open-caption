# Comédies musicales

Les sous-titres d'une comédie musicale fonctionnent mieux quand les paroles se distinguent facilement des répliques parlées et que l'opérateur peut suivre la musique. Cette page rassemble ce qui marche.

## Marquer les paroles

Les paroles s'affichent **en italique entre des notes de musique**, par exemple ♪ Où est l'amour ? ♪. Choisissez la façon de les écrire qui vous convient :

| Écrivez | Résultat |
|---------|----------|
| `~ Where is love? ~` | ♪ Where is love? ♪ (le `~` de fermeture est facultatif) |
| `OLIVER: ~ Where is love?` | pareil, avec le nom du personnage au-dessus |
| `[song: Title]` … `[/song]` | chaque ligne à l'intérieur est une parole ; le titre est réservé à l'opérateur |
| `♪ saisi ou collé ♪` | exactement ce que vous avez saisi, sans italique |
| `~` seul sur une ligne | juste une note : ♪, pour un passage instrumental |

Utilisez `[song]` pour les longs numéros, et `~` pour une seule ligne chantée au milieu d'une scène.

Dans la configuration, les boutons **🎵 Chanson**, **♪** et **♫** au-dessus de l'éditeur du fichier du spectacle les insèrent pour vous. Les caractères ♪ et ♫ peuvent aussi être copiés de n'importe où, car ce sont des caractères de texte ordinaires.

## Deux chanteurs à la fois (contrepoint et duos)

Quand des chanteurs différents chantent des paroles différentes **en même temps**, montrez-les ensemble sur les écrans du public avec `[together]` :

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]
```

- Les deux lignes apparaissent **en même temps** et l'opérateur appuie sur Espace **une seule fois**.
- Sur un grand écran, les voix se placent côte à côte, chacune avec une barre le long de son bord gauche et le nom de son personnage au-dessus ; sur un téléphone, elles s'empilent.
- Ajoutez autant de voix que nécessaire, par exemple un trio ou une réplique parlée sur une partie chantée.
- Pour une réplique rapide à deux voix, vous pouvez écrire `ANNA: ~ I love you ~ || KRIS: ~ Me too ~` sur une seule ligne.
- Les voix qui **entrent l'une après l'autre** (un canon, ou un second chanteur qui répond) sont des lignes séparées ordinaires : une pression sur Espace chacune. Écrivez la première voix, puis la seconde quand elle entre.
- Gardez chaque voix courte. Trois ou quatre mots par voix se lisent plus facilement en même temps qu'une phrase entière.
- Utilisez des noms de personnage pour que le public distingue les voix.

## Conseils pour de bons sous-titres de paroles

- **Une ligne par phrase chantée.** Coupez les lignes là où les chanteurs respirent, pour que la pression sur Espace de l'opérateur suive la musique. Les lignes plus courtes sont plus faciles à suivre.
- **Utilisez des noms de personnage pour les duos** et pour les lignes chantées tour à tour par des personnages différents. Omettez-les pour un chœur.
- **Passages instrumentaux :** utilisez `~` seul, ou `(L'orchestre joue)`, pour les marquer. Une ligne entre parenthèses s'affiche comme une indication de scène.
- **Séparez les lignes parlées et chantées.** Une réplique suivie de paroles avec `~` se lit clairement : le public voit changer les notes et l'italique.
- **Reprises :** écrivez les lignes à chaque fois. L'opérateur appuie une fois sur Espace par ligne dans tous les cas.
- **Effacer entre les numéros :** mettez `[clear]` après la dernière ligne d'une chanson pour que la scène suivante commence sur un écran vide.
- **Rappels pour l'opérateur :** `[note: le tempo s'accélère après le changement de ton]` n'est montré qu'à l'opérateur.
- **Répétez avec la vraie musique.** L'étape Test permet à l'opérateur de s'entraîner à appuyer sur Espace avec l'enregistrement de la distribution.

## Un exemple

```showfile
[section: Act 1, Number 4]
[clear]
(The orchestra begins.)
[song: Where Is Love]
OLIVER: Where is love?
Does it fall from skies above?
[note: wait for the held note]
Is it underneath the willow tree
That I've been dreaming of?
[/song]
[clear]
MRS. CORNEY: That's quite enough of that!
```

Voir aussi [Écrire un fichier du spectacle](topic:admin-writing-a-show-file).
