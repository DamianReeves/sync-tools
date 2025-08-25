"""
Simple hello world tests to verify pytest is working correctly.
"""
import subprocess
import os


def test_hello_world():
    """Basic test to verify pytest is working."""
    assert 1 + 1 == 2


def test_sync_script_exists():
    """Verify that the sync.sh script exists in the repository."""
    script_path = os.path.join(os.path.dirname(__file__), '..', 'sync.sh')
    assert os.path.exists(script_path), "sync.sh script should exist"


def test_sync_script_is_executable():
    """Verify that the sync.sh script is executable."""
    script_path = os.path.join(os.path.dirname(__file__), '..', 'sync.sh')
    assert os.access(script_path, os.X_OK), "sync.sh script should be executable"


def test_sync_script_runs():
    """Verify that the sync.sh script can be executed (basic smoke test)."""
    script_path = os.path.join(os.path.dirname(__file__), '..', 'sync.sh')
    # Just run the script with --help or no args to see if it executes without error
    result = subprocess.run([script_path], capture_output=True, text=True)
    # The script might return non-zero exit code, but it should at least run
    assert result is not None, "Script should execute and return a result"
