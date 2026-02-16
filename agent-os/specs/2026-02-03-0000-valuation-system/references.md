# References: Valuation System

## Intern kod att referera till

### Databas-lager
**Fil:** `internal/db/postgres.go:527`
**Typ:** `GetAllListings()` funktion
**Användning:** Mönster för SQL-frågor och row-scanning

**Fil:** `internal/db/postgres.go`
**Typ:** `SaveListing()`, `UpdateListingStatus()`
**Användning:** Mönster för CRUD-operationer

### Models
**Fil:** `internal/models/models.go:45`
**Typ:** `Listing` struct
**Användning:** Mönster för struct-definition med JSON/db-tags

**Fil:** `internal/models/models.go:7`
**Typ:** `Product` struct
**Användning:** Mönster för relaterad entitet

### API-handlers
**Fil:** `cmd/api/main.go:165`
**Typ:** `listingsHandler`
**Användning:** Mönster för HTTP handler med GET/POST

**Fil:** `cmd/api/main.go:146`
**Typ:** `getListings()`
**Användning:** Mönster för att hämta data och returnera JSON

### Frontend
**Fil:** `frontend/pages/ads.vue`
**Typ:** `/ads` vy
**Användning:** Här ska valuation visas

**Fil:** `frontend/types/database.ts:35`
**Typ:** `Listing` interface
**Användning:** Ska uppdateras med valuation-fält

## Externa resurser

### Tradera Valuation
**URL:** https://www.tradera.com/valuation
**Användning:** Referens för värderingsdata att hämta

### LLM-integration
**Fil:** Sök efter `llm` i kodbasen
**Användning:** Existerande LLM-service att återanvända för nypris-generering

## Kod att studera

1. Hur `formatCurrency()` fungerar i ads.vue
2. Hur `statusClass()` används för badge-styling
3. Felhantering i `useAsyncData` för listings
