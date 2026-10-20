# Guia de rede do local

Para quem cuida do WiFi do teatro. O computador das legendas não precisa de nada da internet. Só precisa ser **alcançável pelos dispositivos do público**.

## O que o público faz

1. Conecta-se ao WiFi de visitantes.
2. Abre um endereço, por exemplo `http://192.168.1.10/` (ou escaneia o código QR).

## Requisitos

| Requisito | Por quê |
|-----------|---------|
| A rede de visitantes consegue alcançar o computador das legendas | Os dispositivos buscam a página e mantêm uma conexão aberta com ele |
| O computador tem um endereço fixo (reserva DHCP ou estático) | Para que códigos QR e cartazes impressos continuem válidos entre espetáculos |
| O ponto de acesso aguenta mais de 100 dispositivos ao mesmo tempo (com 5 GHz ativado) | Uma sala cheia são cerca de 150 dispositivos |
| Não é preciso saída para a internet | Tudo é servido pelo computador das legendas |

**O isolamento de clientes entre visitantes é aceitável e recomendado.** Os dispositivos do público falam só com o computador das legendas, nunca entre si. O que *não* pode ser bloqueado é o tráfego dos visitantes para o servidor. Se a rede de visitantes é isolada da rede local principal, ligue o computador das legendas à rede de visitantes (uma porta livre da VLAN de visitantes ou o SSID de visitantes) ou acrescente uma regra de firewall que permita aos visitantes alcançar o endereço do computador pela porta TCP 80 (e 443 se o HTTPS estiver ativado).

## Conferir antes do espetáculo

De um telefone no WiFi de visitantes, abra o endereço mostrado em Configuração → **Compartilhar**. Se a página carregar e disser «As legendas começarão em breve» (ou mostrar o pré-espetáculo), está tudo certo. Se não, veja «Nada carrega» mais abaixo.

## Deixar o endereço mais fácil

Em ordem de esforço:

1. **Código QR** — Configuração → Compartilhar → «Imprimir um cartaz com código QR». Ponha no saguão e no programa. Não é preciso mudar nada na rede.
2. **Porta 80** — se o programa roda na porta 80, o endereço não tem `:porta`. No Linux, permita-o com `sudo setcap 'cap_net_bind_service=+ep' ./open-caption`.
3. **Um nome local** — acrescente no roteador uma entrada DNS para o computador das legendas, por exemplo `captions.theater` → `192.168.1.10`. A maioria dos roteadores (OpenWrt, Pi-hole, UniFi, Omada, baseados em dnsmasq) permite um «registro DNS local» ou «substituição de host».
4. **Mandar os visitantes para lá automaticamente.** Opções:
   - **Página de boas-vindas / portal cativo** no ponto de acesso apontando para o endereço do público. O recurso de «portal de visitantes» embutido do ponto de acesso funciona com uma URL externa (UniFi: Hotspot → portal externo; OpenWrt: nodogsplash `RedirectURL`). Garanta que ele deixe a página carregar sem login.
   - **Opção 114 do DHCP (RFC 8910) e `/captive-api`.** O programa responde a `GET /captive-api` como a RFC 8908 descreve: `{"captive": false, "user-portal-url": "http://<server-address>/"}`. Se a sua rede já distribui a opção 114 («Captive-Portal API URI»), você pode apontá-la para esse endereço. Saiba o que isso faz: diz aos dispositivos que a rede *não* é um portal cativo e publica o endereço da página. **Não** faz os telefones abrirem sozinhos a página de legendas (isso exigiria `"captive": true`, que o programa nunca envia). A RFC também exige que o endereço da API seja **https com um certificado em que os dispositivos confiem**, então um endereço de opção 114 em `http://` simples ou o certificado autoassinado do programa será ignorado. Isso é experimental e não foi testado em dispositivos reais: use o cartaz QR, ou a página de boas-vindas do seu ponto de acesso, para levar as pessoas à página.
