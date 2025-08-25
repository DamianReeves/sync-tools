Feature: Ignore and unignore behavior
  Scenario: .syncignore prevents files from being copied unless unignored
    Given a source directory with files:
      | filename      | content  |
      | keep.txt      | keep     |
      | secret.log    | secret   |
    And a source .syncignore with:
      """
      *.log
      !secret.log
      """
    And an empty destination directory
    When I run sync.sh in one-way mode
    Then the destination directory should contain the files:
      | filename      | content  |
      | keep.txt      | keep     |
      | secret.log    | secret   |
