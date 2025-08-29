Feature: Hello World
  As a developer
  I want to verify that the test framework works
  So that I can build more complex tests

  Scenario: Basic test framework validation
    Given the sync-tools binary exists
    When I run sync-tools with help
    Then it should display help information
    And the exit code should be 0