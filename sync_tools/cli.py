import click
import subprocess
import tempfile
import urllib.request
import shutil
import os
import logging
import json
from pathlib import Path
from typing import List

try:
    import tomllib as _toml
except Exception:  # pragma: no cover - fallback for <3.11
    import tomli as _toml  # type: ignore

from .rsync_wrapper import build_filter_file, run_rsync
from .config import load_toml_path, validate_config


def _configure_logging(level: int, log_file: str | None, log_format: str = "text") -> logging.Logger:
    logger = logging.getLogger("sync_tools")
    logger.setLevel(level)
    if logger.handlers:
        return logger

    if log_format == "json":
        class SimpleJSONFormatter(logging.Formatter):
            def format(self, record):
                d = {
                    "time": self.formatTime(record),
                    "level": record.levelname,
                    "name": record.name,
                    "msg": record.getMessage(),
                }
                return json.dumps(d)
        formatter = SimpleJSONFormatter()
    else:
        formatter = logging.Formatter("%(asctime)s %(levelname)s %(name)s: %(message)s")

    handler = logging.FileHandler(log_file) if log_file else logging.StreamHandler()
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    return logger


@click.group()
def cli():
    """sync-tools CLI - wrapper around rsync with .syncignore, whitelist, and filters"""
    pass


