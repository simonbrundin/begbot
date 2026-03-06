Feature: Authentication

  Användaren kan autentisera sig och se information därefter.

  Scenario: Lyckad inloggning
    Given användaren har ett giltigt konto med e-post och lösenord
    When användaren anger korrekt e-post och lösenord
    Then ska användaren loggas in framgångsrikt
    And användaren ska omdirigeras till startsidan

  Scenario: Fel lösenord
    Given användaren har ett giltigt konto med e-post och lösenord
    When användaren anger fel e-post eller lösenord
    Then ska inloggningen misslyckas
    And ett felmeddelande ska visas

  Scenario: Ej existerande användare
    Given ingen användare finns med den angivna e-posten
    When användaren försöker logga in med icke-existerande e-post
    Then ska inloggningen misslyckas
    And ett felmeddelande visas

  Scenario: Utloggning
    Given användaren är inloggad
    When användaren loggar ut
    Then ska användaren loggas ut
    And ska omdirigeras till inloggningssidan