5. **mDNS** (`open-caption.local`) é fornecido e vem ativado por padrão: os dispositivos que conseguem usá-lo abrem `http://open-caption.local/` sem configurar nada no roteador. Exige que a rede de visitantes permita **multicast** (UDP 5353 para 224.0.0.251) entre o computador das legendas e os dispositivos do público. Muitas redes de visitantes o bloqueiam («filtragem de multicast» ou «bloqueio de mDNS/Bonjour» no ponto de acesso), e então o nome simplesmente não resolve; o endereço numérico e o código QR continuam funcionando. Se quiser que o nome funcione, ative o encaminhamento mDNS/Bonjour ou desligue a filtragem de multicast para a rede de visitantes. O nome pode ser mudado, ou desligado, na configuração, em Configurações → Rede (*Nome na rede*).

O programa não executa seu próprio servidor DHCP ou DNS, de propósito: ele entraria em conflito com o roteador do local.

## HTTP e HTTPS

Use `http://` simples para o público. Um certificado HTTPS autoassinado faz cada dispositivo mostrar um aviso de segurança. O HTTPS está disponível (`https.enabled: true` em `config.yaml`) para locais que consigam instalar um certificado em que seus dispositivos confiem. Se o local tem um nome de domínio próprio, o programa pode obter e renovar sozinho um certificado confiável; veja o [guia do certificado HTTPS confiável](https-certificate-guide.md) (`docs/https-certificate-guide.md`).

Um pequeno custo do HTTP simples: os navegadores só permitem «manter a tela acesa» em páginas seguras. Com HTTP a página mantém acesa a tela do telefone com um pequeno vídeo em laço, silencioso e invisível. Um telefone em modo de economia de energia pode recusá-lo; a página então mostra, uma só vez, uma dica pedindo para desligar o bloqueio automático.

## Segurança

- A Configuração (`/admin`) e a página do operador (`/operator`) exigem cada uma uma senha de `config.yaml`. A página do público não exige nenhuma e não pode mudar nada.
- Não exponha a porta à internet.
- As senhas de operador e de administrador ficam em texto puro em `config.yaml`, na pasta de dados do programa. Mantenha essa pasta privada para quem conduz o espetáculo.

## Firewalls no computador das legendas

- **Windows:** quando o programa iniciar pela primeira vez, permita-o em **Redes privadas**. Se o aviso foi dispensado: Segurança do Windows → Firewall → Permitir um aplicativo.
- **macOS:** permita as conexões de entrada para o programa quando perguntado, ou adicione-o em Ajustes do Sistema → Rede → Firewall.
- **Linux:** abra a porta, por exemplo `sudo ufw allow 80/tcp`.

## Nada carrega

1. O telefone está no WiFi de visitantes (e não nos dados móveis)?
2. O endereço inclui a porta se o programa não está na porta 80? A etapa Compartilhar mostra o endereço correto.
3. Outro computador no WiFi de visitantes consegue alcançá-lo (`curl http://192.168.1.10/healthz` deve imprimir `ok`)?
4. O tráfego dos visitantes para o servidor está bloqueado pelo roteador ou ponto de acesso (isolamento de clientes em «bloquear tudo» em vez de «bloquear entre clientes»)?
5. O firewall do computador está bloqueando o programa?
6. O computador continua ligado, e no mesmo endereço? Procure uma mudança de endereço em Compartilhar.

Configuração → Compartilhar → **Baixar um pacote de suporte** reúne as configurações (sem senhas), linhas recentes do registro e dados básicos do computador.

## Capacidade

O programa foi testado sob carga com 200 e 500 conexões de público simultâneas em um só computador (latência abaixo de 10 ms na mesma máquina). Num local real o limite é o WiFi, não o programa: use um ponto de acesso dimensionado para a multidão, ative 5 GHz e mantenha as imagens pequenas (abaixo de 1 MB) para que muitos dispositivos possam buscá-las ao mesmo tempo.
