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
