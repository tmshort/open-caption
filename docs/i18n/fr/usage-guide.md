# Guide d'utilisation d'open-caption

open-caption affiche des sous-titres en direct sur les téléphones, tablettes et écrans du public pendant une pièce ou une comédie musicale. Un **opérateur** bénévole suit le spectacle et appuie sur la barre d'espace à chaque ligne ; tous les appareils du WiFi de la salle affichent la ligne un instant plus tard. Rien n'est installé sur les appareils du public, et personne n'a rien à toucher.

Ce guide explique quoi acheter et comment l'organiser ([Déploiement](#recommandations-de-deploiement)), comment préparer un spectacle, comment le conduire, et donne une liste de vérification pour le jour J. Le bouton **Aide** du programme couvre le même terrain en morceaux plus petits, et [venue-network-guide.md](venue-network-guide.md) est écrit pour la personne qui gère le WiFi du bâtiment.

Sommaire : [Les éléments](#les-elements) · [Recommandations de déploiement](#recommandations-de-deploiement) · [Installation et démarrage](#installation-et-demarrage) · [Préparer un spectacle](#preparer-un-spectacle) · [Conduire le spectacle](#conduire-le-spectacle) · [Liste de vérification du jour J](#liste-de-verification-du-jour-j) · [Dépannage](#depannage) · [Maintenance](#maintenance) · [Aide-mémoire de l'opérateur](#aide-memoire-de-l-operateur)

## Les éléments

| Élément | Ce que c'est |
|---------|--------------|
| **Le serveur** | Le programme `open-caption`, qui tourne sur un ordinateur (macOS, Windows ou Linux). Il conserve le fichier du spectacle et envoie les lignes à tout le monde. |
| **Page de configuration** (`/admin`) | Là où les bénévoles préparent un spectacle : fichier du spectacle, images, apparence, répétition. Mot de passe administrateur. |
| **Page de l'opérateur** (`/operator`) | Là où l'opérateur suit le spectacle et envoie les lignes. Mot de passe opérateur. L'étape **Partager** de la configuration et son lien du menu l'ouvrent. |
| **Page de suivi** (`/follow`) | Un second appareil, comme la tablette du régisseur, qui ne fait que regarder le spectacle : il voit les lignes, les notes et la position et ne peut rien envoyer. Mot de passe spectateur facultatif ; les mots de passe opérateur et administrateur l'ouvrent aussi. L'étape **Partager** de la configuration la présente. |
| **Page du public** (`/`) | Ce que le public ouvre sur ses appareils. Sans mot de passe. Un navigateur sur un projecteur ou un grand moniteur peut afficher la même page pour toute la salle. |

Tout tourne sur un seul ordinateur. Le public, l'opérateur et les bénévoles de la configuration utilisent des navigateurs web ordinaires.

## Recommandations de déploiement

### La version courte

1. **Utilisez un point d'accès WiFi dédié aux sous-titres** sauf si le réseau invités de la salle est connu pour fonctionner (voir plus bas). Cela supprime le plus gros risque : un réseau que vous ne maîtrisez pas.
2. **Mettez le serveur sur une connexion filaire** vers ce point d'accès, avec une adresse réservée.
3. **Utilisez un point d'accès conçu pour les foules** (prévu pour plus de 100 appareils) placé dans la zone du public.
4. **Gardez l'ordinateur serveur branché, éveillé et sinon inactif** pendant le spectacle.
5. **Imprimez le code QR** et placez-le dans le programme et dans le hall.
6. **Répétez avec quelques vrais téléphones** dans la vraie salle avant la première.

### Options de réseau

| Option | Quand l'utiliser | Remarques |
|--------|------------------|-----------|
| **A. Point d'accès dédié (recommandé)** | Toute salle, surtout si vous ne pouvez pas modifier le réseau du bâtiment | Vous apportez un routeur ou un point d'accès. Le public rejoint son WiFi (par exemple « Theater Captions »). Aucune connexion internet n'est nécessaire. Vous maîtrisez la capacité, les réservations et les pare-feu. |
| **B. Le WiFi invités de la salle** | Le personnel réseau de la salle confirme que les appareils invités peuvent joindre votre ordinateur | Le moins cher. Demandez une adresse réservée pour votre ordinateur et que le trafic des invités vers le serveur soit autorisé. Voir [venue-network-guide.md](venue-network-guide.md). |
| **C. Le point d'accès de votre propre portable** | Répétitions et très petits publics uniquement | Les points d'accès de portable ne gèrent généralement qu'une poignée d'appareils. Pas pour une salle comble. |

### L'option A en détail (point d'accès dédié)

```
                 ┌───────────────────────┐
                 │  Access point/router  │  WiFi "Theater Captions"
                 │  (in the audience     │ ─ ─ ─ ─ ─ ─ ─ ─  phones / tablets
                 │   area, 5 GHz on)     │                  projector browser
                 └───────────┬───────────┘
                    Ethernet │
                 ┌───────────┴───────────┐
                 │ Captioning computer   │   ← operator uses this screen,
                 │  (open-caption)       │     or a second device on the same WiFi
                 └───────────────────────┘
```

- **Matériel :** un point d'accès ou routeur professionnel ou semi-professionnel conçu pour de nombreux appareils simultanés (cherchez plus de 100 clients dans ses caractéristiques et la prise en charge du 5 GHz), avec un port Ethernet. Les routeurs domestiques bon marché peinent souvent au-delà de 30 à 50 appareils. Pour une salle d'environ 150 places, prévoyez un point d'accès capable, et un second pour une salle profonde ou large.
- **Réglages :**
  - Créez un SSID avec un nom court et une phrase de passe WPA2 simple et imprimable, ou sans phrase de passe si la salle autorise un réseau ouvert. Imprimez-la sur l'affiche QR.
  - Activez le **5 GHz**, et privilégiez-le. Utilisez des largeurs de canal de 20 ou 40 MHz dans un bâtiment encombré.
  - Laissez l'**isolation des clients** activée si elle est proposée. Les appareils du public ne parlent qu'au serveur, jamais entre eux.
  - Donnez à l'ordinateur de sous-titrage une **réservation DHCP** (ou une adresse fixe) pour que l'adresse ne change jamais, et imprimez l'adresse sur l'affiche.
  - Vous n'avez pas besoin de connexion internet, et le programme n'en utilise jamais.
- **Emplacement :** placez le point d'accès dans la zone du public, en hauteur et dégagé des murs, pas dans une arrière-salle ni dans le hall. Ne le mettez pas derrière la scène ni le rack d'éclairage. Un portable sur une table éloignée convient : il est relié par câble.
- **Alimentation :** utilisez un parasurtenseur. Un petit onduleur sur le point d'accès et le serveur supprime le risque d'une prise débranchée.

### L'option B en détail (le WiFi invités de la salle)

Le programme se moque du réseau sur lequel il est, tant que les appareils du public peuvent le joindre. Avant de vous y fier :

- Demandez un **accès invités vers le serveur** : votre ordinateur doit être joignable depuis le SSID invités, sur le port TCP 80 (et 443 si vous activez HTTPS). L'« isolation des clients » entre invités convient ; empêcher les invités de joindre *votre ordinateur*, non.
- Demandez une **adresse réservée** pour l'adaptateur réseau de votre ordinateur.
- Branchez l'ordinateur au réseau par **Ethernet** si c'est du tout possible.
- Testez avec un téléphone sur le WiFi invités bien avant le spectacle (l'étape Partager affiche l'adresse).
- Ayez l'option A dans un sac comme solution de secours si le spectacle compte.

### L'ordinateur de sous-titrage

- **N'importe quel portable ou mini-PC ordinaire** convient. Le programme utilise très peu de mémoire (quelques dizaines de mégaoctets) et presque aucun processeur. Utilisez un ordinateur que vous pouvez garder branché.
- **Système d'exploitation :** macOS, Windows ou Linux. Utilisez le même ordinateur pour tous les spectacles si possible, pour que les mots de passe, les spectacles et l'adresse restent les mêmes.
- **Rendez-le ennuyeux :** réglez l'ordinateur pour qu'il ne se mette *jamais en veille* quand il est branché, désactivez les redémarrages automatiques pour mises à jour autour des dates de spectacle, fermez les autres programmes et gardez l'écran ouvert ou réglé pour ne pas se mettre en veille à la fermeture.
- **Pare-feu :** autorisez le programme sur les **réseaux privés** (Windows le demande la première fois ; sous macOS, autorisez les connexions entrantes pour lui).
- **Adresse :** filaire et réservée (voir plus haut). Si l'adresse change, les codes QR imprimés cessent de fonctionner.
- **Heure :** l'horloge n'a pas d'importance.

### Qui s'assoit où

- **L'opérateur** a besoin d'une vue dégagée sur la scène et d'un appareil fiable. L'arrangement le plus simple est l'ordinateur de sous-titrage lui-même au pupitre technique. La page de l'opérateur est pensée pour un ordinateur avec clavier (Espace envoie la ligne suivante). Utilisez l'ordinateur de sous-titrage ou un autre portable du même réseau (`http://<adresse de l'ordinateur>/operator`) pour que l'opérateur puisse s'asseoir là où il voit la scène. Un téléphone ou une tablette afficheront la page, mais ce n'est **pas recommandé** : un écran tactile se touche trop facilement par erreur pendant le spectacle. Une télécommande Bluetooth ou un clavier est une bonne alternative.
- **Un seul opérateur à la fois pilote le spectacle.** Toute autre personne qui ouvre la page de l'opérateur peut regarder, et peut utiliser **Reprendre la main** si l'opérateur la lui cède.
- **Les bénévoles de la configuration** utilisent la page de configuration depuis n'importe quel ordinateur du réseau avant le spectacle, ou depuis l'ordinateur de sous-titrage.
- **Un écran partagé** pour toute la salle : ouvrez l'adresse du public sur un ordinateur relié à un projecteur ou à un grand moniteur, ajoutez `?display=projector` à l'adresse pour un texte grand et aucun encombrement à l'écran, et passez le navigateur en plein écran (`F11`, ou lancez Chrome avec `--kiosk`).

### Faire tourner le serveur

Pour la plupart des théâtres, la façon la plus simple est la meilleure : lancez le programme à la main avant chaque spectacle, laissez sa fenêtre ouverte et appuyez sur `Ctrl+C` (ou fermez-la) ensuite. La position du spectacle est enregistrée en continu, redémarrer au milieu d'un spectacle est donc sans risque (voir [Dépannage](#depannage)).

Si l'ordinateur doit démarrer le serveur tout seul après un redémarrage, le dossier [`deploy/`](../../../deploy/) contient des fichiers prêts à l'emploi pour Linux (systemd), macOS (launchd) et Windows (Planificateur de tâches). Chacun commence par ses propres étapes d'installation.

### Amener les gens à la page

Par ordre de fiabilité :

1. **Le code QR et l'affiche** (Configuration → Partager → *Imprimer une affiche avec code QR pour le hall*). Mettez-la au mur, dans le programme et sur la carte au dos des fauteuils. C'est sur cela qu'il faut compter.
2. **Une adresse courte.** Sur le port 80, il n'y a pas de `:port` à saisir. Le programme s'annonce aussi sous le nom **`open-caption.local`** (changez-le dans Configuration → Réglages → Réseau → *Utiliser un nom*) : là où le réseau invités le permet, `http://open-caption.local/` fonctionne sur la plupart des téléphones sans rien configurer. Certains réseaux invités bloquent ce type d'annonce ; n'imprimez donc pas seulement le nom : l'affiche QR utilise l'adresse numérique. Un nom dans le DNS propre du routeur (comme `captions.theater`) fonctionne partout mais demande que la salle l'ajoute.
3. **La page d'accueil du point d'accès** (un « portail invités » ou « portail captif » qui redirige vers votre adresse après connexion). Cela se configure sur le routeur ou le point d'accès, pas dans ce programme. Voir [venue-network-guide.md](venue-network-guide.md).
4. **Option 114 du DHCP et `/captive-api` (avancé, expérimental).** L'option 114 est un réglage DHCP qui oriente les appareils vers une « API de portail captif ». Le programme répond à cette API à `/captive-api` (par exemple `https://<adresse>/captive-api`) par « ce réseau n'est pas captif » et l'adresse de la page. Deux mises en garde : cela **ne fait pas ouvrir la page de sous-titres aux téléphones d'eux-mêmes**, et la norme (RFC 8910/8908) exige du **HTTPS avec un certificat auquel les appareils font confiance** : un réglage en `http://` simple ou le certificat auto-signé du programme est donc ignoré. Ne la configurez que si la salle utilise déjà l'option 114 et peut fournir un certificat de confiance. Elle n'a pas été testée sur de vrais appareils.

### Ports et adresses

- Par défaut, le serveur écoute sur le **port 80**, de sorte que l'adresse du public n'a pas besoin de numéro de port, par exemple `http://192.168.1.10/`. Si autre chose utilise le port 80, ou si le système d'exploitation ne l'autorise pas, le serveur vous le dit et utilise le **port 8080** ; l'adresse du public se termine alors par `:8080`. Sous Linux, le port 80 demande une autorisation (`sudo setcap 'cap_net_bind_service=+ep' ./open-caption`) sauf si vous utilisez le fichier systemd, qui l'accorde.
- Vous pouvez choisir vous-même les ports avec `--http-addr :8081`, ou dans `config.yaml`.
- **HTTPS est facultatif.** Le `http://` simple fonctionne sur tous les appareils sans avertissement et est l'adresse recommandée pour le public. Une adresse HTTPS auto-signée fait afficher un avertissement de sécurité à chaque appareil. N'activez HTTPS que si vous pouvez installer un certificat auquel vos appareils font confiance. Si le lieu possède un nom de domaine, le programme peut obtenir et renouveler lui-même un certificat de confiance ; voir le [guide du certificat HTTPS de confiance](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

### Sécurité

- La page du public ne peut que lire. Elle ne peut pas modifier le spectacle ni atteindre les pages de configuration et de l'opérateur.
- La configuration et l'opérateur exigent chacune leur propre mot de passe (conservé dans `config.yaml` uniquement sous forme d'empreinte, un hachage à sens unique, donc impossible à relire ; le fichier est en outre créé lisible par vous seul). Un mot de passe **spectateur** facultatif (dans **Réglages**) laisse un deuxième appareil, comme la tablette d'un régisseur, ouvrir la page de suivi (`/follow` ; les mots de passe opérateur et administrateur l'ouvrent aussi) et regarder le spectacle sans rien pouvoir envoyer. Donnez le mot de passe opérateur à l'opérateur, pas le mot de passe administrateur. Les mots de passe peuvent être changés depuis **Réglages** dans la configuration.
- Sur l'ordinateur qui exécute le programme, la Configuration et la page de l'opérateur s'ouvrent **sans mot de passe** lorsque l'adresse commence par `localhost`, `127.0.0.1` ou `[::1]` (la personne devant cet ordinateur est considérée comme l'administrateur). Les autres ordinateurs, et toute requête relayée par un proxy, se connectent toujours avec un mot de passe. Mettez `access.local_auto_login: false` dans `config.yaml` si d'autres personnes peuvent ouvrir une session sur cet ordinateur.
- **Toute personne qui peut joindre le serveur peut ouvrir les pages de connexion de la configuration et de l'opérateur, et en `http://` simple les mots de passe circulent en clair sur le WiFi.** Sur un WiFi invités partagé, cela signifie qu'un spectateur curieux pourrait capturer un mot de passe et prendre le contrôle des sous-titres. Réduisez le risque, dans cet ordre :
  1. **Limitez les ordinateurs qui peuvent les ouvrir.** Dans `config.yaml` :
     ```yaml
     access:
       admin_allow: [localhost]            # Setup only on the captioning computer itself
       operator_allow: [192.168.1.50]      # the operator's computer (or a network such as 192.168.1.0/24)
     ```
     Les entrées peuvent être `localhost`, `private` (plages d'adresses domestiques ou de bureau), une seule adresse, ou un réseau comme `192.168.1.0/24`. Par défaut, `admin_allow` vaut `[localhost]` (la configuration ne s'ouvre que sur l'ordinateur de sous-titrage) ; une liste vide signifie n'importe quel ordinateur, avec un avertissement. `operator_allow` est vide par défaut. Une adresse de `admin_allow` peut aussi utiliser la page de l'opérateur. Au lieu de modifier le fichier, vous pouvez changer tout cela dans la configuration, sous **Réglages → Accès → Qui peut ouvrir les pages** ; le changement s'applique aussitôt et la configuration refuse une liste qui exclurait l'ordinateur que vous utilisez. Si vous modifiez le fichier à la main, redémarrez le programme ; une entrée erronée arrête le programme avec un message au lieu d'autoriser tout le monde.
  2. **Mettez l'opérateur sur un réseau filaire ou séparé** que les appareils du public ne peuvent pas rejoindre, ou donnez un mot de passe WPA2 au réseau des sous-titres.
  3. **Utilisez HTTPS** (`https.enabled` dans `config.yaml`) si vous pouvez donner aux appareils un certificat auquel ils font confiance ; un certificat auto-signé fait afficher un avertissement à chaque appareil, et une adresse de public non chiffrée est plus simple.
- Un site web, ou un autre service du même ordinateur, ne peut pas faire modifier le spectacle par un navigateur dans votre dos : les requêtes qui ne viennent pas des pages du programme lui-même sont refusées.
- Le programme est destiné à un réseau privé. **Ne l'exposez pas à internet.**

## Installation et démarrage

1. **Procurez-vous le programme.** Téléchargez la version pour votre ordinateur depuis la page *Releases* du dépôt et décompressez-la (ou construisez-la vous-même : voir les guides de compilation dans le dossier `docs` du dépôt source). Placez-la dans un dossier que vous garderez, par exemple `Documents/open-caption`.
2. **Premier démarrage.** Lancez-le : double-clic sous Windows, ou dans un terminal `./open-caption` (macOS/Linux). Le programme n'est pas signé : le premier lancement peut demander un clic de plus. Sous macOS, *clic droit → Ouvrir* (ou **Ouvrir quand même** dans Réglages Système → Confidentialité et sécurité) ; sous Windows, *Informations complémentaires → Exécuter quand même* dans la fenêtre SmartScreen.
3. **Notez les mots de passe.** La première fois, le programme affiche les mots de passe **administrateur** et **opérateur** et l'emplacement de `config.yaml`. Ils restent dans ce fichier.
4. La page de configuration s'ouvre dans votre navigateur. Sinon, ouvrez `http://localhost/admin` (ou `http://localhost:8080/admin` si la console dit qu'elle a utilisé le port 8080).
5. **Choisissez où vivent les données (facultatif).** Par défaut, les spectacles et réglages sont dans le dossier de configuration de votre utilisateur (macOS `~/Library/Application Support/open-caption`, Linux `~/.config/open-caption`, Windows `%AppData%\open-caption`). Pour tout garder dans un dossier de votre choix, démarrez avec `--data <dossier>`. Utilisez toujours le même dossier.

Options utiles : `--data <dossier>`, `--http-addr :8080`, `--https-addr :443`, `--no-browser`, `--log-file <fichier>`, `--log-level debug|info|warn|error`, `--version`. Toutes les options : `open-caption --help`, la page de manuel `docs/open-caption.1` (en anglais seulement) ou `docs/configuration-reference.md`.

Si `config.yaml` contient une erreur (un réglage mal orthographié, un nombre hors limites, une adresse incorrecte), le programme liste tous les problèmes avec leur ligne ou le nom du réglage et s'arrête ; corrigez le fichier et redémarrez. (Un fichier marqué d'un `config_version` plus récent que ce que le programme comprend fait exception : les réglages inconnus sont ignorés avec un avertissement, vous pouvez donc revenir à un programme plus ancien.)

