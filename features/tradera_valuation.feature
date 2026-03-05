Feature: Tradera-värdering
  Som ett värderingssystem
  Vill jag uppskatta produktpriser från sålda Tradera-annonser
  Så att jag kan ge korrekta köprekommendationer

  Scenario: Tolka Tradera API-svar och beräkna värdering
    Given Tradera-värdering är aktiverad
    And Tradera-API:et returnerar ett genomsnittspris på 2045 SEK med 901 sålda artiklar
    When jag värderar en produkt på Tradera
    Then värderingsvärdet ska vara 2045
    And förtroendet ska vara minst 0.84 för 901 artiklar

  Scenario: Returnera fel när inga priser hittas
    Given Tradera-värdering är aktiverad
    And Tradera-API:et returnerar noll genomsnittspris och noll artiklar
    When jag värderar en produkt på Tradera
    Then ett Tradera-fel ska returneras
    And Tradera-värderingen ska vara null

  Scenario: Returnera null när Tradera är inaktiverat
    Given Tradera-värdering är inaktiverad
    When jag värderar en produkt på Tradera
    Then Tradera-värderingen ska vara null
    And inget Tradera-fel ska returneras
