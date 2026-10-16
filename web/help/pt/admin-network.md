# Rede e conexão

Os espectadores conectam seus próprios telefones e tablets ao **WiFi de visitantes** do local e abrem um endereço. Nada precisa ser instalado.

## O que o local precisa fornecer

- Este computador e os dispositivos do público devem estar na mesma rede, ou a rede deve permitir que os dispositivos de visitantes alcancem este computador. É uma configuração do sistema de WiFi do local, não deste programa.
- Muitas redes de visitantes isolam os dispositivos uns dos outros. Tudo bem. Os dispositivos só falam com este computador, nunca entre si, desde que este computador esteja acessível.
- Uma conexão por cabo é o melhor para este computador. Peça um endereço reservado (uma «reserva DHCP») para que o endereço não mude entre espetáculos.
- Nenhuma conexão com a internet é necessária.

Teste antes do primeiro espetáculo: conecte um telefone ao WiFi de visitantes e abra o endereço mostrado na etapa **Conduzir**.

## Deixar o endereço fácil

1. **Código QR** — imprima o cartaz na etapa Conduzir e coloque no saguão e no programa. É a opção mais fácil.
2. **Um endereço curto** — se o programa puder usar a porta 80, o endereço não tem número de porta, como `http://192.168.1.10/`.
3. **Um nome** — o programa se anuncia como **`open-caption.local`**, então `http://open-caption.local/` funciona na maioria dos telefones e computadores da mesma rede, sem configurar nada. Mude o nome, ou desligue-o, em **Configurações → Nome de rede**. Algumas redes WiFi de visitantes bloqueiam esse tipo de anúncio; então use o endereço numérico ou o código QR. O local também pode acrescentar no roteador um nome próprio (como `captions.theater`), que funciona em qualquer lugar.
4. **Página automática** — um local pode mandar os visitantes à página quando entram no WiFi, com a página de boas-vindas do ponto de acesso («portal cativo»). O programa também responde ao endereço padrão de detecção de rede `/captive-api` para locais que usam a opção 114 do DHCP, mas isso apenas informa que a rede não é cativa: não abre a página sozinho, e exige um certificado HTTPS confiável. Veja o [guia de rede do local](topic:admin-venue-network-guide).

## HTTP e HTTPS

Endereços `http://` simples funcionam em todos os dispositivos e são o padrão. O HTTPS pode ser ativado em `config.yaml`; com um certificado autoassinado, cada dispositivo mostra um aviso de segurança na primeira vez. O HTTP simples é mais simples para um público.

## Se a porta estiver ocupada

Se a porta 80 já estiver sendo usada por outro programa, o servidor inicia em outra porta e informa no console. Os endereços do público precisam então incluí-la, por exemplo `http://192.168.1.10:8080/`; a etapa Conduzir mostra o endereço certo.
