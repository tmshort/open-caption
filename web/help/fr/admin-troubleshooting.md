# Dépannage

## Le bouton Dépannage

Le bouton **Dépannage**, en haut de la configuration, à côté de Réglages, est l'endroit où commencer quand des personnes ne peuvent pas se connecter ou que quelque chose semble anormal. Il liste ce que le programme sait : si cet ordinateur a une adresse réseau à donner, si des adresses de plusieurs réseaux sont données, quels ports (HTTP et HTTPS) le serveur web a réellement obtenus, si HTTPS et son certificat sont en ordre, si le nom `.local` est annoncé, si le fichier de réglages peut être enregistré, si chaque langue correspond au fichier de spectacle principal, si une liste d'accès refuse des ordinateurs (ou laisse la page Suivre ouverte), si l'horloge de l'ordinateur concorde avec celle du navigateur, si le fichier du spectacle chargé contient des erreurs ou s'il manque une image, si le dossier de données est accessible en écriture et a de la place, si les mots de passe sont identiques ou en clair, si le spectacle est en attente ou à l'antenne (ce que voit le public maintenant) et combien d'appareils sont connectés. Chaque vérification est marquée **OK**, **Note**, **À vérifier** ou **Problème**, et celles qu'on peut corriger ont un bouton qui mène au bon réglage. Sous les vérifications se trouvent les derniers avertissements et erreurs du journal, un lien vers le journal complet et le paquet d'assistance. **Tester la connexion** interroge le serveur web à chaque adresse donnée par le programme et indique lesquelles répondent et à quelle vitesse ; un pare-feu de cet ordinateur ou du réseau peut quand même bloquer les autres appareils.

Quand le programme enregistre de nouveaux avertissements ou erreurs, un nombre apparaît sur le bouton jusqu'à ce que vous l'ouvriez.

## Personne ne peut ouvrir l'adresse

- Vérifiez que le téléphone est sur le WiFi invités de la salle, pas sur les données mobiles.
- Vérifiez que l'adresse correspond à celle de l'étape **Partager**. Si elle comporte un numéro de port, incluez-le.
- Demandez à la salle si les appareils du WiFi invités ont le droit d'atteindre cet ordinateur.
- Le pare-feu de votre ordinateur bloque peut-être le programme. Sous Windows, autorisez-le sur les **réseaux privés** quand on vous le demande. Sous macOS, autorisez les connexions entrantes pour lui.
- Ouvrez l'adresse sur cet ordinateur avec `localhost` pour vérifier le programme lui-même.

## Le nombre d'appareils connectés est plus faible que prévu

Les appareils qui se sont endormis se déconnectent et se reconnectent tout seuls au réveil. Demandez aux gens de désactiver le verrouillage automatique pendant le spectacle si leur écran s'assombrit.

## Les images s'affichent lentement

Les grandes images prennent du temps sur un WiFi chargé. Cherchez à l'étape Images les fichiers marqués « volumineuse » et remplacez-les par des plus petits.

Pour la liste complète des vérifications côté réseau, voir le [guide réseau de la salle](../../../docs/i18n/fr/venue-network-guide.md).

## Une langue manque pour le public

Ouvrez l'étape Fichier du spectacle et regardez l'onglet de cette langue : un ⚠ signifie que son fichier du spectacle n'a plus les mêmes tops que le principal. Corrigez l'erreur affichée ; la langue revient dès que vous enregistrez. Voir [Langues](admin-languages.md).

## Problèmes du fichier du spectacle

Voir « Ce que signifient les avertissements » dans [Écrire un fichier du spectacle](admin-writing-a-show-file.md).

## Mots de passe

Pour changer l'un ou l'autre mot de passe, utilisez **Réglages** en haut de la page de configuration. Le nouveau mot de passe prend effet immédiatement et est enregistré dans `config.yaml`. Toute personne connectée avec un mot de passe qui a changé est déconnectée ; vous restez connecté. Les mots de passe sont conservés sous forme d'empreintes : la configuration ne peut donc en afficher aucun, notez le mot de passe en le définissant. Un **mot de passe spectateur** facultatif permet à un deuxième appareil, comme la tablette d'un régisseur, d'ouvrir la page de suivi (`/follow`) en simple consultation : elle affiche les lignes, les notes et la position, et rien ne peut en être envoyé. Les mots de passe opérateur et administrateur ouvrent aussi cette page.

Si vous ne pouvez pas du tout vous connecter, les deux mots de passe se trouvent dans `config.yaml`, dans le dossier de données du programme. Ouvrez le fichier avec n'importe quel éditeur de texte. Son emplacement est affiché au démarrage du programme.

## Consulter le journal

**Dépannage → Ouvrir le journal complet…** montre ce que le programme a enregistré récemment. Choisissez *Problèmes seulement* pour voir les avertissements et les erreurs, comme des connexions refusées ou un nom déjà utilisé sur le réseau. La même fenêtre règle la quantité de détails enregistrés (**Niveau de détail du journal**) ; *Débogage* ajoute chaque requête, ce qui aide quand quelque chose d'étrange se produit. Les mots de passe ne sont jamais écrits dans le journal.

## Demander de l'aide

Téléchargez le **paquet de support** avec le bouton **Dépannage** en haut de la configuration. Il contient les réglages (sans mots de passe), les dernières lignes du journal et des informations de base sur cet ordinateur, et il est facile à envoyer à la personne qui maintient le programme.
