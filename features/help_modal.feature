# Language: en
# Help Modal Behavior in Interactive Wizard

Feature: Help Modal Behavior
  As a user navigating the sync wizard
  I want to open and close help modals easily
  So that I can get assistance without being forced to quit

  Background:
    Given I am in the sync wizard
    And I am on the source selection screen

  Scenario: Open and close help modal successfully
    When I press "?" to open help
    Then I should see the help modal displayed
    And I should see "Source Directory Selection:" in the help text
    And I should see navigation instructions in the help text
    When I press any key to close help
    Then the help modal should be closed
    And I should return to the source selection screen
    And I should still be able to navigate directories

  Scenario: Help modal shows context-specific information
    Given I am on the source selection screen
    When I press "?" to open help
    Then I should see "Source Directory Selection:" in the help text
    And I should see "t            Manual path entry mode" in the help text
    And I should see "~            Jump to home directory" in the help text

  Scenario: Help modal can be closed with various keys
    When I press "?" to open help
    Then I should see the help modal displayed
    When I press "h" to close help
    Then the help modal should be closed
    And I should return to the source selection screen

  Scenario: Help modal does not interfere with navigation after closing
    When I press "?" to open help
    And I press "escape" to close help
    Then I should be able to press "j" to navigate down
    And I should be able to press "k" to navigate up
    And I should be able to press "t" to open manual path entry

  Scenario: Multiple help modal open/close cycles work correctly
    When I press "?" to open help
    And I press "enter" to close help
    Then the help modal should be closed
    When I press "?" to open help again
    Then I should see the help modal displayed
    When I press "space" to close help
    Then the help modal should be closed
    And I should return to the source selection screen
