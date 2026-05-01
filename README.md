# WorkPal

[中文说明](README-cn.md) | English

WorkPal is a Go + React office collaboration project that now runs as a real microservice system. The frontend talks only to the API gateway, while backend domain services own their own runtime and storage boundaries.

## What is in the project

- seeded acceptance accounts: `admin`, `emma.chen`, `liam.wang`, `sofia.zhao`
- bilingual UI: `English / 简体中文`
- light and dark theme, message sound toggle, density toggle
- overview, chat, tasks, schedule, files, and directory modules
- direct chat, group chat, group announcement, and group files
- backend-backed tasks, schedule, files, and directory search
- gateway governance, Redis-backed service registry, Redis-backed IM cluster fanout, and outbox-backed Redis Streams search indexing

## Stack

- Backend: Go, Gin, GORM, PostgreSQL, Redis, Redis Streams, Bleve
- Frontend: React 18, Vite, TypeScript, Zustand
- File storage: MinIO with local-file fallback
- Realtime: WebSocket through the IM service, with Redis Pub/Sub fanout for multi-instance delivery

## Ports

| Component | URL | Notes |
| --- | --- | --- |
| Frontend | `http://localhost:3000` | Vite dev server |
| API Gateway | `http://localhost:8080` | the only backend URL the frontend uses |
| User Service | `http://localhost:8081` | auth, users, departments, employees |
| IM Service | `http://localhost:8082` | conversations, messages, WebSocket |
| File Service | `http://localhost:8083` | personal files and group files |
| Search Service | `http://localhost:8084` | message indexing and search |
| Workspace Service | `http://localhost:8085` | tasks and schedule |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | default no password |
| MinIO API | `http://localhost:9000` | object storage |
| MinIO Console | `http://localhost:9001` | `workpal / workpal123456` |

## Microservice topology

| Service | Storage boundary | Main responsibility |
| --- | --- | --- |
| Gateway | stateless | ingress, route catalog, service catalog, service discovery fallback, rate limit, retry, circuit breaker, health |
| User Service | `workpal_user` | login, users, departments, employees, development seed data |
| IM Service | `workpal_im` | direct chat, group chat, announcements, messages, WebSocket, Redis fanout, message outbox |
| File Service | `workpal_file` | file metadata, upload, share, delete |
| Search Service | Bleve + Redis Streams | message indexing and search |
| Workspace Service | `workpal_workspace` | tasks and schedule |

## Gateway learning surface

Gateway now exposes:

- `GET /health/live`
- `GET /health/ready`
- `GET /health`
- `GET /gateway/routes`
- `GET /gateway/services`

Gateway now implements:

- request ID propagation
- explicit route catalog
- Redis-backed service discovery with static-config fallback
- service-level timeouts
- retries for idempotent read requests only
- per-service circuit breakers
- in-memory rate limiting

`/gateway/services` now shows discovered instances when Redis registry data is available, and falls back to the configured static upstream when registry data is missing.

## Microservice learning mapping

If you are learning from a Spring Cloud Alibaba perspective, the current Go project maps roughly like this:

- Spring Cloud Gateway -> `backend/cmd/gateway`
- Nacos-like service registry -> Redis-backed registrations in `backend/internal/platform/registry.go`
- Sentinel-like ingress governance -> gateway rate limit, retry, breaker, readiness checks
- Feign-like service calls -> `backend/internal/clients/*`
- RocketMQ-like async search update path -> IM outbox plus Redis Streams into Search

## Quick start

### 1. Make sure Docker is running

```powershell
docker version
```

Continue only when the output contains both `Client` and `Server`.

### 2. Start the full stack with Docker Compose

From the repo root:

```powershell
docker compose -f docker/docker-compose.yaml build
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

Expected result:

- `postgres`, `redis`, and `minio` are `Up` or `healthy`
- `gateway`, `user-service`, `im-service`, `file-service`, `search-service`, and `workspace-service` are `Up`

Compose waits for Redis before starting registry-enabled backend services, so `/gateway/services` can show discovered instances instead of immediately falling back to static URLs.

Each backend service automatically ensures the databases it owns exist:

- `workpal_user`
- `workpal_im`
- `workpal_file`
- `workpal_workspace`

### 3. Start services from source for debugging

Start infrastructure only:

```powershell
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

Then run these in separate terminals:

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

### 4. Start the frontend

```powershell
cd frontend
npm ci
npm run dev -- --host 127.0.0.1
```

Then open `http://localhost:3000`.

## Acceptance accounts

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

You should see:

- gateway liveness
- gateway readiness across downstream services
- explicit route catalog
- service catalog with discovery mode, discovered instances, timeout, retry, and breaker metadata

## Notes about current frontend data

- tasks, schedule, files, and directory are backend-backed
- the files module no longer mixes frontend-only seeded documents into the main document list
- seeded accounts remain intentionally exposed on the login screen for acceptance and debugging

## Tests

### Backend

```powershell
cd backend
go test ./...
```

### Frontend

```powershell
cd frontend
npm run lint
npm test
npm run build
```

### Continuous integration

GitHub Actions runs on pushes and pull requests to `main`. The current workflow validates backend build, `go vet`, golangci-lint, race-enabled Go tests, frontend lint, frontend unit and component tests, frontend production build, and Docker Compose configuration.

### End-to-end smoke test

Make sure backend and frontend are already running, then execute:

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```
