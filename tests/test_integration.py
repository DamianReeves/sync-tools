import os
import json
import tempfile
from pathlib import Path
from click.testing import CliRunner
from sync_tools.cli import cli


def test_integration_basic_dump(monkeypatch, tmp_path):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        # create a file in source
        open(os.path.join(src, "keep.txt"), "w").write("keep")

        # mock subprocess.check_call to avoid calling real rsync
        called = {}
        def fake_check_call(cmd, *a, **kw):
            called['cmd'] = cmd
            return 0
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.check_call', fake_check_call)

        dump_file = tmp_path / "cmd_dump.json"
        result = runner.invoke(cli, ['sync', '--source', src, '--dest', dst, '--dry-run', '--dump-commands', str(dump_file)])
        assert result.exit_code == 0
        assert dump_file.exists()
        payload = json.loads(dump_file.read_text())
        assert 'cmd' in payload
        assert payload['src'] == str(Path(src).resolve())
        assert payload['dst'] == str(Path(dst).resolve())
        assert 'src_filter' in payload and 'dst_filter' in payload


def test_integration_whitelist_generates_filters(monkeypatch, tmp_path):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        # create files in source
        open(os.path.join(src, "keep.txt"), "w").write("keep")
        open(os.path.join(src, "drop.txt"), "w").write("drop")

        called = {}
        def fake_check_call(cmd, *a, **kw):
            called['cmd'] = cmd
            return 0
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.check_call', fake_check_call)

        dump_file = tmp_path / "cmd_dump.json"
        # Request whitelist-only for keep.txt
        result = runner.invoke(cli, ['sync', '--source', src, '--dest', dst, '--dry-run', '--only', 'keep.txt', '--dump-commands', str(dump_file)])
        assert result.exit_code == 0
        p = json.loads(dump_file.read_text())
        # src_filter lines should include include for keep.txt and a '- *' at end
        src_lines = p['src_filter']['lines']
        assert any(l.startswith('+') and 'keep.txt' in l for l in src_lines)
        assert '- *' in src_lines
