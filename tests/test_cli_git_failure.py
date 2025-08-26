import subprocess
import tempfile
import os
from unittest import mock

import pytest

from sync_tools import cli


def test_git_clone_failure(monkeypatch, tmp_path):
    # Simulate git clone failing
    def fake_run(cmd, check, stdout, stderr, text):
        class R:
            returncode = 1
            stdout = ""
            stderr = "fatal: repository not found"
        return R()

    monkeypatch.setattr(subprocess, "run", fake_run)

    # Ensure tempfile.mkdtemp returns a path
    monkeypatch.setattr(tempfile, "mkdtemp", lambda prefix=None: str(tmp_path / 'repo'))

    called = {}
    def fake_run_rsync(src, dst, rsync_opts, **kwargs):
        called['ran'] = True

    monkeypatch.setattr(cli, 'run_rsync', fake_run_rsync)

    dst = str(tmp_path / 'dest')
    os.makedirs(dst, exist_ok=True)

    with pytest.raises(Exception):
        cli.sync.callback(config=None, source='https://github.com/nonexistent/repo.git', dest=dst, mode=None, dry_run=True,
                          use_source_gitignore=False, exclude_hidden_dirs=False, only_syncignore=False,
                          ignore_src=(), ignore_dest=(), only_items=(), v=0, log_level=None,
                          log_file=None, dump_commands=None, log_format='text', report=None, list_filtered=None)

    assert 'ran' not in called
