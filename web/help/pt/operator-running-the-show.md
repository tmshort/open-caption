# Conduzir o espetáculo

A página mostra o arquivo do espetáculo como uma lista. As linhas já enviadas ficam esmaecidas, a linha **na tela agora** tem um ponto verde, e a linha **seguinte** é grande e leva uma seta amarela.

## As três coisas para lembrar

1. **Espaço** envia a próxima linha.
2. **Backspace** desfaz a última coisa que você enviou.
3. **Shift+Q** para um laço de pré-espetáculo ou de intervalo e deixa a tela em branco.

Acompanhe os atores e aperte **Espaço** a cada fala. Você pode usar um controle remoto; os botões de próximo e anterior dele funcionam como Espaço e Backspace.

## Em branco e Limpar

Ambos deixam vazias as telas do público, mas não são a mesma coisa:

| | **Em branco** (`B`) | **Limpar** (`C`) |
|---|---|---|
| O que o público vê | Nada: só o fundo | Uma área de legendas vazia, pronta para novas linhas |
| As linhas anteriores | **Mantidas.** Aperte `B` de novo, ou envie a próxima linha, e elas ainda estão lá em cima | **Apagadas.** A próxima linha começa numa tela vazia |
| Use para | Um apagão, uma pausa, ou esconder as legendas por um momento | Uma nova cena, ou depois de uma música, para que as linhas velhas não se acumulem |
| Desfazer | `B` de novo, ou **Backspace** | **Backspace** |

Um arquivo do espetáculo pode fazer o mesmo com os comandos `[blank]` e `[clear]`; eles aparecem na sua lista como `TELA EM BRANCO` e `LIMPAR TELA`.

## Palavras desta página

- **Em espera**: o espetáculo está carregado, mas o público vê apenas a tela de espera do local.
- **No ar**: o público vê o espetáculo. **INICIAR** (START) o coloca no ar; **Colocar em espera…** o devolve à espera, mantendo a posição.
- **Reiniciar espetáculo…**: volta à primeira linha e à espera.
- **Em branco** oculta as legendas mas mantém as linhas anteriores; **Limpar** as apaga (veja acima).

## Antes do espetáculo

- Use a página do operador num computador com teclado (um notebook serve). Ela foi pensada para a tecla **Espaço**, e um telefone ou tablet não é recomendado: numa tela sensível ao toque é fácil errar o toque durante o espetáculo.
- Entre na página do operador. Você verá «Você tem o controle». Se disser que outro operador tem o controle, use **Assumir** somente se você deve ser o operador.
- Um segundo dispositivo (o tablet do contrarregra) pode assistir pela página de acompanhamento, `/follow`, com a senha de espectador (ou a de operador ou administrador). Ele mostra o que você vê, notas incluídas, mas não pode enviar nada; a etapa **Compartilhar** da configuração mostra o endereço.
- Confira a contagem de espectadores (o ícone de pessoas). Ela mostra quantos dispositivos estão conectados.
- Um espetáculo carregado fica **em espera**: o público vê a tela de espera do local (um logotipo e uma mensagem) e nada do espetáculo, e esta página diz **EM ESPERA**. Pressione **Espaço** (o botão diz **INICIAR**) para colocar o espetáculo no ar. Se o arquivo do espetáculo começar com um laço de pré-espetáculo, ele começa então.

## Durante o espetáculo

- **Ritmo:** quando uma linha tem um `[pace]` no arquivo do espetáculo, o tempo aparece ao lado da próxima linha e uma barra fina, na borda inferior da faixa de estado, se enche enquanto a linha está na tela. Quando a próxima pressão é devida, a próxima linha ganha uma moldura amarela e a barra fica amarela. Nada é enviado sozinho e não há som.
- As linhas marcadas como notas ou seções são para você. Elas nunca gastam um toque em Espaço.
- Imagens e mensagens em tela cheia aparecem aqui só como um rótulo, como `IMAGEM: cortina.jpg`. Você nunca vê a imagem em si.
- **Músicas:** as linhas cantadas têm notas musicais (♪) em volta e aparecem em itálico nas telas do público. O título de uma música aparece como um marcador azul só para você. Procure apertar Espaço no começo de cada frase.
- **Idiomas:** se o espetáculo tem traduções, você acompanha o idioma principal e tudo o que envia vai a todos os idiomas ao mesmo tempo. A contagem do público no cabeçalho lista cada idioma, por exemplo `83 (en 60 · es 23)`.
- **Várias vozes ao mesmo tempo:** uma linha marcada `JUNTOS` mostra dois ou mais cantores nas telas ao mesmo tempo. Um toque em Espaço envia todos.
- A faixa abaixo do cabeçalho mostra quem tem o controle, **Item n de total** e **Último envio há …**, para você ver de relance que o espetáculo e a sua tela continuam em compasso. Ele fica amarelo depois de dois minutos sem uma linha.
- **Ir para a seção…** (ao lado da caixa de busca) lista os marcadores do arquivo do espetáculo; escolher um o seleciona, e você confirma o salto.
- Para avançar ou voltar, clique numa linha e confirme «Pular para cá». Você também pode buscar com `/`.
- Um aviso avulso como «O espetáculo recomeça em instantes» pode ser enviado com **Aviso…**. Aperte **Voltar** para removê-lo. Enquanto está nas telas, um aviso azul desta página o cita.
- **Colocar em espera…** (no alto, à direita) tira o espetáculo das telas do público: elas voltam a ver a tela de espera e o espetáculo mantém o seu lugar. **INICIAR** o coloca de volta no ar exatamente onde estava.

Veja [o teclado](operator-keyboard.md) e [Se algo der errado](operator-if-something-goes-wrong.md).

## Depois do espetáculo: o relatório do público

Quando chegar ao fim do espetáculo, o aviso diz **Fim do espetáculo.** com um link **Relatório do público** (a etapa **Relatório** da configuração o mostra depois). O relatório diz quantos dispositivos diferentes usaram a página do público enquanto o espetáculo estava no ar, não importa quanto cada um ficou: um telefone que caiu do WiFi e voltou conta uma vez. Mostra os totais (telefones, tablets, computadores, telas de projetor, o máximo conectado ao mesmo tempo, quantos ainda estão conectados), os idiomas e uma linha por dispositivo. Também desenha um gráfico de quantos dispositivos estavam conectados ao longo do tempo, com os laços antes do espetáculo e do intervalo marcados, para ver quantas pessoas havia no começo e quantas ficaram durante a pausa. Continua contando até você voltar à espera, reiniciar ou carregar outro espetáculo, e pode ser impresso ou baixado como arquivo CSV para uma planilha.

Um segundo dispositivo que só acompanha o espetáculo (a página de acompanhamento) aparece na mesma tabela, marcado como **Espectador**, mas não é contado em nenhum número nem no gráfico. Os dispositivos são distinguidos pelo endereço de rede e pelo tipo de navegador, então os números são uma estimativa: telefones que compartilham os dois contam como um, e um telefone cujo endereço muda conta duas vezes. Cada relatório é salvo como um arquivo na pasta `reports` da pasta de dados; os 30 mais recentes são mantidos, e você pode excluir um com **Excluir…** na etapa **Relatório** da configuração, ou apagar os arquivos lá.
