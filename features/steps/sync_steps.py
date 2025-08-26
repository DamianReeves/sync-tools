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


@given('a source .gitignore with:')
def step_source_gitignore(context):
    """Write a .gitignore file in the source directory from a multiline block."""
    text = context.text or ""
    fpath = os.path.join(context.source_dir, '.gitignore')
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

@when('I run sync-tools sync in one-way mode')
def step_run_sync_one_way(context):
    """Execute the Python CLI with one-way mode (legacy wording kept in feature)."""
    _run_sync_with_mode(context, mode='one-way')


@when('I run sync-tools sync in two-way mode')
def step_run_sync_two_way(context):
    """Execute the Python CLI with two-way mode (legacy wording kept in feature)."""
    _run_sync_with_mode(context, mode='two-way')


def _run_sync_with_mode(context, mode='one-way'):
    # Prefer venv python if present, else fallback to system python3
    venv_py = os.path.join(os.getcwd(), '.venv', 'bin', 'python')
    py = venv_py if os.path.exists(venv_py) else 'python3'
    # Build command to run the package module (triggers __main__.py -> CLI)
    cmd = [py, '-m', 'sync_tools', 'sync', '--source', context.source_dir, '--dest', context.dest_dir, '--mode', mode]
    # Include any whitelist items passed in context
    if hasattr(context, 'only_items') and context.only_items:
        for item in context.only_items:
            cmd += ['--only', item]
    # Include any extra CLI args passed via context.extra_args
    if hasattr(context, 'extra_args') and context.extra_args:
        for a in context.extra_args:
            # if a is a tuple (arg, val) expand, else append single token
            if isinstance(a, (list, tuple)):
                cmd += [str(a[0]), str(a[1])]
            else:
                cmd += [str(a)]
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


@then('the destination directory should not contain the files:')
def step_verify_dest_files_not_present(context):
    """Verify that the destination directory does NOT contain the listed files."""
    for row in context.table:
        filename = row['filename']
        dest_file = os.path.join(context.dest_dir, filename)
        assert not os.path.exists(dest_file), f"Unexpected file present: {dest_file}"


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
@when('I whitelist the paths:')
def step_whitelist_paths(context):
    """Set whitelist items for the upcoming sync run from a table of paths."""
    context.only_items = [row['path'] for row in context.table]


@given('I add extra args:')
@when('I add extra args:')
def step_add_extra_args(context):
    """Add extra CLI arguments for the next sync run from a table of arg/value."""
    context.extra_args = []
    for row in context.table:
        arg = row.get('arg')
        val = row.get('value')
        if val is None or val == '':
            context.extra_args.append(arg)
        else:
            context.extra_args.append((arg, val))

# New step for initializing a git repository as source
@given('a git repository with files:')
def step_git_repo_with_files(context):
    import tempfile, os
    # Create a temp directory for the git repo named with .git suffix to trigger clone
    parent = tempfile.TemporaryDirectory(prefix='gitparent-')
    repo_dir = os.path.join(parent.name, 'source.git')
    os.makedirs(repo_dir, exist_ok=True)
    context.gitparent = parent
    context.git_repo_dir = repo_dir
    # Populate files
    for row in context.table:
        filename = row['filename']
        content = row['content']
        path = os.path.join(repo_dir, filename)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, 'w') as f:
            f.write(content)
    # Initialize git repo and commit
    subprocess.run(['git', 'init'], cwd=repo_dir, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    subprocess.run(['git', 'add', '.'], cwd=repo_dir, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    subprocess.run(['git', 'commit', '-m', 'initial'], cwd=repo_dir, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    # Set source_dir to the git URL path
    context.source_dir = repo_dir


@then('the report file should contain:')
def step_report_file_contains(context):
    import os
    # Find report file argument from extra_args
    report_path = None
    for arg in getattr(context, 'extra_args', []):
        if isinstance(arg, (list, tuple)) and arg[0] == '--report':
            report_path = arg[1]
            break
    assert report_path, 'No --report argument provided'
    abs_path = os.path.abspath(report_path)
    assert os.path.exists(abs_path), f'Report file not found: {abs_path}'
    content = open(abs_path).read()
    for row in context.table:
        line = row['line']
        assert line in content, f'Expected line {line!r} in report'
    # Cleanup report file
    try:
        os.unlink(abs_path)
    except Exception:
        pass
