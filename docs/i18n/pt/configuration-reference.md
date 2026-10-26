# Referência do `config.yaml`

O `config.yaml` guarda tudo sobre *este computador e este local*: as senhas, as portas, o HTTPS, quem pode abrir a configuração e os valores padrão que um espetáculo usa quando o seu próprio arquivo não diz nada. Tudo sobre *um espetáculo* (título, cores, duração dos laços, mensagem pré-espetáculo…) fica, em vez disso, no arquivo desse espetáculo (veja [Escrever um arquivo do espetáculo](../../../web/help/admin-writing-a-show-file.md), «Configurações do espetáculo»); as camadas são explicadas mais abaixo em «Onde uma configuração é decidida».

## O arquivo

- **Onde fica.** `<pasta de dados>/config.yaml`; a pasta de dados é `data_dir`, `--data`, ou a padrão `<sua pasta de configuração de usuário>/open-caption` (no macOS `~/Library/Application Support/open-caption`, no Windows `%AppData%\open-caption`, no Linux `~/.config/open-caption`). As primeiras linhas que o programa imprime dizem qual arquivo ele está usando. `--config` indica outro arquivo.
- **Primeira execução.** Se não houver arquivo, o programa escreve um com senhas aleatórias de **administrador** e de **operador** e as imprime uma única vez no console. O arquivo tem modo 0600 (privado para a sua conta) e as senhas ficam em **texto simples**; mantenha-o privado.
- **Você só precisa das chaves que mudar.** Uma chave ausente mantém o valor padrão (as tabelas abaixo). O arquivo gerado lista tudo.
- **Rigoroso.** Uma chave escrita errada, um valor de tipo errado ou fora do intervalo faz o programa parar ao iniciar, listando **todos** os problemas, cada um com o nome da configuração (por exemplo `audience.max_clients: 0 is not a client count: use 1 to 100000`). Nada é adivinhado. A única exceção: num arquivo marcado com um `config_version` mais novo do que o programa conhece, as chaves desconhecidas são ignoradas, com um aviso.
- **A configuração edita uma parte.** Mudar as senhas, o nome `.local`, a mensagem de espera ou o nível do registro na configuração reescreve essas chaves no arquivo e mantém os seus comentários e todas as outras configurações. Todo o resto você edita à mão e depois **reinicia o programa**. (Um arquivo de certificado renovado é a única coisa aproveitada sem reiniciar; veja `https.cert_file`.)
- **Os booleanos** são `yes`/`no`/`true`/`false`. **As listas** se escrevem `[a, b]` ou com um `- item` por linha.

## Linha de comando e ambiente

Uma opção, ou a sua variável de ambiente, tem prioridade sobre o arquivo naquela execução. (O arquivo não é alterado.) Uma opção pode ser escrita com um ou dois hifens (`-data` ou `--data`). `open-caption --help` imprime esta lista, e a **página de manual** (`docs/open-caption.1`; leia com `man ./docs/open-caption.1`, ou copie para uma pasta `man1` como `/usr/local/share/man/man1/`) descreve as mesmas opções, além de arquivos, sinais e status de saída.

| Opção | Ambiente | Efeito |
|---|---|---|
| `--config FILE` | `OPEN_CAPTION_CONFIG` | Usa este arquivo de configurações (padrão `<pasta de dados>/config.yaml`) |
| `--data FOLDER` | `OPEN_CAPTION_DATA` | A pasta de dados (espetáculos, sessões, `live.json`, certificados). Tem prioridade sobre `data_dir` |
| `--http-addr ADDR` | `OPEN_CAPTION_HTTP_ADDR` | Tem prioridade sobre `http.addr`, por exemplo `:80` ou `:8080` |
| `--https-addr ADDR` | `OPEN_CAPTION_HTTPS_ADDR` | **Liga** o HTTPS neste endereço (tem prioridade sobre `https.enabled` e `https.addr`) |
| `--log-level LEVEL` | `OPEN_CAPTION_LOG_LEVEL` | Tem prioridade sobre `log.level` toda vez que o programa inicia (a configuração avisa quando está em vigor) |
| `--log-file FILE` | `OPEN_CAPTION_LOG` | Acrescenta também o registro a este arquivo |
| `--network-interfaces LIST` | `OPEN_CAPTION_NETWORK_INTERFACES` | Tem prioridade sobre `network.interfaces` toda vez que o programa inicia: uma lista separada por vírgulas de nomes de interface, como `Ethernet` ou `en0,en4` (a configuração avisa quando está em vigor) |
| `--network-restrict` | `OPEN_CAPTION_NETWORK_RESTRICT` | Também recusa conexões feitas a endereços que não são oferecidos, toda vez que o programa inicia (como `network.restrict_listening: yes`). A variável vale quando é `yes`, `true`, `on` ou `1` |
| `--no-browser` | | Não abre a configuração em um navegador ao iniciar (como `defaults.open_browser: no`) |
| `--version` | | Imprime a versão e os provedores de DNS incluídos para os certificados automáticos, e sai |
| `--help` | | Imprime as opções e as variáveis de ambiente, e onde ler mais, e sai |
| `--check-show FILE [TRANSLATION ...]` | | **Confere um arquivo do espetáculo** e imprime um relatório, como um «lint», e sai: sem servidor, sem `config.yaml`. Veja «Conferir um arquivo do espetáculo» mais abaixo |
| `--assets FOLDER` | | A pasta de imagens para `--check-show` (padrão: a pasta `assets` ao lado do arquivo do espetáculo) |
| `--strict` | | Com `--check-show`: os avisos também dão o status de saída 1 |
| `--dev`, `--show-file` | | **Somente para desenvolvedores**: pontos de teste sem autenticação neste computador, sem espera; nunca use em um local |

