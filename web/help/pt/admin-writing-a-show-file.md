# Escrever um arquivo do espetáculo

Um arquivo do espetáculo é um arquivo de texto simples (`.txt` ou `.md`) escrito numa forma simples de Markdown. Se você colar um texto de teatro comum, ele já funciona: **cada linha é uma legenda**, e um toque na barra de espaço envia uma linha. As linhas em branco são ignoradas, então use-as para manter o arquivo legível.

As linhas longas quebram sozinhas na tela. Você decide onde dividir uma fala longa em linhas, de acordo com o estilo da apresentação. O programa nunca divide nem junta linhas por você.

Para verificar um arquivo do espetáculo, ou prepará-lo no seu próprio editor de texto, veja [Preparar um arquivo do espetáculo](admin-preparing-a-show-file.md).

## Formatação

| Escreva | Você obtém |
|---------|------------|
| `**forte**` | texto em **negrito** |
| `*aparte*` ou `_aparte_` | texto em *itálico* |
| `***very loud***` | texto em negrito e itálico |
| `~~old line~~` | texto riscado (uma correção, ou uma piada riscada) |
| `(thunder rumbles)` | uma linha inteira entre parênteses é mostrada como um som ou uma indicação de cena |
| `HAMLET: To be, or not to be` | um rótulo de personagem acima da linha (o nome deve estar em MAIÚSCULAS seguido de dois-pontos; um grupo também serve: `ALICE & BOB:`, `ALL (EXCEPT BOB):`) |

**Riscar texto.** Coloque dois til de cada lado, `~~assim~~`, e as palavras são mostradas riscadas. Logo dentro de cada par deve haver um caractere que não seja espaço (`~~ x ~~` fica como está). Uma linha cantada começa com um único `~` seguido de espaço, então as duas coisas nunca se confundem; para começar uma linha comum com til, escreva `\~`.

Todo o resto é mostrado como foi digitado. Links, listas, tabelas e títulos não são suportados. HTML puro também não: se você digitar `<b>`, ele aparece literalmente, e a página mostra uma nota sugerindo a forma Markdown. Sublinhado, sobrescrito e subscrito não têm marcação, mas qualquer caractere que você consiga digitar funciona como está: escreva H₂O, x², 1ˢᵗ ou ½ diretamente.

```showfile
HAMLET: To be, or not to be, **that** is the question.
(A bell tolls in the distance.)
OPHELIA: My *lord*?
```

## Conferir os nomes dos personagens

Clique em **Personagens…** acima do editor para ver todos os nomes usados como personagem, os mais usados primeiro, com o número de linhas de cada um. É a forma rápida de achar um nome escrito errado: **ALCE** com uma linha ao lado de **ALICE** com quarenta é marcado com *Você quis dizer ALICE?*. Clique em um nome para ir à primeira linha dele. Um grupo como `ALICE & BOB:` conta para cada um de seus nomes.

## Voltar a uma versão anterior

A cada salvamento, a versão substituída é guardada (as últimas 50 de cada idioma). Clique em **Versões anteriores…** acima do editor, escolha um horário e clique em **Carregar**. Isso só coloca o texto antigo no editor; nada muda até você clicar em **Salvar**, então você pode olhar antes. **Comparar** mostra, linha a linha, o que mudaria ao carregar uma versão. Uma versão a que você dá um nome (na caixa ao lado, ou com **Guardar** para dar nome ao que está salvo agora) nunca é esquecida, por mais salvamentos que se sigam. A exportação não inclui as versões anteriores.

## Configurações do espetáculo: o título e a aparência

Tudo sobre um espetáculo além das imagens fica guardado no arquivo do espetáculo, incluindo o seu **título** e a **aparência** das legendas. Isso vai num bloco `[settings]`, normalmente no início, com um `nome: valor` por linha:

```showfile
[settings]
title: The Garden Party
background: #000000
text color: #ffffff
speaker name color: #ffffff
font: dyslexic
text size: 120%
speaker names: color
idle message: Captions will begin shortly
loop seconds: 10
scrollback lines: 200
[/settings]
```

