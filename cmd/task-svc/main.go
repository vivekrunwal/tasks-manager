package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"task-svc/internal/config"
	httphandlers "task-svc/internal/http"
	"task-svc/internal/platform/db"
	applog "task-svc/internal/platform/log"
	"task-svc/internal/platform/metrics"
	"task-svc/internal/repo"
	"task-svc/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	logger := applog.Setup(cfg.App)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to the database
	dbPool, err := db.New(ctx, cfg.DB)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Database is connected
	isDBConnected := true

	// Setup repositories
	taskRepo := repo.NewTaskRepo(dbPool)

	// Setup services
	taskService := service.NewTaskService(taskRepo)

	// Setup HTTP handlers
	taskHandler := httphandlers.NewTaskHandler(taskService, cfg.Pagination, logger)

	// Create router
	r := chi.NewRouter()

	// Setup middlewares
	httphandlers.Middleware(r, logger)
	r.Use(metrics.Middleware())

	// Register routes
	r.Route("/v1", func(r chi.Router) {
		// Apply database connection middleware to API routes
		r.Use(httphandlers.DBConnectionMiddleware(&isDBConnected))
		
		// Register task routes
		taskHandler.RegisterRoutes(r)
	})

	// Serve OpenAPI UI and spec
	r.Get("/docs", httphandlers.SwaggerUI)
	r.Get("/docs/openapi.yaml", httphandlers.OpenAPISpec)

	// Health and metrics endpoints
	r.Get("/healthz", httphandlers.Health)
	r.Get("/readyz", httphandlers.Ready(nil))
	r.Handle("/metrics", promhttp.Handler())

	// Create the server
	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting server", "addr", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down server...")

	// Shutdown server with timeout
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// Set database connection to false
	isDBConnected = false
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Server gracefully stopped")
}
