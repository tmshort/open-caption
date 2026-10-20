# Configurar um certificado HTTPS confiável (opcional)

Este guia é para a pessoa que cuida do computador das legendas. Ele é **opcional**: as legendas funcionam sem ele. (As configurações que ele usa estão na [referência de configuração](../../configuration-reference.md).)

## Você precisa disto?

Provavelmente não. Os telefones do público abrem `http://…` sem nenhum aviso, e esse é o endereço que o código QR usa. Um endereço seguro `https://` só é necessário para alguns recursos do navegador limitados a páginas seguras (como manter a tela do telefone acesa do jeito certo; a página do público tem uma alternativa), ou se o local quiser.

Sem este guia, o HTTPS ainda funciona com um certificado **autoassinado**, mas cada telefone mostra um aviso de segurança. Com ele, o programa obtém um certificado real e gratuito do Let's Encrypt e **o renova sozinho**, de modo que `https://captions.mytheater.org` abre sem aviso em todos os dispositivos.

## O que você precisa antes de começar

| Você precisa de | Para quê |
|---|---|
| Um **nome de domínio seu**, por exemplo `mytheater.org`, e um nome dentro dele para as legendas, como `captions.mytheater.org` | Um certificado é emitido para um nome público. `open-caption.local` e números soltos não podem obtê-lo |
| O **DNS dele hospedado em um provedor compatível com este programa** | O programa prova que o nome é seu adicionando um registro temporário no seu DNS. O programa padrão inclui **Cloudflare, GoDaddy e EasyDNS**; execute `open-caption --version` para ver o que o seu tem. Outro provedor precisa de outra compilação (veja o final) |
| **Credenciais de API** dessa conta de DNS | Elas permitem ao programa adicionar e remover esse único registro |
| Que o computador tenha **acesso à Internet** (somente saída) | Ele conversa com o Let's Encrypt e com o seu provedor de DNS. Nada precisa chegar ao local *vindo de fora* |
| Um **endereço fixo** para o computador na rede do local (uma reserva de DHCP serve) | O nome público apontará para ele |

## Passos

**1. Obtenha as credenciais de API.** Na conta do seu provedor de DNS:

