# Premiers pas

Préparer un spectacle demande six étapes (une septième, **Rapport**, est pour après le spectacle). La rangée de boutons en haut de la page les suit dans l'ordre, et une coche apparaît lorsqu'une étape semble terminée. Vous pouvez passer à n'importe quelle étape à tout moment.

1. **Spectacles** — créez un spectacle (ou le spectacle d'exemple pour découvrir).
2. **Fichier du spectacle** — collez ou ouvrez votre fichier du spectacle et corrigez ce que la page signale.
3. **Images** — téléversez les images utilisées par le fichier du spectacle.
4. **Apparence** — choisissez les couleurs et la taille du texte.
5. **Test** — répétez en privé avant que personne ne soit dans la salle.
6. **Partager** — partagez l'adresse avec le public et ouvrez la page de l'opérateur, d'où le spectacle est lancé. Facultatif : la page de suivi (`/follow`) permet à un second appareil de regarder sans pouvoir rien envoyer.
7. **Rapport** — après le spectacle : combien d'appareils ont utilisé la page du public, avec un graphique au fil du temps.

La première fois que la configuration est ouverte sans aucun spectacle, une courte **visite** présente chaque étape et les boutons de l'en-tête ; appuyez sur Échap pour la passer. Pour la revoir, ouvrez l'aide et choisissez **Faire la visite**.

## Attente et passage à l'antenne

Ouvrir la page de l'opérateur charge le spectacle **en attente** : le public voit l'écran d'attente du lieu (un logo et un message, réglés une seule fois sous **Réglages → Écran du public**) et rien du spectacle. L'opérateur appuie sur **Espace** (le bouton indique **DÉMARRER**) pour mettre le spectacle à l'antenne, et **Mettre en attente…** le retire sans perdre sa place. La préparation, la répétition et la modification ne sont jamais devant le public tant que le spectacle est en attente, et un spectacle en attente peut être supprimé.

## Où tout est rangé

Un spectacle est un dossier contenant son **fichier du spectacle** (un simple fichier texte, `show.md`), ses éventuelles traductions et ses images. Le fichier du spectacle contient tout ce qui concerne le spectacle sauf les images : son titre, son apparence et les lignes. Les réglages que le fichier du spectacle ne mentionne pas viennent de `config.yaml`, puis des valeurs intégrées du programme.

## Fini avec un spectacle ?

Renommer, Dupliquer, Exporter, Archiver et Supprimer se trouvent dans le menu **Plus ▾** de chaque spectacle. À l'étape Spectacles, **Archiver** range un spectacle : il est conservé intégralement mais listé sous *Spectacles archivés* et ne peut pas passer à l'antenne. **Restaurer** le remet en place. **Supprimer** l'efface définitivement (exportez-le d'abord si vous pourriez en avoir encore besoin). Le spectacle à l'antenne ne peut être ni archivé ni supprimé.

## Avant de commencer

- Cet ordinateur doit être sur le même réseau WiFi que les téléphones et tablettes du public. Voir [Réseau et connexion](admin-network.md).
- Le mot de passe administrateur et le mot de passe opérateur sont conservés dans `config.yaml` uniquement sous forme d'empreintes ; ils ont été affichés dans la console, une seule fois, au premier démarrage du programme. Pour en définir de nouveaux, utilisez **Réglages → Mots de passe**. La page de l'opérateur a son propre mot de passe, pour que vous puissiez le confier à quelqu'un d'autre sans lui donner accès à la configuration. Sur l'ordinateur qui exécute le programme, la Configuration et la page de l'opérateur s'ouvrent **sans mot de passe** lorsque l'adresse commence par `localhost`, `127.0.0.1` ou `[::1]` (la personne devant cet ordinateur est considérée comme l'administrateur). Les autres ordinateurs, et toute requête relayée par un proxy, se connectent toujours avec un mot de passe. Mettez `access.local_auto_login: false` dans `config.yaml` si d'autres personnes peuvent ouvrir une session sur cet ordinateur.
- La configuration, la page de l'opérateur et cette aide sont disponibles en anglais, espagnol, français, allemand, italien et portugais. Utilisez le menu de langue en haut de chaque page ; le choix est mémorisé dans ce navigateur.
- Votre travail est enregistré sur cet ordinateur. Utilisez **Exporter** (menu **Plus ▾** d'un spectacle) à l'étape Spectacles pour faire une sauvegarde que vous pourrez emporter sur un autre ordinateur. L'exportation est un `.zip` contenant `show.md` (le fichier du spectacle, avec son titre et son apparence), un `show.<langue>.md` pour chaque traduction et un dossier `assets` d'images ; **Importer** transforme un tel zip en nouveau spectacle, et accepte aussi un fichier de spectacle en texte brut (`.md` ou `.txt`, de n'importe quel nom) comme nouveau spectacle.

## Nouveau avec les fichiers de spectacle ?

Un fichier du spectacle est simplement du texte. Chaque ligne est un sous-titre, et une pression sur la barre d'espace envoie une ligne. Pour tout le travail, du texte au fichier terminé (y compris la modification dans votre propre programme et le retour du fichier), lisez [Préparer un fichier de spectacle](admin-preparing-a-show-file.md). Les marquages sont énumérés dans [Écrire un fichier du spectacle](admin-writing-a-show-file.md). Ou créez le spectacle d'exemple et regardez comment il est écrit. Un modèle à télécharger se trouve à l'adresse [sample-show.md](../../sample/sample-show.md).

## Prévoir le matériel

Pour savoir quoi acheter, où placer le point d'accès WiFi et disposer d'une liste de vérification pour le jour J, lisez le [guide d'utilisation](../../../docs/i18n/fr/usage-guide.md). Le [guide réseau de la salle](../../../docs/i18n/fr/venue-network-guide.md) s'adresse à la personne qui s'occupe du WiFi du bâtiment.

## Où aller ensuite

- [Comédies musicales](admin-musicals.md)
- [Écrire un fichier du spectacle](admin-writing-a-show-file.md)
- [Boucles : avant le spectacle et entracte](admin-loops.md)
- [Dépannage](admin-troubleshooting.md)
