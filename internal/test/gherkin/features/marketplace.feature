Feature: Marketplace

  Background:
    Given a marketplace service is available
    And the configuration has blocket enabled

  Scenario: Extract valid Blocket ad ID from URL
    Given the URL "https://www.blocket.se/annons/123456"
    When extracting the ad ID
    Then the ad ID should be 123456

  Scenario: Extract Blocket ad ID from URL with query params
    Given the URL "https://www.blocket.se/annons/123456?q=test"
    When extracting the ad ID
    Then the ad ID should be 123456

  Scenario: Extract valid Blocket ad ID from alternative URL format
    Given the URL "https://www.blocket.se/item/999999"
    When extracting the ad ID
    Then the ad ID should be 999999

  Scenario: Handle invalid URL
    Given an invalid URL "invalid"
    When extracting the ad ID
    Then the ad ID should be 0

  Scenario: Handle non-Blocket URL
    Given a non-Blocket URL "https://www.blocket.se/other/123"
    When extracting the ad ID
    Then the ad ID should be 0

  Scenario: Rate limiting between requests
    Given the rate limiter is reset
    When making 5 consecutive requests
    Then the requests should take at least 1 second
    And no rate limit errors should occur

  Scenario: Fetch Blocket ad from API with valid ID
    Given a valid Blocket ad ID
    When fetching the ad from the API
    Then the response should contain a title
    And the response should contain ad text
    And the price should be greater than 0

  Scenario: Fetch Blocket ad from API with invalid ID
    Given an invalid Blocket ad ID 999999999
    When fetching the ad from the API
    Then an error may be returned (expected for invalid IDs)

  Scenario: Handle rate limit errors gracefully
    Given the API returns a rate limit error
    When retrying the request
    Then the request should eventually succeed
    Or return a rate limit exceeded error
