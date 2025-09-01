Feature: SyncFile PREPEND Post-Sync Action
  As a DevOps engineer
  I want to prepend content to files after sync operations using SyncFile
  So that I can add headers, licenses, and metadata to the beginning of files

  Background:
    Given I have a source directory with files:
      | path           | content                           |
      | config/app.yml | app_name: myapp\nport: 8080      |
      | src/main.js    | console.log('hello');            |
    And I have an empty destination directory

  Scenario: Prepend inline content to single file
    Given I have a SyncFile "PrependTest" containing:
      """
      SYNC ../source ../dest
      PREPEND config/app.yml:
        # Configuration Header
        # Generated: 2025-08-31
        # Version: v1.0
        
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependTest"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      # Configuration Header
      # Generated: 2025-08-31
      # Version: v1.0
      
      app_name: myapp
      port: 8080
      """

  Scenario: Prepend content from external file
    Given I have a file "header.yml" containing:
      """
      # Auto-generated header
      # Build: ${BUILD_ID}
      # Deployed: ${DEPLOY_TIME}
      
      """
    And I have a SyncFile "PrependFromFile" containing:
      """
      SYNC ../source ../dest
      PREPEND --file header.yml config/app.yml
      """
    When I run sync-tools with arguments "syncfile PrependFromFile"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      # Auto-generated header
      # Build: ${BUILD_ID}
      # Deployed: ${DEPLOY_TIME}
      
      app_name: myapp
      port: 8080
      """

  Scenario: Prepend to multiple files
    Given I have a SyncFile "PrependMultiple" containing:
      """
      SYNC ../source ../dest
      PREPEND config/app.yml:
        # App Configuration
        
      END PREPEND
      PREPEND src/main.js:
        // Main Application Entry Point
        
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependMultiple"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      # App Configuration
      
      app_name: myapp
      port: 8080
      """
    And the destination directory should contain "src/main.js" with content:
      """
      // Main Application Entry Point
      
      console.log('hello');
      """

  Scenario: Prepend with backup creates backup file
    Given I have a SyncFile "PrependBackup" containing:
      """
      SYNC ../source ../dest
      PREPEND --backup config/app.yml:
        # Backed Up Configuration
        
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependBackup"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      # Backed Up Configuration
      
      app_name: myapp
      port: 8080
      """
    And a backup file matching pattern "config/app.yml.backup*" should exist in destination

  Scenario: Prepend fails when target file doesn't exist
    Given I have a SyncFile "PrependMissingFile" containing:
      """
      SYNC ../source ../dest
      PREPEND nonexistent/config.yml:
        # Header for missing file
        
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependMissingFile"
    Then the command should fail
    And the error should contain "target file not found: nonexistent/config.yml"

  Scenario: Multiple PREPEND operations in sequence
    Given I have a SyncFile "PrependSequence" containing:
      """
      SYNC ../source ../dest
      PREPEND config/app.yml:
        # First header
        
      END PREPEND
      PREPEND config/app.yml:
        # Second header
        
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependSequence"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      # Second header
      
      # First header
      
      app_name: myapp
      port: 8080
      """

  Scenario: PREPEND with dry-run shows what would be prepended
    Given I have a SyncFile "PrependDryRun" containing:
      """
      SYNC ../source ../dest
      PREPEND --dry-run config/app.yml:
        # Dry run header
        
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependDryRun --dry-run"
    Then the command should succeed
    And the output should contain "Would prepend to"
    And the output should contain "config/app.yml"

  Scenario: PREPEND without newline option
    Given I have a SyncFile "PrependNoNewline" containing:
      """
      SYNC ../source ../dest
      PREPEND --newline=false config/app.yml:
        # Header
      END PREPEND
      """
    When I run sync-tools with arguments "syncfile PrependNoNewline"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      # Headerapp_name: myapp
      port: 8080
      """