<div align="center">

# task-svc • Task Management Service

Manage tasks with a clean, production-ready Go service powered by Chi and PostgreSQL.

<br/>

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white) 
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-4169E1?logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)

</div>

---

## Table of Contents

- Quick Start
- Features
- Project Structure
- Run Options
  - Docker (recommended)
  - Native (Go installed)
- API Docs (Swagger UI)
- Endpoints
- Examples
- Configuration
- Makefile targets
- Troubleshooting

---

## Quick Start

```bash
# Start everything with Docker
docker compose up -d

# Apply DB migrations
make migrate-up

# Open docs
open http://localhost:8080/docs
```

Create a task:
```bash
curl -s -X POST localhost:8080/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"My first task"}' | jq
```

---

## Features

- RESTful API for task management
- PostgreSQL migrations (SQL-first)
- Validation, pagination, sorting, filtering
- Optimistic concurrency (versioning)
- Health checks and Prometheus metrics
- OpenAPI/Swagger docs
- Dockerized for easy local dev

---

## Project Structure

```
task-svc/
  ├── cmd/task-svc/         # Application entry point
  ├── internal/             # Internal packages
  │   ├── config/           # Configuration management
  │   ├── domain/           # Domain models
  │   ├── http/             # Handlers & middleware (incl. Swagger UI)
  │   ├── platform/         # DB, logging, metrics
  │   ├── repo/             # Data access layer
  │   └── service/          # Business logic
  ├── pkg/                  # Reusable packages
  ├── db/migrations/        # Database migrations
  ├── docs/openapi.yaml     # OpenAPI specification
  ├── Dockerfile            # Container image
  ├── docker-compose.yaml   # Local stack
  ├── Makefile              # Dev tooling
  └── README.md             # You are here
```

---

## Run Options

### Docker (recommended)

```bash
docker compose up -d
make migrate-up
```

Open:
- Swagger UI: http://localhost:8080/docs
- Health: http://localhost:8080/healthz

Stop:
```bash
docker compose down
```

Tip: `bash ./run_dev.sh` will bring up Postgres, apply migrations, and run the service.

### Native (Go installed)

```bash
# Start only Postgres (Docker)
docker compose up -d postgres

# Configure app connection
export DB_DSN=postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable

# Apply migrations
go run ./cmd/migrate

# Run the service (with hot reload if you have 'air')
make dev
# or
go run ./cmd/task-svc
```

---

## API Docs (Swagger UI)

- Open: http://localhost:8080/docs
- Raw spec: http://localhost:8080/docs/openapi.yaml

---

## Endpoints

| Method | Endpoint        | Description                              |
|--------|------------------|------------------------------------------|
| GET    | /v1/tasks        | List tasks (filters, pagination, sorting)|
| POST   | /v1/tasks        | Create a new task                        |
| GET    | /v1/tasks/{id}   | Get a task                               |
| PUT    | /v1/tasks/{id}   | Update a task (full update)              |
| PATCH  | /v1/tasks/{id}   | Partially update a task                  |
| DELETE | /v1/tasks/{id}   | Delete a task                            |
| GET    | /healthz         | Health check                             |
| GET    | /metrics         | Prometheus metrics                       |

---

## Examples

Create:
```bash
curl -s -X POST localhost:8080/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Complete the project","status":"InProgress","due_date":"2025-12-31T23:59:59Z"}' | jq
```

List (filters + pagination):
```bash
curl -s "localhost:8080/v1/tasks?status=InProgress&page=1&page_size=10&sort=-created_at" | jq
```

Update (optimistic locking):
```bash
curl -s -X PUT localhost:8080/v1/tasks/{id} \
  -H 'Content-Type: application/json' \
  -d '{"title":"Updated","status":"Completed","version":1}' | jq
```

Patch with If-Match header:
```bash
curl -s -X PATCH localhost:8080/v1/tasks/{id} \
  -H 'Content-Type: application/json' \
  -H 'If-Match: 2' \
  -d '{"status":"Completed"}' | jq
```

Delete:
```bash
curl -s -X DELETE localhost:8080/v1/tasks/{id} -w "%{http_code}\n"
```

---

## Configuration

Environment variables (with defaults):

