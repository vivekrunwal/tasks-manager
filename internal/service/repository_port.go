package service

import (
    "context"

    "github.com/google/uuid"

    "task-svc/internal/domain"
)

// TaskRepository defines the dependency needed by the service layer.
// It allows the service to depend on an abstraction rather than on the concrete repository implementation.
type TaskRepository interface {
    Create(ctx context.Context, task *domain.Task) error
    List(ctx context.Context, filter domain.TaskFilter) (*domain.TaskList, error)
    Get(ctx context.Context, id uuid.UUID) (*domain.Task, error)
    Update(ctx context.Context, task *domain.Task) error
    Patch(ctx context.Context, id uuid.UUID, patch *domain.PatchTaskRequest) (*domain.Task, error)
    Delete(ctx context.Context, id uuid.UUID) error
}