## Préparer un spectacle

Ouvrez la page de configuration et suivez les six étapes en haut (une septième, **Rapport**, est pour après le spectacle). Le bouton **Aide** explique chacune.

1. **Spectacles.** Créez un spectacle (ou le spectacle d'exemple, pour essayer).
2. **Fichier du spectacle.** Collez le fichier du spectacle ou déposez un fichier `.txt` ou `.md`. Chaque ligne devient un sous-titre. Corrigez ce qui est listé dans l'onglet *Problèmes* à côté de l'éditeur ; cliquez sur l'un d'eux pour aller à sa ligne. Enregistrez.
3. **Images.** Téléversez les images utilisées par le fichier du spectacle (PNG, JPG, GIF ou WebP, idéalement de moins de 1 Mo chacune).
4. **Apparence.** Choisissez les couleurs, la police et la taille du texte. Les valeurs par défaut (blanc sur noir, grand) sont les meilleures pour la plupart des gens. Deux polices de lisibilité sont intégrées : une adaptée à la dyslexie (OpenDyslexic) et une de haute lisibilité (Atkinson Hyperlegible, conçue pour la basse vision). Chaque spectateur peut aussi choisir l'une ou l'autre pour son propre appareil avec le petit contrôle **Aa**.
5. **Test.** Faites une répétition privée : vous appuyez sur Espace comme le ferait l'opérateur, et voyez l'écran du public à côté. Rien n'atteint le public.
6. **Partager.** Vérifiez la liste de préparation, imprimez l'affichette QR et appuyez sur **Ouvrir la page Conduire**.

**Travailler sur un spectacle n'est pas la même chose que le rendre actif.** *Travailler sur ce spectacle* choisit seulement le spectacle que la configuration modifie (son nom est dans l'en-tête). Le spectacle que reçoit le public, l'**actif** (marqué ACTIF à l'étape Spectacles et montré dans l'en-tête par la pastille Actif, ambre en attente et rouge à l'antenne), ne change que lorsque vous appuyez sur **Ouvrir la page Conduire** à l'étape Partager de ce spectacle. Un spectacle à l'antenne n'est remplacé qu'après une demande de confirmation.

