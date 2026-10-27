#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../backend"
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "Unformatted Go files:"
  echo "$unformatted"
  exit 1
fi
