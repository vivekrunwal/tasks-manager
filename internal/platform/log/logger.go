package log

import (
	"io"
	"log/slog"
	"os"

	"task-svc/internal/config"
)

// Setup initializes the logger based on the application environment
func Setup(cfg config.AppConfig) *slog.Logger {
	var handler slog.Handler

	// Configure logging format based on environment
	if cfg.Env == "prod" {
		// JSON format for production
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Text format for development
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// Writer returns an io writer for use with HTTP server logging middleware
func Writer() io.Writer {
	return os.Stdout
}
