# v2rayA 项目进度

## 已完成的工作

### 记忆库初始化 (2026-01-04)
- [x] 创建 projectbrief.md - 项目概述
- [x] 创建 productContext.md - 产品上下文
- [x] 创建 systemPatterns.md - 系统架构模式
- [x] 创建 techContext.md - 技术上下文
- [x] 创建 activeContext.md - 当前活动上下文
- [x] 创建 progress.md - 进度追踪

### 项目探索完成
- [x] 阅读 README.md 了解项目概况
- [x] 分析后端入口 (main.go, pre.go)
- [x] 探索 API 路由结构 (server/router/)
- [x] 了解核心功能模块 (core/)
- [x] 分析前端结构 (gui/, ngui/)
- [x] 理解构建和配置系统

### httpupgrade 传输类型支持修复 (2026-01-04)
- [x] 识别问题：v2rayA 不支持 httpupgrade 传输类型，报错 "unexpected transport type: httpupgrade"
- [x] 添加 HTTPUpgradeSettings 结构体到 coreObj.go
- [x] 在 StreamSettings 中添加 HTTPUpgradeSettings 字段
- [x] 在 v2ray.go Configuration() 方法中添加 httpupgrade case
- [x] 在 v2ray.go ExportToURL() 方法中添加 httpupgrade case
- [x] 构建 GUI 前端并嵌入后端
- [x] 编译 Go 后端 (版本: unstable-httpupgrade)
- [x] 部署到系统 (/usr/bin/v2raya)
- [x] 验证节点导入功能正常

**修改的文件：**
1. `service/core/coreObj/coreObj.go`
   - 添加 HTTPUpgradeSettings 结构体
   - 在 StreamSettings 中添加对应字段

2. `service/core/serverObj/v2ray.go`
   - Configuration() 方法添加 httpupgrade case
   - ExportToURL() 方法添加 httpupgrade case

**代码片段：**
```go
// coreObj.go
type HTTPUpgradeSettings struct {
    Path string `json:"path,omitempty"`
    Host string `json:"host,omitempty"`
}

// StreamSettings 中
HTTPUpgradeSettings *HTTPUpgradeSettings `json:"httpupgradeSettings,omitempty"`

// v2ray.go Configuration()
case "httpupgrade":
    core.StreamSettings.HTTPUpgradeSettings = &coreObj.HTTPUpgradeSettings{
        Path: v.Path,
        Host: v.Host,
    }

// v2ray.go ExportToURL()
case "httpupgrade":
    setValue(&query, "path", v.Path)
    setValue(&query, "host", v.Host)
```

## 当前进行中
- 无活跃开发任务

## 待开始的工作
- 无计划任务

## 项目状态概览

### 模块成熟度
| 模块 | 状态 | 说明 |
|------|------|------|
| 后端核心 | 稳定 | Go 服务运行良好 |
| 旧前端 (gui) | 稳定 | Vue 2 生产可用 |
| 新前端 (ngui) | 开发中 | Nuxt 3 重构 |
| 透明代理 | 稳定 | iptables 支持完善 |
| nftables | 实验性 | 可选启用 |
| TUN 模式 | 稳定 | 需要 gvisor |

### 协议/传输支持状态
| 协议 | 状态 |
|------|------|
| VMess | ✓ 完整支持 |
| VLESS | ✓ 完整支持 |
| Shadowsocks | ✓ 完整支持 |
| ShadowsocksR | ✓ 完整支持 |
| Trojan | ✓ 完整支持 |
| TUIC | ✓ 支持 |
| Juicity | ✓ 支持 |
| HTTP | ✓ 支持 |
| SOCKS5 | ✓ 支持 |

| 传输类型 | 状态 |
|----------|------|
| tcp | ✓ 完整支持 |
| ws (websocket) | ✓ 完整支持 |
| grpc | ✓ 完整支持 |
| kcp/mkcp | ✓ 完整支持 |
| h2/http | ✓ 完整支持 |
| quic | ✓ 完整支持 |
| xhttp | ✓ 完整支持 |
| **httpupgrade** | ✓ **新增支持 (2026-01-04)** |

## 已知问题
1. ngui 新前端功能尚未完全对等 gui
2. 部分文档需要更新

## 里程碑

### v2.0 (已发布)
- Web GUI 界面
- 多协议支持
- 透明代理
- 订阅管理

### 后续版本规划
- 新前端完成并替换旧版
- nftables 默认支持
- 更多协议扩展

## 变更日志

### 2026-01-04
- 初始化项目记忆库
- 完成项目结构文档化
- **修复 httpupgrade 传输类型支持**
  - 添加 HTTPUpgradeSettings 结构体和相关配置逻辑
  - 成功编译部署版本 unstable-httpupgrade
  - 验证节点导入功能正常
