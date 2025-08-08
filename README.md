# task-svc: Task Management Service

A production-ready microservice for managing tasks, built with Go, Chi router, and PostgreSQL.

## Features

- RESTful API for task management
- PostgreSQL database with migrations
- Robust error handling and validation
- Pagination, filtering, and sorting
- Optimistic concurrency control
- Health checks and metrics
- API documentation with OpenAPI/Swagger
- Containerization with Docker

## Project Structure

The project follows a clean architecture approach with layers:

```
task-svc/
  ├── cmd/task-svc/         # Application entry point
  ├── internal/             # Internal packages
  │   ├── config/           # Configuration management
  │   ├── domain/           # Domain models
  │   ├── http/             # HTTP handlers and middleware
  │   ├── platform/         # Platform-specific code
  │   │   ├── db/           # Database connectivity
  │   │   ├── log/          # Logging utilities
  │   │   └── metrics/      # Metrics collection
  │   ├── repo/             # Repository layer
  │   └── service/          # Service layer (business logic)
  ├── pkg/                  # Reusable packages
  │   └── pagination/       # Pagination utilities
  ├── db/migrations/        # Database migrations
  ├── docs/                 # Documentation
  │   └── openapi.yaml      # OpenAPI specification
  ├── Dockerfile            # Docker container definition
  ├── docker-compose.yaml   # Docker Compose configuration
  ├── Makefile              # Build and run commands
  ├── go.mod                # Go modules
  └── README.md             # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Docker & Docker Compose (optional)

### Running with Docker Compose

The easiest way to start the service is using Docker Compose:

```bash
# Start PostgreSQL and the service
docker-compose up -d

# Run database migrations
make migrate-up

# The service will be available at http://localhost:8080
```

### Running Locally

```bash
# Set up the database connection (adjust as needed)
export DB_DSN=postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable

# Run migrations
make migrate-up

# Start the service
make dev
```

### API Documentation

The API documentation is available at `/docs` when the service is running.

## API Endpoints

| Method | Endpoint           | Description                               |
|--------|-------------------|------------------------------------------|
| GET    | /v1/tasks         | List tasks with pagination and filtering  |
| POST   | /v1/tasks         | Create a new task                         |
| GET    | /v1/tasks/{id}    | Get a specific task                       |
| PUT    | /v1/tasks/{id}    | Update a task (full update)               |
| PATCH  | /v1/tasks/{id}    | Partially update a task                   |
| DELETE | /v1/tasks/{id}    | Delete a task                             |
| GET    | /healthz          | Health check                              |
| GET    | /readyz           | Readiness check                           |
| GET    | /metrics          | Prometheus metrics                        |

## Example API Usage

### Create a task

```bash
curl -s -X POST localhost:8080/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Complete project documentation","status":"InProgress","due_date":"2023-12-31T23:59:59Z"}' | jq
```

Response:
```json
{
  "id": "3c0d938d-1f5d-4c8e-a6f8-b9e139c053d7",
  "title": "Complete project documentation",
  "status": "InProgress",
  "due_date": "2023-12-31T23:59:59Z",
  "created_at": "2023-06-15T10:30:00Z",
  "updated_at": "2023-06-15T10:30:00Z",
  "version": 1
}
```

### List tasks (with filtering and pagination)

```bash
curl -s "localhost:8080/v1/tasks?status=InProgress&page=1&page_size=10&sort=-created_at" | jq
```

Response:
```json
{
  "items": [
    {
      "id": "3c0d938d-1f5d-4c8e-a6f8-b9e139c053d7",
      "title": "Complete project documentation",
      "status": "InProgress",
      "due_date": "2023-12-31T23:59:59Z",
      "created_at": "2023-06-15T10:30:00Z",
      "updated_at": "2023-06-15T10:30:00Z",
      "version": 1
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 10,
    "total_items": 1,
    "total_pages": 1
  }
}
```

### Update a task with optimistic locking

```bash
curl -s -X PUT localhost:8080/v1/tasks/3c0d938d-1f5d-4c8e-a6f8-b9e139c053d7 \
  -H 'Content-Type: application/json' \
  -d '{"title":"Complete project documentation","description":"Include all API examples","status":"Completed","version":1}' | jq
```

### Partially update a task

```bash
curl -s -X PATCH localhost:8080/v1/tasks/3c0d938d-1f5d-4c8e-a6f8-b9e139c053d7 \
  -H 'Content-Type: application/json' \
  -H 'If-Match: 2' \
  -d '{"status":"Completed"}' | jq
```

### Delete a task

```bash
curl -s -X DELETE localhost:8080/v1/tasks/3c0d938d-1f5d-4c8e-a6f8-b9e139c053d7 -w "%{http_code}\n"
```

## Design Decisions

- **Chi Router**: Lightweight, idiomatic HTTP router with built-in middleware support
- **Repository Pattern**: Separates data access logic from business logic
- **Service Layer**: Encapsulates business rules and domain logic
- **PostgreSQL**: Robust relational database with ENUM type support for task status
- **Optimistic Locking**: Prevents concurrent updates with version field
- **Structured Logging**: Using Go's built-in `slog` package for structured logs
- **Graceful Shutdown**: Proper handling of shutdown signals for zero-downtime deployments
- **Prometheus Metrics**: Collect and expose metrics for monitoring

## Future Enhancements

- Add user authentication and authorization
- Implement task assignments to users
- Add task search functionality
- Integrate event publishing for task status changes
- Support file attachments for tasks
- Add webhooks for task events
