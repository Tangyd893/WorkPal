# WorkPal

[中文说明](README-cn.md) | English

WorkPal is a Go microservices + React office collaboration demo. The current version is no longer just a chat shell. It now includes:

- seeded admin and employee accounts for acceptance
- bilingual UI: `English / 简体中文`
- light and dark theme, message sound toggle, compact density toggle
- overview, chat, tasks, schedule, files, and directory modules
- department and employee seed data in the backend database
- direct chat, group chat, group announcement, and group files

This README is written for the current code in this repository and is intended to be followed step by step for local startup and debugging.

## Stack

- Backend: Go microservices, Gin, GORM, PostgreSQL, Redis Streams, Bleve
- Frontend: React, Vite, Zustand
- File storage: local storage by default, MinIO supported
- Realtime: WebSocket through the IM service and API gateway

## Prerequisites

- Go `1.22+`
- Node.js `18+`
- npm
- Docker Desktop or Docker Engine

Before doing anything else, confirm Docker is actually running:

```powershell
docker version
```

Only continue if the output contains both `Client` and `Server`.

## Default Ports

| Service | URL | Notes |
|---|---|---|
| Frontend | `http://localhost:3000` | Vite dev server |
| API Gateway | `http://localhost:8080` | single frontend entry for HTTP and WebSocket |
| User Service | `http://localhost:8081` | auth, users, departments |
| IM Service | `http://localhost:8082` | conversations, messages, WebSocket |
| File Service | `http://localhost:8083` | personal files and group files |
| Search Service | `http://localhost:8084` | Bleve search API and message-index consumer |
| Health check | `http://localhost:8080/health` | gateway health endpoint |
| PostgreSQL | `localhost:5432` | `workpal / workpal123` |
| Redis | `localhost:6379` | no password by default |
| MinIO API | `http://localhost:9000` | object storage |
| MinIO Console | `http://localhost:9001` | `workpal / workpal123456` |

## Quick Start

### 1. Start infrastructure dependencies

From the repo root:

```powershell
docker compose -f docker/docker-compose.yaml up -d
docker compose -f docker/docker-compose.yaml ps
```

Expected result:

- `postgres` is `Up` or `healthy`
- `redis` is `Up`
- `minio` is `Up`

### 2. Start the backend microservices

