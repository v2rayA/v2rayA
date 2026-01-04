# v2rayA 技术上下文

## 开发环境要求

### 后端开发
- **Go**: 1.21.0+
- **构建标签**: `with_gvisor` (用于 TUN 模式)

### 前端开发 (gui - 旧版)
- **Node.js**: 16+
- **包管理器**: yarn
- **框架**: Vue 2.7 + Vue CLI 5

### 前端开发 (ngui - 新版)
- **Node.js**: 18+
- **包管理器**: pnpm
- **框架**: Nuxt 3 + TypeScript

## 主要依赖

### Go 后端核心依赖
```
github.com/gin-gonic/gin           # Web 框架
github.com/v2fly/v2ray-core/v5     # V2Ray 核心 (fork)
go.etcd.io/bbolt                   # BoltDB 嵌入式数据库
github.com/gorilla/websocket       # WebSocket 支持
github.com/dgrijalva/jwt-go/v4     # JWT 认证
github.com/miekg/dns               # DNS 库
github.com/sagernet/sing-tun       # TUN 设备支持
github.com/google/gopacket         # 网络包处理
github.com/v2rayA/shadowsocksR     # SSR 协议支持
```

### Vue 前端 (gui) 依赖
```
vue@^2.7.14                        # Vue 2 框架
buefy@^0.9.22                      # Bulma UI 组件库
vue-router@^3.0.6                  # 路由
vuex@^3.0.1                        # 状态管理
axios@^0.21.1                      # HTTP 客户端
vue-i18n@^8.15.3                   # 国际化
```

### Nuxt 前端 (ngui) 依赖
```
nuxt@3.2.0                         # Nuxt 3 框架
element-plus@^2.2.30               # Element Plus UI
@vueuse/core@^9.12.0               # VueUse 工具库
@unocss/nuxt@^0.49.4               # UnoCSS 原子化 CSS
@nuxtjs/i18n@8.0.0-beta.9          # 国际化
```

## 构建流程

### 完整构建 (build.sh)
```bash
# 1. 构建前端 (输出到 service/server/router/web)
cd gui && yarn && OUTPUT_DIR=../service/server/router/web yarn build

# 2. 构建后端 (前端嵌入二进制)
cd service && CGO_ENABLED=0 go build -tags "with_gvisor" \
    -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" \
    -o ../v2raya
```

### 开发模式
```bash
# 后端开发
cd service && go run .

# 前端开发 (gui)
cd gui && yarn serve

# 前端开发 (ngui)
cd ngui && pnpm dev
```

## 配置系统

### 环境变量 (V2RAYA_ 前缀)
| 变量 | 默认值 | 描述 |
|------|--------|------|
| V2RAYA_ADDRESS | 0.0.0.0:2017 | 监听地址 |
| V2RAYA_CONFIG | /etc/v2raya | 配置目录 |
| V2RAYA_V2RAY_BIN | 自动检测 | v2ray 可执行文件路径 |
| V2RAYA_V2RAY_ASSETSDIR | - | geoip.dat 等资源目录 |
| V2RAYA_WEBDIR | - | Web 文件目录（使用外部文件） |
| V2RAYA_LOG_LEVEL | info | 日志级别 |
| V2RAYA_LOG_FILE | - | 日志文件路径 |
| V2RAYA_LITE | false | 精简模式（非 root） |

### 命令行参数
支持与环境变量对应的命令行参数，如 `--address`、`--config` 等。

### 数据存储
- **配置数据库**: `{CONFIG_DIR}/bolt.db` (BoltDB)
- **V2Ray 配置**: `{CONFIG_DIR}/config.json`
- **资源文件**: geoip.dat, geosite.dat

## 代码风格

### Go 代码
- 使用标准 Go 代码风格
- 错误处理使用显式返回
- 日志使用 `pkg/util/log` 包

### 前端代码
- Vue 组件使用 Options API (gui) / Composition API (ngui)
- 使用 ESLint + Prettier 格式化

## 协议插件系统

协议支持通过 `pkg/plugin/` 目录下的插件实现：
- `ss/` - Shadowsocks
- `ssr/` - ShadowsocksR
- `trojanc/` - Trojan
- `tuic/` - TUIC
- `juicity/` - Juicity
- `socks5/` - SOCKS5
- 等等

每个插件在 `init()` 中注册自己的创建函数。

## 测试

### 后端测试
```bash
cd service && go test ./...
```

### Docker 开发环境
```bash
docker-compose -f docker-compose.dev.yml up
```

## 常见问题

### 1. TProxy 不可用
需要加载内核模块：`modprobe xt_TPROXY`

### 2. geoip.dat 缺失
程序会自动下载，或手动放置到资源目录

### 3. 端口被占用
默认端口 2017，可通过 `--address` 参数修改
