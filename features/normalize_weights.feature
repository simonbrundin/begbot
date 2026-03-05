Feature: Normalisera värderingsvikter
  Som ett system som hanterar värderingstyper
  Vill jag normalisera vikter så att aktiva typer alltid summerar till 100 %
  Så att den viktade värderingsberäkningen blir korrekt

  Scenario: Jämn fördelning bland aktiva typer
    Given 3 aktiva värderingstyper med vikten 0 var
    When jag normaliserar vikterna
    Then ska varje aktiv typ ha vikten ungefär 33.33

  Scenario: Aktiva vikter summerar alltid till 100
    Given värderingstyper:
      | type | active | weight |
      | 1    | true   | 40     |
      | 2    | true   | 60     |
      | 3    | false  | 0      |
    When jag normaliserar vikterna
    Then ska summan av aktiva vikter vara 100
    And den inaktiva typens vikt ska vara 0

  Scenario: Inaktivering av en typ omfördelar till kvarvarande aktiva typer
    Given värderingstyper:
      | type | active | weight |
      | 1    | true   | 50     |
      | 2    | false  | 50     |
    When jag normaliserar vikterna
    Then ska typ 1 ha vikten 100
    And ska typ 2 ha vikten 0

  Scenario: Återaktivering av en typ ger lika andel åt alla
    Given värderingstyper:
      | type | active | weight |
      | 1    | true   | 60     |
      | 2    | true   | 40     |
      | 3    | true   | 0      |
    When jag normaliserar vikterna
    Then ska alla aktiva typer ha lika vikt ungefär 33.33

  Scenario: Återaktivering av en av två typer ger 50/50-fördelning
    Given värderingstyper:
      | type | active | weight |
      | 1    | true   | 100    |
      | 2    | true   | 0      |
    When jag normaliserar vikterna
    Then ska typ 1 ha vikten 50
    And ska typ 2 ha vikten 50

  Scenario: Inga aktiva typer innebär att alla vikter är noll
    Given värderingstyper:
      | type | active | weight |
      | 1    | false  | 50     |
      | 2    | false  | 50     |
    When jag normaliserar vikterna
    Then ska alla typer ha vikten 0

  Scenario: En enda aktiv typ får 100 % vikt
    Given 1 aktiv värderingstyp med vikten 30
    When jag normaliserar vikterna
    Then ska den typen ha vikten 100

  Scenario: Ingående lista muteras inte
    Given värderingstyper:
      | type | active | weight |
      | 1    | true   | 30     |
      | 2    | true   | 70     |
    When jag normaliserar vikterna
    Then ska de ursprungliga indata-vikterna vara oförändrade
