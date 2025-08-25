#!/usr/bin/env python3
"""rsync parity harness

This small utility runs rsync in dry-run mode with the same filter file
the Python code would pass, captures the list of files rsync would transfer,
and compares that to the decisions returned by `decision_for_path`.

Usage: python tools/rsync_parity_harness.py --src SRC --patterns file.txt
or:    python tools/rsync_parity_harness.py --src SRC --pattern "*.py" --pattern "!keep.py"
"""
import argparse
import subprocess
import tempfile
import os
from pathlib import Path
from typing import List, Set, Optional
import json
from datetime import datetime, timezone

from sync_tools.rsync_wrapper import build_filter_file, decision_for_path


def run_rsync_dryrun(src: str, filter_path: str) -> Set[str]:
    """Run rsync dry-run and return set of relative paths rsync would transfer.

    We run rsync from src -> a temporary empty dest directory and capture
    --out-format lines which include the filename. Files not listed are
    considered excluded by rsync under these filters.
    """
    tmp_dest = tempfile.TemporaryDirectory(prefix="rsync_parity_dest_")
    cmd = [
        "rsync",
        "-r",
        "--dry-run",
        "--out-format=%i %n",
        "--filter",
        f". {filter_path}",
        src.rstrip('/') + '/',
        tmp_dest.name.rstrip('/') + '/',
    ]
    proc = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    out = proc.stdout.splitlines()
    included = set()
    for ln in out:
        if not ln.strip():
            continue
        parts = ln.split(None, 1)
        if len(parts) == 2:
            name = parts[1].strip()
            # rsync prints paths relative to source root in this mode
            included.add(name)
    tmp_dest.cleanup()
    return included


def parity_check(src: str, patterns: List[str], dump_json: Optional[str] = None) -> int:
    """Build filter file, run rsync dry-run, compare to local matcher.

    Returns non-zero exit code if mismatches found.
    """
    filter_path, lines = build_filter_file(patterns)
    try:
        rsync_included = run_rsync_dryrun(src, filter_path)

        mismatches = []
        decisions = {}
        # Walk source and compare decisions
        for root, dirs, files in os.walk(src):
            for name in files:
                full = os.path.join(root, name)
                rel = os.path.relpath(full, src)
                py_dec = decision_for_path(rel, lines)
                rsync_sees = rel in rsync_included
                decisions[rel] = py_dec
                # Interpret rsync inclusion as 'include' if it's in the included set
                if py_dec == 'exclude' and rsync_sees:
                    mismatches.append({'path': rel, 'py': py_dec, 'rsync': 'include'})
                elif py_dec in ('neutral', 'include') and not rsync_sees:
                    mismatches.append({'path': rel, 'py': py_dec, 'rsync': 'exclude'})

        # Prepare diagnostic payload
        payload = {
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'src': src,
            'patterns': patterns,
            'filter_path': filter_path,
            'filter_lines': lines,
            'rsync_included': sorted(rsync_included),
            'python_decisions': decisions,
            'mismatches': mismatches,
        }

        # Write JSON diagnostics if requested or if mismatches found and no explicit path given
        if dump_json or mismatches:
            out_path = dump_json or (os.path.join(tempfile.gettempdir(), f"rsync_parity_{int(datetime.now().timestamp())}.json"))
            try:
                Path(out_path).write_text(json.dumps(payload, indent=2))
                print(f"Wrote parity diagnostics to {out_path}")
            except Exception as e:
                print(f"Failed to write diagnostics to {out_path}: {e}")

        if mismatches:
            print(f"Found {len(mismatches)} mismatches (written to {dump_json or 'auto'}):")
            for m in mismatches[:200]:
                print(" -", m['path'], 'py=', m['py'], 'rsync=', m['rsync'])
            print('\nNote: rsync inclusion set is based on --dry-run --out-format output; files not listed are treated as excluded by rsync under these filters.')
            return 2
        else:
            print("No mismatches: Python matcher and rsync dry-run decisions agree for all files.")
            return 0
    finally:
        try:
            os.unlink(filter_path)
        except Exception:
            pass


def main():
    ap = argparse.ArgumentParser(description="Run an rsync-backed parity check for filter matching")
    ap.add_argument("--src", required=True, help="Source directory to evaluate")
    group = ap.add_mutually_exclusive_group(required=True)
    group.add_argument("--patterns", help="Path to a newline-separated pattern file")
    group.add_argument("--pattern", dest="pattern", action='append', help="Inline pattern (can be passed multiple times)")
    args = ap.parse_args()

    if args.patterns:
        p = Path(args.patterns).read_text().splitlines()
        patterns = [x for x in (s.strip() for s in p) if x]
    else:
        patterns = args.pattern or []

    code = parity_check(args.src, patterns)
    raise SystemExit(code)


if __name__ == '__main__':
    main()
