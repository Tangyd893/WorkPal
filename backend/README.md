# WorkPal Backend

This directory contains the Go backend for WorkPal. The backend now runs only in microservice form.

## Service map

| Service | Entry | Storage boundary | Responsibility |
| --- | --- | --- | --- |
| Gateway | `cmd/gateway` | stateless | ingress, route catalog, service catalog, service discovery fallback, rate limit, retry, circuit breaker, health |
| User Service | `cmd/user-service` | `workpal_user` | auth, users, departments, employees, seeded dev data |
| IM Service | `cmd/im-service` | `workpal_im` | conversations, messages, announcements, WebSocket, Redis fanout, outbox publishing |
| File Service | `cmd/file-service` | `workpal_file` | file metadata and object storage access |
| Search Service | `cmd/search-service` | Bleve + Redis Streams | message indexing and search |
| Workspace Service | `cmd/workspace-service` | `workpal_workspace` | tasks and schedule |

## Startup

For the simplest full-project path, use the repo root [README.md](../README.md).

If you want to run services manually, start infrastructure first:

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

IM Service now writes message events into a `message_outbox` table in the same transaction as message changes, and a background worker publishes those outbox records into Redis Streams for Search Service to consume.

## Redis-backed infrastructure

Redis now carries three different roles in the backend:

1. service registry data for discovery
2. IM cluster fanout events for multi-instance WebSocket delivery
3. Redis Streams for asynchronous message indexing into Search Service
4. durable message outbox publishing inside IM Service

This makes Redis a useful learning pivot in the project: not just a cache, but part of the microservice runtime fabric.

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

## Gateway management surface

Gateway exposes a small control plane for learning and debugging:

| Endpoint | Purpose |
| --- | --- |
| `GET /health/live` | liveness |
| `GET /health/ready` | readiness across downstream services |
| `GET /health` | aggregated health |
| `GET /gateway/routes` | route catalog |
| `GET /gateway/services` | downstream service catalog, discovery mode, and breaker state |

Gateway now applies:

- request IDs
- basic rate limiting
- service-level timeouts
- retries for idempotent read requests only
- per-service circuit breakers
- Redis-backed service discovery with static fallback

## Gateway code reading order

If gateway behavior is what you want to study first, read these files in order:

1. `cmd/gateway/main.go`
2. `cmd/gateway/app.go`
3. `cmd/gateway/rate_limit.go`
4. `cmd/gateway/transport.go`
5. `cmd/gateway/breaker.go`
6. `cmd/gateway/gateway_test.go`

## Discovery and cluster-realtime reading order

If your focus is microservice infrastructure instead of CRUD logic, read these next:

1. `internal/platform/registry.go`
2. `cmd/user-service/main.go`
3. `cmd/im-service/main.go`
4. `internal/im/ws/cluster.go`
5. `internal/im/ws/hub.go`
6. `internal/im/service/outbox_publisher.go`
7. `internal/im/handler/message_handler.go`

That path shows how service registration, gateway discovery, multi-instance IM fanout, and the outbox-to-Streams indexing path connect together.

## Package layout

Key folders:

- `configs`: config model and sample config
- `cmd`: service entrypoints
- `internal/platform`: shared runtime bootstrap, registry, and seed helpers
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
