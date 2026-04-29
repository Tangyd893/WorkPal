# WorkPal Backend

This directory contains the Go backend for WorkPal. The backend now runs only in microservice form; the legacy all-in-one server entrypoint is gone.

## Service map

| Service | Entry | Storage boundary | Responsibility |
| --- | --- | --- | --- |
| Gateway | `cmd/gateway` | stateless | ingress, reverse proxy, rate limit, aggregated health |
| User Service | `cmd/user-service` | `workpal_user` | auth, users, departments, employees, dev seed data |
| IM Service | `cmd/im-service` | `workpal_im` | conversations, messages, announcements, WebSocket |
| File Service | `cmd/file-service` | `workpal_file` | file metadata and object storage access |
| Search Service | `cmd/search-service` | Bleve + Redis Streams | message indexing and search |
| Workspace Service | `cmd/workspace-service` | `workpal_workspace` | tasks and schedule |

## Startup

If you want the simplest full-project path, use the repo root [README.md](../README.md).

If you want to run backend services manually, start infrastructure first:

```powershell
docker compose -f docker/docker-compose.yaml up -d postgres redis minio
```

Then, from `backend`, run each service in its own terminal:

```powershell
go run ./cmd/user-service
```

```powershell
go run ./cmd/im-service
```

```powershell
go run ./cmd/file-service
```

```powershell
go run ./cmd/search-service
```

```powershell
go run ./cmd/workspace-service
```

```powershell
go run ./cmd/gateway
```

## Config lookup

Each service resolves config in this order:

1. `CONFIG_PATH`
2. `backend/configs/config.yaml`
3. `backend/configs/config.example.yaml`

The sample config is enough for local development unless you need overrides.

## Database ownership

The services no longer share one application database. On first startup they ensure the databases they own exist:

- User Service -> `workpal_user`
- IM Service -> `workpal_im`
- File Service -> `workpal_file`
- Workspace Service -> `workpal_workspace`

Search Service does not own a PostgreSQL database in the current design. It keeps search state in a Bleve index and consumes message events from Redis Streams.

## Seeded development data

When `server.mode != release`, User Service automatically ensures:

- departments
- employee profiles
- acceptance accounts

| Username | Password | Suggested role |
| --- | --- | --- |
| `admin` | `admin123` | admin |
| `emma.chen` | `workpal123` | operations |
| `liam.wang` | `workpal123` | engineering |
| `sofia.zhao` | `workpal123` | design |

## Health endpoints

Every service exposes:

- `GET /`
- `GET /health`
- `GET /metrics`

Gateway health is aggregated. `GET http://localhost:8080/health` actively checks all downstream services.

## Package layout

Key folders:

- `configs`: config model and sample config
- `cmd`: service entrypoints
- `internal/platform`: shared runtime bootstrap and development seed helpers
- `internal/user`: auth and directory domain
- `internal/im`: conversations, messages, WebSocket
- `internal/file`: uploads, file metadata, share and delete flows
- `internal/search`: Bleve search logic
- `internal/workspace`: tasks and schedule
- `pkg/msgqueue`: Redis Streams abstraction
- `pkg/auth`: JWT helpers
- `pkg/cache`: Redis-backed cache bootstrap

## Tests

Run the backend test suite:

```powershell
cd backend
go test ./...
```

Optional build check:

```powershell
make build-services
```
