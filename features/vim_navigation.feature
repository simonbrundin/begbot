Feature: Vim Navigation
  As a user browsing lists in the application
  I want to use vim-style keyboard navigation (j/k keys)
  So that I can navigate quickly without using the mouse

  Scenario: Move selection down with j key
    Given a list with 5 items
    And the navigation is focused
    When I press j
    Then the selected index should be 0
    When I press j again
    Then the selected index should be 1

  Scenario: Move selection up with k key
    Given a list with 5 items
    And the navigation is focused
    When I press k
    Then the selected index should be 4
    When I press k again
    Then the selected index should be 3

  Scenario: Visual selection is displayed when selected
    Given a list with 3 items
    And the navigation is focused
    When I press j
    Then the selected index should not be null
    And the selected index should be 0

  Scenario: Stay on first item when pressing k at top
    Given a list with 5 items
    And the navigation is focused
    When I navigate down to index 0
    And I press k
    Then the selected index should remain 0

  Scenario: Stay on last item when pressing j at bottom
    Given a list with 5 items
    And the navigation is focused
    When I navigate to the last item
    And I press j again
    Then the selected index should remain 4

  Scenario: Handle single item list
    Given a list with 1 item
    And the navigation is focused
    When I press j
    Then the selected index should be 0
    When I press j again
    Then the selected index should be 0
    When I press k
    Then the selected index should be 0

  Scenario: Do not respond when not focused
    Given a list with 5 items
    And the navigation is not focused
    When I press j
    Then the selected index should be null
    When I press k
    Then the selected index should be null

  Scenario: Track focus state changes
    Given a list with 5 items
    When I set focus to true and press j
    Then the selected index should be 0
    When I set focus to false and press j
    Then the selected index should remain 0
    When I set focus to true and press j
    Then the selected index should be 1

  Scenario: Handle empty list without crashing
    Given a list with 0 items
    And the navigation is focused
    When I press j
    Then no navigation error should occur
    And the selected index should be null

  Scenario: Clear selection with ESC
    Given a list with 5 items
    And the navigation is focused
    When I press j to select index 0
    And I clear the selection
    Then the selected index should be null

  Scenario: Handle item count change to zero
    Given a list with 5 items
    And the navigation is focused
    When I navigate to index 0
    And the item count changes to 0
    Then the selected index should be null

  Scenario: Handle item count change where selected index becomes invalid
    Given a list with 5 items
    And the navigation is focused
    When I navigate to index 2
    And the item count changes to 2
    Then the selected index should be 1
