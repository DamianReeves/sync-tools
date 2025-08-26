Feature: Destination filters
  These scenarios test the behavior of the --ignore-dest option to prevent specific destination files or folders from being updated during one-way sync.

  Scenario: Ignore a specific destination file
    Given a source directory with files:
      | filename   | content |
      | file0.txt  | NEW0    |
      | file1.txt  | NEW1    |
    And a destination directory with files:
      | filename   | content |
      | file0.txt  | OLD0    |
      | file1.txt  | OLD1    |
    When I add extra args:
      | arg            | value     |
      | --ignore-dest  | file0.txt |
    And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename   | content |
      | file0.txt  | OLD0    |
      | file1.txt  | NEW1    |

  Scenario: Ignore a specific destination folder
    Given a source directory with files:
      | filename                  | content |
      | folder1/a.txt             | NEWA    |
      | folder2/b.txt             | NEWB    |
    And a destination directory with files:
      | filename                  | content |
      | folder1/a.txt             | OLDA    |
      | folder2/b.txt             | OLDB    |
    When I add extra args:
      | arg            | value       |
      | --ignore-dest  | folder1/    |
    And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename                  | content |
      | folder1/a.txt             | OLDA    |
      | folder2/b.txt             | NEWB    |  
