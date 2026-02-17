Feature: Configuration Management
  As a system administrator
  I want to load configuration from YAML files
  So that the application can be configured flexibly

  Background:
    Given a temporary directory for config files

  Scenario: Load LLM configuration with models
    Given a config file with LLM models:
      """
      provider: "openrouter"
      api_key: "test-key"
      site_url: "http://localhost:3000"
      site_name: "Begbot"
      default_model: "deepseek/deepseek-v3.2"
      models:
        ExtractProductInfo: "anthropic/claude-sonnet-4-20250514"
        ValidateProduct: "google/gemini-2.5-pro"
        CheckProductCondition: "openai/gpt-4o-mini"
        EstimateNewPrice: "deepseek/deepseek-chat"
      """
    When I load the configuration
    Then the LLM provider should be "openrouter"
    And the API key should be "test-key"
    And the site URL should be "http://localhost:3000"
    And the site name should be "Begbot"
    And the default model should be "deepseek/deepseek-v3.2"
    And the number of models should be "4"
    And the ExtractProductInfo model should be "anthropic/claude-sonnet-4-20250514"
    And the ValidateProduct model should be "google/gemini-2.5-pro"
    And the CheckProductCondition model should be "openai/gpt-4o-mini"
    And the EstimateNewPrice model should be "deepseek/deepseek-chat"

  Scenario: Load LLM configuration without models
    Given a config file with LLM but no models:
      """
      provider: "openrouter"
      api_key: "test-key"
      default_model: "openai/gpt-4o"
      """
    When I load the configuration
    Then the LLM provider should be "openrouter"
    And the API key should be "test-key"
    And the default model should be "openai/gpt-4o"
    And the number of models should be "0"
