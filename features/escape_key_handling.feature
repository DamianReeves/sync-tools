# Language: en
# Escape Key Handling Across All Modals

Feature: Escape Key Cancellation
  As a user navigating the sync wizard
  I want the escape key to consistently cancel out of all dialogs
  So that I have a reliable way to cancel any modal operation

  Background:
    Given I am in the sync wizard
    And I am on the source selection screen

  Scenario: Escape key cancels help modal
    When I press "?" to open help
    Then I should see the help modal displayed
    When I press "escape" to close help
    Then the help modal should be closed
    And I should return to the source selection screen

  Scenario: Escape key cancels manual path entry dialog
    When I press "t" to open manual path entry
    Then I should see the manual path entry dialog displayed
    When I press "escape" to close the dialog
    Then the manual path entry dialog should be closed
    And I should return to the source selection screen

  Scenario: Escape key cancels home navigation modal
    When I press "~" to open home navigation
    Then I should see the home navigation modal displayed
    When I press "escape" to cancel navigation
    Then the home navigation modal should be closed
    And I should return to the source selection screen

  Scenario: Enter key in home navigation closes modal after selection
    When I press "~" to open home navigation
    Then I should see the home navigation modal displayed
    When I press "enter" to select bookmark
    Then the home navigation modal should be closed
    And the selected path should be applied
