Feature: Search History

  Background:
    Given a search history service is available
    And the database is connected

  Scenario: Record a new search
    When a user searches for "iPhone 15 Pro" with URL "https://blocket.se/iphone"
    And the search finds 10 results with 3 new ads
    Then the search should be saved successfully
    And the search should have a valid ID
    And the search term description should be "iPhone 15 Pro"
    And the results found should be 10
    And the new ads found should be 3

  Scenario: Get search history with data
    Given the database has 2 search records
    When the user requests search history for page 1 with 20 items per page
    Then the response should contain 2 search records
    And the total count should be 2
    And the first record should have search term "iPhone 15"

  Scenario: Get empty search history
    Given the database has no search records
    When the user requests search history
    Then the response should contain 0 search records
    And the total count should be 0

  Scenario: Paginate search history
    Given the database has 5 search records
    When the user requests page 1 with 2 items per page
    Then the response should contain 2 items
    And the total count should be 5
    When the user requests page 2 with 2 items per page
    Then the response should contain 2 items
    And the first item on page 2 should have ID 3
    When the user requests page 3 with 2 items per page
    Then the response should contain 1 item

  Scenario: Handle invalid pagination parameters
    Given the database has no search records
    When the user requests page 0
    Then the request should succeed
    And the count should be 0
    When the user requests page -1
    Then the request should succeed
    And the count should be 0

  Scenario: Handle database error when recording search
    Given the database is unavailable
    When the user attempts to record a search
    Then an error should be returned

  Scenario: Handle database error when getting history
    Given the database is unavailable
    When the user requests search history
    Then an error should be returned

  Scenario: Large page size handling
    Given the database has no search records
    When the user requests page 1 with 200 items per page
    Then the request should succeed
