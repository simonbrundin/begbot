Feature: Konfiguration

  Background:
    Given ett konfigurationssystem är tillgängligt

  Scenario: Ladda konfiguration med LLM-modeller definierade
    Given en konfigurationsfil med leverantören "openrouter"
    And API-nyckel "test-key"
    And webbplatsens URL "http://localhost:3000"
    And webbplatsens namn "Begbot"
    And standardmodell "deepseek/deepseek-v3.2"
    And modeller:
      | task                    | model                              |
      | ExtractProductInfo      | anthropic/claude-sonnet-4-20250514 |
      | ValidateProduct         | google/gemini-2.5-pro              |
      | CheckProductCondition   | openai/gpt-4o-mini                 |
      | EstimateNewPrice        | deepseek/deepseek-chat             |
    When konfigurationen laddas
    Then leverantören ska vara "openrouter"
    And API-nyckeln ska vara "test-key"
    And webbplatsens URL ska vara "http://localhost:3000"
    And webbplatsens namn ska vara "Begbot"
    And standardmodellen ska vara "deepseek/deepseek-v3.2"
    And det ska finnas 4 modeller definierade

  Scenario: Ladda konfiguration utan LLM-modeller
    Given en konfigurationsfil med leverantören "openrouter"
    And API-nyckel "test-key"
    And standardmodell "openai/gpt-4o"
    And inga modeller definierade
    When konfigurationen laddas
    Then antalet modeller ska vara 0

  Scenario: Ladda konfiguration med databasinställningar
    Given en konfigurationsfil med:
      | setting   | value      |
      | host      | localhost  |
      | port      | 5432       |
      | user      | test       |
      | password  | test       |
      | name      | testdb     |
      | sslmode   | require    |
    When konfigurationen laddas
    Then databashosten ska vara "localhost"
    And databasporten ska vara 5432

  Scenario: Ladda konfiguration med skrapningsinställningar
    Given en konfigurationsfil med skrapning:
      | marketplace | enabled | timeout |
      | tradera     | false   | 10s     |
      | blocket     | false   | 10s     |
    When konfigurationen laddas
    Then Tradera ska vara inaktiverat
    And Blocket ska vara inaktiverat

  Scenario: Ladda konfiguration med värderingsinställningar
    Given en konfigurationsfil med värdering:
      | setting            | value |
      | target_sell_days   | 14    |
      | min_profit_margin  | 0.15  |
      | safety_margin      | 0.2   |
    When konfigurationen laddas
    Then målet för säljdagar ska vara 14
    And minsta vinstmarginal ska vara 0.15
    And säkerhetsmarginalen ska vara 0.2

  Scenario: Ladda konfiguration med e-postinställningar
    Given en konfigurationsfil med e-post:
      | setting      | value          |
      | smtp_host    | localhost      |
      | smtp_port    | 587            |
      | smtp_username| test           |
      | smtp_password| test           |
      | from         | test@example.com |
      | recipients   | test@example.com |
    When konfigurationen laddas
    Then SMTP-hosten ska vara "localhost"
    And SMTP-porten ska vara "587"
    And avsändaradressen ska vara "test@example.com"

  Scenario: Hantera saknad konfigurationsfil
    Given en icke-existerande konfigurationsfil
    When konfigurationen laddas
    Then ett konfigurationsladdningsfel ska returneras

  Scenario: Hantera ogiltigt konfigurationsformat
    Given en konfigurationsfil med ogiltigt YAML
    When konfigurationen laddas
    Then ett konfigurationsladdningsfel ska returneras
