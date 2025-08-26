#!/usr/bin/env bash
# Build a single-file binary using PyInstaller.
# Produces dist/sync-tools suitable for Linux on the build machine.

set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT=$(cd "$HERE/.." && pwd)
DIST="$ROOT/dist"
OUT_NAME="sync-tools"
OUT_BIN="$DIST/$OUT_NAME"

mkdir -p "$DIST"

echo "Building PyInstaller binary at $OUT_BIN"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

python3 -m venv "$TMP/venv"
VENV_PY="$TMP/venv/bin/python"
VENV_PIP="$TMP/venv/bin/pip"
"$VENV_PIP" install --upgrade pip setuptools wheel >/dev/null
"$VENV_PIP" install pyinstaller build >/dev/null

# Ensure a wheel exists so we install exactly this project version
"$VENV_PY" -m build --wheel >/dev/null
WHEEL=$(ls "$DIST"/*.whl | head -n1)
if [ -z "${WHEEL:-}" ]; then
  echo "ERROR: No wheel found in $DIST after build" >&2
  exit 1
fi

# Install our wheel into the build venv (runtime deps only)
"$VENV_PIP" install "$WHEEL" >/dev/null

# Create a tiny runner script as the PyInstaller entrypoint
RUNNER="$TMP/run_sync_tools.py"
cat > "$RUNNER" <<'PY'
from sync_tools import main
if __name__ == "__main__":
    main()
PY

# Build single-file binary
"$TMP/venv/bin/pyinstaller" \
  --onefile \
  --name "$OUT_NAME" \
  --clean \
  --distpath "$DIST" \
  --workpath "$TMP/build" \
  --specpath "$TMP/spec" \
  "$RUNNER" >/dev/null

chmod +x "$OUT_BIN"

echo "Built $OUT_BIN"

echo "Try it:"
echo "  $OUT_BIN --help"
