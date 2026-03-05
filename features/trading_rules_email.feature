Feature: E-postnotifiering för handelsregler
  Som ett handelssystem
  Vill jag skicka e-postnotifieringar när annonser uppfyller handelsreglerna
  Så att jag kan agera på lönsamma möjligheter

  Background:
    Given handelsregler med lägsta vinst på 500 SEK och lägsta rabatt på 10%

  Scenario: E-post skickas när annonsen uppfyller alla handelsregler
    Given en annons med priset 6000 SEK och värderingen 8000 SEK
    When annonsen utvärderas mot handelsreglerna
    Then ska vinsten vara 2000 SEK
    And ska rabatten vara ungefär 25%
    And annonsen ska uppfylla handelsreglerna

  Scenario: E-posten innehåller alla obligatoriska fält
    Given en annons med:
      | field       | value                          |
      | price       | 5000                           |
      | valuation   | 8000                           |
      | link        | https://blocket.se/item/123    |
      | description | Säljer iPhone 15 Pro i fint skick |
    And ett nypris på 15000 SEK
    When e-postnotifieringen förbereds
    Then ska e-posten inkludera köppriset
    And ska e-posten inkludera värderingen
    And ska e-posten inkludera rabattprocenten
    And ska e-posten inkludera nypriset
    And ska e-posten inkludera vinsten
    And ska e-posten inkludera beskrivningen
    And ska e-posten inkludera länken

  Scenario: Ingen e-post när vinsten är för låg
    Given en annons med priset 7500 SEK och värderingen 8000 SEK
    When annonsen utvärderas mot handelsreglerna
    Then ska vinsten vara 500 SEK
    And annonsen ska inte uppfylla minimivinstgränsen på 500 SEK

  Scenario: Ingen e-post när rabatten är för låg
    Given en annons med priset 7300 SEK och värderingen 8000 SEK
    When annonsen utvärderas mot handelsreglerna
    Then ska rabatten vara ungefär 8.75%
    And annonsen ska inte uppfylla minimirabattgränsen på 10%

  Scenario: E-post skickas asynkront
    Given en giltig annons som uppfyller handelsreglerna
    When e-postnotifieringen utlöses
    Then ska e-posten skickas asynkront utan att blockera

  Scenario: Hantera saknad e-postkonfiguration
    Given ingen e-postkonfiguration är inställd
    When en annons uppfyller handelsreglerna
    Then ska inget krascha
