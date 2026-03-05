Feature: Sökhistorik

  Background:
    Given en sökhistoriktjänst är tillgänglig
    And databasen är ansluten

  Scenario: Registrera en ny sökning
    When en användare söker efter "iPhone 15 Pro" med URL:en "https://blocket.se/iphone"
    And sökningen hittar 10 resultat med 3 nya annonser
    Then ska sökningen sparas
    And sökningen ska ha ett giltigt ID
    And sökbeskrivningen ska vara "iPhone 15 Pro"
    And antalet hittade resultat ska vara 10
    And antalet nya annonser ska vara 3

  Scenario: Hämta sökhistorik med data
    Given databasen har 2 sökposter
    When användaren begär sökhistorik för sida 1 med 20 poster per sida
    Then ska svaret innehålla 2 sökposter
    And det totala antalet ska vara 2
    And den första posten ska ha söktermen "iPhone 15"

  Scenario: Hämta tom sökhistorik
    Given databasen har inga sökposter
    When användaren begär sökhistorik
    Then ska svaret innehålla 0 sökposter
    And det totala antalet ska vara 0

  Scenario: Paginera sökhistorik
    Given databasen har 5 sökposter
    When användaren begär sida 1 med 2 poster per sida
    Then ska svaret innehålla 2 poster
    And det totala antalet ska vara 5
    When användaren begär sida 2 med 2 poster per sida
    Then ska svaret innehålla 2 poster
    And den första posten på sida 2 ska ha ID 3
    When användaren begär sida 3 med 2 poster per sida
    Then ska svaret innehålla 1 post

  Scenario: Hantera ogiltiga pagineringsparametrar
    Given databasen har inga sökposter
    When användaren begär sida 0
    Then ska förfrågan lyckas
    And antalet ska vara 0
    When användaren begär sida -1
    Then ska förfrågan lyckas
    And antalet ska vara 0

  Scenario: Hantera databasfel vid registrering av sökning
    Given databasen är otillgänglig
    When användaren försöker registrera en sökning
    Then ett sökhistorikfel ska returneras

  Scenario: Hantera databasfel vid hämtning av historik
    Given databasen är otillgänglig
    When användaren begär sökhistorik
    Then ett sökhistorikfel ska returneras

  Scenario: Hantera stor sidstorlek
    Given databasen har inga sökposter
    When användaren begär sida 1 med 200 poster per sida
    Then ska förfrågan lyckas
