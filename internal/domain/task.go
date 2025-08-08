package domain

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a task
type Status string

// Status constants
const (
	StatusPending    Status = "Pending"
	StatusInProgress Status = "InProgress"
	StatusCompleted  Status = "Completed"
	StatusCancelled  Status = "Cancelled"
)

// Task represents a task in the system
type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description *string    `json:"description,omitempty"`
	Status      Status     `json:"status" validate:"oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Version     int        `json:"version"`
}

// CreateTaskRequest represents the payload for creating a new task
type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description *string    `json:"description,omitempty"`
	Status      *Status    `json:"status,omitempty" validate:"omitempty,oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// UpdateTaskRequest represents the payload for updating an existing task
type UpdateTaskRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description *string    `json:"description,omitempty"`
	Status      Status     `json:"status" validate:"required,oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Version     int        `json:"version" validate:"required,min=1"`
}

// PatchTaskRequest represents the payload for patching a task
type PatchTaskRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string    `json:"description,omitempty"`
	Status      *Status    `json:"status,omitempty" validate:"omitempty,oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Version     *int       `json:"version,omitempty" validate:"omitempty,min=1"`
}

// TaskList represents a paginated list of tasks
type TaskList struct {
	Items []Task    `json:"items"`
	Meta  PageMeta  `json:"meta"`
}

// PageMeta contains pagination metadata
type PageMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// TaskFilter represents filter options for listing tasks
type TaskFilter struct {
	Status *Status `json:"status,omitempty"`
	Page   int     `json:"page"`
	Size   int     `json:"size"`
	Sort   string  `json:"sort"`
}
