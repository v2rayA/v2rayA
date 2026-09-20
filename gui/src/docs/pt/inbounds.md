# Entradas e compartilhamento

Uma entrada é uma porta na qual o núcleo aceita conexões de aplicativos ou de outros dispositivos.

## Entradas integradas

**Configurações → Endereços e portas** define essas entradas; 0 fecha uma entrada de proxy e permite que a porta da API seja escolhida aleatoriamente.

| Entrada           | Padrão  | Roteamento                                                         |
| ----------------- | ------- | ------------------------------------------------------------------ |
| SOCKS5            | `20170` | tudo pelo `proxy`                                                  |
| HTTP              | `20171` | tudo pelo `proxy`                                                  |
| SOCKS5 com regras | desativada | o modo da porta de regras                                       |
| HTTP com regras   | `20172` | o modo da porta de regras                                          |
| VMess com regras  | desativada | uma entrada VMess para outros dispositivos; a página mostra seu link de compartilhamento |
| API               | aleatória | a API do próprio núcleo, usada pelo v2rayA para estatísticas e para o balanceamento |

As entradas de proxy escutam em `127.0.0.1`. **Compartilhamento de portas** (Configurações → Proxy) faz com que escutem em todas as interfaces para que celulares e outras máquinas da rede local possam usá-las; a porta da API permanece no loopback. Em uma rede na qual você não confia, adicione uma entrada personalizada com nome de usuário e senha para os outros dispositivos e feche as portas SOCKS e HTTP integradas, que não exigem senha, ou bloqueie-as no firewall.

## Entradas personalizadas

**Configurações → Portas de entrada personalizadas** adiciona entradas SOCKS ou HTTP com sua própria porta, nome de usuário e senha opcionais e seu próprio roteamento: um grupo de proxy fixo ou regras RoutingA escritas para aquela entrada, com o grupo escolhido como destino quando nenhuma regra corresponder (`default:` é ignorado ali). Assim, duas entradas personalizadas podem enviar tráfego para dois grupos diferentes, e uma entrada cujas regras terminam em `network(tcp, udp) -> direct` não usa o proxy para nenhum aplicativo configurado para usá-la.

A tag é o nome da entrada na configuração do núcleo e nas condições `inboundTag(...)`.

## Docker e mapeamento de portas

Em um contêiner iniciado com `--network=host`, as entradas e o proxy transparente atuam no host. Com a rede em modo bridge, publique a porta 2017 e as portas de entrada que você usa e ative Compartilhamento de portas para que escutem nas interfaces do contêiner; nesse caso, o proxy transparente vê apenas o contêiner. O v2rayA não consegue determinar se uma porta mapeada está livre no host.
