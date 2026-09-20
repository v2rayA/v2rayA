# Flags and environment variables

Every flag has an environment variable of the same name: `--log-level` is `V2RAYA_LOG_LEVEL`. The service units set `V2RAYA_LOG_FILE`; the rest keep their defaults unless the unit or a `/etc/default/v2raya` file says otherwise. The table below is read from the running service.

## Hooks

`--transparent-hook` names an executable that runs around the transparent proxy: at `pre-start`, `post-start`, `pre-stop` and `post-stop`, with `--transparent-type=<redirect|tproxy|tun|system_proxy>`, `--stage=<stage>` and `--v2raya-confdir=<directory>` as arguments. With **Auto Route** off, this is where the routes for `tun` are installed and removed.

`--core-hook` runs the same way around the core process, with `--stage` only.

## Directories

| What | Where |
| --- | --- |
| configuration and `v2raya.db` | `--config`; default `/etc/v2raya` on Linux and macOS, `%ProgramData%\SYSTEM\v2rayA` for the Windows service, the user's config directory in lite mode |
| generated core config | `config.json` in the configuration directory |
| rule data (`geoip.dat`, `geosite.dat`) | `--v2ray-assetsdir`; default `/usr/share/v2raya` or `/usr/local/share/v2raya` |
| the core | `--v2ray-bin`; default `v2raya_core` next to `v2raya` or in `PATH` |
| logs | `--log-file` or `V2RAYA_LOG_FILE`; the service units use `/var/log/v2raya/v2raya.log` |
