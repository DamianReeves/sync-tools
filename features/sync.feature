Feature: One-way sync functionality
  As a user
  I want to synchronize files from source to destination in one-way mode
  So that changes in source are reflected in destination

  Scenario: Basic one-way sync copies all files
    Given a source directory with files:
      | filename      | content  |
      | file1.txt     | hello    |
      | dir/file2.txt | world    |
    And an empty destination directory
  When I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename      | content  |
      | file1.txt     | hello    |
      | dir/file2.txt | world    |
