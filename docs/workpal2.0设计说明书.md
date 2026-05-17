# WorkPal 2.0 设计说明书

> 基于 Jira Data Center + 飞书的融合重构方案
> 当前版本：WorkPal 1.0（Go + React 微服务办公协作平台）

---

## 一、产品愿景

WorkPal 2.0 将 Jira DC 的项目管理能力与飞书的协作沟通能力深度融合，打造**一站式团队工作空间**。核心命题是：**项目管理的结构化与团队沟通的即时性不再割裂**。

### 1.1 对标分析

| 维度 | Jira Data Center | 飞书 (Lark) | WorkPal 2.0 目标 |
|------|-----------------|------------|-----------------|
| 核心能力 | 事务追踪、敏捷管理、报表 | 即时通讯、文档协作、日程 | 两者融合 + 上下文串联 |
| 工作流 | 可配置状态流转 | 审批流引擎 | 统一可编排工作流 |
| 沟通模型 | 事务评论区 | 聊天/频道/视频会议 | 事务+频道双向关联 |
| 文档 | Confluence 集成 | 飞书文档 | 内置协作文档 |
| 开放生态 | Marketplace 插件 | 飞书开放平台 | Webhook + API + Bot |
| 部署形态 | 私有化 DC 集群 | SaaS 多租户 | 两者均支持 |

### 1.2 用户画像

- **研发团队**：项目经理用看板追踪进度，开发者用 IDE 插件更新状态
- **产品团队**：PRD 文档 + 需求评审 + 版本规划联动
- **运营团队**：飞书式的快节奏沟通 + Jira 式的工作项追踪
- **管理层**：跨项目报表、团队效能仪表盘

---

## 二、功能模块全景

### 2.1 模块总览

