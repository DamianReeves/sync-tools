import tempfile
import os
from unittest import mock

import shutil

import pytest

from sync_tools import cli


def test_sync_from_zip_url(monkeypatch, tmp_path):
    # Prepare a fake extracted dir
    fake_extracted = tmp_path / "extracted"
    fake_extracted.mkdir()
    (fake_extracted / "docs").mkdir()
    (fake_extracted / "docs" / "readme.md").write_text("x")

    # Mock urllib.request.urlopen to return a file-like object
    class FakeResp:
        def __init__(self, path):
            self._f = open(path, "rb")
        def __enter__(self):
            return self._f
        def __exit__(self, exc_type, exc, tb):
            self._f.close()

    # Monkeypatch urlopen to return a real small zip created on-the-fly
    zip_path = tmp_path / "archive.zip"
    shutil.make_archive(str(zip_path.with_suffix('')), 'zip', root_dir=str(fake_extracted))

    monkeypatch.setattr('urllib.request.urlopen', lambda url: FakeResp(str(zip_path)))

    # Monkeypatch tempfile.NamedTemporaryFile to produce a real filename
    tmpfile = tmp_path / "dl.zip"
    def _named_tempfile(**kwargs):
        class TF:
            def __init__(self, name):
                self.name = name
            def close(self):
                pass
        return TF(str(tmpfile))
    monkeypatch.setattr(tempfile, 'NamedTemporaryFile', _named_tempfile)

    # Monkeypatch tempfile.mkdtemp to return our extraction dir parent
    extract_root = tmp_path / "extract_root"
    extract_root.mkdir()
    # We'll let the real shutil.unpack_archive run which will create a dir under extract_root
    monkeypatch.setattr(tempfile, 'mkdtemp', lambda prefix=None: str(extract_root))

    # Patch run_rsync to capture source/dest
    called = {}
    def fake_run_rsync(src, dst, rsync_opts, **kwargs):
        called['src'] = src
        called['dst'] = dst
    monkeypatch.setattr(cli, 'run_rsync', fake_run_rsync)

    dst = str(tmp_path / 'dest')
    os.makedirs(dst, exist_ok=True)

    # Use a fake URL
    test_url = 'https://example.com/archive.zip'

    # Call CLI sync function
    cli.sync.callback(config=None, source=test_url, dest=dst, mode=None, dry_run=True,
                      use_source_gitignore=False, exclude_hidden_dirs=False, only_syncignore=False,
                      ignore_src=(), ignore_dest=(), only_items=(), v=0, log_level=None,
                      log_file=None, dump_commands=None, log_format='text', report=None, list_filtered=None)

    assert 'src' in called
    assert called['dst'] == dst
    # ensure src points inside extract_root
    assert str(extract_root) in str(called['src'])
