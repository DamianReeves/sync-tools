#!/usr/bin/env bash
# Build a shiv artifact that includes the package and resolved runtime dependencies.
# shiv resolves and vendors dependencies into a self-extracting zipapp.

set -euo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT=$(cd "$HERE/.." && pwd)
DIST="$ROOT/dist"
OUT="$DIST/sync-tools.shiv"

mkdir -p "$DIST"

echo "Building shiv at $OUT"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

python3 -m venv "$TMP/venv"
VENV_PY="$TMP/venv/bin/python"
VENV_PIP="$TMP/venv/bin/pip"
"$VENV_PIP" install --upgrade pip setuptools wheel >/dev/null
"$VENV_PIP" install shiv build >/dev/null

# Build our wheel first so shiv can include it directly
"$VENV_PY" -m build --wheel >/dev/null
WHEEL=$(ls "$DIST"/*.whl | head -n1)
if [ -z "${WHEEL:-}" ]; then
  echo "ERROR: No wheel found in $DIST after build" >&2
  exit 1
fi

# Read runtime deps from pyproject.toml and filter out test/dev packages
RUNTIME_DEPS=$("$VENV_PY" - <<'PY'
import tomllib, pathlib
p = pathlib.Path('pyproject.toml')
with p.open('rb') as f:
    deps = list(tomllib.load(f).get('project',{}).get('dependencies', []))
SKIP = { 'pytest', 'pytest-bdd', 'behave' }
print('\n'.join(d for d in deps if d.split()[0] not in SKIP))
PY
)

REQS_ARGS=()
if [ -n "$RUNTIME_DEPS" ]; then
  while IFS= read -r dep; do
    [ -n "$dep" ] && REQS_ARGS+=("$dep")
  done <<<"$RUNTIME_DEPS"
fi

# Build shiv: pass requirements and our wheel so shiv resolves and vendors them
"$TMP/venv/bin/shiv" "${REQS_ARGS[@]}" "$WHEEL" -o "$OUT" -e sync_tools:main -p "/usr/bin/env python3"
chmod +x "$OUT"

echo "Built $OUT"

echo "Try it:"
echo "  $OUT --help"
