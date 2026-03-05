Feature: Normalize Valuation Weights
  As a system managing valuation types
  I want to normalize weights so active types always sum to 100%
  So that the weighted valuation calculation is correct

  Scenario: Equal distribution among active types
    Given 3 active valuation types with weight 0 each
    When I normalize the weights
    Then each active type should have weight approximately 33.33

  Scenario: Active weights always sum to 100
    Given valuation types:
      | type | active | weight |
      | 1    | true   | 40     |
      | 2    | true   | 60     |
      | 3    | false  | 0      |
    When I normalize the weights
    Then the sum of active weights should be 100
    And the inactive type weight should be 0

  Scenario: Deactivating a type redistributes to remaining active types
    Given valuation types:
      | type | active | weight |
      | 1    | true   | 50     |
      | 2    | false  | 50     |
    When I normalize the weights
    Then type 1 should have weight 100
    And type 2 should have weight 0

  Scenario: Reactivating a type gives equal share to all
    Given valuation types:
      | type | active | weight |
      | 1    | true   | 60     |
      | 2    | true   | 40     |
      | 3    | true   | 0      |
    When I normalize the weights
    Then all active types should have equal weight approximately 33.33

  Scenario: Reactivate one of two types results in 50/50 split
    Given valuation types:
      | type | active | weight |
      | 1    | true   | 100    |
      | 2    | true   | 0      |
    When I normalize the weights
    Then type 1 should have weight 50
    And type 2 should have weight 50

  Scenario: No active types means all weights are zero
    Given valuation types:
      | type | active | weight |
      | 1    | false  | 50     |
      | 2    | false  | 50     |
    When I normalize the weights
    Then all types should have weight 0

  Scenario: Single active type gets 100% weight
    Given 1 active valuation type with weight 30
    When I normalize the weights
    Then that type should have weight 100

  Scenario: Input slice is not mutated
    Given valuation types:
      | type | active | weight |
      | 1    | true   | 30     |
      | 2    | true   | 70     |
    When I normalize the weights
    Then the original input weights should be unchanged
