Feature: Sidebar Icons
  As a user of the application
  I want to see icons next to each sidebar menu item
  So that I can quickly identify different sections

  Background:
    Given the Nuxt frontend application

  Scenario: @nuxt/icon module is installed
    When I check the package.json
    Then nuxt icon module should be in devDependencies

  Scenario: @nuxt/icon is configured in nuxt.config.ts
    When I check nuxt.config.ts
    Then nuxt config should include the icon module

  Scenario: Lucide icon collection is installed
    When I check the package.json
    Then lucide icon collection should be installed

  Scenario: Översikt menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:home icon

  Scenario: Produkter menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:package icon

  Scenario: Mina annonser menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:list icon

  Scenario: Transaktioner menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:arrow-left-right icon

  Scenario: Marknadsanalys menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:bar-chart icon

  Scenario: Scraping menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:spider icon

  Scenario: Historik menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:history icon

  Scenario: Hittade annonser menu item has an icon
    When I check the sidebar layout
    Then it should contain a lucide:megaphone icon

  Scenario: Icons have consistent size styling
    When I check the sidebar layout
    Then all Icon components should have consistent size styling

  Scenario: Icons have consistent color styling
    When I check the sidebar layout
    Then all Icon components should not have explicit color overrides

  Scenario: Hover styling is preserved
    When I check the sidebar layout
    Then the layout should contain "hover:bg-slate-700"

  Scenario: Active state styling is preserved
    When I check the sidebar layout
    Then the layout should contain active-class styling

  Scenario: Icons are inside NuxtLink elements
    When I check the sidebar layout
    Then there should be at least 8 NuxtLink elements with icons

  Scenario: Icons use local collection not remote
    When I check nuxt.config.ts
    Then it should not use remote server bundle
