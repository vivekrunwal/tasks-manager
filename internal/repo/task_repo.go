package repo

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"task-svc/internal/domain"
	"task-svc/internal/platform/db"
)

// Common errors
var (
	ErrNotFound      = errors.New("task not found")
	ErrVersionConflict = errors.New("version conflict")
)

// TaskRepo handles database operations for tasks
type TaskRepo struct {
	db *db.Pool
}

// NewTaskRepo creates a new task repository
func NewTaskRepo(db *db.Pool) *TaskRepo {
	return &TaskRepo{db: db}
}

// Create inserts a new task into the database
func (r *TaskRepo) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, due_date, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		task.CreatedAt,
		task.UpdatedAt,
		task.Version,
	)

	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}

	return nil
}

// Get retrieves a task by ID
func (r *TaskRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, title, description, status, due_date, created_at, updated_at, version
		FROM tasks
		WHERE id = $1
	`

	var task domain.Task
	var description pgtype.Text
	var dueDate pgtype.Timestamptz

	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&description,
		&task.Status,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("selecting task: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		desc := description.String
		task.Description = &desc
	}

	if dueDate.Valid {
		due := dueDate.Time
		task.DueDate = &due
	}

	return &task, nil
}

// List retrieves tasks with pagination and filtering
func (r *TaskRepo) List(ctx context.Context, filter domain.TaskFilter) (*domain.TaskList, error) {
	// First get the total count for pagination metadata
	countQuery := `
		SELECT count(*) 
		FROM tasks 
		WHERE ($1::task_status IS NULL OR status = $1)
	`

	var total int
	err := r.db.QueryRow(ctx, countQuery, filter.Status).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting tasks: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(filter.Size)))
	
	// If there are no items, return an empty list with metadata
	if total == 0 {
		return &domain.TaskList{
			Items: []domain.Task{},
			Meta: domain.PageMeta{
				Page:       filter.Page,
				PageSize:   filter.Size,
				TotalItems: total,
				TotalPages: totalPages,
			},
		}, nil
	}

	// Otherwise fetch the requested page
	query := `
		SELECT id, title, description, status, due_date, created_at, updated_at, version
		FROM tasks
		WHERE ($1::task_status IS NULL OR status = $1)
		ORDER BY 
			CASE WHEN $2 = 'created_at' THEN created_at END ASC,
			CASE WHEN $2 = '-created_at' THEN created_at END DESC,
			CASE WHEN $2 = 'due_date' THEN due_date END ASC,
			CASE WHEN $2 = '-due_date' THEN due_date END DESC,
			created_at DESC
		LIMIT $3 OFFSET $4
	`

	offset := (filter.Page - 1) * filter.Size
	rows, err := r.db.Query(ctx, query, filter.Status, filter.Sort, filter.Size, offset)
	if err != nil {
		return nil, fmt.Errorf("selecting tasks: %w", err)
	}
	defer rows.Close()

	tasks := []domain.Task{}
	for rows.Next() {
		var task domain.Task
		var description pgtype.Text
		var dueDate pgtype.Timestamptz

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&description,
			&task.Status,
			&dueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			desc := description.String
			task.Description = &desc
		}

		if dueDate.Valid {
			due := dueDate.Time
			task.DueDate = &due
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating task rows: %w", err)
	}

	return &domain.TaskList{
		Items: tasks,
		Meta: domain.PageMeta{
			Page:       filter.Page,
			PageSize:   filter.Size,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

// Update updates a task with optimistic locking
func (r *TaskRepo) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET title = $2, description = $3, status = $4, due_date = $5
		WHERE id = $1 AND version = $6
		RETURNING version
	`

	var newVersion int
	err := r.db.QueryRow(ctx, query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		task.Version,
	).Scan(&newVersion)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Check if the task exists at all
			exists, checkErr := r.exists(ctx, task.ID)
			if checkErr != nil {
				return fmt.Errorf("checking task existence: %w", checkErr)
			}
			if !exists {
				return ErrNotFound
			}
			// Task exists but version doesn't match
			return ErrVersionConflict
		}
		return fmt.Errorf("updating task: %w", err)
	}

	task.Version = newVersion
	return nil
}

// Delete removes a task by ID
func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Patch applies a partial update to a task
func (r *TaskRepo) Patch(ctx context.Context, id uuid.UUID, patch *domain.PatchTaskRequest) (*domain.Task, error) {
	// First get the current task
	task, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check version if provided
	if patch.Version != nil && *patch.Version != task.Version {
		return nil, ErrVersionConflict
	}

	// Apply patch changes
	if patch.Title != nil {
		task.Title = *patch.Title
	}
	if patch.Description != nil {
		task.Description = patch.Description
	}
	if patch.Status != nil {
		task.Status = *patch.Status
	}
	if patch.DueDate != nil {
		task.DueDate = patch.DueDate
	}

	// Update in database
	query := `
		UPDATE tasks
		SET title = $2, description = $3, status = $4, due_date = $5
		WHERE id = $1 AND version = $6
		RETURNING version, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.DueDate,
		task.Version,
	).Scan(&task.Version, &task.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("patching task: %w", err)
	}

	return task, nil
}

// exists checks if a task with the given ID exists
func (r *TaskRepo) exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)`
	
	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking task existence: %w", err)
	}
	
	return exists, nil
}