Ouvrir la page de l'opérateur charge le spectacle **en attente** : le public voit l'écran d'attente du lieu (un logo et un message, réglés une seule fois sous **Réglages → Écran du public**) et rien du spectacle ; vous pouvez donc préparer, répéter et vérifier la page de l'opérateur tranquillement. Quand vous êtes prêt, appuyez sur **Espace** (le bouton indique **DÉMARRER**) sur la page de l'opérateur pour mettre le spectacle à l'antenne. **Mettre en attente…** le retire de nouveau sans perdre sa place.

### Écrire le fichier du spectacle

La personne qui prépare le fichier du spectacle n'a pas besoin de lire le reste de ce guide : [Préparer un fichier de spectacle](../../../web/help/admin-preparing-a-show-file.md) couvre le travail du texte au fichier terminé, y compris la modification dans votre propre programme, et [Premiers pas](../../../web/help/admin-getting-started.md) parcourt les six étapes de la Configuration.

Un fichier du spectacle est du texte simple. La version courte :

```text
<!-- Comments like this are never shown. -->
[loop: 10s]
[image: welcome.png | 15s]
Welcome! Please silence your phones.
[/loop]

[section: Act 1, Scene 1]
[note: lights come up slowly]
[clear]
(A garden gate creaks open.)
MARGARET: You're late, Edwin.
EDWIN: The *train* was late.

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[/song]
~ Tea for two, and two for tea ~

[fullscreen: **INTERMISSION**]
```

