Feature: Git Patch Generation
  As a developer
  I want to generate git patch files instead of synchronizing
  So that I can review and apply changes manually with version control

  Scenario: Generate patch file from source differences
    Given I have a source directory with files
    And I have a destination directory with some matching and some different files
    When I run sync-tools with patch generation to "changes.patch"
    Then a git patch file should be created at "changes.patch"
    And the patch file should contain differences between source and destination
    And no files should be synchronized
    And the exit code should be 0

  Scenario: Generate patch with only new files
    Given I have a source directory with files
    And I have an empty destination directory
    When I run sync-tools with patch generation to "new-files.patch"
    Then a git patch file should be created at "new-files.patch" 
    And the patch file should contain all new files from source
    And the patch should show files as new additions
    And the exit code should be 0

  Scenario: Generate patch with deleted files
    Given I have an empty source directory
    And I have a destination directory with files
    When I run sync-tools with patch generation to "deletions.patch"
    Then a git patch file should be created at "deletions.patch"
    And the patch file should contain file deletions
    And the patch should show files as removals
    And the exit code should be 0

  Scenario: Generate patch with ignore patterns
    Given I have a source directory with files
    And I have a destination directory with different files
    And I have a .syncignore file in the source directory
    When I run sync-tools with patch generation to "filtered.patch"
    Then a git patch file should be created at "filtered.patch"
    And the patch file should not contain ignored files
    And the exit code should be 0

  Scenario: Patch generation respects whitelist mode
    Given I have a source directory with files
    And I have a destination directory with different files
    When I run sync-tools with patch generation to "whitelist.patch" and only mode for "hello.txt"
    Then a git patch file should be created at "whitelist.patch"
    And the patch file should only contain changes for whitelisted files
    And the exit code should be 0

  Scenario: Dry-run patch generation shows what would be patched
    Given I have a source directory with files
    And I have a destination directory with different files
    When I run sync-tools with patch generation to "preview.patch" and dry-run
    Then it should show what would be included in the patch
    And no patch file should be created
    And the exit code should be 0