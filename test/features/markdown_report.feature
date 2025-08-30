Feature: Markdown Report Generation
  As a user of sync-tools
  I want to generate markdown reports of sync operations
  So that I can review and document what changes will be made

  Background:
    Given I have a source directory "source"
    And I have a destination directory "dest"

  Scenario: Generate markdown report for one-way sync
    Given the source directory contains:
      | path          | content     |
      | file1.txt     | Hello World |
      | dir1/file2.txt| Test File   |
    And the destination directory is empty
    When I run sync-tools with arguments "sync --source source --dest dest --report sync_report.md --dry-run"
    Then the command should succeed
    And a file "sync_report.md" should exist
    And the file "sync_report.md" should contain "# Sync Report"
    And the file "sync_report.md" should contain "## Configuration"
    And the file "sync_report.md" should contain "## Summary Statistics"
    And the file "sync_report.md" should contain "Files Created"

  Scenario: Generate markdown report with actual sync
    Given the source directory contains:
      | path          | content     |
      | file1.txt     | Hello World |
    And the destination directory is empty
    When I run sync-tools with arguments "sync --source source --dest dest --report changes.md"
    Then the command should succeed
    And a file "changes.md" should exist
    And the destination directory should contain "file1.txt"
    And the file "changes.md" should contain "# Sync Report"

  Scenario: Generate markdown report for sync with updates
    Given the source directory contains:
      | path          | content          |
      | file1.txt     | Updated Content  |
      | file2.txt     | New File         |
    And the destination directory contains:
      | path          | content          |
      | file1.txt     | Original Content |
      | file3.txt     | To Be Deleted    |
    When I run sync-tools with arguments "sync --source source --dest dest --report update_report.md --dry-run"
    Then the command should succeed
    And a file "update_report.md" should exist
    And the file "update_report.md" should contain "Files to Update"
    And the file "update_report.md" should contain "Files/Directories to Create"
    And the file "update_report.md" should contain "Files/Directories to Delete"

  Scenario: Report shows no changes when directories are identical
    Given the source directory contains:
      | path          | content     |
      | file1.txt     | Same Content|
    And the destination directory contains:
      | path          | content     |
      | file1.txt     | Same Content|
    When I run sync-tools with arguments "sync --source source --dest dest --report no_changes.md --dry-run"
    Then the command should succeed
    And a file "no_changes.md" should exist
    And the file "no_changes.md" should contain "No changes detected"