```
┌───────────────────────────────────────────────────────────┐
│                    WorkPal 2.0 功能地图                      │
├──────────────┬──────────────┬──────────────┬──────────────┤
│   项目空间    │   即时通讯    │   协作文档    │   日程会议    │
│ (Jira 基因)  │ (飞书 基因)   │ (飞书 基因)   │ (飞书 基因)   │
├──────────────┼──────────────┼──────────────┼──────────────┤
│  Issue 追踪  │  频道/私聊   │  在线文档     │  日历视图     │
│  敏捷看板    │  消息富文本  │  表格/知识库  │  视频会议     │
│  Epic/版本   │  群公告/文件 │  协同编辑     │  会议纪要     │
│  自定义工作流 │  表情/贴纸   │  模板市场     │  周期日程     │
├──────────────┼──────────────┼──────────────┼──────────────┤
│   审批中心    │   全局搜索    │   效能报表    │   AI 助手     │
│ (Jira+飞书)  │ (Jira+飞书)  │ (Jira 基因)  │   (新增)      │
├──────────────┼──────────────┼──────────────┼──────────────┤
│  自定义审批流 │  跨模块搜索  │  燃尽/燃起图 │  智能摘要     │
│  字段级权限   │  上下文检索  │  吞吐量看板  │  代码审查     │
│  操作审计     │  联邦搜索    │  团队效能    │  任务建议     │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

### 2.2 详细模块设计

#### 2.2.1 项目空间（对应 Jira Project）

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| Issue 类型体系 | Epic → Story → Task → Sub-task → Bug，支持自定义类型 | P0 |
| 敏捷看板 | Scrum 看板（Backlog/To Do/In Progress/Done）+ 泳道 | P0 |
| 甘特图/时间线 | 版本进度甘特图，依赖关系可视化 | P1 |
| 自定义工作流 | 状态流转规则、条件、校验器、后处理函数 | P0 |
| 字段配置 | 自定义字段类型（文本/下拉/日期/用户选择器/多选） | P0 |
| 版本规划 | 版本创建、Issue 分配、进度追踪 | P1 |
| 权限模型 | 项目角色 + 权限方案 = Jira 权限模式的三层模型 | P0 |
| 仪表盘 | 可配置小组件（图表/筛选器/统计卡片） | P1 |
| 通知方案 | 事件驱动的通知模板（关注/指派/评论/@提及） | P0 |

#### 2.2.2 即时通讯（对应飞书 IM）

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 频道体系 | 公开/私有频道，支持话题（Thread）聚合讨论 | P0 |
| 私聊/群聊 | 1对1、群组、跨项目频道 | P0 |
| 消息类型 | 文本/富文本/图片/文件/代码块/投票/@提及 | P0 |
| 消息操作 | 编辑、撤回、引用回复、表情回应、收藏、转发 | P0 |
| 消息搜索 | 全文搜索 + 时间范围 + 对象过滤 | P0 |
| 免打扰 | 按频道/用户/时间段设置 | P1 |
| 多端同步 | Web / Desktop / Mobile 推送与状态同步 | P1 |
| Issue 卡片 | 在聊天中内嵌 Issue 卡片（状态/指派人/优先级） | P0 |
| 频道机器人 | Webhook 推送 Issue 变更到频道 | P1 |

#### 2.2.3 协作文档（对应飞书文档）

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 在线文档 | 富文本协同编辑，支持 Markdown 快捷输入 | P0 |
| 数据表格 | 多维表格（对标飞书多维表格 / Notion Database） | P0 |
| 知识库 | 按项目空间组织树形目录 | P1 |
| 协同编辑 | OT/CRDT 实时同步，光标协作 | P0 |
| 版本历史 | 文档历史版本对比与回滚 | P1 |
| 模板市场 | 会议纪要、PRD、技术方案、周报模板 | P2 |
| 文档关联 | 文档可关联 Issue、频道、日程 | P0 |

#### 2.2.4 日程会议（对应飞书日历）

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 个人日历 | 日/周/月视图，拖拽创建 | P0 |
| 团队日历 | 多成员叠加显示，空闲/忙碌状态 | P0 |
| 视频会议 | 内置 WebRTC 音视频 + 屏幕共享 | P1 |
| 会议纪要 | 会议关联文档，自动生成纪要模板 | P2 |
| 周期日程 | 每日/每周/每月重复规则 | P1 |
| 日程提醒 | 多级提醒 + 飞书式强弹窗 | P1 |
| Issue 关联 | 日程可关联 Issue，评审会/站会等 | P0 |

#### 2.2.5 审批中心（融合创新）

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 审批模板 | 请假/报销/采购/发布/代码合并等 | P1 |
| 审批流引擎 | 自定义节点（审批人/条件分支/并行/加签/转交） | P1 |
| 移动审批 | 通知推送 + 一句话审批 | P1 |
| 审批关联 | 审批可关联 Issue / 文档 / 代码仓库 | P1 |

#### 2.2.6 AI 助手（新增能力）

| 功能点 | 说明 | 优先级 |
|--------|------|--------|
| 智能搜索 | 自然语言查询："我上周三改过哪些 bug？" | P1 |
| 任务总结 | 自动生成项目周报、Sprint 回顾 | P1 |
| 代码审查 | PR 关联 Issue，AI 生成 Review 摘要 | P2 |
| 智能分配 | 基于负载和历史推荐 Issue 指派人 | P2 |

---

## 三、技术架构设计

### 3.1 总体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                         WorkPal 2.0 技术架构                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    前端层 (React 18 + TypeScript)                │  │
│  │  ┌─────┐  ┌────────┐  ┌────────┐  ┌────────┐  ┌───────────┐  │  │
│  │  │项目 │  │  沟通   │  │  文档   │  │  日历   │  │  审批/AI   │  │
│  │  │空间 │  │  频道   │  │  知识库 │  │  会议   │  │  助手      │  │
│  │  └─────┘  └────────┘  └────────┘  └────────┘  └───────────┘  │  │
│  │        Zustand 状态管理 │ TipTap/ProseMirror 编辑器              │  │
│  │        WebSocket 实时连接池 │ React Query 服务端缓存             │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│              HTTP/2 + WebSocket + Server-Sent Events                 │
│                              │                                       │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                    API Gateway (:8080)                           │  │
│  │  统一鉴权 │ 限流熔断 │ 路由转发 │ 请求聚合 │ WebSocket 代理       │  │
│  │  OpenTelemetry 追踪注入 │ Prometheus 指标暴露                    │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              │                                       │
│     ┌────────┬────────┬────────┬────────┬────────┬──────────┐      │
│     ▼        ▼        ▼        ▼        ▼        ▼          ▼      │
│  ┌──────┐┌──────-┐┌──────-┐┌──────┐┌──────┐┌────────┐┌───────┐   │
│  │User  ││Project││  IM   ││ Docs ││Calen-││Approval││Notifi-│   │
│  │Service││Service││Service││Service││Service││Service││cation │   │
│  │:8081 ││ :8086 ││ :8082 ││ :8087 ││ :8088 ││ :8089 ││ :8090 │   │
│  └──┬───┘└──┬───┘└──┬───┘└──┬───┘└──┬───┘└───┬───┘└───┬───┘   │
│     │       │       │       │       │        │        │       │      │
│  ┌──┴───────┴───────┴───────┴───────┴────────┴────────┴────┐      │
│  │                    基础设施层                                │      │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────────────────┐ │      │
│  │  │ PostgreSQL │  │   Redis    │  │ Kafka / Redpanda      │ │      │
│  │  │ (独立库)   │  │ (缓存/注册)│  │ (事件总线)            │ │      │
│  │  └───────────┘  └───────────┘  └───────────────────────┘ │      │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────────────────┐ │      │
│  │  │ MinIO / S3 │  │Elasticsearch│  │ etcd (配置中心)       │ │      │
│  │  │ (文件存储) │  │ (全文搜索) │  │                    │ │      │
│  │  └───────────┘  └───────────┘  └───────────────────────┘ │      │
│  └──────────────────────────────────────────────────────────┘      │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    可观测性                                    │   │
│  │  Prometheus + Grafana + Jaeger + OpenTelemetry Collector      │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 微服务清单（1.0 → 2.0 演进）

| 服务 | 1.0 状态 | 2.0 状态 | 数据库 | 关键变化 |
|------|---------|---------|--------|---------|
| Gateway | 已有 | 增强 | 无状态 | 增加请求聚合、GraphQL Federation 网关 |
| User Service | 已有 | 增强 | `workpal_user` | 增加角色/权限体系、组织架构树 |
| IM Service | 已有 | 增强 | `workpal_im` | 增加频道体系、话题线程、Issue 卡片消息 |
| File Service | 已有 | 增强 | `workpal_file` | 文件版本管理、文档附件关联 |
| Search Service | 已有 (Bleve) | 升级 | Elasticsearch | 替换为 ES，支持跨模块联邦搜索 |
| Workspace Service | 已有 | 拆分 | — | 拆分为 Task Service + Calendar Service |
| **Project Service** | — | **新增** | `workpal_project` | Issue 体系、工作流、看板、版本、权限 |
| **Docs Service** | — | **新增** | `workpal_docs` | 文档 CRDT 引擎、协同编辑、知识库 |
| **Approval Service** | — | **新增** | `workpal_approval` | 审批流引擎、模板、实例 |
| **Notification Service** | — | **新增** | `workpal_notification` | 多通道通知（站内/推送/邮件/Webhook） |
| **AI Service** | — | **新增** | 无状态 | LLM 网关，对接模型推理 |

### 3.3 关键架构决策

#### 3.3.1 事件驱动架构

```
┌──────────┐    ┌──────────┐    ┌───────────────┐
│ 业务服务  │───▶│  Kafka    │───▶│ 消费者服务      │
│ (Producer)│    │ (Topic)   │    │ (Consumer)     │
└──────────┘    └──────────┘    └───────────────┘
                                       │
     ┌─────────────────────────────────┤
     ▼           ▼           ▼          ▼
 Notification  Search    Analytics    Audit
  Service     Service    Service     Service
