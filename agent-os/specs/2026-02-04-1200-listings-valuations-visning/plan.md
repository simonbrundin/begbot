# Plan: Visa delvärderingar i annonslistan

## Steg 1: Klargöra vad vi bygger
- **Funktion**: Visa individuella värderingar (delvärderingar) för varje annons i listvyn
- **Output**: Enkelt format - "500 kr - Egen databas" per värdering
- **Placering**: Inline i listvyn bredvid nuvarande sammanslagna värdering

## Steg 2: Samla visuella referenser
- Inga visuella tillhandahållna

## Steg 3: Identifiera referensimplementationer
- `internal/services/valuation.go` - Valuation service
- `internal/models/models.go` - Valuation modell
- `/api/listings` endpoint i cmd/api/

## Steg 4: Kontrollera produktkontext
- Mission: "Den hittar precis de produkter jag tror jag kan tjäna pengar på att köpa och sälja vidare"
- Roadmap Phase 1: Värdering av produkt, beräkning av potentiell vinst

## Steg 5: Applicera standards
- `swedish-text`: All frontend-text på svenska
- `currency-storage`: Valörer som heltal (ören)

## Steg 6: Generera spec-mapp
- `agent-os/specs/2026-02-04-1200-listings-valuations-visning/`

## Steg 7: Struktura planen

### Task 1: Spara spec-dokumentation
### Task 2: Uppdatera API `/api/listings` för att inkludera delvärderingar
### Task 3: Uppdatera frontend för att visa delvärderingar
### Task 4: Verifiera implementation

## Steg 8: Komplettera planen

### Implementation Tasks

**Task 1: Spara spec-dokumentation** ✓

**Task 2: Uppdatera API `/api/listings`**
- Modifiera `GetListings` handler för att hämta relaterade valuations
- Lägg till JOIN eller separat query för valuations per product_id
- Strukturera response med `valuations: []` array
- Varje valuating: `{ type: string, value: int }`

**Task 3: Uppdatera frontend**
- Lägg till visning i listings-komponenten
- Format: "X kr - Typ" per valuating
- Placera bredvid nuvarande "potential profit" eller "valuation"

**Task 4: Verifiera**
- Kör API och verifiera response
- Verifiera frontend-visning
