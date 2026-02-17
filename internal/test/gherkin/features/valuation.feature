Feature: Product Valuation
  As a user of the begbot system
  I want to value products using multiple methods
  So that I can determine fair prices for buying and selling

  Background:
    Given a valuation service is available

  Scenario: Database valuation method has correct name and priority
    When I check the database valuation method
    Then the name should be "Egen databas"
    And the priority should be "1"

  Scenario: LLM new price method has correct name and priority
    When I check the LLM new price method
    Then the name should be "Nypris (LLM)"
    And the priority should be "2"

  Scenario: Tradera valuation method has correct name and priority
    When I check the Tradera valuation method
    Then the name should be "Tradera"
    And the priority should be "3"

  Scenario: Sold ads valuation method has correct name and priority
    When I check the sold ads valuation method
    Then the name should be "eBay/Marknadsplatser"
    And the priority should be "4"

  Scenario: Calculate weighted average with multiple inputs
    Given I have the following valuation inputs:
      | type     | value | confidence |
      | Method1  | 1000  | 0.8        |
      | Method2  | 1200  | 0.6        |
      | Method3  | 1100  | 0.7        |
    When I compile the weighted average
    Then the recommended price should be between "1000" and "1200"
    And the confidence should be between "0.6" and "0.8"

  Scenario: Calculate weighted average with single input
    Given I have a single valuation input with value "1500" and confidence "0.9"
    When I compile the weighted average
    Then the recommended price should be "1500"
    And the confidence should be "0.9"

  Scenario: Calculate weighted average with no valid inputs
    Given I have no valuation inputs
    When I compile the weighted average
    Then the recommended price should be "0"
    And the confidence should be "0"

  Scenario: Historical valuation calculates price for days
    Given a historical valuation with K value "-10" and intercept "1500"
    When I calculate price for "0" days
    Then the price should be "1500"
    When I calculate price for "7" days
    Then the price should be "1430"
    When I calculate price for "30" days
    Then the price should be "1200"

  Scenario: Historical valuation with no data
    Given a historical valuation with no data
    When I calculate price for "7" days
    Then the price should be "0"

  Scenario: Calculate profit
    Given buy price "500", shipping cost "50", and estimated sell price "1000"
    When I calculate the profit
    Then the profit should be "450"

  Scenario: Calculate profit margin
    Given profit "450", buy price "500", and shipping cost "50"
    When I calculate the profit margin
    Then the margin should be approximately "0.818"

  Scenario: Calculate profit margin with zero cost
    Given profit "100", buy price "0", and shipping cost "0"
    When I calculate the profit margin
    Then the margin should be "0"

  Scenario: Estimate sell probability with negative K value
    Given K value is "-10"
    When I estimate sell probability for "7" days on market with target "30"
    Then the probability should be "0.95"
    When I estimate sell probability for "30" days on market with target "30"
    Then the probability should be "0.5"

  Scenario: Estimate sell probability with positive K value
    Given K value is "10"
    When I estimate sell probability for "7" days on market with target "30"
    Then the probability should be "0.1"
    When I estimate sell probability for "30" days on market with target "30"
    Then the probability should be "0.5"

  Scenario: Calculate confidence based on number of items
    Given I have "0" sold items
    When I calculate confidence
    Then the confidence should be "0"
    Given I have "2" sold items
    When I calculate confidence
    Then the confidence should be "0.3"
    Given I have "4" sold items
    When I calculate confidence
    Then the confidence should be "0.5"
    Given I have "8" sold items
    When I calculate confidence
    Then the confidence should be "0.7"

  Scenario: Compile with no inputs
    Given I have no valuation inputs
    When I compile the valuations
    Then the recommended price should be "0"
    And the confidence should be "0"

  Scenario: Compile with invalid inputs (zero value or confidence)
    Given I have the following invalid valuation inputs:
      | type     | value | confidence |
      | Method1  | 0     | 0.8        |
      | Method2  | 1000  | 0          |
    When I compile the valuations
    Then the recommended price should be "0"

  Scenario: Database price in ören should be converted to kronor
    Given sold items with prices in ören: "10000", "15000", "12500"
    When I calculate the weighted average
    Then the result should be in SEK (kronor), not ören
