# 分流规则

RoutingA 是 v2rayA 的规则语言。**设置 → RoutingA** 打开编辑器；规则作用于模式为 RoutingA 的规则端口（`20172`）、策略为“与规则端口相同”的透明代理，以及绑定到 RoutingA 的自定义入站。

## 语法

每行一条规则：若干条件以 `&&` 连接，再接 `->` 和出口。规则自上而下匹配，首条匹配的规则生效；`default:` 指定未匹配流量的出口。`#` 开始注释。

```
default: proxy
domain(geosite:category-ads-all) -> block
domain(geosite:cn) -> direct
ip(geoip:private, geoip:cn) -> direct
```

| 条件 | 匹配对象 |
| --- | --- |
| `domain(...)` | `full:` 完全匹配，`domain:` 域名及其子域名（默认），`contains:`，`regexp:`，`geosite:<类别>`，`ext:"文件.dat:标签"` |
| `ip(...)` | 地址、CIDR、`geoip:<代码>`、`ext:"文件.dat:标签"` |
| `port(...)` | 目标端口与范围，如 `port(80, 443, 1000-2000)` |
| `sourcePort(...)` | 源端口 |
| `network(...)` | `tcp`、`udp` |
| `protocol(...)` | 探测到的协议：`http`、`tls`、`bittorrent`（需要开启**嗅探**） |
| `source(...)` | 源地址与 CIDR |
| `inboundTag(...)` | 连接进入的入站 |

## 出口

`proxy`、`direct`、`block` 内置。其他代理分组在有已连接成员后即可按名称作为出口，否则规则无法应用。v2rayA 之外的 SOCKS 或 HTTP 服务器声明一次后像分组一样使用：

```
outbound: office = socks(address: 192.168.1.10, port: 1080)
outbound: gateway = http(address: 10.0.0.1, port: 8080, user: 'name', pass: 'secret')
domain(domain: corp.example) -> office
```

## 类别

`geosite:` 类别来自 v2fly 的 [domain-list-community](https://github.com/v2fly/domain-list-community)，`geoip:` 代码来自 [v2fly/geoip](https://github.com/v2fly/geoip)；v2rayA 首次启动时下载这两个文件。常用的有 `geosite:cn`、`geosite:geolocation-!cn`、`geosite:private`、`geosite:category-ads-all`、`geosite:netflix`、`geosite:youtube`、`geosite:telegram`、`geosite:openai`；`geoip:cn`、`geoip:private` 以及任意国家代码。属性可缩小类别范围：`geosite:apple@cn`、`geosite:category-games@cn`。

只存在于 Loyalsoldier 构建中的名称（`gfw`、`greatfire`、`apple-cn`、`geoip:telegram`）不在这两个文件里，内核会以 `illegal domain rule` 拒绝。GFWList 模式会把 Loyalsoldier 的文件下载为 `LoyalsoldierSite.dat`，之后规则可以写成 `domain(ext:"LoyalsoldierSite.dat:gfw")`。

## 编辑器

列表模式用表单编辑每条规则；文本模式提供行号、着色和逐行检查。编辑器旁的语法参考可把示例插入到光标处，其中**模板**一节是常见策略的完整规则集：插入到自己的规则之前，或者用它替换整个规则集。规则可导入导出为文本文件。

保存后内核以新规则重启；内核拒绝规则时保留原规则，错误信息指出出错的行。