| Configuração | O que faz |
|--------------|-----------|
| `title` | O nome do espetáculo, como listado na Configuração |
| `background` | A cor da tela atrás das legendas, como `#000000` (um nome como `black` ou `yellow` também funciona) |
| `text color` | A cor das legendas, como `#ffffff` |
| `speaker name color` | A cor dos nomes dos personagens (quando cada um não tem uma cor própria) |
| `font` | `system`, `serif`, `mono`, `dyslexic` ou `hyperlegible` |
| `text size` | Uma porcentagem de 50 a 300; 100 é o normal |
| `speaker names` | `label` (uma cor), `color` (uma cor diferente para cada personagem) ou `hidden` |
| `speaker position` | `above` (o nome acima da linha), `beside` (na mesma linha, antes do texto) ou `column` (em uma coluna à esquerda do texto) |
| `idle message` | As palavras mostradas antes de a primeira linha ser enviada |
| `loop seconds` | Quanto tempo cada item de um `[loop]` fica quando não diz nada |
| `scrollback lines` | Quantas legendas um espectador pode rolar para trás (de 20 a 2000) |
| `archived` | `yes` se o espetáculo está arquivado; os botões **Arquivar** e **Restaurar** da Configuração o acrescentam e o removem |

**Você não precisa escrever nada disso.** A etapa **Aparência** da Configuração preenche o bloco para você (e **Renomear** define o título), e tudo o que você deixar de fora simplesmente usa um valor padrão. Os padrões vêm em duas camadas: o programa tem valores embutidos, que o arquivo de configurações `config.yaml` pode mudar para todos os espetáculos (a seção `defaults`), e então cada arquivo do espetáculo substitui o que define. O mais específico vence: **arquivo do espetáculo, depois config.yaml, depois os valores embutidos**.

Algumas regras: o arquivo do espetáculo de uma tradução só pode definir `idle message` (a aparência é definida uma vez, para o espetáculo todo, no arquivo principal do espetáculo); uma configuração escrita errada ou com um valor sem sentido é apontada com o número da linha e ignorada; e os comentários dentro do bloco não são mantidos quando a Configuração o reescreve, então coloque os comentários acima. Vários blocos `[settings]` são combinados, e os valores posteriores vencem.

## Músicas e trilha

Para um musical, mostre as letras entre notas musicais (♪). Há três maneiras, da mais rápida à mais flexível:

- **Digite ou cole as notas você mesmo.** ♪ e ♫ funcionam em qualquer lugar do arquivo do espetáculo. No editor, os botões **♪** e **♫** as inserem no cursor.
- **Comece uma linha com `~`.** A linha é tratada como uma letra cantada: é mostrada em itálico com um ♪ em cada ponta, e o próprio `~` não é mostrado. Só o `~` inicial é necessário: o ♪ do fim é acrescentado para você, e um `~` de fechamento é permitido, mas não muda nada. Também funciona depois do nome de um personagem.
- **Envolva uma música inteira em `[song]` e `[/song]`.** Toda linha dentro é cantada, então você não precisa de um `~` em cada uma. O título depois dos dois-pontos é um marcador só para o operador. Notas, clear e blank ainda podem ficar dentro.

```showfile
OLIVER: ~ Where is love? ~
~ Row, row, row your boat

[song: The Tea Song]
MARGARET: Oh, the kettle sings its song,
EDWIN: Whistling all day long
[note: hold the last chord]
[/song]
```

### Duas ou mais vozes ao mesmo tempo (duetos e contraponto)

Quando cantores cantam linhas diferentes **ao mesmo tempo**, ponha-as em `[together]` … `[/together]`. Cada linha dentro é mostrada de uma vez, lado a lado numa tela larga e empilhadas num telefone, e o operador as envia com um único toque em Espaço. Cada uma mantém o próprio nome de personagem. Para uma troca curta você pode escrever as vozes numa linha, separadas por ` || ` (um espaço, duas barras verticais, um espaço). `[duet]` e `[counterpoint]` significam o mesmo que `[together]`.

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]