- **Une ligne = un sous-titre** = une pression sur Espace. Vous décidez où couper les longues répliques.
- Les **noms de personnage** en majuscules suivis de deux-points s'affichent au-dessus de la ligne.
- Les **indications de scène** entre `(parenthèses)` s'affichent en italique.
- Les **paroles** entre des notes `♪`, en italique : commencez une ligne par `~`, ou entourez la chanson de `[song]`…`[/song]`. Pour deux chanteurs en même temps (un duo ou un contrepoint), placez leurs lignes entre `[together]` et `[/together]`, ou joignez-les avec ` || ` sur une seule ligne ; les deux s'affichent en même temps d'une seule pression sur Espace. Les boutons **♪**, **♫** et **🎵 Chanson** de l'éditeur les insèrent. Voir la rubrique d'aide *Comédies musicales*.
- Les **boucles** (avant-spectacle et entracte) répètent des images et du texte jusqu'à ce que l'opérateur passe à la suite. Une boucle tout au début du fichier du spectacle démarre quand l'opérateur met le spectacle à l'antenne (DÉMARRER).
- Les **notes** `[note: …]` et les signets `[section: …]` ne sont que pour l'opérateur. `[pace: 8s]` indique à l'opérateur combien de temps la ligne suivante reste d'ordinaire affichée, et une ligne qui commence par `# ` est un titre qui fonctionne comme un signet.
- La rubrique d'aide **Écrire un fichier du spectacle** liste tout, avec des exemples, et l'onglet *Aide-mémoire* à côté de l'éditeur le résume.

