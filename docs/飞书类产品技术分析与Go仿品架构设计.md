# 飞书类产品技术分析报告 & Go 仿品架构设计

> 本文档对飞书、钉钉、企业微信等企业协作平台进行全方位技术拆解，并给出用 Go 语言构建仿品的技术架构设计方案。

---

## 目录

1. [产品级分析](#1-产品级分析)
2. [竞品横向对比](#2-竞品横向对比)
3. [核心系统架构](#3-核心系统架构)
4. [技术栈选型](#4-技术栈选型)
5. [项目结构设计](#5-项目结构设计)
6. [分阶段开发路线](#6-分阶段开发路线)
7. [数据库设计概要](#7-数据库设计概要)
8. [API 设计概要](#8-api-设计概要)

---

## 1. 产品级分析

### 1.1 飞书产品定位

飞书（Feishu/Lark）是字节跳动自研的企业协作平台，定位为 **"下一代工作方式"**，集成了 IM、文档、云盘、日历、OKR、审批等全套企业协作能力。核心技术壁垒在于：

- **多端实时同步**：iOS/Android/macOS/Windows/Web 五端消息、文档、状态完全同步
- **云文档协作编辑**：支持多人实时协同编辑文档、表格、幻灯片
- **开放平台能力**：小程序、机器人、Webhook、开放 API 支撑企业定制化

### 1.2 核心功能矩阵

| 功能域 | 具体功能 | 用户价值 |
|--------|----------|----------|
| **IM 即时通讯** | 私聊、群聊、话题群、多人通话 | 日常沟通协作 |
| **消息** | 文字/图片/文件/代码块/卡片/表情、@成员、回复、撤回、编辑、已读回执 | 信息传递丰富度 |
| **组织架构** | 部门树、成员列表、管理员角色、跨企业搜索 | 人员管理 |
| **云文档** | 文档（类 Notion）、表格、幻灯片、思维导图 | 内容协作 |
| **日历** | 日程创建、会议邀请、会议室预约 | 时间协作 |
| **任务** | 任务创建/分配/截止日期/优先级、项目管理视图 | 目标追踪 |
| **视频会议** | 1080P 会议、屏幕共享、录制、云录制 | 远程协作 |
| **审批** | 假期/报销/加班等自定义审批流 | 流程自动化 |
| **搜索** | 全局搜索：消息/文档/人员/群聊 | 知识获取 |
| **开放平台** | 小程序、机器人、Webhook、SSO | 生态扩展 |

### 1.3 飞书技术特色

#### 1.3.1 消息系统设计

飞书消息采用 **Channel-based** 架构（类似 Slack）：

```
用户 → 消息队列（Kafka/NATS）→ 消息服务 → 存储服务
                              ↓
              WebSocket Gateway → 推送至客户端
```

- **写扩散** vs **读扩散** 的权衡：飞书采用混合模式
  - 私聊：写扩散（发1条，存1份，实时推）
  - 群聊：大群采用读扩散（消息入库，接收者各自拉取）
- **消息 ID**：使用雪花算法（Snowflake）生成全局唯一 ID，支持客户端本地去重
- **消息漫游**：基于分页的漫游协议，消息云端持久化，本地只存最近 N 条

#### 1.3.2 文档协作（OT 算法）

飞书文档使用 **Operational Transformation (OT)** 算法实现多人实时编辑：

```
用户A编辑 → 本地预览 → 发送 Operation 到服务器
用户B编辑 → 本地预览 → 发送 Operation 到服务器
服务器收到所有 Operation → 转换（Transform）冲突 → 广播给所有用户
```

关键点：
- 每次编辑不发送全量文本，只发送操作（insert/delete/retain）
- 服务端维护文档的完整操作历史（Append-only log）
- 支持离线编辑，重连后自动合并

#### 1.3.3 全文搜索

飞书搜索支持 **消息+文档+人员+群组**，底层基于全文搜索引擎：

```
Elasticsearch 集群
  ├── 消息索引（分词：单字/词组）
  ├── 文档索引（支持富文本）
  └── 用户/群组索引（精确匹配）
```

---

## 2. 竞品横向对比

### 2.1 核心竞品对比表

| 维度 | 飞书 | 钉钉 | 企业微信 | Slack | Discord |
|------|------|------|----------|-------|---------|
| **母公司** | 字节跳动 | 阿里巴巴 | 腾讯 | Salesforce | Discord Inc. |
| **生态绑定** | 字节全家桶 | 阿里全家桶 | 微信生态 | 国际化/开放 | 游戏社区 |
| **文档协作** | ✅ 强大 | ⚠️ 一般 | ❌ 依赖腾讯文档 | ⚠️ 弱 | ❌ 无 |
| **IM 实时性** | ✅ 优秀 | ✅ 优秀 | ✅ 优秀 | ✅ 优秀 | ✅ 优秀 |
| **开放平台** | ✅ 小程序+机器人 | ✅ 钉钉小程序 | ✅ 企业微信小程序 | ✅ Slack App | ⚠️ Bot |
| **视频会议** | ✅ 1080P/录制 | ✅ 1080P | ✅ 腾讯会议集成 | ✅ Huddle | ✅ Stage |
| **国际化** | ✅ Lark 海外 | ❌ | ❌ | ✅ 全球 | ✅ 全球 |
| **定价** | 免费+增值 | 免费+增值 | 免费 | $8/user/月 | 免费+ Nitro |

### 2.2 技术架构流派

| 类型 | 代表产品 | 架构特点 |
|------|----------|----------|
| **重量级微服务** | 飞书、钉钉 | K8s + 100+ 微服务 + Service Mesh |
| **单体 + 插件化** | 企业微信早期 | 单体 + 动态插件加载 |
| **Serverless 优先** | 新型 SaaS | 大量使用云函数 + BaaS |
| **消息驱动** | Slack | Kafka 为核心，所有事件消息化 |
| **边缘计算** | 飞书国际版 | 多 Region 边缘节点，本地化部署 |

---

## 3. 核心系统架构

### 3.1 整体架构图

```
                           ┌─────────────────────────────────────────────┐
                           │                   客户端                      │
                           │  iOS / Android / macOS / Windows / Web       │
                           └────────────────────┬────────────────────────┘
                                                │ HTTPS / WSS
                           ┌────────────────────▼────────────────────────┐
                           │               API Gateway / Nginx            │
                           │         (路由 / 鉴权 / 限流 / TLS)            │
                           └──┬──────────┬──────────┬──────────┬──────────┘
                              │          │          │          │
                   ┌─────────▼──┐  ┌───▼────┐ ┌──▼────┐ ┌──▼──────┐
                   │   用户服务   │  │ 消息服务 │ │文档服务 │ │ 文件服务 │
                   │  (User Svc) │  │ (IM Svc) │ │(Doc Svc)│ │(File Svc)│
                   └──────┬──────┘  └────┬───┘ └───────┘ └─────────┘
                          │              │
              ┌───────────▼──────────────▼──────────────┐
              │              Message Bus (Kafka/NATS)     │
              │   (消息流转 / 事件驱动 / 异步解耦)        │
              └──┬────────┬──────────────┬──────────────┘
                 │        │              │
          ┌──────▼──┐ ┌──▼───┐  ┌──────▼──────┐
          │  消息存储 │ │ 推送  │  │ 搜索引擎    │
          │(Postgres)│ │(APNs)│  │(Elasticsearch)│
          └──────────┘ └──────┘  └─────────────┘
```

### 3.2 服务模块划分

| 服务 | 职责 | 技术选型建议 |
|------|------|-------------|
| **API Gateway** | 路由、鉴权、限流、日志 | Kong / APISIX / 自研（Gin） |
| **User Service** | 用户注册/登录/组织架构/权限 | Gin + GORM |
| **IM Service** | 消息收发、群聊管理、已读未读 | Gin + WebSocket + Kafka |
| **Doc Service** | 云文档 CRUD、版本管理、权限 | Gin + OT 算法库 |
| **File Service** | 文件上传/下载/预览/转码 | Gin + MinIO |
| **Search Service** | 全文索引、搜索建议 | Gin + Elasticsearch |
| **Notify Service** | 推送（APNs/FCM/厂商通道） | Worker + Kafka Consumer |
| **Calendar Service** | 日程、会议、邀请 | Gin + GORM |

### 3.3 关键架构决策

#### 3.3.1 消息可靠投递（AT LEAST ONCE）

```
Producer → Kafka → Consumer → ACK
             ↑
         消息持久化（未 ACK 不删除）
```

- 生产者：发送消息到 Kafka，等待 ISR 副本确认
- 消费者：处理完消息后手动 ACK，防止丢消息
- 补偿机制：消费者幂等处理，消息去重

#### 3.3.2 WebSocket 长连接管理

```
Client → Nginx/LB → WebSocket Gateway → Redis (Session) → Kafka/Redis PubSub
```

- **连接数**：单机 WebSocket Gateway 可维护 10W+ 长连接（基于 epoll/kqueue）
- **消息路由**：Redis PubSub 广播，Gateway 节点间通过 Redis 同步
- **心跳检测**：客户端每 30s 发心跳，服务端标记在线状态到 Redis

#### 3.3.3 水平扩展策略

```
                    ↑ 请求
              ┌─────────────┐
              │ Load Balancer│
              └──────┬──────┘
        ┌────────────┼────────────┐
   ┌────▼────┐ ┌───▼────┐ ┌───▼────┐
   │ Node 1  │ │ Node 2 │ │ Node 3 │
   └─────────┘ └────────┘ └─────────┘
        ↑           ↑           ↑
   所有节点无状态，可随时扩缩容
```

---

## 4. 技术栈选型

### 4.1 基础层（必选）

| 组件 | 推荐选型 | 原因 |
|------|----------|------|
| **语言/框架** | Go 1.21+ / Gin / Echo / Fiber | 高并发、简洁、部署简单 |
| **数据库** | PostgreSQL 16 | 事务强一致、JSON 支持、PostGIS |
| **缓存** | Redis 7 | String/Hash/List/SortedSet 多种数据结构 |
| **消息队列** | Kafka 或 NATS | Kafka 生态完善，NATS 更轻量 |
| **WebSocket** | gorilla/websocket / melean/rwebsocket | 稳定、广泛使用 |
| **ORM** | GORM / sqlx | GORM 便捷，sqlx 性能更好 |
| **配置** | Viper | 支持 YAML/TOML/ENV/命令行 |
| **日志** | zap + logrus | 结构化日志 + 彩色输出 |
| **API 文档** | Swagger / Protobuf | go-swagger / grpc-gateway |
| **容器化** | Docker + Docker Compose | 开发/测试环境标准化 |

### 4.2 进阶级（按需选型）

| 组件 | 推荐选型 | 适用场景 |
|------|----------|----------|
| **全文搜索** | Elasticsearch / Bleve | 消息/文档搜索 |
| **文件存储** | MinIO / 阿里云 OSS | 私有部署 / 云部署 |
| **服务通信** | gRPC + Protobuf | 内部微服务间调用 |
| **API 网关** | Kong / APISIX | 路由/鉴权/限流/插件化 |
| **链路追踪** | Jaeger / Zipkin | 分布式追踪 |
| **指标监控** | Prometheus + Grafana | 运行时指标 |
| **日志聚合** | Loki + Grafana / ELK | 日志收集查询 |
| **服务网格** | Istio / Linkerd | 微服务治理（可选） |
| **对象检测** | 阿里云内容安全 | 图片/文本内容审核 |

### 4.3 工具链

| 用途 | 工具 |
|------|------|
| **API 测试** | Apifox / Postman / curl |
| **数据库管理** | pgAdmin / DBeaver |
| **Redis 管理** | RedisInsight |
| **Docker 管理** | Portainer |
| **代码规范** | golangci-lint + pre-commit |
| **CI/CD** | GitHub Actions / GitLab CI |
| **代码生成** | `go generate` + stringer + mockgen |

---

## 5. 项目结构设计

### 5.1 推荐目录结构（DDD + 洋葱架构）

```
feishu-clone/
├── cmd/                          # 入口点
│   └── server/
│       └── main.go               # 主程序入口
│   └── migrator/
│       └── main.go               # 数据库迁移工具
│
├── internal/                     # 私有应用代码（不可被外部 import）
│   ├── user/                     # 用户模块（DDD 风格）
│   │   ├── handler/              # HTTP Handler / gRPC Server
│   │   │   ├── user.go
│   │   │   └── auth.go
│   │   ├── service/              # 业务逻辑层
│   │   │   ├── user_svc.go
│   │   │   └── auth_svc.go
│   │   ├── repo/                 # 数据访问层
│   │   │   └── user_repo.go
│   │   ├── model/                 # 数据模型（DO / Entity）
│   │   │   └── user.go
│   │   └── pkg/                  # 模块内部工具包
│   │
│   ├── im/                       # IM 即时通讯模块
│   │   ├── handler/
│   │   │   ├── message.go        # 消息 CRUD
│   │   │   ├── conversation.go   # 会话管理
│   │   │   └── ws.go             # WebSocket Handler
│   │   ├── service/
│   │   │   ├── message_svc.go
│   │   │   └── presence_svc.go    # 在线状态
│   │   ├── repo/
│   │   │   ├── message_repo.go
│   │   │   └── conversation_repo.go
│   │   ├── model/
│   │   │   ├── message.go
│   │   │   └── conversation.go
│   │   └── ws/
│   │       ├── hub.go            # WebSocket Hub（连接管理器）
│   │       ├── client.go         # 单个连接客户端
│   │       └── broadcast.go      # 广播逻辑
│   │
│   ├── org/                      # 组织架构模块
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repo/
│   │   └── model/
│   │
│   ├── file/                     # 文件存储模块
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repo/
│   │   └── model/
│   │
│   └── common/                   # 公共代码（可被 internal 所有模块 import）
│       ├── middleware/           # 中间件（JWT 鉴权/日志/限流/CORS）
│       ├── response/              # 统一响应结构
│       ├── errors/               # 自定义错误类型
│       ├── pagination/           # 分页工具
│       └── validator/            # 参数校验
│
├── pkg/                          # 公共工具包（可被外部项目 import）
│   ├── auth/                     # JWT / 密码加密工具
│   ├── cache/                    # Redis 封装
│   ├── msgqueue/                 # Kafka/NATS 封装
│   ├── oss/                      # 对象存储封装
│   └── tracer/                   # 链路追踪封装
│
├── configs/                      # 配置文件
│   ├── config.yaml
│   ├── config.prod.yaml
│   └── config.test.yaml
│
├── deployments/                  # 部署配置
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yaml
│   └── k8s/
│       ├── deployment.yaml
│       └── service.yaml
│
├── migrations/                   # 数据库迁移（sqlc / goose）
│   ├── 001_create_users.sql
│   ├── 002_create_conversations.sql
│   └── 003_create_messages.sql
│
├── proto/                        # Protobuf 定义（gRPC 接口）
│   ├── user.proto
│   ├── im.proto
│   └── generate.sh
│
├── scripts/                      # 工具脚本
│   ├── gen_token.go              # 生成测试 JWT
│   └── mock_data.go              # 构造测试数据
│
├── Makefile                      # make 命令集合
├── go.mod
├── go.sum
└── README.md
```

### 5.2 各层职责说明

```
Handler 层（接口层）
  职责：解析请求参数、参数校验、调用 Service、返回响应
  注意：禁止直接操作数据库，禁止写业务逻辑

Service 层（业务逻辑层）
  职责：实现核心业务逻辑、事务控制、跨 Repo 组合
  注意：一个 Service 方法对应一个业务用例

Repo 层（数据访问层）
  职责：数据库 CRUD 操作，向上对 Service 透明
  注意：每个 Repo 只操作一个实体，不跨表 join

Model 层（数据模型层）
  职责：定义数据结构体、ORM tag、与 DB schema 对应
```

---

## 6. 分阶段开发路线

### 第一阶段：基础骨架（Week 1-2）—— 练 Go 基础

**目标**：跑通 Web API 闭环，理解 Go 项目结构

```
✅  项目初始化（go mod init）
✅  Gin 框架集成（路由/中间件/日志）
✅  配置管理（Viper 读取 YAML）
✅  PostgreSQL 连接 + GORM AutoMigrate
✅  用户注册 / 登录（bcrypt + JWT）
✅  HTTP API：获取用户信息、修改资料
✅  单元测试覆盖率 > 60%
```

**关键文件**：

```
cmd/server/main.go          # 入口
internal/user/handler/     # 用户接口
internal/user/service/     # 业务逻辑
internal/user/repo/        # 数据访问
internal/common/middleware/ # JWT 中间件
```

**验证方式**：

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","nickname":"测试用户"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
# 返回 JWT token
```

---

### 第二阶段：IM 核心（Week 3-4）—— 练并发 + WebSocket

**目标**：实现私聊/群聊，理解并发编程和长连接

```
✅  会话模型（Conversation）：私聊/群聊
✅  消息模型（Message）：文字/图片/文件
✅  WebSocket 长连接（gorilla/websocket）
✅  消息发送/接收（实时推送）
✅  群聊创建/加入/退出
✅  未读消息计数 + 已读回执
✅  在线状态管理（Redis Sorted Set）
✅  消息历史分页查询
```

**技术要点**：

```go
// WebSocket Hub 设计（核心）
type Hub struct {
    clients    map[*Client]bool      // 所有连接
    rooms      map[string]map[*Client]bool  // 房间（群）-> 客户端
    register   chan *Client
    unregister chan *Client
    broadcast  chan *Message
    mu         sync.RWMutex
}

// 客户端消息格式（JSON）
type WSMessage struct {
    Type    string      `json:"type"`    // "chat"/"presence"/"ack"
    From    string      `json:"from"`
    To      string      `json:"to"`       // user_id 或 room_id
    Content interface{} `json:"content"`
    Seq     int64       `json:"seq"`      // 客户端本地序列号
}
```

**验证方式**：

```bash
# 使用 wscat 测试
wscat -c ws://localhost:8080/ws?token=xxx

# 发送消息
{"type":"chat","to":"user_id","content":"hello"}

# 收到服务器推送
{"type":"chat","from":"user_id","content":"hello","seq":1}
```

---

### 第三阶段：生产级扩展（Week 5-8）—— 练系统设计

**目标**：引入消息队列、分库分表、搜索等生产级组件

```
✅  Kafka 消息总线（消息异步处理、削峰）
✅  消息多级缓存（Redis L1 + DB L2）
✅  Elasticsearch 聊天记录全文搜索
✅  MinIO 文件上传/预览（图片/文档）
✅  gRPC 微服务拆分（User/IM/File 服务独立部署）
✅  API 网关（鉴权/路由/限流）
✅  Prometheus 指标暴露 + Grafana 看板
✅  Jaeger 分布式链路追踪
✅  Docker Compose 一键部署
```

---

### 第四阶段：高级特性（Week 9-12）—— 向飞书看齐

```
✅  云文档协作编辑（OT 算法）
✅  音视频通话（WebRTC SFU）
✅  表情回复/线程消息
✅  消息转发/引用/合并转发
✅  消息免打扰/置顶/折叠
✅  多因素认证（MFA）
✅  开放平台（机器人/Webhook/小程序）
✅  国际化（i18n）
```

---

## 7. 数据库设计概要

### 7.1 核心表结构

```sql
-- 用户表
CREATE TABLE users (
    id            BIGINT PRIMARY KEY DEFAULT nextval('users_id_seq'),
    username      VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname      VARCHAR(128),
    avatar_url    VARCHAR(512),
    email         VARCHAR(255) UNIQUE,
    phone         VARCHAR(32) UNIQUE,
    status        SMALLINT DEFAULT 1,  -- 1正常 2禁用
    created_at    TIMESTAMP DEFAULT NOW(),
    updated_at    TIMESTAMP DEFAULT NOW()
);

-- 组织表（部门）
CREATE TABLE departments (
    id         BIGINT PRIMARY KEY,
    name       VARCHAR(128) NOT NULL,
    parent_id  BIGINT REFERENCES departments(id),  -- 树形自关联
    leader_id  BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 用户部门关联表
CREATE TABLE user_departments (
    user_id      BIGINT REFERENCES users(id),
    dept_id      BIGINT REFERENCES departments(id),
    PRIMARY KEY (user_id, dept_id)
);

-- 会话表（Conversation）
CREATE TABLE conversations (
    id          BIGINT PRIMARY KEY,
    type        SMALLINT NOT NULL,      -- 1=私聊 2=群聊
    name        VARCHAR(256),            -- 群名（私聊为空）
    avatar_url  VARCHAR(512),
    owner_id    BIGINT,                  -- 群主（私聊为空）
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

-- 会话成员表
CREATE TABLE conversation_members (
    conv_id   BIGINT REFERENCES conversations(id),
    user_id   BIGINT REFERENCES users(id),
    joined_at  TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (conv_id, user_id)
);

-- 消息表
CREATE TABLE messages (
    id            BIGINT PRIMARY KEY,  -- 雪花算法 ID
    conv_id       BIGINT REFERENCES conversations(id),
    sender_id     BIGINT REFERENCES users(id),
    type          SMALLINT NOT NULL,   -- 1=文字 2=图片 3=文件 4=代码 5=卡片
    content       TEXT,
    metadata      JSONB,               -- 扩展字段（图片宽高、文件大小等）
    reply_to      BIGINT REFERENCES messages(id),  -- 回复的消息 ID
    created_at    TIMESTAMP DEFAULT NOW(),
    updated_at    TIMESTAMP DEFAULT NOW(),
    deleted_at    TIMESTAMP
);

-- 消息索引（按时间+会话分片）
CREATE INDEX idx_messages_conv_time ON messages(conv_id, created_at DESC);

-- 已读状态表
CREATE TABLE message_reads (
    user_id   BIGINT REFERENCES users(id),
    conv_id   BIGINT REFERENCES conversations(id),
    read_at   TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, conv_id)
);

-- 用户在线状态（Redis 更适合，此处仅作记录）
CREATE TABLE user_presences (
    user_id    BIGINT PRIMARY KEY REFERENCES users(id),
    status     SMALLINT DEFAULT 0,   -- 0=离线 1=在线 2=忙碌
    last_seen  TIMESTAMP DEFAULT NOW()
);
```

### 7.2 分库分表策略（生产级）

```
消息表按 conv_id 分片（按会话 ID 哈希）：
  messages_0  (conv_id % 4 == 0)
  messages_1  (conv_id % 4 == 1)
  messages_2  (conv_id % 4 == 2)
  messages_3  (conv_id % 4 == 3)

热点群聊（大群）单独拆库：
  messages_hot_0  (万人群专属)
  messages_hot_1
```

---

## 8. API 设计概要

### 8.1 RESTful API 路由

```
认证：
POST   /api/v1/auth/register         注册
POST   /api/v1/auth/login            登录
POST   /api/v1/auth/refresh           刷新 Token
POST   /api/v1/auth/logout             登出

用户：
GET    /api/v1/users/me              当前用户信息
PUT    /api/v1/users/me               更新个人资料
GET    /api/v1/users/:id              获取用户信息
GET    /api/v1/users/search           搜索用户
GET    /api/v1/departments/tree       获取组织架构树

会话：
GET    /api/v1/conversations          获取会话列表
POST   /api/v1/conversations          创建会话（私聊/群聊）
GET    /api/v1/conversations/:id       获取会话详情
PUT    /api/v1/conversations/:id       更新会话（群名等）
POST   /api/v1/conversations/:id/members  添加成员
DELETE /api/v1/conversations/:id/members/:uid  移除成员
DELETE /api/v1/conversations/:id       解散会话

消息：
GET    /api/v1/conversations/:id/messages  获取历史消息（分页）
POST   /api/v1/conversations/:id/messages  发送消息
PUT    /api/v1/messages/:id           编辑消息
DELETE /api/v1/messages/:id           撤回消息
POST   /api/v1/messages/:id/read      标记已读

文件：
POST   /api/v1/files/upload           上传文件
GET    /api/v1/files/:id               下载文件

WebSocket：
WSS   /ws?token=xxx                  WebSocket 长连接
```

### 8.2 统一响应格式

```json
// 成功
{
    "code": 0,
    "message": "success",
    "data": { ... }
}

// 分页响应
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [...],
        "total": 100,
        "page": 1,
        "page_size": 20
    }
}

// 错误
{
    "code": 40001,
    "message": "用户名或密码错误",
    "data": null
}
```

### 8.3 错误码规范

| 错误码区间 | 含义 |
|-----------|------|
| 0 | 成功 |
| 40001-40099 | 参数/请求错误 |
| 40101-40199 | 认证错误 |
| 40301-40399 | 权限错误 |
| 40401-40499 | 资源不存在 |
| 50001-50099 | 服务器内部错误 |

---

## 附录

### A. 学习资源推荐

| 资源 | 链接 |
|------|------|
| Go Web 编程（书籍）| 人民邮电出版社 |
| 7天用Go从零实现分布式消息队列 | github.com/semodaddy/mqueue |
| 飞书技术博客 |.feishu.cn/tech |
| Slack 技术博客 | slack.engineering |
| gorilla/websocket 文档 | github.com/gorilla/websocket |

### B. 快速启动命令

```bash
# 克隆项目后
make deps          # 安装依赖
make migrate      # 运行数据库迁移
make run          # 启动服务
make test         # 运行测试
make docker       # 构建 Docker 镜像

# 或一键启动（需要 Docker）
docker compose -f deployments/docker/docker-compose.yaml up -d
```

### C. 开发环境要求

```
Go 1.21+
PostgreSQL 16+
Redis 7+
Kafka 3.x (第三阶段)
Docker & Docker Compose
```
