Feature: Tradera Valuation Method
  As a valuation system
  I want to estimate product prices from Tradera sold listings
  So that I can provide accurate purchase recommendations

  Scenario: Parse Tradera API response and calculate valuation
    Given Tradera valuation is enabled
    And the Tradera API returns average price 2045 SEK with 901 sold items
    When I valuate a product on Tradera
    Then the valuation value should be 2045
    And the confidence should be at least 0.84 for 901 items

  Scenario: Return error when no prices found
    Given Tradera valuation is enabled
    And the Tradera API returns zero average price and zero items
    When I valuate a product on Tradera
    Then a Tradera error should be returned
    And the Tradera valuation should be nil

  Scenario: Return nil when Tradera is disabled
    Given Tradera valuation is disabled
    When I valuate a product on Tradera
    Then the Tradera valuation should be nil
    And no Tradera error should be returned
