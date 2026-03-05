Feature: Värdering

  Background:
    Given en värderingskompilator är tillgänglig

  Scenario: Beräkna viktat genomsnitt med flera indata
    Given följande värderingsindata:
      | type     | value | confidence |
      | Method1  | 1000  | 0.8        |
      | Method2  | 1200  | 0.6        |
      | Method3  | 1100  | 0.7        |
    When kompilatorn beräknar det viktade genomsnittet
    Then ska det rekommenderade priset vara mellan 1000 och 1200
    And förtroendet ska vara mellan 0.6 och 0.8

  Scenario: Beräkna viktat genomsnitt med ett enda indata
    Given ett enda värderingsindata med värdet 1500 och förtroendet 0.9
    When kompilatorn beräknar det viktade genomsnittet
    Then ska det rekommenderade priset vara 1500
    And förtroendet ska vara 0.9

  Scenario: Hantera tomma indata
    Given inga värderingsindata
    When kompilatorn beräknar det viktade genomsnittet
    Then ska det rekommenderade priset vara 0
    And förtroendet ska vara 0

  Scenario: Beräkna historiskt pris för dagar
    Given en historisk värdering med K-värde -10 och interceptet 1500
    When priset för 0 dagar beräknas
    Then ska priset vara 1500
    When priset för 7 dagar beräknas
    Then ska priset vara 1430
    When priset för 30 dagar beräknas
    Then ska priset vara 1200

  Scenario: Hantera historisk värdering utan data
    Given en historisk värdering utan data
    When priset för 7 dagar beräknas
    Then ska priset vara 0

  Scenario: Beräkna vinst
    Given ett inköpspris på 500 SEK
    And fraktkostnad på 50 SEK
    And uppskattat säljpris på 1000 SEK
    When vinsten beräknas
    Then ska beräknad vinst vara 450 SEK

  Scenario: Beräkna vinstmarginal
    Given en vinst på 450 SEK
    And en totalkostnad på 550 SEK
    When vinstmarginalen beräknas
    Then ska marginalen vara ungefär 0.818

  Scenario: Hantera nollkostnad för vinstmarginal
    Given en vinst på 100 SEK
    And en totalkostnad på 0 SEK
    When vinstmarginalen beräknas
    Then ska marginalen vara 0

  Scenario: Uppskatta säljsannolikhet med negativt K-värde
    Given K-värdet är -10 (priset sjunker med tid)
    When säljsannolikheten uppskattas för 7 dagar med mål 30 dagar
    Then ska sannolikheten vara 0.95
    When säljsannolikheten uppskattas för 30 dagar med mål 30 dagar
    Then ska sannolikheten vara 0.5

  Scenario: Uppskatta säljsannolikhet med positivt K-värde
    Given K-värdet är 10 (priset stiger med tid)
    When säljsannolikheten uppskattas för 7 dagar med mål 30 dagar
    Then ska sannolikheten vara 0.1
    When säljsannolikheten uppskattas för 30 dagar med mål 30 dagar
    Then ska sannolikheten vara 0.5

  Scenario Outline: Prioritet för databasvärderingsmetod
    Given en databasvärderingsmetod
    When metodnamnet hämtas
    Then ska namnet vara "Egen databas"
    And prioriteten ska vara <priority>

    Examples:
      | priority |
      | 1        |

  Scenario Outline: LLM nyprismetod
    Given en LLM-nyprismetod
    When metodnamnet hämtas
    Then ska namnet vara "Nypris (LLM)"
    And prioriteten ska vara <priority>

    Examples:
      | priority |
      | 2        |

  Scenario Outline: Tradera-värderingsmetod
    Given en Tradera-värderingsmetod
    When metodnamnet hämtas
    Then ska namnet vara "Tradera"
    And prioriteten ska vara <priority>

    Examples:
      | priority |
      | 3        |

  Scenario Outline: Metod för sålda annonser
    Given en metod för sålda annonser
    When metodnamnet hämtas
    Then ska namnet vara "eBay/Marknadsplatser"
    And prioriteten ska vara <priority>

    Examples:
      | priority |
      | 4        |

  Scenario: Beräkna förtroende med inga artiklar
    Given en databasvärderingsmetod med 0 sålda artiklar
    When förtroendet beräknas
    Then ska förtroendet vara 0

  Scenario: Beräkna förtroende med 2 artiklar
    Given en databasvärderingsmetod med 2 sålda artiklar
    When förtroendet beräknas
    Then ska förtroendet vara 0.3

  Scenario: Beräkna förtroende med 4 artiklar
    Given en databasvärderingsmetod med 4 sålda artiklar
    When förtroendet beräknas
    Then ska förtroendet vara 0.5

  Scenario: Beräkna förtroende med 8 artiklar
    Given en databasvärderingsmetod med 8 sålda artiklar
    When förtroendet beräknas
    Then ska förtroendet vara 0.7

  Scenario: Priset ska vara i SEK och inte i öre
    Given sålda artiklar med priserna 100 SEK, 150 SEK och 125 SEK
    When det uppskattade priset beräknas
    Then ska priset vara i SEK och inte i öre

  Scenario: Kompilera utan giltiga indata
    Given värderingsindata med nollvärde eller nollförtroende
    When värderingen kompileras
    Then ska det rekommenderade priset vara 0

  Scenario: Validera värderingsgränser - normalt fall
    Given värderingsindata med värdet 1500 och förtroendet 0.7
    And nypriset är 2000
    When det viktade genomsnittet kompileras
    Then ska inget värderingsfel inträffa

  Scenario: Validera värderingsgränser - orimligt fall
    Given ett värderingsindata med värdet 150000 och förtroendet 0.7
    And nypriset är 2000
    When det viktade genomsnittet kompileras
    Then ska en varning loggas för orimlig värdering
