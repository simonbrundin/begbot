Feature: Trading Rules Email Notification
  As a trading system
  I want to send email notifications when listings pass trading rules
  So that I can act on profitable opportunities

  Background:
    Given trading rules with minimum profit 500 SEK and minimum discount 10%

  Scenario: Email is sent when listing passes all trading rules
    Given a listing with price 6000 SEK and valuation 8000 SEK
    When evaluating the listing against trading rules
    Then profit should be 2000 SEK
    And discount should be approximately 25%
    And the listing should pass the trading rules

  Scenario: Email contains all required fields
    Given a listing with:
      | field       | value                          |
      | price       | 5000                           |
      | valuation   | 8000                           |
      | link        | https://blocket.se/item/123    |
      | description | Säljer iPhone 15 Pro i fint skick |
    And a new price of 15000 SEK
    When preparing the email notification
    Then the email should include the purchase price
    And the email should include the valuation
    And the email should include the discount percent
    And the email should include the new price
    And the email should include the profit
    And the email should include the description
    And the email should include the link

  Scenario: No email when profit is too low
    Given a listing with price 7500 SEK and valuation 8000 SEK
    When evaluating the listing against trading rules
    Then profit should be 500 SEK
    And the listing should not pass the minimum profit threshold of 500 SEK

  Scenario: No email when discount is too low
    Given a listing with price 7300 SEK and valuation 8000 SEK
    When evaluating the listing against trading rules
    Then discount should be approximately 8.75%
    And the listing should not pass the minimum discount threshold of 10%

  Scenario: Email is sent asynchronously
    Given a valid listing that passes trading rules
    When the email notification is triggered
    Then the email should be sent asynchronously without blocking

  Scenario: Handle missing email configuration
    Given no email configuration is set
    When a listing passes trading rules
    Then no crash should occur
