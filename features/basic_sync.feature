Feature: Basic Sync Operations
  As a user
  I want to synchronize files between directories
  So that I can keep my data in sync

  Scenario: One-way sync with dry-run
    Given I have a source directory with files
    And I have an empty destination directory
    When I run sync-tools with one-way sync and dry-run
    Then it should show what files would be copied
    And no files should actually be copied
    And the exit code should be 0

  Scenario: One-way sync execution
    Given I have a source directory with files
    And I have an empty destination directory
    When I run sync-tools with one-way sync
    Then files should be copied to destination
    And the destination should match source
    And the exit code should be 0

  Scenario: Two-way sync with conflict detection
    Given I have a source directory with files
    And I have a destination directory with different files
    When I run sync-tools with two-way sync
    Then files should be synchronized in both directions
    And conflicts should be handled appropriately
    And the exit code should be 0