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

Na página **Proxies**, **Importar** aceita links de compartilhamento ou uma assinatura: escolha **Link do servidor** para links `vmess://`, `vless://`, `ss://`, `trojan://`, `hysteria2://`, `tuic://`, `juicity://`, `anytls://`, `wireguard://`, `socks5://`, `http://` e `https://`, um por linha, ou uma imagem de código QR; escolha **Endereço da assinatura** para uma assinatura. Links ShadowsocksR são recusados, assim como links Shadowsocks com cifra de fluxo (`rc4-md5`, `aes-*-cfb`, `chacha20-ietf`) ou `none`: o núcleo aceita apenas as cifras AEAD e os métodos 2022-blake3. `allow_insecure` em um link é ignorado: o v2rayA nunca pula a verificação do certificado; para um servidor autoassinado, fixe o SHA-256 do certificado no formulário do nó.

As assinaturas têm sua própria página, **Assinaturas** (no celular, a documentação passa para o menu da barra superior para abrir espaço). Elas mantêm seus nós agrupados e podem ser atualizadas manualmente ou por agendamento (**Configurações → Atualizar assinaturas automaticamente**). A configuração de modo ao lado determina se a atualização passa pelo proxy.

## Grupos

Os nós são usados por meio de grupos de proxy. O grupo `proxy` sempre existe; outros são criados, configurados e excluídos no botão de grupos da barra superior, o único lugar para isso. Um nó entra em um grupo pelo seu menu, ou selecione vários na página Proxies e escolha o grupo em **Adicionar ao grupo de proxy**. Um grupo com vários membros conectados usa o de menor latência medida (**Automático (menor latência)**); o menu do nó no painel fixa um membro para que o grupo use apenas ele, **Adicionar ou remover nós** altera os membros, e as configurações do grupo definem o endereço e o intervalo de sondagem.

As regras de roteamento usam os nomes dos grupos como saídas: `proxy` por padrão, e qualquer outro grupo pelo próprio nome assim que tiver um membro conectado.

## Iniciar o núcleo

**Iniciar** no painel, ou o botão de estado no topo da página, inicia o `v2raya_core` com a configuração atual. As alterações na página Configurações entram em vigor com **Salvar e aplicar**: um núcleo em execução é reiniciado com elas, e um núcleo parado as adota na próxima inicialização.

Quando o núcleo está em execução, os aplicativos o acessam pelas entradas locais:

| Entrada         | Endereço          | Roteamento                                                       |
| --------------- | ----------------- | ---------------------------------------------------------------- |
| SOCKS5          | `127.0.0.1:20170` | tudo pelo `proxy`                                                |
| HTTP            | `127.0.0.1:20171` | tudo pelo `proxy`                                                |
| HTTP com regras | `127.0.0.1:20172` | a configuração **Modo de divisão de tráfego da porta de regras** |

As portas são alteradas em **Configurações → Endereços e portas**; 0 fecha uma entrada. Para atender aos aplicativos sem configurar cada um, ative o proxy transparente.

## Roteamento

**Configurações → Divisão de tráfego** define como a porta de regras e o proxy transparente dividem o tráfego: usar o proxy para tudo, exceto sites chineses, usar o proxy apenas para a GFWList ou usar suas próprias regras RoutingA. O RoutingA é descrito em sua própria seção.

## Teclado

- Página Proxies, visão em lista: `Ctrl`+`A` seleciona todos os nós listados e `Esc` limpa a seleção; `/` vai para a busca nas páginas Proxies e Registros.
- Configurações: `Ctrl`+`S` salva. Editor RoutingA: `Ctrl`+`S` salva, `Tab` indenta.
- Diálogos com área de texto (importar, assinatura, scripts de rota, listas): `Ctrl`+`Enter` envia; `Esc` fecha qualquer diálogo.
- Registros: com foco, `Home`, `End`, `PgUp` e `PgDn` percorrem o registro.

No macOS, `Cmd` corresponde a `Ctrl`.

## Atualização automática de assinaturas

Cada assinatura tem seu próprio modo de atualização:

- **Desativada:** atualizar somente quando solicitado manualmente.
- **Ao iniciar o serviço:** atualizar uma vez sempre que o v2rayA iniciar.
- **Em um intervalo:** atualizar ao iniciar e depois do número de minutos configurado.
- **Em um intervalo com recuperação de falha:** seguir a programação regular e também verificar os servidores salvos no intervalo de falha. Quando nenhum funcionar, atualizar a assinatura nesse intervalo até que pelo menos um fique disponível.

Downloads com erro ou vazios preservam a lista de servidores salva. Uma tentativa de recuperação não adia a programação regular, e uma execução demorada nunca ocorre em paralelo com a próxima. No Linux, os downloads de recuperação usam conexões marcadas para que o proxy transparente não os encaminhe de volta ao proxy com falha. Esse desvio não é garantido no modo TUN do macOS ou Windows. As verificações podem iniciar processos temporários do núcleo, mas a atualização automática não inicia o núcleo principal que o usuário parou manualmente.

Na atualização, o modo global antigo **ao iniciar** é atribuído a todas as assinaturas existentes como **Ao iniciar o serviço**. O modo antigo por intervalo passa a **Em um intervalo**, com as horas convertidas em minutos. A recuperação de falha nunca é ativada pela migração; selecione-a manualmente onde for necessária.
