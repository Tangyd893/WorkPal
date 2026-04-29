# WorkPal

[中文说明](README-cn.md) | English

WorkPal is a Go + React office collaboration demo that now runs as a real microservice project rather than a hybrid monolith. The frontend talks to a single API gateway, while domain services own their own runtime and storage boundaries.

## What the project includes

- seeded admin and employee acceptance accounts
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

| Service | Storage boundary | Main responsibility |
| --- | --- | --- |
| Gateway | stateless | ingress, route catalog, service catalog, rate limit, retry, circuit breaker, health |
| User Service | `workpal_user` | login, users, departments, employees, dev seed data |
| IM Service | `workpal_im` | direct chat, group chat, messages, announcements, WebSocket |
| File Service | `workpal_file` | file metadata, upload, share, delete |
| Search Service | Bleve + Redis Streams | message indexing and search |
| Workspace Service | `workpal_workspace` | tasks and schedule |

## Gateway learning surface

The gateway is now a real microservice entry layer instead of a plain reverse proxy. It exposes:

- `GET /health/live`
- `GET /health/ready`
- `GET /health`
- `GET /gateway/routes`
- `GET /gateway/services`

It also implements:

- request IDs
- basic rate limiting
- service-level timeouts
- retries for idempotent read requests only
- per-service circuit breakers

If you know Spring Cloud Alibaba, you can read this layer as a lightweight Go mapping of Gateway + Sentinel concepts, with a static service catalog instead of a dynamic registry.

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

On first boot, backend services automatically create the service-owned databases they need:

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

## Quick verification

```powershell
Invoke-RestMethod http://localhost:8080/health/live
Invoke-RestMethod http://localhost:8080/health/ready
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/gateway/routes
Invoke-RestMethod http://localhost:8080/gateway/services
```

You should see a live gateway, a ready gateway with healthy downstreams, a route catalog, and a service catalog with timeout / retry / circuit-breaker metadata.

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
