Feature: Two-Phased Interactive Sync
  As a DevOps engineer
  I want to review and modify sync operations before execution
  So that I can have fine-grained control over file synchronization

  Background:
    Given I have a source directory with files:
      | path               | content              | modified             |
      | config/app.yml     | app_config_v1       | 2025-08-30T10:30:00  |
      | src/main.js        | console.log('v1')   | 2025-08-30T10:35:00  |
      | docs/README.md     | # Version 1         | 2025-08-30T09:00:00  |
    And I have a destination directory with files:
      | path               | content              | modified             |
      | config/app.yml     | app_config_v0       | 2025-08-30T09:30:00  |
      | config/db.yml      | database_config     | 2025-08-30T10:00:00  |
      | docs/README.md     | # Version 2         | 2025-08-30T10:00:00  |

  Scenario: Generate basic sync plan
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan sync.plan"
    Then the command should succeed
    And a file "sync.plan" should be created
    And the plan file should contain:
      """
      # Sync Plan Generated:
      # Generated from: sync-tools sync --source
      # Mode: one-way
      """
    And the plan file should contain sync operations with visual aliases

  Scenario: Plan shows different change types with visual aliases
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan changes.plan"
    Then the command should succeed
    And the plan file "changes.plan" should contain:
      """
      << file src/main.js
      <> file config/app.yml
      >> file config/db.yml
      <> file docs/README.md
      """

  Scenario: Filter plan by change types - new files only
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan new-only.plan --include-changes new-in-source"
    Then the command should succeed
    And the plan file "new-only.plan" should contain:
      """
      << file src/main.js
      """
    And the plan file should not contain "config/app.yml"
    And the plan file should not contain "config/db.yml"
    And the plan file should not contain "docs/README.md"

  Scenario: Filter plan by change types - conflicts only
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan conflicts.plan --include-changes conflicts"
    Then the command should succeed
    And the plan file "conflicts.plan" should contain:
      """
      <> file config/app.yml
      <> file docs/README.md
      """
    And the plan file should not contain "src/main.js"
    And the plan file should not contain "config/db.yml"

  Scenario: Filter plan by change types - updates and conflicts
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan updates.plan --include-changes updates,conflicts"
    Then the command should succeed
    And the plan file "updates.plan" should contain:
      """
      <> file config/app.yml
      <> file docs/README.md
      """

  Scenario: Execute a simple sync plan with visual aliases
    Given I have a plan file "execute.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: one-way
      
      << file   src/main.js                    17B   2025-08-30T10:35:00  [new-in-source]
      << file   config/app.yml                13B   2025-08-30T10:30:00  [update: newer-in-source]
      # >> file   config/db.yml                15B   2025-08-30T10:00:00  [new-in-dest]
      """
    When I run sync-tools with arguments "sync --apply-plan execute.plan"
    Then the command should succeed
    And the destination directory should contain "src/main.js" with content "console.log('v1')"
    And the destination directory should contain "config/app.yml" with content "app_config_v1"
    And the destination directory should contain "config/db.yml"

  Scenario: Execute plan with bidirectional operation
    Given I have a plan file "bidirectional.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      << file   src/main.js                    17B   2025-08-30T10:35:00  [new-in-source]
      >> file   config/db.yml                 15B   2025-08-30T10:00:00  [new-in-dest]
      """
    When I run sync-tools with arguments "sync --apply-plan bidirectional.plan"
    Then the command should succeed
    And the source directory should contain "config/db.yml" with content "database_config"
    And the destination directory should contain "src/main.js" with content "console.log('v1')"

  Scenario: Skip commented lines in plan execution
    Given I have a plan file "with-skips.plan" containing:
      """
      # Sync Plan Generated: 2025-08-30T10:45:00
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: one-way
      
      << file   src/main.js                    17B   2025-08-30T10:35:00  [new-in-source]
      # << file   config/app.yml              13B   2025-08-30T10:30:00  [update: newer-in-source]
      """
    When I run sync-tools with arguments "sync --apply-plan with-skips.plan"
    Then the command should succeed
    And the destination directory should contain "src/main.js" with content "console.log('v1')"
    And the destination file "config/app.yml" should contain "app_config_v0"

  Scenario: Plan generation with SyncFile integration
    Given I have a SyncFile "TestSyncFile" containing:
      """
      SYNC ./test_source ./test_dest
      MODE one-way
      EXCLUDE *.tmp
      GITIGNORE true
      """
    When I run sync-tools with arguments "syncfile TestSyncFile --plan syncfile.plan"
    Then the command should succeed
    And the plan file "syncfile.plan" should contain:
      """
      # Generated from: sync-tools syncfile TestSyncFile --plan syncfile.plan
      # SyncFile: TestSyncFile
      """

  Scenario: Plan generation excludes unchanged files by default
    Given I have a source directory with files:
      | path               | content              | modified             |
      | same.txt           | identical content   | 2025-08-30T10:00:00  |
      | different.txt      | source version      | 2025-08-30T10:30:00  |
    And I have a destination directory with files:
      | path               | content              | modified             |
      | same.txt           | identical content   | 2025-08-30T10:00:00  |
      | different.txt      | dest version        | 2025-08-30T10:20:00  |
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan filtered.plan --exclude-changes unchanged"
    Then the command should succeed
    And the plan file should contain "different.txt"
    And the plan file should not contain "same.txt"

  Scenario: Plan validation catches invalid syntax
    Given I have a plan file "invalid.plan" containing:
      """
      # Invalid plan file
      invalid-command file test.txt
      << 
      >> file missing-size
      """
    When I run sync-tools with arguments "sync --apply-plan invalid.plan"
    Then the command should fail
    And the error should contain "invalid plan syntax"

  Scenario: Plan shows summary statistics
    When I run sync-tools with arguments "sync --source ./test_source --dest ./test_dest --plan summary.plan"
    Then the command should succeed
    And the plan file "summary.plan" should contain:
      """
      # Summary:
      # Files matching filter: 4
      # New in source: 1
      # New in dest: 1
      # Updates: 0
      # Conflicts: 2
      """