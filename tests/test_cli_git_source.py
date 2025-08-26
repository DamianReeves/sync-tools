import subprocess
import tempfile
import os
from unittest import mock

import pytest

from sync_tools import cli


def test_sync_from_git_url(monkeypatch, tmp_path):
    # Prepare a fake git clone by creating a temp dir that will act as clone_dir
    fake_clone = tmp_path / "repo-clone"
    fake_clone.mkdir()
    # create a dummy file inside to make it non-empty
    (fake_clone / "README.md").write_text("hello")

    # Mock subprocess.run to simulate successful git clone
    def fake_run(cmd, check, stdout, stderr, text):
        class R:
            returncode = 0
            stdout = "cloned"
            stderr = ""
        return R()

    monkeypatch.setattr(subprocess, "run", fake_run)

    # Monkeypatch tempfile.mkdtemp to return our fake_clone path as the clone dir
    monkeypatch.setattr(tempfile, "mkdtemp", lambda prefix=None: str(fake_clone))

    # Patch run_rsync so we can capture what source path it was called with
    called = {}

    def fake_run_rsync(src, dst, rsync_opts, **kwargs):
        called['src'] = src
        called['dst'] = dst

    monkeypatch.setattr(cli, "run_rsync", fake_run_rsync)

    # Run CLI sync with a git URL (we use a dummy .git URL)
    test_url = "https://github.com/example/repo.git"
    # Use tmp_path as destination
    dst = str(tmp_path / "dest")
    os.makedirs(dst, exist_ok=True)

    # Call the sync function directly
    cli.sync.callback = cli.sync.callback if hasattr(cli.sync, 'callback') else cli.sync
    # The CLI command function expects Click to call it, but we can call the function object directly
    cli.sync.callback(config=None, source=test_url, dest=dst, mode=None, dry_run=True,
                      use_source_gitignore=False, exclude_hidden_dirs=False, only_syncignore=False,
                      ignore_src=(), ignore_dest=(), only_items=(), v=0, log_level=None,
                      log_file=None, dump_commands=None, log_format='text', report=None, list_filtered=None)

    assert 'src' in called
    assert called['src'].endswith('repo-clone') or 'repo-clone' in called['src']
    assert called['dst'] == dst
