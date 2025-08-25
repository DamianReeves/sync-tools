import os
from sync_tools.rsync_wrapper import to_filter_lines, build_filter_file


def test_to_filter_lines_basic_include_and_exclude():
    patterns = ["!/.git", "node_modules", "!docs/manual"]
    lines = to_filter_lines(patterns)
    # Expect includes for /.git and its recursive children
    assert "+ /.git" in lines
    assert "+ /.git/**" in lines
    # Exclude should be emitted as-is
    assert "- node_modules" in lines
    # docs/manual parent includes
    assert "+ /docs" in lines
    assert "+ /docs/manual" in lines
    assert "+ /docs/manual/**" in lines


def test_build_filter_file_whitelist(tmp_path):
    only = ["src", "README.md"]
    path, lines = build_filter_file(only, only=True, default_excludes=["/.git/"])
    assert os.path.exists(path)
    # whitelist should end with - *
    assert lines[-1] == "- *"
    # default excludes appended
    assert "- /.git/" in lines
    # cleanup
    os.unlink(path)


def test_to_filter_lines_edge_cases():
    # nested parent includes
    patterns = ["!a/b/c", "!/x/y"]
    lines = to_filter_lines(patterns)
    # parents for a/b/c should include /a, /a/b, /a/b/c
    assert "+ /a" in lines
    assert "+ /a/b" in lines
    assert "+ /a/b/c" in lines
    # leading slash input for /x/y should be normalized but include parents
    assert "+ /x" in lines
    assert "+ /x/y" in lines

    # pattern-only excludes (no slashes)
    patterns2 = ["node_modules", "build"]
    lines2 = to_filter_lines(patterns2)
    assert "- node_modules" in lines2
    assert "- build" in lines2
