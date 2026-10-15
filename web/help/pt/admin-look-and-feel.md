# Aparência

A etapa **Aparência** define como as legendas aparecem em cada dispositivo do público. Os padrões, texto branco sobre fundo preto, são os mais fáceis de ler numa sala escura e os melhores para a acessibilidade.

- **Cores:** uma leitura do contraste mostra o quanto a combinação é legível. Busque 7:1 ou mais; a página avisa quando está baixo demais.
- **Tamanho do texto:** escala todo o texto. Os dispositivos ajustam o texto à tela, então é uma mudança relativa, não um tamanho em pontos.
- **Fonte:** limpa (a fonte habitual do dispositivo), com serifa, monoespaçada, ou **amigável para dislexia (OpenDyslexic)**, uma fonte com a base das letras mais pesada e formas de letra marcantes que alguns leitores com dislexia acham mais fáceis. Ela vem embutida no programa, então não precisa de internet, e um dispositivo a baixa (cerca de 115 KB) só quando ela é usada. As linhas ganham um pouco de altura com ela. Escolha **alta legibilidade (Atkinson Hyperlegible)** para uma fonte limpa projetada pelo Braille Institute para baixa visão, com letras fáceis de distinguir (por exemplo o I maiúsculo, o l minúsculo e o número 1). As duas fontes vêm embutidas no programa, e um dispositivo baixa uma só se a usar (cerca de 17 KB a Atkinson Hyperlegible). Os espectadores também podem escolher qualquer uma das duas para o próprio dispositivo no pequeno controle **Aa**, seja qual for a fonte que você escolheu aqui.
- **Nomes dos personagens:** mostrados acima da linha em uma só cor (a *cor dos nomes dos personagens*), em uma **cor diferente para cada personagem**, ou ocultos. As cores dos personagens são escolhidas automaticamente, na ordem em que cada personagem fala pela primeira vez, a partir de uma paleta que funciona para espectadores daltônicos, e ajustadas para continuarem legíveis sobre o seu fundo.
- **Mensagem antes de o espetáculo começar:** mostrada até a primeira linha ser enviada, se o arquivo do espetáculo não começar com um laço.

As mudanças são salvas com **Salvar aparência** e chegam logo aos dispositivos conectados. **Salvar aparência grava as suas escolhas no bloco `[settings]` no início do arquivo do espetáculo** (veja [Escrever um arquivo do espetáculo](topic:admin-writing-a-show-file)), de modo que o arquivo do espetáculo leva sua própria aparência para onde for copiado; você também pode editar esse bloco à mão. Tudo o que o arquivo do espetáculo não define usa os padrões de `config.yaml` (a seção `defaults`), que por sua vez partem dos valores embutidos do programa. Se o editor do arquivo do espetáculo tiver alterações não salvas, salve-as antes de salvar a aparência.

## Uma tela para a sala toda

Se um navegador estiver mostrando as legendas num projetor ou monitor grande, abra o endereço do público com `?display=projector` no final. O texto fica maior, o ponteiro do mouse some e nada mais é mostrado.

## Reler linhas anteriores

Qualquer pessoa pode rolar as legendas para cima para reler algo que perdeu. As linhas novas esperam então atrás de um botão **↓ Voltar ao vivo (3)** (que conta o que chegou) em vez de arrastar a página, e as linhas mais antigas param de esmaecer enquanto alguém as lê. Tocar no botão, ou rolar até o fim, volta às legendas ao vivo. Uma tela com `?display=projector` nunca rola.

## Controles do público

Os espectadores nunca precisam fazer nada. Um pequeno controle discreto **Aa** no canto superior direito da página do público permite que cada pessoa aumente ou diminua o próprio texto, mude para escuro sobre claro, escolha a fonte amigável para dislexia ou a de alta legibilidade, ou desligue **A tela fica acesa** (a página mantém acesa a tela do telefone durante o espetáculo, e quem está com pouca bateria pode preferir não), mas somente no próprio dispositivo.
