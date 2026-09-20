# Regras de roteamento

RoutingA é a linguagem de regras do v2rayA. **Configurações → RoutingA** abre o editor; as regras se aplicam à porta de regras (`20172`) quando seu modo é RoutingA e ao proxy transparente quando sua política é _Igual ao da porta de regras_. Uma entrada personalizada vinculada ao RoutingA tem seu próprio texto de regras, veja abaixo.

## Sintaxe

Uma regra por linha: condições, unidas por `&&`, seguidas de `->` e uma saída. As regras são avaliadas de cima para baixo e a primeira correspondência prevalece; `default:` indica a saída para tudo que não corresponder a nenhuma regra. Uma linha que começa com `#` é um comentário. O v2rayA coloca algumas regras próprias antes das demais: os endereços dos servidores proxy e o serviço de notificações push da Apple usam conexão direta, e uma condição `inboundTag(...)` substitui o escopo de entradas habitual da regra.

```
default: proxy
domain(geosite:category-ads-all) -> block
domain(geosite:cn) -> direct
ip(geoip:private, geoip:cn) -> direct
```

| Condição          | Corresponde a                                                                                                                                                            |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `domain(...)`     | `domain:` o domínio e seus subdomínios, `full:` correspondência exata, `regexp:`, `geosite:<category>`, `ext:"file.dat:tag"`; um valor sem prefixo, como `example.com`, corresponde a qualquer domínio que contenha esse texto |
| `ip(...)`         | um endereço, um CIDR (coloque IPv6 entre aspas: `ip("2001:db8::/32")`), `geoip:<code>`, `ext:"file.dat:tag"` |
| `port(...)`       | portas e intervalos de destino, `port(80, 443, 1000-2000)` |
| `sourcePort(...)` | portas de origem |
| `network(...)`    | `tcp`, `udp` |
| `protocol(...)`   | protocolo identificado pela inspeção: `http`, `tls`, `quic`, `bittorrent` (exige **Inspeção de tráfego** ativada) |
| `source(...)`     | endereços e CIDRs de origem |
| `inboundTag(...)` | a entrada pela qual a conexão chegou |

## Saídas

`proxy`, `direct` e `block` são integradas. Qualquer outro grupo de proxy é uma saída com seu próprio nome assim que tem um membro conectado; caso contrário, as regras não podem ser aplicadas. Um servidor SOCKS ou HTTP externo ao v2rayA é declarado uma vez e usado como um grupo (um arquivo `ext:` fica no diretório de dados de regras):

```
outbound: office = socks(address: 192.168.1.10, port: 1080)
outbound: gateway = http(address: 10.0.0.1, port: 8080, user: 'name', pass: 'secret')
domain(domain: corp.example) -> office
```

## Categorias

As categorias `geosite:` vêm do [domain-list-community](https://github.com/v2fly/domain-list-community) do v2fly, e os códigos `geoip:`, do [v2fly/geoip](https://github.com/v2fly/geoip); o v2rayA baixa os arquivos que faltam na primeira inicialização, mas um pacote pode fornecer sua própria compilação. Usados com frequência: `geosite:cn`, `geosite:geolocation-!cn`, `geosite:private`, `geosite:category-ads-all`, `geosite:greatfire`, `geosite:netflix`, `geosite:youtube`, `geosite:telegram`, `geosite:openai`; `geoip:cn`, `geoip:private` e os códigos de países. Um atributo restringe uma categoria: `geosite:apple@cn`, `geosite:category-games@cn`.

Nomes que existem apenas nas compilações de Loyalsoldier (`gfw`, `apple-cn`, `google-cn`, `geoip:telegram`) não estão nesses arquivos; o núcleo os recusa com `illegal domain rule` ou `illegal ip rule`. **Atualizar GFWList** baixa essa compilação (de `v2rayA/dist-v2ray-rules-dat`) como `LoyalsoldierSite.dat`, e as regras podem então usá-la como `domain(ext:"LoyalsoldierSite.dat:gfw")`.

## O editor

O modo de lista edita cada regra com um formulário; o modo de texto mostra números de linha, realce de sintaxe e uma verificação por linha. A referência ao lado do editor insere exemplos na posição do cursor, e sua seção **Modelos** contém conjuntos completos de regras para as políticas comuns: insira um antes de suas próprias regras ou substitua todo o conjunto de regras por ele. As regras são importadas e exportadas como um arquivo de texto.

Salvar recarrega um núcleo em execução com as novas regras; quando o núcleo as rejeita, o v2rayA restaura as regras anteriores e mostra o erro do núcleo, que indica o texto problemático. Um núcleo parado adota as regras na próxima inicialização.

## Entradas personalizadas

Uma entrada personalizada vinculada ao RoutingA tem seu próprio texto de regras, que se aplica apenas a essa entrada. Ela aceita regras, não definições: linhas `default:` e `outbound:` são ignoradas, e o tráfego que não corresponde a nenhuma regra vai para o grupo escolhido no formulário. Termine o texto com uma regra que abranja tudo, como `network(tcp, udp) -> direct`, para definir por conta própria o roteamento de todo o restante.