```

核心事件主题：

| Topic | 生产者 | 消费者 | 说明 |
|-------|--------|--------|------|
| `issue.events` | Project Service | Notification, Search, Analytics | Issue 创建/更新/删除/状态变更 |
| `message.events` | IM Service | Notification, Search | 消息发送/编辑/撤回 |
| `document.events` | Docs Service | Notification, Search | 文档创建/编辑/评论 |
| `approval.events` | Approval Service | Notification | 审批提交/通过/驳回/转交 |
| `calendar.events` | Calendar Service | Notification | 日程创建/提醒/变更 |
| `notification.commands` | 各服务 | Notification Service | 统一通知发送指令 |

#### 3.3.2 搜索架构升级

```
1.0 架构：Bleve（进程内嵌入，无分布式能力）
2.0 架构：Elasticsearch 集群（分布式索引 + 跨模块联邦搜索）

索引结构：
- 统一索引：workpal-unified-{date}（按天滚动）
- 字段：entity_type, entity_id, project_id, title, content, 
        creator_id, assignee_id, status, created_at, updated_at
- 权限过滤：searchable_by 字段记录可见用户/角色列表
```

#### 3.3.3 协同编辑引擎

```
技术选型：Yjs (CRDT) + y-websocket
替代方案：Operational Transform (ShareJS 风格)

