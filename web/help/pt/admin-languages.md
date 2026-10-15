# Idiomas

Um espetáculo pode oferecer suas legendas em mais de um idioma. Cada idioma é o seu **próprio arquivo do espetáculo**, e o operador acompanha só o principal: um toque em Espaço envia o mesmo momento em todos os idiomas ao mesmo tempo.

## Adicionar um idioma

1. Na etapa **Arquivo do espetáculo**, clique em **＋ Adicionar um idioma** ao lado das abas de idioma.
2. Escolha o idioma (ou «Outro idioma…» e digite um código curto como `cy`).
3. Deixe marcada a caixa **Começar com uma cópia do arquivo principal do espetáculo**. Você recebe uma cópia na forma certa, então só precisa trocar as palavras.
4. Traduza e depois **Salve**.

Cada tradução começa com uma linha que nomeia o seu idioma. O editor a acrescenta para você:

```showfile
[language: es | Español]
Hola, Margaret.
```

O nome depois do `|` é o que os espectadores veem no menu de idiomas, então escreva-o no próprio idioma (Español, Français, Deutsch).

## A única regra: manter a mesma forma

Cada idioma deve ter **as mesmas deixas na mesma ordem**: uma legenda continua sendo uma legenda, um `[clear]` continua sendo um `[clear]`, um laço mantém o mesmo número de itens, e vozes mostradas juntas continuam sendo o mesmo número de vozes. Só as palavras mudam. Nomes de personagens, notas, nomes de seções, nomes de imagens e tempos de laços podem ser diferentes.

Se uma tradução deixar de coincidir, o editor mostra um erro com a linha em que ela se desencontra, por exemplo «A deixa 4 não coincide com o arquivo do espetáculo em inglês: aquele arquivo tem [clear] aqui, e este tem uma legenda». Um idioma que não coincide **não é oferecido ao público** até ser corrigido. O arquivo do espetáculo em inglês (principal) não é afetado, e a etapa Ao vivo lista o problema.

Para dizer ao público quais palavras vão com qual linha, mantenha as linhas em compasso: se o inglês divide uma fala em três linhas, o espanhol também divide, mesmo que não seja assim que o idioma a quebraria.

## O que o público vê

- Cada dispositivo recebe o melhor idioma que o espetáculo tem. Olham-se as configurações de idioma do telefone ou do navegador (por exemplo o espanhol do México recebe espanhol), e usa-se o idioma principal quando nada coincide.
- Um pequeno controle **🌐** no canto superior esquerdo lista os idiomas. Escolher um muda aquele dispositivo, e o endereço passa a terminar em `?lang=es`, de modo que você pode imprimir um link ou código QR para cada idioma.
- As próprias palavras da página («Reconectando…», tamanho do texto etc.) também estão traduzidas para espanhol, francês, alemão, italiano e português. Os outros idiomas mostram esses poucos rótulos em inglês; as legendas em si estão sempre no idioma do arquivo do espetáculo.
- Idiomas escritos da direita para a esquerda (árabe, hebraico) se alinham à direita.

## Outras coisas a saber

- **Imagens:** todos os idiomas podem usar as mesmas imagens. Para uma imagem com palavras, envie uma versão por idioma e indique cada uma no arquivo do espetáculo desse idioma.
- **Mensagem de pré-espetáculo:** cada idioma tem a sua, guardada no arquivo do espetáculo desse idioma (`idle message:` no seu bloco `[settings]`). A etapa Aparência tem uma caixa para cada idioma e as escreve para você. O arquivo do espetáculo de uma tradução não pode definir mais nada: a aparência é definida uma só vez, no arquivo principal do espetáculo.
- **As mensagens do operador** («Mensagem…») são enviadas exatamente como digitadas a todos os idiomas, então digite-as nos idiomas de que precisar.
- **Alterar um arquivo do espetáculo com o espetáculo em andamento:** adicionar ou corrigir uma tradução vale na hora e mantém o lugar do operador. Remover um idioma não é permitido enquanto o espetáculo está ao vivo.
- **Contagens:** a página do operador mostra quantos dispositivos estão lendo cada idioma, por exemplo `👥 83 (en 60 · es 23)`.
- **Cópias de segurança:** Exportar guarda todos os idiomas; Duplicar também os copia.

Veja também [Escrever um arquivo do espetáculo](topic:admin-writing-a-show-file).
