# Laços: antes do espetáculo e intervalo

Um **laço** é um conjunto de imagens e textos em tela cheia que se repete até o operador seguir adiante. Use um antes de o espetáculo começar e outro no intervalo.

```showfile
[loop: 10s]
[image: sponsors.png | 15s]
**Welcome!**
Please silence your phones
[/loop]
```

Cada linha dentro do laço é um item:

- uma linha de texto simples é mostrada como texto em tela cheia;
- `[image: arquivo.png]`, `![legenda](arquivo.png)` e `[fullscreen: texto]` também funcionam;
- `[blank: 3s]` é uma pequena pausa com a tela em branco;
- acrescente `| 15s` ao fim de um item para mudar quanto tempo ele fica;
- `[loop: 10s]` define o tempo dos itens que não dizem nada. Se você não der tempo algum, usa-se o padrão de `config.yaml` (10 segundos).

Legendas, `[clear]`, seções, notas e outros laços não podem ficar dentro de um laço.

## Quando um laço começa e termina

- Se a **primeira coisa** do arquivo do espetáculo for um laço, ele começa sozinho assim que o espetáculo é iniciado, e o público vê a pré-exibição enquanto procura seus lugares.
- Qualquer outro laço (por exemplo no intervalo) começa quando o operador aperta a barra de espaço sobre ele.
- A barra de espaço encerra o laço e envia a próxima linha. **Shift+Q** para o laço e deixa a tela em branco. **Voltar** retorna ao laço e o reinicia a partir do primeiro item.

A lista do operador mostra um laço como uma única entrada que se expande para listar os itens. Veja [Laços para o operador](topic:operator-loops).

## Um temporizador de contagem regressiva

Dê um **temporizador** a um laço e as telas do público podem mostrar quanto tempo falta, contando ao vivo («Voltamos em 14:32»). Escreva `timer` nas opções do laço e ponha `{timer}` em qualquer texto desse laço onde o tempo deva aparecer.

```showfile
[loop: timer 15 minutes]
[fullscreen: **INTERMISSION**]
Back in {timer}
![Refreshments are in the lobby](lobby.png)
[/loop]
```

O valor de um temporizador é uma de duas coisas:

- Um **número de minutos**, como `timer 15` ou `timer 15 minutes`: para um intervalo. A contagem começa no momento em que o laço começa (quando o operador pressiona a barra de espaço nele) e recomeça com o tempo todo se o operador voltar para o laço.
- Uma **hora do dia**, como `timer 7:30pm`, `timer 7pm` ou `timer 19:30`: para o antes do espetáculo. As telas contam até essa hora de hoje. Usam-se o relógio e o fuso horário do computador que executa o programa, então confira-os. Uma hora que já passou mostra 0:00.

As opções dividem os colchetes, separadas por vírgulas, em qualquer ordem.

```showfile
[loop: 10s, timer 7:30pm]
**Welcome!**
The show begins in {timer}
[/loop]
```

- O tempo restante é mostrado em minutos e segundos (`14:32`), com horas quando houver mais de uma. Em zero fica em `0:00`; o laço continua até o operador seguir adiante.
- Só os itens cujo texto contém `{timer}` o mostram, então você pode alternar uma mensagem com a contagem.
- Um telefone com o relógio errado mostra mesmo assim o tempo certo, porque se usa o relógio do programa.
- Salvar o arquivo do espetáculo durante um intervalo não reinicia a contagem, nem reiniciar o programa.
- Em uma tradução, o laço deve ter o mesmo temporizador que o arquivo principal do espetáculo; traduza as palavras em volta de `{timer}`.
- O operador também vê o tempo restante no aviso verde. Um `{timer}` em um laço sem temporizador é mostrado como foi escrito, com um aviso.
