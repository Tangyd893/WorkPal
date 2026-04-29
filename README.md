# WorkPal

[中文说明](README-cn.md) | English

WorkPal is a Go + React office collaboration demo that now runs as a real microservice project rather than a hybrid monolith. The backend has a single public entry at the API gateway, while domain services own their own data and runtime boundaries.

## What the project includes

- seeded acceptance accounts for admin and employees
- bilingual UI: `English / 简体中文`
- light and dark theme, message sound toggle, density toggle
- overview, chat, tasks, schedule, files, and directory modules
- direct chat, group chat, group announcement, and group files
- backend-backed tasks, schedule, personal files, and directory search
- API gateway, domain services, Redis Streams, and Bleve search

## Stack

- Backend: Go, Gin, GORM, PostgreSQL, Redis, Redis Streams, Bleve
- Frontend: React 18, Vite, TypeScript, Zustand
- File storage: MinIO with local-file fallback
- Realtime: WebSocket through the IM service

## Prerequisites

- Go `1.22+`
- Node.js `18+`
- npm
- Docker Desktop or Docker Engine

Confirm Docker is actually running before you start:

```powershell
docker version
```

Continue only when the output contains both `Client` and `Server`.

## Ports

| Component | URL | Notes |
| --- | --- | --- |
| Frontend | `http://localhost:3000` | Vite dev server |
| API Gateway | `http://localhost:8080` | only backend URL the frontend uses |
| User Service | `http://localhost:8081` | auth, users, departments, employees |
| IM Service | `http://localhost:8082` | conversations, messages, WebSocket |
| File Service | `http://localhost:8083` | personal files and group files |
| Search Service | `http://localhost:8084` | message search and indexing |
| Workspace Service | `http://localhost:8085` | tasks and schedule |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | no password by default |
| MinIO API | `http://localhost:9000` | object storage |
| MinIO Console | `http://localhost:9001` | `workpal / workpal123456` |

## Microservice topology

The backend is organized around one gateway and five domain services:

| Service | Storage boundary | Main responsibility |
| --- | --- | --- |
| Gateway | stateless | single ingress, reverse proxy, rate limit, aggregated health |
| User Service | `workpal_user` | login, users, departments, employees, dev seed data |
| IM Service | `workpal_im` | direct chat, group chat, messages, announcements, WebSocket |
| File Service | `workpal_file` | file metadata, upload, share, delete |
| Search Service | Bleve + Redis Streams | message indexing and search |
| Workspace Service | `workpal_workspace` | tasks and schedule |

The gateway health endpoint is aggregated. `GET /health` on port `8080` checks every downstream service, so it reflects the real backend state rather than only the gateway process.

## Quick start

### 1. Start the full stack with Docker Compose

From the repo root:

```powershell
docker compose -f docker/docker-compose.yaml build
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

Expected result:

- `postgres`, `redis`, and `minio` are `Up` or `healthy`
- `gateway`, `user-service`, `im-service`, `file-service`, `search-service`, and `workspace-service` are `Up`

On first boot, the backend services automatically create the service-owned databases they need:

- `workpal_user`
- `workpal_im`
- `workpal_file`
- `workpal_workspace`

### 2. Start services from source instead of Docker

If you want to debug service processes manually, start infrastructure only:

```powershell
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

Then open separate terminals and run:

```powershell
cd backend
go run ./cmd/user-service
```

```powershell
cd backend
go run ./cmd/im-service
```

```powershell
cd backend
go run ./cmd/file-service
```

```powershell
cd backend
go run ./cmd/search-service
```

```powershell
cd backend
go run ./cmd/workspace-service
```

```powershell
cd backend
go run ./cmd/gateway
```

### 3. Start the frontend

Open another terminal:

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

Then open:

```text
http://localhost:3000
```

## Acceptance accounts

In the default development mode, User Service automatically ensures these accounts exist:

| Role | Username | Password |
| --- | --- | --- |
| Admin | `admin` | `admin123` |
| Employee | `emma.chen` | `workpal123` |
| Employee | `liam.wang` | `workpal123` |
| Employee | `sofia.zhao` | `workpal123` |

Seeded department and employee profile data is also created automatically, so directory filtering and search work out of the box.

## Quick verification

### Gateway health

```powershell
Invoke-RestMethod http://localhost:8080/health
```

Expected result:

- HTTP `200`
- `status` is `ok`
- `components` lists the five downstream services

### Login API

```powershell
$body = @{
  username = "admin"
  password = "admin123"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

Expected result:

- `code` is `0`
- `data.token` exists

### Frontend login

Open `http://localhost:3000`, log in with `admin / admin123`, and confirm the app redirects to `/workspace/overview`.

## Tests

### Backend

```powershell
cd backend
go test ./...
```

### Frontend

```powershell
cd frontend
npm test
npm run build
```

### End-to-end smoke test

Make sure backend and frontend are already running, then execute:

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

## Related docs

- [README-cn.md](README-cn.md)
- [backend/README.md](backend/README.md)
- [frontend/README.md](frontend/README.md)
- [docs/测试手册.md](docs/测试手册.md)
- [docs/技术选型文档.md](docs/技术选型文档.md)
- [docs/架构设计.md](docs/架构设计.md)
- [docs/学习手册.md](docs/学习手册.md)

## License

MIT
