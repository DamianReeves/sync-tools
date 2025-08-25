import tempfile
import shlex
import subprocess
import json
import logging
import os
from pathlib import Path
from datetime import datetime, timezone
from typing import List, Tuple, Optional
import fnmatch


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
    includes: List[str] = []
    excludes: List[str] = []
    for pat in patterns or []:
        if not pat:
            continue
        pat = pat.strip()
        if pat.startswith("!"):
            base = pat[1:]
            # strip any trailing /** used by users
            base = base.rstrip("/**").rstrip("/")
            base = _ensure_slash_prefix(base)
            # ensure top-level directory is included so rsync will traverse
            includes.append("+ /")
            # emit parent includes for each ancestor so rsync will traverse
            parts = base.lstrip("/").split("/")
            acc = ""
            for p in parts:
                acc += "/" + p
                includes.append(f"+ {acc}")
            # include children recursively
            includes.append(f"+ {base}/**")
        else:
            # allow the user to specify patterns as absolute (/foo) or relative
            # but write them as-is (rsync accepts both). Keep a dash prefix.
            excludes.append(f"- {pat}")

    # Emit includes first so rsync will traverse into directories we may later
    # re-include; then emit excludes in original order.
    return includes + excludes


def parse_filter_lines(lines: List[str]) -> List[Tuple[str, str]]:
    """Parse lines like '+ /path' or '- pattern' into tuples (action, pattern)."""
    out = []
    for ln in lines or []:
        s = ln.strip()
        if not s:
            continue
        if s.startswith('+ '):
            out.append(('+', s[2:].strip()))
        elif s.startswith('- '):
            out.append(('-', s[2:].strip()))
        else:
            # allow raw patterns
            if s[0] in ('+', '-'):
                out.append((s[0], s[1:].strip()))
            else:
                out.append(('-', s))
    return out


def _match_pattern(pattern: str, relpath: str) -> bool:
    """Match a simplified rsync pattern against a relative path.

    This is intentionally conservative: it supports absolute-leading slash
    patterns (/a/b or /a/b/**) and shell-style wildcards. It does not
    implement every rsync nuance.
    """
    # normalize pattern
    pat = pattern.strip()
    if pat.startswith('./'):
        pat = pat[2:]
    # remove leading slash for internal matching
    absolute = pat.startswith('/')
    pat = pat.lstrip('/')

    # If pattern contains a slash, match against the full relative path.
    # Otherwise match against the basename only (rsync semantics).
    try:
        contains_slash = '/' in pat
    except Exception:
        contains_slash = False

    # treat /** as recursive wildcard for prefix matching
    if pat.endswith('/**'):
        base = pat[:-3].rstrip('/')
        if relpath == base or relpath.startswith(base + '/'):
            return True
        return False

    # trailing slash means directory
    if pat.endswith('/'):
        base = pat.rstrip('/')
        return relpath == base or relpath.startswith(base + '/')
    # choose subject to match
    subject = relpath if contains_slash else os.path.basename(relpath)

    # Convert rsync-style glob to regex. Rules:
    #  - '**' -> '.*' (match across slashes)
    #  - '*' -> '[^/]*' (do not match slashes)
    #  - '?' -> '[^/]'
    # Anchor to start/end depending on whether pattern was absolute
    regex = ''
    i = 0
    while i < len(pat):
        if pat[i:i+2] == '**':
            regex += '.*'
            i += 2
        else:
            c = pat[i]
            if c == '*':
                regex += '[^/]*'
            elif c == '?':
                regex += '[^/]'
            else:
                # escape literal characters (slashes, dots, etc.)
                try:
                    import re
                    regex += re.escape(c)
                except Exception:
                    regex += c
            i += 1

    # wrap regex
    rx = r'^' + regex + r'$'

    try:
        import re
        cre = re.compile(rx)
        return cre.match(subject) is not None
    except Exception:
        # fallback to simple fnmatch for safety
        glob_pat = pat.replace('**', '*')
        return fnmatch.fnmatchcase(subject, glob_pat)


