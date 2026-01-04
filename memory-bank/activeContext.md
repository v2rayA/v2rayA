# v2rayA 当前活动上下文

## 最近完成的工作

### httpupgrade 传输类型支持修复 (2026-01-04)

**问题描述：**
v2rayA 不支持 httpupgrade 传输类型，导入包含 `type=httpupgrade` 的节点时报错：
```
unexpected transport type: httpupgrade
```

**解决方案：**
在后端代码中添加对 httpupgrade 传输类型的支持。

**修改的文件：**

1. **service/core/coreObj/coreObj.go**
   - 添加 HTTPUpgradeSettings 结构体
   - 在 StreamSettings 结构体中添加 HTTPUpgradeSettings 字段

2. **service/core/serverObj/v2ray.go**
   - 在 Configuration() 方法的 switch 语句中添加 httpupgrade case
   - 在 ExportToURL() 方法的 switch 语句中添加 httpupgrade case

**代码变更详情：**

```go
// service/core/coreObj/coreObj.go
type HTTPUpgradeSettings struct {
    Path string `json:"path,omitempty"`
    Host string `json:"host,omitempty"`
}

// 在 StreamSettings 结构体中添加
HTTPUpgradeSettings *HTTPUpgradeSettings `json:"httpupgradeSettings,omitempty"`
```

```go
// service/core/serverObj/v2ray.go - Configuration() 方法
case "httpupgrade":
    core.StreamSettings.HTTPUpgradeSettings = &coreObj.HTTPUpgradeSettings{
        Path: v.Path,
        Host: v.Host,
    }
```

```go
// service/core/serverObj/v2ray.go - ExportToURL() 方法
case "httpupgrade":
    setValue(&query, "path", v.Path)
    setValue(&query, "host", v.Host)
```

**构建与部署：**
- 版本标识：`unstable-httpupgrade`
- 二进制位置：`/usr/bin/v2raya`
- 备份位置：`/usr/bin/v2raya.bak`
- GUI 已嵌入后端

**验证结果：**
- ✅ 节点导入成功，不再报 "unexpected transport type: httpupgrade" 错误
- 测试节点：
  - `vless://...@173.245.59.132:8880?type=httpupgrade` (无 TLS)
  - `vless://...@85.192.48.101:443?type=httpupgrade&security=tls` (带 TLS)

## 当前状态

- 无活跃开发任务
- v2rayA 服务正常运行 (版本 unstable-httpupgrade)

## 技术笔记

### httpupgrade 传输类型
httpupgrade 是一种类似于 websocket 的传输协议，主要区别：
- 使用 HTTP Upgrade 机制建立连接
- 配置参数：path（路径）和 host（主机名）
- 可与 TLS 配合使用

### 代码架构参考
- 传输类型配置在 `core/coreObj/coreObj.go` 中定义结构体
- 配置生成逻辑在 `core/serverObj/v2ray.go` 的 Configuration() 方法
- URL 导出逻辑在 `core/serverObj/v2ray.go` 的 ExportToURL() 方法
- 添加新传输类型时需要同时修改这三处

## 待办事项

无

## 上下文刷新时间
2026-01-04 22:40 CST