### Réglages du spectacle (titre, apparence et plus)

Tout ce qui concerne un spectacle hormis ses images est rangé dans son fichier du spectacle, y compris son titre et l'apparence des sous-titres, dans un bloc `[settings]` près du début :

```text
[settings]
title: The Garden Party
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
[/settings]
```

L'étape **Apparence** de la configuration écrit ce bloc pour vous, et vous pouvez le modifier à la main. Tout ce qu'un fichier de spectacle omet utilise une valeur par défaut : le programme a des valeurs intégrées, `config.yaml` (sa section `defaults`) peut les changer pour tous les spectacles de cet ordinateur, et chaque fichier de spectacle remplace les deux. Les ports, mots de passe, HTTPS et listes d'accès ne sont pas des réglages de spectacle et restent dans `config.yaml`. Comme le fichier du spectacle emporte son apparence, copier ou exporter le spectacle emporte aussi l'apparence. Voir la rubrique d'aide *Écrire un fichier du spectacle* pour tous les réglages.

### Vérifier les noms des personnages

Cliquez sur **Fichier ▾ → Personnages…** au-dessus de l'éditeur pour voir tous les noms utilisés comme personnage, les plus utilisés d'abord, avec le nombre de lignes de chacun. C'est le moyen rapide de repérer un nom mal orthographié : **ALCE** avec une ligne à côté de **ALICE** avec quarante est signalé par *Vouliez-vous dire ALICE ?*. Cliquez sur un nom pour aller à sa première ligne. Un groupe comme `ALICE & BOB:` compte pour chacun de ses noms.