## Conferir um arquivo do espetáculo

`open-caption --check-show show.md` analisa um arquivo do espetáculo sem carregá-lo no programa: executa o mesmo analisador e as mesmas verificações da lista de problemas da etapa Arquivo do espetáculo da configuração e imprime o que essa lista mostraria.

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

- Cada problema tem a forma `file:line: severity: message` (gravidade `error`, `warning` ou `note`), a que os editores e os sistemas de integração contínua entendem. Uma linha sobre o arquivo inteiro não tem número de linha.
- Os arquivos depois do primeiro são **traduções**: cada um é analisado como tradução e comparado com o arquivo principal (o mesmo número e tipo de marcações, na mesma ordem), como faz a configuração. O idioma é a linha `[language]` do arquivo ou o seu nome (`show.es.md`); dois arquivos para o mesmo idioma são um erro.
- As imagens são conferidas na pasta `assets` ao lado do arquivo principal, ou em `--assets FOLDER`. Sem pasta, não são conferidas, e o relatório diz isso. Uma imagem ausente nomeia a pasta em que foi procurada (o editor da configuração diz, em vez disso, para enviá-la na etapa Imagens).
- Uma imagem da pasta que nenhum arquivo do espetáculo mostra é listada como `unused images (…): name.png` e contada como nota, não como aviso (a etapa Imagens da configuração diz «Não usada» embaixo). Imagens usadas apenas por uma tradução contam como usadas.
- `-` como nome de arquivo lê a entrada padrão.
- **Status de saída:** `0` sem erros, `1` erros (ou avisos com `--strict`), `2` um arquivo não pôde ser lido. Coloque as outras opções (`--assets`, `--strict`) **antes** de `--check-show`: os arquivos que vêm depois são traduções.

## Configurações

Os valores padrão são os que você obtém sem a chave. «Configuração» diz onde a configuração pode alterá-la.

### Nível superior

| Chave | Padrão | Significado |
|---|---|---|
| `config_version` | `1` | Escrita pelo programa. Um número maior do que o programa conhece significa que o arquivo veio de uma versão mais nova |
| `admin_password` | gerada | A senha da configuração (`/admin`). Não pode ficar vazia (6 ou mais caracteres ao mudar na configuração). *Configuração: Configurações → Segurança* |
| `operator_password` | gerada | A senha da página do operador (`/operator`); a do administrador também vale ali. Não pode ficar vazia. *Configuração: Configurações → Segurança* |
| `data_dir` | vazio | A pasta de dados. Vazio significa a pasta padrão acima. Não pode ser um arquivo |

### `http`

| Chave | Padrão | Significado |
|---|---|---|
| `http.addr` | `:80` | Onde o servidor HTTP sem criptografia escuta: `:80` (todos os endereços, porta 80), `192.168.1.20:80` ou `127.0.0.1:8080` (somente este computador). Não pode ficar vazio |
| `http.fallback_addr` | `:8080` | Usado, com um aviso visível, se `addr` não puder ser usado (a porta 80 costuma estar ocupada ou exigir privilégios). Vazio significa sem alternativa |

### `https`

O HTTPS vem **desligado** por padrão. Há três jeitos de obter um certificado; escolha um.

