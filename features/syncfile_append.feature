Feature: SyncFile APPEND Post-Sync Action
  As a DevOps engineer
  I want to append content to files after sync operations using SyncFile
  So that I can add environment-specific configuration and metadata

  Background:
    Given I have a source directory with files:
      | path           | content                           |
      | config/app.yml | app_name: myapp\nport: 8080      |
      | src/main.js    | console.log('hello');            |
    And I have an empty destination directory

  Scenario: Append inline content to single file
    Given I have a SyncFile "AppendTest" containing:
      """
      SYNC ../source ../dest
      APPEND config/app.yml:

        # Production configuration
        production:
          debug: false
          log_level: info
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendTest"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      app_name: myapp
      port: 8080

      # Production configuration
      production:
        debug: false
        log_level: info
      """

  Scenario: Append content from external file
    Given I have a file "footer.yml" containing:
      """

      # Auto-generated footer
      build_time: ${BUILD_TIME}
      version: ${APP_VERSION}
      """
    And I have a SyncFile "AppendFromFile" containing:
      """
      SYNC ../source ../dest
      APPEND --file footer.yml config/app.yml
      """
    When I run sync-tools with arguments "syncfile AppendFromFile"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      app_name: myapp
      port: 8080

      # Auto-generated footer
      build_time: ${BUILD_TIME}
      version: ${APP_VERSION}
      """

  Scenario: Append to multiple files
    Given I have a SyncFile "AppendMultiple" containing:
      """
      SYNC ../source ../dest
      APPEND config/app.yml:

        environment: production
      END APPEND
      APPEND src/main.js:

        module.exports = app;
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendMultiple"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      app_name: myapp
      port: 8080

      environment: production
      """
    And the destination directory should contain "src/main.js" with content:
      """
      console.log('hello');

      module.exports = app;
      """

  Scenario: Append with backup creates backup file
    Given I have a SyncFile "AppendBackup" containing:
      """
      SYNC ../source ../dest
      APPEND --backup config/app.yml:

        backup_test: true
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendBackup"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      app_name: myapp
      port: 8080

      backup_test: true
      """
    And a backup file matching pattern "config/app.yml.backup*" should exist in destination

  Scenario: Append fails when target file doesn't exist
    Given I have a SyncFile "AppendMissingFile" containing:
      """
      SYNC ../source ../dest
      APPEND nonexistent/config.yml:

        test: value
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendMissingFile"
    Then the command should fail
    And the error should contain "target file not found: nonexistent/config.yml"

  Scenario: Multiple APPEND operations in sequence
    Given I have a SyncFile "AppendSequence" containing:
      """
      SYNC ../source ../dest
      APPEND config/app.yml:

        # First append
        stage: development
      END APPEND
      APPEND config/app.yml:

        # Second append
        debug: true
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendSequence"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      app_name: myapp
      port: 8080

      # First append
      stage: development

      # Second append
      debug: true
      """

  Scenario: APPEND with dry-run shows what would be appended
    Given I have a SyncFile "AppendDryRun" containing:
      """
      SYNC ../source ../dest
      APPEND --dry-run config/app.yml:

        dry_run_test: true
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendDryRun --dry-run"
    Then the command should succeed
    And the output should contain "Would append to"
    And the output should contain "config/app.yml"

  Scenario: APPEND without newline option
    Given I have a SyncFile "AppendNoNewline" containing:
      """
      SYNC ../source ../dest
      APPEND --newline=false config/app.yml:
        production: true
      END APPEND
      """
    When I run sync-tools with arguments "syncfile AppendNoNewline"
    Then the command should succeed
    And the destination directory should contain "config/app.yml" with content:
      """
      app_name: myapp
      port: 8080production: true
      """
