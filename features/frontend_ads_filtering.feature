Feature: Annonsfiltrering i frontend
  Som en användare som tittar på annonssidan
  Vill jag filtrera annonser via flik
  Så att jag kan se alla annonser eller enbart de med potential vinst

  Scenario: Visa alla annonser i fliken "Alla"
    Given frontend-annonser finns med varierande priser och värderingar
    When jag applicerar frontend-filtret "all"
    Then ska alla frontend-annonser returneras

  Scenario: Tomt resultat när inga annonser finns
    Given inga frontend-annonser finns
    When jag applicerar frontend-filtret "all"
    Then ska 0 frontend-annonser returneras

  Scenario: Visa bara lönsamma annonser i fliken "Potentiella"
    Given frontend-annonser med vinstdata:
      | id | potentialProfit | discountPercent |
      | 1  | 200             | 20              |
      | 2  | -100            | -10             |
    When jag applicerar frontend-filtret "potential"
    Then ska 1 frontend-annons returneras
    And den returnerade frontend-annonsen ska ha id 1

  Scenario: Returnera tomt när inga annonser har positiv potential vinst
    Given en frontend-annons med id 1 och ingen vinstdata
    When jag applicerar frontend-filtret "potential"
    Then ska 0 frontend-annonser returneras
