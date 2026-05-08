#!/usr/bin/env bash
# Run the Go backend test suite.
# Usage: ./scripts/test.sh
# Must be invoked from the repository root.
set -euo pipefail

cd backend
go test ./...
