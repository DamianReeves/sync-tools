import os
import tempfile
from click.testing import CliRunner
from sync_tools.cli import cli


def test_sync_dry_run_basic(monkeypatch):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        # create a file
        open(os.path.join(src, "keep.txt"), "w").write("keep")

        # mock subprocess.check_call to avoid needing rsync in test env
        called = {}
        def fake_check_call(cmd, *a, **kw):
            called['cmd'] = cmd
            return 0
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.check_call', fake_check_call)

        result = runner.invoke(cli, ['sync','--source', src, '--dest', dst, '--dry-run'])
        assert result.exit_code == 0
        assert 'rsync' in called['cmd'][0]


def test_sync_writes_dump(monkeypatch, tmp_path):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        open(os.path.join(src, "keep.txt"), "w").write("keep")

        called = {}
        def fake_check_call(cmd, *a, **kw):
            called['cmd'] = cmd
            return 0
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.check_call', fake_check_call)

        dump_file = tmp_path / "dump.json"
        result = runner.invoke(cli, ['sync','--source', src, '--dest', dst, '--dry-run', '--dump-commands', str(dump_file)])
        assert result.exit_code == 0
        assert dump_file.exists()
        txt = dump_file.read_text()
        assert 'cmd' in txt
        assert 'src' in txt
