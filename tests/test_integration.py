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


def test_cli_list_filtered(monkeypatch, caplog):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        # create files in source
        open(os.path.join(src, "keep.txt"), "w").write("keep")
        open(os.path.join(src, "drop.txt"), "w").write("drop")

        # no rsync call expected; invoke CLI with list-filtered
        result = runner.invoke(cli, ['sync', '--source', src, '--dest', dst, '--list-filtered', 'src'])
        assert result.exit_code == 0
        # capture either stdout/stderr or logger output
        out = result.output
        if 'Filtered items' not in out:
            # fall back to logger capture
            logs = "\n".join([r.getMessage() for r in caplog.records])
            assert 'Filtered items' in logs


def test_cli_report_writes_file(monkeypatch, tmp_path):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        open(os.path.join(src, "keep.txt"), "w").write("keep")

        # mock subprocess.run to simulate rsync output
        class FakeProc:
            stdout = "+f+++++++++ keep.txt\n"
            stderr = ""
        def fake_run(cmd, stdout, stderr, text):
            return FakeProc()
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.run', fake_run)

        report_path = tmp_path / "report.md"
        result = runner.invoke(cli, ['sync', '--source', src, '--dest', dst, '--report', str(report_path)])
        assert result.exit_code == 0
        assert report_path.exists()
        txt = report_path.read_text()
        assert 'Added' in txt or 'Updated' in txt or 'Deleted' in txt
