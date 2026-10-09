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
| `redirect`       | Linux                                                              | `REDIRECT` do iptables/nftables; apenas TCP, além de DNS na porta 53 redirecionado ao módulo DNS do núcleo (porta 52353) no modo *serviço e interceptação* |
| `tproxy`         | Linux                                                              | `TPROXY` do iptables/nftables; TCP e UDP                                                                                 |
| `tun`            | Linux, Windows, macOS                                              | o núcleo abre um dispositivo TUN; TCP e UDP; exclui o v2rayA e o próprio núcleo automaticamente                          |
| Proxy do sistema | Windows; Linux e macOS quando não é executado como root (`--lite`) | configura o proxy do ambiente gráfico (GNOME e KDE no Linux); abrange apenas os aplicativos que as respeitam             |

`redirect` e `tproxy` exigem root e `iptables` ou `nftables`; `tun` exige `/dev/net/tun` e `ip` no Linux e privilégios de administrador no Windows e no macOS.

**Prefixos de interfaces excluídos** mantém o tráfego que chega pelas interfaces indicadas (bridges do Docker, túneis VPN; `docker*`, `veth*`, `wg*`, `ppp*` por padrão) fora de `redirect` e `tproxy`; o DNS dessas interfaces só é interceptado no modo *serviço e interceptação*.

`--redirect-respect-bound-device` (Linux) deixa as conexões TCP vinculadas a uma interface com `SO_BINDTODEVICE` fora do `redirect`: as verificações de conectividade do NetworkManager são conexões desse tipo e, redirecionadas, informam uma conexão limitada e mantêm os aplicativos offline. Desligado por padrão. Quando ligado, o serviço marca esses sockets com `0x80` por cgroup BPF; isso requer kernel 5.14 ou mais recente e cgroup v2, e sem eles o `redirect` não inicia.

## TUN

O núcleo cria o dispositivo TUN e atribui seu endereço. Com **Rota automática** ativada, o v2rayA instala as rotas e, no modo *serviço e interceptação*, direciona o resolvedor do sistema para o núcleo; com ela desativada, os scripts de configuração e remoção em **Configurar script de rota** fazem isso.

Nunca entram no TUN: as conexões do próprio núcleo, a saída `direct` e as consultas do módulo DNS aos servidores upstream (por marca de socket no Linux, por vinculação à interface física no Windows e no macOS). O v2rayA e o núcleo são sempre excluídos; **Processos excluídos do TUN** exclui outros pelo nome do executável, um por linha, para conexões cujo processo responsável pode ser identificado; no modo *serviço e interceptação*, o DNS desses processos ainda vai para o módulo DNS do núcleo. Rotas mais específicas que a padrão (redes conectadas, rotas estáticas) também não passam pelo TUN.

No modo *serviço e interceptação*, o DNS sem criptografia na porta 53 que chega ao TUN é respondido pelo módulo DNS do núcleo de acordo com as regras de DNS; o DNS criptografado não é interceptado. Nesse caso, no Windows, o resolvedor do sistema é direcionado para o gateway do TUN; no macOS, para o serviço do núcleo que escuta em `127.0.0.1`, que precisa da porta 53 livre.

Nos outros dois modos o TUN não monta esse relay e o resolvedor do sistema não é alterado em nenhuma das duas plataformas; uma consulta que ainda chegar sai direto do dispositivo. O mesmo vale com a **Rota automática** desligada e o script de rota no comando da rede: o v2rayA não muda a configuração do resolvedor, então o DNS também cabe ao script.

Limitação conhecida: no Windows e no macOS, um aplicativo que consulta diretamente um resolvedor da rede local ainda não passa pelo TUN.

## DNS

**Configurações → Configurações de DNS** contém o modo do DNS e as regras que o módulo DNS do núcleo segue.

| Modo                  | Módulo DNS | Consultas do sistema                                                       |
| --------------------- | ---------- | ------------------------------------------------------------------------- |
| Desligado             | não roda   | não são alteradas                                                          |
| Somente serviço       | roda       | não são alteradas: o resolvedor, as regras de DNS do firewall e o relay da TUN ficam como estão |
| Serviço e interceptação | roda      | são direcionadas a ele: o resolvedor é reapontado e a porta 53 é desviada   |

O modo responde a duas perguntas que antes se confundiam: se o módulo roda e se as consultas do sistema vão para ele — e o modo *somente serviço* responde de forma diferente às duas. *Desligado* e *somente serviço* deixam o TCP/UDP simples na porta 53 passar em `redirect`, `tproxy` e no proxy do sistema, e o `tun` o envia direto para fora; o DNS criptografado nunca vai para o módulo, então não há motivo para abdicar da interceptação por causa dele. Escolha *somente serviço* quando outro programa já aponta o DNS para o v2rayA, e *desligado* quando nada deve responder DNS por você.

