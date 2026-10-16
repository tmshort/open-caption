# Imagens

Envie imagens na etapa **Imagens** e use-as no arquivo do espetáculo com `[image: nome.png]`.

- Tipos aceitos: PNG, JPG, GIF e WebP. Outros tipos, incluindo SVG, não são aceitos.
- Cada arquivo pode ter até 10 MB, mas **mantenha-os abaixo de cerca de 1 MB**. Cada dispositivo do público baixa a imagem inteira, e 100 dispositivos buscando uma imagem grande ao mesmo tempo podem ficar lentos. Cerca de 1920 pixels de largura bastam.
- O programa não redimensiona as imagens. Os navegadores as reduzem para caber em cada tela.
- Uma imagem **sem legenda preenche a tela inteira**, de borda a borda: ela é ajustada sem nunca ser cortada, e a cor de fundo aparece nas faixas se as proporções forem diferentes. Adicione uma legenda (`[image: curtain.jpg | Act 1]`) e a imagem mantém uma margem com as palavras abaixo.
- As próximas imagens são baixadas em segundo plano antes de serem mostradas, então aparecem na hora.

## Nomes

O arquivo do espetáculo se refere a uma imagem pelo nome do arquivo, como `cortina.jpg`. Nomes com espaços funcionam. Se o arquivo do espetáculo usar um nome que você não enviou, a etapa Imagens o lista em «Imagens que o arquivo do espetáculo usa mas que você não enviou», com um botão para enviar o arquivo certo com esse nome. Um erro pequeno como `cortna.jpg` é apontado com uma sugestão.

## Substituir e excluir

Enviar um arquivo com o mesmo nome o substitui. Excluir uma imagem que o arquivo do espetáculo ainda usa é permitido após um aviso; o arquivo do espetáculo passa então a mostrar um aviso de «não encontrada» para ela.

## Em um laço

As imagens de um laço se alternam na tela: `[image: patrocinadores.png | 15s]` mostra essa por 15 segundos. Veja [Laços](topic:admin-loops).
