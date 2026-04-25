# AI 协作开发记录

> 本文件记录 WorkPal 项目中 AI 辅助开发的实际情况：哪些模块由 AI 生成、哪些由人工调整，以及 AI 代码中发现的典型问题。

## 参与比例

| 模块 | AI 生成 | 人工调整 |
|------|---------|---------|
| 后端 CRUD（Repo/Service/Handler）| ~95% | 错误处理、并发加固、接口抽象 |
| WebSocket Hub | ~80% | Mutex 细化、race 检测 |
| 前端页面组件 | ~90% | Zustand store 重构、响应格式兼容 |
| 配置文件生成 | ~60% | YAML key 映射问题修复 |
| 基础设施（Docker/CI）| ~40% | GitHub Actions 配置、docker-compose 整理 |

---

## AI 踩坑记录

以下问题均来自真实调试经历，非理论推测。

### 1. JWT 配置项 YAML → Go struct 映射失败

**问题描述**：登录时 Token 一直报 signature invalid，所有 Token 签名失效。

**根因**：AI 生成的 `config.yaml` 使用下划线命名 `jwt_secret`，但 Go struct 字段是驼峰 `jwtSecret`，viper 没有 `mapstructure` 标签，映射失败后该字段为空字符串，导致所有 Token 用空密钥签名。

**修复**：
```yaml
# 错误（AI 生成）
jwt_secret: "xxx"
jwt_expiry_hours: 72

# 正确
jwtSecret: "xxx"
jwtExpiryHours: 72
```

**教训**：viper 项目中 YAML key 必须与 Go struct 字段名完全一致，或显式加 `mapstructure` 标签。

---

### 2. Zustand persist 导致 hydration 时数据丢失

**问题描述**：用户登录成功后跳转到聊天页面，但立即显示未登录（token 为空），刷新页面后反而正常。

**根因**：AI 生成的 `useAuthStore` 使用了 `persist` middleware。axios interceptor 在非 React 上下文中调用 `getState()` 时，localStorage 数据尚未完成 hydration，得到空状态，导致写入后 store 被覆盖为空。

**修复**：移除 `persist` middleware，改用手动 `localStorage.getItem/setItem`，在 interceptor 和组件中都直接读本地存储而非通过 store 状态。

**教训**：Zustand 的 `persist` 在 SSR 场景或非组件代码中使用时有严重的时序问题，不适合这类需要在 interceptor 里读 token 的架构。

---

### 3. WebSocket 认证 Token 位置

**问题描述**：WebSocket 连接建立后立即断开，客户端报 401。

**根因**：AI 生成的 WebSocket handler 把 JWT token 放在 HTTP Header 里让客户端传递。但前端把所有认证信息放在 URL query parameter（`/ws?token=xxx`），服务端从未读取。

**修复**：在 `UpgradeToWebSocket` 之前从 query string 解析 token：
```go
tokenStr := r.URL.Query().Get("token")
claims, err := auth.ParseToken(tokenStr)
userID := claims.UserID
```

**教训**：AI 对 WebSocket 认证的常见误解是沿用 HTTP 的 Header 方案，但浏览器 WebSocket API 只支持 URL 参数传递 token。

---

### 4. AI 生成的 Hub 并发代码存在 race condition

**问题描述**：AI 生成的 WebSocket Hub，`clients` map（`map[int64]*Client`）的读写没有锁保护。并发注册/注销客户端时 race detector 报红。

**根因**：AI 写代码时对 Go map 的并发安全性认知不足，`sync.RWMutex` 写了但用在错误的粒度上。

**修复**：
- Hub 结构体内 `clients` 读写用 `sync.RWMutex` 保护
- `SendToUser` / `SendToUsers` 在 `RLock` 区间内完成 Send 调用
- Client 的 `SendCh` 有 256 缓冲，满载时做背压丢弃（而非直接 panic）

```go
// 修复后
func (h *Hub) SendToUser(userID int64, content []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    if client, ok := h.clients[userID]; ok {
        client.Send(content)  // 仍然并发安全
    }
}
```

**教训**：AI 的并发代码必须用 `go test -race` 验证，否则生产环境必出问题。

---

### 5. 错误处理被忽略

**问题描述**：golangci-lint 跑出 13 个 `errcheck` 错误，所有 `AddMember`、`MkdirAll`、`json.Unmarshal` 的错误返回值均被丢弃。

**根因**：AI 生成的代码里写了 `err := ...` 但立即忽略不用，看起来像是占位符忘了补。

**修复**：逐个补全错误检查，关键路径（文件写入、DB操作）不允许错误吞掉。

**教训**：AI 代码中的"占位符式错误处理"是高频问题，必须配合 lint 工具检查。

---

## AI 辅助开发的最佳实践

基于以上踩坑总结：

1. **YAML/配置文件**：AI 生成后立刻人工确认 key 名与代码中的字段名一致
2. **并发相关代码**：AI 写完后必须跑 `go test -race`，不信任 AI 的 mutex 逻辑
3. **状态管理**：避免在非组件上下文中使用有 hydration 时序问题的方案（如 Zustand persist）
4. **错误处理**：跑 golangci-lint 强制检查，AI 生成的 `errcheck` 问题比想象中更多
5. **API 契约**：前后端接口文档先确认，再让 AI 生成代码，避免 token 位置、响应格式这类对接问题

---

## 项目中的 AI 工作流

```
需求 → AI 生成初稿 → 人工 Review（重点：安全+并发）→ 跑 lint + race → 单测覆盖 → 提交
```

目前 WorkPal 的单测（`auth_svc`、`message_svc`、`conversation_svc`、`presence_svc`）均为人工事后补充，非 AI 生成。