As regras dizem qual servidor upstream responde por quais domínios e se a consulta sai diretamente. Os padrões enviam nomes privados para `127.0.0.1:53` (`localhost`, que precisa de um resolvedor escutando ali), `geosite:cn` diretamente para `223.5.5.5` e todo o resto para `1.0.0.1` pelo proxy. Uma saída `direct` consulta diretamente; qualquer outro valor envia a consulta pela entrada SOCKS local, de modo que ela é roteada como tráfego SOCKS. A regra com uma lista de domínios vazia responde por todos os domínios que nenhuma outra regra indica. Um servidor upstream é um endereço (`8.8.8.8`, `dns.google`), `tcp://host`, `tls://host` para DNS sobre TLS ou `https://host/dns-query` para DNS sobre HTTPS; DNS sobre QUIC não é suportado e é recusado ao salvar.

*Desligado* esconde as regras e as mantém: nada as leria de qualquer forma, e apagá-las faria com que religar o modousse uma configuração diferente.

Salvar grava as regras e o modo em duas requisições, e a caixa de diálogo informa qual delas o serviço aceitou. Uma recusa depois de as regras terem sido gravadas é dita como tal, para que a nova tentativa envie apenas o que falta.


**DNS para resolução de nós**, ao lado das configurações de DNS, seleciona o resolvedor global dos nomes de servidores dos nós de proxy. Aplica-se a novas conexões e testes de latência TCP/HTTP nos modos Somente serviço e Serviço e interceptação, independentemente da interceptação. O modo Desligado desliga o módulo e usa o caminho existente de resolução do sistema para os nós. Nesse modo, a lista de DNS dos nós e o botão de atualização ficam desativados sem apagar a seleção; voltam a ficar disponíveis após salvar um modo ativo. Nós com endereço IP não precisam de consulta DNS. Nomes de sites, servidores de assinatura e bootstrap dos upstreams de DNS mantêm seus mecanismos existentes.

Nesses modos ativos, as conexões dos nós e os testes de latência TCP/HTTP aguardam o módulo DNS interno começar a atender consultas e consultam seu listener local. O módulo consulta diretamente o upstream escolhido ou retorna o resultado em cache; a resolução dos nós não usa o resolvedor do sistema. O limite de tempo de inicialização, o cancelamento ou a saída do core interrompem a consulta. Uma falha do upstream é informada como erro de consulta DNS, sem recorrer a outro resolvedor. Com o core parado, os testes de latência iniciam um core temporário com a configuração DNS selecionada e o encerram após o teste.

A lista contém endpoints UDP, TCP, DoT e DoH com host IP das regras DNS, da configuração original do sistema e da lista alternativa interna. No Linux, o DNS do sistema é lido de `/etc/resolv.conf` ou do backup original durante o sequestro de DNS. As categorias indicam a origem: **grupo direto**, **DNS local**, **grupo de proxy**, **DNS alternativo**. Todas as consultas DNS dos nós são diretas, mesmo quando a origem é um grupo de proxy. Certificados DoT/DoH são verificados. As opções indicam origens válidas, não conectividade já testada.

**automático** escolhe o primeiro endpoint do primeiro grupo não vazio, na ordem direto → local → alternativo, após remover origens duplicadas. O rótulo mostra a URL completa, incluindo protocolo e porta, como um endpoint explícito; por exemplo, `udp://223.5.5.5:53 (automático)`. Uma consulta que falha retorna erro DNS sem mudar de endpoint. Atualizar e salvar configurações ou regras DNS recarregam as origens. Uma origem salva que desaparece permanece no formulário com aviso; alterações que a invalidariam são recusadas. Escolha outro endpoint ou automático primeiro. Cache e pré-busca de DNS continuam ativos; resultados novos afetam conexões futuras, enquanto conexões existentes mantêm seu IP.

## Compartilhamento com a rede local

**Compartilhamento de portas** faz as entradas SOCKS, HTTP e personalizadas escutarem em todas as interfaces em vez de `127.0.0.1`, para que outros dispositivos possam usar esta máquina como proxy; a porta da API permanece no loopback. Para rotear o tráfego desses dispositivos também pelo proxy transparente, use `redirect`, `tproxy` ou `tun`, ative **Encaminhamento de IP**, mantenha a interface pela qual o tráfego chega fora dos prefixos excluídos e direcione o gateway padrão dos dispositivos para esta máquina; os modos de proxy do sistema não podem fazer isso.
