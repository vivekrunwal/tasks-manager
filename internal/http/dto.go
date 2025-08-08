package http

import (
	"time"

	"task-svc/internal/domain"
)

// HTTP-layer DTOs decouple API payload/response from domain models.

// CreateTaskPayload represents the HTTP request body to create a task
type CreateTaskPayload struct {
	Title       string         `json:"title" validate:"required,min=1,max=200"`
	Description *string        `json:"description,omitempty"`
	Status      *domain.Status `json:"status,omitempty" validate:"omitempty,oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
}

// UpdateTaskPayload represents the HTTP request body to update a task
type UpdateTaskPayload struct {
	Title       string        `json:"title" validate:"required,min=1,max=200"`
	Description *string       `json:"description,omitempty"`
	Status      domain.Status `json:"status" validate:"required,oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
	Version     int           `json:"version" validate:"required,min=1"`
}

// PatchTaskPayload represents the HTTP request body to partially update a task
type PatchTaskPayload struct {
	Title       *string        `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string        `json:"description,omitempty"`
	Status      *domain.Status `json:"status,omitempty" validate:"omitempty,oneof=Pending InProgress Completed Cancelled"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
	Version     *int           `json:"version,omitempty" validate:"omitempty,min=1"`
}

// ToDomain converts CreateTaskPayload to domain.CreateTaskRequest
func (p CreateTaskPayload) ToDomain() domain.CreateTaskRequest {
	return domain.CreateTaskRequest{
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		DueDate:     p.DueDate,
	}
}

// ToDomain converts UpdateTaskPayload to domain.UpdateTaskRequest
func (p UpdateTaskPayload) ToDomain() domain.UpdateTaskRequest {
	return domain.UpdateTaskRequest{
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		DueDate:     p.DueDate,
		Version:     p.Version,
	}
}

// ToDomain converts PatchTaskPayload to domain.PatchTaskRequest
func (p PatchTaskPayload) ToDomain() domain.PatchTaskRequest {
	return domain.PatchTaskRequest{
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		DueDate:     p.DueDate,
		Version:     p.Version,
	}
}

// TaskResponse is the response shape for a task
type TaskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	DueDate     *string `json:"due_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Version     int     `json:"version"`
}

// TaskListResponse wraps a list of tasks along with pagination metadata
type TaskListResponse struct {
	Items []TaskResponse `json:"items"`
	Meta  PageMeta       `json:"meta"`
}

// PageMeta mirrors domain.PageMeta but keeps the HTTP boundary explicit
type PageMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// fromDomainTask maps a domain.Task to TaskResponse
func fromDomainTask(t domain.Task) TaskResponse {
	var dueStr *string
	if t.DueDate != nil {
		s := t.DueDate.UTC().Format(time.RFC3339)
		dueStr = &s
	}
	return TaskResponse{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		DueDate:     dueStr,
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.UTC().Format(time.RFC3339),
		Version:     t.Version,
	}
}

// fromDomainTaskList maps a domain.TaskList to TaskListResponse
func fromDomainTaskList(l *domain.TaskList) TaskListResponse {
	items := make([]TaskResponse, 0, len(l.Items))
	for _, it := range l.Items {
		items = append(items, fromDomainTask(it))
	}
	return TaskListResponse{
		Items: items,
		Meta: PageMeta{
			Page:       l.Meta.Page,
			PageSize:   l.Meta.PageSize,
			TotalItems: l.Meta.TotalItems,
			TotalPages: l.Meta.TotalPages,
		},
	}
}

// no custom time layout constant; using time.RFC3339