| Variable                      | Default                                                         |
|------------------------------|-----------------------------------------------------------------|
| APP_ENV                       | dev                                                             |
| HTTP_ADDR                     | :8080                                                           |
| HTTP_READ_TIMEOUT             | 10s                                                             |
| HTTP_WRITE_TIMEOUT            | 10s                                                             |
| HTTP_IDLE_TIMEOUT             | 60s                                                             |
| DB_DSN                        | postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable |
| DB_MAX_CONNS                  | 10                                                              |
| DB_MAX_IDLE_CONNS             | 5                                                               |
| DB_MAX_LIFETIME               | 5m                                                              |
| PAGINATION_DEFAULT_SIZE       | 20                                                              |
| PAGINATION_MAX_SIZE           | 100                                                             |

---

## Makefile targets

```bash
make build         # Build binary into ./bin
make dev           # Run app (uses air if available, else go run)
make test          # Run tests
make lint          # Run golangci-lint
make migrate-up    # Apply migrations to running Postgres container
make migrate-down  # Rollback migrations
make compose-up    # docker compose up -d
make compose-down  # docker compose down
```

---

## Troubleshooting

- 404 on /v1/tasks
  - Ensure you’re on the latest container: `docker compose down && docker compose build --no-cache app && docker compose up -d`
- Migration errors like “already exists”
  - Safe to ignore when re-applying the same SQL (schema is already present)
- Cannot connect to DB
  - Ensure Postgres is up: `docker compose ps`; connection string uses `localhost:5432`
- Health endpoint
  - `curl http://localhost:8080/healthz` should return `{ "status": "ok" }`

---

### Architecture

