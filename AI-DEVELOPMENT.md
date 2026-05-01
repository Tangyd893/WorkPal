# AI 协作开发记录

> 本文件记录 WorkPal 项目中 AI 辅助开发的实际情况：哪些模块由 AI 生成、哪些由人工调整，以及 AI 代码中发现的典型问题。

## 参与比例

| 模块 | AI 生成 | 人工调整 |
|------|---------|---------|
| 后端 CRUD（Repo/Service/Handler）| ~95% | 错误处理、并发加固、接口抽象 |
| WebSocket Hub | ~80% | Mutex 细化、投递链路、Origin 校验、race 检测 |
| 前端页面组件 | ~90% | Zustand store 重构、统一响应解包、消息去重、可访问性语义 |
| 配置文件生成 | ~60% | YAML key 映射、默认配置路径、样例配置 |
| 基础设施（Docker/CI）| ~60% | GitHub Actions 配置、docker-compose 整理、npm 工具链统一、Compose 配置校验 |

---

## 当前项目基线

- 后端：Go 1.22 + Gin + GORM + PostgreSQL 16 + Redis 7 + Redis Streams + Bleve + MinIO/本地文件双模式。
- 前端：React 18 + TypeScript + Vite 5 + Zustand + Axios，包管理统一为 npm/`package-lock.json`。
- 配置：`backend/configs/config.example.yaml` 为样例，真实 `config.yaml` 不提交；未设置 `CONFIG_PATH` 时优先读 `configs/config.yaml`，再回退到样例配置。
- API：HTTP 响应统一为 `{ code, message, data }`，失败响应带对应 HTTP 状态码；前端 Axios 拦截器直接返回 `data`。
- WebSocket：连接使用 `/ws?token=` 鉴权，握手后加入当前用户已有会话；聊天消息通过 HTTP API 落库后广播，避免未持久化 WS 消息。
- 权限：会话消息、搜索、文件上传/下载/列表均按当前用户的会话成员身份或文件所有者做校验。

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

**修复**：在 `/ws` 升级前从 query string 解析 token，并保留 Header 作为兼容路径：
```go
tokenStr := c.Query("token")
claims, err := auth.ParseToken(tokenStr)
userID := claims.UserID
```

**教训**：浏览器 WebSocket API 不能像普通 HTTP 请求一样自由设置自定义 Header；如果采用 URL token，必须配合 HTTPS/WSS、日志脱敏和 Origin 校验。

---

### 4. AI 生成的 Hub 并发代码存在 race condition

**问题描述**：AI 生成的 WebSocket Hub，`clients` map（`map[int64]*Client`）的读写没有锁保护。并发注册/注销客户端时 race detector 报红。

**根因**：AI 写代码时对 Go map 的并发安全性认知不足，`sync.RWMutex` 写了但用在错误的粒度上。

**修复**：
- Hub 结构体内 `clients` 读写用 `sync.RWMutex` 保护
- `SendToUser` / `SendToUsers` 在 `RLock` 区间内完成 Send 调用
- Client 的 `SendCh` 有 256 缓冲，满载时做背压丢弃（而非直接 panic）
- 只保留一个写循环消费 `SendCh`，避免多个 goroutine 竞争读取同一个 channel
- `Send` 在连接关闭时直接返回，并兜底 recover，避免向已关闭 channel 发送导致 panic

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

### 6. 前后端统一响应契约不一致

**问题描述**：后端统一返回 `{code,message,data}`，但前端聊天页有的地方按 `res.items` 读，有的把 `res` 当消息对象，有的又读 `res.data`。接口成功时页面仍可能显示空列表、搜索崩溃或插入错误消息结构。

**根因**：AI 同时生成了后端统一响应和前端页面代码，但没有固定 API client 的返回契约，导致页面直接猜测后端结构。

**修复**：Axios 响应拦截器统一检查 `code`，非 0 抛错，成功时只返回 `data`；页面只处理业务数据：
```ts
request.interceptors.response.use((res) => {
  const body = res.data
  if (body.code !== 0) throw new Error(body.message)
  return body.data ?? null
})
```

