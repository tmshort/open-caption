# Musicais

As legendas de um musical funcionam melhor quando as letras se distinguem com facilidade das falas e o operador consegue acompanhar a música. Esta página reúne o que funciona.

## Marcar as letras

As letras são mostradas **em itálico entre notas musicais**, por exemplo ♪ Onde está o amor? ♪. Escolha a forma de escrevê-las que lhe convier:

| Escreva | Resultado |
|---------|-----------|
| `~ Where is love? ~` | ♪ Where is love? ♪ (o `~` de fechamento é opcional) |
| `OLIVER: ~ Where is love?` | o mesmo, com o nome do personagem acima |
| `[song: Title]` … `[/song]` | cada linha dentro é uma letra; o título é só para o operador |
| `♪ digitado ou colado ♪` | exatamente o que você digitou, sem itálico |
| `~` sozinho numa linha | só uma nota: ♪, para uma passagem instrumental |

Use `[song]` para números longos, e `~` para uma única linha cantada no meio de uma cena.

Na Configuração, os botões **🎵 Música**, **♪** e **♫** acima do editor do arquivo do espetáculo os inserem para você. Os caracteres ♪ e ♫ também podem ser copiados de qualquer lugar, porque são texto comum.

## Dois cantores ao mesmo tempo (contraponto e duetos)

Quando cantores diferentes cantam palavras diferentes **ao mesmo tempo**, mostre-os juntos nas telas do público com `[together]`:

```showfile
[together]
ANNA: ~ I love you, I love you ~
KRIS: ~ Me too, me too ~
[/together]
```

- As duas linhas aparecem **juntas** e o operador aperta Espaço **uma vez**.
- Numa tela larga as vozes ficam lado a lado, cada uma com uma barra na borda esquerda e o nome do personagem acima; num telefone elas se empilham.
- Acrescente quantas vozes precisar, por exemplo um trio ou uma fala por cima de uma parte cantada.
- Para uma linha rápida a duas vozes você pode escrever `ANNA: ~ I love you ~ || KRIS: ~ Me too ~` numa única linha.
- Vozes que **entram uma depois da outra** (um cânone, ou um segundo cantor que responde) são linhas separadas comuns: um toque em Espaço cada. Escreva a primeira voz e depois a segunda quando entrar.
- Mantenha cada voz curta. Três ou quatro palavras por voz são mais fáceis de ler ao mesmo tempo do que uma frase inteira.
- Use nomes de personagens para que o público distinga as vozes.

## Dicas para boas legendas de letras

- **Uma linha por frase cantada.** Quebre as linhas onde os cantores respiram, para que o toque do operador em Espaço combine com a música. Linhas mais curtas são mais fáceis de acompanhar.
- **Use nomes de personagens nos duetos** e nas linhas cantadas por personagens diferentes em sequência. Deixe-os de fora num coro.
- **Trechos instrumentais:** use `~` sozinho, ou `(A orquestra toca)`, para marcá-los. Uma linha entre parênteses é mostrada como indicação de cena.
- **Mantenha separadas as linhas faladas e as cantadas.** Uma linha de personagem seguida de letras com `~` se lê com clareza: o público vê as notas e o itálico mudarem.
- **Repetições:** escreva as linhas todas as vezes. O operador aperta Espaço uma vez por linha de qualquer jeito.
- **Limpar entre os números:** ponha `[clear]` depois da última linha de uma música para que a cena seguinte comece numa tela vazia.
- **Lembretes para o operador:** `[note: o andamento acelera depois da mudança de tom]` é mostrado só ao operador.
- **Ensaie com a música de verdade.** A etapa Teste permite que o operador pratique apertar Espaço com a gravação do elenco.

## Um exemplo

```showfile
[section: Act 1, Number 4]
[clear]
(The orchestra begins.)
[song: Where Is Love]
OLIVER: Where is love?
Does it fall from skies above?
[note: wait for the held note]
Is it underneath the willow tree
That I've been dreaming of?
[/song]
[clear]
MRS. CORNEY: That's quite enough of that!
```

Veja também [Escrever um arquivo do espetáculo](topic:admin-writing-a-show-file).