### Revenir en arrière

Chaque enregistrement conserve la version qu'il a remplacée (les 50 dernières). **Fichier ▾ → Versions antérieures…** au-dessus de l'éditeur les liste ; **Charger** en place une dans l'éditeur, et **Enregistrer** la rend courante. **Comparer** montre ligne à ligne ce que changerait le chargement de l'une d'elles, et une version à laquelle vous donnez un nom (ou **Conserver** pour ce qui est enregistré maintenant) n'est jamais oubliée. Exportez quand même avant le spectacle : les versions antérieures vivent dans le dossier de données, pas dans l'exportation.

### Deux personnes qui modifient

Si quelqu'un d'autre enregistre un fichier de spectacle pendant que vous l'avez ouvert, le programme vous prévient à l'enregistrement au lieu d'écraser son travail, et vous laisse prendre sa version ou garder la vôtre.

### Plusieurs langues

Un spectacle peut proposer des sous-titres en plusieurs langues. Dans Configuration → Fichier du spectacle, ouvrez le menu de langue à gauche de la barre d'outils et choisissez **＋ Ajouter une langue**, laissez cochée la case « Partir d'une copie du fichier principal du spectacle » et traduisez les mots. Chaque langue est son propre fichier de spectacle, avec les mêmes lignes, commandes et écrans vides dans le même ordre : une pression sur Espace montre donc le même moment dans toutes les langues. L'opérateur ne suit que la langue principale. Les appareils du public reçoivent la langue réglée sur leur téléphone (ou la principale) et peuvent changer depuis le petit menu 🌐, ou recevoir un lien se terminant par `?lang=es`. Une traduction qui cesse de correspondre au fichier principal du spectacle est signalée et retenue jusqu'à correction. Le spectacle d'exemple inclut une traduction espagnole à regarder. Voir la rubrique d'aide *Langues*.

### Changer des choses pendant les répétitions et les spectacles

Vous pouvez modifier et réenregistrer un fichier de spectacle pendant qu'un spectacle est en direct (par exemple pour corriger une coquille). Le programme vous demande de saisir le nom du spectacle par sécurité, conserve la position de l'opérateur par numéro de ligne, et prévient si le nombre de lignes a changé.

## Conduire le spectacle

### Avant l'arrivée du public

1. Démarrez le serveur et ouvrez la **page de l'opérateur** ; connectez-vous. Elle doit dire **Vous avez la main**.
2. Dans Configuration → Partager, appuyez sur **Ouvrir la page Conduire**. Le spectacle est chargé en attente : le public voit l'écran d'attente du lieu. Quand vous êtes prêt, appuyez sur **Espace** (DÉMARRER) sur la page de l'opérateur pour le mettre à l'antenne. Une boucle d'avant-spectacle, si le fichier du spectacle en a une, commence alors sur tous les écrans.
3. Ouvrez l'adresse du public sur votre propre téléphone et, si vous en avez un, sur le projecteur. Vérifiez que le nombre de spectateurs augmente sur la page de l'opérateur.
4. Placez l'affiche QR là où les gens peuvent la voir.

### Pendant le spectacle

- **Espace** envoie la ligne suivante (aussi Flèche droite, Page suivante, Entrée, et la plupart des télécommandes).
- **Retour arrière** annule la dernière chose envoyée.
- **B** vide les écrans du public (masque tout ; les mêmes sous-titres reviennent quand vous appuyez de nouveau sur **B**). **C** les efface (efface les sous-titres, pour que la ligne suivante commence sur un écran vide). **Retour arrière** annule l'un ou l'autre.
- **Maj+Q** arrête une boucle d'entracte ou d'avant-spectacle (écran vide). **Espace** continue.
- **Annonce…** place un avis ponctuel sur chaque écran (« Le spectacle reprendra bientôt »). **Retour** le retire.
- Cliquez sur une ligne puis confirmez pour sauter n'importe où. **/** recherche dans le fichier du spectacle.
- La ligne avec la flèche jaune est la *suivante*. Le point vert marque ce qui est à l'écran maintenant.

### Quand le spectacle se termine

