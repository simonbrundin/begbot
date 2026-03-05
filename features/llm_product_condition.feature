Feature: LLM Product Condition Assessment
  As a bot system
  I want to assess whether a product is in acceptable condition
  So that I can decide whether to purchase it

  Scenario: Product is intact with no issues
    Given a product condition result with isIntact true and no scratches
    And no issues found
    Then the product should be valid for purchase

  Scenario: Product has major issues and is not intact
    Given a product condition result with isIntact false
    And issues found: "sprucken skärm", "vattenskada"
    Then the product should not be valid for purchase
    And there should be 2 issues found

  Scenario: Product has minor scratches but is still intact
    Given a product condition result with isIntact true and minor scratches
    Then the product should be valid for purchase
    And the minor scratches flag should be true

  Scenario: Product is valid for purchase when intact without scratches
    Given a product with isIntact true and no scratches
    When evaluating purchase validity
    Then the product should be valid for purchase

  Scenario: Product is valid for purchase when intact with minor scratches
    Given a product with isIntact true and minor scratches
    When evaluating purchase validity
    Then the product should be valid for purchase

  Scenario: Product is not valid for purchase when not intact
    Given a product with isIntact false and issues "sprucken skärm"
    When evaluating purchase validity
    Then the product should not be valid for purchase
