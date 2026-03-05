Feature: Weighted Valuation Calculation
  As a product valuation system
  I want to compute a weighted average of valuations from multiple sources
  So that I can recommend the best purchase price

  Scenario: Type is active when no configs exist (backward compatible)
    Given no valuation configs exist for product 1
    When I check if type 1 is active for product 1
    Then it should be active

  Scenario: Type is active when configs are empty for product
    Given an empty config list for product 1
    When I check if type 1 is active for product 1
    Then it should be active

  Scenario: Type is active when no config for specific type
    Given a config that only deactivates type 2 for product 1
    When I check if type 1 is active for product 1
    Then it should be active

  Scenario: Type is inactive when deactivated in config
    Given a config that deactivates type 1 for product 1
    When I check if type 1 is active for product 1
    Then it should be inactive

  Scenario: Type is active when explicitly activated in config
    Given a config that activates type 1 for product 1
    When I check if type 1 is active for product 1
    Then it should be active

  Scenario: Return null when no enabled valuation types
    Given no enabled valuation types
    When I compute weighted valuation for product 1
    Then the result should be null

  Scenario: Return null when product has no valuations
    Given enabled types: 1, 2
    And product 1 has no valuations
    When I compute weighted valuation for product 1
    Then the result should be null

  Scenario: Calculate correct average with equal weights
    Given enabled types: 1, 2
    And product 1 has valuation 1000 for type 1 and 2000 for type 2
    And both types have weight 1
    When I compute weighted valuation for product 1
    Then the average should be 1500

  Scenario: Calculate correct average with custom weights
    Given enabled types: 1, 2
    And product 1 has valuation 1000 for type 1 and 3000 for type 2
    And type 1 has weight 1 and type 2 has weight 3
    When I compute weighted valuation for product 1
    Then the average should be 2500

  Scenario: Safety is 100% when only one valuation exists
    Given enabled types: 1, 2
    And product 1 has only valuation 1000 for type 1
    When I compute weighted valuation for product 1
    Then the safety percent should be 100

  Scenario: Safety is lower when values diverge significantly
    Given enabled types: 1, 2
    And product 1 has valuation 100 for type 1 and 900 for type 2
    And both types have weight 1
    When I compute weighted valuation for product 1
    Then the safety percent should be less than 100

  Scenario: Return null when total weight is zero
    Given enabled types: 1
    And product 1 has valuation 1000 for type 1 with weight 0
    When I compute weighted valuation for product 1
    Then the result should be null

  Scenario: Ignore types without a valuation for the product
    Given enabled types: 1, 2
    And product 1 has only valuation 2000 for type 1
    And type 1 has weight 1 and type 2 has weight 1
    When I compute weighted valuation for product 1
    Then the average should be 2000
    And the safety percent should be 100

  Scenario: Exclude deactivated types from weighted average
    Given enabled types: 1, 2
    And product 1 has valuation 1000 for type 1 and 3000 for type 2
    And type 2 is deactivated for product 1
    And both types have weight 1
    When I compute weighted valuation for product 1
    Then the average should be 1000

  Scenario: Return null when all types are deactivated for product
    Given enabled types: 1, 2
    And product 1 has valuation 1000 for type 1 and 2000 for type 2
    And both type 1 and type 2 are deactivated for product 1
    When I compute weighted valuation for product 1
    Then the result should be null

  Scenario: Use all types when no config exists (backward compatible)
    Given enabled types: 1, 2
    And product 1 has valuation 1000 for type 1 and 2000 for type 2
    And no configs exist
    When I compute weighted valuation for product 1
    Then the average should be 1500

  Scenario: Last active type gets full weight when others deactivated
    Given enabled types: 1, 2, 3
    And product 1 has valuation 1000 for type 1, 2000 for type 2, 3000 for type 3
    And types 1 and 2 are deactivated for product 1
    When I compute weighted valuation for product 1
    Then the average should be 3000
