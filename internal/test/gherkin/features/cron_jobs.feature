Feature: Cron Jobs Service
  As a bot user
  I want to manage scheduled scraper jobs
  So that I can automate when the scraper runs

  Background:
    Given a cron job service with mock database

  Scenario: Create a new cron job successfully
    When I create a cron job with name "Daily iPhone scan", expression "0 8 * * *", search term IDs [1, 2], and active true
    Then the cron job should be saved successfully
    And the cron job should have ID set
    And the cron job name should be "Daily iPhone scan"
    And the cron job expression should be "0 8 * * *"
    And the cron job should be active

  Scenario: Get all cron jobs
    Given the database has cron job records
      | id | name               | cron_expression | search_term_ids | is_active |
      | 1  | Daily iPhone scan  | 0 8 * * *       | [1,2]           | true      |
      | 2  | Hourly check        | 0 * * * *       | []              | false     |
    When I get all cron jobs
    Then I should receive 2 cron job records
    And the first cron job should have name "Daily iPhone scan"

  Scenario: Update a cron job
    Given the database has cron job records
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Daily scan       | 0 8 * * *       | [1]             | true      |
    When I update cron job 1 to name "Twice daily scan", expression "0 8,20 * * *", search term IDs [1, 2], active false
    Then the cron job should be updated successfully
    And the cron job name should be "Twice daily scan"
    And the cron job expression should be "0 8,20 * * *"
    And the cron job should be inactive

  Scenario: Delete a cron job
    Given the database has cron job records
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Old job          | 0 8 * * *       | [1]             | true      |
    When I delete cron job 1
    Then the cron job should be deleted successfully
    And there should be 0 cron jobs in the database

  Scenario: Toggle cron job active status
    Given the database has cron job records
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Test job         | 0 * * * *       | []              | true      |
    When I toggle cron job 1 active status
    Then the cron job should be updated successfully
    And the cron job should be inactive

  Scenario: Create cron job with empty search term IDs (all terms)
    When I create a cron job with name "All terms scan", expression "0 6 * * *", search term IDs [], active true
    Then the cron job should be saved successfully
    And the cron job should have empty search term IDs

  Scenario: Invalid cron expression
    Given the database has cron job records
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Test job         | 0 8 * * *       | []              | true      |
    When I update cron job 1 to name "Bad job", expression "invalid", search term IDs [], active true
    Then an error should be returned
    And the error message should contain "invalid cron expression"

  Scenario: Get cron job by ID
    Given the database has cron job records
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Specific job     | 0 8 * * *       | [1,2,3]         | true      |
      | 2  | Other job        | 0 * * * *       | []              | false     |
    When I get cron job by ID 1
    Then I should receive 1 cron job record
    And the cron job name should be "Specific job"
    And the cron job expression should be "0 8 * * *"

  Scenario: Cron expression with special characters
    When I create a cron job with name "Weekday scan", expression "0 9 * * 1-5", search term IDs [1], active true
    Then the cron job should be saved successfully
    And the cron job expression should be "0 9 * * 1-5"
