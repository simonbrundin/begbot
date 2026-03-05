Feature: Blocket Valuation Method
  As a valuation system
  I want to estimate product prices from Blocket listings
  So that I can provide accurate purchase recommendations

  Scenario: Parse API response and calculate valuation
    Given Blocket valuation is enabled
    And the Blocket API returns 12 listings with prices between 1000 and 2200 SEK
    When I valuate a product on Blocket
    Then the valuation should have a positive value
    And the confidence should be at least 0.5 for 10 or more items

  Scenario: Filter outliers using IQR method
    Given Blocket valuation is enabled
    And the Blocket API returns listings with outliers:
      | listing | price |
      | 1       | 1000  |
      | 2       | 1100  |
      | 3       | 1200  |
      | 4       | 1300  |
      | 5       | 100   |
      | 6       | 10000 |
      | 7       | 1250  |
      | 8       | 1150  |
    When I valuate a product on Blocket
    Then the valuation should be less than 2000
    And outlier prices should be filtered out

  Scenario: Return nil when Blocket is disabled
    Given Blocket valuation is disabled
    When I valuate a product on Blocket
    Then the valuation should be nil
    And no Blocket error should be returned

  Scenario: Return error when no prices found
    Given Blocket valuation is enabled
    And the Blocket API returns no listings
    When I valuate a product on Blocket
    Then a Blocket error should be returned
    And the valuation should be nil

  Scenario: Cache results to avoid duplicate API calls
    Given Blocket valuation is enabled
    And the Blocket API is available
    When I valuate the same product twice
    Then only one API request should be made
    And both valuations should have the same value

  Scenario: Valuate with only model name (no manufacturer)
    Given Blocket valuation is enabled
    And the Blocket API returns listings for the model
    When I valuate a product with only a model name
    Then the valuation should have a positive value

  Scenario: Calculate quartiles for price statistics
    Given prices: 100, 200, 300, 400, 500, 600, 700, 800
    When I calculate quartiles
    Then Q1 should be 200
    And Q3 should be 600
    And IQR should be 400

  Scenario: Filter outliers from a price list
    Given prices with outlier: 100, 200, 300, 400, 500, 10000
    When I filter outliers using IQR
    Then 5 prices should remain

  Scenario: Calculate median of prices
    Given an odd-length price list: 1, 2, 3, 4, 5
    When I calculate the median
    Then the median should be 3

  Scenario: Calculate median of even-length list
    Given an even-length price list: 1, 2, 3, 4
    When I calculate the median
    Then the median should be 2
