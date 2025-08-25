"""TOML config loading and validation for sync-tools."""
from pathlib import Path
from typing import Any, Dict, List


def load_toml_path(path: Path) -> Dict[str, Any]:
    try:
        with path.open("rb") as f:
            try:
                import tomllib as _toml
            except Exception:
                import tomli as _toml  # type: ignore
            return _toml.load(f)
    except FileNotFoundError:
        return {}


def validate_config(cfg: Dict[str, Any]) -> None:
    """Validate the TOML configuration dict.

    Raises ValueError with a helpful message when validation fails.
    Supported keys and expected types:
      - source: str
      - dest: str
      - mode: 'one-way'|'two-way'
      - dry_run: bool
      - use_source_gitignore: bool
      - exclude_hidden_dirs: bool
      - ignore_src: array of strings
      - ignore_dest: array of strings
      - only: array of strings
    """
    if not isinstance(cfg, dict):
        raise ValueError("config must be a table/object in TOML")

    def expect_type(key: str, typ, choices: List[str] | None = None):
        if key in cfg:
            val = cfg[key]
            if not isinstance(val, typ):
                raise ValueError(f"config key '{key}' must be of type {typ.__name__}")
            if choices and val not in choices:
                raise ValueError(f"config key '{key}' must be one of {choices}")

    expect_type("source", str)
    expect_type("dest", str)
    expect_type("mode", str, choices=["one-way", "two-way"]) if "mode" in cfg else None
    expect_type("dry_run", bool)
    expect_type("use_source_gitignore", bool)
    expect_type("exclude_hidden_dirs", bool)

    for arr_key in ("ignore_src", "ignore_dest", "only"):
        if arr_key in cfg:
            val = cfg[arr_key]
            if not isinstance(val, list):
                raise ValueError(f"config key '{arr_key}' must be an array/list of strings")
            for i, itm in enumerate(val):
                if not isinstance(itm, str):
                    raise ValueError(f"items in '{arr_key}' must be strings (index {i})")
