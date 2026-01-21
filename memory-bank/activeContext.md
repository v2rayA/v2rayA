# v2rayA 当前活动上下文

## 最近完成的工作

### URL 查询参数空格问题全面修复 (2026-01-22)

**问题描述：**
订阅链接中 `+` 在 `application/x-www-form-urlencoded` 编码中被解码为空格，导致多个参数解析失败，使 xray 核心崩溃。

**遇到的具体错误：**
1. `type=ws+` → `"unexpected transport type: ws "`
2. `type=tcp+` → `"unknown transport protocol: tcp "`
3. `fp=chrome+` → `"unknown \"fingerprint\": chrome "`
4. `net=none` → `"unexpected transport type: none"`
5. `type=raw` → `"unexpected transport type: raw"`
6. `type=splithttp` → `"unexpected transport type: splithttp"`
7. `sid=2404+` → `"invalid \"shortId\": 2404"` (REALITY 配置)

**解决方案：**
对所有从 `url.Query().Get()` 获取的参数统一应用 `strings.TrimSpace()` 清理首尾空格，同时添加传输协议别名转换。

**修改的文件：**

1. **service/core/serverObj/v2ray.go - ParseVlessURL 函数**
   - 所有字段应用 TrimSpace：`aid`, `type`, `headerType`, `host`, `sni`, `path`, `security`, `fp`, `pbk`, `sid`, `spx`, `flow`, `alpn`, `allowInsecure`, `key`
   - 后续赋值也应用 TrimSpace：`serviceName`, `host`, `seed`, `quicSecurity`
   - 添加传输协议别名转换：`raw`→`tcp`, `splithttp`→`xhttp`

2. **service/core/serverObj/v2ray.go - ParseVmessURL 函数**
   - 添加 `none`→`tcp` 转换

3. **service/core/serverObj/trojan.go - ParseTrojanURL 函数**
   - 所有字段应用 TrimSpace：`allowInsecure`, `peer`, `sni`, `alpn`, `type`, `path`, `serviceName`, `encryption`, `host`

**技术要点：**
- URL 中 `+` 在 `application/x-www-form-urlencoded` 编码中被解码为空格是标准行为
- `strings.TrimSpace()` 只去除首尾空白字符，不影响有效内容
- 这些参数值本身不应包含首尾空格，因此修复是安全的

## 当前状态

- 无活跃开发任务
- 代码编译验证通过

## 技术笔记

### URL 参数解析最佳实践
对于所有从 `url.Query().Get()` 获取的参数，应统一应用 `strings.TrimSpace()` 进行防御性处理，避免因订阅源编码问题导致的解析失败。

### 传输协议别名映射
- `raw` → `tcp`
- `none` → `tcp`
- `splithttp` → `xhttp`
- `websocket` → `ws`

## 上下文刷新时间
2026-01-22 00:18 CST
