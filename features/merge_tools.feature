Feature: Interactive Merge Tool Integration
  As a developer
  I want to use familiar merge tools to resolve conflicts
  So that I can handle complex file conflicts effectively

  Background:
    Given I have a source directory with files:
      | path               | content                    | modified             |
      | config/settings.json | {"version": "1.0", "src": true} | 2025-08-30T10:30:00  |
      | src/module.js      | function src() { return 1; } | 2025-08-30T10:35:00  |
    And I have a destination directory with files:
      | path               | content                    | modified             |
      | config/settings.json | {"version": "2.0", "dest": true} | 2025-08-30T10:40:00  |
      | src/module.js      | function dest() { return 2; } | 2025-08-30T10:25:00  |

  Scenario: Execute plan with conflict resolution strategy flag
    Given I have a plan file "conflicts.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   config/settings.json          32B   2025-08-30T10:40:00  [CONFLICT: both-modified, auto:newest]
      <> file   src/module.js                 28B   2025-08-30T10:35:00  [CONFLICT: both-modified, auto:source]
      """
    When I run sync-tools with arguments "sync --apply-plan conflicts.plan"
    Then the command should succeed
    And the source file "config/settings.json" should contain "version": "2.0"
    And the destination file "src/module.js" should contain "function src()"

  Scenario: Interactive merge with vimdiff
    Given I have a plan file "merge.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   config/settings.json          32B   2025-08-30T10:40:00  [CONFLICT: both-modified, merge-tool:vimdiff]
      """
    And the environment variable "EDITOR" is set to "vimdiff"
    When I run sync-tools with arguments "sync --apply-plan merge.plan --interactive-merge"
    Then the command should prompt for merge tool launch
    And the merge tool "vimdiff" should be invoked with source and destination files

  Scenario: Interactive merge with custom merge tool
    Given I have a plan file "custom-merge.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   src/module.js                 28B   2025-08-30T10:35:00  [CONFLICT: both-modified]
      """
    When I run sync-tools with arguments "sync --apply-plan custom-merge.plan --interactive-merge --merge-tool meld"
    Then the command should prompt for merge tool launch
    And the merge tool "meld" should be invoked with source and destination files

  Scenario: Backup creation during conflict resolution
    Given I have a plan file "backup-conflict.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   config/settings.json          32B   2025-08-30T10:40:00  [CONFLICT: both-modified, auto:backup]
      """
    When I run sync-tools with arguments "sync --apply-plan backup-conflict.plan"
    Then the command should succeed
    And a backup file matching pattern "config/settings.json.conflict-*" should exist in destination
    And the conflict should be resolved using newest-wins strategy

  Scenario: Batch conflict resolution
    Given I have a plan file "batch-conflicts.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   config/settings.json          32B   2025-08-30T10:40:00  [CONFLICT: both-modified]
      <> file   src/module.js                 28B   2025-08-30T10:35:00  [CONFLICT: both-modified]
      """
    When I run sync-tools with arguments "sync --apply-plan batch-conflicts.plan --conflict-strategy newest-wins"
    Then the command should succeed
    And all conflicts should be resolved using newest-wins strategy
    And no merge tool prompts should appear

  Scenario: Skip conflicts and generate resolution plan
    Given I have a source file "src/new.js" with content "console.log('new');" 
    And I have a source file "config/settings.json" with content "{\"version\": \"1.0\"}"
    And I have a destination file "config/settings.json" with content "{\"version\": \"2.0\"}"
    And I have a plan file "skip-conflicts.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      << file   src/new.js                    15B   2025-08-30T10:35:00  [new-in-source]
      <> file   config/settings.json          32B   2025-08-30T10:40:00  [CONFLICT: both-modified]
      """
    And all source files and destination files are accessible
    And the plan references only existing files
    When I run sync-tools with arguments "sync --apply-plan skip-conflicts.plan --skip-conflicts --generate-conflict-plan resolve.plan"
    Then the command should succeed
    And the file "src/new.js" should be synced to destination
    And the file "config/settings.json" should remain unchanged
    And a new plan file "resolve.plan" should be created containing only conflicts

  Scenario: Merge tool timeout handling
    Given I have a plan file "timeout.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   config/settings.json          32B   2025-08-30T10:40:00  [CONFLICT: both-modified, merge-tool:slow-tool]
      """
    And a merge tool "slow-tool" that takes longer than timeout
    When I run sync-tools with arguments "sync --apply-plan timeout.plan --interactive-merge --merge-timeout 5"
    Then the command should handle the timeout gracefully
    And fall back to the default conflict strategy

  Scenario: Binary file conflict resolution
    Given I have a source directory with files:
      | path               | content                    | modified             | type   |
      | assets/logo.png    | <binary-content-v1>       | 2025-08-30T10:30:00  | binary |
    And I have a destination directory with files:
      | path               | content                    | modified             | type   |
      | assets/logo.png    | <binary-content-v2>       | 2025-08-30T10:40:00  | binary |
    And I have a plan file "binary-conflict.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   assets/logo.png               12KB  2025-08-30T10:40:00  [CONFLICT: both-modified]
      """
    When I run sync-tools with arguments "sync --apply-plan binary-conflict.plan --interactive-merge"
    Then the command should not invoke a text merge tool
    And should use binary conflict resolution strategy (newest-wins by default)

  Scenario: Three-way merge with git integration
    Given I have a git repository with common ancestor
    And I have a plan file "three-way.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file   src/module.js                 28B   2025-08-30T10:35:00  [CONFLICT: both-modified]
      """
    When I run sync-tools with arguments "sync --apply-plan three-way.plan --interactive-merge --use-git-base"
    Then the command should detect the git repository
    And invoke the merge tool with three-way merge (base, source, dest)
    And provide better conflict resolution context