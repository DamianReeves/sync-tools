import click
import subprocess
import tempfile
import urllib.request
import shutil
import os
import shlex
import logging
import json
from pathlib import Path
from typing import List

try:
    import tomllib as _toml
except Exception:
    import tomli as _toml  # type: ignore

from .rsync_wrapper import build_filter_file, run_rsync
from .config import load_toml_path, validate_config


def _configure_logging(level: int, log_file: str | None, log_format: str = "text") -> logging.Logger:
    logger = logging.getLogger("sync_tools")
    logger.setLevel(level)
    # if handlers already exist (tests) avoid duplicate handlers
    if not logger.handlers:
        if log_format == "json":
            try:
                import json_log_formatter
                formatter = json_log_formatter.JSONFormatter()
            except Exception:
                # fallback to simple JSON-ish formatter
                class SimpleJSONFormatter(logging.Formatter):
                    def format(self, record):
                        d = {"time": self.formatTime(record), "level": record.levelname, "name": record.name, "msg": record.getMessage()}
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
def sync(config, source, dest, mode, dry_run, use_source_gitignore, exclude_hidden_dirs, only_syncignore, ignore_src, ignore_dest, only_items, v, log_level, log_file, dump_commands, log_format, report, list_filtered):
    """Perform a sync between SOURCE and DEST using rsync with layered filters.

    You can specify defaults in a TOML config and override them on the command line.
    """
    cfg = {}

    # Support passing a GitHub download / archive URL (zip) as the SOURCE.
    # If a URL is provided we download and unpack it into a temporary
    # directory and point `src` at the extracted tree so later discovery
    # (config, .syncignore, etc) works as if the user had supplied a path.
    downloaded_zip_path = None
    extracted_dir = None
    clone_dir = None
    # Heuristic to detect git repository URLs (ssh/git protocol, .git suffix, or common hosts)
    def _looks_like_git_url(u: str) -> bool:
        if not u:
            return False
        return u.startswith("git@") or u.startswith("git://") or u.endswith(".git") or ("github.com" in u and ".zip" not in u)
    # If source looks like a git repo, perform a shallow clone and use that as source
    if source and _looks_like_git_url(source):
        try:
            clone_dir = tempfile.mkdtemp(prefix="sync_tools_git_clone_")
            # shallow clone; allow upstream to fail and surface a Click error
            res = subprocess.run(["git", "clone", "--depth", "1", source, clone_dir], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            if getattr(res, "returncode", 1) != 0:
                raise click.BadParameter(f"Failed to clone git repo {source}: {getattr(res, 'stderr', '') or getattr(res, 'stdout', '')}")
            source = str(Path(clone_dir).resolve())
        except click.BadParameter:
            raise
        except Exception as e:
            raise click.BadParameter(f"Failed to clone git repo {source}: {e}")
    if source and (source.startswith("http://") or source.startswith("https://")):
        # Only handle obvious zip/archive URLs; treat other URLs as errors
        try:
            # download to a temp file
            tf = tempfile.NamedTemporaryFile(delete=False, suffix=".zip")
            tf.close()
            with urllib.request.urlopen(source) as resp, open(tf.name, "wb") as out:
                shutil.copyfileobj(resp, out)
            downloaded_zip_path = tf.name
            # extract to a temp dir
            extracted_dir = tempfile.mkdtemp(prefix="sync_tools_zip_")
            shutil.unpack_archive(downloaded_zip_path, extracted_dir)
            # If the archive created a single top-level dir (common for GitHub), use it
            parts = [p for p in Path(extracted_dir).iterdir() if p.exists()]
            if len(parts) == 1 and parts[0].is_dir():
                source = str(parts[0].resolve())
            else:
                source = str(Path(extracted_dir).resolve())
        except Exception as e:
            # If download or extraction fails, raise a clear Click error
            raise click.BadParameter(f"Failed to download/extract source URL: {e}")

    # If the user provided a config file use it. Otherwise we'll try to
    # auto-discover a config file in the source directory (if provided).
    if config:
        with open(config, "rb") as f:
            cfg = _toml.load(f)

    def _cfg(key, default=None):
        return cfg.get(key, default)

    # Merge CLI args with config (CLI takes precedence when provided)
    src = str(Path(source).resolve()) if source else None
    dst = str(Path(dest).resolve()) if dest else None

    # Auto-discover config in source dir if none supplied and source is known
    if not cfg and src:
        cand1 = Path(src) / "sync.toml"
        cand2 = Path(src) / ".sync.toml"
        for c in (cand1, cand2):
            if c.exists():
                cfg = load_toml_path(c)
                break

    # If still no config, attempt to discover in current working directory
    if not cfg:
        cand_cwd1 = Path.cwd() / "sync.toml"
        cand_cwd2 = Path.cwd() / ".sync.toml"
        for c in (cand_cwd1, cand_cwd2):
            if c.exists():
                cfg = load_toml_path(c)
                break

    # validate config shape if present
    if cfg:
        try:
            validate_config(cfg)
        except ValueError as e:
            raise click.BadParameter(f"Invalid config file: {e}")

    # If after discovery the caller still didn't provide src/dst via CLI,
    # pull them from the config file
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

    # Logging level: if user passed --log-level use it; otherwise compute from -v count
    if log_level:
        level = getattr(logging, log_level)
    else:
        # default INFO (20); each -v lowers by 10 until DEBUG (10)
        level = max(10, 20 - (10 * v))

    logger = _configure_logging(level, log_file, log_format)
    logger.debug("CLI options after merge: %s", {"src": src, "dst": dst, "mode": mode, "dry_run": dry_run})

    # construct rsync opts
    rsync_opts = ["-a", "--human-readable", "--itemize-changes", "--partial"]
    if dry_run:
        rsync_opts.append("--dry-run")

    # default excludes such as .git
    default_excludes = ["/.git/"]
    if exclude_hidden_dirs:
        default_excludes.append(".*/")

    src_filter = None
    dst_filter = None
    src_tmp = None
    dst_tmp = None
    try:
        # only_list implies whitelist-only mode for source
        if only_list:
            src_tmp, src_lines = build_filter_file(only_list, only=True, default_excludes=default_excludes)
            src_filter = (src_tmp, src_lines)
        elif ignore_src_list:
            src_tmp, src_lines = build_filter_file(ignore_src_list, only=False, default_excludes=default_excludes)
            src_filter = (src_tmp, src_lines)

        if ignore_dest_list:
            dst_tmp, dst_lines = build_filter_file(ignore_dest_list, only=False, default_excludes=default_excludes)
            dst_filter = (dst_tmp, dst_lines)

        # wire logger into run_rsync so it can log prepared commands
        run_rsync(src, dst, rsync_opts, src_filter=src_filter, dst_filter=dst_filter, dump_commands=dump_commands, logger=logger, report_path=report, list_filtered=list_filtered)
    finally:
        # cleanup downloaded/extracted temp paths if used
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
