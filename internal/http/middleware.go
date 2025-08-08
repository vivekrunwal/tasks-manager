package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

// RequestIDKey is the context key for the request ID
type contextKey string
const RequestIDKey = contextKey("requestID")

// Middleware sets up all middlewares for the API
func Middleware(r chi.Router, logger *slog.Logger) {
	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware for development
	r.Use(corsMiddleware)
}

// requestLogger logs the request details
func requestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			// Get the request ID from the context and add it to response header
			requestID := middleware.GetReqID(r.Context())
			if requestID != "" {
				ww.Header().Set("X-Request-ID", requestID)
				// Store it in the context for downstream handlers
				ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
				r = r.WithContext(ctx)
			}
			
			defer func() {
				// Log request details after completion
				logger.Info("Request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"size", ww.BytesWritten(),
					"duration", time.Since(start).String(),
					"request_id", requestID,
				)
			}()
			
			next.ServeHTTP(ww, r)
		})
	}
}

// corsMiddleware handles Cross-Origin Resource Sharing
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Request-ID, If-Match")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// DBConnectionMiddleware ensures the DB connection is available
func DBConnectionMiddleware(isDBConnected *bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isDBConnected == nil || !*isDBConnected {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Database connection not available"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
