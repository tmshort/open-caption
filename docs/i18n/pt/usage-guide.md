# Guia de uso do open-caption

O open-caption mostra legendas ao vivo nos telefones, tablets e telas do público durante uma peça ou um musical. Uma pessoa voluntária, o **operador**, acompanha o espetáculo e aperta a barra de espaço a cada linha; todos os dispositivos no WiFi do local mostram a linha um instante depois. Nada é instalado nos dispositivos do público, e ninguém precisa tocar em nada.

Este guia trata do que comprar e de como organizar ([Instalação](#recomendacoes-de-instalacao)), de como preparar um espetáculo, de como conduzi-lo e traz uma lista de verificação para o dia. O botão **Ajuda** do programa cobre o mesmo assunto em pedaços menores, e [venue-network-guide.md](venue-network-guide.md) é escrito para quem administra o WiFi do prédio.

Conteúdo: [As peças](#as-pecas) · [Recomendações de instalação](#recomendacoes-de-instalacao) · [Instalação e início](#instalacao-e-inicio) · [Preparar um espetáculo](#preparar-um-espetaculo) · [Conduzir o espetáculo](#conduzir-o-espetaculo) · [Lista de verificação do dia](#lista-de-verificacao-do-dia) · [Solução de problemas](#solucao-de-problemas) · [Manutenção](#manutencao) · [Cartão rápido do operador](#cartao-rapido-do-operador)

## As peças

| Peça | O que é |
|------|---------|
| **O servidor** | O programa `open-caption`, rodando em um computador (macOS, Windows ou Linux). Guarda o arquivo do espetáculo e envia as linhas a todos. |
| **Página de Configuração** (`/admin`) | Onde os voluntários preparam um espetáculo: arquivo do espetáculo, imagens, aparência, ensaio. Senha de administrador. |
| **Página do operador** (`/operator`) | Onde o operador acompanha o espetáculo e envia as linhas. Senha de operador. |
| **Página do público** (`/`) | O que o público abre nos seus dispositivos. Sem senha. Um navegador num projetor ou monitor grande pode mostrar a mesma página para a sala toda. |

Tudo roda em um computador. O público, o operador e os voluntários da configuração usam navegadores web comuns.

## Recomendações de instalação

### A versão curta

1. **Use um ponto de acesso WiFi dedicado às legendas**, a menos que a rede de visitantes do local seja reconhecidamente confiável (veja abaixo). Isso elimina o maior risco: uma rede que você não controla.
2. **Ponha o servidor numa conexão por cabo** com esse ponto de acesso, com um endereço reservado.
3. **Use um ponto de acesso feito para multidões** (dimensionado para mais de 100 dispositivos) colocado na área do público.
4. **Mantenha o computador servidor ligado na tomada, acordado e sem outras tarefas** durante o espetáculo.
5. **Imprima o código QR** e coloque-o no programa e no saguão.
6. **Ensaie com alguns telefones de verdade** na sala de verdade antes da estreia.

### Opções de rede

| Opção | Quando usar | Observações |
|-------|-------------|-------------|
| **A. Ponto de acesso dedicado (recomendado)** | Qualquer local, principalmente se você não pode mudar a rede do prédio | Você traz um roteador ou ponto de acesso. O público entra no WiFi dele (por exemplo «Theater Captions»). Não é preciso conexão com a internet. Você controla a capacidade, as reservas e os firewalls. |
| **B. O WiFi de visitantes do local** | A equipe de rede do local confirma que os dispositivos de visitantes alcançam o seu computador | O mais barato. Peça um endereço reservado para o seu computador e que o tráfego dos visitantes para o servidor seja permitido. Veja [venue-network-guide.md](venue-network-guide.md). |
| **C. O hotspot do seu próprio notebook** | Só ensaios e públicos minúsculos | Hotspots de notebook costumam aguentar só um punhado de dispositivos. Não serve para uma sala cheia. |

### A opção A em detalhe (ponto de acesso dedicado)

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

- **Equipamento:** um ponto de acesso ou roteador profissional ou semiprofissional projetado para muitos dispositivos simultâneos (procure mais de 100 clientes nas especificações e suporte a 5 GHz), com uma porta Ethernet. Roteadores domésticos baratos costumam penar acima de 30–50 dispositivos. Para uma casa de cerca de 150 lugares, conte com um ponto de acesso capaz, e um segundo para uma sala funda ou larga.
- **Configurações:**
  - Crie um SSID com um nome curto e uma senha WPA2 simples e imprimível, ou sem senha se o local permitir uma rede aberta. Imprima-a no cartaz QR.
  - Ligue o **5 GHz** e prefira-o. Use larguras de canal de 20 ou 40 MHz num prédio congestionado.
  - Deixe o **isolamento de clientes** ligado se for oferecido. Os dispositivos do público falam só com o servidor, nunca entre si.
  - Dê ao computador das legendas uma **reserva DHCP** (ou um endereço fixo) para que o endereço nunca mude, e imprima o endereço no cartaz.
  - Você não precisa de conexão com a internet, e o programa nunca usa uma.
- **Posição:** ponha o ponto de acesso na área do público, no alto e longe de paredes, não num fundo de sala nem no saguão. Não o coloque atrás do palco nem do rack de iluminação. Um notebook numa mesa distante serve: ele está ligado por cabo.
- **Energia:** use um protetor contra surtos. Um pequeno nobreak no ponto de acesso e no servidor elimina o risco de um plugue desconectado.

### A opção B em detalhe (o WiFi de visitantes do local)

O programa não se importa com a rede em que está, desde que os dispositivos do público consigam alcançá-lo. Antes de contar com ela:

- Peça **acesso dos visitantes ao servidor**: o seu computador deve ser alcançável a partir do SSID de visitantes, na porta TCP 80 (e 443 se você ativar o HTTPS). O «isolamento de clientes» entre visitantes é aceitável; impedir que os visitantes alcancem *o seu computador* não é.
- Peça um **endereço reservado** para o adaptador de rede do seu computador.
- Ligue o computador à rede por **Ethernet** se for de algum modo possível.
- Teste com um telefone no WiFi de visitantes bem antes do espetáculo (a etapa Conduzir mostra o endereço).
- Tenha a opção A numa bolsa como reserva se o espetáculo for importante.

### O computador das legendas

- **Qualquer notebook ou mini PC comum** serve. O programa usa muito pouca memória (algumas dezenas de megabytes) e quase nenhum processador. Use um computador que você possa manter ligado na tomada.
- **Sistema operacional:** macOS, Windows ou Linux. Use o mesmo computador para todos os espetáculos, se puder, para que as senhas, os espetáculos e o endereço continuem os mesmos.
- **Deixe-o entediante:** configure o computador para *nunca suspender* quando ligado na tomada, desative as reinicializações automáticas para atualizações perto das datas de espetáculo, feche outros programas e mantenha a tampa aberta ou configurada para não suspender.
- **Firewall:** permita o programa em **redes privadas** (o Windows pergunta na primeira vez; no macOS, permita as conexões de entrada para ele).
- **Endereço:** por cabo e reservado (veja acima). Se o endereço mudar, os códigos QR impressos deixam de funcionar.
- **Hora:** o relógio não importa.

### Quem fica onde

- **O operador** precisa de uma boa visão do palco e de um dispositivo confiável. O arranjo mais simples é o próprio computador das legendas na mesa técnica. A página do operador foi pensada para um computador com teclado (Espaço envia a próxima linha). Use o computador das legendas ou outro notebook da mesma rede (`http://<endereço do computador>/operator`), para o operador sentar onde veja o palco. Um telefone ou tablet mostra a página, mas **não é recomendado**: numa tela sensível ao toque é fácil errar o toque durante o espetáculo. Um controle Bluetooth ou um teclado é uma boa alternativa.
- **Só um operador controla o espetáculo por vez.** Qualquer outra pessoa que abrir a página do operador pode assistir, e pode usar **Assumir** se o operador passar a vez.
- **Os voluntários da configuração** usam a página de Configuração de qualquer computador da rede antes do espetáculo, ou do computador das legendas.
- **Uma tela compartilhada** para a sala toda: abra o endereço do público num computador ligado a um projetor ou monitor grande, acrescente `?display=projector` ao endereço para ter texto grande e nada de poluição na tela, e ponha o navegador em tela cheia (`F11`, ou rode o Chrome com `--kiosk`).

### Executar o servidor

Para a maioria dos teatros, o jeito mais simples é o melhor: inicie o programa à mão antes de cada espetáculo, deixe a janela dele aberta e aperte `Ctrl+C` (ou feche-a) depois. A posição do espetáculo é salva continuamente, então reiniciar no meio de um espetáculo é seguro (veja [Solução de problemas](#solucao-de-problemas)).

Se o computador deve iniciar o servidor sozinho depois de uma reinicialização, a pasta [`deploy/`](../deploy/) traz arquivos prontos para Linux (systemd), macOS (launchd) e Windows (Agendador de Tarefas). Cada um começa com seus próprios passos de instalação.

### Levar as pessoas à página

Em ordem de confiabilidade:

1. **O código QR e o cartaz** (Configuração → Conduzir → *Imprimir um cartaz com código QR para o saguão*). Ponha-o na parede, no programa e no cartão do encosto das poltronas. É nisso que se deve confiar.
2. **Um endereço curto.** Na porta 80 não há `:porta` para digitar. O programa também se anuncia como **`open-caption.local`** (mude em Configuração → Conduzir → *Usar um nome*): onde a rede de visitantes permite, `http://open-caption.local/` funciona na maioria dos telefones sem configurar nada. Algumas redes de visitantes bloqueiam esse tipo de anúncio, então não imprima só o nome; o cartaz QR usa o endereço numérico. Um nome no DNS do próprio roteador (como `captions.theater`) funciona em qualquer lugar, mas exige que o local o acrescente.
3. **A página de boas-vindas do ponto de acesso** (um «portal de visitantes» ou «portal cativo» que redireciona ao seu endereço depois de entrar). Isso se configura no roteador ou ponto de acesso, não neste programa. Veja [venue-network-guide.md](venue-network-guide.md).
4. **Opção 114 do DHCP e `/captive-api` (avançado, experimental).** A opção 114 é uma configuração do DHCP que aponta os dispositivos para uma «API de portal cativo». O programa responde a essa API em `/captive-api` (por exemplo `https://<endereço>/captive-api`) com «esta rede não é cativa» e o endereço da página. Duas ressalvas: **ela não faz os telefones abrirem sozinhos a página de legendas**, e o padrão (RFC 8910/8908) exige **HTTPS com um certificado em que os dispositivos confiem**, então uma configuração com `http://` simples ou o certificado autoassinado do programa é ignorada. Só configure se o local já usa a opção 114 e pode fornecer um certificado confiável. Não foi testada em dispositivos reais.

### Portas e endereços

- Por padrão o servidor escuta na **porta 80**, então o endereço do público não precisa de número de porta, por exemplo `http://192.168.1.10/`. Se outra coisa usa a porta 80, ou o sistema operacional não a permite, o servidor avisa e usa a **porta 8080**; o endereço do público passa então a terminar em `:8080`. No Linux, a porta 80 precisa de permissão (`sudo setcap 'cap_net_bind_service=+ep' ./open-caption`), a menos que você use o arquivo do systemd, que a concede.
- Você pode escolher as portas com `--http-addr :8081`, ou em `config.yaml`.
- **O HTTPS é opcional.** O `http://` simples funciona em todos os dispositivos sem avisos e é o endereço recomendado para o público. Um endereço HTTPS autoassinado faz cada dispositivo mostrar um aviso de segurança. Ative o HTTPS só se puder instalar um certificado em que seus dispositivos confiem. Se o local tem um nome de domínio próprio, o programa pode obter e renovar sozinho um certificado confiável; veja `docs/https-certificate-guide.md` (somente em inglês).

### Segurança

- A página do público só pode ler. Não pode mudar o espetáculo nem alcançar as páginas de configuração e do operador.
- Configuração e operador exigem cada uma a sua senha (guardada em texto puro em `config.yaml`, que é criado legível só por você). Dê a senha de operador ao operador, não a de administrador. As senhas podem ser mudadas em **Configurações** na Configuração.
- No computador que executa o programa, Configuração e a página do operador abrem **sem senha** quando o endereço começa com `localhost`, `127.0.0.1` ou `[::1]` (a pessoa nesse computador é tratada como administrador). Os outros computadores, e qualquer requisição retransmitida por um proxy, continuam entrando com senha. Defina `access.local_auto_login: false` em `config.yaml` se outras pessoas puderem entrar neste computador.
- **Qualquer pessoa que alcance o servidor pode abrir as páginas de entrada da Configuração e do operador, e com `http://` simples as senhas trafegam sem criptografia pelo WiFi.** Num WiFi de visitantes compartilhado isso significa que um espectador curioso poderia capturar uma senha e assumir o controle das legendas. Reduza o risco, nesta ordem:
  1. **Limite quais computadores podem abri-las.** Em `config.yaml`:
     ```yaml
     access:
       admin_allow: [localhost]            # Setup only on the captioning computer itself
       operator_allow: [192.168.1.50]      # the operator's computer (or a network such as 192.168.1.0/24)
     ```
     As entradas podem ser `localhost`, `private` (faixas de endereços domésticos/de escritório), um único endereço, ou uma rede como `192.168.1.0/24`. Uma lista vazia (o padrão) significa qualquer computador. Um endereço em `admin_allow` também pode usar a página do operador. Reinicie o programa depois de editar. Uma entrada errada para o programa com uma mensagem em vez de permitir todos. Configurações na Configuração mostra o que está permitido agora.
  2. **Ponha o operador numa rede por cabo ou separada** à qual os dispositivos do público não possam se juntar, ou dê uma senha WPA2 à rede das legendas.
  3. **Use HTTPS** (`https.enabled` em `config.yaml`) se puder dar aos dispositivos um certificado em que confiem; um autoassinado faz cada dispositivo mostrar um aviso, e um endereço do público sem criptografia é mais simples.
- Um site, ou outro serviço no mesmo computador, não pode fazer um navegador mudar o espetáculo pelas suas costas: as requisições que não vieram das páginas do próprio programa são recusadas.
- O programa foi feito para uma rede privada. **Não o exponha à internet.**

## Instalação e início

1. **Obtenha o programa.** Baixe a versão para o seu computador (ou rode `make dist` para construir todas). Ponha-a numa pasta que você vá manter, por exemplo `Documentos/open-caption`.
2. **Primeiro início.** Execute-o: duplo clique no Windows, ou num terminal `./open-caption` (macOS/Linux; no macOS a primeira execução pode exigir *clique direito → Abrir* porque o programa não é assinado).
3. **Anote as senhas.** Na primeira vez, o programa imprime as senhas de **administrador** e de **operador** e onde está `config.yaml`. Elas ficam nesse arquivo.
4. A página de configuração abre no seu navegador. Se não abrir, abra `http://localhost/admin` (ou `http://localhost:8080/admin` se o console disser que usou a porta 8080).
5. **Escolha onde ficam os dados (opcional).** Por padrão, os espetáculos e as configurações ficam na pasta de configuração do seu usuário (macOS `~/Library/Application Support/open-caption`, Linux `~/.config/open-caption`, Windows `%AppData%\open-caption`). Para manter tudo numa pasta que você escolher, inicie com `--data <pasta>`. Use sempre a mesma pasta.

Opções úteis: `--data <pasta>`, `--http-addr :8080`, `--https-addr :443`, `--no-browser`, `--log-file <arquivo>`, `--log-level debug|info|warn|error`, `--version`.

Se `config.yaml` tiver um erro (uma configuração escrita errada, um número fora do intervalo, um endereço ruim), o programa lista todos os problemas com a linha ou o nome da configuração e para; corrija o arquivo e inicie de novo. (Um arquivo marcado com um `config_version` mais novo do que o programa entende é a exceção: as configurações desconhecidas são ignoradas com um aviso, de modo que você pode voltar a um programa mais antigo.)

## Preparar um espetáculo

Abra a página de configuração e siga as seis etapas no alto. O botão **Ajuda** explica cada uma.

1. **Espetáculos.** Crie um espetáculo (ou o espetáculo de exemplo, para experimentar).
2. **Arquivo do espetáculo.** Cole o arquivo do espetáculo ou solte um arquivo `.txt` ou `.md`. Cada linha vira uma legenda. Corrija o que aparecer em *Problemas e avisos*; clique em um para ir à linha dele. Salve.
3. **Imagens.** Envie as imagens que o arquivo do espetáculo usa (PNG, JPG, GIF ou WebP, de preferência com menos de 1 MB cada).
4. **Aparência.** Escolha cores, fonte e tamanho do texto. Os padrões (branco sobre preto, grande) são os melhores para a maioria das pessoas. Duas fontes de legibilidade vêm embutidas: uma amigável para dislexia (OpenDyslexic) e uma de alta legibilidade (Atkinson Hyperlegible, projetada para baixa visão). Cada espectador também pode escolher qualquer uma para o próprio dispositivo no pequeno controle **Aa**.
5. **Teste.** Faça um ensaio particular: você aperta Espaço como o operador faria, e vê a tela do público ao lado. Nada chega ao público.
6. **Conduzir.** Confira a lista de preparação, imprima o cartaz com o QR e pressione **Abrir a página Conduzir**.

Abrir a página do operador carrega o espetáculo **em espera**: o público vê a tela de espera do local (um logotipo e uma mensagem, definidos uma só vez em **Configurações → Tela de espera**) e nada do espetáculo, então você pode preparar, ensaiar e conferir a página do operador com calma. Quando estiver pronto, pressione **Espaço** (o botão diz **INICIAR**) na página do operador para colocar o espetáculo no ar. **Colocar em espera…** o retira de novo sem perder o lugar.

### Escrever o arquivo do espetáculo

Quem prepara o arquivo do espetáculo não precisa ler o resto deste guia: [Preparar um arquivo do espetáculo](../web/help/admin-preparing-a-show-file.md) cobre o trabalho do texto ao arquivo pronto, incluindo editá-lo no seu próprio programa, e [Primeiros passos](../web/help/admin-getting-started.md) percorre as seis etapas da Configuração.

Um arquivo do espetáculo é texto simples. A versão curta:

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

- **Uma linha = uma legenda** = um toque em Espaço. Você decide onde quebrar as falas longas.
- Os **nomes dos personagens** em maiúsculas seguidos de dois-pontos aparecem acima da linha.
- As **indicações de cena** entre `(parênteses)` aparecem em itálico.
- As **letras de música** entre notas `♪`, em itálico: comece uma linha com `~`, ou envolva a música em `[song]`…`[/song]`. Para dois cantores ao mesmo tempo (um dueto ou contraponto), ponha as linhas deles entre `[together]` e `[/together]`, ou una-as com ` || ` numa linha; as duas aparecem juntas com um toque em Espaço. Os botões **♪**, **♫** e **🎵 Música** do editor as inserem. Veja o tópico de ajuda *Musicais*.
- Os **laços** (pré-espetáculo e intervalo) repetem imagens e texto até o operador seguir adiante. Um laço bem no início do arquivo do espetáculo começa quando o operador coloca o espetáculo no ar (INICIAR).
- As **notas** `[note: …]` e os marcadores `[section: …]` são só para o operador.
- O tópico de ajuda **Escrever um arquivo do espetáculo** lista tudo, com exemplos, e a *cola do arquivo do espetáculo* do editor o resume.

### Configurações do espetáculo (título, aparência e mais)

Tudo sobre um espetáculo além das imagens fica guardado no arquivo do espetáculo, incluindo o título e a aparência das legendas, num bloco `[settings]` perto do início:

```text
[settings]
title: The Garden Party
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
[/settings]
```

A etapa **Aparência** da Configuração escreve esse bloco para você, e você pode editá-lo à mão. Tudo o que um arquivo do espetáculo deixa de fora usa um padrão: o programa tem valores embutidos, `config.yaml` (a seção `defaults`) pode mudá-los para todos os espetáculos deste computador, e cada arquivo do espetáculo substitui os dois. Portas, senhas, HTTPS e as listas de acesso não são configurações do espetáculo e ficam em `config.yaml`. Como o arquivo do espetáculo leva a própria aparência, copiar ou exportar o espetáculo leva a aparência também. Veja o tópico de ajuda *Escrever um arquivo do espetáculo* para todas as configurações.

### Voltar atrás

Cada salvamento guarda a versão que substituiu (as últimas 50). **Versões anteriores…** acima do editor as lista; **Carregar** coloca uma no editor, e **Salvar** a torna atual. **Comparar** mostra linha a linha o que mudaria ao carregar uma, e uma versão a que você dá um nome (ou **Guardar** para o que está salvo agora) nunca é esquecida. Exporte antes do espetáculo de qualquer jeito: as versões anteriores ficam na pasta de dados, não na exportação.

### Duas pessoas editando

Se outra pessoa salvar um arquivo do espetáculo enquanto você o tem aberto, o programa avisa você ao salvar em vez de sobrescrever o trabalho dela, e deixa você ficar com a versão dela ou manter a sua.

### Mais de um idioma

Um espetáculo pode oferecer legendas em vários idiomas. Em Configuração → Arquivo do espetáculo, clique em **＋ Adicionar um idioma**, deixe marcada a caixa «Começar com uma cópia do arquivo principal do espetáculo» e traduza as palavras. Cada idioma é o seu próprio arquivo do espetáculo, com as mesmas linhas, comandos e telas em branco na mesma ordem, então um toque em Espaço mostra o mesmo momento em todos os idiomas. O operador acompanha só o idioma principal. Os dispositivos do público recebem o idioma em que o telefone está configurado (ou o principal) e podem trocar pelo pequeno menu 🌐, ou receber um link terminado em `?lang=es`. Uma tradução que deixa de coincidir com o arquivo principal do espetáculo é apontada e retida até ser corrigida. O espetáculo de exemplo inclui uma tradução para o espanhol para você ver. Veja o tópico de ajuda *Idiomas*.

### Mudar coisas durante os ensaios e espetáculos

Você pode editar e salvar de novo um arquivo do espetáculo com o espetáculo ao vivo (por exemplo para corrigir um erro de digitação). O programa pede que você digite o nome do espetáculo para ter certeza, mantém o lugar do operador pelo número da linha e avisa se o número de linhas mudou.

## Conduzir o espetáculo

### Antes de o público chegar

1. Inicie o servidor e abra a **página do operador**; entre. Ela deve dizer **Você tem o controle**.
2. Em Configuração → Conduzir, pressione **Abrir a página Conduzir**. O espetáculo é carregado em espera: o público vê a tela de espera do local. Quando estiver pronto, pressione **Espaço** (INICIAR) na página do operador para colocá-lo no ar. Um laço de pré-espetáculo, se o arquivo do espetáculo tiver um, começa então em todas as telas.
3. Abra o endereço do público no seu próprio telefone e, se tiver um, no projetor. Confira que a contagem de espectadores na página do operador sobe.
4. Ponha o cartaz QR onde as pessoas possam vê-lo.

### Durante o espetáculo

- **Espaço** envia a próxima linha (também Seta para a direita, Page Down, Enter e a maioria dos controles remotos).
- **Backspace** desfaz a última coisa que você enviou.
- **B** deixa em branco as telas do público (esconde tudo; as mesmas legendas voltam quando você aperta **B** de novo). **C** as limpa (apaga as legendas, para que a próxima linha comece numa tela vazia). **Backspace** desfaz qualquer um dos dois.
- **Shift+Q** para um laço de intervalo ou pré-espetáculo (tela em branco). **Espaço** continua.
- **Aviso…** põe um aviso avulso em todas as telas («O espetáculo recomeça em instantes»). **Voltar** o remove.
- Clique numa linha e confirme para pular para qualquer lugar. **/** busca no arquivo do espetáculo.
- A linha com a seta amarela é a *seguinte*. O ponto verde marca o que está na tela agora.

### Quando o espetáculo termina

Pressione `Ctrl+C` na janela do programa (ou pare o serviço). A posição é salva. Para refazer o espetáculo desde o início, use **Reiniciar espetáculo…** na página do operador, ou **Reiniciar o espetáculo** em Configuração → Conduzir. Ambos o devolvem à espera.

## Lista de verificação do dia

**Alguns dias antes**
- [ ] Ensaie com o arquivo do espetáculo de verdade e alguns telefones de verdade na rede de verdade na sala de verdade.
- [ ] Exporte o espetáculo (Configuração → Espetáculos → **Exportar**) e guarde o `.zip` em local seguro, como um pen drive.
- [ ] Confirme que o endereço do computador das legendas está reservado e que o cartaz QR o mostra.

**No local, uma hora antes**
- [ ] Energia: computador e ponto de acesso ligados na tomada; computador configurado para nunca suspender.
- [ ] Ponto de acesso ligado, no lugar, com o nome certo; o computador ligado a ele por cabo.
- [ ] Inicie o programa; anote o endereço impresso (e o de Conduzir).
- [ ] Um telefone no WiFi do público consegue abrir o endereço e mostra «As legendas começarão em breve».
- [ ] Configuração → Teste: faça um ensaio rápido; as teclas do operador funcionam.
- [ ] Navegador do projetor (se usado) aberto em `…/?display=projector`, em tela cheia.
- [ ] Pressione **INICIAR** na página do operador. O laço de pré-espetáculo (se houver) aparece em todas as telas.
- [ ] O operador sentado onde veja o palco; controle remoto ou teclado testado.
- [ ] Cartaz QR colocado; um voluntário sabe o nome do WiFi e o endereço para quem perguntar.

**Depois do espetáculo**
- [ ] `Ctrl+C` na janela do programa.
- [ ] Exporte o espetáculo de novo se mudou algo.

## Solução de problemas

| Problema | O que tentar |
|----------|--------------|
| Ninguém consegue abrir o endereço | O telefone está no WiFi certo (e não nos dados móveis)? O endereço inclui `:8080` se o programa usou essa porta? O tráfego dos visitantes para o servidor é permitido? O firewall do computador permite o programa em redes privadas? Veja [venue-network-guide.md](venue-network-guide.md). |
| Alguns dispositivos se conectam e outros não | O ponto de acesso pode estar no limite de dispositivos. Acrescente um segundo ponto de acesso, ou leve o primeiro para dentro da sala. |
| As legendas chegam tarde ou travam em alguns telefones | Sinal fraco ou canal ocupado. Esses telefones se reconectam sozinhos. Melhore a posição do ponto de acesso; use 5 GHz. |
| Alguém perdeu uma linha | Pode rolar as legendas para cima para relê-la; as linhas novas esperam atrás de um botão **Voltar ao vivo** até que ele o toque. |
| Uma página diz «Este navegador é antigo demais» | O navegador do dispositivo está abaixo do que a página precisa: cerca de iOS 12 para a página do público, iOS 14.1 / Chrome 87 para a Configuração e a página do operador. Atualize o dispositivo, ou use outro. |
| As telas dos telefones escurecem | Os telefones bloqueiam a tela. A página mantém a tela acesa com o bloqueio de ativação do navegador onde consegue e, onde não consegue (HTTP simples), com um pequeno vídeo invisível em laço. Quem preferir deixar o telefone dormir (bateria fraca, por exemplo) pode desligar *A tela fica acesa* no painel «Aa». Alguns telefones em modo de economia de energia recusam os dois: a página então mostra uma dica para desligar o bloqueio automático durante a noite. |
| A linha errada foi enviada | **Backspace** a desfaz. |
| A página do operador diz que outro operador tem o controle | Use **Assumir**, se você é o operador. |
| O computador travou ou foi reiniciado no meio do espetáculo | Inicie o programa de novo. A página do operador se reconecta, continua conectada e retoma o controle sozinha, o espetáculo está onde você o deixou, e os dispositivos do público se reconectam sozinhos. Um laço recomeça do primeiro item. **Backspace** ainda pode desfazer as últimas coisas enviadas antes da reinicialização. |
| Esqueci uma senha | Entre com a outra e use **Configurações** na Configuração, ou abra `config.yaml` na pasta de dados com qualquer editor de texto. |
| Outra coisa | Configuração → Conduzir → **Baixar um pacote de suporte**; ele contém as configurações (sem senhas) e linhas recentes do registro. |

## Manutenção

- **Cópias de segurança:** *Exportar* em Configuração → Espetáculos salva um espetáculo (arquivo do espetáculo, imagens, aparência) como um `.zip`. *Importar* o restaura, neste ou em outro computador. Você também pode copiar a pasta de dados inteira.
- **Mudar para outro computador:** instale o programa, importe os espetáculos exportados, ajuste o endereço do computador novo e, se quiser manter as senhas antigas, copie `config.yaml`.
- **Atualizar o programa:** pare-o (`Ctrl+C`), substitua o arquivo do programa pelo mais novo e inicie-o de novo com a mesma pasta `--data`. Não atualize logo antes de um espetáculo.
- **Reutilizar um espetáculo:** Configuração → Espetáculos tem **Renomear…** e **Duplicar**. Duplicar faz uma cópia do arquivo do espetáculo, das imagens e da aparência para adaptar à próxima temporada, e deixa o original intacto. Renomear, Duplicar, Exportar, Arquivar e Excluir estão no menu **Mais ▾** de cada espetáculo.
- **Limpar:** exclua os espetáculos de que não precisa mais em Configuração → Espetáculos. Os `.zip` exportados são o seu arquivo.
- **Registros:** Configuração → Conduzir → **Registro recente…** mostra o que o programa registrou há pouco, filtrado por nível (escolha *Somente problemas* para ver avisos e erros); as linhas recentes do registro também estão no pacote de suporte; acrescente `--log-file` para manter um arquivo. Cada linha começa com o seu nível (`INFO`, `WARN`, `ERROR` ou `DEBUG`). O `info` padrão registra entradas, quem assumiu o controle da página do operador, espetáculos ativados e eventos parecidos; `warn` mantém só os problemas; `debug` registra também cada requisição, o que ajuda quando algo estranho acontece. Defina com `log: {level: debug}` em `config.yaml`, com `--log-level debug`, ou com o programa em execução em Configuração → Configurações → *Nível de detalhe do registro*. Se o programa foi iniciado com `--log-level`, essa opção manda a cada início: a Configuração avisa, uma mudança feita lá dura só até a próxima reinicialização, e o registro de início avisa quando a opção e o arquivo discordam. Um arquivo de registro mantido com `--log-file` começa um novo aos 5 MB e guarda três anteriores (`open-caption.log.1` a `.3`), de modo que nunca enche o disco.
- **Arquivar:** em Configuração → Espetáculos, **Arquivar** guarda de lado um espetáculo encerrado (ele é mantido, listado em «Espetáculos arquivados», e não pode ir ao ar) e **Restaurar** o traz de volta. **Excluir** remove um espetáculo de vez.

## Cartão rápido do operador

Imprima e deixe junto do operador.

| Tecla | Faz |
|-------|-----|
| **Espaço** | Enviar a próxima linha |
| **Backspace** | Desfazer a última coisa enviada |
| **Shift+Q** | Parar um laço (tela em branco) |
| **B** | Em branco: esconde tudo, as legendas voltam ao desfazer o branco |
| **C** | Limpar: apaga as legendas para um recomeço |
| **/** | Buscar no arquivo do espetáculo |
| **?** | Ajuda (as teclas ficam em pausa enquanto ela está aberta) |

- Seta amarela = próxima linha. Ponto verde = na tela agora.
- Linha errada? **Backspace.** Perdeu o lugar? Clique na linha que quer e confirme.
- Precisa dizer algo ao público? **Aviso…**
- Espetáculo interrompido? **B** para deixar as telas em branco, **Aviso…** para explicar.
