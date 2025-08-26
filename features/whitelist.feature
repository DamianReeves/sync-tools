Feature: Whitelist (only) mode
  Scenario: Whitelist restricts synced paths to only the listed items
    Given a source directory with files:
      | filename      | content  |
      | keep.txt      | keep     |
      | other.txt     | other    |
      | docs/readme.adoc | doc   |
    And an empty destination directory
    When I whitelist the paths:
      | path         |
      | keep.txt     |
      | docs/readme.adoc |
  And I run sync-tools sync in one-way mode
    Then the destination directory should contain the files:
      | filename      | content  |
      | keep.txt      | keep     |
      | docs/readme.adoc | doc   |
