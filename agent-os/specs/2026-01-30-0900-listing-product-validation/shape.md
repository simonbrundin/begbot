# Shape: Listing Product Validation

## Problembeskrivning

Idag sparas alla annonser som hittas utan att verifiera att produkten matchar något i produktkatalogen. Detta leder till:
- Skal, laddare och andra tillbehör sparas som listings
- Fulldata som inte kan användas för värdering
- Onödig belastning på databasen

## Beslut

### Databasschema - Nya kolumner i `products`

```sql
ALTER TABLE products ADD COLUMN category TEXT;        -- t.ex. "phone", "case", "charger", "tablet"
ALTER TABLE products ADD COLUMN model_variant TEXT;   -- t.ex. "pro", "base", "mini" (valfritt)
```

**Motivering**:
- Möjliggör kategoribaserad matchning
- Förhindrar att "skal till iPhone 16 Pro" matchar "iPhone 16 Pro"
- Samma brand+name kan ha olika category (t.ex. "iPhone 15" som phone vs "iPhone 15" som reservdel)

### Valideringsstrategi

1. LLM extraherar: `brand`, `name`, `category` från annons
2. Slå upp produkt i `products` med `brand + name + category`
3. Om matchning finns → fortsätt spara med `ProductID`
4. Om ingen matchning → hoppa över med loggning

### Matchningslogik

```go
// Exempel på produktkatalog:
products = [
    {brand: "Apple", name: "iPhone 16 Pro", category: "phone"},
    {brand: "Apple", name: "iPhone 15", category: "phone"},
]

// LLM extraherar från annons:
{brand: "Apple", name: "iPhone 16 Pro", category: "phone"} → MATCH → spara
{brand: "Apple", name: "iPhone 16 Pro", category: "case"}  → NO MATCH → hoppa över
```

## Unika Identifierare

Produkter identifieras genom **tre** fält:
- `brand` (t.ex. "Apple")
- `name` (t.ex. "iPhone 16 Pro")
- `category` (t.ex. "phone", "case", "charger", "tablet", "watch", "headphones")

## Kontext

### Ingångar
- `RawAd` från marketplace scraping (ad text, link, price)

### Utgångar
- Sparad `Listing` med `ProductID` satt (bara om produkt matchar)
- Eller hoppad annons med loggning (om ingen matchning)

### Beroenden
- `LLMService.ExtractProductInfo()` - Uppdaterad att returnera Category
- `Postgres.FindProduct()` - NY databasfunktion
- `Postgres.SaveListing()`
