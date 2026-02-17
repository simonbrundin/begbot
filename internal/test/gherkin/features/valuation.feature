Feature: Valuation Service
  As a valuation system
  I want to calculate item values using multiple methods
  So that I can provide accurate price recommendations

  Background:
    Given a valuation compiler

  Scenario: Database valuation method returns correct name
    Given a DatabaseValuationMethod
    When I get the method name
    Then it should return "Egen databas"

  Scenario: Database valuation method returns correct priority
    Given a DatabaseValuationMethod
    When I get the method priority
    Then it should return 1

  Scenario: LLM new price method returns correct name
    Given a LLMNewPriceMethod
    When I get the method name
    Then it should return "Nypris (LLM)"

  Scenario: LLM new price method returns correct priority
    Given a LLMNewPriceMethod
    When I get the method priority
    Then it should return 2

  Scenario: Tradera valuation method returns correct name
    Given a TraderaValuationMethod
    When I get the method name
    Then it should return "Tradera"

  Scenario: Tradera valuation method returns correct priority
    Given a TraderaValuationMethod
    When I get the method priority
    Then it should return 3

  Scenario: Sold ads valuation method returns correct name
    Given a SoldAdsValuationMethod
    When I get the method name
    Then it should return "eBay/Marknadsplatser"

  Scenario: Sold ads valuation method returns correct priority
    Given a SoldAdsValuationMethod
    When I get the method priority
    Then it should return 4

  Scenario: Compile weighted average with multiple inputs
    Given I have valuation inputs:
      | type    | value | confidence |
      | Method1 | 1000  | 0.8        |
      | Method2 | 1200  | 0.6        |
      | Method3 | 1100  | 0.7        |
    When I compile weighted average
    Then I should get a recommended price between 1000 and 1200
    And I should get a confidence between 0.6 and 0.8

  Scenario: Compile weighted average with single input
    Given I have valuation inputs:
      | type    | value | confidence |
      | Method1 | 1500  | 0.9        |
    When I compile weighted average
    Then the recommended price should be 1500
    And the confidence should be 0.9

  Scenario: Compile weighted average with no valid inputs
    Given I have no valuation inputs
    When I compile weighted average
    Then the recommended price should be 0
    And the confidence should be 0

  Scenario: Compile with no inputs returns zero
    When I compile with empty inputs
    Then the recommended price should be 0
    And the confidence should be 0

  Scenario: Compile with invalid inputs returns zero
    Given I have valuation inputs:
      | type    | value | confidence |
      | Method1 | 0     | 0.8        |
      | Method2 | 1000  | 0           |
    When I compile
    Then the recommended price should be 0

  Scenario: Calculate price for days with negative K value
    Given a historical valuation with:
      | field           | value |
      | has_data        | true  |
      | k_value         | -10.0 |
      | intercept       | 1500.0|
      | average_price   | 1400.0|
    When I calculate price for 0 days
    Then the price should be 1500.0
    When I calculate price for 7 days
    Then the price should be 1430.0
    When I calculate price for 30 days
    Then the price should be 1200.0

  Scenario: Calculate price for days with no data
    Given a historical valuation with:
      | field           | value |
      | has_data        | false |
      | k_value         | -10.0 |
      | intercept       | 1500.0|
      | average_price   | 1400.0|
    When I calculate price for 7 days
    Then the price should be 0

  Scenario: Calculate profit correctly
    Given buy price 500, shipping cost 50, estimated sell price 1000
    When I calculate profit
    Then the profit should be 450.0

  Scenario: Calculate profit margin correctly
    Given buy price 500, shipping cost 50, profit 450
    When I calculate profit margin
    Then the margin should be approximately 0.818

  Scenario: Calculate profit margin with zero cost
    Given buy price 0, shipping cost 0, profit 100
    When I calculate profit margin
    Then the margin should be 0

  Scenario: Estimate sell probability with negative K value
    Given K value -10.0 and target days 30
    When I estimate sell probability for 7 days on market
    Then the probability should be 0.95
    When I estimate sell probability for 30 days on market
    Then the probability should be 0.5

  Scenario: Estimate sell probability with positive K value
    Given K value 10.0 and target days 30
    When I estimate sell probability for 7 days on market
    Then the probability should be 0.1
    When I estimate sell probability for 30 days on market
    Then the probability should be 0.5

  Scenario: Calculate weight for recent vs old items
    Given a DatabaseValuationMethod
    When I calculate weight for item sold yesterday
    And I calculate weight for item sold 200 days ago
    Then recent item should have higher weight than old item

  Scenario: Calculate confidence based on number of items
    Given a DatabaseValuationMethod
    When I calculate confidence for 0 items
    Then confidence should be 0
    When I calculate confidence for 2 items
    Then confidence should be 0.3
    When I calculate confidence for 4 items
    Then confidence should be 0.5
    When I calculate confidence for 8 items
    Then confidence should be 0.7

  Scenario: Price should be in SEK not ören
    Given a DatabaseValuationMethod
    When I calculate weighted price with items priced in ören
    Then the result should be in SEK (not ören)
