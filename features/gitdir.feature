Feature: .git directory default exclusion
  Scenario: .git directory is excluded by default
    Given a source directory with files:
      | filename       | content |
      | .git/config    | cfg     |
      | keep.txt       | keep    |
    And an empty destination directory
  When I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename       | content |
      | keep.txt       | keep    |

  Scenario: re-include .git on source with ignore-src unignore
    Given a source directory with files:
      | filename       | content |
      | .git/config    | cfg     |
      | keep.txt       | keep    |
    And an empty destination directory
    When I add extra args:
      | arg           | value         |
      | --ignore-src  | !/.git/**     |
  And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename       | content |
      | keep.txt       | keep    |
      | .git/config    | cfg     |
