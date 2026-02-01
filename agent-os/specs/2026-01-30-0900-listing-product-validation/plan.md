# Plan: Listing Product Validation

## Steg 1: Clarify What We're Building

**Scope**: Ny valideringsfunktion som säkerställer att listings bara sparas om motsvarande produkt finns i `products`-tabellen.

**Förväntat resultat**: Endast annonser för produkter som finns i produktkatalogen (t.ex. "iPhone 16 Pro") ska sparas. Skal, laddare och andra tillbehör ska filtreras bort.

**Constraints**:
- Måste fungera med befintlig kodstruktur
- Ska använda befintliga databasfunktioner
- Måste vara bakåtkompatibelt

## Steg 2: Gather Visuals

Inga visuals tillgängliga.

## Steg 3: Identify Reference Implementations

- `bot.go:115-123` - `ValidateProduct()` LLM-validering
- `postgres.go:301-317` - `GetProductByName()` databasfunktion
- `postgres.go:319-338` - `GetOrCreateProduct()` pattern

## Steg 4: Check Product Context

- **Mission**: Hitta lönsamma produkter för omförsäljning
- **Phase 1 MVP**: Värdering, vinstberäkning, emailnotifikation
- Denna funktion stödjer missionen genom att säkerställa datakvalitet

## Steg 5: Surface Relevant Standards

- `backend/llm-service-functions.md` - En LLM-funktion per uppgift
- `database/currency-storage.md` - Valutahantering (SEK öre)

## Steg 6: Generate Spec Folder Name

`2026-01-30-0900-listing-product-validation/`

## Steg 7-8: Structure & Complete the Plan

### Tasks

1. ~~**Save spec documentation**~~ ✅ COMPLETED
2. ~~**Migrera databasen**~~ ✅ COMPLETED
   - Lägg till `category` och `model_variant` kolumner till `products`
   - Kolumnerna läggs till via `postgres.go:Migrate()` vid appstart
3. ~~**Uppdatera Product-modellen**~~ ✅ COMPLETED
   - `Category` och `ModelVariant` fält finns redan i `internal/models/models.go`
4. ~~**Skapa databasfunktion**~~ ✅ COMPLETED
   - `FindProduct(brand, name, category)` finns i `postgres.go:352-368`
5. ~~**Skapa LLM-funktion**~~ ✅ COMPLETED
   - `ExtractProductInfo()` returnerar redan `category` i `ProductInfo` struct (`llm.go:21-30`)
6. ~~**Skapa tjänstelager-validering**~~ ✅ COMPLETED
   - `ValidateListing()` finns i `bot.go:203-232`
   - Validerar att produkt finns i katalog med rätt category
7. ~~**Integrera validering i listings-flödet**~~ ✅ COMPLETED
   - `processAd()` använder nu `ValidateListing()` för att validera produkter
   - Endast produkter som finns i katalog med rätt category sparas
8. ~~**Uppdatera befintliga produkter**~~ ✅ COMPLETED
   - Migration sätter `category = 'phone'` för produkter med NULL (`postgres.go:151`)
9. ~~**Lägg till enhetstester**~~ ✅ COMPLETED
   - Tester i `internal/services/validate_listing_test.go`
   - Testar kategorimatchning, produktvalidering, och edge cases
10. ~~**Verifiera att allt fungerar**~~ ✅ COMPLETED
    - Alla tester passerar
    - `go vet` och `gofmt` godkända
