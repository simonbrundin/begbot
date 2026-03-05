Feature: Blocket-värdering
  Som ett värderingssystem
  Vill jag uppskatta produktpriser från Blocket-annonser
  Så att jag kan ge korrekta köprekommendationer

  Scenario: Tolka API-svar och beräkna värdering
    Given Blocket-värdering är aktiverad
    And Blocket-API:et returnerar 12 annonser med priser mellan 1000 och 2200 SEK
    When jag värderar en produkt på Blocket
    Then värderingen ska ha ett positivt värde
    And förtroendet ska vara minst 0.5 för 10 eller fler artiklar

  Scenario: Filtrera extremvärden med IQR-metoden
    Given Blocket-värdering är aktiverad
    And Blocket-API:et returnerar annonser med extremvärden:
      | listing | price |
      | 1       | 1000  |
      | 2       | 1100  |
      | 3       | 1200  |
      | 4       | 1300  |
      | 5       | 100   |
      | 6       | 10000 |
      | 7       | 1250  |
      | 8       | 1150  |
    When jag värderar en produkt på Blocket
    Then värderingen ska vara mindre än 2000
    And extrempriser ska ha filtrerats bort

  Scenario: Returnera null när Blocket är inaktiverat
    Given Blocket-värdering är inaktiverad
    When jag värderar en produkt på Blocket
    Then Blocket-värderingen ska vara null
    And inget Blocket-fel ska returneras

  Scenario: Returnera fel när inga priser hittas
    Given Blocket-värdering är aktiverad
    And Blocket-API:et returnerar inga annonser
    When jag värderar en produkt på Blocket
    Then ett Blocket-fel ska returneras
    And Blocket-värderingen ska vara null

  Scenario: Cacha resultat för att undvika duplicerade API-anrop
    Given Blocket-värdering är aktiverad
    And Blocket-API:et är tillgängligt
    When jag värderar samma produkt två gånger
    Then ska bara ett API-anrop göras
    And båda värderingarna ska ha samma värde

  Scenario: Värdera med enbart modellnamn (ingen tillverkare)
    Given Blocket-värdering är aktiverad
    And Blocket-API:et returnerar annonser för modellen
    When jag värderar en produkt med enbart ett modellnamn
    Then värderingen ska ha ett positivt värde

  Scenario: Beräkna kvartiler för prisstatistik
    Given priser: 100, 200, 300, 400, 500, 600, 700, 800
    When jag beräknar kvartilerna
    Then Q1 ska vara 200
    And Q3 ska vara 600
    And IQR ska vara 400

  Scenario: Filtrera extremvärden från en prislista
    Given priser med extremvärde: 100, 200, 300, 400, 500, 10000
    When jag filtrerar extremvärden med IQR
    Then ska 5 priser återstå

  Scenario: Beräkna median för priser
    Given en prislista med ojämnt antal element: 1, 2, 3, 4, 5
    When jag beräknar medianen
    Then medianen ska vara 3

  Scenario: Beräkna median för lista med jämnt antal element
    Given en prislista med jämnt antal element: 1, 2, 3, 4
    When jag beräknar medianen
    Then medianen ska vara 2
