# Integration and BDD Test Suite for sync.sh

This project uses [pytest](https://docs.pytest.org/), [behave](https://behave.readthedocs.io/), and [pytest-bdd](https://pytest-bdd.readthedocs.io/) to provide integration and behavior-driven tests for the `sync.sh` script.

## Structure

- `features/` - Gherkin feature files for BDD tests
- `features/steps/` - Step definitions for Behave
- `tests/` - Additional Python integration/unit tests

## Setup

This repository provides a `Makefile` to create the virtualenv and run tests.

```bash
# Create venv and install the project + deps
make install

# Run behave tests
make bdd

# Run pytest tests
make pytest
```

If you want to run the tools manually, activate the venv first:

```bash
source .venv/bin/activate
behave --no-capture
pytest -v
```

## Writing Tests
- Add `.feature` files to `features/`
- Add step definitions to `features/steps/`
- Add Python tests to `tests/`

## Notes
- The test suite is designed to test the `sync.sh` script as a black box, using temporary directories and subprocess calls.
