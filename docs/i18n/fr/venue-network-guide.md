# Guide réseau de la salle

Pour la personne qui s'occupe du WiFi du théâtre. L'ordinateur de sous-titrage n'a besoin de rien d'internet. Il doit seulement être **joignable depuis les appareils du public**.

## Ce que fait le public

1. Se connecter au WiFi invités.
2. Ouvrir une adresse, par exemple `http://192.168.1.10/` (ou scanner le code QR).

## Exigences

| Exigence | Pourquoi |
|----------|----------|
| Le réseau invités peut joindre l'ordinateur de sous-titrage | Les appareils récupèrent la page et gardent une connexion ouverte avec lui |
| L'ordinateur a une adresse fixe (réservation DHCP ou statique) | Pour que les codes QR et affiches imprimés restent valables d'un spectacle à l'autre |
| Le point d'accès gère plus de 100 appareils simultanés (5 GHz activé) | Une salle comble, ce sont environ 150 appareils |
| Aucune liaison internet n'est nécessaire | Tout est servi par l'ordinateur de sous-titrage |

**L'isolation des clients entre invités est acceptable et recommandée.** Les appareils du public ne parlent qu'à l'ordinateur de sous-titrage, jamais entre eux. Ce qui ne doit *pas* être bloqué, c'est le trafic des invités vers le serveur. Si le réseau invités est isolé du réseau local principal, branchez l'ordinateur de sous-titrage sur le réseau invités (un port libre du VLAN invités ou le SSID invités) ou ajoutez une règle de pare-feu autorisant les invités à joindre l'adresse de l'ordinateur sur le port TCP 80 (et 443 si HTTPS est activé).

## Vérifier avant le spectacle

Depuis un téléphone sur le WiFi invités, ouvrez l'adresse affichée dans Configuration → **Partager**. Si la page se charge et dit « Les sous-titres vont bientôt commencer » (ou montre l'avant-spectacle), c'est bon. Sinon, voir « Rien ne se charge » plus bas.

## Rendre l'adresse plus facile

Par ordre d'effort :

