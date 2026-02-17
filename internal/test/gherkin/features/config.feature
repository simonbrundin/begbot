Feature: Config Loading
  As a configuration system
  I want to load settings from YAML files
  So that I can configure the application

  Background:
    Given temporary config directory

  Scenario: Load LLM config with models
    Given a config file with LLM section containing:
      | field         | value                              |
      | provider      | openrouter                         |
      | api_key       | test-key                           |
      | site_url      | http://localhost:3000              |
      | site_name     | Begbot                             |
      | default_model | deepseek/deepseek-v3.2             |
    And models section with:
      | key                    | value                              |
      | ExtractProductInfo     | anthropic/claude-sonnet-4-20250514 |
      | ValidateProduct        | google/gemini-2.5-pro              |
      | CheckProductCondition | openai/gpt-4o-mini                |
      | EstimateNewPrice       | deepseek/deepseek-chat             |
    When I load the config
    Then LLM provider should be "openrouter"
    And LLM api_key should be "test-key"
    And LLM site_url should be "http://localhost:3000"
    And LLM site_name should be "Begbot"
    And LLM default_model should be "deepseek/deepseek-v3.2"
    And there should be 4 models configured

  Scenario: Load LLM config without models section
    Given a config file with LLM section containing:
      | field         | value         |
      | provider      | openrouter    |
      | api_key       | test-key      |
      | default_model | openai/gpt-4o |
    And no models section
    When I load the config
    Then models should be empty or nil
