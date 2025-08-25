import os
import subprocess
import sys
from pathlib import Path


def test_parity_harness_runs(tmp_path):
    # Ensure the harness script executes and returns either 0 (no mismatches)
    # or 2 (mismatches found) so CI can record parity diffs without failing.
    harness = Path('tools/rsync_parity_harness.py')
    assert harness.exists(), 'harness script missing'

    # create a small tree
    d = tmp_path / "src"
    d.mkdir()
    (d / "a.txt").write_text("hello")
    (d / "keep.txt").write_text("keep")

    env = os.environ.copy()
    env['PYTHONPATH'] = os.getcwd()
    cmd = [sys.executable, str(harness), '--src', str(d), '--pattern', '*.txt', '--pattern', '!keep.txt']
    proc = subprocess.run(cmd, env=env)
    assert proc.returncode in (0, 2), f'harness exited with unexpected code {proc.returncode}'
