Feature: Search History Service
  As a bot user
  I want to record and retrieve search history
  So that I can track what searches have been performed

  Background:
    Given a search history service with mock database

  Scenario: Record a new search successfully
    When I record a search with term ID 1, description "iPhone 15 Pro", URL "https://blocket.se/...", results 10, new ads 3
    Then the search should be saved successfully
    And the search should have ID set
    And the search term description should be "iPhone 15 Pro"
    And results found should be 10
    And new ads found should be 3

  Scenario: Get search history with data
    Given the database has search history records
      | id | search_term_id | search_term_desc | url                    | results_found | new_ads_found |
      | 1  | 1              | iPhone 15        | https://blocket.se/iphone | 15            | 5             |
      | 2  | 2              | MacBook Pro      | https://blocket.se/macbook | 8             | 2             |
    When I get history for page 1 with page size 20
    Then I should receive 2 history records
    And total count should be 2
    And first record should have description "iPhone 15"

  Scenario: Get search history with empty database
    Given the database has no search history
    When I get history for page 1 with page size 20
    Then I should receive 0 history records
    And total count should be 0

  Scenario: Get search history with pagination
    Given the database has 5 search history records
    When I get page 1 with page size 2
    Then I should receive 2 records
    And total count should be 5
    When I get page 2 with page size 2
    Then I should receive 2 records
    And first record on page 2 should have ID 3
    When I get page 3 with page size 2
    Then I should receive 1 record

  Scenario: Get search history with invalid page
    Given the database has no search history
    When I get history for page 0 with page size 20
    Then I should receive 0 history records
    When I get history for page -1 with page size 20
    Then I should receive 0 history records

  Scenario: Get search history with large page size
    Given the database has no search history
    When I get history for page 1 with page size 200
    Then no error should occur

  Scenario: Get search history with database error
    Given the database returns an error "database unavailable"
    When I get history for page 1 with page size 20
    Then an error should be returned

  Scenario: Record search with database error
    Given the database returns an error "database unavailable"
    When I record a search with term ID 1, description "Test", URL "https://...", results 10, new ads 2
    Then an error should be returned

  Scenario: Empty state detection
    Given the database has no search history
    When I get history for page 1 with page size 20
    Then it should indicate empty state