原因：
- CRDT 不需要中央协调服务器，离线编辑更友好
- Yjs 生态成熟（Quill / ProseMirror / Monaco 绑定）
- 更简单的冲突解决（数学保证最终一致）
```

#### 3.3.4 工作流引擎

```
借鉴 Jira 工作流设计模型：
- 工作流定义：JSON/YAML DSL 描述状态和转换
- 状态机执行：用 go-statemachine 或自行实现
- 条件评估：Groovy/JS 脚本沙箱 或 结构化条件 DSL

示例 DSL：
workflow:
  name: "研发工作流"
  statuses: [Open, In Progress, In Review, Done, Reopened]
  transitions:
    - from: Open
      to: In Progress
      conditions:
        - field: assignee
          operator: not_null
      validators:
        - class: PermissionValidator
          args: {role: "developer"}
      post_functions:
        - class: UpdateHistory
        - class: NotificationEvent
```

#### 3.3.5 权限模型

```
三层权限模型（继承 Jira DC 思路）：

第一层 - 全局权限：
  - 系统管理员、用户管理员、全局浏览

第二层 - 项目角色：
  - 管理员、开发者、观察者、报告者
  - 支持自定义角色 + 权限组合

第三层 - Issue 安全级别：
  - 按 Issue 级别隐藏（只读/不可见）
  - 适用于外包场景、机密项目

实现方案：
  - RBAC 模型存储在 PostgreSQL
  - 鉴权中间件缓存热数据到 Redis
  - 前端通过 /api/v2/permissions/check 批量校验
