# Preparar um arquivo do espetáculo

Esta página é para quem transforma o texto da peça em legendas. Você não precisa operar a apresentação nem saber nada sobre redes. Um arquivo do espetáculo é um arquivo de texto comum: você pode escrevê-lo no programa que preferir, em qualquer computador, e trazê-lo para a Configuração quando estiver pronto. [Escrever um arquivo do espetáculo](topic:admin-writing-a-show-file) lista todas as marcações que você pode usar; esta página trata de como fazer o trabalho.

## A versão curta

1. Coloque o texto da peça em um arquivo de texto simples. Se quiser, comece pelo espetáculo de exemplo em [/help/sample-show.md](/help/sample-show.md).
2. Faça com que **uma linha do arquivo seja uma legenda**: um toque na barra de espaço.
3. Acrescente o que ajuda o público e o operador (nomes de personagens, descrições de sons, seções, notas).
4. Traga o arquivo para a Configuração, leia a lista de problemas, corrija-os e salve.
5. Ensaie na etapa Teste e corrija o que notar.

## Que tipo de arquivo

- Qualquer arquivo de texto simples serve: um nome terminado em `.txt` ou `.md`, e qualquer nome, como `Nossa cidade - legendas v3.txt`.
- Salve como **UTF-8** se o programa perguntar. Assim letras acentuadas, aspas curvas e notas musicais são preservadas. Um arquivo salvo como «ANSI» do Windows (Windows-1252) também é entendido. Arquivos de processador de texto (`.doc`, `.docx`, `.pdf`) não são arquivos de texto: copie o texto e cole primeiro em um editor de texto simples, ou use Salvar como e escolha texto simples.
- As quebras de linha do Windows (CRLF) e as do Mac ou Linux (LF) funcionam, assim como a marca invisível que alguns editores colocam no início do arquivo.
- Um `.zip` só é necessário para um backup completo, que inclui as imagens. O arquivo do espetáculo é só o texto.

## Decidir onde cada legenda começa e termina

O programa nunca divide nem junta as suas linhas. Uma legenda é exatamente o que você digitou em uma linha, então você controla quanto o público lê de uma vez.

- Procure linhas que se leiam confortavelmente de relance: uma frase ou parte de uma frase. Falas muito longas devem ser cortadas em pausas naturais, onde o ator respira.
- Linhas em branco não significam nada, então use-as à vontade entre as falas para manter o arquivo fácil de ler.
- Se duas ou mais linhas devem aparecer juntas, una-as com uma barra invertida no fim da linha, ou coloque-as entre `[[` e `]]`.
- Use `HAMLET: fala` para mostrar quem está falando. Escreva o nome em maiúsculas, depois dois-pontos.
- Descreva sons e indicações de cena entre parênteses, em uma linha própria, por exemplo `(trovões ribombam)`. Eles são mostrados ao público em outro estilo.

```showfile
[section: Act 1]
[note: house lights down, then start]
GEORGE: Morning, Emily.
EMILY: Oh, hello, George.
(a train whistle, far away)
```

## O que só o operador vê

Marque o seu lugar e deixe lembretes sem incomodar o público: `[section: Ato 2]` acrescenta um marcador na lista do operador, e `[note: esperar os aplausos]` mostra um lembrete. Nenhum dos dois gasta um toque na barra de espaço. Um comentário entre `<!--` e `-->` é para você e ninguém o vê.

## Trabalhar no seu próprio editor

Você nunca é obrigado a digitar no editor da etapa Arquivo do espetáculo. Em vez disso:

1. Na etapa **Arquivo do espetáculo**, pressione **Baixar o arquivo do espetáculo**. Você recebe um arquivo `.md` de texto simples.
2. Abra-o no seu próprio programa, edite e salve. Mantenha-o como texto simples.
3. De volta à etapa Arquivo do espetáculo, use **Enviar um arquivo do espetáculo…** e escolha o seu. O texto dele substitui o do editor e a lista de problemas é atualizada na hora. Nada muda em definitivo até você pressionar **Salvar**.
4. Pressione **Salvar**. A versão substituída fica guardada em **Versões anteriores**, onde você pode compará-la com a nova, então um erro sempre pode ser desfeito.

O mesmo vale para uma tradução: escolha primeiro a aba do idioma, depois baixe e abra o arquivo desse idioma. Você também pode soltar um arquivo sobre o editor.

Para começar um **espetáculo novo** a partir de um arquivo, use **Importar um espetáculo** na etapa Espetáculo e escolha o arquivo. Ele vira um espetáculo novo, com o nome da linha `title:` do seu bloco `[settings]`, ou o nome do arquivo se não houver. As imagens não fazem parte de um arquivo de texto: envie-as na etapa Imagens, com os mesmos nomes das suas linhas `[image: ...]`.

## Conferir o seu trabalho

- A etapa Arquivo do espetáculo lista cada problema com o número da linha. Clique em um para ir até ele. Os erros nunca impedem você de salvar, mas um item com erro é ignorado, então corrija-os antes da apresentação.
- O painel ao lado do editor mostra o que o operador vai percorrer, uma linha por toque na barra de espaço. Confira se o número de linhas é o que você espera.
- Execute a etapa **Teste** para ver as legendas como o público as verá, em um telefone além de uma tela grande.
- Se outras pessoas também trabalham no arquivo do espetáculo, veja *Se duas pessoas editam o mesmo arquivo* em [Escrever um arquivo do espetáculo](topic:admin-writing-a-show-file).

## Antes de entregar

- Leia o arquivo uma vez de cima a baixo como o público faria, não como um revisor.
- Confira se cada nome em `[image: ...]` corresponde a uma imagem enviada, e se cada `[loop]`, `[song]`, `[together]` e `[fullscreen]` tem a sua linha de fechamento.
- Dê ao operador uma cópia do arquivo e os nomes das seções, para que ele saiba onde encontrar o seu lugar. [Canções e música](topic:admin-musicals), [Repetições](topic:admin-loops) e [Idiomas](topic:admin-languages) detalham essas partes.
