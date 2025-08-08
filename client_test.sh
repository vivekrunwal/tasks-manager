#!/bin/bash
# Simple bash client to test task-svc functionality

BASE_URL="http://localhost:8080"

echo "Task Service Test Client"
echo "========================"
echo

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed. Please install it to continue."
    exit 1
fi

# Test health endpoint
echo "Testing health endpoint..."
curl -s ${BASE_URL}/healthz | jq
echo

# Create a task
echo "Creating a new task..."
TASK=$(curl -s -X POST ${BASE_URL}/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test the API",
    "description": "Make sure all endpoints are working correctly",
    "status": "InProgress",
    "due_date": "2023-12-31T23:59:59Z"
  }')
echo "$TASK" | jq
echo

# Extract the task ID and version
TASK_ID=$(echo $TASK | jq -r '.id')
VERSION=$(echo $TASK | jq -r '.version')

# List all tasks
echo "Listing all tasks..."
curl -s "${BASE_URL}/v1/tasks" | jq
echo

# Get specific task
echo "Getting task with ID: $TASK_ID"
curl -s "${BASE_URL}/v1/tasks/$TASK_ID" | jq
echo

# Update the task
echo "Updating task with ID: $TASK_ID"
UPDATED_TASK=$(curl -s -X PUT "${BASE_URL}/v1/tasks/$TASK_ID" \
  -H "Content-Type: application/json" \
  -d "{
    \"title\": \"Test the API thoroughly\",
    \"description\": \"Make sure all endpoints are working correctly with detailed tests\",
    \"status\": \"InProgress\",
    \"version\": $VERSION
  }")
echo "$UPDATED_TASK" | jq
echo

# Extract new version
NEW_VERSION=$(echo $UPDATED_TASK | jq -r '.version')

# Patch the task
echo "Patching task with ID: $TASK_ID"
PATCHED_TASK=$(curl -s -X PATCH "${BASE_URL}/v1/tasks/$TASK_ID" \
  -H "Content-Type: application/json" \
  -H "If-Match: $NEW_VERSION" \
  -d '{
    "status": "Completed"
  }')
echo "$PATCHED_TASK" | jq
echo

# List tasks with filter and pagination
echo "Listing completed tasks..."
curl -s "${BASE_URL}/v1/tasks?status=Completed&page=1&page_size=10" | jq
echo

# Delete the task
echo "Deleting task with ID: $TASK_ID"
DELETE_RESPONSE=$(curl -s -X DELETE -w "%{http_code}" "${BASE_URL}/v1/tasks/$TASK_ID")
if [[ $DELETE_RESPONSE -eq 204 ]]; then
    echo "Task deleted successfully (HTTP 204)"
else
    echo "Failed to delete task: $DELETE_RESPONSE"
fi
echo

# Check metrics endpoint
echo "Checking metrics endpoint..."
curl -s ${BASE_URL}/metrics | head -10
echo

echo "All tests completed!"
