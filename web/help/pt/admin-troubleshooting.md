# Solução de problemas

## Ninguém consegue abrir o endereço

- Confira se o telefone está no WiFi de visitantes do local, e não nos dados móveis.
- Confira se o endereço é igual ao da etapa **Ao vivo**. Se tiver número de porta, inclua-o.
- Pergunte ao local se os dispositivos do WiFi de visitantes podem alcançar este computador.
- O firewall do computador pode estar bloqueando o programa. No Windows, permita-o em **redes privadas** quando perguntado. No macOS, permita as conexões de entrada para ele.
- Abra o endereço neste computador usando `localhost` para testar o próprio programa.

## O número de dispositivos conectados é menor que o esperado

Os dispositivos que adormeceram se desconectam e se reconectam sozinhos quando acordam. Peça às pessoas que desliguem o bloqueio automático durante o espetáculo se a tela escurecer.

## As imagens aparecem devagar

Imagens grandes demoram num WiFi cheio. Procure na etapa Imagens os arquivos marcados como «grande» e troque-os por outros menores.

Para a lista completa de verificações do lado da rede, veja o [guia de rede do local](topic:admin-venue-network-guide).

## Falta um idioma para o público

Abra a etapa Arquivo do espetáculo e olhe a aba desse idioma: um ⚠ significa que o arquivo do espetáculo dele já não tem as mesmas deixas do principal. Corrija o erro mostrado; o idioma volta assim que você salvar. Veja [Idiomas](topic:admin-languages).

## Problemas no arquivo do espetáculo

Veja «O que significam os avisos» em [Escrever um arquivo do espetáculo](topic:admin-writing-a-show-file).

## Senhas

Para mudar qualquer uma das senhas, use **Configurações** no alto da página de Configuração. A nova senha vale na hora e é salva em `config.yaml`. Quem estiver conectado com uma senha que mudou é desconectado; você continua conectado. A mesma janela pode mostrar a senha de operador atual para que você a entregue ao operador.

Se você não consegue entrar de jeito nenhum, as duas senhas estão em `config.yaml`, na pasta de dados do programa. Abra o arquivo em qualquer editor de texto. O local dele é mostrado quando o programa inicia.

## Olhar o registro

**Configurações → Registro recente…** mostra o que o programa registrou há pouco. Escolha *Somente problemas* para ver avisos e erros, como entradas recusadas ou um nome já usado na rede. A mesma janela define quanto detalhe é registrado (**Nível de detalhe do registro**); *Depuração* acrescenta cada requisição, o que ajuda quando algo estranho acontece. As senhas nunca são escritas no registro.

## Pedir ajuda

Baixe o **pacote de suporte** no fim da etapa Ao vivo. Ele contém as configurações (sem senhas), linhas recentes do registro e informações básicas sobre este computador, e é fácil de enviar a quem mantém o programa.
