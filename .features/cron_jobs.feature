Feature: Schemalagda jobb
  Som en botanvändare
  Vill jag hantera schemalagda skrapningsjobb
  Så att jag kan automatisera när skraparen körs

  Background:
    Given en schemalagd-jobb-tjänst med mockdatabas

  Scenario: Skapa ett nytt schemalagt jobb
    When jag skapar ett schemalagt jobb med namn "Daily iPhone scan", uttryck "0 8 * * *", sökterms-ID:n "[1, 2]", och aktivt true
    Then det schemalagda jobbet ska sparas
    And det schemalagda jobbet ska ha ett ID
    And jobbnamnet ska vara "Daily iPhone scan"
    And jobbuttrycket ska vara "0 8 * * *"
    And jobbet ska vara aktivt

  Scenario: Hämta alla schemalagda jobb
    Given databasen har schemalagda jobbposter
      | id | name               | cron_expression | search_term_ids | is_active |
      | 1  | Daily iPhone scan  | 0 8 * * *       | [1,2]           | true      |
      | 2  | Hourly check        | 0 * * * *       | []              | false     |
    When jag hämtar alla schemalagda jobb
    Then ska jag få 2 schemalagda jobbposter
    And det första schemalagda jobbet ska ha namn "Daily iPhone scan"

  Scenario: Uppdatera ett schemalagt jobb
    Given databasen har schemalagda jobbposter
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Daily scan       | 0 8 * * *       | [1]             | true      |
    When jag uppdaterar schemalagt jobb 1 till namn "Twice daily scan", uttryck "0 8,20 * * *", sökterms-ID:n "[1, 2]", och aktivt false
    Then det schemalagda jobbet ska uppdateras
    And jobbnamnet ska vara "Twice daily scan"
    And jobbuttrycket ska vara "0 8,20 * * *"
    And jobbet ska vara inaktivt

  Scenario: Ta bort ett schemalagt jobb
    Given databasen har schemalagda jobbposter
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Old job          | 0 8 * * *       | [1]             | true      |
    When jag tar bort schemalagt jobb 1
    Then det schemalagda jobbet ska tas bort
    And det ska finnas 0 schemalagda jobb i databasen

  Scenario: Växla aktivt tillstånd för ett schemalagt jobb
    Given databasen har schemalagda jobbposter
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Test job         | 0 * * * *       | []              | true      |
    When jag växlar aktivt tillstånd för schemalagt jobb 1
    Then det schemalagda jobbet ska uppdateras
    And jobbet ska vara inaktivt

  Scenario: Skapa schemalagt jobb med tomma sökterms-ID:n (alla termer)
    When jag skapar ett schemalagt jobb med namn "All terms scan", uttryck "0 6 * * *", sökterms-ID:n "[]", och aktivt true
    Then det schemalagda jobbet ska sparas
    And sökterms-ID:na ska vara tomma

  Scenario: Ogiltigt cron-uttryck
    Given databasen har schemalagda jobbposter
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Test job         | 0 8 * * *       | []              | true      |
    When jag uppdaterar schemalagt jobb 1 till namn "Bad job", uttryck "invalid", sökterms-ID:n "[]", och aktivt true
    Then ett fel ska returneras
    And felmeddelandet ska innehålla "invalid cron expression"

  Scenario: Hämta schemalagt jobb via ID
    Given databasen har schemalagda jobbposter
      | id | name             | cron_expression | search_term_ids | is_active |
      | 1  | Specific job     | 0 8 * * *       | [1,2,3]         | true      |
      | 2  | Other job        | 0 * * * *       | []              | false     |
    When jag hämtar schemalagt jobb med ID 1
    Then ska jag få 1 schemalagd jobbpost
    And jobbnamnet ska vara "Specific job"
    And jobbuttrycket ska vara "0 8 * * *"

  Scenario: Cron-uttryck med specialtecken
    When jag skapar ett schemalagt jobb med namn "Weekday scan", uttryck "0 9 * * 1-5", sökterms-ID:n "[1]", och aktivt true
    Then det schemalagda jobbet ska sparas
    And jobbuttrycket ska vara "0 9 * * 1-5"
