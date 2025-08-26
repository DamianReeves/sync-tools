#!/usr/bin/env bash
# Optional helper to run the installed console script or fallback to module runner.
# Usage: ./tools/sync-tools-launcher.sh sync --source ./src --dest ./dst

set -euo pipefail

# Prefer an installed 'sync-tools' on PATH
if command -v sync-tools >/dev/null 2>&1; then
	exec sync-tools "$@"
fi

# Otherwise run using the repository source as PYTHONPATH
export PYTHONPATH="$(pwd)"
exec python3 -m sync_tools.cli "$@"
