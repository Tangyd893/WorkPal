# Testing

## 目录结构

```
testing/
├── README.md          # 本文件
├── e2e/
│   └── playwright.mjs # E2E 测试脚本
└── specs/            # API 测试规格文件（TODO）
```

## 启动测试环境

```bash
# 1. 启动基础设施（PostgreSQL + Redis + MinIO）
docker compose -f docker-compose.yaml up -d

# 2. 启动后端
cd backend
GOTOOLCHAIN=local go run cmd/server/main.go &

# 3. 启动前端
cd frontend && npm run dev &
```

## E2E 测试（Playwright）

```bash
cd testing/e2e
node playwright.mjs
```

需要先安装 Playwright 浏览器：
```bash
cd frontend
npx playwright install chromium
```

## 单元测试（Go）

```bash
cd backend

# 运行所有测试
go test -v ./...

# 带 race 检测
go test -race ./...

# 只跑 service 层测试
go test -race ./internal/im/service/...
go test -race ./internal/user/service/...
```

## 测试覆盖率

```bash
cd backend
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## CI

GitHub Actions 自动在每次 PR/Push 时运行：
- `go test -race ./...`
- `golangci-lint`
- 前端 `tsc --noEmit`
