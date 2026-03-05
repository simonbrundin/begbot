Feature: Product Model JSON Serialization
  As a frontend consuming the API
  I want product data to be correctly serialized to JSON
  So that I can display it without errors

  Scenario: Product with null created_at should omit field from JSON
    Given a product with null created_at
    When I serialize the product to JSON
    Then the JSON should not contain "created_at"

  Scenario: Product with valid created_at should include it in JSON
    Given a product with created_at "2024-01-15T10:30:00Z"
    When I serialize the product to JSON
    Then the JSON should contain "created_at"

  Scenario: Empty brand and name should be preserved in JSON
    Given a product with empty brand and name
    When I serialize the product to JSON
    Then the JSON should contain brand as empty string
    And the JSON should contain name as empty string

  Scenario: Enabled field should be boolean true or false in JSON
    Given a product with enabled set to true
    When I serialize the product to JSON
    Then the JSON should contain "enabled":true

  Scenario: Product with all null fields should not have zero time in JSON
    Given a product with all null optional fields
    When I serialize the product to JSON
    Then the JSON should not contain the zero time "0001-01-01T00:00:00Z"
