#!/usr/bin/env bash
# Optional helper to run the installed console script from source tree without installing.
# Usage: ./tools/sync-tools-launcher.sh sync --source ./src --dest ./dst
PYTHONPATH="$(pwd)" python -m sync_tools.cli "$@"
