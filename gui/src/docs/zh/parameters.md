# 启动参数与环境变量

每个参数都有同名的环境变量：`--log-level` 对应 `V2RAYA_LOG_LEVEL`。Linux 的服务单元设置了 `V2RAYA_LOG_FILE` 并读取 `/etc/default/v2raya`；其余保持默认。下表从正在运行的服务读取，显示的是声明的默认值而不是当前生效值；默认值为空的参数在启动时计算。

## 钩子

`--transparent-hook` 指定一个可执行文件，在透明代理安装前后和卸载前后执行，参数为 `--transparent-type=<redirect|tproxy|tun|system_proxy>`、`--stage=<pre-start|post-start|pre-stop|post-stop>` 与 `--v2raya-confdir=<目录>`。`--core-hook` 在内核启动前后和停止前后执行，参数为 `--stage` 与 `--v2raya-confdir`。**自动路由**关闭时 `tun` 使用的路由脚本是另一项设置：**配置路由脚本**。

## 目录

| 内容                                   | 位置                                                                                                                                                                                                  |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 配置与 `v2raya.db`                     | `--config`；默认 Linux 与 macOS 为 `/etc/v2raya`，Windows 服务为 `%ProgramData%\SYSTEM\v2rayA`，lite 模式为 `~/.config/v2raya`（Windows 为 `%AppData%\v2rayA`）                                       |
| 生成的内核配置                         | 配置目录下的 `config.json`                                                                                                                                                                            |
| 规则数据（`geoip.dat`、`geosite.dat`） | `--v2ray-assetsdir`；未指定时按 XDG 数据目录查找（`/usr/share/v2raya`、`/usr/local/share/v2raya`、`~/.local/share/v2raya`），下载的文件放入用户的数据目录；Windows 为安装程序的 `data` 目录或配置目录 |
| 内核                                   | `--v2ray-bin`；默认为 `v2raya` 同目录或 `PATH` 中的 `v2raya_core`                                                                                                                                     |
| 日志                                   | `--log-file` 或 `V2RAYA_LOG_FILE`，未指定时输出到控制台；Linux 系统单元使用 `/var/log/v2raya/v2raya.log`，用户单元使用 `~/.local/state/v2raya/v2raya.log`，Windows 安装程序使用 `%TEMP%\v2raya.log`   |
