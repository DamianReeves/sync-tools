Feature: Git source sync with report
  These scenarios use a git repository as the source and sync only selected folders, producing a report.

  Scenario: Sync only docs folder from git repo and generate report
    Given a git repository with files:
      | filename         | content    |
      | docs/readme.md   | HELLO DOC  |
      | src/main.py      | print('hi')|
    And an empty destination directory
    When I add extra args:
      | arg        | value        |
      | --only     | docs/        |
      | --report   | report.md    |
    And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename       | content    |
      | docs/readme.md | HELLO DOC  |
    And the report file should contain:
      | line                        |
    | - docs/readme.md           |
