# WorkPal Frontend

这是 WorkPal 的 React + Vite 前端。

如果你要从零启动整套项目，请优先看仓库根目录的 [README.md](../README.md)。本文件主要说明前端自己的启动方式、代理关系和当前页面结构。

## 环境要求

- Node.js 18+
- npm

## 启动前提

前端默认依赖：

- 本地后端运行在 `http://localhost:8080`
- Vite dev server 运行在 `http://localhost:3000`

代理规则定义在 [vite.config.ts](vite.config.ts)：

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

## 启动前端

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

打开：

```text
http://localhost:3000
```

## 当前页面结构

登录后会进入多板块工作台，而不是单一聊天页：

- `Overview / 总览`
- `Chat / 沟通`
- `Tasks / 任务`
- `Schedule / 日程`
- `Files / 文件`
- `Directory / 通讯录`

同时支持以下偏好设置：

- `English / 简体中文`
- 浅色 / 深色主题
- 消息提示音开关
- 舒适 / 紧凑密度

## 预置验收账号

这些账号由后端开发模式自动确保存在，前端登录页也会直接展示：

| 用户名 | 密码 |
|---|---|
| `admin` | `admin123` |
| `emma.chen` | `workpal123` |
| `liam.wang` | `workpal123` |
| `sofia.zhao` | `workpal123` |

## 哪些模块是后端联调，哪些是前端演示

### 直接依赖后端

- 登录
- 当前用户 / 用户列表
- 私聊 / 群聊
- 消息发送
- 消息搜索
- WebSocket 实时状态

### 当前为前端预置协作演示

- 总览摘要
- 任务看板
- 日程面板
- 文件与知识面板

这些模块的存在是为了让项目在验收时具备更完整的办公协作平台形态，不再只有沟通板块。

## 常用脚本

```powershell
cd frontend
npm run dev -- --host 127.0.0.1
npm test
npm run build
```

## E2E 冒烟

要求前后端都已启动：

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

这个脚本会验证预置账号登录、工作台导航、语言切换、通讯录和聊天入口。
