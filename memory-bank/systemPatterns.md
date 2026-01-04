# v2rayA 系统架构模式

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        用户浏览器                            │
│                    (Web GUI / ngui)                         │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP/WebSocket (Port 2017)
┌─────────────────────▼───────────────────────────────────────┐
│                    v2rayA Service                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Gin Web Framework                       │   │
│  │  ┌─────────────────┐  ┌─────────────────────────┐   │   │
│  │  │   Router (API)  │  │   Static File Server   │   │   │
│  │  │  /api/*         │  │   (Embedded Web GUI)   │   │   │
│  │  └─────────────────┘  └─────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    Core Layer                        │   │
│  │  ┌───────────┐ ┌───────────┐ ┌─────────────────┐    │   │
│  │  │ ServerObj │ │   v2ray   │ │   iptables/     │    │   │
│  │  │  (协议)   │ │ (进程管理) │ │   nftables      │    │   │
│  │  └───────────┘ └───────────┘ └─────────────────┘    │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   Database Layer                     │   │
│  │                    (BoltDB)                          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  v2ray-core / xray-core                     │
│              (实际的代理核心进程)                            │
└─────────────────────────────────────────────────────────────┘
```

## 目录结构设计

```
v2rayA/
├── service/              # Go 后端服务
│   ├── main.go           # 程序入口
│   ├── pre.go            # 初始化逻辑
│   ├── conf/             # 配置管理
│   ├── core/             # 核心功能模块
│   │   ├── serverObj/    # 服务器对象（协议抽象）
│   │   ├── v2ray/        # v2ray-core 管理
│   │   ├── iptables/     # 防火墙规则管理
│   │   ├── dns/          # DNS 配置
│   │   ├── tun/          # TUN 设备管理
│   │   └── specialMode/  # 特殊模式（FakeDNS 等）
│   ├── db/               # 数据库层
│   │   └── configure/    # 配置数据模型
│   ├── server/           # HTTP 服务层
│   │   ├── router/       # 路由定义
│   │   ├── controller/   # 控制器
│   │   └── service/      # 业务逻辑
│   └── pkg/              # 工具包
│       ├── plugin/       # 协议插件
│       ├── server/       # 服务器工具
│       └── util/         # 通用工具
├── gui/                  # Vue 2 前端（旧版）
│   ├── src/
│   │   ├── components/   # Vue 组件
│   │   ├── plugins/      # 插件
│   │   ├── locales/      # 国际化
│   │   └── store/        # Vuex 状态管理
│   └── public/           # 静态资源
├── ngui/                 # Nuxt 3 前端（新版）
│   ├── pages/            # 页面组件
│   ├── components/       # 共享组件
│   ├── composables/      # 组合式函数
│   ├── middleware/       # 中间件
│   └── locales/          # 国际化
└── install/              # 安装脚本和配置
```

## 核心设计模式

### 1. 服务器对象模式 (ServerObj)
使用接口抽象不同协议的服务器：
```go
type ServerObj interface {
    Configuration(info PriorInfo) (c Configuration, err error)
    ExportToURL() string
    NeedPluginPort() bool
    ProtoToShow() string
    GetProtocol() string
    GetHostname() string
    GetPort() int
    GetName() string
    SetName(name string)
}
```

支持的协议通过工厂模式注册：
- `FromLinkRegister()` - 从链接创建
- `EmptyRegister()` - 创建空对象

### 2. 进程管理模式
`ProcessManager` 管理 v2ray-core 进程的生命周期：
- 启动/停止核心进程
- 配置文件生成
- 透明代理规则管理
- 健康检查

### 3. 数据访问模式
使用 BoltDB 键值存储：
- `db/boltdb.go` - 基础 DB 操作
- `db/listOp.go` - 列表操作
- `db/setOp.go` - 集合操作
- `db/configure/` - 业务数据模型

### 4. 嵌入式 Web 资源
使用 Go 1.16+ embed 特性：
```go
//go:embed web
var webRoot embed.FS
```
构建时将前端文件嵌入二进制。

## API 设计

### 认证
- JWT Token 认证
- 首次启动需要创建账户

### 主要 API 端点
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/login | 用户登录 |
| POST | /api/account | 创建账户 |
| GET | /api/touch | 获取服务器和订阅列表 |
| POST | /api/import | 导入服务器/订阅 |
| POST | /api/connection | 连接服务器 |
| DELETE | /api/connection | 断开服务器 |
| POST | /api/v2ray | 启动代理核心 |
| DELETE | /api/v2ray | 停止代理核心 |
| GET | /api/setting | 获取设置 |
| PUT | /api/setting | 更新设置 |

## 透明代理实现

### Linux TProxy 模式
1. 创建路由表和规则
2. 配置 iptables/nftables TPROXY 规则
3. 重定向流量到 v2ray-core 监听端口

### Linux Redirect 模式
1. 配置 iptables REDIRECT 规则
2. 重定向 TCP 流量到本地端口

### Windows/macOS
使用系统 API 设置系统代理
