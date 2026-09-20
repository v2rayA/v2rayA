# Início rápido

Esta página no navegador configura o v2rayA e mostra seu estado; o serviço é executado como root (como administrador no Windows; sem privilégios com `--lite`) e inicia o próprio `v2raya_core`.

## Entrar

A primeira conta registrada é a de administrador; não há outra conta. O registro pede um nome de usuário e uma senha de 6 a 32 caracteres e não exige credenciais existentes; por isso, em uma rede compartilhada, restrinja quem pode acessar a porta 2017 antes de se registrar, ou execute o serviço com `--address 127.0.0.1:2017` para uso apenas local.

Para redefinir uma senha esquecida, use a linha de comando com o serviço parado:

```sh
v2raya --reset-password
```

Execute o comando com a conta e o diretório `--config` usados pelo serviço (também com `--lite` no modo lite); no Windows, o serviço é executado como SYSTEM e seu diretório é diferente do diretório de um usuário com privilégios elevados. Inicie o serviço novamente e registre uma nova conta.

## Importar nós

Na página **Proxies**, **Importar** aceita links de compartilhamento ou uma assinatura: escolha **Link do servidor** para links `vmess://`, `vless://`, `ss://`, `trojan://`, `hysteria2://`, `tuic://`, `juicity://`, `anytls://`, `wireguard://`, `socks5://`, `http://` e `https://`, um por linha, ou uma imagem de código QR; escolha **Endereço da assinatura** para uma assinatura. Links ShadowsocksR são recusados.

As assinaturas mantêm seus nós agrupados e podem ser atualizadas manualmente ou por agendamento (**Configurações → Atualizar assinaturas automaticamente**). A configuração de modo ao lado determina se a atualização passa pelo proxy.

## Grupos

Os nós são usados por meio de grupos de proxy. O grupo `proxy` sempre existe; outros podem ser adicionados na página Proxies. Adicione um nó a um grupo pelo menu do nó ou selecionando vários nós. Um grupo com vários membros conectados usa o de menor latência medida (**Automático (menor latência)**); o menu do nó no painel fixa um membro para que o grupo use apenas ele, **Editar grupo** altera os membros, e as configurações do grupo definem o endereço e o intervalo de sondagem.

As regras de roteamento usam os nomes dos grupos como saídas: `proxy` por padrão, e qualquer outro grupo pelo próprio nome assim que tiver um membro conectado.

## Iniciar o núcleo

**Iniciar** no painel, ou o botão de estado no topo da página, inicia o `v2raya_core` com a configuração atual. As alterações na página Configurações entram em vigor com **Salvar e aplicar**: um núcleo em execução é reiniciado com elas, e um núcleo parado as adota na próxima inicialização.

Quando o núcleo está em execução, os aplicativos o acessam pelas entradas locais:

| Entrada         | Endereço          | Roteamento                                          |
| --------------- | ----------------- | --------------------------------------------------- |
| SOCKS5          | `127.0.0.1:20170` | tudo pelo `proxy`                                   |
| HTTP            | `127.0.0.1:20171` | tudo pelo `proxy`                                   |
| HTTP com regras | `127.0.0.1:20172` | a configuração **Modo de divisão de tráfego da porta de regras** |

As portas são alteradas em **Configurações → Endereços e portas**; 0 fecha uma entrada. Para atender aos aplicativos sem configurar cada um, ative o proxy transparente.

## Roteamento

**Configurações → Divisão de tráfego** define como a porta de regras e o proxy transparente dividem o tráfego: usar o proxy para tudo, exceto sites chineses, usar o proxy apenas para a GFWList ou usar suas próprias regras RoutingA. O RoutingA é descrito em sua própria seção.
