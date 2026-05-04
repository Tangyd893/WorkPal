# WorkPal

[中文说明](README-cn.md) | English

WorkPal is a Go + React office collaboration platform running as a real microservice system. The frontend communicates exclusively through the API gateway, while each backend domain service owns its own runtime and storage boundary.

## What is in the project

- seeded acceptance accounts: `admin`, `emma.chen`, `liam.wang`, `sofia.zhao`
- bilingual UI: `English / 简体中文`
- light and dark theme, message sound toggle, density toggle
- overview, chat, tasks, schedule, files, and directory modules
- direct chat, group chat, group announcements, group files
- message editing, message recall, and inline editing in the chat pane
- backend-backed tasks, schedule, files, and directory search
- gateway governance, Redis-backed service registry, Redis-backed IM cluster fanout, outbox-backed Redis Streams search indexing
- versioned database migrations for all four domain services

## Stack

- **Backend:** Go 1.22, Gin, GORM, PostgreSQL 16, Redis 7, Redis Streams, Bleve, golang-migrate
- **Frontend:** React 18, Vite 5, TypeScript 5.4, Zustand 4.5
- **File storage:** MinIO with local-file fallback
- **Realtime:** WebSocket through the IM service, with Redis Pub/Sub fanout for multi-instance delivery

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
| IM Service | `workpal_im` | direct chat, group chat, announcements, messages, message edit/recall, WebSocket, Redis fanout, message outbox |
| File Service | `workpal_file` | file metadata, upload, share, delete |
| Search Service | Bleve + Redis Streams | message indexing and search |
| Workspace Service | `workpal_workspace` | tasks and schedule |

## Gateway learning surface

Gateway exposes:

- `GET /health/live`
- `GET /health/ready`
- `GET /health`
- `GET /gateway/routes`
- `GET /gateway/services`

Gateway implements:

- request ID propagation
- explicit route catalog
- Redis-backed service discovery with static-config fallback
- service-level timeouts
- retries for idempotent read requests only
- per-service circuit breakers
- in-memory rate limiting

## Database migrations

Each service has versioned SQL migrations under `backend/migrations/<service>/`:

| Service | Migration | Tables |
| --- | --- | --- |
| user-service | `001_init` | `users`, `departments`, `employees` |
| im-service | `001_init` | `conversations`, `conversation_members`, `messages`, `message_outbox`, `message_reads` |
| file-service | `001_init` | `files` |
| workspace-service | `001_init` | `tasks`, `schedule_events` |

Run migrations manually:

```powershell
cd backend
make migrate-install
make migrate-up SERVICE=user-service
make migrate-down SERVICE=user-service
```

Or create new migrations:

```powershell
make migrate-create SERVICE=im-service NAME=add_message_attachments
```

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

- tasks, schedule, files, chat, and directory are backend-backed
- message editing, recall, and mark-read are supported via the IM API
- the files module no longer mixes frontend-only seeded documents into the main document list
- seeded accounts remain intentionally exposed on the login screen for acceptance and debugging

## Tests

### Backend

```powershell
cd backend
go vet ./...
go test -race ./...
```

### Frontend

```powershell
cd frontend
npm run lint
npm test
npm run build
```

### Continuous integration

GitHub Actions runs on pushes and pull requests to `main`. The pipeline includes:

- **Backend:** build, `go vet`, `golangci-lint`, race-enabled Go tests
- **Frontend:** TypeScript type check, ESLint, Vitest component tests, production build
- **E2E:** starts Compose services, runs Playwright API smoke tests (health, login, chat)
- **Compose:** Docker Compose configuration validation

### End-to-end smoke test

Make sure backend and frontend are already running, then:

```powershell
cd testing/e2e
npm install
npx playwright install chromium
node playwright.mjs
```
