Feature: Hidden directory exclusion
  Scenario: Exclude hidden directories when --exclude-hidden-dirs is used
    Given a source directory with files:
      | filename           | content |
      | .hidden/file.txt   | x       |
      | visible/file.txt   | y       |
    And an empty destination directory
    When I add extra args:
      | arg                   | value |
      | --exclude-hidden-dirs |       |
    And I run sync.sh in one-way mode
    Then the destination directory should contain the files:
      | filename         | content |
      | visible/file.txt | y       |
    And the destination directory should not contain the files:
      | filename         |
      | .hidden/file.txt |
