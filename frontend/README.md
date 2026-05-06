# WorkPal 前端

本目录是 WorkPal 的 React + Vite 前端。前端只访问 API 网关，开发代理由 `vite.config.ts` 维护：

- `/api/*` 转发到 `http://localhost:8080`
- `/ws` 转发到 `ws://localhost:8080`

## 启动方式

前端使用 npm 作为唯一包管理器，请保留 `package-lock.json`。

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

默认访问地址：

```text
http://localhost:3000
```

## 工作台模块

登录后进入工作台，当前模块包括：

- 总览
- 沟通
- 任务
- 日程
- 文件
- 通讯录

`/workspace/chat/:conversationId` 支持直达指定会话，其他模块使用 `/workspace/:section`。

## 目录结构

- `src/pages`：路由页面，包含登录、工作台、沟通和 404 页面。
- `src/components`：通用组件、工作台组件、聊天组件。
- `src/hooks`：工作台数据、业务操作、偏好设置、聊天控制器和 Toast 状态。
- `src/stores`：认证、WebSocket 消息、会话状态。
- `src/i18n`：中英文语言包。
- `src/styles`：按设计令牌、重置、组件、布局、页面和工具类拆分的样式。
- `src/test`：组件测试渲染工具。

## 关键交互

- 侧边栏固定高度，主内容区独立滚动。
- 侧边栏分组折叠，移动端可收起导航。
- Toast 操作反馈、危险操作确认框、文件上传进度条。
- 任务可拖拽变更状态，日程支持列表和日历视图。
- 文件支持图片/PDF 内联预览。
- `Ctrl/Cmd + K` 打开模块切换器，`Ctrl/Cmd + /` 打开偏好设置。

## 自检命令

```powershell
cd frontend
npm run lint
npm test
npm run build
```

当前项目还在 CI 中执行类型检查、lint、测试和构建。若在受限沙箱中遇到 esbuild `spawn EPERM`，需要在允许子进程启动的环境中复跑 Vitest 或 Vite build。
