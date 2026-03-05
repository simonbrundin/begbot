Feature: Söktermstjänst
  Som ett skrapningssystem
  Vill jag hantera och använda söktermer för marknadsplatssökningar
  Så att jag kan hitta relevanta annonser

  Scenario: Skapa en giltig sökterm med marknadsplats
    Given en sökterm med beskrivningen "iPhone Search" och URL:en "https://blocket.se/?q=iphone"
    And söktermen är kopplad till marknadsplats-ID 1
    Then ska beskrivningen vara "iPhone Search"
    And URL:en ska inte vara tom
    And marknadsplats-ID:et ska vara 1

  Scenario: Skapa ett sökjobb med marknadsplats
    Given en sökterm "iPhone" kopplad till Blocket-marknadsplatsen
    When ett sökjobb skapas
    Then jobbets URL ska inte vara tom
    And marknadsplatsen ska inte vara null

  Scenario: Sökjobb för Tradera-marknadsplats
    Given en sökterm "Lego Star Wars" kopplad till Tradera-marknadsplatsen
    When ett sökjobb skapas
    Then marknadsplatsens namn ska vara "Tradera"

  Scenario: Inaktiv sökterm
    Given en sökterm med isActive false
    Then ska söktermen vara inaktiv

  Scenario: Sökterm utan marknadsplats
    Given en sökterm utan marknadsplats-ID
    Then ska marknadsplats-ID:et vara null

  Scenario: Tjänsten accepterar kontext
    Given en söktermstjänst
    When en kontext skickas
    Then ska inget fel inträffa