- **EasyDNS:** ative o acesso à API e crie um **token e uma chave** de API. As configurações são `EASYDNS_TOKEN` e `EASYDNS_KEY`. (O EasyDNS tem um ambiente de testes para a API; veja a documentação deles para os passos atuais.)
- **Cloudflare:** crie um **token de API** limitado à sua zona com a permissão *Zone → DNS → Edit*. A configuração é `CLOUDFLARE_DNS_API_TOKEN`.
- **GoDaddy:** crie uma **chave e um segredo** de API de produção. As configurações são `GODADDY_API_KEY` e `GODADDY_API_SECRET`. (A GoDaddy pode limitar sua API de DNS a alguns tipos de conta; confira as regras atuais.)
- Qualquer outro provedor: os nomes exatos das configurações estão na [documentação do lego](https://go-acme.github.io/lego/dns/).

Um token de DNS geralmente pode alterar **todos** os registros da zona, não só o temporário, então dê a ele o mínimo de poder que o provedor permitir. Para a melhor proteção, use um CNAME (veja «Credenciais mais seguras» abaixo).

**2. Faça o nome apontar para o computador.** No seu provedor de DNS, adicione um registro `A`: `captions.mytheater.org` → o endereço do computador na rede do local (por exemplo `192.168.1.20`). Não há problema em um nome público apontar para um endereço privado; o certificado não se importa.

**3. Guarde as credenciais em um arquivo privado.** Crie `/etc/open-caption/dns.env` (qualquer lugar privado serve; no Windows, uma pasta que só a sua conta possa ler) com uma linha por configuração, por exemplo para o EasyDNS:

```
EASYDNS_TOKEN=paste-the-token-here
EASYDNS_KEY=paste-the-key-here
```

No macOS e no Linux, torne-o privado: `chmod 600 /etc/open-caption/dns.env`. O programa avisa se outros usuários puderem lê-lo.

**4. Edite `config.yaml`** (as primeiras linhas que o programa imprime dizem onde ele está; cada configuração é explicada na [referência](../../configuration-reference.md)):

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

**5. Pratique primeiro com o serviço de testes.** `letsencrypt-staging` é o serviço de testes do Let's Encrypt: funciona exatamente como o real, mas seus certificados **não são confiáveis para os navegadores** (um aviso é esperado) e ele tem limites generosos. Inicie o programa.

**6. Confira se funcionou.** Abra **Configuração → Configurações → Segurança → Certificado**. O estado passa de *Aguardando o primeiro certificado* para **Válido** em poucos minutos, e mostra o nome, quem o emitiu e a data de validade. O registro do programa diz `HTTPS: got a certificate for … valid until …`. Se disser *A última tentativa falhou*, leia **Último problema** e a tabela abaixo.

**7. Passe para o certificado real.** Troque `directory: letsencrypt-staging` por `directory: letsencrypt` (ou apague a linha) e reinicie o programa. O certificado real chega do mesmo jeito. Abra `https://captions.mytheater.org/` em um telefone no WiFi do local: sem aviso. (Se a porta 443 estava ocupada, o programa usou a reserva, `:8443`; o endereço então precisa da porta, `https://captions.mytheater.org:8443/`, e o registro mostra o endereço exato.)

**8. Dê o endereço às pessoas.** A etapa Compartilhar da configuração lista os endereços numéricos e o nome `.local`, mas ainda não este. Para imprimir um cartaz com código QR dele, com a sessão iniciada na configuração, abra
`http://localhost/admin?print=qr&url=https%3A%2F%2Fcaptions.mytheater.org%2F` (troque o nome; acrescente a porta se necessário) e imprima essa página.

## Depois

- **A renovação é automática.** O programa verifica cerca de duas vezes por dia e renova quando faltam 30 dias ou menos. Você não precisa fazer nada.
- **Se a Internet cair no local,** nada quebra: o programa continua servindo o certificado que tem e tenta de novo uma hora depois. Antes de o primeiríssimo certificado chegar (ou se ele nunca chegar), os telefones recebem o autoassinado.
- **Renovar agora** (Configuração → Configurações → Segurança → Certificado) obtém um novo certificado imediatamente. Use depois de resolver um problema.
- **Trocar o token de DNS:** edite `/etc/open-caption/dns.env`. Ele é lido de novo na próxima tentativa, então não é preciso reiniciar.
- **Desligar:** defina `acme: enabled: no` e reinicie. O HTTPS passa a usar o certificado autoassinado (ou os seus próprios arquivos).

## Se algo der errado

A configuração mostra **Último problema**; o registro tem o mesmo texto. Os mais comuns:

| Você vê | Causa provável e solução |
|---|---|
| `the DNS provider "…" is not built into this program` (o programa para ao iniciar) | Esta cópia do programa foi compilada sem esse provedor. Use `cloudflare`, `godaddy` ou `easydns`, ou consiga uma compilação com o seu (abaixo) |
| `… credentials information are missing: EASYDNS_TOKEN,EASYDNS_KEY` | Os nomes no arquivo estão errados ou o arquivo não pode ser lido. Os nomes são em maiúsculas, `NOME=valor`, um por linha |
| Um erro do provedor de DNS como *unauthorized*, *forbidden* ou um número 4xx | O token ou a chave estão errados, expiraram ou não têm permissão para editar essa zona. Crie novos |
| *Aguardando o primeiro certificado* por muito tempo, ou uma mensagem de que o registro não foi encontrado | O novo registro DNS demora a se propagar. Acrescente `propagation_wait: 3m` em `dns:` e aperte **Renovar agora** |
| `too many certificates already issued` ou `rate limit` | O Let's Encrypt limita repetições. Espere (a mensagem diz quanto) e use `letsencrypt-staging` enquanto testa |
| O navegador ainda avisa | Você está usando o serviço de testes (mude para `letsencrypt`), ou abriu o endereço numérico ou outro nome em vez de `https://captions…`, ou o relógio do computador ou do telefone está errado |
| Um telefone não consegue abrir o nome | Algumas redes WiFi de visitantes e roteadores bloqueiam nomes públicos que respondem com endereços privados («proteção contra DNS rebinding»). Peça ao responsável pela rede que permita `captions.mytheater.org`, ou use o endereço numérico e o código QR como antes |
| `https.acme: cannot be used together with cert_file/key_file` | Você definiu os dois. Remova um: certificados automáticos **ou** os seus próprios arquivos |
| O programa se recusa a iniciar e lista problemas em `config.yaml` | Cada linha nomeia a configuração e o que está errado. Corrija todos e inicie de novo |

## Credenciais mais seguras (opcional)

Um token de DNS que pode alterar toda a sua zona é um risco se o computador for perdido ou compartilhado. Duas formas de reduzi-lo:

- Mantenha os segredos **fora do `config.yaml`**: no arquivo de credenciais (passo 3), ou como variáveis de ambiente do serviço que inicia o programa (`EASYDNS_TOKEN_FILE=/path/to/file` também funciona: a configuração é lida desse arquivo).
- Use a **delegação por CNAME**: no seu DNS, adicione um registro permanente `_acme-challenge.captions.mytheater.org  CNAME  _acme-challenge.other-zone.example`, onde `other-zone.example` é uma zona pequena e, fora isso, vazia. O programa o segue automaticamente, então o token só precisa de acesso a essa zona pequena, não à principal.

As credenciais nunca são escritas no registro, no pacote de suporte nem na configuração.

## Outro provedor de DNS

O programa só contém os provedores de DNS com que foi compilado. Quem o compila pode escolher outros, ou todos, com `make build ACME_PROVIDERS="…"` (os nomes são os da [lista do lego](https://go-acme.github.io/lego/dns/)). Compilar exige o código-fonte e uma conexão com a Internet (os guias de compilação para macOS, Windows e Linux estão no repositório do código; no Windows use `build.bat -AcmeProviders "…"`). O programa cresce a cada provedor: cerca de 10,9 MB sem certificados automáticos, 12,6 MB com os três padrão (Cloudflare, GoDaddy, EasyDNS), 15,0 MB com EasyDNS e Route 53 e 60,9 MB com todos os provedores. Se o seu provedor não for compatível de forma alguma, `exec` (executa um script seu) e `httpreq` (chama um endereço web) podem controlar quase qualquer serviço de DNS.
