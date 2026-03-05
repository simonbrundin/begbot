Feature: LLM Produkttillståndsbedömning
  Som ett botsystem
  Vill jag bedöma om en produkt är i acceptabelt skick
  Så att jag kan avgöra om jag ska köpa den

  Scenario: Produkten är hel och utan problem
    Given ett produkttillståndsresultat med isIntact true och inga repor
    And inga problem hittades
    Then ska produkten vara giltig för köp

  Scenario: Produkten har allvarliga problem och är inte hel
    Given ett produkttillståndsresultat med isIntact false
    And problem hittades: "sprucken skärm", "vattenskada"
    Then ska produkten inte vara giltig för köp
    And det ska ha hittats 2 problem

  Scenario: Produkten har smärre repor men är fortfarande hel
    Given ett produkttillståndsresultat med isIntact true och smärre repor
    Then ska produkten vara giltig för köp
    And flaggan för smärre repor ska vara true

  Scenario: Produkten är giltig för köp när hel och utan repor
    Given en produkt med isIntact true och inga repor
    When köpgiltigheten utvärderas
    Then ska produkten vara giltig för köp

  Scenario: Produkten är giltig för köp när hel med smärre repor
    Given en produkt med isIntact true och smärre repor
    When köpgiltigheten utvärderas
    Then ska produkten vara giltig för köp

  Scenario: Produkten är inte giltig för köp när den inte är hel
    Given en produkt med isIntact false och problem "sprucken skärm"
    When köpgiltigheten utvärderas
    Then ska produkten inte vara giltig för köp
