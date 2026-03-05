Feature: Frontend Ads Filtering
  As a user viewing the ads page
  I want to filter listings by tab
  So that I can see all listings or only those with potential profit

  Scenario: Show all listings in "Alla" tab
    Given frontend listings exist with various prices and valuations
    When I apply the frontend filter "all"
    Then all frontend listings should be returned

  Scenario: Empty result when no listings exist
    Given no frontend listings exist
    When I apply the frontend filter "all"
    Then 0 frontend listings should be returned

  Scenario: Show only profitable listings in "Potentiella" tab
    Given frontend listings with profit data:
      | id | potentialProfit | discountPercent |
      | 1  | 200             | 20              |
      | 2  | -100            | -10             |
    When I apply the frontend filter "potential"
    Then 1 frontend listing should be returned
    And the returned frontend listing should have id 1

  Scenario: Return empty when no listings have positive potential profit
    Given a frontend listing with id 1 and no potential profit data
    When I apply the frontend filter "potential"
    Then 0 frontend listings should be returned