OLIVER: ~ Where is love? ~ || NANCY: ~ Who will buy? ~
```

As vozes que entram **uma depois da outra** são linhas separadas: aperte Espaço para cada uma. Use `[together]` só para linhas que aparecem no mesmo instante. Um bloco `[together]` contém só linhas de texto, precisa de pelo menos duas, e dentro de um `[song]` toda voz é cantada.

Um til no meio de uma linha, como `about ~5 minutes`, é só texto. Para começar uma linha comum com um, escreva `\~`. Veja [Musicais](admin-musicals.md) para mais conselhos.

## Várias linhas em uma legenda

Termine uma linha com uma barra invertida para juntá-la à seguinte, ou ponha as linhas entre `[[` e `]]`.

```showfile
GHOST: Remember me. \
Remember me.

[[
Row, row, row your boat,
Gently down the stream
]]
```

## Comentários e notas

- `<!-- assim -->` é um **comentário**. Ninguém o vê, e ele pode ocupar várias linhas.
- `[note: texto]` é um lembrete mostrado **só ao operador**, por exemplo `[note: hold for applause]`.
- `[section: Act 2, Scene 1]` acrescenta um marcador na lista do operador. Notas e seções nunca gastam um toque na barra de espaço.
- `[pace: 8s]` diz quanto tempo a linha seguinte costuma ficar na tela, como lembrete **só para o operador**. A página de controle mostra isso ao lado dessa linha (`⏱ about 8 s`) e, depois que a linha é enviada, faz a contagem regressiva até o momento de o próximo toque ser devido, ficando amarela quando é. Nunca envia nada sozinha. O tempo é maior que 0 segundo. Pertence à próxima linha que o operador envia (seções e notas são puladas) e não pode ficar dentro de um laço nem de um bloco `[together]`. Se o tempo for muito maior do que o necessário para ler a linha, a lista de problemas acrescenta uma nota, caso seja um erro de digitação.
- Uma linha que começa com `#` e um espaço é um **título** e funciona como `[section: …]`: `# Act 2` (ou `## Scene 1`, com até seis `#`) acrescenta um marcador na lista do operador e nunca é mostrada ao público. Um `#` sem espaço depois (uma hashtag, por exemplo) é ignorado, com uma nota; para começar uma legenda com `#`, escreva `\#`. Os títulos não podem ficar dentro de um laço.

```showfile
<!-- Act 1 captions, version 3 -->
[section: Act 1, Scene 1]
[note: lights come up slowly, wait for the first line]
HAMLET: Who's there?
```

## Comandos

Uma linha entre colchetes é um comando. Os nomes não diferenciam maiúsculas de minúsculas.

| Comando | O que faz |
|---------|-----------|
| `[clear]` | Apaga as legendas da tela |
| `[blank]` | Deixa a tela em branco (as legendas são mantidas) |
| `[image: file.png]` | Mostra uma imagem em tela cheia. Acrescente uma legenda com `[image: file.png \| text]` |
| `![caption](file.png)` | O mesmo, escrito à maneira do Markdown |
| `[fullscreen: text]` | Mostra um texto grande em tela cheia. Para várias linhas use `[fullscreen]` … `[/fullscreen]` |
| `[loop]` … `[/loop]` | Um conjunto repetido de itens em tela cheia para o pré-espetáculo ou o intervalo. Veja [Laços](admin-loops.md) |
| `[together]` … `[/together]` | Duas ou mais vozes mostradas ao mesmo tempo. Veja [Músicas e trilha](admin-musicals.md) |
| `[song: title]` … `[/song]` | Uma música: toda linha dentro é mostrada como letra cantada. Veja [Músicas e trilha](admin-musicals.md) |
| `[section: name]` | Um marcador para o operador |
| `[note: text]` | Uma nota para o operador |
| `[pace: 8s]` | Um lembrete de ritmo para o operador (veja acima) |
| `[language: es]` | Marca um arquivo do espetáculo traduzido. Veja [Idiomas](admin-languages.md) |
| `[settings]` … `[/settings]` | O título e a aparência do espetáculo. Veja *Configurações do espetáculo* acima |

