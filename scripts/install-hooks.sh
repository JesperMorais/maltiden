#!/usr/bin/env bash
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
HOOK_SRC="$REPO_ROOT/scripts/pre-commit"
HOOK_DEST="$REPO_ROOT/.git/hooks/pre-commit"

chmod +x "$HOOK_SRC"
ln -sf "$HOOK_SRC" "$HOOK_DEST"

echo "Installed pre-commit hook: $HOOK_DEST -> $HOOK_SRC"
