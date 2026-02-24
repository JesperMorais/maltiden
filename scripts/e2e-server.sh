#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Build frontend
echo "=== Building frontend ==="
npm run --prefix frontend build

# Copy build output to backend/static for SPA serving
echo "=== Setting up static files ==="
rm -rf backend/static
cp -r frontend/dist backend/static

# Ensure data directory exists
mkdir -p backend/data

# Remove old test DB for fresh state (migrations + seed data re-applied on startup)
rm -f backend/data/e2e-test.db

# Start Go server with test configuration
echo "=== Starting Go server on :8080 ==="
cd backend
DATABASE_PATH=./data/e2e-test.db \
JWT_SECRET=e2e-test-secret-key-must-be-at-least-32-characters \
CORS_ORIGINS=http://localhost:8080 \
exec go run cmd/server/main.go
