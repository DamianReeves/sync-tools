from sync_tools.rsync_wrapper import decision_for_path, parse_filter_lines


def test_basename_vs_path_matching():
    # pattern without slash matches basename only
    lines = ['- foo']
    assert decision_for_path('foo', lines) == 'exclude'
    assert decision_for_path('a/foo', lines) == 'exclude'
    assert decision_for_path('a/bar', lines) == 'neutral'

    # pattern with slash matches full path
    lines = ['- /a/b']
    assert decision_for_path('a/b', lines) == 'exclude'
    assert decision_for_path('a/b/c', lines) == 'neutral'


def test_recursive_and_directory_patterns():
    lines = ['- /a/b/**']
    assert decision_for_path('a/b', lines) == 'exclude'
    assert decision_for_path('a/b/c.txt', lines) == 'exclude'
    assert decision_for_path('a/b/c/d.txt', lines) == 'exclude'

    lines = ['- dir/']
    assert decision_for_path('dir/file.txt', lines) == 'exclude'
    assert decision_for_path('dir', lines) == 'exclude'
    assert decision_for_path('otherdir/dirfile', lines) == 'neutral'


def test_wildcards_and_question_mark():
    lines = ['- *.txt']
    assert decision_for_path('readme.txt', lines) == 'exclude'
    assert decision_for_path('sub/readme.txt', lines) == 'exclude'  # basename match

    lines = ['- **/*.txt']
    assert decision_for_path('sub/readme.txt', lines) == 'exclude'
    assert decision_for_path('deep/a/b/c.log', lines) == 'neutral'

    lines = ['- file?.txt']
    assert decision_for_path('file1.txt', lines) == 'exclude'
    assert decision_for_path('file10.txt', lines) == 'neutral'


def test_precedence_and_override():
    # later rules override earlier ones
    lines = ['- *.log', '+ important.log']
    assert decision_for_path('a/important.log', lines) == 'include'
    assert decision_for_path('other.log', lines) == 'exclude'

    # reverse order -> exclude wins
    lines = ['+ *.log', '- important.log']
    assert decision_for_path('a/important.log', lines) == 'exclude'
    assert decision_for_path('b/other.log', lines) == 'include'


def test_parse_filter_lines_various_formats():
    raw = ['+ /a', '+ /a/**', '- node_modules', '- *.py']
    parsed = parse_filter_lines(raw)
    assert parsed[0][0] == '+' and parsed[1][0] == '+'
    assert parsed[2][0] == '-'
    assert parsed[3][1] == '*.py'


def test_complex_mixed_patterns():
    lines = [
        '- **/*.tmp',
        '- /build/',
        '+ important/*.tmp',
        '- secret?.key',
    ]
    # secret1.key excluded
    assert decision_for_path('secret1.key', lines) == 'exclude'
    # tmp under any path excluded except those under important/
    assert decision_for_path('a/b/x.tmp', lines) == 'exclude'
    assert decision_for_path('important/x.tmp', lines) == 'include'
    # build directory excluded
    assert decision_for_path('build/output.o', lines) == 'exclude'