| Chave | Padrão | Significado |
|---|---|---|
| `https.enabled` | `no` | Serve HTTPS além de HTTP, em portas separadas (não há redirecionamento). Precisa de um dos três abaixo |
| `https.addr` | `:443` | A porta HTTPS. Obrigatória quando ligado |
| `https.fallback_addr` | `:8443` | Usada se `addr` não puder ser aberta |
| `https.self_signed` | `yes` | Sem arquivos de certificado e sem `acme`: o programa cria um certificado próprio (guardado em `<pasta de dados>/tls/`). Os navegadores avisam sobre ele até que seja aceito. `no` sem arquivos é um erro |
| `https.cert_file`, `https.key_file` | vazio | O seu próprio certificado e a chave (PEM). Defina **os dois** ou nenhum; os dois arquivos precisam existir. **Os arquivos são vigiados**: quando são substituídos (por exemplo pelo `certbot`), o novo par é usado nas novas conexões em cerca de 30 segundos, sem reiniciar. Um par que não pode ser carregado mantém o certificado antigo e registra um aviso |

### `https.acme` (certificados automáticos)

Desligado, salvo `enabled: yes`. Precisa de `https.enabled: yes`, de um nome de domínio público seu e de um provedor de DNS com o qual este programa foi compilado. Não pode ser combinado com `cert_file`/`key_file`. Como configurar, passo a passo: o [guia do certificado HTTPS confiável](https-certificate-guide.md); as opções de provedores de DNS ao compilar, e os tamanhos, estão no final desse guia.

