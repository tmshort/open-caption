# Mettre en place un certificat HTTPS de confiance (facultatif)

Ce guide s'adresse à la personne qui s'occupe de l'ordinateur de sous-titrage. Il est **facultatif** : les sous-titres fonctionnent sans lui. (Les réglages qu'il utilise sont dans la [référence de configuration](11-configuration-reference.md), et la conception et les tests dans le [document 10](10-acme-lets-encrypt.md) ; tous deux n'existent qu'en anglais.)

## En avez-vous besoin ?

Probablement pas. Les téléphones du public ouvrent `http://…` sans aucun avertissement, et c'est l'adresse que le code QR utilise. Une adresse sécurisée `https://` n'est nécessaire que pour quelques fonctions de navigateur réservées aux pages sécurisées (comme garder l'écran du téléphone allumé de la bonne manière ; la page du public a une solution de repli), ou si le lieu le souhaite.

Sans ce guide, HTTPS fonctionne quand même avec un certificat **auto-signé**, mais chaque téléphone affiche un avertissement de sécurité. Avec lui, le programme obtient un vrai certificat gratuit de Let's Encrypt et **le renouvelle tout seul** : `https://captions.mytheater.org` s'ouvre sans avertissement sur tous les appareils.

## Ce qu'il vous faut avant de commencer

| Il vous faut | Pourquoi |
|---|---|
| Un **nom de domaine qui vous appartient**, par exemple `mytheater.org`, et un nom sous celui-ci pour les sous-titres, tel que `captions.mytheater.org` | Un certificat est délivré pour un nom public. `open-caption.local` et les simples numéros n'en obtiennent pas |
| Son **DNS hébergé chez un fournisseur que ce programme prend en charge** | Le programme prouve que le nom est à vous en ajoutant un enregistrement temporaire dans votre DNS. Le programme par défaut inclut **Cloudflare, GoDaddy et EasyDNS** ; lancez `open-caption --version` pour voir ce que contient le vôtre. Un autre fournisseur demande une autre version compilée (voir la fin) |
| Des **identifiants d'API** pour ce compte DNS | Ils permettent au programme d'ajouter et de retirer cet unique enregistrement |
| Que l'ordinateur ait un **accès à Internet** (sortant seulement) | Il parle à Let's Encrypt et à votre fournisseur DNS. Rien n'a besoin d'atteindre le lieu *de l'extérieur* |
| Une **adresse fixe** pour l'ordinateur sur le réseau du lieu (une réservation DHCP convient) | Le nom public pointera vers elle |

## Étapes

**1. Obtenez les identifiants d'API.** Dans le compte de votre fournisseur DNS :

