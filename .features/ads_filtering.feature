Feature: Annonsfiltrering
  Som en användare som tittar på annonssidan
  Vill jag filtrera annonser via flikval
  Så att jag snabbt kan hitta prisvärda erbjudanden

  Background:
    Given jag har en annonsfiltreringstjänst

  Scenario: Filtret visar alla annonser i fliken "Alla"
    Given följande annonser:
      | listing_id | price | valuation |
      | 1          | 1000  | 800       |
      | 2          | 500   | 1000      |
      | 3          | 2000  | 2500      |
    When jag filtrerar på fliken "all"
    Then ska jag få 3 annonser

  Scenario: Filtret visar bara prisvärda annonser i fliken "Prisvärda"
    Given följande annonser:
      | listing_id | price | valuation |
      | 1          | 1000  | 800       |
      | 2          | 500   | 1000      |
      | 3          | 2000  | 2500      |
    When jag filtrerar på fliken "good-value"
    Then ska jag få 2 annonser

  Scenario: Inga prisvärda annonser när priset är lika med värderingen
    Given följande annonser:
      | listing_id | price | valuation |
      | 1          | 1000  | 1000      |
    When jag filtrerar på fliken "good-value"
    Then ska jag få 0 annonser

  Scenario: Inga prisvärda annonser när priset överstiger värderingen
    Given följande annonser:
      | listing_id | price | valuation |
      | 1          | 1500  | 1000      |
    When jag filtrerar på fliken "good-value"
    Then ska jag få 0 annonser

  Scenario: Annons utan värdering inkluderas inte som prisvärd
    Given följande annonser:
      | listing_id | price | valuation |
      | 1          | 500   |           |
    When jag filtrerar på fliken "good-value"
    Then ska jag få 0 annonser

  Scenario: Alla annonser är prisvärda
    Given följande annonser:
      | listing_id | price | valuation |
      | 1          | 500   | 1000      |
      | 2          | 800   | 1500      |
      | 3          | 1000  | 2000      |
    When jag filtrerar på fliken "good-value"
    Then ska jag få 3 annonser

  Scenario: Tom annonslista
    Given det finns inga annonser
    When jag filtrerar på fliken "all"
    Then ska jag få 0 annonser
    When jag filtrerar på fliken "good-value"
    Then ska jag få 0 annonser
