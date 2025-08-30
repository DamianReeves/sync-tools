Feature: Sync From Subcommand
  As a user of sync-tools
  I want to use "sync from" to sync from a source directory to my current directory
  So that I can quickly sync content into my working directory

  Background:
    Given I have a source directory "source"
    And I have a destination directory "dest"

  Scenario: Basic sync from with dry-run
    Given the source directory contains:
      | path          | content     |
      | file1.txt     | Hello World |
      | file2.txt     | Test File   |
    And the destination directory is empty
    And I am in the destination directory
    When I run sync-tools with arguments "sync from ../source --dry-run"
    Then the command should succeed
    And the destination directory should be empty

  Scenario: Actual sync from operation
    Given the source directory contains:
      | path          | content     |
      | file1.txt     | Hello World |
      | dir1/file2.txt| Nested File |
    And the destination directory is empty
    And I am in the destination directory
    When I run sync-tools with arguments "sync from ../source"
    Then the command should succeed
    And the destination directory should contain "file1.txt"
    And the destination directory should contain "dir1/file2.txt"

  Scenario: Sync from with markdown report
    Given the source directory contains:
      | path          | content     |
      | test.txt      | Content     |
    And the destination directory is empty
    And I am in the destination directory
    When I run sync-tools with arguments "sync from ../source --report sync_report.md --dry-run"
    Then the command should succeed
    And a file "sync_report.md" should exist
    And the file "sync_report.md" should contain "# Sync Report"
    And the file "sync_report.md" should contain "Files Created"

  Scenario: Sync from with filters
    Given the source directory contains:
      | path          | content      |
      | important.txt | Important    |
      | temp.log      | Log content  |
      | .hidden       | Hidden file  |
    And the destination directory is empty
    And I am in the destination directory
    When I run sync-tools with arguments "sync from ../source --only '*.txt' --dry-run"
    Then the command should succeed

  Scenario: Sync from with preview
    Given the source directory contains:
      | path          | content      |
      | preview.txt   | Preview this |
    And the destination directory is empty  
    And I am in the destination directory
    When I run sync-tools with arguments "sync from ../source --preview"
    Then the command should succeed

  Scenario: Error when source directory doesn't exist
    Given I am in the destination directory
    When I run sync-tools with arguments "sync from /nonexistent/path"
    Then the command should fail
    And the output should contain "does not exist"

  Scenario: Error when trying to sync directory to itself
    Given the source directory contains:
      | path          | content     |
      | file.txt      | Test        |
    And I am in the source directory
    When I run sync-tools with arguments "sync from ."
    Then the command should fail
    And the output should contain "cannot sync directory to itself"