1. **Code QR** — Configuration → Partager → « Imprimer une affiche avec code QR ». Placez-la dans le hall et dans le programme. Aucun changement de réseau n'est nécessaire.
2. **Port 80** — si le programme tourne sur le port 80, l'adresse n'a pas de `:port`. Sous Linux, autorisez-le avec `sudo setcap 'cap_net_bind_service=+ep' ./open-caption`.
3. **Un nom local** — ajoutez sur le routeur une entrée DNS pour l'ordinateur de sous-titrage, par exemple `captions.theater` → `192.168.1.10`. La plupart des routeurs (OpenWrt, Pi-hole, UniFi, Omada, à base de dnsmasq) permettent un « enregistrement DNS local » ou une « substitution d'hôte ».
4. **Envoyer automatiquement les invités là-bas.** Options :
   - **Page d'accueil / portail captif** sur le point d'accès pointant vers l'adresse du public. La fonction « portail invités » intégrée du point d'accès fonctionne avec une URL externe (UniFi : Hotspot → portail externe ; OpenWrt : nodogsplash `RedirectURL`). Assurez-vous qu'elle laisse la page se charger sans connexion.
   - **Option 114 du DHCP (RFC 8910) et `/captive-api`.** Le programme répond à `GET /captive-api` comme le décrit la RFC 8908 : `{"captive": false, "user-portal-url": "http://<server-address>/"}`. Si votre réseau distribue déjà l'option 114 (« Captive-Portal API URI »), vous pouvez la faire pointer vers cette adresse. Sachez ce que cela fait : cela dit aux appareils que le réseau n'est *pas* un portail captif et publie l'adresse de la page. Cela ne fait **pas** ouvrir la page de sous-titres aux téléphones d'eux-mêmes (il faudrait `"captive": true`, que le programme n'envoie jamais). La RFC exige aussi que l'adresse de l'API soit en **https avec un certificat auquel les appareils font confiance** : une adresse d'option 114 en `http://` simple ou le certificat auto-signé du programme sera donc ignoré. C'est expérimental et non testé sur de vrais appareils : utilisez l'affiche QR, ou la page d'accueil de votre point d'accès, pour amener les gens à la page.
5. **mDNS** (`open-caption.local`) est fourni et activé par défaut : les appareils qui peuvent l'utiliser ouvrent `http://open-caption.local/` sans rien configurer sur le routeur. Il faut que le réseau invités autorise le **multicast** (UDP 5353 vers 224.0.0.251) entre l'ordinateur de sous-titrage et les appareils du public. Beaucoup de réseaux invités le bloquent (« filtrage multicast » ou « blocage mDNS/Bonjour » sur le point d'accès), et le nom ne se résout alors simplement pas ; l'adresse numérique et le code QR fonctionnent toujours. Si vous voulez que le nom fonctionne, activez la redirection mDNS/Bonjour ou désactivez le filtrage multicast pour le réseau invités. Le nom peut être changé, ou désactivé, dans la configuration, sous Réglages → Réseau (*Nom sur le réseau*).

Le programme n'exécute pas son propre serveur DHCP ou DNS, volontairement : il entrerait en conflit avec le routeur de la salle.

## HTTP et HTTPS

Utilisez du `http://` simple pour le public. Un certificat HTTPS auto-signé fait afficher un avertissement de sécurité à chaque appareil. HTTPS est disponible (`https.enabled: true` dans `config.yaml`) pour les sites qui peuvent installer un certificat auquel leurs appareils font confiance. Si le lieu possède un nom de domaine, le programme peut obtenir et renouveler lui-même un certificat de confiance ; voir le [guide du certificat HTTPS de confiance](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

Un petit inconvénient du HTTP simple : les navigateurs n'autorisent « garder l'écran allumé » que sur les pages sécurisées. En HTTP, la page garde l'écran du téléphone allumé avec une minuscule vidéo en boucle, silencieuse et invisible. Un téléphone en mode économie d'énergie peut la refuser ; la page affiche alors, une seule fois, une astuce demandant de désactiver le verrouillage automatique.

## Sécurité

- La configuration (`/admin`) et la page de l'opérateur (`/operator`) exigent chacune un mot de passe tiré de `config.yaml`. La page du public n'en exige aucun et ne peut rien modifier. Un mot de passe spectateur facultatif ouvre la page de suivi (`/follow`), qui permet seulement de regarder. Limitez les ordinateurs qui peuvent l'ouvrir avec `access.viewer_allow`.
- N'exposez pas le port sur internet.
- Les mots de passe opérateur et administrateur sont conservés sous forme d'empreintes dans `config.yaml`, dans le dossier de données du programme, donc impossibles à relire : notez-les quand ils s'affichent, ou définissez-en de nouveaux dans la configuration. Gardez ce dossier privé pour les personnes qui conduisent le spectacle.

## Pare-feu sur l'ordinateur de sous-titrage

- **Windows :** au premier démarrage du programme, autorisez-le sur les **réseaux privés**. Si l'invite a été ignorée : Sécurité Windows → Pare-feu → Autoriser une application.
- **macOS :** autorisez les connexions entrantes pour le programme quand on vous le demande, ou ajoutez-le dans Réglages Système → Réseau → Coupe-feu.
- **Linux :** ouvrez le port, par exemple `sudo ufw allow 80/tcp`.

## Rien ne se charge

1. Le téléphone est-il sur le WiFi invités (pas sur les données mobiles) ?
2. L'adresse inclut-elle le port si le programme n'est pas sur le port 80 ? L'étape Partager montre la bonne adresse.
3. Un autre ordinateur du WiFi invités peut-il le joindre (`curl http://192.168.1.10/healthz` doit afficher `ok`) ?
4. Le trafic des invités vers le serveur est-il bloqué par le routeur ou le point d'accès (isolation des clients réglée sur « tout bloquer » au lieu de « bloquer entre clients ») ?
5. Le pare-feu de l'ordinateur bloque-t-il le programme ?
6. L'ordinateur est-il toujours éveillé, et toujours à la même adresse ? Cherchez un changement d'adresse dans Partager.

Configuration → Dépannage → **Télécharger un paquet de support** rassemble les réglages (sans mots de passe), les dernières lignes du journal et des informations de base sur l'ordinateur.

## Capacité

Le programme a été testé en charge avec 200 et 500 connexions de public simultanées sur un seul ordinateur (latence inférieure à 10 ms sur la même machine). Dans une vraie salle, la limite est le WiFi, pas le programme : utilisez un point d'accès prévu pour la foule, activez le 5 GHz et gardez les images petites (moins de 1 Mo) pour que de nombreux appareils puissent les récupérer en même temps.
