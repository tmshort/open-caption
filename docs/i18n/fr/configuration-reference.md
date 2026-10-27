# Référence de `config.yaml`

`config.yaml` contient tout ce qui concerne *cet ordinateur et ce lieu* : les mots de passe, les ports, HTTPS, qui peut ouvrir la configuration, et les valeurs par défaut qu'un spectacle utilise quand son propre fichier n'en dit rien. Tout ce qui concerne *un spectacle* (titre, couleurs, durée des boucles, message d'avant-spectacle…) se trouve au contraire dans le fichier de ce spectacle (voir [Écrire un fichier de spectacle](../../../web/help/admin-writing-a-show-file.md), « Réglages du spectacle ») ; les couches sont expliquées plus bas dans « Où un réglage est décidé ».

## Le fichier

- **Où il se trouve.** `<dossier de données>/config.yaml` ; le dossier de données est `data_dir`, `--data`, ou par défaut `<votre dossier de configuration utilisateur>/open-caption` (sous macOS `~/Library/Application Support/open-caption`, sous Windows `%AppData%\open-caption`, sous Linux `~/.config/open-caption`). Les premières lignes affichées par le programme disent quel fichier il utilise. `--config` désigne un autre fichier.
- **Premier démarrage.** S'il n'y a pas de fichier, le programme en écrit un avec des mots de passe aléatoires d'**administrateur** et d'**opérateur** et les affiche une seule fois dans la console. Le fichier a le mode 0600 (privé pour votre compte) et les mots de passe sont en **clair** (sauf si vous conservez une empreinte, voir « Mots de passe » plus bas) ; gardez-le donc privé.
- **Vous n'avez besoin que des clés que vous changez.** Une clé absente garde sa valeur par défaut (les tableaux ci-dessous). Le fichier généré liste tout.
- **Strict.** Une clé mal orthographiée, une valeur d'un mauvais type ou hors limites arrête le programme au démarrage en listant **tous** les problèmes, chacun avec le nom du réglage (par exemple `audience.max_clients: 0 is not a client count: use 1 to 100000`). Rien n'est deviné. Une seule exception : dans un fichier marqué d'un `config_version` plus récent que celui que connaît le programme, les clés inconnues sont ignorées, avec un avertissement.
- **La configuration en modifie une partie.** Changer les mots de passe, le nom `.local`, le message d'attente ou le niveau du journal dans la configuration réécrit ces clés dans le fichier et conserve vos commentaires et tous les autres réglages. Tout le reste se modifie à la main, puis il faut **redémarrer le programme**. (Un fichier de certificat renouvelé est la seule chose prise en compte sans redémarrage ; voir `https.cert_file`.)
- **Les booléens** sont `yes`/`no`/`true`/`false`. **Les listes** s'écrivent `[a, b]` ou avec un `- élément` par ligne.

## Ligne de commande et environnement

Une option, ou sa variable d'environnement, l'emporte sur le fichier pour cette exécution. (Le fichier n'est pas modifié.) Une option peut s'écrire avec un ou deux tirets (`-data` ou `--data`). `open-caption --help` affiche cette liste, et la **page de manuel** (`docs/open-caption.1` ; lisez-la avec `man ./docs/open-caption.1`, ou copiez-la dans un dossier `man1` comme `/usr/local/share/man/man1/`) décrit les mêmes options, ainsi que les fichiers, les signaux et le statut de sortie.

| Option | Environnement | Effet |
|---|---|---|
| `--config FILE` | `OPEN_CAPTION_CONFIG` | Utilise ce fichier de réglages (par défaut `<dossier de données>/config.yaml`) |
| `--data FOLDER` | `OPEN_CAPTION_DATA` | Le dossier de données (spectacles, sessions, `live.json`, certificats). L'emporte sur `data_dir` |
| `--http-addr ADDR` | `OPEN_CAPTION_HTTP_ADDR` | L'emporte sur `http.addr`, par exemple `:80` ou `:8080` |
| `--https-addr ADDR` | `OPEN_CAPTION_HTTPS_ADDR` | **Active** HTTPS à cette adresse (l'emporte sur `https.enabled` et `https.addr`) |
| `--log-level LEVEL` | `OPEN_CAPTION_LOG_LEVEL` | L'emporte sur `log.level` à chaque démarrage du programme (la configuration l'indique quand il est en vigueur) |
| `--log-file FILE` | `OPEN_CAPTION_LOG` | Ajoute aussi le journal à ce fichier |
| `--network-interfaces LIST` | `OPEN_CAPTION_NETWORK_INTERFACES` | L'emporte sur `network.interfaces` à chaque démarrage du programme : une liste séparée par des virgules de noms d'interface, comme `Ethernet` ou `en0,en4` (la configuration l'indique quand il est en vigueur) |
| `--network-restrict` | `OPEN_CAPTION_NETWORK_RESTRICT` | Refuse aussi les connexions faites à des adresses non proposées, à chaque démarrage du programme (comme `network.restrict_listening: yes`). La variable compte quand elle vaut `yes`, `true`, `on` ou `1` |
| `--no-browser` | | N'ouvre pas la configuration dans un navigateur au démarrage (comme `defaults.open_browser: no`) |
| `--version` | | Affiche la version et les fournisseurs DNS inclus pour les certificats automatiques, puis quitte |
| `--help` | | Affiche les options et les variables d'environnement, et où en lire plus, puis quitte |
| `--check-show FILE [TRANSLATION ...]` | | **Vérifie un fichier de spectacle** et affiche un rapport, comme un « lint », puis quitte : pas de serveur, pas de `config.yaml`. Voir « Vérifier un fichier de spectacle » plus bas |
| `--assets FOLDER` | | Le dossier d'images pour `--check-show` (par défaut : le dossier `assets` à côté du fichier de spectacle) |
| `--strict` | | Avec `--check-show` : les avertissements donnent aussi le statut de sortie 1 |
| `--hash-password` | | Lit un mot de passe sur l'entrée standard et l'affiche sous forme d'empreinte `sha256:` pour `admin_password` ou `operator_password`, puis s'arrête : ni serveur ni `config.yaml`. Voir « Mots de passe » plus bas |
| `--dev`, `--show-file` | | **Pour les développeurs seulement** : points d'accès de test sans authentification sur cet ordinateur, sans attente ; ne les utilisez jamais dans un lieu |

## Mots de passe

`admin_password` et `operator_password` sont **en clair** par défaut, c'est pourquoi le fichier est privé pour votre compte. Pour conserver un mot de passe dans le fichier sous forme d'**empreinte** (un hachage à sens unique, de sorte qu'une copie de `config.yaml` ne révèle pas le mot de passe), écrivez l'empreinte à sa place :

| Écrit comme | Signifie |
|---|---|
| `hunter22` | le mot de passe lui-même (sans préfixe) |
| `plain:hunter22` | le mot de passe lui-même, écrit avec un préfixe ; à utiliser pour un mot de passe qui commence par `sha256:`, `sha512:`, `sha1:` ou `plain:` |
| `sha256:5b11…77d6` | l'empreinte SHA-256 du mot de passe, 64 chiffres hexadécimaux (`sha512:` et `sha1:` fonctionnent aussi, mais `sha1:` est faible et le programme avertit) |

Fabriquez l'empreinte avec le programme, en lisant le mot de passe sur l'entrée standard pour qu'il reste hors de l'historique de votre shell :

```
$ printf '%s' 'the password' | open-caption --hash-password
sha256:5b11618c2e44027877d0cd0921ed166b9f176f50587fc91e7534dd2946db77d6
```

(`printf '%s' 'the password' | shasum -a 256` donne les mêmes chiffres sous macOS et Linux ; ajoutez `sha256:` devant.) Pour se connecter, on tape toujours le mot de passe lui-même. Une empreinte ne peut pas redevenir le mot de passe : **Réglages → Sécurité** dans la configuration ne peut donc pas afficher un mot de passe d'opérateur conservé sous forme d'empreinte, seulement le remplacer, et le programme ne peut pas le réafficher : gardez-le en lieu sûr. Un mot de passe changé dans la configuration est conservé sous la même forme que celui qu'il remplace (une empreinte reste une empreinte, le clair reste en clair) ; le programme ne convertit jamais un mot de passe que vous n'avez pas changé. Une valeur qui commence par `sha256:`, `sha512:` ou `sha1:` sans être une empreinte de la bonne longueur arrête le programme au démarrage, en nommant le réglage.

## Vérifier un fichier de spectacle

`open-caption --check-show show.md` analyse un fichier de spectacle sans le charger dans le programme : il exécute le même analyseur et les mêmes contrôles que la liste des problèmes de l'étape Fichier du spectacle de la configuration et affiche ce que cette liste montrerait.

```
$ open-caption --check-show show.md show.es.md
show.md: info: 4 captions, 1 images, 0 fullscreen, 1 loops, 1 sections, 0 notes.
show.md:8: warning: Image "curtain.jpg" not found in assets.
show.md:13: error: [song] was opened but never closed with [/song]. The lyrics so far were kept.
show.md: 8 cues; 2 speakers: ALICE, BOB
show.es.md: 8 cues; language es

images checked against assets (1 found)
missing images: curtain.jpg (line 8 of show.md)
1 error, 1 warning, 1 note
```

- Chaque problème a la forme `file:line: severity: message` (gravité `error`, `warning` ou `note`), celle que comprennent les éditeurs et les systèmes d'intégration continue. Une ligne qui concerne tout le fichier n'a pas de numéro de ligne.
- Les fichiers après le premier sont des **traductions** : chacun est analysé comme une traduction et comparé au fichier principal (le même nombre et les mêmes types de repères, dans le même ordre), comme le fait la configuration. La langue est la ligne `[language]` du fichier ou son nom (`show.es.md`) ; deux fichiers pour une même langue sont une erreur.
- Les images sont vérifiées dans le dossier `assets` à côté du fichier principal, ou dans `--assets FOLDER`. Sans dossier, elles ne sont pas vérifiées, et le rapport le dit. Une image manquante indique le dossier où elle a été cherchée (l'éditeur de la configuration dit plutôt de la téléverser à l'étape Images).
- Une image du dossier qu'aucun fichier de spectacle n'affiche est listée comme `unused images (…): name.png` et comptée comme note, pas comme avertissement (l'étape Images de la configuration indique « Non utilisée » dessous). Les images utilisées seulement par une traduction comptent comme utilisées.
- `-` comme nom de fichier lit l'entrée standard.
- **Statut de sortie :** `0` aucune erreur, `1` des erreurs (ou des avertissements avec `--strict`), `2` un fichier n'a pas pu être lu. Mettez les autres options (`--assets`, `--strict`) **avant** `--check-show` : les fichiers qui le suivent sont des traductions.

## Réglages

Les valeurs par défaut sont celles que vous obtenez sans la clé. « Configuration » indique où la configuration peut la modifier.

### Niveau supérieur

| Clé | Par défaut | Signification |
|---|---|---|
| `config_version` | `1` | Écrit par le programme. Un numéro plus grand que celui que connaît le programme signifie que le fichier vient d'une version plus récente |
| `admin_password` | générée | Le mot de passe de la configuration (`/admin`). En clair, ou une empreinte telle que `sha256:…` (voir « Mots de passe » plus bas). Ne doit pas être vide (6 caractères ou plus quand on le change dans la configuration). *Configuration : Réglages → Sécurité* |
| `operator_password` | générée | Le mot de passe de la page de l'opérateur (`/operator`) ; celui de l'administrateur y fonctionne aussi. En clair ou une empreinte, comme `admin_password` ; une empreinte ne peut pas être affichée dans la configuration. Ne doit pas être vide. *Configuration : Réglages → Sécurité* |
| `viewer_password` | *(aucun)* | Un troisième mot de passe facultatif qui permet à un deuxième appareil (la tablette d'un régisseur) de **regarder** la page de l'opérateur : il se connecte sur `/operator` avec ce mot de passe et voit les lignes, les notes et la position, mais ne peut rien envoyer ni prendre le contrôle. Vide = pas de connexion spectateur. En clair ou une empreinte, différent des deux autres. *Configuration : Réglages → Sécurité* |
| `data_dir` | vide | Le dossier de données. Vide signifie le dossier par défaut ci-dessus. Ne doit pas être un fichier |

### `http`

| Clé | Par défaut | Signification |
|---|---|---|
| `http.addr` | `:80` | Où écoute le serveur HTTP non chiffré : `:80` (toutes les adresses, port 80), `192.168.1.20:80`, ou `127.0.0.1:8080` (cet ordinateur seulement). Ne doit pas être vide |
| `http.fallback_addr` | `:8080` | Utilisée, avec un avertissement visible, si `addr` ne peut pas l'être (le port 80 est souvent occupé ou demande des privilèges). Vide signifie pas de repli |

### `https`

HTTPS est **désactivé** par défaut. Il y a trois façons d'obtenir un certificat ; choisissez-en une.

| Clé | Par défaut | Signification |
|---|---|---|
| `https.enabled` | `no` | Sert HTTPS en plus de HTTP, sur des ports distincts (il n'y a pas de redirection). Demande l'une des trois ci-dessous |
| `https.addr` | `:443` | Le port HTTPS. Obligatoire si activé |
| `https.fallback_addr` | `:8443` | Utilisée si `addr` ne peut pas être ouvert |
| `https.self_signed` | `yes` | Sans fichiers de certificat et sans `acme` : le programme fabrique son propre certificat (conservé dans `<dossier de données>/tls/`). Les navigateurs le signalent tant qu'il n'est pas accepté. `no` sans fichiers est une erreur |
| `https.cert_file`, `https.key_file` | vide | Votre propre certificat et sa clé (PEM). Indiquez **les deux** ou aucun ; les deux fichiers doivent exister. **Les fichiers sont surveillés** : quand ils sont remplacés (par exemple par `certbot`), la nouvelle paire sert aux nouvelles connexions en une trentaine de secondes, sans redémarrage. Une paire impossible à charger garde l'ancien certificat et consigne un avertissement |

### `https.acme` (certificats automatiques)

Désactivé sauf `enabled: yes`. Demande `https.enabled: yes`, un nom de domaine public qui vous appartient et un fournisseur DNS avec lequel ce programme a été compilé. Ne peut pas être combiné avec `cert_file`/`key_file`. Comment le mettre en place, pas à pas : le [guide du certificat HTTPS de confiance](https-certificate-guide.md) ; les choix de fournisseurs DNS à la compilation, et les tailles, sont à la fin de ce guide.

| Clé | Par défaut | Signification |
|---|---|---|
| `https.acme.enabled` | `no` | Obtient et renouvelle un certificat automatiquement (ACME, DNS-01) |
| `https.acme.domain` | | Le nom public, par exemple `captions.example.org`. Un seul nom : pas de joker, pas de `.local`/`.lan`/`.internal`, pas d'adresse IP. Il doit pointer, pour les téléphones du public, vers l'adresse de cet ordinateur |
| `https.acme.email` | | Où l'autorité de certification envoie les avis d'expiration |
| `https.acme.agree_to_terms` | `no` | **Doit valoir `yes`** : cela consigne que vous acceptez les conditions d'utilisation de l'autorité de certification |
| `https.acme.directory` | `letsencrypt` | `letsencrypt`, `letsencrypt-staging` (pour essayer : les navigateurs ne lui font pas confiance), ou l'adresse `https://` d'un autre annuaire ACME |
| `https.acme.renew_before_days` | `30` | Renouvelle quand il reste ce nombre de jours ou moins (1 à 60) |
| `https.acme.dns.provider` | | Le fournisseur DNS, par son [nom dans lego](https://go-acme.github.io/lego/dns/) (`cloudflare`, `godaddy`, `easydns`, …). Doit être inclus dans le programme (`--version` les liste) |
| `https.acme.dns.credentials` | vide | Les réglages propres au fournisseur, sous forme de table : pour EasyDNS `EASYDNS_TOKEN: …` et `EASYDNS_KEY: …`. Les noms sont en majuscules, chiffres et tirets bas. L'environnement et `credentials_file` l'emportent sur ceux-ci |
| `https.acme.dns.credentials_file` | vide | Un fichier de lignes `NAME=value` avec les mêmes réglages, relu à chaque tentative. Doit exister ; protégez-le avec `chmod 600` |
| `https.acme.dns.propagation_wait` | vide | Une attente fixe (`90s`, `2m`, au plus `30m`) pour que le nouvel enregistrement DNS se propage, au lieu de le vérifier |

### `defaults`

Ce qu'un spectacle utilise quand son propre fichier ne le dit pas. Le bloc `[settings]` du fichier de spectacle l'emporte sur toutes les valeurs d'ici (voir [Écrire un fichier de spectacle](../../../web/help/admin-writing-a-show-file.md)). L'étape Apparence de la configuration écrit le fichier de spectacle, pas celles-ci.

| Clé | Par défaut | Signification |
|---|---|---|
| `defaults.language` | `en` | La langue principale d'un spectacle dont le fichier n'en nomme pas : deux ou trois lettres, avec une région facultative (`en`, `es`, `pt-br`) |
| `defaults.loop_item_seconds` | `10` | Combien de temps reste chaque élément d'une boucle d'avant-spectacle ou d'entracte, en secondes (1 à 3600), sauf si l'élément en dit autrement |
| `defaults.idle_message` | `Captions will begin shortly` | Ce que lit le public avant l'envoi de la première ligne quand un spectacle n'a pas de message d'avant-spectacle (jusqu'à 200 caractères) |
| `defaults.open_browser` | `yes` | Ouvre la configuration dans un navigateur au démarrage du programme (ce n'est pas un réglage de spectacle ; `--no-browser` le désactive pour une exécution) |
| `defaults.look.background` | `#000000` | Le fond des sous-titres : `#rgb`, `#rrggbb`, ou `black`, `white`, `yellow`, `red`, `green`, `blue`, `cyan`, `magenta`, `orange`, `gray` |
| `defaults.look.text_color` | `#ffffff` | La couleur du texte des sous-titres (mêmes formes) |
| `defaults.look.speaker_name_color` | `#ffffff` | La couleur des noms de personnages (mêmes formes) |
| `defaults.look.font` | `system` | `system`, `serif`, `mono`, `dyslexic` ou `hyperlegible` |
| `defaults.look.text_size_percent` | `100` | Taille du texte des sous-titres, de 50 à 300 |
| `defaults.look.speaker_names` | `label` | `label` (noms d'une seule couleur), `color` (une couleur différente pour chaque personnage) ou `hidden` |
| `defaults.look.speaker_position` | `above` | `above` (le nom au-dessus de la ligne), `beside` (sur la même ligne, avant le texte) ou `column` (dans une colonne à gauche du texte) |

### `audience`

| Clé | Par défaut | Signification |
|---|---|---|
| `audience.max_clients` | `500` | Le maximum d'appareils du public connectés **par langue** (1 à 100000) ; les suivants sont refusés jusqu'à ce que certains partent |
| `audience.scrollback_lines` | `200` | Combien de lignes précédentes un appareil peut remonter (20 à 2000) |

### `access`

Qui peut ouvrir la configuration et la page de l'opérateur. La page du public est toujours ouverte à tous.

| Clé | Par défaut | Signification |
|---|---|---|
| `access.admin_allow` | `[localhost]` | Ordinateurs autorisés à ouvrir la configuration ; par défaut, seulement l'ordinateur qui exécute le programme. Une liste vide (`[]`) signifie que **n'importe quel ordinateur du réseau peut essayer le mot de passe** ; le programme le signale par un avertissement. Entrées : `localhost`, `private` (les réseaux domestiques et de bureau usuels), une adresse (`192.168.1.50`) ou un réseau (`192.168.1.0/24`). Exemple : `[localhost]` |
| `access.operator_allow` | `[]` | Pareil, pour la page de l'opérateur |
| `access.local_auto_login` | `yes` | Un navigateur sur l'ordinateur qui exécute le programme (ouvert à `localhost`, `127.0.0.1` ou `[::1]`) est connecté comme administrateur sans mot de passe. Les autres ordinateurs utilisent toujours un mot de passe. Désactivez-le sur un ordinateur où d'autres personnes peuvent ouvrir une session |

### `log`, `mdns`, `standby`

| Clé | Par défaut | Signification |
|---|---|---|
| `log.level` | `info` | `debug` (chaque requête), `info` (événements et connexions), `warn` ou `error`. *Configuration : Réglages → À propos* |
| `mdns.enabled` | `yes` | Annonce cet ordinateur sous le nom `<nom>.local` sur le réseau local |
| `mdns.name` | `open-caption` | Le nom : lettres, chiffres et traits d'union seulement, jusqu'à 63 caractères (`.local` est ajouté). *Configuration : Réglages → Réseau* |
| `standby.message` | vide | La ligne que voit le public sur l'écran d'attente. Vide signifie « Captions will begin shortly » dans la langue de chaque spectateur. Une ligne, jusqu'à 200 caractères. *Configuration : Réglages → Écran du public* (le logo est un fichier choisi là, pas un réglage) |

### `network`

| Clé | Par défaut | Signification |
|---|---|---|
| `network.interfaces` | `[]` | Les interfaces réseau utilisées, par nom (`Ethernet`, `Wi-Fi`, `en0`, `eth0`) : seules leurs adresses sont données au public, listées dans la configuration, mises dans un nouveau certificat HTTPS autosigné et annoncées comme `<nom>.local`. Vide signifie qu'aucun choix n'a été fait : toutes les interfaces sont utilisées. Une fois un choix fait, une interface qui apparaît plus tard (un adaptateur USB, un VPN) reste désactivée jusqu'à son ajout. Quand deux interfaces utilisées sont sur le même réseau (un câble et le Wi-Fi à la fois), une seule adresse est proposée : la filaire, sinon celle de la connexion principale ; laissez l'autre de côté pour choisir. *Configuration : Réglages → Réseau* |
| `network.restrict_listening` | `no` | **Refuse aussi les connexions** faites à une adresse d'une interface qui n'est pas utilisée (un téléphone qui l'utilise est renvoyé), au lieu de seulement écarter l'adresse des listes. Les connexions à cet ordinateur lui-même (`localhost`) sont toujours acceptées : on peut donc ouvrir la configuration pour le changer. La configuration ne désactive jamais l'interface par laquelle vous l'utilisez. *Configuration : Réglages → Réseau* |

Les adresses d'une interface non utilisée ne sont pas listées à l'étape Partager ni au démarrage, ne sont pas données au public, ne sont pas annoncées comme `<nom>.local` et ne sont pas mises dans un nouveau certificat HTTPS autosigné. Seules les interfaces connectées sont listées, et une seule adresse est proposée par réseau. Par défaut cela ne choisit que ce qui est *proposé* : le serveur accepte toujours les connexions sur toutes les adresses, tant que `network.restrict_listening` n'est pas activé.

## Où un réglage est décidé

Du plus faible au plus fort : la valeur intégrée du programme → `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) → le `[settings]` du fichier de spectacle. Une option de ligne de commande ou une variable d'environnement l'emporte sur le fichier. Les réglages propres au lieu (ports, HTTPS, mots de passe, `access`, `data_dir`, `audience.max_clients`, `mdns`, `network`, `standby`, `log`) n'existent qu'ici.

## Un exemple complet

```yaml
config_version: 1
admin_password: "set-a-long-one"
operator_password: "another-one"
http:
  addr: ":80"
  fallback_addr: ":8080"
https:
  enabled: yes
  addr: ":443"
  fallback_addr: ":8443"
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org
    agree_to_terms: yes
    directory: letsencrypt
    dns:
      provider: easydns
      credentials_file: /etc/open-caption/dns.env
mdns:
  enabled: yes
  name: open-caption
network:
  interfaces: []
  restrict_listening: no
standby:
  message: "Welcome to the Little Theatre"
defaults:
  language: en
  loop_item_seconds: 10
  idle_message: "Captions will begin shortly"
  look:
    font: hyperlegible
    text_size_percent: 120
audience:
  max_clients: 500
  scrollback_lines: 200
access:
  admin_allow: [localhost]
  operator_allow: [private]
  local_auto_login: yes
log:
  level: info
```
