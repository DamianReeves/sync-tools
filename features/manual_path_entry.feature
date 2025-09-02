# Language: en
# Manual Path Entry Dialog Behavior

Feature: Manual Path Entry Dialog
  As a user navigating the sync wizard
  I want to manually enter file paths using a dialog
  So that I can specify exact paths without navigating through directories

  Background:
    Given I am in the sync wizard
    And I am on the source selection screen

  Scenario: Open and close manual path entry dialog successfully
    When I press "t" to open manual path entry
    Then I should see the manual path entry dialog displayed
    And I should see a path input field
    When I press "escape" to close the dialog
    Then the manual path entry dialog should be closed
    And I should return to the source selection screen

  Scenario: Enter a valid path and confirm
    When I press "t" to open manual path entry
    And I type "/home/user/documents" in the path field
    And I press "enter" to confirm
    Then the manual path entry dialog should be closed
    And the current path should be updated to "/home/user/documents"

  Scenario: Handle invalid path entry
    When I press "t" to open manual path entry
    And I type "/invalid/nonexistent/path" in the path field
    And I press "enter" to confirm
    Then I should see an error message about invalid path
    And the dialog should remain open

  Scenario: Handle long path display
    When I press "t" to open manual path entry
    And I type a very long path
    Then the path display should be properly formatted
    And should not span multiple lines
    And should show the end of the path clearly
