# Integration and BDD Test Suite for sync.sh

This project uses [pytest](https://docs.pytest.org/), [behave](https://behave.readthedocs.io/), and [pytest-bdd](https://pytest-bdd.readthedocs.io/) to provide integration and behavior-driven tests for the `sync.sh` script.

## Structure

- `features/` - Gherkin feature files for BDD tests
- `features/steps/` - Step definitions for Behave
- `tests/` - Additional Python integration/unit tests

## Setup

1. Install dependencies (already handled by `pyproject.toml`):
   ```bash
   pip install -r requirements.txt  # or use a tool like pipx/poetry if preferred
   ```
2. Run Behave BDD tests:
   ```bash
   behave
   ```
3. Run pytest tests:
   ```bash
   pytest
   ```

## Writing Tests
- Add `.feature` files to `features/`
- Add step definitions to `features/steps/`
- Add Python tests to `tests/`

## Notes
- The test suite is designed to test the `sync.sh` script as a black box, using temporary directories and subprocess calls.
