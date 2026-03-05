Feature: Utility Functions
  As a frontend application
  I want utility functions for formatting and calculation
  So that I can display data consistently

  Scenario: Format öre to kronor correctly
    Given the amount 10000 öre
    When I format it as currency
    Then the result should be "100.00 kr"

  Scenario: Format small amount to kronor
    Given the amount 500 öre
    When I format it as currency
    Then the result should be "5.00 kr"

  Scenario: Format zero amount
    Given the amount 0 öre
    When I format it as currency
    Then the result should be "0.00 kr"

  Scenario: Format negative amount
    Given the amount -5000 öre
    When I format it as currency
    Then the result should be "-50.00 kr"

  Scenario: Format ISO date to Swedish format
    Given the date "2024-01-15"
    When I format it as a date
    Then the result should be "2024-01-15"

  Scenario: Calculate profit correctly
    Given a trade item with:
      | field                  | value |
      | buy_price              | 5000  |
      | buy_shipping_cost      | 500   |
      | sell_price             | 8000  |
      | sell_packaging_cost    | 200   |
      | sell_postage_cost      | 100   |
      | sell_shipping_collected| 0     |
    When I calculate the profit
    Then the profit should be 2200

  Scenario: Calculate profit with missing sell values
    Given a trade item with buy_price 5000 and no sell price
    When I calculate the profit
    Then the profit should be -5000
