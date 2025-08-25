import os
import tempfile
from pathlib import Path
from click.testing import CliRunner
from sync_tools.cli import cli


def test_report_contents(monkeypatch, tmp_path):
    runner = CliRunner()
    with tempfile.TemporaryDirectory() as src, tempfile.TemporaryDirectory() as dst:
        Path(src, 'keep.txt').write_text('keep')
        # simulate rsync output lines that would indicate an added file
        class FakeProc:
            stdout = "+f+++++++++ keep.txt\n"
            stderr = ""
        def fake_run(cmd, stdout, stderr, text):
            return FakeProc()
        monkeypatch.setattr('sync_tools.rsync_wrapper.subprocess.run', fake_run)

        report_path = tmp_path / 'report.md'
        result = runner.invoke(cli, ['sync', '--source', src, '--dest', dst, '--report', str(report_path)])
        assert result.exit_code == 0
        text = report_path.read_text()
        assert '## Added' in text
        assert '## Excluded by filters' in text