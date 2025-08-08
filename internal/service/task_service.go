package service

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "task-svc/internal/domain"
    "task-svc/internal/repo"
)

// TaskService defines the interface for task operations
type TaskService interface {
	Create(ctx context.Context, req domain.CreateTaskRequest) (*domain.Task, error)
	List(ctx context.Context, filter domain.TaskFilter) (*domain.TaskList, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateTaskRequest) (*domain.Task, error)
	Patch(ctx context.Context, id uuid.UUID, req domain.PatchTaskRequest) (*domain.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// taskService implements the TaskService interface
type taskService struct {
    repository TaskRepository
}

// NewTaskService creates a new task service
func NewTaskService(repo *repo.TaskRepo) TaskService {
    // Accept the concrete repo but depend on the TaskRepository interface internally.
    return &taskService{repository: repo}
}

// Create creates a new task
func (s *taskService) Create(ctx context.Context, req domain.CreateTaskRequest) (*domain.Task, error) {
	now := time.Now().UTC()
	
	// Set default status if not provided
	status := domain.StatusPending
	if req.Status != nil {
		status = *req.Status
	}
	
	task := &domain.Task{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
	
    if err := s.repository.Create(ctx, task); err != nil {
		return nil, err
	}
	
	return task, nil
}

// List retrieves tasks with pagination and filtering
func (s *taskService) List(ctx context.Context, filter domain.TaskFilter) (*domain.TaskList, error) {
    return s.repository.List(ctx, filter)
}

// Get retrieves a task by ID
func (s *taskService) Get(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
    t, err := s.repository.Get(ctx, id)
    if err != nil {
        if errors.Is(err, repo.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return t, nil
}

// Update fully updates a task
func (s *taskService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateTaskRequest) (*domain.Task, error) {
	// First get the existing task
    task, err := s.repository.Get(ctx, id)
	if err != nil {
        if errors.Is(err, repo.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
	}
	
	// Check version for optimistic locking
    if task.Version != req.Version {
        return nil, ErrVersionConflict
	}
	
	// Update the task with new values
	task.Title = req.Title
	task.Description = req.Description
	task.Status = req.Status
	task.DueDate = req.DueDate
	
	// Save the updated task
    if err := s.repository.Update(ctx, task); err != nil {
        if errors.Is(err, repo.ErrVersionConflict) {
            return nil, ErrVersionConflict
        }
        if errors.Is(err, repo.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
	}
	
	// Refresh the task to get the updated version and timestamps
    return s.repository.Get(ctx, id)
}

// Patch partially updates a task
func (s *taskService) Patch(ctx context.Context, id uuid.UUID, req domain.PatchTaskRequest) (*domain.Task, error) {
    t, err := s.repository.Patch(ctx, id, &req)
    if err != nil {
        if errors.Is(err, repo.ErrVersionConflict) {
            return nil, ErrVersionConflict
        }
        if errors.Is(err, repo.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return t, nil
}

// Delete removes a task by ID
func (s *taskService) Delete(ctx context.Context, id uuid.UUID) error {
    err := s.repository.Delete(ctx, id)
	if err != nil {
        if errors.Is(err, repo.ErrNotFound) {
			// It's okay if the task is already gone
			return nil
		}
		return err
	}
	
	return nil
}