Appuyez sur `Ctrl+C` dans la fenêtre du programme (ou arrêtez le service). La position est enregistrée. Pour relancer le spectacle depuis le début, utilisez **Redémarrer le spectacle…** sur la page de l'opérateur, ou **Redémarrer le spectacle…** dans Configuration → Partager. Les deux le remettent en attente. **Combien de personnes l'ont utilisé ?** Quand le spectacle arrive à sa fin, le bandeau de la page de l'opérateur indique *Fin du spectacle.* avec un lien **Rapport sur le public** (l'étape **Rapport** de la configuration affiche le dernier). Il compte les appareils différents qui ont utilisé la page du public pendant que le spectacle était à l'antenne, quelle qu'ait été la durée de chacun (un téléphone qui a perdu le WiFi puis est revenu compte une fois), avec un graphique des appareils au fil du temps (qui montre quand ont tourné les boucles d'avant-spectacle et d'entracte), le moment de plus forte affluence, les langues et une ligne par appareil ; on peut l'imprimer ou le télécharger en fichier CSV. Les appareils sont distingués par leur adresse et leur type de navigateur : c'est donc une estimation. Les rapports sont conservés dans le dossier `reports` du dossier de données.

## Liste de vérification du jour J

**Quelques jours avant**
- [ ] Répétez avec le vrai fichier du spectacle et quelques vrais téléphones sur le vrai réseau dans la vraie salle.
- [ ] Exportez le spectacle (Configuration → Spectacles → **Plus ▾** → **Exporter**) et gardez le `.zip` en lieu sûr, comme une clé USB.
- [ ] Confirmez que l'adresse de l'ordinateur de sous-titrage est réservée et que l'affiche QR l'indique.

