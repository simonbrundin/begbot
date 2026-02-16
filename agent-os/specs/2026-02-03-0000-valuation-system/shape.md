# Shaping: Valuation System

## Scope

### In Scope
- Databas-schema för valuations och valuation_types
- API för CRUD-operationer på värderingar
- Valuation collection från 4 källor:
  1. Egen databas (sålda varor analys)
  2. Tradera valuation tool
  3. Marketplace/eBay sold listings
  4. LLM-genererat nypris
- LLM compilation service (sammanställer till pris + säkerhetsprocent)
- Integration i /ads vy (visar produktnamn, frakt, pris, värdering)

### Out of Scope
- Automatisk inköpsprocess baserat på värdering
- Real-time scraping av externa källor (behöver external tools)
- Historik-grafer i UI (kan läggas till senare)
- Caching av LLM-svar (version 2)

## Designbeslut

### 1. Databas-design
```
valuation_types:
  - id (PK)
  - name (e.g. "Egen databas", "Tradera", "eBay", "Nypris (LLM)")

valuations:
  - id (PK)
  - product_id (FK)
  - valuation_type_id (FK)
  - valuation (int, SEK öre)
  - metadata (JSON, kan innehålla extra data som source_url, days_to_sell)
  - created_at (TIMESTAMP)
```

### 2. API-design
```
GET /api/valuation-types
  - Returnerar alla valuation types

GET /api/valuations?product_id={id}
  - Returnerar alla värderingar för en produkt

POST /api/valuations
  - Skapa ny värdering
  - Body: { product_id, valuation_type_id, valuation, metadata? }
```

### 3. LLM Compilation Prompt
```
Given these valuations for {product_name}:
- Database analysis: {valuation} SEK (based on {n} sold items)
- Tradera valuation: {valuation} SEK
- eBay sold: {valuation} SEK
- New price (LLM): {valuation} SEK

Suggest:
1. Recommended selling price
2. Safety margin percentage (0-100)
3. Brief reasoning
```

### 4. UI-exempel för /ads
```
┌─────────────────────────────────────┐
│ [Produktnamn]                       │
│ 1 500 kr                            │ ← Price
│ Frakt: 59 kr                        │ ← Shipping (nytt)
│ Värdering: 1 350 kr (säkerhet 15%)  │ ← Valuation (nytt)
│ [Visa] →                           │
└─────────────────────────────────────┘
```

## Öppna frågor
1. Ska vi använda SQL-funktion för databas-analys eller Go-kod?
2. Behövs transaktionshantering vid multipla valuation-källor?
3. Ska valuation kopplas till TradedItem eller Product?
