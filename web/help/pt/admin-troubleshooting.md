# Solução de problemas

## O botão Solução de problemas

O botão **Solução de problemas**, no alto da configuração, ao lado de Configurações, é o lugar para começar quando as pessoas não conseguem se conectar ou algo parece errado. Abre em três abas: **Verificações**, **Teste de ligação** e **Avisos e erros**. As verificações listam o que o programa sabe: se este computador tem um endereço de rede para dar, se são fornecidos endereços de mais de uma rede, quais portas (HTTP e HTTPS) o servidor web realmente obteve, se o HTTPS e o certificado estão em ordem, se o nome `.local` é anunciado, se o arquivo de configurações pode ser salvo, se cada idioma corresponde ao arquivo principal do espetáculo se uma lista de acesso recusa computadores (ou deixa aberta a página Acompanhar), se o relógio do computador concorda com o do navegador, se o arquivo do espetáculo carregado tem erros ou falta uma imagem, se a pasta de dados pode ser escrita e tem espaço, se as palavras-passe são iguais ou estão em texto simples, se o espetáculo está em espera ou no ar (o que o público vê agora) e quantos dispositivos estão conectados. Cada verificação é marcada **OK**, **Nota**, **Verificar** ou **Problema**, e as que podem ser corrigidas têm um botão que leva à configuração certa. Abaixo ficam os avisos e erros mais recentes do registro, um link para o registro completo, **Copiar resumo** (as verificações como texto, para colar num e-mail) e o pacote de suporte. **Testar a ligação** pede o servidor web em cada endereço fornecido pelo programa e diz quais respondem e com que rapidez; uma firewall neste computador ou na rede pode mesmo assim bloquear outros dispositivos.

Quando o programa registra novos avisos ou erros, aparece um número no botão até você abri-lo.

## Ninguém consegue abrir o endereço

- Confira se o telefone está no WiFi de visitantes do local, e não nos dados móveis.
- Confira se o endereço é igual ao da etapa **Compartilhar**. Se tiver número de porta, inclua-o.
- Pergunte ao local se os dispositivos do WiFi de visitantes podem alcançar este computador.
- O firewall do computador pode estar bloqueando o programa. No Windows, permita-o em **redes privadas** quando perguntado. No macOS, permita as conexões de entrada para ele.
- Abra o endereço neste computador usando `localhost` para testar o próprio programa.

## O número de dispositivos conectados é menor que o esperado

Os dispositivos que adormeceram se desconectam e se reconectam sozinhos quando acordam. Peça às pessoas que desliguem o bloqueio automático durante o espetáculo se a tela escurecer.

## As imagens aparecem devagar

Imagens grandes demoram num WiFi cheio. Procure na etapa Imagens os arquivos marcados como «grande» e troque-os por outros menores.

Para a lista completa de verificações do lado da rede, veja o [guia de rede do local](../../../docs/i18n/pt/venue-network-guide.md).

## Falta um idioma para o público

Abra a etapa Arquivo do espetáculo e olhe a aba desse idioma: um ⚠ significa que o arquivo do espetáculo dele já não tem as mesmas deixas do principal. Corrija o erro mostrado; o idioma volta assim que você salvar. Veja [Idiomas](admin-languages.md).

## Problemas no arquivo do espetáculo

Veja «O que significam os avisos» em [Escrever um arquivo do espetáculo](admin-writing-a-show-file.md).

## Senhas

Para mudar qualquer uma das senhas, use **Configurações** no alto da página de Configuração. A nova senha vale na hora e é salva em `config.yaml`. Quem estiver conectado com uma senha que mudou é desconectado; você continua conectado. As senhas são guardadas como hashes, então a configuração não pode mostrar nenhuma: anote a senha ao defini-la. Uma **senha de espectador** opcional permite que um segundo dispositivo, como o tablet de um contrarregra, abra a página de acompanhamento (`/follow`) só para acompanhar: ela mostra as linhas, as notas e a posição, e dela não se pode enviar nada. As senhas de operador e de administrador também abrem essa página.

Se você não consegue entrar de jeito nenhum, as duas senhas estão em `config.yaml`, na pasta de dados do programa. Abra o arquivo em qualquer editor de texto. O local dele é mostrado quando o programa inicia.

## Olhar o registro

**Solução de problemas → Avisos e erros** mostra o que o programa registrou há pouco; abre em *Somente problemas* (avisos e erros, como entradas recusadas ou um nome já usado na rede), e o menu de nível amplia para *Tudo*. **Configurações → Sobre** define quanto detalhe é registrado (**Nível de detalhe do registro**); *Depuração* acrescenta cada requisição, o que ajuda quando algo estranho acontece. As senhas nunca são escritas no registro.

## Pedir ajuda

Baixe o **pacote de suporte** com o botão **Solução de problemas** no alto da configuração. Ele contém as configurações (sem senhas), linhas recentes do registro e informações básicas sobre este computador, e é fácil de enviar a quem mantém o programa.
