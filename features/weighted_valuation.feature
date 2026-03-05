Feature: Viktad värderingsberäkning
  Som ett produktvärderingssystem
  Vill jag beräkna ett viktat genomsnitt av värderingar från flera källor
  Så att jag kan rekommendera det bästa köppriset

  Scenario: Typ är aktiv när inga konfigurationer finns (bakåtkompatibel)
    Given inga värderingskonfigurationer finns för produkt 1
    When jag kontrollerar om typ 1 är aktiv för produkt 1
    Then ska den vara aktiv

  Scenario: Typ är aktiv när konfigurationslistan är tom för produkten
    Given en tom konfigurationslista för produkt 1
    When jag kontrollerar om typ 1 är aktiv för produkt 1
    Then ska den vara aktiv

  Scenario: Typ är aktiv när ingen konfiguration finns för den specifika typen
    Given en konfiguration som bara inaktiverar typ 2 för produkt 1
    When jag kontrollerar om typ 1 är aktiv för produkt 1
    Then ska den vara aktiv

  Scenario: Typ är inaktiv när den är inaktiverad i konfigurationen
    Given en konfiguration som inaktiverar typ 1 för produkt 1
    When jag kontrollerar om typ 1 är aktiv för produkt 1
    Then ska den vara inaktiv

  Scenario: Typ är aktiv när den är explicit aktiverad i konfigurationen
    Given en konfiguration som aktiverar typ 1 för produkt 1
    When jag kontrollerar om typ 1 är aktiv för produkt 1
    Then ska den vara aktiv

  Scenario: Returnera null när inga aktiverade värderingstyper finns
    Given inga aktiverade värderingstyper
    When jag beräknar viktad värdering för produkt 1
    Then ska resultatet vara null

  Scenario: Returnera null när produkten saknar värderingar
    Given aktiverade typer: 1, 2
    And produkt 1 saknar värderingar
    When jag beräknar viktad värdering för produkt 1
    Then ska resultatet vara null

  Scenario: Beräkna korrekt genomsnitt med lika vikter
    Given aktiverade typer: 1, 2
    And produkt 1 har värderingen 1000 för typ 1 och 2000 för typ 2
    And båda typerna har vikten 1
    When jag beräknar viktad värdering för produkt 1
    Then ska genomsnittet vara 1500

  Scenario: Beräkna korrekt genomsnitt med anpassade vikter
    Given aktiverade typer: 1, 2
    And produkt 1 har värderingen 1000 för typ 1 och 3000 för typ 2
    And typ 1 har vikten 1 och typ 2 har vikten 3
    When jag beräknar viktad värdering för produkt 1
    Then ska genomsnittet vara 2500

  Scenario: Säkerhetsnivån är 100% när bara en värdering finns
    Given aktiverade typer: 1, 2
    And produkt 1 har bara värderingen 1000 för typ 1
    When jag beräknar viktad värdering för produkt 1
    Then ska säkerhetsprocenten vara 100

  Scenario: Säkerhetsnivån är lägre när värdena skiljer sig avsevärt
    Given aktiverade typer: 1, 2
    And produkt 1 har värderingen 100 för typ 1 och 900 för typ 2
    And båda typerna har vikten 1
    When jag beräknar viktad värdering för produkt 1
    Then ska säkerhetsprocenten vara lägre än 100

  Scenario: Returnera null när den totala vikten är noll
    Given aktiverade typer: 1
    And produkt 1 har värderingen 1000 för typ 1 med vikten 0
    When jag beräknar viktad värdering för produkt 1
    Then ska resultatet vara null

  Scenario: Ignorera typer utan värdering för produkten
    Given aktiverade typer: 1, 2
    And produkt 1 har bara värderingen 2000 för typ 1
    And typ 1 har vikten 1 och typ 2 har vikten 1
    When jag beräknar viktad värdering för produkt 1
    Then ska genomsnittet vara 2000
    And ska säkerhetsprocenten vara 100

  Scenario: Exkludera inaktiverade typer från viktat genomsnitt
    Given aktiverade typer: 1, 2
    And produkt 1 har värderingen 1000 för typ 1 och 3000 för typ 2
    And typ 2 är inaktiverad för produkt 1
    And båda typerna har vikten 1
    When jag beräknar viktad värdering för produkt 1
    Then ska genomsnittet vara 1000

  Scenario: Returnera null när alla typer är inaktiverade för produkten
    Given aktiverade typer: 1, 2
    And produkt 1 har värderingen 1000 för typ 1 och 2000 för typ 2
    And både typ 1 och typ 2 är inaktiverade för produkt 1
    When jag beräknar viktad värdering för produkt 1
    Then ska resultatet vara null

  Scenario: Använd alla typer när ingen konfiguration finns (bakåtkompatibel)
    Given aktiverade typer: 1, 2
    And produkt 1 har värderingen 1000 för typ 1 och 2000 för typ 2
    And inga konfigurationer finns
    When jag beräknar viktad värdering för produkt 1
    Then ska genomsnittet vara 1500

  Scenario: Sista aktiva typen får full vikt när andra är inaktiverade
    Given aktiverade typer: 1, 2, 3
    And produkt 1 har värderingen 1000 för typ 1, 2000 för typ 2, 3000 för typ 3
    And typ 1 och typ 2 är inaktiverade för produkt 1
    When jag beräknar viktad värdering för produkt 1
    Then ska genomsnittet vara 3000
