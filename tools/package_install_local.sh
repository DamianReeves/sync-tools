#!/usr/bin/env bash
set -euo pipefail

# Build distributions and install from dist/ to user site (or venv when available)
echo "[package_install_local] Building distributions and installing from dist/ to user site"
echo "[package_install_local] Removing any existing ~/.local/bin/sync-tools symlink to avoid stale targets"
rm -f "${HOME}/.local/bin/sync-tools" || true

# Determine python executor: prefer project venv python if present
VENV=".venv"
PY_VENV="$VENV/bin/python"
if [ -x "$PY_VENV" ]; then
  PYEXEC="$PY_VENV"
else
  PYEXEC=python3
fi

# Ensure the chosen python has pip
if ! $PYEXEC -m pip --version >/dev/null 2>&1; then
  echo "ERROR: pip is not available for $PYEXEC. Please install system pip or python3-venv and retry." >&2
  exit 1
fi

# Ensure build tool
$PYEXEC -m pip install --upgrade build
$PYEXEC -m build --sdist --wheel

# Choose wheel if available, otherwise fall back to sdist
WHEEL=$(ls dist/*.whl 2>/dev/null | head -n1 || true)
if [ -n "$WHEEL" ]; then
  TARGET="$WHEEL"
else
  SDIST=$(ls dist/*.tar.gz 2>/dev/null | head -n1 || true)
  TARGET="$SDIST"
fi

if [ -z "$TARGET" ]; then
  echo "ERROR: No distribution found in dist/ to install" >&2
  exit 1
fi

if [ "$PYEXEC" = "$PY_VENV" ]; then
  $PYEXEC -m pip install "$TARGET"
else
  $PYEXEC -m pip install --user "$TARGET"
fi

mkdir -p "$HOME/.local/bin"
# If venv produced an executable, symlink that directly into ~/.local/bin
if [ -x "$VENV/bin/sync-tools" ]; then
  ln -sf "$(pwd)/$VENV/bin/sync-tools" "$HOME/.local/bin/sync-tools"
  echo "Symlinked venv executable to ~/.local/bin/sync-tools"
else
  EXE=$($PYEXEC -c 'import shutil,sys; print(shutil.which("sync-tools") or "")')
  if [ -n "$EXE" ]; then
    ln -sf "$EXE" "$HOME/.local/bin/sync-tools"
    echo "Installed from dist/ to user site; launcher symlinked to ~/.local/bin/sync-tools"
  else
    echo "Failed to locate installed 'sync-tools' executable after package install"
    exit 1
  fi
fi
