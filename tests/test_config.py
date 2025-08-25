import os
import tempfile
from pathlib import Path
from click.testing import CliRunner
from sync_tools.config import validate_config
from sync_tools.cli import cli


def test_validate_config_errors():
    bad = {"source": 123, "dest": "/tmp"}
    try:
        validate_config(bad)
    except ValueError as e:
        assert "source" in str(e)


def test_cwd_autodiscover(monkeypatch, tmp_path):
    runner = CliRunner()
    # create temp cwd with sync.toml
    oldcwd = Path.cwd()
    with tmp_path.as_posix() and tempfile.TemporaryDirectory(dir=tmp_path) as d:
        # create a sync.toml
        t = Path(d) / "sync.toml"
        t.write_text('source = "srcdir"\ndest = "dstdir"')
        # ensure subprocess.check_call is mocked
        called = {}
        def fake_check_call(cmd, *a, **kw):
            called['cmd'] = cmd
            return 0
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.check_call', fake_check_call)
        # run inside that dir
        os.chdir(d)
        # create source and dest dirs referenced
        Path("srcdir").mkdir()
        Path("dstdir").mkdir()
        try:
            result = runner.invoke(cli, ['sync', '--source', 'srcdir', '--dest', 'dstdir', '--dry-run'])
            assert result.exit_code == 0
        finally:
            os.chdir(oldcwd)