![Architecture Diagram (SVG)](https://www.plantuml.com/plantuml/svg/JP5FZzem4CNl_XIZd4ArY6sFgwhQ2gsW4YieeUgXxM5m1cFXsEdO2OkgVlTw55Bb46I_tvi_VZp7X9owngbuiRv5nWPqQDhWoQSgtHm1aqxeZQE99Pwn3bVh3PpGnIcvChRVRPtEyE7nKIMR7C6s9qRdO3asXi-ippRbnXsJKgXJO2YMpcPM04OOZagrg2ze28g4yJAFlOPo5MO5_540FzQ7mmxMw6j77A7C7MU_f7YTKYlyrTOPNP4f_lJbShkl7c0JZCbhFy0h9ROw3XNPAGKLSMyrmGPln0AWwhM0FRKXsXJuDc4Y2bm6Vx86IlBzq6rvCg9NLpI66BvzWN9HXt7MNHR7VOUCc-asTraSiYunbD45Cy2igBzArPZmY96Ws8MO35FrTLC8twtKxuww3BPTEeOtA4-Tf3mJ2eEcVAYbaaEUUwOkBbQC9_G89RqbOezki8j3gU7FWFdy4qYqy0Jxb-e0Y_xzufqI1GH9Xla3EKji-PFp49Fyo4tM6qtZCyUxl6moQepl2LOlUPW4g6GV-3YVZxKPHitOqZZj_i2ijztHsNY6EZHeFL2V-yR_OQXdydNNk7y0)

![Sequence Diagram(SVG)](https://www.plantuml.com/plantuml/svg/LP51Jy8m5CVl_HGld32681uz619Kq1YOR4SlZ_M93LqxsyS8tzw-bJWvhCds_xt-V-tcaJ7miR4g7enhCM03yHkrFnctXo-qaHGRjWIBGCW45SO3l5X_KWfhzW4Mrf1ZrM9WKviM7SDeLLK5hD1fIs446675t5uZ9ONErDIdPLTVXrjgVJJimxdFvLLfPDnX91WLFl8-KFdntgV5Kgai0PF7lWaUeDYK5KoxsPIJR_nqR-Lc3Jklps9jEcJAJhB8M598K_cCuq0_HufyCx1Yc1uXHUOFbTiODrwJ7U2iAulWnJB1h-loULk1kAZqOJ9i4_m9Re5Da_gmQUEqen2DTqhJrNwvVJYV5CDJjHcuB2cnS9VVaiZxkzb5LgSckOV_CyAhEJbzREml)
## Architecture Notes (Microservices Readiness)


### Single Responsibility & Layers
- Handlers focus on HTTP concerns; services encapsulate business logic; repos handle data access; platform packages isolate infra (DB, logs, metrics).

### Scalability
- **Stateless processes**: The app is 12‑factor friendly; all state is in PostgreSQL. This enables running many replicas behind a load balancer (Docker Swarm/Kubernetes/NGINX/Traefik).
- **Horizontal scaling**:
  - Run N replicas; use readiness (`/healthz`) and liveness checks for safe rolling updates.
  - Prefer surge/rolling or blue‑green/canary deployments for zero downtime.
- **Connection pooling**:
  - Tune `DB_MAX_CONNS` and `DB_MAX_IDLE_CONNS` per pod to avoid exhausting DB connections.
  - Optionally place PgBouncer in front of PostgreSQL for transaction pooling.
- **Database scaling path**:
  - Start single instance → add read replicas for GET traffic → partition/ shard if write throughput demands it.
  - Add targeted indexes and avoid N+1 queries. Keep queries predictable; paginate large lists.
- **Caching**:
  - Layer a cache (e.g., Redis) for hot reads like list endpoints; use cache keys derived from query params.
  - Use entity‑level cache with version (optimistic locking) to invalidate precisely.
- **Resilience patterns**:
  - Timeouts everywhere (HTTP client and DB), retries with exponential backoff for transient errors.
  - Circuit breakers/bulkheads (e.g., `sony/gobreaker`) to protect downstreams under failure.
  - Idempotency keys for POST/PUT to tolerate retries without duplications.
- **Observability**:
  - Prometheus metrics at `/metrics`, structured logs (slog), and recommended tracing via OpenTelemetry (Jaeger/Tempo).
  - Propagate correlation IDs (`X-Request-ID`) across boundaries; include them in logs and spans.
- **Migrations without downtime**:
  - Use expand/contract strategy: add new columns non‑null with defaults, backfill, deploy, then remove old paths.
  - Ensure new app versions are backward compatible with old schema during rollout.
- **Security & config**:
  - Strictly via env vars; manage secrets with Vault/Secrets Manager. Prefer mTLS at the mesh/ingress for east‑west traffic.

### Inter-Service Communication
- **Synchronous (request/response)**:
  - **REST/JSON** (current): simple, debuggable, widely compatible. Version APIs (`/v1`, `/v2`), maintain backward compatibility, and use standard HTTP semantics.
  - **gRPC/Protobuf**: strongly typed, fast, supports streaming and deadlines; ideal for high‑throughput internal calls. Consider a thin REST->gRPC gateway for external clients.
  - Set client‑side timeouts (e.g., 200–500ms per hop), use retries only for safe, idempotent methods.
- **Asynchronous (event‑driven)**:
  - Publish domain events to a broker (Kafka/RabbitMQ/NATS) for decoupled reactions (e.g., notifications, analytics).
  - Use the **Outbox pattern** to avoid dual‑write anomalies: write domain change and an outbox row in the same DB transaction; a background relay publishes from the outbox.
  - Ensure **idempotency** and **ordering** per aggregate (e.g., partition by `task_id`) to process events exactly‑once effectively (at‑least‑once delivery + dedupe key).
  - Version event schemas (Protobuf/Avro/JSON‑Schema) and keep them backward compatible; use a schema registry where available.
- **API Gateway / Service Mesh**:
  - Gateway centralizes auth (JWT/OIDC), rate limiting, request shaping, and routing.
  - A mesh (Envoy/Istio/Linkerd) provides mTLS, traffic policies, and uniform telemetry without app changes.
- **Security**:
  - Propagate identity via JWTs or mTLS SPIFFE IDs; enforce RBAC at gateway and service.
  - Validate inputs at the edge; sanitize logs; avoid leaking PII.
- **Consistency patterns**:
  - For cross‑service transactions, prefer **Sagas** (choreography or orchestration). Define compensations for failed steps.
  - Use **CQRS/read models** for aggregated views spanning multiple services.

Example event (JSON) for status changes (publish on updates):

```json
{
  "event_name": "task.status.changed",
  "event_id": "a31d9d9b-7c2c-4b32-9b2e-2a0c1f7b5a10",
  "occurred_at": "2025-08-08T12:34:56Z",
  "task": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "old_status": "Pending",
    "new_status": "InProgress",
    "version": 2
  }
}
```

### Data Ownership
- Each service owns its schema (`tasks` here). Cross-service queries avoided; use APIs/events to share state.

### Extending with a User Service
- Task service stores `user_id` foreign key; reads user details via gRPC/REST or caches from user events.
- Commands mutate only owning service; queries aggregate via API composition or separate read models.

---
