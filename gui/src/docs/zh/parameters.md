# 启动参数与环境变量

每个参数都有同名的环境变量：`--log-level` 对应 `V2RAYA_LOG_LEVEL`。服务单元设置了 `V2RAYA_LOG_FILE`；其余保持默认，除非单元或 `/etc/default/v2raya` 另有指定。下表从正在运行的服务读取。

## 钩子

`--transparent-hook` 指定一个可执行文件，在透明代理的 `pre-start`、`post-start`、`pre-stop`、`post-stop` 四个阶段执行，参数为 `--transparent-type=<redirect|tproxy|tun|system_proxy>`、`--stage=<阶段>` 与 `--v2raya-confdir=<目录>`。**自动路由**关闭时，`tun` 的路由就在这里安装和移除。

`--core-hook` 以同样方式围绕内核进程执行，只带 `--stage`。

## 目录

| 内容 | 位置 |
| --- | --- |
| 配置与 `v2raya.db` | `--config`；默认 Linux 与 macOS 为 `/etc/v2raya`，Windows 服务为 `%ProgramData%\SYSTEM\v2rayA`，lite 模式为用户配置目录 |
| 生成的内核配置 | 配置目录下的 `config.json` |
| 规则数据（`geoip.dat`、`geosite.dat`） | `--v2ray-assetsdir`；默认 `/usr/share/v2raya` 或 `/usr/local/share/v2raya` |
| 内核 | `--v2ray-bin`；默认为 `v2raya` 同目录或 `PATH` 中的 `v2raya_core` |
| 日志 | `--log-file` 或 `V2RAYA_LOG_FILE`；服务单元使用 `/var/log/v2raya/v2raya.log` |
