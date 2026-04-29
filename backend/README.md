# WorkPal Backend

This is the Go backend for WorkPal.

If your goal is to boot the whole project locally, start with the repo root [README.md](../README.md). This document focuses on backend-specific structure and debugging notes.

## What the backend provides

- authentication and JWT issuance
- user, employee, and department directory data
- direct chat and group chat
- message search and WebSocket realtime delivery
- personal files and group files
- workspace task and schedule APIs
- health checks and Prometheus metrics

## Main startup path

For normal local development, use the integrated server:

```powershell
cd backend
go run ./cmd/server
```

That entrypoint wires together:

- user module
- IM module
- search module
- file module
- workspace module

## Config resolution

The backend reads config in this order:

1. `CONFIG_PATH`
2. `backend/configs/config.yaml`
3. `backend/configs/config.example.yaml`

So the sample config is enough to start locally unless you need custom overrides.

## Local dependencies

The backend expects:

- PostgreSQL
- Redis
- MinIO or local file storage fallback

Recommended startup from the repo root:

```powershell
docker compose -f docker/docker-compose.yaml up -d
```

## Seed data in development mode

When `server.mode != release`, the backend automatically ensures:

- departments
- employees
- acceptance accounts

Seeded accounts:

| Username | Password | Role |
|---|---|---|
| `admin` | `admin123` | admin |
| `emma.chen` | `workpal123` | operations |
| `liam.wang` | `workpal123` | engineering |
| `sofia.zhao` | `workpal123` | design |

This seed data is what powers the frontend directory filters and multi-user chat testing.

## Package layout

Core backend code lives under `backend/internal`:

- `user`: auth, user data, employee and department directory
- `im`: conversations, messages, presence, WebSocket
- `file`: upload, download, share, delete
- `search`: Bleve-backed message search
- `workspace`: tasks and schedule APIs
- `common`: middleware, response helpers, pagination, errors
- `platform`: runtime bootstrap and dev seed helpers

Supporting packages:

- `pkg/auth`: JWT helpers
- `pkg/cache`: Redis cache bootstrap
- `pkg/msgqueue`: Redis Streams abstraction

## API highlights

Key routes include:

- `POST /api/v1/auth/login`
- `GET /api/v1/users/me`
- `GET /api/v1/users`
- `GET /api/v1/departments`
- `POST /api/v1/conversations`
- `POST /api/v1/conversations/:id/messages`
- `PUT /api/v1/conversations/:id/announcement`
- `GET /api/v1/conversations/:id/files`
- `POST /api/v1/files/upload`
- `GET /api/v1/tasks`
- `POST /api/v1/tasks`
- `GET /api/v1/schedule`
- `POST /api/v1/schedule`
- `GET /health`
- `GET /metrics`

## Realtime behavior

WebSocket endpoint:

```text
/ws?token=<jwt>
```

The server loads all conversations for the authenticated user and joins those rooms on connect.

## Tests

Run the backend test suite:

```powershell
cd backend
go test ./...
```

Optional race test:

```powershell
go test -race ./...
```

## About the other `cmd/*` services

This repo also contains entrypoints such as:

- `cmd/gateway`
- `cmd/user-service`
- `cmd/im-service`
- `cmd/file-service`
- `cmd/search-service`
- `cmd/workspace-service`

They reflect a microservice-oriented evolution path. For everyday local debugging, `cmd/server` is still the simplest and most complete entrypoint.
