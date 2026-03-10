Feature: JSON-serialisering av produktmodell
  Som ett frontend som konsumerar API:et
  Vill jag att produktdata serialiseras korrekt till JSON
  Så att jag kan visa det utan fel

  Scenario: Produkt med null i created_at ska utelämna fältet från JSON
    Given en produkt med null i created_at
    When jag serialiserar produkten till JSON
    Then ska JSON inte innehålla "created_at"

  Scenario: Produkt med giltigt created_at ska inkludera det i JSON
    Given en produkt med created_at "2024-01-15T10:30:00Z"
    When jag serialiserar produkten till JSON
    Then ska JSON innehålla "created_at"

  Scenario: Tomt varumärke och namn ska bevaras i JSON
    Given en produkt med tomt varumärke och namn
    When jag serialiserar produkten till JSON
    Then ska JSON innehålla varumärket som tom sträng
    And ska JSON innehålla namnet som tom sträng

  Scenario: Fältet enabled ska vara booleskt true eller false i JSON
    Given en produkt med enabled satt till true
    When jag serialiserar produkten till JSON
    Then ska JSON innehålla "enabled":true

  Scenario: Produkt med alla null-fält ska inte ha nolltid i JSON
    Given en produkt med alla null-fält
    When jag serialiserar produkten till JSON
    Then ska JSON inte innehålla nolltiden "0001-01-01T00:00:00Z"
