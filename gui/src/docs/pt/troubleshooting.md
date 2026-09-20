# Solução de problemas

Consulte primeiro o registro do serviço: `/var/log/v2raya/v2raya.log` quando executado pela unidade de serviço do Linux, `journalctl -u v2raya` para o que a própria unidade imprime e, nos demais casos, o caminho definido por `--log-file`. A página **Registros** mostra o mesmo fluxo; a configuração **Nível de registro** aumenta o detalhamento do núcleo. A configuração gerada do núcleo é `config.json`, no diretório de configuração.

## O núcleo não inicia

`CORE_START_FAILED` significa que não foi possível iniciar o `v2raya_core`, que ele encerrou logo após iniciar ou que não abriu sua porta de API dentro de `--core-startup-timeout`; as linhas anteriores no registro dizem o motivo. Causas comuns:

- **Incompatibilidade de versão do núcleo** — o aviso no topo da página. O `v2raya_core` deve ter a mesma versão do `v2raya`; um binário `xray` ou `v2ray` em seu lugar não é aceito. Instale ambos da mesma versão publicada.
- **`illegal domain rule: geosite:...`** — uma regra indica uma categoria que não está em `geosite.dat`. Veja na seção de roteamento quais nomes existem.
- **address already in use** — uma porta de entrada está ocupada. Altere-a em Endereços e portas ou pare o outro programa. A porta 53 é diferente: quando está ocupada, o módulo DNS não abre seu serviço de escuta nessa porta e informa isso no registro; no macOS, o resolvedor existente continua respondendo.
- **`v2raya_core executable not found`** — o núcleo não está ao lado de `v2raya` nem em `PATH`; passe `--v2ray-bin`.

## Nada se conecta

- O painel mostra o grupo e a latência de seus membros. `TIMEOUT` significa que a conexão TCP desta máquina com o servidor falhou; tente outro nó ou verifique a assinatura.
- Teste uma entrada diretamente: `curl -x socks5h://127.0.0.1:20170 https://example.com`. Se funcionar, mas o proxy transparente não, verifique a configuração do proxy transparente, as regras de DNS e o grupo para o qual o proxy transparente roteia antes de atribuir o problema ao nó.
- Com `redirect` ou `tproxy`, o Docker ou um firewall pode ter substituído as regras do iptables; pare e inicie o núcleo para reinstalá-las.
- No Windows e no macOS, um aplicativo que consulta DNS diretamente em um resolvedor da rede local não passa pelo TUN.

## Dados de regras

Ao iniciar, depois de localizar o núcleo, o v2rayA baixa do GitHub o arquivo `geoip.dat` ou `geosite.dat` que estiver ausente e encerra sua execução se o download falhar. Copie os arquivos manualmente para o diretório de dados de regras (veja a seção de parâmetros) ou instale um pacote que os inclua.

## Após uma atualização

Um banco de dados BoltDB de uma versão anterior é migrado na primeira inicialização; o arquivo antigo permanece como `bolt.db.bak` (`bolt.db.bak.1` e seguintes quando esse nome já está em uso), e as contas precisam ser registradas novamente. Pare o serviço e copie o diretório de configuração antes de atualizar; assim, a migração poderá ser repetida a partir dessa cópia.

## Redefinir a senha

Pare o serviço, execute `v2raya --reset-password` com a conta usada pelo serviço e com as mesmas opções `--config` e `--lite` (no Windows, o diretório do serviço fica em `%ProgramData%\SYSTEM`, não no diretório do usuário com privilégios elevados), inicie-o novamente e registre-se.
