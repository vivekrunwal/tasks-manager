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

## License

MIT
