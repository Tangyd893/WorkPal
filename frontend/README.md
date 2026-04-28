# WorkPal 前端

React 18 + Vite + TypeScript 企业协作平台前端。

## 技术栈

| 类别 | 技术 |
|------|------|
| 框架 | React 18 + TypeScript |
| 构建 | Vite 5 |
| 路由 | React Router v6 |
| 状态管理 | Zustand（全局状态）+ React State（局部） |
| HTTP | Axios |
| 样式 | CSS（自定义变量，参考 antd 设计规范） |

## 目录结构

```
frontend/
├── src/
│   ├── api/           # Axios 封装，统一请求拦截
│   ├── components/     # 公共组件（Layout 等）
│   ├── hooks/         # 自定义 Hook（useAuthStore 等）
│   ├── pages/         # 页面组件
│   ├── styles/        # 全局样式
│   ├── App.tsx        # 根组件
│   └── main.tsx       # 入口
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## 快速开始

```bash
# 安装依赖
npm ci

# 开发模式
npm run dev    # 启动在 http://localhost:3000

# 构建生产版本
npm run build

# 单元测试
npm test
```

## 环境变量

项目根目录创建 `.env.local`：

```env
VITE_API_BASE_URL=/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

## 接口代理

Vite 配置了 `/api` 到后端 `localhost:8080` 的代理，开发环境无需处理跨域。

## API 响应约定

后端统一返回 `{ code, message, data }`。`src/api/request.ts` 会在 `code === 0` 时自动返回 `data`，非 0 时抛出错误；页面代码不需要再手动读取外层 `data`。
