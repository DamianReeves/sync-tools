import tempfile
import shlex
import subprocess
import json
import logging
from pathlib import Path
from datetime import datetime, timezone
from typing import List, Tuple, Optional


def _ensure_slash_prefix(p: str) -> str:
    if not p.startswith("/"):
        return "/" + p
    return p


def to_filter_lines(patterns: List[str]) -> List[str]:
    """Convert a list of patterns (supporting leading !) to rsync filter lines.

    Rules:
    - Patterns starting with '!' are treated as explicit includes. For each
      such pattern we emit parent-directory includes, the path itself and a
      recursive include for children ("+ /path" and "+ /path/**").
    - Other patterns become simple exclude rules: "- pattern".
    - Empty/None entries are ignored.
    """
    out: List[str] = []
    for pat in patterns or []:
        if not pat:
            continue
        pat = pat.strip()
        if pat.startswith("!"):
            base = pat[1:]
            # strip any trailing /** used by users
            base = base.rstrip("/**").rstrip("/")
            base = _ensure_slash_prefix(base)
            # emit parent includes for each ancestor so rsync will traverse
            parts = base.lstrip("/").split("/")
            acc = ""
            for p in parts:
                acc += "/" + p
                out.append(f"+ {acc}")
            # include children recursively
            out.append(f"+ {base}/**")
        else:
            # allow the user to specify patterns as absolute (/foo) or relative
            # but write them as-is (rsync accepts both). Keep a dash prefix.
            out.append(f"- {pat}")
    return out


def build_filter_file(patterns: List[str], only: bool = False, default_excludes: Optional[List[str]] = None) -> Tuple[str, List[str]]:
    """Create a temporary filter file from patterns.

    Returns (filepath, lines_written). If `only` is True then this will
    generate whitelist-style filters (multiple + lines followed by `- *`).
    `default_excludes` can be appended to the output as plain filter lines.
    """
    lines: List[str] = []
    if only:
        # For whitelist mode, the provided patterns are treated as the only
        # things to include. Convert each to include lines (supporting '!')
        for pat in patterns or []:
            if not pat:
                continue
            p = pat.strip()
            # if user provided leading '!' normalize it away; in whitelist mode
            # everything is an include
            if p.startswith("!"):
                p = p[1:]
            p = p.rstrip("/")
            p = _ensure_slash_prefix(p)
            # emit parent includes
            parts = p.lstrip("/").split("/")
            acc = ""
            for part in parts:
                acc += "/" + part
                lines.append(f"+ {acc}")
            lines.append(f"+ {p}/**")
        # append default excludes before blocking everything else
        for ex in default_excludes or []:
            if ex:
                lines.append(f"- {ex}" if not ex.startswith(('+', '-')) else ex)
        # finally block everything else
        lines.append("- *")
    else:
        lines = to_filter_lines(patterns or [])


    tf = tempfile.NamedTemporaryFile(delete=False, mode="w", prefix="sync_filter_")
    tf.write("\n".join(lines))
    tf.flush()
    tf.close()
    return tf.name, lines


def run_rsync(src: str, dst: str, opts: List[str], src_filter: Optional[Tuple[str, List[str]]] = None,
              dst_filter: Optional[Tuple[str, List[str]]] = None, dump_commands: Optional[str] = None,
              logger: Optional[logging.Logger] = None):
    """Run rsync with the given options and optional filter files.

    src_filter and dst_filter are tuples (path, lines) as returned by
    `build_filter_file`. If `dump_commands` is a path we will write a JSON
    object containing the final rsync command and metadata which makes it
    easy to assert in tests.
    """
    cmd = ["rsync"] + list(opts)
    # place source filters first so source-side include/unignore rules take precedence
    if src_filter:
        cmd += ["--filter", f". {src_filter[0]}"]
    if dst_filter:
        cmd += ["--filter", f". {dst_filter[0]}"]
    # ensure trailing slash semantics (caller may provide either)
    cmd += [src.rstrip('/') + '/', dst.rstrip('/') + '/']

    if logger:
        logger.debug("Prepared rsync command: %s", " ".join(shlex.quote(x) for x in cmd))
    else:
        print("Running:", " ".join(shlex.quote(x) for x in cmd))

    if dump_commands:
        payload = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "src": src,
            "dst": dst,
            "opts": opts,
            "cmd": cmd,
            "src_filter": {"path": src_filter[0] if src_filter else None, "lines": src_filter[1] if src_filter else []},
            "dst_filter": {"path": dst_filter[0] if dst_filter else None, "lines": dst_filter[1] if dst_filter else []},
        }
        try:
            Path(dump_commands).write_text(json.dumps(payload, indent=2))
            if logger:
                logger.info("Wrote rsync command dump to %s", dump_commands)
        except Exception as e:
            if logger:
                logger.warning("Failed to write dump file %s: %s", dump_commands, e)

    subprocess.check_call(cmd)