| Chave | Padrão | Significado |
|---|---|---|
| `https.acme.enabled` | `no` | Obtém e renova um certificado automaticamente (ACME, DNS-01) |
| `https.acme.domain` | | O nome público, por exemplo `captions.example.org`. Um só nome: sem curinga, sem `.local`/`.lan`/`.internal`, sem endereço IP. Ele precisa resolver, para os telefones do público, ao endereço deste computador |
| `https.acme.email` | | Para onde a autoridade certificadora envia os avisos de vencimento |
| `https.acme.agree_to_terms` | `no` | **Precisa ser `yes`**: registra que você aceita os termos de serviço da autoridade certificadora |
| `https.acme.directory` | `letsencrypt` | `letsencrypt`, `letsencrypt-staging` (para experimentar: os navegadores não confiam nele) ou o endereço `https://` de outro diretório ACME |
| `https.acme.renew_before_days` | `30` | Renova quando faltam estes dias ou menos (1 a 60) |
| `https.acme.dns.provider` | | O provedor de DNS, pelo seu [nome no lego](https://go-acme.github.io/lego/dns/) (`cloudflare`, `godaddy`, `easydns`, …). Precisa estar incluído no programa (`--version` lista os incluídos) |
| `https.acme.dns.credentials` | vazio | As configurações próprias do provedor, como um mapa: para o EasyDNS `EASYDNS_TOKEN: …` e `EASYDNS_KEY: …`. Os nomes são letras maiúsculas, dígitos e sublinhados. O ambiente e `credentials_file` prevalecem sobre estas |
| `https.acme.dns.credentials_file` | vazio | Um arquivo de linhas `NAME=value` com as mesmas configurações, lido de novo a cada tentativa. Precisa existir; proteja-o com `chmod 600` |
| `https.acme.dns.propagation_wait` | vazio | Uma espera fixa (`90s`, `2m`, no máximo `30m`) para o novo registro DNS se propagar, em vez de verificar |

### `defaults`

O que um espetáculo usa quando o seu próprio arquivo não diz. O bloco `[settings]` do arquivo do espetáculo prevalece sobre todos os valores daqui (veja [Escrever um arquivo do espetáculo](../../../web/help/admin-writing-a-show-file.md)). A etapa Aparência da configuração escreve o arquivo do espetáculo, não estes.

| Chave | Padrão | Significado |
|---|---|---|
| `defaults.language` | `en` | O idioma principal de um espetáculo cujo arquivo não indica nenhum: duas ou três letras, com região opcional (`en`, `es`, `pt-br`) |
| `defaults.loop_item_seconds` | `10` | Quanto tempo cada item de um laço pré-espetáculo ou de intervalo fica, em segundos (1 a 3600), a menos que o item diga outra coisa |
| `defaults.idle_message` | `Captions will begin shortly` | O que o público lê antes de a primeira linha ser enviada quando um espetáculo não tem mensagem pré-espetáculo (até 200 caracteres) |
| `defaults.open_browser` | `yes` | Abre a configuração em um navegador quando o programa inicia (não é uma configuração do espetáculo; `--no-browser` a desliga por uma execução) |
| `defaults.look.background` | `#000000` | O fundo das legendas: `#rgb`, `#rrggbb` ou `black`, `white`, `yellow`, `red`, `green`, `blue`, `cyan`, `magenta`, `orange`, `gray` |
| `defaults.look.text_color` | `#ffffff` | A cor do texto das legendas (mesmas formas) |
| `defaults.look.speaker_name_color` | `#ffffff` | A cor dos nomes dos personagens (mesmas formas) |
| `defaults.look.font` | `system` | `system`, `serif`, `mono`, `dyslexic` ou `hyperlegible` |
| `defaults.look.text_size_percent` | `100` | Tamanho do texto das legendas, de 50 a 300 |
| `defaults.look.speaker_names` | `label` | `label` (nomes em uma só cor), `color` (uma cor diferente para cada personagem) ou `hidden` |
| `defaults.look.speaker_position` | `above` | `above` (o nome acima da linha), `beside` (na mesma linha, antes do texto) ou `column` (em uma coluna à esquerda do texto) |

### `audience`

| Chave | Padrão | Significado |
|---|---|---|
| `audience.max_clients` | `500` | O máximo de dispositivos do público conectados **por idioma** (1 a 100000); os demais são recusados até que alguns saiam |
| `audience.scrollback_lines` | `200` | Quantas linhas anteriores um dispositivo pode rolar para trás (20 a 2000) |

### `access`

Quem pode abrir a configuração e a página do operador. A página do público é sempre aberta a todos.

| Chave | Padrão | Significado |
|---|---|---|
| `access.admin_allow` | `[localhost]` | Computadores autorizados a abrir a configuração; por padrão, apenas o computador que executa o programa. Uma lista vazia (`[]`) significa que **qualquer computador da rede pode tentar a senha**; o programa avisa sobre isso. Entradas: `localhost`, `private` (as redes domésticas e de escritório usuais), um endereço (`192.168.1.50`) ou uma rede (`192.168.1.0/24`). Exemplo: `[localhost]` |
| `access.operator_allow` | `[]` | O mesmo, para a página do operador |
| `access.local_auto_login` | `yes` | Um navegador no computador que executa o programa (aberto em `localhost`, `127.0.0.1` ou `[::1]`) entra como administrador sem senha. Os outros computadores sempre usam senha. Desligue em um computador em que outras pessoas possam entrar |

### `log`, `mdns`, `standby`

| Chave | Padrão | Significado |
|---|---|---|
| `log.level` | `info` | `debug` (cada requisição), `info` (eventos e entradas), `warn` ou `error`. *Configuração: Configurações → Sobre* |
| `mdns.enabled` | `yes` | Anuncia este computador como `<nome>.local` na rede local |
| `mdns.name` | `open-caption` | O nome: somente letras, dígitos e hifens, até 63 caracteres (`.local` é acrescentado). *Configuração: Configurações → Rede* |
| `standby.message` | vazio | A linha que o público vê na tela de espera. Vazio significa «Captions will begin shortly» no idioma de cada espectador. Uma linha, até 200 caracteres. *Configuração: Configurações → Tela do público* (o logotipo é um arquivo escolhido ali, não uma configuração) |

### `network`

| Chave | Padrão | Significado |
|---|---|---|
| `network.interfaces` | `[]` | As interfaces de rede usadas, por nome (`Ethernet`, `Wi-Fi`, `en0`, `eth0`): só os endereços delas são dados ao público, listados na configuração, colocados em um novo certificado HTTPS autoassinado e anunciados como `<nome>.local`. Vazio significa que ainda não houve escolha: todas as interfaces são usadas. Feita a escolha, uma interface que aparece depois (um adaptador USB, uma VPN) fica desligada até ser adicionada. Quando duas interfaces usadas estão na mesma rede (um cabo e o Wi-Fi ao mesmo tempo), só um endereço é oferecido: o da cabeada e, se não, o da conexão principal; deixe a outra de fora para escolher. *Configuração: Configurações → Rede* |
| `network.restrict_listening` | `no` | Também **recusa conexões** feitas a um endereço de uma interface que não é usada (um telefone que a usa é recusado), em vez de apenas deixar o endereço fora das listas. Conexões ao próprio computador (`localhost`) são sempre atendidas, então a configuração pode ser aberta para mudar isto. A configuração nunca desliga a interface pela qual você a está usando. *Configuração: Configurações → Rede* |

Os endereços de uma interface que não é usada não aparecem na etapa Compartilhar nem na inicialização, não são dados ao público, não são anunciados como `<nome>.local` e não entram em um novo certificado HTTPS autoassinado. Só as interfaces conectadas são listadas, e só um endereço é oferecido por rede. Por padrão isto só escolhe o que é *oferecido*: o servidor continua aceitando conexões em todos os endereços, até que `network.restrict_listening` seja ligado.

## Onde uma configuração é decidida

Do mais fraco ao mais forte: o valor embutido do programa → `config.yaml` (`defaults`, `defaults.look`, `audience.scrollback_lines`) → o `[settings]` do arquivo do espetáculo. Uma opção da linha de comando ou uma variável de ambiente prevalece sobre o arquivo. As configurações do local (portas, HTTPS, senhas, `access`, `data_dir`, `audience.max_clients`, `mdns`, `network`, `standby`, `log`) existem somente aqui.

## Um exemplo completo

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
