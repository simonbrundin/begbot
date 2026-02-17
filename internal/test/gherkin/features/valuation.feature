Feature: Valuation

  Background:
    Given a valuation compiler is available

  Scenario: Calculate weighted average with multiple inputs
    Given the following valuation inputs:
      | type     | value | confidence |
      | Method1  | 1000  | 0.8        |
      | Method2  | 1200  | 0.6        |
      | Method3  | 1100  | 0.7        |
    When the compiler calculates the weighted average
    Then the recommended price should be between 1000 and 1200
    And the confidence should be between 0.6 and 0.8

  Scenario: Calculate weighted average with single input
    Given a single valuation input with value 1500 and confidence 0.9
    When the compiler calculates the weighted average
    Then the recommended price should be 1500
    And the confidence should be 0.9

  Scenario: Handle empty inputs
    Given no valuation inputs
    When the compiler calculates the weighted average
    Then the recommended price should be 0
    And the confidence should be 0

  Scenario: Calculate historical price for days
    Given a historical valuation with K-value -10 and intercept 1500
    When calculating the price for 0 days
    Then the price should be 1500
    When calculating the price for 7 days
    Then the price should be 1430
    When calculating the price for 30 days
    Then the price should be 1200

  Scenario: Handle historical valuation with no data
    Given a historical valuation with no data
    When calculating the price for 7 days
    Then the price should be 0

  Scenario: Calculate profit
    Given a purchase price of 500 SEK
    And shipping cost of 50 SEK
    And estimated sell price of 1000 SEK
    When calculating the profit
    Then the profit should be 450 SEK

  Scenario: Calculate profit margin
    Given a profit of 450 SEK
    And total cost of 550 SEK
    When calculating the profit margin
    Then the margin should be approximately 0.818

  Scenario: Handle zero cost for profit margin
    Given a profit of 100 SEK
    And total cost of 0 SEK
    When calculating the profit margin
    Then the margin should be 0

  Scenario: Estimate sell probability with negative K value
    Given K value is -10 (price drops over time)
    When estimating sell probability for 7 days with target 30 days
    Then the probability should be 0.95
    When estimating sell probability for 30 days with target 30 days
    Then the probability should be 0.5

  Scenario: Estimate sell probability with positive K value
    Given K value is 10 (price increases over time)
    When estimating sell probability for 7 days with target 30 days
    Then the probability should be 0.1
    When estimating sell probability for 30 days with target 30 days
    Then the probability should be 0.5

  Scenario Outline: Database valuation method priority
    Given a database valuation method
    When getting the method name
    Then the name should be "Egen databas"
    And the priority should be <priority>

    Examples:
      | priority |
      | 1        |

  Scenario Outline: LLM new price method
    Given an LLM new price method
    When getting the method name
    Then the name should be "Nypris (LLM)"
    And the priority should be <priority>

    Examples:
      | priority |
      | 2        |

  Scenario Outline: Tradera valuation method
    Given a Tradera valuation method
    When getting the method name
    Then the name should be "Tradera"
    And the priority should be <priority>

    Examples:
      | priority |
      | 3        |

  Scenario Outline: Sold ads valuation method
    Given a sold ads valuation method
    When getting the method name
    Then the name should be "eBay/Marknadsplatser"
    And the priority should be <priority>

    Examples:
      | priority |
      | 4        |

  Scenario: Calculate confidence with no items
    Given a database valuation method with 0 sold items
    When calculating confidence
    Then the confidence should be 0

  Scenario: Calculate confidence with 2 items
    Given a database valuation method with 2 sold items
    When calculating confidence
    Then the confidence should be 0.3

  Scenario: Calculate confidence with 4 items
    Given a database valuation method with 4 sold items
    When calculating confidence
    Then the confidence should be 0.5

  Scenario: Calculate confidence with 8 items
    Given a database valuation method with 8 sold items
    When calculating confidence
    Then the confidence should be 0.7

  Scenario: Price should be in SEK not ören
    Given sold items with prices 100 SEK, 150 SEK, and 125 SEK
    When calculating the estimated price
    Then the price should be in SEK (not ören)

  Scenario: Compile with no valid inputs
    Given valuation inputs with zero value or confidence
    When compiling the valuation
    Then the recommended price should be 0

  Scenario: Validate valuation bounds - normal case
    Given valuation inputs with value 1500 and confidence 0.7
    And new price of 2000
    When compiling the weighted average
    Then no error should occur

  Scenario: Validate valuation bounds - unreasonable case
    Given a valuation input with value 150000 and confidence 0.7
    And new price of 2000
    When compiling the weighted average
    Then a warning should be logged for unreasonable valuation
