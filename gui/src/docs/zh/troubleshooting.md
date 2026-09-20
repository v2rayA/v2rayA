# 故障排除

先看服务日志：服务单元下是 `/var/log/v2raya/v2raya.log`，单元自身的输出用 `journalctl -u v2raya` 查看，其他情况看 `--log-file` 指定的路径。**日志**页显示同一份内容。`--log-level debug` 会额外记录内核配置和每个请求。

## 内核启动失败

`CORE_START_FAILED` 表示 `v2raya_core` 启动后立即退出，日志里它上面一行说明原因。常见原因：

- **内核版本不匹配**：页面顶部的横幅。`v2raya_core` 必须与 `v2raya` 版本相同；放在它位置上的 `xray` 或 `v2ray` 不被接受。从同一个发布安装两者。
- **`illegal domain rule: geosite:...`**：规则用了 `geosite.dat` 里没有的类别。可用的名称见分流规则一节。
- **address already in use**：入站端口或 53 端口（`redirect` 与 macOS 的 TUN 解析器需要）被占用。在地址与端口里改端口，或停止占用的程序。
- **`v2raya_core executable not found`**：内核不在 `v2raya` 同目录也不在 `PATH` 中；用 `--v2ray-bin` 指定。

## 无法连接

- 仪表板显示分组及成员延迟。成员显示 `timeout` 表示本机无法连通它；换一个节点或检查订阅。
- 直接测试一个入站：`curl -x socks5h://127.0.0.1:20170 https://example.com`。入站可用而透明代理不可用，问题在透明代理的配置而不是节点。
- 使用 `redirect` 或 `tproxy` 时，Docker 或防火墙可能改写了 iptables 规则；停止再启动内核可重新安装。
- Windows 与 macOS 上直接向局域网 DNS 查询的应用会绕过 TUN。

## 规则数据

首次启动时若没有 `geoip.dat` 与 `geosite.dat`，服务从 GitHub 下载，下载失败则退出。可手动把两个文件复制到 `/usr/share/v2raya`，或安装自带它们的软件包。

## 升级之后

2.4 之前的数据库（BoltDB）在首次启动时迁移；旧文件保留为 `bolt.db.bak`，账户需要重新注册。升级前停止服务并复制配置目录，迁移就能从副本重做。

## 重置密码

停止服务，用服务所用的 `--config` 与 `--lite` 参数执行 `v2raya --reset-password`，再启动服务并注册。
