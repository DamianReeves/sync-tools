Feature: Plan Generation and Execution
  As a sync-tools user
  I want to generate and execute sync plans
  So that I can review and control sync operations

  Background:
    Given I have a source directory with files:
      | path               | content              | modified             |
      | file1.txt          | source content 1    | 2025-08-30T10:30:00  |
      | file2.txt          | source content 2    | 2025-08-30T10:35:00  |
      | shared.txt         | source shared       | 2025-08-30T10:40:00  |
    And I have a destination directory with files:
      | path               | content              | modified             |
      | shared.txt         | dest shared         | 2025-08-30T10:20:00  |
      | dest-only.txt      | dest only content   | 2025-08-30T10:25:00  |

  Scenario: Generate plan with all changes
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan all.plan"
    Then the command should succeed
    And a file "all.plan" should be created
    And the plan file should contain:
      """
      << file file1.txt
      << file file2.txt
      <> file shared.txt
      >> file dest-only.txt
      """

  Scenario: Filter plan to show only new files in source
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan new.plan --include-changes new-in-source"
    Then the command should succeed
    And the plan file "new.plan" should contain:
      """
      << file file1.txt
      << file file2.txt
      """
    And the plan file should not contain "shared.txt"
    And the plan file should not contain "dest-only.txt"

  Scenario: Filter plan to show only conflicts
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan conflicts.plan --include-changes conflicts"
    Then the command should succeed
    And the plan file "conflicts.plan" should contain:
      """
      <> file shared.txt
      """
    And the plan file should not contain "file1.txt"
    And the plan file should not contain "file2.txt"
    And the plan file should not contain "dest-only.txt"

  Scenario: Execute plan with source-to-dest operations
    Given I have a plan file "s2d.plan" containing:
      """
      # Source: ./source
      # Destination: ./dest
      # Mode: one-way
      
      << file file1.txt 17B 2025-08-30T10:30:00 [new-in-source]
      << file file2.txt 17B 2025-08-30T10:35:00 [new-in-source]
      """
    When I run sync-tools with arguments "sync --apply-plan s2d.plan"
    Then the command should succeed
    And the destination directory should contain "file1.txt" with content "source content 1"
    And the destination directory should contain "file2.txt" with content "source content 2"

  Scenario: Execute plan with dest-to-source operations
    Given I have a plan file "d2s.plan" containing:
      """
      # Source: ./source
      # Destination: ./dest
      # Mode: one-way
      
      >> file dest-only.txt 18B 2025-08-30T10:25:00 [new-in-dest]
      """
    When I run sync-tools with arguments "sync --apply-plan d2s.plan"
    Then the command should succeed
    And the source directory should contain "dest-only.txt" with content "dest only content"

  Scenario: Execute plan with bidirectional conflict resolution using newest-wins
    Given I have a source file "conflict.txt" with content "source version" modified at "2025-08-30T10:40:00"
    And I have a destination file "conflict.txt" with content "dest version" modified at "2025-08-30T10:50:00"
    And I have a plan file "bidir.plan" containing:
      """
      # Source: ./source
      # Destination: ./dest
      # Mode: two-way
      
      <> file conflict.txt 14B 2025-08-30T10:50:00 [CONFLICT: both-modified]
      """
    When I run sync-tools with arguments "sync --apply-plan bidir.plan"
    Then the command should succeed
    And both source and destination should contain "conflict.txt" with content "dest version"

  Scenario: Interactive editor opens when using --interactive with --plan
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan edit.plan --interactive --editor echo"
    Then the command should succeed
    And the editor "echo" should have been invoked with the plan file
    And the plan file "edit.plan" should exist

  Scenario: Custom editor specified with --editor flag
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan custom.plan --interactive --editor nano"
    Then the command should attempt to open "nano" editor
    And the plan file "custom.plan" should exist

  Scenario: Plan validation detects invalid operations
    Given I have a plan file "invalid.plan" containing:
      """
      # Source: ./source
      # Destination: ./dest
      
      invalid-op file test.txt
      << file
      """
    When I run sync-tools with arguments "sync --apply-plan invalid.plan"
    Then the command should fail
    And the error should contain "invalid plan syntax"

  Scenario: Exclude unchanged files from plan
    Given I have identical files in source and destination:
      | path               | content              | modified             |
      | identical.txt      | same content        | 2025-08-30T10:00:00  |
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan filtered.plan --exclude-changes unchanged"
    Then the command should succeed
    And the plan file should not contain "identical.txt"

  Scenario: Multiple change type filters
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan multi.plan --include-changes new-in-source,conflicts"
    Then the command should succeed
    And the plan file "multi.plan" should contain:
      """
      << file file1.txt
      << file file2.txt
      <> file shared.txt
      """
    And the plan file should not contain "dest-only.txt"

  Scenario: Conflict resolution with backup strategy
    Given I have a plan file "backup.plan" containing:
      """
      # Source: ./source
      # Destination: ./dest
      # Mode: two-way
      
      <> file shared.txt 12B 2025-08-30T10:40:00 [CONFLICT: both-modified, auto:backup]
      """
    When I run sync-tools with arguments "sync --apply-plan backup.plan"
    Then the command should succeed
    And a backup file matching "shared.txt.conflict-*" should exist in destination
    And the conflict should be resolved

  Scenario: Plan shows accurate summary statistics
    When I run sync-tools with arguments "sync --source ./source --dest ./dest --plan stats.plan"
    Then the command should succeed
    And the plan file "stats.plan" should contain:
      """
      # Summary:
      # Files matching filter: 4
      # New in source: 2
      # New in dest: 1
      # Updates: 0
      # Conflicts: 1
      """

  Scenario: Dry-run mode for plan execution
    Given I have a plan file "dry.plan" containing:
      """
      # Source: ./source
      # Destination: ./dest
      
      << file file1.txt 17B 2025-08-30T10:30:00 [new-in-source]
      """
    When I run sync-tools with arguments "sync --apply-plan dry.plan --dry-run"
    Then the command should succeed
    And the destination directory should not contain "file1.txt"
    And the output should indicate dry-run mode