#!/usr/bin/env bash
# Run the Go backend test suite.
# Usage: ./scripts/test.sh
# Must be invoked from the repository root.
set -euo pipefail

cd backend
# -v + -race per CLAUDE.md's canonical CI command — race detector catches
# concurrent-access bugs locally that CI would otherwise be the first to see.
go test -v -race ./...