def decision_for_path(relpath: str, filter_lines: List[str]) -> str:
    """Return 'include'|'exclude'|'neutral' for a relative path according to filter_lines order."""
    parsed = parse_filter_lines(filter_lines)
    decision = 'neutral'
    for action, pat in parsed:
        if _match_pattern(pat, relpath):
            decision = 'include' if action == '+' else 'exclude'
    return decision


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
            # ensure top-level directory is included so rsync will traverse
            lines.append(f"+ /")
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
              logger: Optional[logging.Logger] = None, report_path: Optional[str] = None,
              list_filtered: Optional[str] = None):
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

    # If user asked to list only filtered items, perform that analysis by
    # walking the filesystem and evaluating our filter rules locally. We do
    # not invoke rsync for this mode.
    if list_filtered:
        lists = {}
        if list_filtered in ("src", "both"):
            src_lines = src_filter[1] if src_filter else []
            found = []
            for root, dirs, files in os.walk(src):
                for name in files:
                    full = os.path.join(root, name)
                    rel = os.path.relpath(full, src)
                    dec = decision_for_path(rel, src_lines)
                    if dec == 'exclude':
                        found.append(rel)
            lists['src'] = found
        if list_filtered in ("dst", "both"):
            dst_lines = dst_filter[1] if dst_filter else []
            found = []
            for root, dirs, files in os.walk(dst):
                for name in files:
                    full = os.path.join(root, name)
                    rel = os.path.relpath(full, dst)
                    dec = decision_for_path(rel, dst_lines)
                    if dec == 'exclude':
                        found.append(rel)
            lists['dst'] = found
        # write a simple markdown report to stdout or logger
        md = [f"# Filtered items (mode={list_filtered})\n"]
        for k, arr in lists.items():
            md.append(f"## {k}\n")
            if not arr:
                md.append("(none)\n")
            else:
                for p in sorted(arr):
                    md.append(f"- {p}\n")
        out = "\n".join(md)
        if logger:
            logger.info("Filtered list:\n%s", out)
        else:
            print(out)
        return

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

    # If a report is requested, capture rsync output and create a markdown
    # summary. Otherwise, run as before.
    if report_path:
        # ensure rsync prints itemized changes and names
        cmd_with_out = cmd + ["--out-format=%i %n"]
        proc = subprocess.run(cmd_with_out, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        out = proc.stdout.splitlines()
        added, updated, deleted = [], [], []
        for ln in out:
            if not ln.strip():
                continue
            parts = ln.split(None, 1)
            if len(parts) == 1:
                continue
            it, name = parts[0], parts[1].strip()
            if it.startswith("*") or it.startswith("deleting") or 'deleting' in ln:
                deleted.append(name)
            elif '+' in it:
                # created or transferred
                added.append(name)
            else:
                updated.append(name)

        # Determine excluded items by evaluating filters over the source tree
        excluded = []
        src_lines = src_filter[1] if src_filter else []
        for root, dirs, files in os.walk(src):
            for name in files:
                rel = os.path.relpath(os.path.join(root, name), src)
                if decision_for_path(rel, src_lines) == 'exclude':
                    excluded.append(rel)

        md = [f"# Sync report\nGenerated: {datetime.now(timezone.utc).isoformat()}\n", f"- src: {src}", f"- dst: {dst}", f"- dry-run: {'--dry-run' in opts or '--dry-run' in ' '.join(opts)}\n"]
        if added:
            md.append("\n## Added\n")
            md += [f"- {p}" for p in sorted(added)]
        if updated:
            md.append("\n## Updated\n")
            md += [f"- {p}" for p in sorted(updated)]
        if deleted:
            md.append("\n## Deleted\n")
            md += [f"- {p}" for p in sorted(deleted)]
        md.append("\n## Excluded by filters\n")
        if not excluded:
            md.append("(none)\n")
        else:
            md += [f"- {p}" for p in sorted(excluded)]

        Path(report_path).write_text("\n".join(md))
        if logger:
            logger.info("Wrote report to %s", report_path)
        else:
            print(f"Wrote report to {report_path}")
        return
    else:
        subprocess.check_call(cmd)
