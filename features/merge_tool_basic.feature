Feature: Basic Merge Tool Integration
  As a developer
  I want to resolve sync conflicts with different strategies
  So that I can handle conflicting files appropriately

  Scenario: Conflict resolution with newest-wins strategy
    Given I have a source file "conflict.txt" with content "source content" modified at "2025-08-30T10:30:00"
    And I have a destination file "conflict.txt" with content "dest content" modified at "2025-08-30T10:40:00"
    And I have a plan file "resolve.plan" containing:
      """
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file conflict.txt 14B 2025-08-30T10:40:00 [CONFLICT: both-modified, auto:newest]
      """
    When I run sync-tools with arguments "sync --apply-plan resolve.plan"
    Then the command should succeed
    And both source and destination should contain "conflict.txt" with content "dest content"

  Scenario: Conflict resolution with source-wins strategy  
    Given I have a source file "test.txt" with content "source wins"
    And I have a destination file "test.txt" with content "dest loses"
    And I have a plan file "source-wins.plan" containing:
      """
      # Source: ./test_source
      # Destination: ./test_dest
      # Mode: two-way
      
      <> file test.txt 12B 2025-08-30T10:30:00 [CONFLICT: both-modified, auto:source]
      """
    When I run sync-tools with arguments "sync --apply-plan source-wins.plan"
    Then the command should succeed  
    And both source and destination should contain "test.txt" with content "source wins"