# Shape: Frontend Swedish Localization

## Scope Decisions

### Vad inkluderas
- Alla Vue-komponenter i `pages/`
- Layout-filer i `layouts/`
- Navigation och menyer
- Formulärfält och etiketter
- Button-texter och meddelanden

### Vad exkluderas
- Databas-scheman och fältnamn
- API-responser och felmeddelanden från backend
- Externa bibliotek och deren UI

## Tekniska Beslut

### Översättningsstrategi
Direkt replacement av engelska strängar med svenska motsvarigheter.

### T.ex.
- "Inventory" → "Lager"
- "Add Item" → "Lägg till"
- "Sign In" → "Logga in"
- "Total Items" → "Totalt antal"
- "Buy Price" → "Köpris"
- "Sell Price" → "Försäljningspris"

## Kontext

Produkten är byggd för en svensktalande användare (enligt product/mission.md).
Redan ett undantag finns: "Annonser" i navigationen istället för "Ads".
