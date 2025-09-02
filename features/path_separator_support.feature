# Language: en
# Path Separator Support in Manual Path Entry

Feature: Path Separator Support
  As a user entering paths manually
  I want to be able to type paths with forward slashes
  So that I can specify complete directory paths like /home/user/documents

  Background:
    Given I am in the sync wizard
    And I am on the source selection screen

  Scenario: Enter absolute path with forward slashes
    When I press "t" to open manual path entry
    And I type "/home/user/documents" in the path field
    Then the path input should contain "/home/user/documents"
    When I press "enter" to confirm
    Then the manual path entry dialog should be closed

  Scenario: Enter relative path with forward slashes
    When I press "t" to open manual path entry
    And I type "documents/projects/myapp" in the path field
    Then the path input should contain "documents/projects/myapp"
    When I press "enter" to confirm
    Then the manual path entry dialog should be closed

  Scenario: Build path character by character including slashes
    When I press "t" to open manual path entry
    And I build path character by character "home/user"
    Then the path input should contain "home/user"

  Scenario: Home directory shortcut still works at start
    When I press "t" to open manual path entry
    And I press "~" for home directory
    Then the path input should be set to the home directory
    When I type "/documents" in the path field
    Then the path should be extended with "/documents"

  Scenario: Tilde as regular character when not at start
    When I press "t" to open manual path entry
    And I type "backup~old" in the path field
    Then the path input should contain "backup~old"
