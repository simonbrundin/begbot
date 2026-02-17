Feature: Search History Management
  As a user of the begbot system
  I want to record and retrieve search history
  So that I can track search results and analyze trends

  Background:
    Given a search history service with mock database

  Scenario: Record a new search successfully
    When I record a search with term ID "1", description "iPhone 15 Pro", URL "https://blocket.se/...", results "10", new ads "3"
    Then the search should be recorded successfully
    And the search term ID should be "1"
    And the search term description should be "iPhone 15 Pro"
    And the results found should be "10"
    And the new ads found should be "3"
    And the URL should be "https://blocket.se/..."
    And the history ID should be set

  Scenario: Get search history with data
    Given the database has the following search history:
      | id | search_term_id | search_term_desc | url                    | results_found | new_ads_found |
      | 1  | 1              | iPhone 15        | https://blocket.se/1  | 15             | 5             |
      | 2  | 2              | MacBook Pro      | https://blocket.se/2  | 8              | 2             |
    When I get search history for page "1" with page size "20"
    Then I should receive "2" history records
    And the total count should be "2"
    And the first record should have description "iPhone 15"

  Scenario: Get empty search history
    Given the database has no search history
    When I get search history for page "1" with page size "20"
    Then I should receive "0" history records
    And the total count should be "0"

  Scenario: Get search history with pagination
    Given the database has "5" search history records
    When I get search history for page "1" with page size "2"
    Then I should receive "2" history records on page 1
    And the total count should be "5"
    When I get search history for page "2" with page size "2"
    Then I should receive "2" history records on page 2
    And the record should start at ID "3"
    When I get search history for page "3" with page size "2"
    Then I should receive "1" history record on page 3

  Scenario: Handle invalid page numbers
    Given the database has no search history
    When I get search history for page "0" with page size "20"
    Then the request should succeed
    And the count should be "0"
    When I get search history for page "-1" with page size "20"
    Then the request should succeed
    And the count should be "0"

  Scenario: Handle large page size
    Given the database has no search history
    When I get search history for page "1" with page size "200"
    Then the request should succeed

  Scenario: Handle database error when recording search
    Given the database returns error "database unavailable"
    When I try to record a search with term ID "1"
    Then I should receive an error

  Scenario: Handle database error when getting history
    Given the database returns error "database unavailable"
    When I try to get search history
    Then I should receive an error

  Scenario: Empty state detection
    Given the database has no search history
    When I get search history for page "1" with page size "20"
    Then the system should detect empty state
