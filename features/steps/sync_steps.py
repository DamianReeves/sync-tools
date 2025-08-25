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


@given('a destination directory with files:')
def step_dest_dir_with_files(context):
    """Create a temporary destination directory and populate files from the table."""
    # Create TemporaryDirectory for destination if not already created
    if not hasattr(context, 'dest_tmpdir'):
        context.dest_tmpdir = tempfile.TemporaryDirectory(prefix='dest-')
        context.dest_dir = context.dest_tmpdir.name
    for row in context.table:
        filename = row['filename']
        content = row['content']
        file_path = os.path.join(context.dest_dir, filename)
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        with open(file_path, 'w') as f:
            f.write(content)


@given('a source .syncignore with:')
def step_source_syncignore(context):
    """Write a .syncignore file in the source directory from a multiline block."""
    text = context.text or ""
    fpath = os.path.join(context.source_dir, '.syncignore')
    with open(fpath, 'w') as f:
        f.write(text)


@given('a dest .syncignore with:')
def step_dest_syncignore(context):
    text = context.text or ""
    fpath = os.path.join(context.dest_dir, '.syncignore')
    with open(fpath, 'w') as f:
        f.write(text)

@given('an empty destination directory')
def step_empty_dest_dir(context):
    """Create a temporary empty destination directory using TemporaryDirectory."""
    context.dest_tmpdir = tempfile.TemporaryDirectory(prefix='dest-')
    context.dest_dir = context.dest_tmpdir.name

@when('I run sync.sh in one-way mode')
def step_run_sync_one_way(context):
    """Execute the sync.sh script with one-way mode."""
    _run_sync_with_mode(context, mode='one-way')


@when('I run sync.sh in two-way mode')
def step_run_sync_two_way(context):
    """Execute the sync.sh script with two-way mode."""
    _run_sync_with_mode(context, mode='two-way')


def _run_sync_with_mode(context, mode='one-way'):
    script_path = os.path.abspath(os.path.join(os.getcwd(), 'sync.sh'))
    # Build command
    cmd = ['bash', script_path, '--source', context.source_dir, '--dest', context.dest_dir, '--mode', mode]
    # Include any whitelist items passed in context
    if hasattr(context, 'only_items') and context.only_items:
        for item in context.only_items:
            cmd += ['--only', item]
    result = subprocess.run(cmd, capture_output=True, text=True)
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


@then('a conflict file should exist for "{filename}" on the source')
def step_conflict_file_exists(context, filename):
    """Assert that a conflict copy was preserved on the source for filename."""
    # Look for files that match filename.conflict-*
    base = os.path.join(context.source_dir, filename)
    found = False
    for p in os.listdir(os.path.dirname(base) or context.source_dir):
        full = os.path.join(os.path.dirname(base) or context.source_dir, p)
        if p.startswith(os.path.basename(base) + '.conflict-'):
            found = True
            break
    assert found, f"Expected conflict file for {filename} in {context.source_dir}"


@given('I whitelist the paths:')
def step_whitelist_paths(context):
    """Set whitelist items for the upcoming sync run from a table of paths."""
    context.only_items = [row['path'] for row in context.table]