As opções de tempo só fazem sentido dentro de um [laço](admin-loops.md): `| 15s` depois de uma imagem ou de um texto em tela cheia, e `[blank: 3s]`. Fora de um laço, uma imagem ou um texto em tela cheia fica até o operador passar adiante, o tempo é ignorado e apontado como aviso, assim como um tempo em `[blank]`.

```showfile
[clear]
(Lights up on the garden.)
[image: garden.png | The garden at dusk]
[fullscreen: **Please remain seated**]
[fullscreen]
**THE END**
Thank you for coming
[/fullscreen]
[blank]
```

Uma imagem, um texto em tela cheia ou um branco permanece nas telas do público até o operador enviar a linha seguinte.

## Caracteres especiais

Ponha uma barra invertida antes de um caractere para mostrá-lo literalmente em vez de tratá-lo como formatação: `\*`, `\_`, `\#`, `\[`, `\]`, `\:`, `\(`, `\)`, `\!`, `\~` e `\\` (a própria barra invertida). Por exemplo, uma legenda que precisa começar com um jogo da velha se escreve `\# 1 hit`, e `HAMLET\: not a speaker` mostra os dois-pontos sem criar um rótulo de personagem.

```showfile
\# 1 hit of the season
TIME\: ten o'clock
2 \* 3 = 6
```

## O que significam os avisos

A etapa Arquivo do espetáculo lista os problemas com números de linha. Clique em um para ir à linha.

- **«Comando desconhecido [x]»** — uma linha entre colchetes que não é um comando. Se era para ser uma legenda, como um efeito sonoro, escreva-a com parênteses, como `(laughter)`.
- **«Imagem … não encontrada»** — a imagem não foi enviada, ou o nome está escrito de outro jeito. Envie-a na etapa Imagens.
- **«[loop] foi aberto mas nunca fechado»** — acrescente `[/loop]` depois do último item. O mesmo vale para os blocos `[fullscreen]`, os blocos `[[` e os comentários.
- **«HTML não é suportado»** — use `**negrito**` e `*itálico*` em vez de tags.
- **«não pode ficar dentro de um laço»** — os laços contêm só imagens e texto.
- **«[together] precisa de pelo menos duas linhas»** — um bloco `[together]` com uma linha é mostrado como uma legenda comum. **«[together] foi aberto mas nunca fechado»** — acrescente `[/together]`; as linhas que você escreveu são mantidas.
- **«Esta legenda mostra 5 vozes ao mesmo tempo»** — cinco ou mais vozes em um bloco `[together]` (ou numa linha com ` || `) são muitas para um telefone, que as empilha, e para uma tela compartilhada, que tem espaço para cerca de quatro e nunca rola. Divida-a em duas legendas.
- **«Configuração desconhecida» / «não é uma fonte» / «não é uma cor»** — uma linha do bloco `[settings]` que a Configuração não consegue usar; ela é ignorada e vale o padrão. A mensagem diz o que escrever no lugar.
- **«[settings] foi aberto mas nunca fechado»** — acrescente `[/settings]`; tudo o que vem depois de um bloco não fechado é lido como configurações.
- **«[song] foi aberto mas nunca fechado»** — acrescente `[/song]` depois da última letra. As letras que você escreveu continuam sendo mantidas.

Os erros nunca impedem você de salvar.

## Se duas pessoas editarem o mesmo arquivo do espetáculo

Se outra pessoa salvar o arquivo do espetáculo enquanto você o tem aberto, o programa não sobrescreve em silêncio o trabalho dela. Quando você salva, ele avisa que o arquivo do espetáculo foi alterado por outra pessoa e deixa você **usar a versão dela** (substituindo o que está no seu editor) ou **salvar a sua mesmo assim**. Copie o que quiser guardar antes de escolher. A parte com o erro é ignorada, e o resto do arquivo do espetáculo continua funcionando.
