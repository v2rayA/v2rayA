# 故障排除

先看服务日志：Linux 服务单元下是 `/var/log/v2raya/v2raya.log`，单元自身的输出用 `journalctl -u v2raya` 查看，其他情况看 `--log-file` 指定的路径。**日志**页显示同一份内容；**日志等级**设置提高内核的输出详细程度。生成的内核配置是配置目录下的 `config.json`。

## 内核启动失败

`CORE_START_FAILED` 表示 `v2raya_core` 无法启动、启动后立即退出，或在 `--core-startup-timeout` 内没有打开 API 端口，日志里它前面几行说明原因。常见原因：

- **内核版本不匹配**：页面顶部的横幅。`v2raya_core` 必须与 `v2raya` 版本相同；放在它位置上的 `xray` 或 `v2ray` 不被接受。安装同一版本发布中的 `v2raya` 与 `v2raya_core`。
- **`illegal domain rule: geosite:...`**：规则用了 `geosite.dat` 里没有的类别。可用的名称见分流规则一节。
- **address already in use**：入站端口被占用。在地址与端口里改端口，或停止占用的程序。53 端口不同：被占用时 DNS 模块跳过那个监听并写入日志，macOS 上由原有解析器继续回答。
- **`v2raya_core executable not found`**：内核不在 `v2raya` 同目录也不在 `PATH` 中；用 `--v2ray-bin` 指定。

## 无法连接

- 仪表板显示分组及成员延迟。`TIMEOUT` 表示本机到该服务器的 TCP 连接失败；换一个节点或检查订阅。
- 直接测试一个入站：`curl -x socks5h://127.0.0.1:20170 https://example.com`。入站可用而透明代理不可用时，先查透明代理的配置、DNS 规则和透明代理所走的分组，再怀疑节点。
- 使用 `redirect` 或 `tproxy` 时，Docker 或防火墙可能改写了 iptables 规则；停止再启动内核可重新安装。
- Windows 与 macOS 上直接向局域网 DNS 查询的应用会绕过 TUN。

## 规则数据

找到内核后，启动时若缺少 `geoip.dat` 或 `geosite.dat`，服务从 GitHub 下载缺少的文件，下载失败则退出。可手动把文件复制到规则数据目录（见参数一节），或安装自带它们的软件包。

## 升级之后

旧版本的 BoltDB 数据库在首次启动时迁移；旧文件保留为 `bolt.db.bak`（该名称已存在时为 `bolt.db.bak.1` 及后续编号），账户需要重新注册。升级前停止服务并复制配置目录，迁移就能从副本重做。

## 重置密码

停止服务，以服务所用的账户和同样的 `--config` 与 `--lite` 参数执行 `v2raya --reset-password`（Windows 上服务的目录在 `%ProgramData%\SYSTEM` 下，不是提权用户的目录），再启动服务并注册。
