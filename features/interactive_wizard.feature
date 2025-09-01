Feature: Interactive Wizard with Type State Pattern
  As a user
  I want to use an interactive wizard to set up sync operations  
  So that I can easily configure complex sync scenarios without CLI expertise

  Background:
    Given I have a temporary directory for wizard operations

  Scenario: Complete wizard flow with one-way sync
    Given I launch the interactive wizard
    When I am on the welcome screen
    Then I should see sync mode options
    And "One-way sync" should be selected by default
    And "Two-way sync" should be marked as "Coming in future release"
    
    When I press Enter to continue
    Then I should be on the source directory selection screen
    And I should see a directory tree browser
    And I should see the current path display
    
    When I navigate to a test source directory
    And I press Enter to select it
    Then I should be on the destination directory selection screen
    And I should see the selected source path displayed
    And I should see a directory tree browser for destination
    
    When I navigate to a test destination directory
    And I press Enter to select it
    Then I should be on the sync options screen
    And I should see basic sync options
    And I should see advanced sync options
    And "Dry run" should be enabled by default
    
    When I press Enter to continue
    Then I should be on the directory filter selection screen
    And I should see folders from the source directory
    And common exclude folders should be unchecked by default
    
    When I press Enter to continue
    Then I should be on the exclusion patterns screen
    And I should see default exclusion patterns
    And I should be able to add new patterns
    
    When I press Enter to continue
    Then I should be on the preview screen
    And I should see a complete summary of all settings
    And I should see file counts and size estimates
    And I should be able to save the configuration as SyncFile
    
    When I press Enter to start sync
    Then I should be on the progress screen
    And I should see a progress bar
    And I should see current file being processed
    And I should see transfer statistics

  Scenario: Type state safety - cannot access unset data
    Given I launch the interactive wizard
    When I am on the welcome screen
    Then I cannot access source directory path
    And I cannot access destination directory path
    And I cannot access sync configuration
    
    When I select one-way sync mode
    And I continue to source directory selection
    Then I can access the selected sync mode
    But I cannot access source directory path until selected
    And I cannot access destination directory path
    
    When I select a source directory
    And I continue to destination directory selection  
    Then I can access the source directory path
    And I can access the selected sync mode
    But I cannot access destination directory path until selected

  Scenario: Back navigation preserves all state
    Given I launch the interactive wizard
    And I configure sync mode as "one-way"
    And I select source directory "/tmp/test-source"
    And I select destination directory "/tmp/test-dest"
    And I configure sync options with verbose enabled
    And I am on the directory filter selection screen
    
    When I press the back button
    Then I should be on the sync options screen
    And all my sync options should be preserved
    
    When I press the back button again
    Then I should be on the destination directory selection screen
    And the destination path should still be "/tmp/test-dest"
    And the source path should still be displayed
    
    When I press the back button again
    Then I should be on the source directory selection screen  
    And the source path should still be "/tmp/test-source"

  Scenario: Invalid state transitions are prevented
    Given I launch the interactive wizard
    When I am on the welcome screen
    Then I cannot navigate directly to the preview screen
    And I cannot start a sync operation
    
    When I select one-way sync and continue
    And I am on the source directory selection screen
    Then I cannot navigate directly to the destination screen without selecting source
    And I cannot skip to sync options
    
    When I select a source directory and continue
    And I am on the destination directory selection screen
    Then I cannot navigate to preview without selecting destination
    And the sync configuration is not yet accessible

  Scenario: Wizard generates valid SyncFile configuration
    Given I launch the interactive wizard
    And I complete the wizard with the following settings:
      | setting              | value                    |
      | sync_mode           | one-way                  |
      | source_path         | /tmp/wizard-source       |
      | destination_path    | /tmp/wizard-dest         |
      | dry_run             | false                    |
      | verbose             | true                     |
      | use_gitignore       | true                     |
      | selected_folders    | src/, docs/, config/     |
      | exclusion_patterns  | *.log, *.tmp, .DS_Store  |
    
    When I choose to save configuration as SyncFile
    Then a SyncFile should be created with the name "WizardConfig"
    And the SyncFile should contain:
      """
      SYNC /tmp/wizard-source /tmp/wizard-dest
      MODE one-way
      DRYRUN false
      VERBOSE true
      GITIGNORE true
      ONLY src/
      ONLY docs/
      ONLY config/
      EXCLUDE *.log
      EXCLUDE *.tmp
      EXCLUDE .DS_Store
      """

  Scenario: Directory tree browser shows file counts and sizes
    Given I have a test directory structure:
      | path                    | type | size  |
      | /tmp/test-source        | dir  | -     |
      | /tmp/test-source/src    | dir  | -     |
      | /tmp/test-source/src/main.go | file | 1024  |
      | /tmp/test-source/docs   | dir  | -     |
      | /tmp/test-source/docs/readme.md | file | 512 |
      | /tmp/test-source/node_modules | dir | - |
      | /tmp/test-source/node_modules/pkg/index.js | file | 2048 |
    
    And I launch the interactive wizard
    When I navigate to the source directory selection screen
    And I browse to "/tmp/test-source"
    Then I should see folder information:
      | folder       | file_count | size    | selected |
      | src/         | 1          | 1024 B  | true     |
      | docs/        | 1          | 512 B   | true     |
      | node_modules/| 1          | 2048 B  | false    |

  Scenario: Exclusion patterns are validated in real-time
    Given I launch the interactive wizard
    And I navigate to the exclusion patterns screen
    When I try to add the pattern "*.log"
    Then the pattern should be accepted
    And I should see it in the patterns list
    
    When I try to add an invalid pattern "[unclosed"
    Then I should see a validation error
    And the pattern should not be added to the list
    
    When I try to add a duplicate pattern "*.log"
    Then I should see a warning about duplication
    And the pattern should not be added again

  Scenario: Progress screen shows real-time sync status
    Given I launch the interactive wizard
    And I complete the wizard configuration
    And I start the sync operation
    When I am on the progress screen
    Then I should see:
      | element              | status   |
      | progress_bar         | visible  |
      | current_file         | updating |
      | transfer_speed       | visible  |
      | files_completed      | counting |
      | estimated_time       | visible  |
    
    And I should be able to pause the operation
    And I should be able to view detailed logs
    But I should not be able to go back to previous screens

  Scenario: Error handling during wizard flow
    Given I launch the interactive wizard
    When I select a source directory that doesn't exist
    Then I should see an error message
    And I should remain on the source directory selection screen
    
    When I select a destination directory without write permissions
    Then I should see a permission error
    And I should remain on the destination directory selection screen
    
    When I try to sync with no selected folders
    Then I should see a warning about empty selection
    And I should be able to continue or go back to fix it

  Scenario: Wizard respects terminal size constraints
    Given I have a terminal smaller than 80x24
    When I launch the interactive wizard
    Then I should see a minimum size warning
    And I should not be able to use the wizard
    
    Given I have a terminal of exactly 80x24
    When I launch the interactive wizard
    Then the wizard should display correctly
    And all UI elements should be visible
    
    Given I resize the terminal during wizard usage
    When the terminal becomes too small
    Then I should see an adaptive warning
    But I should not lose my current state