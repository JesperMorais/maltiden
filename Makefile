.PHONY: test build fmt help

help:
	@echo "Available targets:"
	@echo "  test  - run backend Go tests"
	@echo "  build - build backend Go packages"
	@echo "  fmt   - format backend Go code"

test:
	cd backend && go test ./...

build:
	cd backend && go build ./...

fmt:
	cd backend && go fmt ./...
