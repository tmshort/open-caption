# Réseau et connexion

Les spectateurs connectent leurs propres téléphones et tablettes au **WiFi invités** de la salle et ouvrent une adresse. Rien n'est à installer.

## Ce que la salle doit fournir

- Cet ordinateur et les appareils du public doivent être sur le même réseau, ou le réseau doit permettre aux appareils invités d'atteindre cet ordinateur. C'est un réglage du système WiFi de la salle, pas de ce programme.
- Beaucoup de réseaux invités isolent les appareils les uns des autres. Ce n'est pas un problème. Les appareils ne parlent qu'à cet ordinateur, jamais entre eux, tant que cet ordinateur est joignable.
- Une connexion filaire est la meilleure pour cet ordinateur. Demandez une adresse réservée (une « réservation DHCP ») pour que l'adresse ne change pas d'un spectacle à l'autre.
- Aucune connexion internet n'est nécessaire.

Testez avant le premier spectacle : connectez un téléphone au WiFi invités et ouvrez l'adresse affichée à l'étape **Partager**. Pour un projecteur ou un grand écran, la même étape propose un lien projecteur : ouvrez-le sur l'ordinateur auquel le projecteur est branché.

## Rendre l'adresse facile

1. **Code QR** — imprimez l'affiche depuis l'étape Partager et placez-la dans le hall et dans le programme. C'est l'option la plus simple.
2. **Une adresse courte** — si le programme peut utiliser le port 80, l'adresse n'a pas de numéro de port, comme `http://192.168.1.10/`.
3. **Un nom** — le programme s'annonce sous le nom **`open-caption.local`**, de sorte que `http://open-caption.local/` fonctionne sur la plupart des téléphones et ordinateurs du même réseau, sans rien configurer. Changez le nom, ou désactivez-le, dans **Réglages** (bouton en haut), onglet **Réseau**, sous **Nom sur le réseau** (la case **Utiliser un nom** et le nom). Certains réseaux WiFi invités bloquent ce type d'annonce ; utilisez alors l'adresse numérique ou le code QR. La salle peut aussi ajouter sur son routeur un nom à elle (comme `captions.theater`), qui fonctionne partout.
4. **Page automatique** — une salle peut envoyer les invités vers la page quand ils rejoignent le WiFi grâce à la page d'accueil du point d'accès (« portail captif »). Le programme répond aussi à l'adresse standard de détection de réseau `/captive-api` pour les salles qui utilisent l'option 114 de DHCP, mais cela indique seulement que le réseau n'est pas captif : cela n'ouvre pas la page tout seul, et cela exige un certificat HTTPS de confiance. Voir le [guide réseau de la salle](topic:admin-venue-network-guide).

## Si cet ordinateur a plus d'une adresse

Un ordinateur branché par câble et en WiFi à la fois, ou avec des machines virtuelles ou un VPN, a plusieurs interfaces réseau. Ouvrez **Réglages** (bouton en haut) et son onglet **Réseau** : il liste chaque interface connectée avec une case. Seules les interfaces cochées sont utilisées, et le public ne reçoit que leurs adresses (dans la liste des adresses, le code QR et le nom `.local`). Ne cochez que les réseaux que le public utilise. Si deux interfaces cochées sont sur le même réseau (un câble et le WiFi, tous deux avec une adresse 192.168.1.x), une seule adresse est proposée, pour que le code QR et le nom `.local` concordent : la filaire, sinon celle de la connexion principale. Pour utiliser l'autre, décochez la première. La case est enregistrée dès que vous cliquez, et l'interface par laquelle vous utilisez la configuration ne peut pas être décochée. Une nouvelle interface (un adaptateur USB, un VPN) reste désactivée jusqu'à ce que vous la cochiez. **Refuser aussi les connexions aux interfaces qui ne sont pas cochées** va plus loin : un téléphone qui utilise une adresse d'une interface décochée est renvoyé ; les connexions à cet ordinateur lui-même sont toujours acceptées, la configuration reste donc accessible.

## HTTP et HTTPS

Les adresses `http://` simples fonctionnent sur tous les appareils et sont le choix par défaut. HTTPS peut être activé dans `config.yaml` ; avec un certificat auto-signé, chaque appareil affiche un avertissement de sécurité la première fois. Le HTTP simple est plus simple pour un public. Lorsque HTTPS est activé, l'étape **Partager** liste les adresses en deux groupes, *Sécurisée (https)* et *Non chiffrée (http)* ; avec un certificat pour un vrai nom (automatique ou le vôtre), ce nom est listé en premier et c'est par lui que commencent la grande adresse, **Copier l'adresse** et le code QR. Les appareils doivent pouvoir trouver cet ordinateur par ce nom (par exemple grâce au service de noms du routeur de la salle).

## Si le port est occupé

Si le port 80 est déjà utilisé par un autre programme, le serveur démarre sur un autre port et l'indique dans la console. Les adresses du public doivent alors l'inclure, par exemple `http://192.168.1.10:8080/` ; l'étape Partager affiche la bonne adresse.
