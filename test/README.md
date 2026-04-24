# 测试目录

## E2E 测试（Playwright）

依赖：`npx playwright install chromium`

```bash
# 启动服务
cd backend/deployments/docker && docker compose up -d
cd backend && GOTOOLCHAIN=local go run cmd/server/main.go &
cd frontend && npm run dev &

# 运行 E2E（需要先安装 playwright）
npx playwright install chromium
node test/e2e/playwright.mjs
```

## 单元测试（Go）

```bash
cd backend
GOTOOLCHAIN=local go test -v ./...
```
