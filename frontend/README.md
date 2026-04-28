# WorkPal Frontend

这是 WorkPal 的 React + Vite 前端。

如果你要从零启动整套项目，请优先看仓库根目录的 [README.md](../README.md)。  
这里重点说明前端自己的启动方式和依赖关系。

## 环境要求

- Node.js 18+
- npm

## 启动前提

前端默认依赖：

- 本地后端运行在 `http://localhost:8080`
- Vite dev server 运行在 `http://localhost:3000`

当前代理规则在 [vite.config.ts](vite.config.ts)：

- `/api/*` -> `http://localhost:8080`
- `/ws` -> `ws://localhost:8080`

## 启动前端

```powershell
cd frontend
npm ci
npm run dev
```

启动后打开：

```text
http://localhost:3000
```

## 重要说明

- 当前前端只有登录页，没有注册页
- 在默认开发配置下，后端启动时会自动确保默认管理员账号存在：
  - 用户名：`admin`
  - 密码：`admin123`
- 登录成功后会进入聊天页

建议先直接用这组账号登录。

如果你需要额外测试账号，再手动创建：

```powershell
$suffix = Get-Date -Format 'MMddHHmmss'
$username = "debug$suffix"

$body = @{
  username = $username
  password = "pass123456"
  nickname = "Debug User"
  email = "$username@example.com"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/register" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

## 常用脚本

```powershell
cd frontend
npm run dev
npm run build
npm test
```

## 调试建议

先确认下面两件事再看前端问题：

1. `http://localhost:8080/health` 返回 200
2. 你能用 `admin / admin123` 登录，或者已经手动创建了一个可登录账号

如果后端没起来，前端页可以打开，但 API 请求和聊天功能一定不通。
