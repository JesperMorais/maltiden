#!/usr/bin/env bash
# Run the Go backend test suite.
# Usage: ./scripts/test.sh
# Must be invoked from the repository root.
set -euo pipefail

cd backend
# -v + -race matches the CI command in .github/workflows/ci.yml; the race
# detector catches concurrent-access bugs locally before CI does.
go test -v -race ./...