**教训**：前后端分离项目必须先稳定 API client 的返回类型，再写页面逻辑；不要在每个页面里重复拆响应。

---

### 7. 搜索与文件接口权限边界缺失

**问题描述**：消息全局搜索未按当前用户的会话过滤；文件下载和会话文件列表只检查登录态，没有检查文件所有者或会话成员身份。

**根因**：AI 更容易生成“功能能跑通”的 CRUD，但跨资源权限常被遗漏，尤其是搜索索引和文件这类不直接挂在当前用户路径下的接口。

**修复**：
- 搜索不指定 `conv_id` 时，先查询当前用户已加入会话，再只在这些会话范围内返回结果。
- 文件上传到会话、下载会话文件、列出会话文件时校验会话成员；个人文件只允许上传者访问。
- 添加群成员增删的群主校验，避免普通成员任意拉人/踢人。

**教训**：权限检查应靠服务端资源关系判断，不能依赖前端隐藏按钮或“用户一般不会传别人的 id”。

---

### 8. 默认启动路径与样例配置缺失

**问题描述**：README 写 `go run ./cmd/server`，但服务启动时去 Go 临时编译目录查找 `configs/config.yaml`，导致配置文件不存在。

**根因**：AI 用 `os.Executable()` 推导配置目录，这对编译后的二进制可能可用，但对 `go run` 会指向临时目录；同时真实 `config.yaml` 被 `.gitignore` 忽略，却没有提交样例文件。

**修复**：
- 新增 `backend/configs/config.example.yaml`。
- 默认配置查找顺序改为当前工作目录下的 `configs/config.yaml`、`configs/config.example.yaml`，并兼容仓库根目录运行。
- `jwtSecret` 为空时直接启动失败，避免空密钥签发 token。

**教训**：启动命令和配置查找必须用真实命令验证，不能只从“打包后目录结构”推断。

---

### 9. 未读 SQL 与数据模型不一致

**问题描述**：`CountUnread` 查询 `message_reads.msg_id`，但当前 `message_reads` 模型只有 `(user_id, conv_id, read_at)`，真实数据库会报列不存在。

**根因**：AI 混用了“每条消息一条已读记录”和“每个会话一个已读水位”的两种模型。

**修复**：统一为会话级已读水位：`message_reads` 以 `(user_id, conv_id)` 为主键，`read_at` 表示该用户在会话中的最后已读时间；未读统计按 `messages.created_at > read_at` 计算。

**教训**：读回执模型必须先选定粒度，否则 Repo SQL、Service 语义和前端展示会互相矛盾。

---

## AI 辅助开发的最佳实践

基于以上踩坑总结：

1. **YAML/配置文件**：AI 生成后立刻人工确认 key 名与代码中的字段名一致
2. **并发相关代码**：AI 写完后必须跑 `go test -race`，不信任 AI 的 mutex 逻辑
3. **状态管理**：避免在非组件上下文中使用有 hydration 时序问题的方案（如 Zustand persist）
4. **错误处理**：跑 golangci-lint 强制检查，AI 生成的 `errcheck` 问题比想象中更多
5. **API 契约**：前后端接口文档先确认，再让 AI 生成代码，避免 token 位置、响应格式这类对接问题
6. **权限边界**：任何按 id 访问资源的接口都要补服务端所有权/成员关系校验
7. **启动脚本**：README 里的命令必须在空环境里跑通，配置样例必须随代码提交

---

## 项目中的 AI 工作流

```
需求 → AI 生成初稿 → 人工 Review（重点：安全+并发+权限）→ gofmt/lint/test/race → 文档同步 → 提交
```

目前 WorkPal 的单测（`auth_svc`、`message_svc`、`conversation_svc`、`presence_svc`、文件服务、搜索服务、Hub 并发）以人工事后补充和修正为主。普通 `go test ./...` 可作为本地基础验证，`go test -race ./...` 建议在 Linux CI 或本机可用 race detector 的环境中运行。GitHub Actions 当前覆盖后端构建、`go vet`、golangci-lint、race 测试、前端 lint、前端单元与组件测试、前端生产构建和 Docker Compose 配置校验。
