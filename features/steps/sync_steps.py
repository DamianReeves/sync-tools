import os
import subprocess
import tempfile
import shutil
from behave import given, when, then

@given('a source directory with files:')
def step_source_dir_with_files(context):
    """Create a temporary source directory and populate files from the table.

    Uses tempfile.TemporaryDirectory so directories are created outside the repo and
    are cleaned up after the scenario.
    """
    # Create a TemporaryDirectory object and save it so we can cleanup later
    context.source_tmpdir = tempfile.TemporaryDirectory(prefix='source-')
    context.source_dir = context.source_tmpdir.name
    for row in context.table:
        filename = row['filename']
        content = row['content']
        file_path = os.path.join(context.source_dir, filename)
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        with open(file_path, 'w') as f:
            f.write(content)

@given('an empty destination directory')
def step_empty_dest_dir(context):
    """Create a temporary empty destination directory using TemporaryDirectory."""
    context.dest_tmpdir = tempfile.TemporaryDirectory(prefix='dest-')
    context.dest_dir = context.dest_tmpdir.name

@when('I run sync.sh in one-way mode')
def step_run_sync_one_way(context):
    """Execute the sync.sh script with one-way mode."""
    script_path = os.path.abspath(os.path.join(os.getcwd(), 'sync.sh'))
    # Run the script explicitly with bash to ensure correct interpreter execution
    result = subprocess.run([
        'bash', script_path,
        '--source', context.source_dir,
        '--dest', context.dest_dir,
        '--mode', 'one-way'
    ], capture_output=True, text=True)
    # Record execution details for debugging; do not assert here to allow file verification
    context.returncode = result.returncode
    context.stdout = result.stdout
    context.stderr = result.stderr

@then('the destination directory should contain the files:')
def step_verify_dest_files(context):
    """Verify that the destination directory has the expected files with correct content."""
    for row in context.table:
        filename = row['filename']
        expected = row['content']
        dest_file = os.path.join(context.dest_dir, filename)
        assert os.path.exists(dest_file), f"Expected file missing: {dest_file}"
        with open(dest_file, 'r') as f:
            data = f.read()
        assert data == expected, f"Content mismatch for {filename}: expected '{expected}', got '{data}'"
    # Clean up TemporaryDirectory objects (they remove directories)
    try:
        context.source_tmpdir.cleanup()
    except Exception:
        pass
    try:
        context.dest_tmpdir.cleanup()
    except Exception:
        pass
