# Opções e variáveis de ambiente

Cada opção tem uma variável de ambiente de mesmo nome: `--log-level` é `V2RAYA_LOG_LEVEL`. A unidade de serviço do Linux define `V2RAYA_LOG_FILE` e lê `/etc/default/v2raya`; as demais mantêm seus padrões. A tabela abaixo é lida do serviço em execução e mostra os valores padrão declarados, não os valores em uso; quando o valor padrão está vazio, ele é calculado na inicialização.

## Hooks

`--transparent-hook` indica um executável que é executado antes e depois da configuração e da remoção do proxy transparente, com `--transparent-type=<redirect|tproxy|tun|system_proxy>`, `--stage=<pre-start|post-start|pre-stop|post-stop>` e `--v2raya-confdir=<directory>` como argumentos. `--core-hook` é executado antes e depois de o núcleo iniciar e parar, com `--stage` e `--v2raya-confdir`. Os scripts de rota para `tun` com **Rota automática** desativada ficam em uma configuração separada, **Configurar script de rota**.

## Diretórios

| Conteúdo                               | Local                                                                                                                                                                                                                                                         |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| configuração e `v2raya.db`              | `--config`; padrão `/etc/v2raya` no Linux e no macOS, `%ProgramData%\SYSTEM\v2rayA` para o serviço do Windows, `~/.config/v2raya` (ou `%AppData%\v2rayA`) no modo lite |
| configuração gerada do núcleo          | `config.json` no diretório de configuração |
| dados de regras (`geoip.dat`, `geosite.dat`) | `--v2ray-assetsdir`; caso contrário, a busca usa os diretórios de dados XDG (`/usr/share/v2raya`, `/usr/local/share/v2raya`, `~/.local/share/v2raya`) e os arquivos baixados ficam no diretório de dados XDG do usuário; no Windows, o diretório `data` do instalador ou o diretório de configuração |
| o núcleo                               | `--v2ray-bin`; padrão `v2raya_core` ao lado de `v2raya` ou em `PATH` |
| registros                              | `--log-file` ou `V2RAYA_LOG_FILE`, caso contrário, o console; a unidade de sistema do Linux usa `/var/log/v2raya/v2raya.log`, a unidade de usuário usa `~/.local/state/v2raya/v2raya.log`, e o instalador do Windows usa `%TEMP%\v2raya.log` |
