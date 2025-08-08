#!/bin/bash
# Script to run the service in development mode

echo "Starting task-svc development environment..."

# Check if docker and docker-compose are installed
if ! command -v docker &> /dev/null || ! command -v docker-compose &> /dev/null; then
    echo "Error: docker and docker-compose are required but not installed."
    exit 1
fi

# Start PostgreSQL if not already running
echo "Starting PostgreSQL container..."
docker-compose up -d postgres
sleep 3  # Give PostgreSQL time to initialize

# Apply migrations (with error handling for already existing objects)
echo "Applying database migrations..."
make migrate-up || echo "Note: Some migration errors are expected if objects already exist."

# Build and start the service in a Docker container
echo "Starting the service..."
docker-compose build app
docker-compose up app

echo "Development environment stopped."