@cli.command()
@click.option("--config", type=click.Path(exists=True), help="Path to a TOML config file to load default options")
@click.option("--source", required=False, type=click.Path(exists=True))
@click.option("--dest", required=False, type=click.Path())
@click.option("--mode", type=click.Choice(["one-way", "two-way"]), default=None)
@click.option("--dry-run", is_flag=True)
@click.option("--use-source-gitignore", is_flag=True)
@click.option("--exclude-hidden-dirs", is_flag=True)
@click.option("--only-syncignore", is_flag=True)
@click.option("--ignore-src", multiple=True)
@click.option("--ignore-dest", multiple=True)
@click.option("--only", "only_items", multiple=True)
@click.option("-v", "-V", count=True)
@click.option("--log-level", type=click.Choice(["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]), default=None,
              help="Explicit log level (overrides -v)")
@click.option("--log-file", type=click.Path(), default=None, help="Path to write logs")
@click.option("--dump-commands", type=click.Path(), default=None, help="Write rsync command and filters to JSON file")
@click.option("--log-format", type=click.Choice(["text", "json"]), default="text", help="Log format")
@click.option("--report", type=click.Path(), default=None, help="Write a markdown sync report to this path")
@click.option("--list-filtered", type=click.Choice(["src", "dst", "both"]), default=None, help="List items that would be filtered (src, dst or both)")
def sync(config, source, dest, mode, dry_run, use_source_gitignore, exclude_hidden_dirs, only_syncignore,
         ignore_src, ignore_dest, only_items, v, log_level, log_file, dump_commands, log_format, report, list_filtered):
    """Perform a sync between SOURCE and DEST using rsync with layered filters.

    You can specify defaults in a TOML config and override them on the command line.
    """
    cfg = {}

    downloaded_zip_path = None
    extracted_dir = None
    clone_dir = None

    def _looks_like_git_url(u: str) -> bool:
        if not u:
            return False
        return u.startswith("git@") or u.startswith("git://") or u.endswith(".git") or ("github.com" in u and ".zip" not in u)

    if source and _looks_like_git_url(source):
        try:
            clone_dir = tempfile.mkdtemp(prefix="sync_tools_git_clone_")
            res = subprocess.run(["git", "clone", "--depth", "1", source, clone_dir], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            if getattr(res, "returncode", 1) != 0:
                raise click.BadParameter(f"Failed to clone git repo {source}: {getattr(res, 'stderr', '') or getattr(res, 'stdout', '')}")
            source = str(Path(clone_dir).resolve())
        except click.BadParameter:
            raise
        except Exception as e:
            raise click.BadParameter(f"Failed to clone git repo {source}: {e}")

    if source and (source.startswith("http://") or source.startswith("https://")):
        try:
            tf = tempfile.NamedTemporaryFile(delete=False, suffix=".zip")
            tf.close()
            with urllib.request.urlopen(source) as resp, open(tf.name, "wb") as out:
                shutil.copyfileobj(resp, out)
            downloaded_zip_path = tf.name
            extracted_dir = tempfile.mkdtemp(prefix="sync_tools_zip_")
            shutil.unpack_archive(downloaded_zip_path, extracted_dir)
            parts = [p for p in Path(extracted_dir).iterdir() if p.exists()]
            if len(parts) == 1 and parts[0].is_dir():
                source = str(parts[0].resolve())
            else:
                source = str(Path(extracted_dir).resolve())
        except Exception as e:
            raise click.BadParameter(f"Failed to download/extract source URL: {e}")

    if config:
        with open(config, "rb") as f:
            cfg = _toml.load(f)

    def _cfg(key, default=None):
        return cfg.get(key, default)

    src = str(Path(source).resolve()) if source else None
    dst = str(Path(dest).resolve()) if dest else None

    if not cfg and src:
        for c in (Path(src) / "sync.toml", Path(src) / ".sync.toml"):
            if c.exists():
                cfg = load_toml_path(c)
                break

    if not cfg:
        for c in (Path.cwd() / "sync.toml", Path.cwd() / ".sync.toml"):
            if c.exists():
                cfg = load_toml_path(c)
                break

    if cfg:
        try:
            validate_config(cfg)
        except ValueError as e:
            raise click.BadParameter(f"Invalid config file: {e}")

    if not src:
        s = _cfg("source")
        src = str(Path(s).resolve()) if s else None
    if not dst:
        d = _cfg("dest")
        dst = str(Path(d).resolve()) if d else None

    if not src or not dst:
        raise click.BadParameter("source and dest must be provided either via CLI or config file")

    mode = mode or _cfg("mode", "one-way")
    dry_run = dry_run or _cfg("dry_run", False)
    use_source_gitignore = use_source_gitignore or _cfg("use_source_gitignore", False)
    exclude_hidden_dirs = exclude_hidden_dirs or _cfg("exclude_hidden_dirs", False)
    only_syncignore = only_syncignore or _cfg("only_syncignore", False)

    ignore_src_list: List[str] = list(ignore_src) if ignore_src else list(_cfg("ignore_src", []))
    ignore_dest_list: List[str] = list(ignore_dest) if ignore_dest else list(_cfg("ignore_dest", []))
    only_list: List[str] = list(only_items) if only_items else list(_cfg("only", []))

    if log_level:
        level = getattr(logging, log_level)
    else:
        level = max(10, 20 - (10 * v))

    logger = _configure_logging(level, log_file, log_format)
    logger.debug("CLI options after merge: %s", {"src": src, "dst": dst, "mode": mode, "dry_run": dry_run})

    # Require rsync only for real executions (not list-filtered, not dry-run).
    # This lets tests run without rsync when monkeypatching subprocess.
    if shutil.which("rsync") is None and not list_filtered and not dry_run:
        raise click.ClickException(
            "rsync not found on PATH. Please install rsync to use sync-tools (e.g., 'sudo apt-get install -y rsync')."
        )

    rsync_opts = ["-a", "--human-readable", "--itemize-changes", "--partial"]
    if dry_run:
        rsync_opts.append("--dry-run")
    rsync_opts.append("--checksum")

    default_excludes = ["- /.git/"]
    if exclude_hidden_dirs:
        default_excludes.append("- .*/")

    src_filter = None
    dst_filter = None
    src_tmp = None
    dst_tmp = None
    try:
        src_patterns: List[str] = []
        syncignore_path = Path(src) / ".syncignore"
        if syncignore_path.exists():
            src_patterns += [ln.strip() for ln in syncignore_path.read_text().splitlines() if ln.strip()]
        if use_source_gitignore:
            gitignore_path = Path(src) / ".gitignore"
            if gitignore_path.exists():
                src_patterns += [ln.strip() for ln in gitignore_path.read_text().splitlines() if ln.strip()]
        src_patterns += list(ignore_src_list)

        if only_list:
            src_tmp, src_lines = build_filter_file(only_list, only=True, default_excludes=default_excludes)
            src_filter = (src_tmp, src_lines)
        else:
            src_tmp, src_lines = build_filter_file(src_patterns, only=False, default_excludes=default_excludes)
            src_filter = (src_tmp, src_lines)

        if ignore_dest_list:
            dst_tmp, dst_lines = build_filter_file(ignore_dest_list, only=False, default_excludes=default_excludes)
            dst_filter = (dst_tmp, dst_lines)

        if mode == "two-way":
            # Pre-flight: if a file exists on both sides and contents differ,
            # preserve the DESTINATION version as a conflict copy on the SOURCE.
            # This matches the BDD expectation: destination changes are preserved
            # as conflict files in the source when both sides differ.
            for root, _, files in os.walk(dst):
                for name in files:
                    d_full = os.path.join(root, name)
                    rel = os.path.relpath(d_full, dst)
                    s_full = os.path.join(src, rel)
                    if os.path.exists(s_full):
                        try:
                            with open(s_full, 'rb') as fs, open(d_full, 'rb') as fd:
                                s_bytes = fs.read()
                                d_bytes = fd.read()
                            if s_bytes != d_bytes:
                                ts = int(Path(d_full).stat().st_mtime)
                                conflict = os.path.join(os.path.dirname(s_full), os.path.basename(s_full) + f".conflict-{ts}")
                                try:
                                    shutil.copy2(d_full, conflict)
                                except Exception:
                                    shutil.copy(d_full, conflict)
                        except Exception:
                            # best-effort; ignore errors in conflict preservation
                            pass
            # First, copy src -> dst (normal direction)
            run_rsync(src, dst, rsync_opts, src_filter=src_filter, dst_filter=dst_filter, dump_commands=dump_commands, logger=logger)
            # Then propagate dst -> src to bring source up to date with destination changes
            run_rsync(dst, src, rsync_opts, src_filter=dst_filter, dst_filter=src_filter, dump_commands=dump_commands, logger=logger)
        else:
            run_rsync(
                src,
                dst,
                rsync_opts,
                src_filter=src_filter,
                dst_filter=dst_filter,
                dump_commands=dump_commands,
                logger=logger,
                report_path=report,
                list_filtered=list_filtered,
            )
    finally:
        for _p in (downloaded_zip_path, extracted_dir):
            if _p:
                try:
                    pth = Path(_p)
                    if pth.exists():
                        if pth.is_dir():
                            shutil.rmtree(pth)
                        else:
                            pth.unlink()
                except Exception:
                    logger.debug("Failed to cleanup temp path %s", _p)
        for p in (src_tmp, dst_tmp):
            if p:
                try:
                    Path(p).unlink()
                except Exception:
                    logger.debug("Failed to unlink %s", p)
