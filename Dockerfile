FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
# If go.sum exists, copy it as well
COPY go.sum* ./
# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
# BuildKit provides TARGETOS/TARGETARCH so we produce a binary for the current platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o /app/task-svc ./cmd/task-svc

# Create a minimal runtime image
FROM alpine:3.18

# Add ca-certificates for secure connections
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' appuser

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/task-svc .
# Copy migrations and docs
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/docs ./docs

# Use the non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/task-svc"]
