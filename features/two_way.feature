Feature: Two-way sync and conflict preservation
  Scenario: Two-way sync preserves destination changes as conflict copy when both sides differ
    Given a source directory with files:
      | filename      | content  |
      | file1.txt     | A        |
    And a destination directory with files:
      | filename      | content  |
      | file1.txt     | B        |
  When I run sync-tools sync in two-way mode
    Then a conflict file should exist for "file1.txt" on the source
