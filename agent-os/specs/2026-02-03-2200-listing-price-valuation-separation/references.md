# Reference Implementations

## Liknande mönster i kodbasen

### Valuation-systemet
- **Fil**: `internal/services/bot.go` (rad 188-261)
- **Beskrivning**: Visar hur `listing.Valuation` sätts till `compiledValuation.RecommendedPrice`
- **Referens för**: Förstå hur valuation beräknas och sparas

### Listing save/load mönster
- **Fil**: `internal/db/postgres.go`
  - `SaveListing` (rad 302): INSERT med både price och valuation
  - `GetAllListings` (rad 586): SELECT som behöver uppdateras
  - `GetListingByProductID` (rad 338): SELECT som behöver uppdateras

### Frontend typ-mönster
- **Fil**: `frontend/types/database.ts`
- **Referens för**: Hur Listing-interface är definierat

### Existerande spec
- **Fil**: `agent-os/specs/2026-02-03-0000-valuation-system/`
- **Relevans**: Innehåller bakgrund om valuingsystemet
