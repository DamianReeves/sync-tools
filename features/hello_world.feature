Feature: Hello World Test
  As a test developer
  I want to verify that the test framework is working
  So that I can proceed with confidence to write real tests

  Scenario: Simple hello world test
    Given I have a test environment
    When I run a simple test
    Then I should see that it passes

  Scenario: Basic math verification
    Given I have the number 5
    And I have the number 3
    When I add them together
    Then the result should be 8
