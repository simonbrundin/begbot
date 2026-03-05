Feature: Search Terms Service
  As a scraper system
  I want to manage and use search terms for marketplace searches
  So that I can find relevant product listings

  Scenario: Create a valid search term with marketplace
    Given a search term with description "iPhone Search" and URL "https://blocket.se/?q=iphone"
    And the search term is linked to marketplace ID 1
    Then the description should be "iPhone Search"
    And the URL should not be empty
    And the marketplace ID should be 1

  Scenario: Create a search job with marketplace
    Given a search term "iPhone" linked to Blocket marketplace
    When creating a search job
    Then the job URL should not be empty
    And the marketplace should not be nil

  Scenario: Search job for Tradera marketplace
    Given a search term "Lego Star Wars" linked to Tradera marketplace
    When creating a search job
    Then the marketplace name should be "Tradera"

  Scenario: Inactive search term
    Given a search term with isActive false
    Then the search term should be inactive

  Scenario: Search term without marketplace
    Given a search term with no marketplace ID
    Then the marketplace ID should be null

  Scenario: Service accepts context
    Given a search term service
    When passing a context
    Then no error should occur
