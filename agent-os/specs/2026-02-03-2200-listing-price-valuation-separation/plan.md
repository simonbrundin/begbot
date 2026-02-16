# Plan: Separera price och valuation i listings

## Steg 1: Save spec documentation (denna fil)

## Steg 2: Uppdatera Go-modeller och databaslager

### 2.1: Fixa GetAllListings query
Lägg till `valuation` i SELECT-satsen:
```sql
SELECT id, product_id, price, valuation, link, ...
```

### 2.2: Fixa GetListingByProductID query
Lägg till `valuation` i SELECT-satsen.

### 2.3: Uppdatera Scan-logik
Se till att `valuation` scannas in i `models.Listing`.

## Steg 3: Uppdatera frontend types

### 3.1: Lägg till valuation i Listing interfacet
```typescript
export interface Listing {
  // ... befintliga fält
  valuation: number | null  // <-- LÄGG TILL
  eligible_for_shipping: boolean | null  // <-- LÄGG TILL (saknas idag)
  seller_pays_shipping: boolean | null    // <-- LÄGG TILL (saknas idag)
  buy_now: boolean | null                // <-- LÄGG TILL (saknas idag)
}
```

## Steg 4: Uppdatera frontend UI (valfritt)

Lägg till visning av `valuation` i listings.vue om användaren vill se båda.

## Steg 5: Verifiera

1. Kör backend: `go run cmd/api/main.go`
2. Kör frontend: `npm run dev`
3. Verifiera att valuation sparas och läses korrekt
