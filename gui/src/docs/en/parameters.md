# Flags and environment variables

Every flag has an environment variable of the same name: `--log-level` is `V2RAYA_LOG_LEVEL`. The Linux service unit sets `V2RAYA_LOG_FILE` and reads `/etc/default/v2raya`; the rest keep their defaults. The table below is read from the running service and shows the declared defaults, not the values in effect; an empty default is computed at start.

## Hooks

`--transparent-hook` names an executable that runs before and after the transparent proxy is set up and torn down, with `--transparent-type=<redirect|tproxy|tun|system_proxy>`, `--stage=<pre-start|post-start|pre-stop|post-stop>` and `--v2raya-confdir=<directory>` as arguments. `--core-hook` runs before and after the core starts and stops, with `--stage` and `--v2raya-confdir`. The route scripts for `tun` with **Auto Route** off are a separate setting, **Configure Route Script**.

## Directories

| What                                   | Where                                                                                                                                                                                                                                                         |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| configuration and `v2raya.db`          | `--config`; default `/etc/v2raya` on Linux and macOS, `%ProgramData%\SYSTEM\v2rayA` for the Windows service, `~/.config/v2raya` (or `%AppData%\v2rayA`) in lite mode                                                                                          |
| generated core config                  | `config.json` in the configuration directory                                                                                                                                                                                                                  |
| rule data (`geoip.dat`, `geosite.dat`) | `--v2ray-assetsdir`; otherwise the XDG data directories are searched (`/usr/share/v2raya`, `/usr/local/share/v2raya`, `~/.local/share/v2raya`) and a download lands in the user's; on Windows the installer's `data` directory or the configuration directory |
| the core                               | `--v2ray-bin`; default `v2raya_core` next to `v2raya` or in `PATH`                                                                                                                                                                                            |
| logs                                   | `--log-file` or `V2RAYA_LOG_FILE`, else the console; the Linux system unit uses `/var/log/v2raya/v2raya.log`, the user unit `~/.local/state/v2raya/v2raya.log`, the Windows installer `%TEMP%\v2raya.log`                                                     |