**À la salle, une heure avant**
- [ ] Alimentation : ordinateur et point d'accès branchés ; ordinateur réglé pour ne jamais se mettre en veille.
- [ ] Point d'accès allumé, en place, avec le bon nom ; ordinateur relié à lui par câble.
- [ ] Lancez le programme ; notez l'adresse affichée (et dans Partager).
- [ ] Un téléphone sur le WiFi du public peut ouvrir l'adresse et affiche « Les sous-titres vont bientôt commencer ».
- [ ] Configuration → Test : faites une courte répétition ; les touches de l'opérateur fonctionnent.
- [ ] Navigateur du projecteur (s'il y en a un) ouvert sur `…/?display=projector`, en plein écran.
- [ ] Appuyez sur **DÉMARRER** sur la page de l'opérateur. La boucle d'avant-spectacle (s'il y en a une) est sur tous les écrans.
- [ ] L'opérateur est assis là où il voit la scène ; télécommande ou clavier testé.
- [ ] Affiche QR en place ; un bénévole connaît le nom du WiFi et l'adresse pour ceux qui demandent.

**Après le spectacle**
- [ ] `Ctrl+C` dans la fenêtre du programme.
- [ ] Exportez de nouveau le spectacle si vous avez changé quelque chose.

## Dépannage

| Problème | Que faire |
|----------|-----------|
| Personne ne peut ouvrir l'adresse | Le téléphone est-il sur le bon WiFi (pas sur les données mobiles) ? L'adresse inclut-elle `:8080` si le programme a utilisé ce port ? Le trafic des invités vers le serveur est-il autorisé ? Le pare-feu de l'ordinateur autorise-t-il le programme sur les réseaux privés ? Voir [venue-network-guide.md](venue-network-guide.md). |
| Certains appareils se connectent et d'autres non | Le point d'accès a peut-être atteint sa limite d'appareils. Ajoutez un second point d'accès, ou rapprochez le premier de la salle. |
| Les sous-titres arrivent en retard ou se figent sur quelques téléphones | Signal faible ou canal occupé. Ces téléphones se reconnectent d'eux-mêmes. Améliorez l'emplacement du point d'accès ; utilisez le 5 GHz. |
| Quelqu'un a manqué une ligne | Il peut faire défiler les sous-titres vers le haut pour la relire ; les nouvelles lignes attendent derrière un bouton **Retour au direct** jusqu'à ce qu'il le touche. |
| Une page dit « Ce navigateur est trop ancien » | Le navigateur de l'appareil est en dessous de ce dont la page a besoin : environ iOS 12 pour la page du public, iOS 14.1 / Chrome 87 pour la configuration et la page de l'opérateur. Mettez l'appareil à jour, ou utilisez-en un autre. |
| Les écrans des téléphones s'assombrissent | Les téléphones verrouillent leur écran. La page garde l'écran allumé avec le verrou d'éveil du navigateur quand elle le peut et, quand elle ne le peut pas (HTTP simple), avec une minuscule vidéo invisible en boucle. Quiconque préfère laisser son téléphone s'endormir (batterie faible, par exemple) peut désactiver *L'écran reste allumé* dans le panneau « Aa ». Certains téléphones en mode économie d'énergie refusent les deux : la page affiche alors une astuce pour désactiver le verrouillage automatique pour la soirée. |
| La mauvaise ligne a été envoyée | **Retour arrière** l'annule. |
| La page de l'opérateur dit qu'un autre opérateur a la main | Utilisez **Reprendre la main**, si vous êtes l'opérateur. |
| L'ordinateur s'est figé ou a redémarré en plein spectacle | Relancez le programme. La page de l'opérateur se reconnecte, reste connectée et reprend la main toute seule, le spectacle est là où vous l'avez laissé, et les appareils du public se reconnectent d'eux-mêmes. Une boucle repart de son premier élément. **Retour arrière** peut encore annuler les dernières choses envoyées avant le redémarrage. |
| J'ai oublié un mot de passe | Connectez-vous avec l'autre et utilisez **Réglages** dans la configuration, ou ouvrez `config.yaml` dans le dossier de données avec n'importe quel éditeur de texte. |
| Autre chose | Configuration → Dépannage → **Télécharger un paquet de support** ; il contient les réglages (sans mots de passe) et les dernières lignes du journal. |

## Maintenance

- **Sauvegardes :** *Exporter* (menu **Plus ▾**) dans Configuration → Spectacles enregistre un spectacle (fichier du spectacle, images, apparence) sous forme de `.zip`. *Importer* le restaure, sur cet ordinateur ou un autre. Vous pouvez aussi copier tout le dossier de données.
- **Passer à un autre ordinateur :** installez le programme, importez les spectacles exportés, réglez l'adresse du nouvel ordinateur et, si vous voulez garder les anciens mots de passe, copiez `config.yaml`.
- **Mettre à jour le programme :** arrêtez-le (`Ctrl+C`), remplacez le fichier du programme par le plus récent et relancez-le avec le même dossier `--data`. Ne mettez pas à jour juste avant un spectacle.
- **Réutiliser un spectacle :** Configuration → Spectacles propose **Renommer…** et **Dupliquer**. Dupliquer fait une copie du fichier du spectacle, des images et de l'apparence à adapter pour la prochaine série, et laisse l'original intact. Renommer, Dupliquer, Exporter, Archiver et Supprimer se trouvent dans le menu **Plus ▾** de chaque spectacle.
- **Faire le ménage :** supprimez les spectacles dont vous n'avez plus besoin dans Configuration → Spectacles. Les `.zip` exportés sont vos archives.
- **Rapports sur le public :** chaque représentation laisse son rapport sur le public sous forme de fichier dans le dossier `reports` du dossier de données (Configuration → Réglages → À propos indique où est le dossier de données). Les 30 plus récents sont conservés et les plus anciens sont supprimés automatiquement. Un rapport liste les adresses réseau des appareils : supprimez donc les rapports que vous ne voulez pas garder (**Supprimer…** à l'étape **Rapport** de Configuration, ou les fichiers eux-mêmes), un seul ou tout le dossier `reports` (il est recréé à la prochaine mise à l'antenne) ; rien d'autre n'en dépend, et les supprimer ne touche pas vos spectacles.
- **Journaux :** Configuration → Dépannage → l'onglet **Avertissements et erreurs** montre ce que le programme a enregistré récemment, filtré par niveau (choisissez *Problèmes seulement* pour voir les avertissements et les erreurs) ; les dernières lignes du journal sont aussi dans le paquet de support ; ajoutez `--log-file` pour conserver un fichier. Chaque ligne commence par son niveau (`INFO`, `WARN`, `ERROR` ou `DEBUG`). Le niveau `info` par défaut enregistre les connexions, qui a pris la main sur la page de l'opérateur, les spectacles activés et des événements similaires ; `warn` ne garde que les problèmes ; `debug` enregistre aussi chaque requête, ce qui aide quand quelque chose d'étrange se produit. Réglez-le avec `log: {level: debug}` dans `config.yaml`, avec `--log-level debug`, ou pendant que le programme tourne depuis Configuration → Réglages → *Niveau de détail du journal*. Si le programme a été lancé avec `--log-level`, cette option l'emporte à chaque démarrage : la configuration le dit, un changement fait là ne dure que jusqu'au prochain redémarrage, et le journal de démarrage prévient quand l'option et le fichier ne sont pas d'accord. Un fichier journal conservé avec `--log-file` en commence un nouveau à 5 Mo et en garde trois plus anciens (`open-caption.log.1` à `.3`), si bien qu'il ne remplit jamais le disque.
- **Archivage :** dans Configuration → Spectacles, **Archiver** range un spectacle terminé (il est conservé, listé sous « Spectacles archivés », et ne peut pas passer à l'antenne) et **Restaurer** le ramène. **Supprimer** retire un spectacle définitivement.

## Aide-mémoire de l'opérateur

Imprimez ceci et gardez-le près de l'opérateur.

| Touche | Fait |
|--------|------|
| **Espace** | Envoyer la ligne suivante |
| **Retour arrière** | Annuler la dernière chose envoyée |
| **S** | Sauter la ligne suivante sans l'envoyer (Retour arrière l'annule) |
| **Maj+Q** | Arrêter une boucle (écran vide) |
| **B** | Écran vide : masque tout, les sous-titres reviennent au réaffichage |
| **C** | Effacer : efface les sous-titres pour repartir à neuf |
| **/** | Rechercher dans le fichier du spectacle |
| **?** | Aide (les touches sont en pause tant qu'elle est ouverte) |

- Flèche jaune = ligne suivante. Point vert = à l'écran maintenant.
- Mauvaise ligne ? **Retour arrière.** Vous avez perdu le fil ? Cliquez sur la ligne voulue puis confirmez.
- Besoin de dire quelque chose au public ? **Annonce…**
- Spectacle interrompu ? **B** pour vider les écrans, **Annonce…** pour expliquer.
