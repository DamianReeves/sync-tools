Feature: Use SOURCE .gitignore patterns
  Scenario: Import SOURCE .gitignore patterns when enabled
    Given a source directory with files:
      | filename      | content |
      | keep.txt      | keep    |
      | node_modules/a | a      |
    And a source .gitignore with:
      """
      node_modules/
      """
    And an empty destination directory
    When I add extra args:
      | arg               | value |
      | --use-source-gitignore |     |
    And I run sync.sh in one-way mode
    Then the destination directory should contain the files:
      | filename      | content |
      | keep.txt      | keep    |
    And the destination directory should not contain the files:
      | filename       |
      | node_modules/a |
