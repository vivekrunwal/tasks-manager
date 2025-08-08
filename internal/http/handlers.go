package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log/slog"

	"task-svc/internal/config"
	"task-svc/internal/domain"
	"task-svc/internal/repo"
	"task-svc/internal/service"
	"task-svc/pkg/pagination"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	service     service.TaskService
	paginationConfig config.PaginationConfig
	logger      *slog.Logger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(service service.TaskService, cfg config.PaginationConfig, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		service:     service,
		paginationConfig: cfg,
		logger:      logger,
	}
}

// errorResponse represents an error response
type errorResponse struct {
	Error struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

// respondWithJSON sends a JSON response
func (h *TaskHandler) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("Error encoding response", "error", err)
		}
	}
}

// respondWithError sends an error response
func (h *TaskHandler) respondWithError(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
	resp := errorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message
	resp.Error.Details = details

	h.respondWithJSON(w, status, resp)
}

// CreateTask handles POST /v1/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "bad_request", "Invalid request body", nil)
		return
	}

	// Validate the request
	if err := Validate.Struct(req); err != nil {
		validationErrors := parseValidationErrors(err)
		h.respondWithError(w, http.StatusBadRequest, "validation_error", "Validation failed", validationErrors)
		return
	}

	task, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create task", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to create task", nil)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, task)
}

// GetTask handles GET /v1/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "Invalid task ID", nil)
		return
	}

	task, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "not_found", "Task not found", nil)
			return
		}
		h.logger.Error("Failed to get task", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve task", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// ListTasks handles GET /v1/tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pageNum, pageSize := pagination.Parse(
		r.URL.Query().Get("page"),
		r.URL.Query().Get("page_size"),
		pagination.DefaultParams{
			DefaultPage: 1,
			DefaultSize: h.paginationConfig.DefaultSize,
			MaxSize:     h.paginationConfig.MaxSize,
		},
	)

	// Parse status filter
	var status *domain.Status
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		s := domain.Status(statusParam)
		if !isValidStatus(s) {
			h.respondWithError(w, http.StatusBadRequest, "invalid_status", "Invalid status value", nil)
			return
		}
		status = &s
	}

	// Parse sort parameter
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "-created_at" // Default sort
	}
	if !isValidSortField(sort) {
		h.respondWithError(w, http.StatusBadRequest, "invalid_sort", "Invalid sort parameter", nil)
		return
	}

	filter := domain.TaskFilter{
		Status: status,
		Page:   pageNum,
		Size:   pageSize,
		Sort:   sort,
	}

	result, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tasks", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to list tasks", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, result)
}

// UpdateTask handles PUT /v1/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "Invalid task ID", nil)
		return
	}

	var req domain.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "bad_request", "Invalid request body", nil)
		return
	}

	// Check for version in header if not in body
	if req.Version == 0 {
		if ifMatch := r.Header.Get("If-Match"); ifMatch != "" {
			version, err := strconv.Atoi(ifMatch)
			if err == nil && version > 0 {
				req.Version = version
			}
		}
	}

	// Validate the request
	if err := Validate.Struct(req); err != nil {
		validationErrors := parseValidationErrors(err)
		h.respondWithError(w, http.StatusBadRequest, "validation_error", "Validation failed", validationErrors)
		return
	}

	task, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "not_found", "Task not found", nil)
			return
		}
		if errors.Is(err, repo.ErrVersionConflict) {
			h.respondWithError(w, http.StatusConflict, "version_conflict", "Task was modified by another request", nil)
			return
		}
		h.logger.Error("Failed to update task", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to update task", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// PatchTask handles PATCH /v1/tasks/{id}
func (h *TaskHandler) PatchTask(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "Invalid task ID", nil)
		return
	}

	var req domain.PatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "bad_request", "Invalid request body", nil)
		return
	}

	// Check for version in header if not in body
	if req.Version == nil {
		if ifMatch := r.Header.Get("If-Match"); ifMatch != "" {
			version, err := strconv.Atoi(ifMatch)
			if err == nil && version > 0 {
				req.Version = &version
			}
		}
	}

	// Validate the request
	if err := Validate.Struct(req); err != nil {
		validationErrors := parseValidationErrors(err)
		h.respondWithError(w, http.StatusBadRequest, "validation_error", "Validation failed", validationErrors)
		return
	}

	task, err := h.service.Patch(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.respondWithError(w, http.StatusNotFound, "not_found", "Task not found", nil)
			return
		}
		if errors.Is(err, repo.ErrVersionConflict) {
			h.respondWithError(w, http.StatusConflict, "version_conflict", "Task was modified by another request", nil)
			return
		}
		h.logger.Error("Failed to patch task", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to patch task", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// DeleteTask handles DELETE /v1/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid_id", "Invalid task ID", nil)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete task", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to delete task", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isValidStatus checks if the status is valid
func isValidStatus(status domain.Status) bool {
	validStatuses := []domain.Status{
		domain.StatusPending,
		domain.StatusInProgress,
		domain.StatusCompleted,
		domain.StatusCancelled,
	}

	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// isValidSortField checks if the sort field is valid
func isValidSortField(sort string) bool {
	validSortFields := []string{"created_at", "-created_at", "due_date", "-due_date"}
	for _, s := range validSortFields {
		if sort == s {
			return true
		}
	}
	return false
}

// RegisterRoutes registers all task routes
func (h *TaskHandler) RegisterRoutes(r chi.Router) {
    r.Route("/tasks", func(r chi.Router) {
		r.Post("/", h.CreateTask)
		r.Get("/", h.ListTasks)
		r.Get("/{id}", h.GetTask)
		r.Put("/{id}", h.UpdateTask)
		r.Patch("/{id}", h.PatchTask)
		r.Delete("/{id}", h.DeleteTask)
	})
}

// Health handles GET /healthz
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Ready handles GET /readyz
// This checks if the application is ready to handle requests
func Ready(db *http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		if db == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
