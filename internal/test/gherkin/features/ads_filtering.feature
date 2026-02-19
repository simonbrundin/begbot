Feature: Ads Page Filtering
  As a user viewing the ads page
  I want to filter listings by tab selection
  So that I can quickly find good value deals

  Background:
    Given I have a listing filter service

  Scenario: Filter shows all listings in "Alla" tab
    Given the following listings:
      | listing_id | price | valuation |
      | 1          | 1000  | 800       |
      | 2          | 500   | 1000      |
      | 3          | 2000  | 2500      |
    When I filter by "all" tab
    Then I should receive 3 listings

  Scenario: Filter shows only good value listings in "Prisvärda" tab
    Given the following listings:
      | listing_id | price | valuation |
      | 1          | 1000  | 800       |
      | 2          | 500   | 1000      |
      | 3          | 2000  | 2500      |
    When I filter by "good-value" tab
    Then I should receive 2 listings

  Scenario: No good value listings when price equals valuation
    Given the following listings:
      | listing_id | price | valuation |
      | 1          | 1000  | 1000      |
    When I filter by "good-value" tab
    Then I should receive 0 listings

  Scenario: No good value listings when price exceeds valuation
    Given the following listings:
      | listing_id | price | valuation |
      | 1          | 1500  | 1000      |
    When I filter by "good-value" tab
    Then I should receive 0 listings

  Scenario: Listing without valuation is not included in good value
    Given the following listings:
      | listing_id | price | valuation |
      | 1          | 500   |           |
    When I filter by "good-value" tab
    Then I should receive 0 listings

  Scenario: All listings are good value
    Given the following listings:
      | listing_id | price | valuation |
      | 1          | 500   | 1000      |
      | 2          | 800   | 1500      |
      | 3          | 1000  | 2000      |
    When I filter by "good-value" tab
    Then I should receive 3 listings

  Scenario: Empty listings array
    Given there are no listings
    When I filter by "all" tab
    Then I should receive 0 listings
    When I filter by "good-value" tab
    Then I should receive 0 listings
