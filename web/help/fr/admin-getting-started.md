# Premiers pas

Préparer un spectacle demande six étapes. La rangée de boutons en haut de la page les suit dans l'ordre, et une coche apparaît lorsqu'une étape semble terminée. Vous pouvez passer à n'importe quelle étape à tout moment.

1. **Spectacles** — créez un spectacle (ou le spectacle d'exemple pour découvrir).
2. **Fichier du spectacle** — collez ou ouvrez votre fichier du spectacle et corrigez ce que la page signale.
3. **Images** — téléversez les images utilisées par le fichier du spectacle.
4. **Apparence** — choisissez les couleurs et la taille du texte.
5. **Test** — répétez en privé avant que personne ne soit dans la salle.
6. **En direct** — partagez l'adresse avec le public et lancez le spectacle.

## Où tout est rangé

Un spectacle est un dossier contenant son **fichier du spectacle** (un simple fichier texte, `show.md`), ses éventuelles traductions et ses images. Le fichier du spectacle contient tout ce qui concerne le spectacle sauf les images : son titre, son apparence et les lignes. Les réglages que le fichier du spectacle ne mentionne pas viennent de `config.yaml`, puis des valeurs intégrées du programme.

## Fini avec un spectacle ?

À l'étape Spectacles, **Archiver** range un spectacle : il est conservé intégralement mais listé sous *Spectacles archivés* et ne peut pas passer à l'antenne. **Restaurer** le remet en place. **Supprimer** l'efface définitivement (exportez-le d'abord si vous pourriez en avoir encore besoin). Le spectacle à l'antenne ne peut être ni archivé ni supprimé.

## Avant de commencer

- Cet ordinateur doit être sur le même réseau WiFi que les téléphones et tablettes du public. Voir [Réseau et connexion](topic:admin-network).
- Le mot de passe administrateur et le mot de passe opérateur se trouvent dans `config.yaml` ; ils ont aussi été affichés dans la console au premier démarrage du programme. La page de l'opérateur a son propre mot de passe, pour que vous puissiez le confier à quelqu'un d'autre sans lui donner accès à la configuration.
- La configuration, la page de l'opérateur et cette aide sont disponibles en anglais, espagnol, français, allemand, italien et portugais. Utilisez le menu de langue en haut de chaque page ; le choix est mémorisé dans ce navigateur.
- Votre travail est enregistré sur cet ordinateur. Utilisez **Exporter** à l'étape Spectacles pour faire une sauvegarde que vous pourrez emporter sur un autre ordinateur. L'exportation est un `.zip` contenant `show.md` (le fichier du spectacle, avec son titre et son apparence), un `show.<langue>.md` pour chaque traduction et un dossier `assets` d'images ; **Importer** transforme un tel zip en nouveau spectacle, et accepte aussi un fichier de spectacle en texte brut (`.md` ou `.txt`, de n'importe quel nom) comme nouveau spectacle.

## Nouveau avec les fichiers de spectacle ?

Un fichier du spectacle est simplement du texte. Chaque ligne est un sous-titre, et une pression sur la barre d'espace envoie une ligne. Pour tout le travail, du texte au fichier terminé (y compris la modification dans votre propre programme et le retour du fichier), lisez [Préparer un fichier de spectacle](topic:admin-preparing-a-show-file). Les marquages sont énumérés dans [Écrire un fichier du spectacle](topic:admin-writing-a-show-file). Ou créez le spectacle d'exemple et regardez comment il est écrit. Un modèle à télécharger se trouve à l'adresse [/help/sample-show.md](/help/sample-show.md).

## Prévoir le matériel

Pour savoir quoi acheter, où placer le point d'accès WiFi et disposer d'une liste de vérification pour le jour J, lisez le [guide d'utilisation](topic:admin-usage-guide). Le [guide réseau de la salle](topic:admin-venue-network-guide) s'adresse à la personne qui s'occupe du WiFi du bâtiment.

## Où aller ensuite

- [Comédies musicales](topic:admin-musicals)
- [Écrire un fichier du spectacle](topic:admin-writing-a-show-file)
- [Boucles : avant le spectacle et entracte](topic:admin-loops)
- [Dépannage](topic:admin-troubleshooting)
