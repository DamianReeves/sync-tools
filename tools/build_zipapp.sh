#!/usr/bin/env bash
# Build a self-contained zipapp (pyz) from the package. This avoids installing
# into a virtualenv and creates an executable archive that runs with system
# Python 3. Requirements and some platform compiled extensions won't be
# included; this is intended for pure-Python packages and the project's own
# modules.

set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT=$(cd "$HERE/.." && pwd)
DIST="$ROOT/dist"
PYZ="$DIST/sync-tools.pyz"

mkdir -p "$DIST"

echo "Building zipapp at $PYZ"

# Ensure package is importable by building a temporary layout
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# Copy package sources and necessary files into tempdir
rsync -a --exclude='*.pyc' --exclude='__pycache__' --exclude='.git' "$ROOT/" "$TMPDIR/" >/dev/null

# Use python -m zipapp to build a pyz that runs as a module (uses __main__.py)
python3 -m zipapp "$TMPDIR" -o "$PYZ" -m 'sync_tools:main'

chmod +x "$PYZ"

echo "Built $PYZ"

# Provide quick usage notes
cat <<EOF
To run the standalone archive:
  python3 $PYZ [args...]
or make it executable and run directly:
  $PYZ [args...]

Note: This bundles package code only. If your project depends on third-party
pure-Python packages, you must vendor them into the repo or use a different
packaging strategy (wheel or installer). Compiled extensions will not be
included.
EOF
