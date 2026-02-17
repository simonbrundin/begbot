Feature: Marketplace Ad Management user of the begbot system
  I want to fetch and manage marketplace
  As a ads
  So that I can track products for buying and selling

  Background:
    Given a marketplace service is configured

  Scenario: Extract blocket ad ID from URL with annons path
    Given the URL "https://www.blocket.se/annons/123456"
    When I extract the ad ID
    Then the ad ID should be "123456"

  Scenario: Extract blocket ad ID from URL with item path
    Given the URL "https://www.blocket.se/item/999999"
    When I extract the ad ID
    Then the ad ID should be "999999"

  Scenario: Extract blocket ad ID from URL with query parameters
    Given the URL "https://www.blocket.se/annons/123456?q=test"
    When I extract the ad ID
    Then the ad ID should be "123456"

  Scenario: Extract blocket ad ID from invalid URL
    Given the URL "invalid"
    When I extract the ad ID
    Then the ad ID should be "0"

  Scenario: Extract blocket ad ID from URL with wrong path
    Given the URL "https://www.blocket.se/other/123"
    When I extract the ad ID
    Then the ad ID should be "0"

  Scenario: Rate limiting works correctly
    Given rate limiting is enabled
    When I make "5" consecutive requests
    Then the requests should take at least "0.4" seconds

  Scenario: Fetch blocket ad from API (integration test)
    Given a valid blocket ad ID "124456789"
    When I fetch the ad from the API
    Then the request should either succeed or return an expected error for invalid ID
    And if successful, the title should not be empty
    And if successful, the ad text should not be empty
    And if successful, the price should be greater than "0"
