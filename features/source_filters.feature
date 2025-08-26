Feature: Source filters
  These scenarios test the behavior of the --ignore-src option to prevent specific source files or folders from being copied during one-way sync.

  Scenario: Ignore a specific source file
    Given a source directory with files:
      | filename   | content |
      | file0.txt  | NEW0    |
      | file1.txt  | NEW1    |
    And an empty destination directory
    When I add extra args:
      | arg           | value      |
      | --ignore-src  | file0.txt  |
    And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename   | content |
      | file1.txt  | NEW1    |
    And the destination directory should not contain the files:
      | filename   |
      | file0.txt  |

  Scenario: Ignore a specific source folder
    Given a source directory with files:
      | filename                  | content |
      | folder1/a.txt             | A_NEW   |
      | folder2/b.txt             | B_NEW   |
    And an empty destination directory
    When I add extra args:
      | arg           | value       |
      | --ignore-src  | folder1/    |
    And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename                  | content |
      | folder2/b.txt             | B_NEW   |
    And the destination directory should not contain the files:
      | filename                  |
      | folder1/a.txt             |
