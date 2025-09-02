# Language: en
# Welcome Screen Navigation

Feature: Welcome Screen Navigation
  As a user starting the sync wizard
  I want to navigate past the welcome screen by pressing Enter
  So that I can proceed to select a source directory

  Scenario: Navigate from welcome screen to source selection
    Given I am in the sync wizard
    And I am on the welcome screen
    When I press "enter" to start
    Then I should be on the source selection screen
    And I should see directory browser with navigation instructions
