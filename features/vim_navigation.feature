Feature: Vim-navigering
  Som en användare som bläddrar i listor i applikationen
  Vill jag använda vim-liknande tangentbordsnavigering (j/k-tangenter)
  Så att jag kan navigera snabbt utan att använda musen

  Scenario: Flytta markering nedåt med j-tangenten
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag trycker på j
    Then ska det valda indexet vara 0
    When jag trycker på j igen
    Then ska det valda indexet vara 1

  Scenario: Flytta markering uppåt med k-tangenten
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag trycker på k
    Then ska det valda indexet vara 4
    When jag trycker på k igen
    Then ska det valda indexet vara 3

  Scenario: Visuell markering visas när ett element är valt
    Given en lista med 3 poster
    And navigeringen är fokuserad
    When jag trycker på j
    Then ska det valda indexet inte vara null
    And ska det valda indexet vara 0

  Scenario: Stanna på första posten vid k-tryckning i toppen
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag navigerar ned till index 0
    And jag trycker på k
    Then ska det valda indexet förbli 0

  Scenario: Stanna på sista posten vid j-tryckning i botten
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag navigerar till sista posten
    And jag trycker på j igen
    Then ska det valda indexet förbli 4

  Scenario: Hantera lista med ett element
    Given en lista med 1 post
    And navigeringen är fokuserad
    When jag trycker på j
    Then ska det valda indexet vara 0
    When jag trycker på j igen
    Then ska det valda indexet vara 0
    When jag trycker på k
    Then ska det valda indexet vara 0

  Scenario: Reagerar inte när navigeringen inte är fokuserad
    Given en lista med 5 poster
    And navigeringen är inte fokuserad
    When jag trycker på j
    Then ska det valda indexet vara null
    When jag trycker på k
    Then ska det valda indexet vara null

  Scenario: Spåra fokusändringar
    Given en lista med 5 poster
    When jag ställer in fokus till true och trycker på j
    Then ska det valda indexet vara 0
    When jag ställer in fokus till false och trycker på j
    Then ska det valda indexet förbli 0
    When jag ställer in fokus till true och trycker på j
    Then ska det valda indexet vara 1

  Scenario: Hantera tom lista utan att krascha
    Given en lista med 0 poster
    And navigeringen är fokuserad
    When jag trycker på j
    Then ska inget navigeringsfel inträffa
    And ska det valda indexet vara null

  Scenario: Rensa markering med ESC
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag trycker på j för att välja index 0
    And jag rensar markeringen
    Then ska det valda indexet vara null

  Scenario: Hantera förändring av antal poster till noll
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag navigerar till index 0
    And antalet poster ändras till 0
    Then ska det valda indexet vara null

  Scenario: Hantera förändring av antal poster där valt index blir ogiltigt
    Given en lista med 5 poster
    And navigeringen är fokuserad
    When jag navigerar till index 2
    And antalet poster ändras till 2
    Then ska det valda indexet vara 1