```

### 3.4 技术栈详细

| 层级 | 1.0 | 2.0 | 变更原因 |
|------|-----|-----|---------|
| 后端语言 | Go 1.22 | Go 1.23+ | 迭代更新 |
| 前端框架 | React 18 + Vite 5 | React 19 + Vite 6 | 服务端组件支持 |
| 状态管理 | Zustand 4.5 | Zustand 5 | 迭代更新 |
| 富文本编辑器 | — | TipTap (ProseMirror) | 文档/评论区 |
| 实时通信 | WebSocket (gorilla/ws) | WebSocket + SSE | 文档/通知推送 |
| ORM | GORM | GORM + sqlx | 关键查询性能优化 |
| 消息队列 | Redis Streams | Kafka / Redpanda | 持久化、重放、分区 |
| 搜索引擎 | Bleve | Elasticsearch 8.x | 分布式、联邦搜索 |
| 文件存储 | MinIO | MinIO（兼容 S3） | 不变 |
| 配置中心 | YAML 文件 | etcd + 热更新 | 动态配置 |
| 协同编辑 | — | Yjs (CRDT) | 文档协同 |
| API 文档 | — | OpenAPI 3.1 (swaggo) | 自动生成 |
| CI/CD | GitHub Actions | GitHub Actions + ArgoCD | GitOps |

---

## 四、数据模型设计

### 4.1 核心聚合根

```
┌──────────────────────────────────────────────────────────────────┐
│                        数据模型关系图                               │
├──────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌──────────┐     ┌──────────────┐     ┌─────────────────────┐   │
│  │  Project │────▶│    Issue     │────▶│  WorkflowTransition  │   │
│  │          │     │              │     │  IssueLink           │   │
│  │  Versions│     │  Comments    │     │  TimeTracking        │   │
│  │  Boards  │     │  Attachments │     │  CustomFieldValue    │   │
│  └──────────┘     └──────┬───────┘     └─────────────────────┘   │
│                          │                                         │
│          ┌───────────────┼───────────────┐                        │
│          ▼               ▼               ▼                        │
│  ┌─────────────┐ ┌──────────────┐ ┌──────────────┐               │
│  │  Channel    │ │  Document    │ │  Calendar    │               │
│  │  ─────────  │ │  ─────────   │ │  ─────────   │               │
│  │  Messages   │ │  Pages       │ │  Events      │               │
│  │  Threads    │ │  Blocks      │ │  Recurrences │               │
│  │  Reactions  │ │  Revisions   │ │  Attendees   │               │
│  └──────┬──────┘ └──────┬───────┘ └──────┬───────┘               │
│         │               │                │                        │
│         └───────────────┼────────────────┘                        │
│                         │                                         │
│                ┌────────┴────────┐                                │
│                │  Association    │  (多对多关联表)                  │
│                │  ─────────────  │                                │
│                │  source_type    │                                │
│                │  source_id      │                                │
│                │  target_type    │                                │
│                │  target_id      │                                │
│                │  link_type      │                                │
│                └─────────────────┘                                │
└──────────────────────────────────────────────────────────────────┘
```

### 4.2 Project Service 数据模型

```sql
-- 项目
CREATE TABLE projects (
    id          BIGSERIAL PRIMARY KEY,
    key         VARCHAR(10) NOT NULL UNIQUE,  -- 如 "WP", "MKT"
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    lead_id     BIGINT NOT NULL,
    icon        VARCHAR(50),
    category    VARCHAR(50),                  -- software / business / marketing
    is_archived BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Issue 类型
CREATE TABLE issue_types (
    id           BIGSERIAL PRIMARY KEY,
    project_id   BIGINT NOT NULL REFERENCES projects(id),
    name         VARCHAR(100) NOT NULL,        -- Epic / Story / Task / Bug
    description  TEXT,
    icon_url     VARCHAR(255),
    hierarchy_level INT DEFAULT 0,            -- 层级（0=Epic, 1=Story, ...）
    is_standard  BOOLEAN DEFAULT TRUE
);

-- 工作流
CREATE TABLE workflows (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    dsl_definition JSONB NOT NULL,             -- 工作流 DSL
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Issue 主表
CREATE TABLE issues (
    id             BIGSERIAL PRIMARY KEY,
    project_id     BIGINT NOT NULL REFERENCES projects(id),
    issue_type_id  BIGINT NOT NULL REFERENCES issue_types(id),
    parent_id      BIGINT REFERENCES issues(id),  -- 父 Issue
    key            VARCHAR(50) NOT NULL UNIQUE,    -- WP-123
    summary        VARCHAR(500) NOT NULL,
    description    TEXT,
    status         VARCHAR(50) NOT NULL,
    priority       VARCHAR(20) NOT NULL DEFAULT 'Medium',
    assignee_id    BIGINT,
    reporter_id    BIGINT NOT NULL,
    due_date       DATE,
    story_points   DECIMAL(5,1),
    resolution     VARCHAR(50),
    sprint_id      BIGINT,
    version_ids    BIGINT[],                       -- 影响的版本
    fix_version_ids BIGINT[],                      -- 修复的版本
    time_estimate  INT,                            -- 预估工时（分钟）
    time_spent     INT DEFAULT 0,                  -- 实际工时（分钟）
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

-- 看板
CREATE TABLE boards (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT NOT NULL REFERENCES projects(id),
    name        VARCHAR(255) NOT NULL,
    board_type  VARCHAR(20) NOT NULL,               -- scrum / kanban
    config      JSONB NOT NULL,                     -- 列配置、泳道配置、过滤条件
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 版本
CREATE TABLE versions (
    id           BIGSERIAL PRIMARY KEY,
    project_id   BIGINT NOT NULL REFERENCES projects(id),
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    start_date   DATE,
    release_date DATE,
    is_archived  BOOLEAN DEFAULT FALSE,
    is_released  BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Issue 变更历史
CREATE TABLE issue_changelogs (
    id          BIGSERIAL PRIMARY KEY,
    issue_id    BIGINT NOT NULL REFERENCES issues(id),
    field       VARCHAR(100) NOT NULL,
    old_value   TEXT,
    new_value   TEXT,
    changed_by  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 自定义字段值（EAV 模型）
CREATE TABLE custom_field_values (
    id           BIGSERIAL PRIMARY KEY,
    issue_id     BIGINT NOT NULL REFERENCES issues(id),
    field_id     BIGINT NOT NULL REFERENCES custom_field_defs(id),
    value_text   TEXT,
    value_number DECIMAL,
    value_date   DATE,
    value_json   JSONB
);
```

### 4.3 Docs Service 数据模型

```sql
-- 文档
CREATE TABLE documents (
    id          BIGSERIAL PRIMARY KEY,
    project_id  BIGINT,
    parent_id   BIGINT REFERENCES documents(id),     -- 知识库目录树
    title       VARCHAR(500) NOT NULL,
    created_by  BIGINT NOT NULL,
    updated_by  BIGINT NOT NULL,
    is_folder   BOOLEAN DEFAULT FALSE,
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 文档内容（版本化）
CREATE TABLE document_revisions (
    id            BIGSERIAL PRIMARY KEY,
    document_id   BIGINT NOT NULL REFERENCES documents(id),
    version       INT NOT NULL,
    content       JSONB NOT NULL,                  -- TipTap/ProseMirror JSON
    created_by    BIGINT NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 协同编辑会话（Redis / 内存表）
-- 使用 Yjs 的内存协议，通过 WebSocket 同步
-- 落盘时写入 document_revisions
```

### 4.4 IM Service 增强模型

```sql
-- 频道（扩展会话概念）
CREATE TABLE channels (
    id           BIGSERIAL PRIMARY KEY,
    project_id   BIGINT,
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    channel_type VARCHAR(20) NOT NULL,      -- public / private / direct_message
    created_by   BIGINT NOT NULL,
    is_archived  BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- 话题线程（Thread）
CREATE TABLE threads (
    id            BIGSERIAL PRIMARY KEY,
    channel_id    BIGINT NOT NULL REFERENCES channels(id),
    parent_msg_id BIGINT NOT NULL REFERENCES messages(id),
    title         VARCHAR(500),
    reply_count   INT DEFAULT 0,
    last_reply_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 消息增强
ALTER TABLE messages
    ADD COLUMN thread_id BIGINT REFERENCES threads(id),
    ADD COLUMN message_type VARCHAR(30) DEFAULT 'text',   -- text / issue_card / file / system
    ADD COLUMN metadata JSONB;                            -- Issue 卡片、文件预览等扩展数据
```

---

## 五、UI/UX 设计原则

### 5.1 设计语言

| 原则 | Jira DC 参考 | 飞书参考 | WorkPal 2.0 策略 |
|------|-------------|---------|-----------------|
| 信息密度 | 高密度信息展示 | 适度留白 | 提供紧凑/舒适/宽松三种密度 |
| 导航模型 | 左侧项目树 + 顶部Tab | 左侧一级 + 二级 | 左侧双栏（项目空间 → 模块） |
| 操作效率 | 键盘快捷键(???) | 快捷指令(⌘K) | 全局命令面板(⌘K) + 快捷键体系 |
| 色彩体系 | 蓝灰主色调（工具体验） | 蓝白主色调（办公体验） | 中性基色 + 功能色语义化 |
| 视觉语言 | 紧凑表格/卡片 | 圆角卡片/Fluent | 圆角8px + 阴影层级 + 功能色强调 |

### 5.2 布局框架

```
┌──────────────────────────────────────────────────────────────┐
│  ┌───────────┐  ┌─────────────────────────────────────────┐ │
│  │           │  │  顶部导航栏（项目切换/搜索/通知/个人中心） │ │
│  │  左侧一级  │  ├─────────────────────────────────────────┤ │
│  │  导航      │  │                                         │ │
│  │           │  │  内容区                                   │ │
│  │  · 首页   │  │  (看板/列表/文档/日历/频道)              │ │
│  │  · 项目   │  │                                         │ │
│  │  · 频道   │  │                                         │ │
│  │  · 文档   │  │                                         │ │
│  │  · 日历   │  │                                         │ │
│  │  · 审批   │  │                                         │ │
│  │           │  │                                         │ │
│  └───────────┘  └─────────────────────────────────────────┘ │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  底部状态栏（连接状态/最后同步时间/快捷操作）             │ │
│  └─────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### 5.3 关键交互

| 交互场景 | 设计方案 |
|---------|---------|
| Issue 拖拽排序 | 看板列内拖拽，自动更新状态+排序，乐观更新+失败回滚 |
| 消息中创建 Issue | 长按消息→"转为 Issue"，标题预填消息内容，自动关联频道 |
| 文档 @提及 Issue | 输入 `WP-` 触发 Issue 搜索，自动生成卡片链接 |
| 全局搜索 | ⌘K 唤起，跨模块搜索（Issue/文档/消息/日程），按相关度排序 |
| 通知聚合 | 同 Issue 多条更新合并为一条摘要通知 |
| 多 Tab 编辑 | 支持多个 Issue/文档页签并行编辑，防止丢失 |

### 5.4 深浅色主题

继承 1.0 的主题切换能力，2.0 增强：
- **深色模式**：富文本编辑器（TipTap）内置暗色 CSS 变量适配
- **系统跟随**：`prefers-color-scheme` 媒体查询自动切换
- **代码块**：深/浅色分别使用对应语法高亮主题
- **数据可视化**：ECharts/Recharts 图表双主题配置

---

## 六、模块交互设计

### 6.1 核心交互链路

#### 6.1.1 Issue 创建全链路

```
用户操作：在项目看板点击"创建 Issue"
         │
         ▼
前端：弹出 Issue 创建表单（类型/摘要/描述/指派人/优先级/截止日期/父Issue）
         │
         ▼
Gateway：POST /api/v2/projects/{projectId}/issues
         │
         ▼
Project Service：
  1. 校验项目权限（创建 Issue 权限）
  2. 校验工作流（当前 Issue 类型允许的初始状态）
  3. 生成 Issue Key（WP-123）
  4. 写入 issues 表
  5. 写入自定义字段值（custom_field_values）
  6. 写入变更历史（issue_changelogs）
  7. 发布 Kafka 事件 {type: "issue.created", ...}
         │
         ▼
事件消费者：
  ┌─→ Notification Service：通知指派人 + 关注者
  ├─→ Search Service：索引 Issue 内容
  └─→ Analytics Service：更新项目统计
```

#### 6.1.2 聊天消息转 Issue

```
用户操作：在频道中右键消息 → "转为 Issue"
         │
         ▼
前端：
  1. 弹出 Issue 创建表单，预填：
     - 摘要 = 消息内容截断
     - 描述 = 完整消息 + 发送者 + 时间
     - 频道关联(metadata.channel_id)
  2. 用户补充类型/优先级/指派人后提交
         │
         ▼
Project Service：
  1. 创建 Issue（同上述流程）
  2. 写入 association 表：
     {source_type: "issue", source_id: 123, 
      target_type: "message", target_id: 456,
      link_type: "derived_from"}
         │
         ▼
IM Service：
  1. 在原消息下方插入系统消息："[已转为 Issue WP-123](链接)"
  2. 发送 Issue 卡片消息到频道
```

#### 6.1.3 文档协同编辑

```
用户 A 打开文档：GET /api/v2/docs/{docId}
         │
         ▼
Docs Service：
  1. 加载最新 document_revisions.content
  2. 建立 WebSocket 连接到 y-websocket 房间
         │
         ▼
用户 B 同时编辑：
  WebSocket 广播 CRDT 增量（Yjs update）
         │
         ▼
实时展示：
  - 用户 A 的编辑器实时显示 B 的修改
  - 通过 Yjs Awareness 显示远程光标位置
         │
         ▼
定时保存（debounce 2s 或 1000 ops）：
  Docs Service 写入新的 document_revisions 版本
         │
         ▼
发布事件：
  {type: "document.updated", docId, version, userId}
  → Search Service 更新索引
  → Notification Service 通知文档订阅者
```

### 6.2 事件流全景图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Kafka 事件流全景                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Project Service ──┬── issue.created ─────→ Notify / Search / Analytics │
│                    ├── issue.updated ─────→ Notify / Search / Analytics │
│                    ├── issue.commented ───→ Notify / Search             │
│                    ├── issue.status_changed → Notify / Analytics        │
│                    └── issue.assigned ────→ Notify                      │
│                                                                          │
│  IM Service ───────┬── message.sent ──────→ Notify / Search             │
│                    ├── message.edited ────→ Search                      │
│                    └── message.deleted ───→ Search                      │
│                                                                          │
│  Docs Service ─────┬── document.updated ──→ Search / Notify             │
│                    └── document.commented → Notify / Search             │
│                                                                          │
│  Calendar Service ─┬── event.created ─────→ Notify                      │
│                    ├── event.reminder ────→ Notify (push/email)         │
│                    └── event.updated ─────→ Notify                      │
│                                                                          │
│  Approval Service ─┬── approval.submitted ─→ Notify                     │
│                    ├── approval.approved ──→ Notify                     │
│                    └── approval.rejected ──→ Notify                     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 七、部署架构

### 7.1 部署拓扑

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster (workpal)                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────────────────┐  ┌─────────────────────────────┐   │
│  │  前端 Pod (nginx:alpine)     │  │  可观测性 Stack               │   │
│  │  replicas: 2                │  │  Prometheus + Grafana        │   │
│  │  HPA: CPU 70%               │  │  Jaeger + OTel Collector     │   │
│  └─────────────────────────────┘  └─────────────────────────────┘   │
│                                                                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │ Gateway × 2 │ │User Svc × 2 │ │Project × 2  │ │ IM Svc × 2  │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │Docs Svc × 2 │ │Calendar × 2 │ │Approval × 2 │ │Notify × 2   │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │                  有状态服务 (StatefulSet)                       │   │
│  │  ┌──────────────────┐  ┌──────────────┐  ┌──────────────────┐│   │
│  │  │ PostgreSQL 16    │  │ Redis 7      │  │ Kafka/Redpanda   ││   │
│  │  │ 主库 + 只读副本  │  │ Cluster 3 节 │  │ 3 Broker         ││   │
│  │  │ 备份: WAL-G + S3  │  │ Sentinel    │  │ 保留: 7 天       ││   │
│  │  └──────────────────┘  └──────────────┘  └──────────────────┘│   │
│  │  ┌──────────────────┐  ┌──────────────┐  ┌──────────────────┐│   │
│  │  │Elasticsearch 8   │  │ MinIO 集群   │  │ etcd 集群        ││   │
│  │  │ 3 节点           │  │ 纠删码模式   │  │ 3 节点           ││   │
│  │  └──────────────────┘  └──────────────┘  └──────────────────┘│   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.2 部署形态

| 形态 | 适用场景 | 组件差异 |
|------|---------|---------|
| **私有化部署** | Jira DC 目标客户、政企 | 全量 K8s 部署，MinIO 替代 S3，内置 OAuth/LDAP |
| **SaaS 多租户** | 飞书目标客户、中小团队 | 共享 K8s 集群，租户隔离（DB Schema / Collection），按量计费 |
| **单机 Docker Compose** | 开发者本地、POC 验证 | 所有服务合并镜像，2C4G 可跑通，用于学习和试用 |

---

## 八、实施路线图

### 第一阶段（MVP — 6 周）

```
目标：跑通核心链路，可以创建项目、Issue、用看板追踪、频道里聊天

Week 1-2: 数据模型 + Project Service 骨架
  - projects / issues / workflows 表 + 迁移
  - 创建项目、创建 Issue、修改状态 API
  - 前端项目列表页 + Issue 详情页（基础表单）

Week 3-4: 看板 + 频道
  - 看板列配置 + 拖拽排序
  - 频道创建 + 消息收发
  - Issue 卡片消息类型

Week 5-6: 搜索 + 通知
  - Elasticsearch 索引管道
  - 通知中心（站内通知）
  - 关联功能（Issue ↔ 频道消息）
```

### 第二阶段（增强 — 6 周）

```
Week 7-8: 工作流引擎 + 权限
  - 工作流 DSL 解析与执行
  - RBAC 权限模型
  - 自定义字段体系

Week 9-10: 文档协作
  - 文档 CRUD + 富文本编辑器
  - Yjs 协同编辑
  - 知识库目录树

Week 11-12: 日历 + 审批
  - 个人/团队日历
  - 审批流引擎
  - 视频会议基础
```

### 第三阶段（完善 — 6 周）

```
Week 13-14: AI 助手
  - 自然语言搜索
  - 智能摘要（项目周报）

Week 15-16: 报表 + 效能
  - 燃尽图 / 吞吐量报表
  - 团队效能仪表盘

Week 17-18: 性能优化 + 多端
  - Electron Desktop 应用
  - 移动端适配
  - 性能压测 + 优化
```

---

## 九、与 WorkPal 1.0 的兼容与迁移

### 9.1 数据迁移

| 1.0 实体 | 2.0 实体 | 迁移方式 |
|----------|---------|---------|
| `tasks` | `issues` (type=task) | 数据转换脚本，补充 project_id / issue_type_id / key |
| `schedule_events` | `Calendar Service` 的 `events` | 迁移到新表的 events |
| `conversations` | `channels` (type=direct_message) | 保留兼容，新频道用 channels 表 |
| `messages` | `messages` (增加 thread_id / message_type) | ALTER TABLE 增量迁移 |
| `files` | `files` (增加 document_id 关联) | 兼容迁移 |

### 9.2 API 兼容

- 1.0 的所有 `/api/v1/*` 路径在 Gateway 层保留转发
- 2.0 新接口统一使用 `/api/v2/*`
- 给前端一个版本的过渡期（两个版本的前端可以同时运行）

### 9.3 前端渐进式升级

- Phase 1：新模块（项目空间/文档）使用新的路由和组件
- Phase 2：沟通/日历模块逐步迁移到新的 UI 框架
- Phase 3：完全移除 1.0 遗留组件

---

## 十、里程碑交付物

| 阶段 | 交付物 | 验收标准 |
|------|--------|---------|
| MVP | 项目空间 + 看板 + 频道聊天 | 可创建项目→创建 Issue→拖拽看板→在频道讨论 Issue |
| 增强 | 工作流 + 文档 + 日历 | 可定义工作流→书写 PRD 文档→安排评审日程 |
| 完善 | AI + 报表 + 多端 | 可自然语言搜索→查看团队效能→在桌面和手机上使用 |

---

> 本文档定义了 WorkPal 2.0 的完整设计方案，以 Jira Data Center 的项目管理能力与飞书的协作沟通能力为蓝图，构建从"会聊天的任务板"到"以项目为中心的工作空间"的跨越式演进。
