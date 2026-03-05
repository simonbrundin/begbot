Feature: Ta bort annons
  Som en användare av annonssidan
  Vill jag kunna ta bort annonser
  Så att jag kan radera oönskade eller felaktiga annonser

  Background:
    Given jag har en annonstabas

  Scenario: Ta bort en befintlig annons
    Given en annons med id "1" finns i databasen
    When jag skickar en DELETE-förfrågan till "/api/listings/1"
    Then ska svarsstatusen vara 204
    And annonsen med id "1" ska inte längre finnas i databasen

  Scenario: Försök ta bort en icke-existerande annons
    Given ingen annons med id "999" finns i databasen
    When jag skickar en DELETE-förfrågan till "/api/listings/999"
    Then ska svarsstatusen vara 404

  Scenario: Ta bort annons med värderingar
    Given en annons med id "2" finns i databasen
    And annonsen har värderingar
    When jag skickar en DELETE-förfrågan till "/api/listings/2"
    Then ska svarsstatusen vara 204
    And annonsen med id "2" ska inte längre finnas i databasen
    And tillhörande värderingar ska också tas bort

  Scenario: Ta bort annons med handlade varor
    Given en annons med id "3" finns i databasen
    And annonsen har handlade varor
    When jag skickar en DELETE-förfrågan till "/api/listings/3"
    Then ska svarsstatusen vara 204
    And annonsen med id "3" ska inte längre finnas i databasen
    And tillhörande handlade varor ska också tas bort

  Scenario: Ogiltigt annons-ID-format
    Given jag har en annonstabas
    When jag skickar en DELETE-förfrågan till "/api/listings/invalid"
    Then ska svarsstatusen vara 400

  Scenario: Ta bort annons med bildlänkar
    Given en annons med id "4" finns i databasen
    And annonsen har bildlänkar
    When jag skickar en DELETE-förfrågan till "/api/listings/4"
    Then ska svarsstatusen vara 204
    And annonsen med id "4" ska inte längre finnas i databasen
    And tillhörande bildlänkar ska också tas bort
