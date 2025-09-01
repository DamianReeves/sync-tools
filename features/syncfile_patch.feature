Feature: SyncFile PATCH Post-Sync Actions
  As a DevOps engineer
  I want to apply patch files to synchronized files after sync operations
  So that I can automate file modifications with version-controlled patches

  Background:
    Given I have a temporary directory for sync operations

  Scenario: Apply patch file to synchronized file
    Given I have a source directory with files:
      | path      | content                    |
      | config.yml | database_host: localhost   |
      | app.log   | INFO: Application started  |
    And I have an empty destination directory
    And I have a patch file "config.patch" containing:
      """
      --- config.yml
      +++ config.yml
      @@ -1 +1,2 @@
       database_host: localhost
      +database_port: 5432
      """
    And I have a SyncFile "TestSyncFile" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH config.patch config.yml
      """
    When I run sync-tools with SyncFile "TestSyncFile"
    Then the sync should succeed
    And the destination directory should contain "config.yml" with content:
      """
      database_host: localhost
      database_port: 5432
      """
    And the destination directory should contain "app.log" with content "INFO: Application started"

  Scenario: Apply multiple patches to different files
    Given I have a source directory with files:
      | path          | content                  |
      | server.conf   | port=8080               |
      | client.conf   | timeout=30              |
    And I have an empty destination directory  
    And I have a patch file "server.patch" containing:
      """
      --- server.conf
      +++ server.conf  
      @@ -1 +1,2 @@
       port=8080
      +ssl=enabled
      """
    And I have a patch file "client.patch" containing:
      """
      --- client.conf
      +++ client.conf
      @@ -1 +1,2 @@
       timeout=30
      +retries=3
      """
    And I have a SyncFile "MultiPatchSync" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH server.patch server.conf
      PATCH client.patch client.conf
      """
    When I run sync-tools with SyncFile "MultiPatchSync"
    Then the sync should succeed
    And the destination directory should contain "server.conf" with content:
      """
      port=8080
      ssl=enabled
      """
    And the destination directory should contain "client.conf" with content:
      """
      timeout=30
      retries=3
      """

  Scenario: Patch fails when target file doesn't exist
    Given I have a source directory with files:
      | path      | content                    |
      | config.yml | database_host: localhost   |
    And I have an empty destination directory
    And I have a patch file "missing.patch" containing:
      """
      --- nonexistent.txt
      +++ nonexistent.txt
      @@ -1 +1,2 @@
       some content
      +additional line
      """
    And I have a SyncFile "FailingPatch" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH missing.patch nonexistent.txt
      """
    When I run sync-tools with SyncFile "FailingPatch"
    Then the sync should fail
    And the error should contain "target file not found"

  Scenario: Patch with dry-run flag shows what would be applied
    Given I have a source directory with files:
      | path      | content                    |
      | config.yml | database_host: localhost   |
    And I have an empty destination directory
    And I have a patch file "config.patch" containing:
      """
      --- config.yml
      +++ config.yml
      @@ -1 +1,2 @@
       database_host: localhost
      +database_port: 5432
      """
    And I have a SyncFile "DryRunPatch" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH --dry-run config.patch config.yml
      """
    When I run sync-tools with SyncFile "DryRunPatch"
    Then the sync should succeed
    And the output should contain "Would apply patch config.patch to config.yml"
    And the destination directory should contain "config.yml" with content "database_host: localhost"

  Scenario: Patch with backup flag creates backup before applying
    Given I have a source directory with files:
      | path      | content                    |
      | config.yml | database_host: localhost   |
    And I have an empty destination directory
    And I have a patch file "config.patch" containing:
      """
      --- config.yml
      +++ config.yml
      @@ -1 +1,2 @@
       database_host: localhost
      +database_port: 5432
      """
    And I have a SyncFile "BackupPatch" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH --backup config.patch config.yml
      """
    When I run sync-tools with SyncFile "BackupPatch"
    Then the sync should succeed
    And the destination directory should contain "config.yml" with content:
      """
      database_host: localhost
      database_port: 5432
      """
    And the destination directory should contain "config.yml.bak" with content "database_host: localhost"

  Scenario: Patch respects source/destination directory context
    Given I have a source directory with files:
      | path          | content                  |
      | config.yml    | host: localhost         |
    And I have a destination directory with files:
      | path          | content                  |
      | config.yml    | host: localhost         |
    And I have a patch file "update.patch" containing:
      """
      --- config.yml
      +++ config.yml
      @@ -1 +1,2 @@
       host: localhost
      +port: 5432
      """
    And I have a SyncFile "ContextPatch" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH update.patch config.yml
      """
    When I run sync-tools with SyncFile "ContextPatch"
    Then the sync should succeed
    And the destination directory should contain "config.yml" with content:
      """
      host: localhost
      port: 5432
      """

  Scenario: Invalid patch file format fails gracefully
    Given I have a source directory with files:
      | path      | content                    |
      | config.yml | database_host: localhost   |
    And I have an empty destination directory
    And I have a patch file "invalid.patch" containing:
      """
      This is not a valid patch file
      Just some random text
      """
    And I have a SyncFile "InvalidPatch" containing:
      """
      SYNC ./test_source ./test_dest
      PATCH invalid.patch config.yml
      """
    When I run sync-tools with SyncFile "InvalidPatch"
    Then the sync should fail
    And the error should contain "invalid patch format"