- **EasyDNS :** activez l'accès à l'API et créez un **jeton et une clé** d'API. Les réglages sont `EASYDNS_TOKEN` et `EASYDNS_KEY`. (EasyDNS propose un bac à sable pour essayer l'API ; voir leur documentation pour les étapes actuelles.)
- **Cloudflare :** créez un **jeton d'API** limité à votre zone avec la permission *Zone → DNS → Edit*. Le réglage est `CLOUDFLARE_DNS_API_TOKEN`.
- **GoDaddy :** créez une **clé et un secret** d'API de production. Les réglages sont `GODADDY_API_KEY` et `GODADDY_API_SECRET`. (GoDaddy peut limiter son API DNS à certains types de compte ; vérifiez leurs règles actuelles.)
- Tout autre fournisseur : les noms exacts des réglages sont dans la [documentation de lego](https://go-acme.github.io/lego/dns/).

Un jeton DNS peut généralement modifier **tous** les enregistrements de la zone, pas seulement le temporaire ; donnez-lui donc le moins de pouvoir possible. Pour la meilleure protection, utilisez un CNAME (voir « Des identifiants plus sûrs » plus bas).

**2. Faites pointer le nom vers l'ordinateur.** Chez votre fournisseur DNS, ajoutez un enregistrement `A` : `captions.mytheater.org` → l'adresse de l'ordinateur sur le réseau du lieu (par exemple `192.168.1.20`). Un nom public peut très bien pointer vers une adresse privée ; le certificat s'en moque.

**3. Gardez les identifiants dans un fichier privé.** Créez `/etc/open-caption/dns.env` (n'importe quel endroit privé convient ; sous Windows, un dossier que seul votre compte peut lire) avec une ligne par réglage, par exemple pour EasyDNS :

```
EASYDNS_TOKEN=paste-the-token-here
EASYDNS_KEY=paste-the-key-here
```

Sous macOS et Linux, rendez-le privé : `chmod 600 /etc/open-caption/dns.env`. Le programme avertit si d'autres utilisateurs peuvent le lire.

**4. Modifiez `config.yaml`** (les premières lignes affichées par le programme disent où il se trouve ; chaque réglage est expliqué dans la [référence](11-configuration-reference.md)) :

```yaml
https:
  enabled: yes
  acme:
    enabled: yes
    domain: captions.mytheater.org
    email: webmaster@mytheater.org      # Let's Encrypt writes here before a certificate expires
    agree_to_terms: yes                 # you accept Let's Encrypt's terms of service
    directory: letsencrypt-staging      # practice first (see step 5)
    dns:
      provider: easydns                 # or cloudflare, godaddy
      credentials_file: /etc/open-caption/dns.env
```

**5. Entraînez-vous d'abord avec le service d'essai.** `letsencrypt-staging` est le service de test de Let's Encrypt : il fonctionne exactement comme le vrai, mais ses certificats **ne sont pas approuvés par les navigateurs** (un avertissement est normal) et ses limites sont généreuses. Démarrez le programme.

**6. Vérifiez que cela a marché.** Ouvrez **Configuration → Réglages → Sécurité → Certificat**. L'état passe de *En attente du premier certificat* à **Valide** en quelques minutes, et affiche le nom, l'émetteur et la date d'expiration. Le journal du programme dit `HTTPS: got a certificate for … valid until …`. S'il dit *La dernière tentative a échoué*, lisez **Dernier problème** et le tableau ci-dessous.

**7. Passez au vrai certificat.** Remplacez `directory: letsencrypt-staging` par `directory: letsencrypt` (ou supprimez la ligne) et redémarrez le programme. Le vrai certificat arrive de la même façon. Ouvrez `https://captions.mytheater.org/` depuis un téléphone sur le WiFi du lieu : aucun avertissement. (Si le port 443 était occupé, le programme a utilisé le port de repli, `:8443` ; l'adresse demande alors le port, `https://captions.mytheater.org:8443/`, et le journal indique l'adresse exacte.)

**8. Donnez l'adresse aux gens.** L'étape Conduire de la configuration liste les adresses numériques et le nom `.local`, mais pas encore celle-ci. Pour imprimer une affiche avec code QR pour elle, une fois connecté à la configuration, ouvrez
`http://localhost/admin?print=qr&url=https%3A%2F%2Fcaptions.mytheater.org%2F` (changez le nom ; ajoutez le port si besoin) et imprimez cette page.

## Ensuite

- **Le renouvellement est automatique.** Le programme vérifie environ deux fois par jour et renouvelle quand il reste 30 jours ou moins. Vous n'avez rien à faire.
- **Si Internet tombe dans le lieu,** rien ne casse : le programme continue de servir le certificat qu'il a et réessaie une heure plus tard. Avant l'arrivée du tout premier certificat (ou s'il n'arrive jamais), les téléphones reçoivent l'auto-signé.
- **Renouveler maintenant** (Configuration → Réglages → Sécurité → Certificat) obtient un nouveau certificat tout de suite. Utilisez-le après avoir réglé un problème.
- **Changer le jeton DNS :** modifiez `/etc/open-caption/dns.env`. Il est relu à la prochaine tentative, donc aucun redémarrage n'est nécessaire.
- **Le désactiver :** mettez `acme: enabled: no` et redémarrez. HTTPS utilise alors le certificat auto-signé (ou vos propres fichiers).

## En cas de problème

La configuration affiche **Dernier problème** ; le journal contient le même texte. Les cas courants :

| Vous voyez | Cause probable et solution |
|---|---|
| `the DNS provider "…" is not built into this program` (le programme s'arrête au démarrage) | Cette version du programme a été compilée sans ce fournisseur. Utilisez `cloudflare`, `godaddy` ou `easydns`, ou procurez-vous une version avec le vôtre (plus bas) |
| `… credentials information are missing: EASYDNS_TOKEN,EASYDNS_KEY` | Les noms dans le fichier sont faux ou le fichier est illisible. Les noms sont en majuscules, `NOM=valeur`, un par ligne |
| Une erreur du fournisseur DNS comme *unauthorized*, *forbidden* ou un numéro 4xx | Le jeton ou la clé sont faux, expirés, ou n'ont pas le droit de modifier cette zone. Créez-en de nouveaux |
| *En attente du premier certificat* très longtemps, ou un message disant que l'enregistrement est introuvable | Le nouvel enregistrement DNS met du temps à se propager. Ajoutez `propagation_wait: 3m` sous `dns:` et appuyez sur **Renouveler maintenant** |
| `too many certificates already issued` ou `rate limit` | Let's Encrypt limite les répétitions. Attendez (le message dit combien de temps) et utilisez `letsencrypt-staging` pendant vos essais |
| Le navigateur avertit toujours | Vous utilisez le service d'essai (passez à `letsencrypt`), ou vous avez ouvert l'adresse numérique ou un autre nom au lieu de `https://captions…`, ou l'horloge de l'ordinateur ou du téléphone est fausse |
| Un téléphone ne peut pas ouvrir le nom | Certains réseaux WiFi invités et routeurs bloquent les noms publics qui répondent avec des adresses privées (« protection contre le DNS rebinding »). Demandez au responsable du réseau d'autoriser `captions.mytheater.org`, ou utilisez l'adresse numérique et le code QR comme avant |
| `https.acme: cannot be used together with cert_file/key_file` | Vous avez réglé les deux. Retirez-en un : certificats automatiques **ou** vos propres fichiers |
| Le programme refuse de démarrer et liste des problèmes dans `config.yaml` | Chaque ligne nomme le réglage et ce qui ne va pas. Corrigez-les tous et redémarrez |

## Des identifiants plus sûrs (facultatif)

Un jeton DNS capable de modifier toute votre zone est un risque si l'ordinateur est perdu ou partagé. Deux façons de le réduire :

- Gardez les secrets **hors de `config.yaml`** : dans le fichier d'identifiants (étape 3), ou comme variables d'environnement du service qui lance le programme (`EASYDNS_TOKEN_FILE=/path/to/file` fonctionne aussi : le réglage est lu dans ce fichier).
- Utilisez la **délégation par CNAME** : dans votre DNS, ajoutez un enregistrement permanent `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.other-zone.example`, où `other-zone.example` est une petite zone par ailleurs vide. Le programme le suit automatiquement ; le jeton n'a donc besoin d'accéder qu'à cette petite zone, pas à votre zone principale.

Les identifiants ne sont jamais écrits dans le journal, le paquet d'assistance ni la configuration.

## Un autre fournisseur DNS

Le programme ne contient que les fournisseurs DNS avec lesquels il a été compilé. Ceux qui le compilent peuvent en choisir d'autres, ou tous, avec `make build ACME_PROVIDERS="…"` (les noms sont ceux de la [liste de lego](https://go-acme.github.io/lego/dns/)) ; les tailles et les étapes sont dans le [document 10](10-acme-lets-encrypt.md). Si votre fournisseur n'est pas pris en charge du tout, `exec` (lance un script de votre choix) et `httpreq` (appelle une adresse web) peuvent piloter presque n'importe quel service DNS.
