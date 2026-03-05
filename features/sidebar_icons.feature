Feature: Ikoner i sidofältet
  Som en användare av applikationen
  Vill jag se ikoner bredvid varje menypost i sidofältet
  Så att jag snabbt kan identifiera olika sektioner

  Background:
    Given Nuxt-frontendapplikationen

  Scenario: Modulen @nuxt/icon är installerad
    When jag kontrollerar package.json
    Then Nuxt-ikonmodulen ska finnas i devDependencies

  Scenario: @nuxt/icon är konfigurerat i nuxt.config.ts
    When jag kontrollerar nuxt.config.ts
    Then Nuxt-konfigurationen ska inkludera ikonmodulen

  Scenario: Lucide-ikonsamlingen är installerad
    When jag kontrollerar package.json
    Then Lucide-ikonsamlingen ska vara installerad

  Scenario: Menyposten Översikt har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:home-ikon

  Scenario: Menyposten Produkter har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:package-ikon

  Scenario: Menyposten Mina annonser har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:list-ikon

  Scenario: Menyposten Transaktioner har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:arrow-left-right-ikon

  Scenario: Menyposten Marknadsanalys har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:bar-chart-ikon

  Scenario: Menyposten Scraping har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:spider-ikon

  Scenario: Menyposten Historik har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:history-ikon

  Scenario: Menyposten Hittade annonser har en ikon
    When jag kontrollerar sidofältslayouten
    Then ska den innehålla en lucide:megaphone-ikon

  Scenario: Ikoner har konsekvent storleksstil
    When jag kontrollerar sidofältslayouten
    Then ska alla ikonkomponenter ha konsekvent storleksstil

  Scenario: Ikoner har konsekvent färgstil
    When jag kontrollerar sidofältslayouten
    Then ska inga ikonkomponenter ha explicita färgöverskridningar

  Scenario: Hover-stilsättning bevaras
    When jag kontrollerar sidofältslayouten
    Then ska layouten innehålla "hover:bg-slate-700"

  Scenario: Aktivt-tillstånd-stilsättning bevaras
    When jag kontrollerar sidofältslayouten
    Then ska layouten innehålla active-class-stilsättning

  Scenario: Ikoner finns inuti NuxtLink-element
    When jag kontrollerar sidofältslayouten
    Then ska det finnas minst 8 NuxtLink-element med ikoner

  Scenario: Ikoner använder lokal samling och inte fjärrserver
    When jag kontrollerar nuxt.config.ts
    Then ska konfigurationen inte använda fjärr-server-bundle
