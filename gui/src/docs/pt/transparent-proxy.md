# Proxy transparente

Com o proxy transparente ativado, o tráfego chega ao núcleo sem nenhuma configuração nos aplicativos; o tráfego abrangido depende da implementação. **Configurações → Proxy transparente/Proxy do sistema** ativa o recurso e seleciona a política de divisão; a configuração abaixo seleciona a implementação.

## Políticas

| Política                    | Tráfego pelo proxy                                                                                                                                                                                                    |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Não dividir                 | tudo                                                                                                                                                                                                                  |
| Proxy exceto sites da China | tudo, exceto domínios e endereços chineses (`geosite:cn`, `geoip:cn`) e endereços privados; sites fora da China, Google e endereços de Hong Kong e Macau passam pelo proxy mesmo quando uma regra chinesa corresponde |
| Proxy apenas para GFWList   | os domínios da GFWList (`gfw` e `greatfire` do arquivo de Loyalsoldier) e as faixas de endereços do Telegram; execute **Atualizar GFWList** primeiro, pois o modo é recusado sem o arquivo                            |
| Igual ao da porta de regras | o modo da porta de regras, incluindo RoutingA                                                                                                                                                                         |

## Implementações

| Implementação    | Plataformas                                                        | Observações                                                                                                              |
| ---------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| `redirect`       | Linux                                                              | `REDIRECT` do iptables/nftables; apenas TCP, além de DNS na porta 53 redirecionado ao módulo DNS do núcleo (porta 52353) |
| `tproxy`         | Linux                                                              | `TPROXY` do iptables/nftables; TCP e UDP                                                                                 |
| `tun`            | Linux, Windows, macOS                                              | o núcleo abre um dispositivo TUN; TCP e UDP; exclui o v2rayA e o próprio núcleo automaticamente                          |
| Proxy do sistema | Windows; Linux e macOS quando não é executado como root (`--lite`) | configura o proxy do ambiente gráfico (GNOME e KDE no Linux); abrange apenas os aplicativos que as respeitam             |

`redirect` e `tproxy` exigem root e `iptables` ou `nftables`; `tun` exige `/dev/net/tun` e `ip` no Linux e privilégios de administrador no Windows e no macOS.

**Prefixos de interfaces excluídos** mantém o tráfego que chega pelas interfaces indicadas (bridges do Docker, túneis VPN; `docker*`, `veth*`, `wg*`, `ppp*` por padrão) fora de `redirect` e `tproxy`; o DNS dessas interfaces ainda é interceptado.

`--redirect-respect-bound-device` (Linux) deixa as conexões TCP vinculadas a uma interface com `SO_BINDTODEVICE` fora do `redirect`: as verificações de conectividade do NetworkManager são conexões desse tipo e, redirecionadas, informam uma conexão limitada e mantêm os aplicativos offline. Desligado por padrão. Quando ligado, o serviço marca esses sockets com `0x80` por cgroup BPF; isso requer kernel 5.14 ou mais recente e cgroup v2, e sem eles o `redirect` não inicia.

## TUN

O núcleo cria o dispositivo TUN e atribui seu endereço. Com **Rota automática** ativada, o v2rayA instala as rotas e direciona o resolvedor do sistema para o núcleo; com ela desativada, os scripts de configuração e remoção em **Configurar script de rota** fazem isso.

Nunca entram no TUN: as conexões do próprio núcleo, a saída `direct` e as consultas do módulo DNS aos servidores upstream (por marca de socket no Linux, por vinculação à interface física no Windows e no macOS). O v2rayA e o núcleo são sempre excluídos; **Processos excluídos do TUN** exclui outros pelo nome do executável, um por linha, para conexões cujo processo responsável pode ser identificado; o DNS desses processos ainda vai para o módulo DNS do núcleo. Rotas mais específicas que a padrão (redes conectadas, rotas estáticas) também não passam pelo TUN.

O DNS sem criptografia na porta 53 que chega ao TUN é respondido pelo módulo DNS do núcleo de acordo com as regras de DNS; o DNS criptografado não é interceptado. No Windows, o resolvedor do sistema é direcionado para o gateway do TUN; no macOS, para o serviço do núcleo que escuta em `127.0.0.1`, que precisa da porta 53 livre.

Limitação conhecida: no Windows e no macOS, um aplicativo que consulta diretamente um resolvedor da rede local ainda não passa pelo TUN.

## DNS

**Configurações → Configurações de DNS** contém as regras que o módulo DNS do núcleo segue: qual servidor upstream responde por quais domínios e se a consulta sai diretamente. Os padrões enviam nomes privados para `127.0.0.1:53` (`localhost`, que precisa de um resolvedor escutando ali), `geosite:cn` diretamente para `223.5.5.5` e todo o resto para `1.0.0.1` pelo proxy. Uma saída `direct` consulta diretamente; qualquer outro valor envia a consulta pela entrada SOCKS local, de modo que ela é roteada como tráfego SOCKS. A regra com uma lista de domínios vazia responde por todos os domínios que nenhuma outra regra indica. Um servidor upstream é um endereço (`8.8.8.8`, `dns.google`), `tcp://host`, `tls://host` para DNS sobre TLS ou `https://host/dns-query` para DNS sobre HTTPS; DNS sobre QUIC não é suportado e é recusado ao salvar.

## Compartilhamento com a rede local

**Compartilhamento de portas** faz as entradas SOCKS, HTTP e personalizadas escutarem em todas as interfaces em vez de `127.0.0.1`, para que outros dispositivos possam usar esta máquina como proxy; a porta da API permanece no loopback. Para rotear o tráfego desses dispositivos também pelo proxy transparente, use `redirect`, `tproxy` ou `tun`, ative **Encaminhamento de IP**, mantenha a interface pela qual o tráfego chega fora dos prefixos excluídos e direcione o gateway padrão dos dispositivos para esta máquina; os modos de proxy do sistema não podem fazer isso.
