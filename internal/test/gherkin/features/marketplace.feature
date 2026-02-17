Feature: Marketplace Service
  As a marketplace service
  I want to extract ad IDs and handle rate limiting
  So that I can fetch ads without being blocked

  Background:
    Given a marketplace service with default: Extract blocket config

  Scenario: ad ID from annons URL
    Given URL "https://www.blocket.se/annons/123456"
    When I extract blocket ad ID
    Then the ad ID should be 123456

  Scenario: Extract blocket ad ID from item URL
    Given URL "https://www.blocket.se/item/999999"
    When I extract blocket ad ID
    Then the ad ID should be 999999

  Scenario: Extract blocket ad ID from URL with query
    Given URL "https://www.blocket.se/annons/123456?q=test"
    When I extract blocket ad ID
    Then the ad ID should be 123456

  Scenario: Extract blocket ad ID from invalid URL
    Given URL "invalid"
    When I extract blocket ad ID
    Then the ad ID should be 0

  Scenario: Extract blocket ad ID from other path
    Given URL "https://www.blocket.se/other/123"
    When I extract blocket ad ID
    Then the ad ID should be 0

  Scenario: Rate limiting works correctly
    Given a marketplace service
    When I make 5 requests with rate limiting
    Then the total time should be at least 800ms

  Scenario: Wait for rate limit with no error
    Given a marketplace service
    When I wait for rate limit
    Then no error should occur
