# Language: en
# Directory Creation in Manual Path Entry

Feature: Directory Creation Prompts
  As a user entering a non-existent path manually
  I want to be prompted to create the directory
  So that I can easily set up new sync destinations

  Background:
    Given I am in the sync wizard
    And I am on the source selection screen

  Scenario: User confirms directory creation
    When I press "t" to open manual path entry
    And I type "/tmp/test-sync-new-dir" in the path field
    And I press "enter" to confirm
    Then I should see a directory creation prompt
    When I press "y" to confirm directory creation
    Then the directory should be created
    And I should proceed to the next step

  Scenario: User declines directory creation and goes back to path entry
    When I press "t" to open manual path entry
    And I type "/tmp/test-sync-nonexistent" in the path field
    And I press "enter" to confirm
    Then I should see a directory creation prompt
    When I press "n" to decline directory creation
    Then I should be back in manual path entry
    And the path field should contain "/tmp/test-sync-nonexistent"

  Scenario: User cancels directory creation entirely
    When I press "t" to open manual path entry
    And I type "/tmp/test-sync-cancel" in the path field
    And I press "enter" to confirm
    Then I should see a directory creation prompt
    When I press "escape" to cancel
    Then I should be back to directory selection
    And no directory should be created

  Scenario: Directory creation fails due to permissions
    When I press "t" to open manual path entry
    And I type "/root/restricted-dir" in the path field
    And I press "enter" to confirm
    Then I should see a directory creation prompt
    When I press "y" to confirm directory creation
    Then I should see an error about directory creation failure
    And I should remain on the current screen
