# Language: en
# Shortcut Character Conflicts in Path Entry

Feature: Path Entry Without Shortcut Interference
  As a user entering file paths manually
  I want to type any characters including those that might be shortcuts
  So that I can specify complete file paths with any valid filename characters

  Background:
    Given I am in the sync wizard
    And I am on the source selection screen

  Scenario: Enter path with tilde in filename
    When I press "t" to open manual path entry
    And I type "backup~old" in the path field
    Then the path input should contain "backup~old"
    When I press "enter" to confirm
    Then the manual path entry dialog should be closed

  Scenario: Enter path with special characters that were shortcuts
    When I press "t" to open manual path entry
    And I type "file~version.txt" in the path field
    Then the path input should contain "file~version.txt"

  Scenario: Type path with various special characters
    When I press "t" to open manual path entry
    And I type "project-v2.1_backup~final.tar.gz" in the path field
    Then the path input should contain "project-v2.1_backup~final.tar.gz"

  Scenario: Enter absolute path with tilde in directory name
    When I press "t" to open manual path entry
    And I type "/home/user/old~versions/file.txt" in the path field
    Then the path input should contain "/home/user/old~versions/file.txt"

  Scenario: Type path with numbers that were shortcuts
    When I press "t" to open manual path entry
    And I type "dir123/file456.txt" in the path field
    Then the path input should contain "dir123/file456.txt"

  Scenario: Backspace and ctrl+u still work as expected
    When I press "t" to open manual path entry
    And I type "wrongpath" in the path field
    Then the path input should contain "wrongpath"
    When I press "backspace" to remove character
    Then the path should be shortened
    When I press "ctrl+u" to clear path
    Then the path input should be empty
