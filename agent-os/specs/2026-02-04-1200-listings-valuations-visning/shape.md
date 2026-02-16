# Shaping Notes

## Scope
Visa alla delvärderingar för en produkt direkt i annonslistvyn.

## Beslut

### 1. API-response struktur
```json
{
  "id": 123,
  "title": "Produktnamn",
  "price": 300,
  "valuation": 500,
  "valuations": [
    { "type": "Egen databas", "value": 500 },
    { "type": "Nypris (LLM)", "value": 800 }
  ]
}
```

### 2. Query-strategi
Valuation-typenames behöver joinas från `valuation_types` tabellen.

### 3. Bakåtkompabilitet
Lägg till `valuations` fält som nyckel - befintliga klienter fortsätter fungera.

## Kontext
- Listings har redan en `product_id` koppling
- Valuations kopplas till `products.id`
- Finns 4 valuation types: Egen databas, Tradera, eBay, Nypris (LLM)

## Constraints
- EFTERNOM: Valuation-data kan vara tomt om inga värderingar finns
- Sortering: Visa högst 3-4 värderingar om många finns
