Feature: Marknadsplats

  Background:
    Given en marknadsplatstjänst är tillgänglig
    And konfigurationen har Blocket aktiverat

  Scenario: Extrahera giltigt Blocket annons-ID från URL
    Given URL:en "https://www.blocket.se/annons/123456"
    When annons-ID:et extraheras
    Then ska annons-ID:et vara 123456

  Scenario: Extrahera Blocket annons-ID från URL med frågeparametrar
    Given URL:en "https://www.blocket.se/annons/123456?q=test"
    When annons-ID:et extraheras
    Then ska annons-ID:et vara 123456

  Scenario: Extrahera giltigt Blocket annons-ID från alternativt URL-format
    Given URL:en "https://www.blocket.se/item/999999"
    When annons-ID:et extraheras
    Then ska annons-ID:et vara 999999

  Scenario: Hantera ogiltig URL
    Given en ogiltig URL "invalid"
    When annons-ID:et extraheras
    Then ska annons-ID:et vara 0

  Scenario: Hantera URL som inte är från Blocket
    Given en icke-Blocket URL "https://www.blocket.se/other/123"
    When annons-ID:et extraheras
    Then ska annons-ID:et vara 0

  Scenario: Hastighetsbegränsning mellan förfrågningar
    Given hastighetsbegränsaren nollställs
    When 5 förfrågningar görs i följd
    Then ska förfrågningarna ta minst 1 sekund
    And inga hastighetsbegränsningsfel ska inträffa

  Scenario: Hämta Blocket-annons från API med giltigt ID
    Given ett giltigt Blocket annons-ID
    When annonsen hämtas från API:et
    Then ska svaret innehålla en rubrik
    And ska svaret innehålla annonstext
    And priset ska vara större än 0

  Scenario: Hämta Blocket-annons från API med ogiltigt ID
    Given ett ogiltigt Blocket annons-ID 999999999
    When annonsen hämtas från API:et
    Then kan ett fel returneras för ogiltiga ID:n

  Scenario: Hantera hastighetsbegränsningsfel på ett smidigt sätt
    Given API:et returnerar ett hastighetsbegränsningsfel
    When förfrågan försöks igen
    Then ska omförsöket lyckas
