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
	@echo "[make] Installing project and test deps into venv (dev extras)"
	@$(PIP) install -e .[dev]

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
	@tools/install_local.sh $(sudo)



package-install-local: clean
	@echo "[make] Building distributions and installing from dist/ to user site"
	@tools/package_install_local.sh

build-standalone: clean
	@echo "[make] Building standalone zipapp artifact"
	@tools/build_zipapp.sh

build-standalone-full: clean
	@echo "[make] Building standalone zipapp with dependencies"
	@tools/build_zipapp_with_deps.sh

build-pex: clean
	@echo "[make] Building PEX artifact"
	@tools/build_pex.sh

build-shiv: clean
	@echo "[make] Building shiv artifact"
	@tools/build_shiv.sh
