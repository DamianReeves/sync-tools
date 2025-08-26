#!/usr/bin/env bash
set -euo pipefail
# Usage: tools/install_local.sh [sudo]
SUDO_FLAG=${1:-}
if [ "${SUDO_FLAG}" = "1" ]; then
  echo "[install_local] Performing system-wide install (requires sudo)"
  sudo python3 -m pip install --upgrade pip setuptools wheel
  sudo python3 -m pip install .
  INSTALLED=$(sudo python3 -c 'import shutil,sys; print(shutil.which("sync-tools") or "")') || INSTALLED=""
  if [ -n "$INSTALLED" ]; then
    sudo ln -sf "$INSTALLED" /usr/local/bin/sync-tools || true
    echo "Installed system-wide (sync-tools available on PATH)"
  else
    echo "Warning: unable to locate installed 'sync-tools' executable as root; it may already be on PATH or require manual symlink"
  fi
else
  echo "[install_local] Performing user-local install"
  python3 -m pip install --user --upgrade pip setuptools wheel
  python3 -m pip install --user .
  mkdir -p "$HOME/.local/bin"
  EXE=$(python3 -c 'import shutil,sys; print(shutil.which("sync-tools") or "")') || EXE=""
  if [ -n "$EXE" ]; then
    ln -sf "$EXE" "$HOME/.local/bin/sync-tools"
    echo "Installed to user site; launcher symlinked to ~/.local/bin/sync-tools"
    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$HOME/.local/bin"; then
      echo "WARNING: ~/.local/bin is not on your PATH. Add this to your shell profile:";
      echo '  export PATH="$HOME/.local/bin:$PATH"'
    fi
  else
    echo "Failed to locate installed 'sync-tools' executable after install"
    exit 1
  fi
fi
