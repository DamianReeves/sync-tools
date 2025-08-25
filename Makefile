SHELL := /bin/bash

.PHONY: help venv install bdd pytest test clean

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
	@echo "  make clean      - remove virtualenv and common temp files"

venv:
	@echo "[make] Ensuring virtualenv exists at $(VENV)"
	@if [ ! -d "$(VENV)" ]; then \
		python -m venv $(VENV); \
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
