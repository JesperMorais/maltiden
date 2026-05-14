# Måltiden dev tasks

.PHONY: help test backend-test frontend-build
.DEFAULT_GOAL := help

help:
	@echo "test             run all tests and frontend build"
	@echo "backend-test     run Go tests with race detector"
	@echo "frontend-build   type-check, lint, and build frontend"

test:
	cd backend && go test ./... -race
	cd frontend && npm run type-check && npm run lint && npm run build

backend-test:
	cd backend && go test ./... -race

frontend-build:
	cd frontend && npm run type-check && npm run lint && npm run build