Open separate terminals for the services below:

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
go run ./cmd/gateway
```

The gateway remains the only backend URL that the frontend needs to know. It routes requests to the domain services:

| Gateway path | Upstream service |
|---|---|
| `/api/v1/auth/*`, `/api/v1/users*`, `/api/v1/departments*` | User Service |
| `/api/v1/conversations*`, `/api/v1/messages*`, `/ws` | IM Service |
| `/api/v1/files*`, `/api/v1/conversations/:id/files` | File Service |
| `/api/v1/search*` | Search Service |

For quick local compatibility, the legacy all-in-one server is still available:

```powershell
cd backend
go run ./cmd/server
```

Important startup behavior in microservice mode:

1. Each service migrates the tables it owns on startup.
2. The User Service ensures the seeded departments, employees, and acceptance accounts exist in non-`release` mode.
3. The IM Service writes messages to PostgreSQL and publishes message index events to Redis Streams.
4. The Search Service consumes those Redis Streams events and updates the Bleve index.
5. You do **not** need to create `backend/configs/config.yaml` unless you want to override the sample config.

Backend config lookup order:

1. `CONFIG_PATH`
2. `backend/configs/config.yaml`
3. `backend/configs/config.example.yaml`

Quick verification:

```powershell
Invoke-WebRequest http://localhost:8080/health -UseBasicParsing
Invoke-WebRequest http://localhost:8080/ -UseBasicParsing
```

Expected result:

- `/health` returns HTTP `200`
- `/` returns JSON similar to:

```json
{"name":"WorkPal","status":"running","version":"0.2.0"}
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

Frontend proxy rules:

- `/api/*` -> `http://localhost:8080` -> API Gateway -> target service
- `/ws` -> `ws://localhost:8080` -> API Gateway -> IM Service

## Microservice Message Flow

Chat messages use HTTP for persistence and Redis Streams for cross-service indexing:

```text
Frontend
  -> API Gateway
  -> IM Service
  -> PostgreSQL
  -> Redis Streams: message.upserted / message.deleted
  -> Search Service
  -> Bleve index
```

This keeps message sending reliable even if search indexing is temporarily unavailable. The source of truth remains PostgreSQL, while Redis Streams decouples IM writes from search indexing.

## Acceptance Accounts

When the backend is started in the default development mode, it automatically ensures these accounts exist:

| Role | Username | Password | Suggested use |
|---|---|---|---|
| Admin | `admin` | `admin123` | full acceptance and settings checks |
| Employee | `emma.chen` | `workpal123` | operations and collaboration flows |
| Employee | `liam.wang` | `workpal123` | engineering and group tests |
| Employee | `sofia.zhao` | `workpal123` | design and release-readiness tests |

Seeded org data also includes departments and employee profiles, so directory search and department filters work out of the box.

If you want to verify login through the API first:

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
- `data.token` is present

## Recommended Acceptance Path

After logging in on the frontend, use this order:

1. `Overview / 总览`
   - confirm the overview loads
   - click the metric cards or module buttons and verify they jump to the matching module

2. `Preferences / 偏好设置`
   - switch `English / 简体中文`
   - switch light and dark theme
   - toggle message sound
   - toggle comfortable and compact density

3. `Directory / 通讯录`
   - verify seeded users are visible
   - use the department filter
   - search by title, phone, or department, not just by username
   - example: filter `Engineering` and search `Platform Engineer`

4. `Chat / 沟通`
   - create a direct chat with `emma.chen`
   - send a message
   - create a group with `emma.chen` and `liam.wang`
   - send a group message
   - update the group announcement
   - upload a group file

5. `Tasks / 任务`
   - create a task
   - move it across columns
   - share it
   - delete it

6. `Schedule / 日程`
   - create an event
   - share it
   - delete it

7. `Files / 文件`
   - upload a personal file
   - open it
   - share it
   - delete it

## What Is Backend-Backed vs Local Demo State

This is important for debugging expectations.

### Backend-backed right now

- login
- current user
- user list and department list
- directory search and department filtering
- direct chat and group chat
- message send and message search
- WebSocket connection
- group announcement
- group files
- personal file upload, list, share, delete

### Frontend-local demo state right now

- overview summary composition
- task board items
- schedule items
- seeded knowledge cards in the files module

That means:

- tasks and schedule are functional in the UI, but are not persisted to the backend yet
- files uploaded through the file service are real backend data
- the overview module reflects current frontend state and backend-loaded people data

## Validation Commands

### Backend and frontend tests

```powershell
cd backend
go test ./...
make build-services

cd ..\frontend
npm test
npm run build
```

### End-to-end smoke test

Make sure backend and frontend are already running, then:

```powershell
cd frontend
npx playwright install chromium
node ..\testing\e2e\playwright.mjs
```

The Playwright smoke test covers:

- `/health`
- `/metrics`
- seeded login API
- private and group chat API flows
- group announcement and group file API flows
- frontend login
- overview jump actions
- directory filter and search
- task creation
- schedule creation
- file upload
- direct chat creation and send

## Stop Services

Stop frontend and backend with `Ctrl + C` in their terminals.

Stop Docker dependencies from the repo root:

```powershell
docker compose -f docker/docker-compose.yaml down
```

## Related Docs

- [README-cn.md](README-cn.md)
- [frontend/README.md](frontend/README.md)
- [docs/acceptance-testing.md](docs/acceptance-testing.md)
- [docs/项目技术特点学习笔记.md](docs/项目技术特点学习笔记.md)

## License

MIT
