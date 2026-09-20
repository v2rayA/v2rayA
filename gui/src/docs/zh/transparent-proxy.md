# 透明代理

开启透明代理后，所有应用的流量都经过内核，应用本身不需要任何设置。**设置 → 透明代理/系统代理**负责开启并选择分流策略，下面一项选择实现方式。

## 分流策略

| 策略 | 走代理的流量 |
| --- | --- |
| 不分流 | 全部 |
| 大陆白名单 | 除国内域名与地址（`geosite:cn`、`geoip:cn`）和私有地址以外的全部 |
| 仅 GFWList | GFWList 上的域名；需要先下载一次列表（**更新GFWList**） |
| 与规则端口相同 | 规则端口的模式，包括 RoutingA |

## 实现方式

| 实现 | 平台 | 说明 |
| --- | --- | --- |
| `redirect` | Linux | iptables/nftables `REDIRECT`；只有 TCP；DNS 拦截需要本机 53 端口空闲；可处理宿主机上 Docker 容器的流量 |
| `tproxy` | Linux | iptables/nftables `TPROXY`；TCP 与 UDP；不处理 Docker 容器的流量 |
| `tun` | Linux、Windows、macOS | 内核打开 TUN 设备；TCP 与 UDP；自动排除 v2rayA 与内核自身 |
| 系统代理 | Windows 始终可用；Linux 与 macOS 在 `--lite` 模式下可用 | 修改桌面的代理设置（Linux 上为 GNOME 与 KDE）；只覆盖遵守该设置的应用 |

`redirect` 与 `tproxy` 需要 root 和 `iptables` 或 `nftables`；`tun` 在 Linux 上需要 `/dev/net/tun`，在 Windows 与 macOS 上需要管理员权限。

**排除的网卡名前缀**让指定网卡（Docker 网桥、VPN 隧道）的流量不经过 `redirect` 与 `tproxy`。

## TUN

内核创建 TUN 设备并配置地址。**自动路由**开启时由 v2rayA 安装路由并把系统解析器指向内核；关闭时由你自己的脚本完成（见参数一节的钩子）。

不会进入 TUN 的流量：内核自己的连接、`direct` 出站与 DNS 模块的上游查询（Linux 通过套接字标记，Windows 与 macOS 通过绑定物理网卡）。v2rayA 与内核始终被排除；**TUN 自定义排除进程**按可执行文件名排除更多进程，每行一个。比默认路由更精确的路由（直连网段、静态路由）同样不经过 TUN。

到达 TUN 的 53 端口明文 DNS 由内核的 DNS 模块按 DNS 规则回答；加密 DNS 不被拦截。Windows 上系统解析器指向 TUN 网关，macOS 上指向内核在 `127.0.0.1` 的监听，后者要求 53 端口空闲。

已知限制：Windows 与 macOS 上直接向局域网 DNS 查询的应用仍会绕过 TUN。

## DNS

**设置 → DNS** 保存内核 DNS 模块遵循的规则：哪个上游回答哪些域名，查询经过哪个出站。默认规则把私有域名交给系统解析器，`geosite:cn` 直连查询 `223.5.5.5`，其余经代理查询 `1.0.0.1`。域名列表为空的规则是兜底。

## 局域网共享

**允许局域网的连接**让 SOCKS、HTTP 与自定义入站监听所有网卡而不是 `127.0.0.1`，其他设备就能把这台机器当代理。要让它们的流量也经过透明代理，打开 **开启IP转发**并把它们的默认网关指向这台机器。
