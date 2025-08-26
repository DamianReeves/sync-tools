SHELL := /bin/bash

.PHONY: help venv install bdd pytest test clean
.PHONY: install-local package-install-local

VENV := .venv
PY := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

BEHAVE_ARGS ?=
PYTEST_ARGS ?=

help:
	@echo "Makefile for sync-tools"
	@echo
	@echo "Targets:"
	@echo "  make venv       - create virtualenv and upgrade packaging tools"
	@echo "  make install    - install project and test dependencies into venv"
	@echo "  make bdd        - run behave BDD tests (set BEHAVE_ARGS to pass extra args)"
	@echo "  make pytest     - run pytest (set PYTEST_ARGS to pass extra args)"
	@echo "  make test       - run behave then pytest"
	@echo "  make install-local [sudo=1] - install the package locally and make the 'sync-tools' launcher available (~/.local/bin), optionally install system-wide with sudo=1"

venv:
	@echo "[make] Ensuring virtualenv exists at $(VENV)"
	@if [ ! -d "$(VENV)" ]; then \
		python3 -m venv $(VENV); \
	fi
	@echo "[make] Upgrading pip, setuptools, wheel in venv"
	@$(PY) -m pip install --upgrade pip setuptools wheel

install: venv
	@echo "[make] Installing project and test deps into venv"
	@$(PIP) install -e .

bdd: venv
	@echo "[make] Running behave BDD tests"
	@$(PY) -m behave $(BEHAVE_ARGS)

pytest: venv
	@echo "[make] Running pytest"
	@$(PY) -m pytest $(PYTEST_ARGS)

test: bdd pytest

clean:
	@echo "[make] Cleaning virtualenv and temporary files"
	@rm -rf $(VENV) .pytest_cache behave-results reports


install-local:
	@echo "[make] Installing package for local user (or system-wide if sudo=1)"
	@if [ "$(sudo)" = "1" ]; then \
		# system-wide install (requires sudo)
		sudo python3 -m pip install --upgrade pip setuptools wheel; \
		sudo python3 -m pip install .; \
		if [ -f "$(VENV)/bin/sync-tools" ]; then true; fi; \
		sudo ln -sf "$(shell python3 -c 'import shutil,sys; print(shutil.which("sync-tools") or "")')" /usr/local/bin/sync-tools || true; \
		echo "Installed system-wide (sync-tools available on PATH)"; \
	else \
		python3 -m pip install --user --upgrade pip setuptools wheel; \
		python3 -m pip install --user .; \
		# ensure ~/.local/bin exists
		mkdir -p $${HOME}/.local/bin; \
		EXE=$$(python3 -c 'import shutil,sys; print(shutil.which("sync-tools") or "")'); \
		if [ -n "$$EXE" ]; then \
			ln -sf "$$EXE" $${HOME}/.local/bin/sync-tools; \
			echo "Installed to user site; launcher symlinked to ~/.local/bin/sync-tools"; \
			if ! echo "$${PATH}" | tr ':' '\n' | grep -qx "$${HOME}/.local/bin"; then \
				echo "WARNING: ~/.local/bin is not on your PATH. Add this to your shell profile:"; \
				echo '  export PATH="$$HOME/.local/bin:$$PATH"'; \
			fi; \
		else \
			echo "Failed to locate installed 'sync-tools' executable after install"; \
		fi; \
	fi



package-install-local: clean
	@echo "[make] Building distributions and installing from dist/ to user site"
	@echo "[make] Removing any existing ~/.local/bin/sync-tools symlink to avoid stale targets"
	@rm -f $${HOME}/.local/bin/sync-tools || true; \
	# Try to create virtualenv; if it fails, fall back to system python3 (if pip is available)
	@if [ ! -x "$(PY)" ]; then \
		if ! $(MAKE) venv; then \
			echo "[make] Warning: failed to create venv, will attempt system python3 fallback"; \
		fi; \
	fi; \
	if [ -x "$(PY)" ]; then \
		PYEXEC="$(PY)"; \
	else \
		PYEXEC=python3; \
	fi; \
	# Ensure the chosen python has pip
	if ! $$PYEXEC -m pip --version >/dev/null 2>&1; then \
		echo "ERROR: pip is not available for $$PYEXEC. Please install system pip or python3-venv and retry." >&2; \
		exit 1; \
	fi; \
	$$PYEXEC -m pip install --upgrade build; \
	$$PYEXEC -m build --sdist --wheel; \
	# Choose wheel if available, otherwise fall back to sdist. Install into venv when using project python,
	# otherwise install to user site. This avoids pip trying to install both wheel+sdist and raising
	# ResolutionImpossible due to duplicate package spec sources.
	WHEEL=$$(ls dist/*.whl 2>/dev/null | head -n1 || true); \
	if [ -n "$$WHEEL" ]; then \
		TARGET="$$WHEEL"; \
	else \
		SDIST=$$(ls dist/*.tar.gz 2>/dev/null | head -n1 || true); \
		TARGET="$$SDIST"; \
	fi; \
	if [ -z "$$TARGET" ]; then \
		echo "ERROR: No distribution found in dist/ to install" >&2; \
		exit 1; \
	fi; \
	if [ "$$PYEXEC" = "$(PY)" ]; then \
		$$PYEXEC -m pip install "$$TARGET"; \
	else \
		$$PYEXEC -m pip install --user "$$TARGET"; \
	fi; \
	mkdir -p $${HOME}/.local/bin; \
	# If venv produced an executable, symlink that directly into ~/.local/bin
	if [ -x "$(VENV)/bin/sync-tools" ]; then \
		ln -sf "$(CURDIR)/$(VENV)/bin/sync-tools" $${HOME}/.local/bin/sync-tools; \
		echo "Symlinked venv executable to ~/.local/bin/sync-tools"; \
	else \
		EXE=$$($$PYEXEC -c 'import shutil,sys; print(shutil.which("sync-tools") or "")'); \
		if [ -n "$$EXE" ]; then \
			ln -sf "$$EXE" $${HOME}/.local/bin/sync-tools; \
			echo "Installed from dist/ to user site; launcher symlinked to ~/.local/bin/sync-tools"; \
		else \
			echo "Failed to locate installed 'sync-tools' executable after package install"; \
		fi; \
	fi
