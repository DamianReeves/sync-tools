Feature: Ignore Patterns
  As a user
  I want to exclude certain files and directories from sync
  So that I can control what gets synchronized

  Scenario: Using .syncignore file
    Given I have a source directory with files
    And I have a .syncignore file in the source directory
    When I run sync-tools with one-way sync
    Then files matching ignore patterns should not be copied
    And files not matching patterns should be copied
    And the exit code should be 0

  Scenario: Using gitignore import
    Given I have a source directory with files
    And I have a .gitignore file in the source directory
    When I run sync-tools with gitignore import enabled
    Then files matching gitignore patterns should not be copied
    And files not matching patterns should be copied
    And the exit code should be 0

  Scenario: Unignore patterns
    Given I have a source directory with files
    And I have ignore patterns with unignore rules
    When I run sync-tools with one-way sync
    Then files matching ignore patterns should not be copied
    And files matching unignore patterns should be copied
    And the exit code should be 0