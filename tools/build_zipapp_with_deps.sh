#!/usr/bin/env bash
# Build a zipapp that includes runtime dependencies so it can run without a
# virtualenv. Only suitable for pure-Python dependencies. Compiled extensions
# and system libraries cannot be bundled this way reliably.

set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT=$(cd "$HERE/.." && pwd)
DIST="$ROOT/dist"
OUT="$DIST/sync-tools-full.pyz"

mkdir -p "$DIST"

echo "Building full zipapp at $OUT"

TMPVENV=$(mktemp -d)
trap 'rm -rf "$TMPVENV"' EXIT

python3 -m venv "$TMPVENV/venv"
# Use the venv pip to install runtime deps into the temporary venv
VENV_PY="$TMPVENV/venv/bin/python"
VENV_PIP="$TMPVENV/venv/bin/pip"

# Upgrade pip and install project runtime dependencies using pyproject.toml
"$VENV_PIP" install --upgrade pip setuptools wheel >/dev/null

# Extract runtime dependencies from pyproject.toml using python
RUNTIME_DEPS=$("$VENV_PY" -c "import tomllib,sys, pathlib
p=pathlib.Path('pyproject.toml')
with p.open('rb') as f:
    data=tomllib.load(f)
print('\n'.join(data.get('project',{}).get('dependencies',[])))")

if [ -n "$RUNTIME_DEPS" ]; then
    echo "Installing runtime deps into temporary venv:"
    echo "$RUNTIME_DEPS"
    printf '%s\n' $RUNTIME_DEPS | xargs "$VENV_PIP" install --no-deps -U >/dev/null
fi

# Prepare a tempdir to assemble the zipapp contents
PAYLOAD=$(mktemp -d)

# Copy site-packages from venv to payload
SITE_PACKAGES_DIR=$("$VENV_PY" -c "import site,sys; print(next(p for p in site.getsitepackages() if 'site-packages' in p))")
rsync -a --exclude='__pycache__' "$SITE_PACKAGES_DIR/" "$PAYLOAD/" >/dev/null

# Copy package source into payload
rsync -a --exclude='*.pyc' --exclude='__pycache__' --exclude='.git' "$ROOT/sync_tools/" "$PAYLOAD/sync_tools/" >/dev/null

# Build the zipapp
python3 -m zipapp "$PAYLOAD" -o "$OUT" -m 'sync_tools:main'
chmod +x "$OUT"

echo "Built $OUT"

cat <<EOF
To run the full standalone archive:
  python3 $OUT [args...]
or make it executable and run directly:
  $OUT [args...]

Note: Only pure-Python runtime dependencies are bundled. If you rely on
compiled extensions, consider using a wheel installer or platform packaging.
EOF
