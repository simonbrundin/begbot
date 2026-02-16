# Shaping Notes: Listing Price/Valuation Separation

## Problem

Användaren rapporterar att "den sammanslagna värderingen sparas i price". Efter undersökning:

1. **Databasen har rätt struktur**: `listings`-tabellen har både `price` och `valuation` kolumner
2. **Go-modellen är korrekt**: `models.Listing` har båda fält
3. **Sparande fungerar**: `SaveListing` sparar båda kolumnerna
4. **Läsning är trasig**: `GetAllListings` och `GetListingByProductID` select:ar INTE `valuation`

## Root Cause

Två trasiga queries i `internal/db/postgres.go`:

- `GetAllListings` (rad 586): SELECT saknar `valuation`
- `GetListingByProductID` (rad 338): SELECT saknar `valuation`

## Scope

### Inkluderat
- Fixa databasqueries för att läsa `valuation`
- Uppdatera frontend TypeScript types
- Verifiera att data flödar korrekt

### Exkluderat (tillsvidare)
- Ändra valuating-beräkningen (redan korrekt i bot.go)
- Ny UI för att visa valuation (kan läggas till separat)

## Beslut

1. **Fokusera på minimal impact**: Endast läsa in valuation, inte ändra affärslogik
2. **Frontend type-first**: Uppdatera TypeScript innan UI
