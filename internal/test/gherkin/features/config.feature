Feature: Configuration

  Background:
    Given a configuration system is available

  Scenario: Load configuration with LLM models defined
    Given a config file with provider "openrouter"
    And API key "test-key"
    And site URL "http://localhost:3000"
    And site name "Begbot"
    And default model "deepseek/deepseek-v3.2"
    And models:
      | task                    | model                              |
      | ExtractProductInfo      | anthropic/claude-sonnet-4-20250514 |
      | ValidateProduct         | google/gemini-2.5-pro              |
      | CheckProductCondition   | openai/gpt-4o-mini                 |
      | EstimateNewPrice        | deepseek/deepseek-chat             |
    When loading the configuration
    Then the provider should be "openrouter"
    And the API key should be "test-key"
    And the site URL should be "http://localhost:3000"
    And the site name should be "Begbot"
    And the default model should be "deepseek/deepseek-v3.2"
    And there should be 4 models defined

  Scenario: Load configuration without LLM models
    Given a config file with provider "openrouter"
    And API key "test-key"
    And default model "openai/gpt-4o"
    And no models defined
    When loading the configuration
    Then the models count should be 0

  Scenario: Load configuration with database settings
    Given a config file with:
      | setting   | value      |
      | host      | localhost  |
      | port      | 5432       |
      | user      | test       |
      | password  | test       |
      | name      | testdb     |
      | sslmode   | require    |
    When loading the configuration
    Then the database host should be "localhost"
    And the database port should be 5432

  Scenario: Load configuration with scraping settings
    Given a config file with scraping:
      | marketplace | enabled | timeout |
      | tradera     | false   | 10s     |
      | blocket     | false   | 10s     |
    When loading the configuration
    Then tradera should be disabled
    And blocket should be disabled

  Scenario: Load configuration with valuation settings
    Given a config file with valuation:
      | setting            | value |
      | target_sell_days   | 14    |
      | min_profit_margin  | 0.15  |
      | safety_margin      | 0.2   |
    When loading the configuration
    Then the target sell days should be 14
    And the minimum profit margin should be 0.15
    And the safety margin should be 0.2

  Scenario: Load configuration with email settings
    Given a config file with email:
      | setting      | value          |
      | smtp_host    | localhost      |
      | smtp_port    | 587            |
      | smtp_username| test           |
      | smtp_password| test           |
      | from         | test@example.com |
      | recipients   | test@example.com |
    When loading the configuration
    Then the SMTP host should be "localhost"
    And the SMTP port should be "587"
    And the from address should be "test@example.com"

  Scenario: Handle missing config file
    Given a non-existent config file
    When loading the configuration
    Then an error should be returned

  Scenario: Handle invalid config format
    Given a config file with invalid YAML
    When loading the configuration
    Then an error should be returned
