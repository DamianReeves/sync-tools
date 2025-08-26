# Developing sync-tools

This document explains how to set up your development environment to work on the sync-tools project and run the test suite.

## Project Structure

```
sync-tools/
├── sync.sh                    # Main bash script
├── ReadMe.adoc               # User documentation
├── LICENSE                   # MIT License
├── pyproject.toml           # Python project configuration
├── TESTING.md               # Test documentation
├── DEVELOPING.md            # This file
├── features/                # Cucumber/Behave feature files
│   ├── hello_world.feature  # Example BDD tests
│   └── steps/              # Step definitions
│       └── hello_world_steps.py
├── tests/                   # Python unit/integration tests
│   └── test_hello_world.py # Example pytest tests
└── test_folders/           # Test data and fixtures (if needed)
```

## Prerequisites

- **Python 3.8+** (3.12+ recommended)
- **Bash 4+** 
- **rsync 3.1+** (3.2+ recommended)
- **Git**

## Setting Up Your Development Environment

### 1. Clone and Navigate to the Repository

```bash
git clone https://github.com/DamianReeves/sync-tools.git
cd sync-tools
```

### 2. Create and Activate a Python Virtual Environment

The project uses Python virtual environments to isolate dependencies. You can use various tools:

#### Using Python's built-in venv (recommended)

```bash
# Create virtual environment
python3 -m venv .venv

# Activate it (Linux/macOS)
source .venv/bin/activate

# Activate it (Windows)
.venv\Scripts\activate
```

#### Using conda

```bash
conda create -n sync-tools python=3.12
conda activate sync-tools
```

### 3. Install Python Dependencies

Preferred: use the provided Makefile which creates the virtualenv and installs the project

```bash
# Create venv and install dependencies
make install

# Alternatively, if you prefer manual steps:
python3 -m venv .venv
source .venv/bin/activate
pip install -e .
```

### 4. Verify Installation

Check that the required tools are available:

```bash
# Check Python and tools
python3 --version
pytest --version
behave --version

# Check bash and rsync
bash --version
rsync --version

# Make sure sync.sh is executable
chmod +x sync.sh
```

## Running Tests

This project includes a `Makefile` with convenient targets. Recommended workflow:

```bash
# Create venv and install dependencies
make install

# Run BDD (behave)
make bdd

# Run pytest
make pytest

# Run both (BDD then pytest)
make test
```

You can also call the underlying tools directly if you prefer. When calling them directly, run them from the activated virtualenv:

```bash
# Activate venv
source .venv/bin/activate

# Behave
behave --no-capture

# Pytest
pytest -v

# With coverage
pip install pytest-cov
pytest --cov=. --cov-report=html
```

## Development Workflow

### 1. Writing Tests

#### BDD Tests (Cucumber/Behave)

1. Create `.feature` files in the `features/` directory using Gherkin syntax
2. Implement step definitions in `features/steps/` directory
3. Run `behave` to execute the tests

Example feature file:
```gherkin
Feature: Sync Script Functionality
  Scenario: Basic one-way sync
    Given I have a source directory with files
    And I have an empty destination directory  
    When I run sync.sh with one-way mode
    Then the destination should contain the source files
```

#### Python Integration Tests

1. Create test files in the `tests/` directory with `test_` prefix
2. Use pytest conventions and fixtures
3. Test the `sync.sh` script as a subprocess for integration testing

Example test:
```python
def test_sync_basic_functionality():
    """Test basic sync functionality."""
    # Setup test directories
    # Run sync.sh subprocess
    # Assert expected behavior
    pass
```

### 2. Testing Your Changes

Always run tests before committing:

```bash
# Quick smoke test
pytest tests/test_hello_world.py

# Full test suite
behave && pytest

# Test specific functionality
behave features/sync_functionality.feature
```

### 3. Commit Workflow

```bash
# Stage your changes
git add .

# Run tests before committing
behave && pytest

# Commit with descriptive message
git commit -m "Add: new sync feature with comprehensive tests"

# Push to your branch
git push origin feature-branch
```

## Development Tips

### Testing the sync.sh Script

Since `sync.sh` is a bash script, integration testing involves:

1. **Creating temporary directories** for testing
2. **Running the script as a subprocess** from Python
3. **Verifying file system state** after operations
4. **Testing various command-line options** and configurations

Example test pattern:
```python
import subprocess
import tempfile
import os

def test_sync_operation():
    with tempfile.TemporaryDirectory() as tmpdir:
        source_dir = os.path.join(tmpdir, "source")
        dest_dir = os.path.join(tmpdir, "dest")
        os.makedirs(source_dir)
        os.makedirs(dest_dir)
        
        # Create test files in source
        with open(os.path.join(source_dir, "test.txt"), "w") as f:
            f.write("test content")
        
        # Run sync.sh
        result = subprocess.run([
            "./sync.sh", 
            "--source", source_dir, 
            "--dest", dest_dir,
            "--dry-run"
        ], capture_output=True, text=True)
        
        # Assert expected behavior
        assert result.returncode == 0
```

### Debugging

- Use `--dry-run` flag extensively when testing sync operations
- Check the output of both stdout and stderr from subprocess calls
- Use pytest's `-s` flag to see print statements: `pytest -s`
- Use behave's `--no-capture` flag to see print statements: `behave --no-capture`

### Code Style

- Follow PEP 8 for Python code
- Use meaningful test names that describe the scenario
- Write clear Gherkin scenarios that are readable by non-developers
- Add docstrings to test functions explaining their purpose

### Virtual Environment Activation

Remember to always activate your virtual environment before working:

```bash
# If you forget to activate, you might see import errors
source .venv/bin/activate  # Linux/macOS
# or
.venv\Scripts\activate     # Windows
```

### IDE Setup

For VS Code, consider installing:
- Python extension
- Gherkin/Cucumber extension
- Bash extension
- Python Test Explorer

## Troubleshooting

### Common Issues

1. **"behave: command not found"**
   - Make sure your virtual environment is activated
   - Verify behave is installed: `pip list | grep behave`

2. **"Permission denied" when running sync.sh**
   - Make the script executable: `chmod +x sync.sh`

3. **Import errors in tests**
   - Ensure virtual environment is activated
   - Install project in editable mode: `pip install -e .`

4. **Tests failing unexpectedly**
   - Check if sync.sh has the right permissions
   - Verify rsync is available: `which rsync`
   - Run with verbose output to see details

### Getting Help

- Check existing tests for examples
- Review the `ReadMe.adoc` for sync.sh usage
- Look at the `TESTING.md` for test-specific information
- Create issues in the GitHub repository for bugs or questions

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Write tests for your changes
4. Implement your changes
5. Ensure all tests pass
6. Commit with clear messages
7. Push and create a Pull Request

Remember: Tests should pass on your local environment before submitting a PR!
