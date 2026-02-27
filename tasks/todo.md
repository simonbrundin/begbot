# Task: Debug and Fix Listings Scraper

**Date:** 2026-02-04
**Status:** ✅ COMPLETED

## Problem

- User reported that no ads were being saved to the listings table in the database
- Frontend was showing "fetch failed" error on /ads page

## Root Cause Analysis

1. **Scraping was working** - system found 53 ads from Blocket
2. **API was working** - /api/listings endpoint returned data
3. **Frontend was working** - page rendered correctly
4. **Issue:** Only 4 listings saved because most products were either not in catalog or not enabled
5. **Secondary issue:** `title` and `valuation` fields were not being saved when listings were created

## Changes Made

### 1. Database Schema Updates

- Added `new_price` column to `products` table
- Updated `listings` table to have non-nullable `title` and `valuation` columns

### 2. Go Backend Changes

**internal/models/models.go:**

- Changed `Listing.Title` from `*string` to `string`
- Changed `Listing.Valuation` from `*int` to `int`
- Added `Product.NewPrice` field

**internal/services/bot.go:**

- Updated `processAd()` to save `ad.Title` and `candidate.EstimatedSell`
- Added valuation collection using `CollectAll()` and `Compile()`
- Fixed foreign key issue when saving valuations to database

**internal/db/postgres.go:**

- Updated all product queries to include `new_price`
- Fixed nil pointer handling for non-nullable fields

**cmd/api/main.go:**

- Updated product queries to include `new_price`

### 3. Frontend Changes

**frontend/pages/ads.vue:**

- Fixed price formatting (removed /100 division)
- Added thousand separators using `toLocaleString("sv-SE")`
- Removed decimal places for cleaner display
- Added `new_price` display from product

### 4. Database Migrations

- Ran SQL to add `new_price` column to products table
- Updated existing listings with default values for title and valuation
- Added NOT NULL constraints

## Results

- ✅ Listings now have proper `title` and `valuation` fields populated
- ✅ Valuations use compiled values from multiple methods (LLM, Tradera, SoldAds)
- ✅ Partial valuations saved to `valuations` table
- ✅ Price formatting shows Swedish format (3 000 kr)
- ✅ Frontend displays correctly on /ads page

## Database Queries Run

```sql
ALTER TABLE products ADD COLUMN new_price INTEGER;
ALTER TABLE listings ALTER COLUMN title SET NOT NULL;
ALTER TABLE listings ALTER COLUMN valuation SET NOT NULL;
```

---

# Task: Visa delvärderingar i annonslistan

**Date:** 2026-02-04
**Status:** 📋 PENDING
**Spec:** `agent-os/specs/2026-02-04-1200-listings-valuations-visning/`

## Problem

Användaren vill se individuella värderingar (delvärderingar) för varje annons, inte bara det sammanslagna värdet.

## Lösning

Visa "X kr - Typ" för varje delvärdering direkt i listvyn.

## Tasks

### Task 1: Spara spec-dokumentation ✅

- Skapad: plan.md, shape.md, standards.md, references.md

### Task 2: Uppdatera API `/api/listings`

- [x] Modifiera `GetListings` handler för att hämta relaterade valuations
- [x] JOIN med `valuation_types` för att få typnamn
- [x] Lägg till `valuations: []` array i response
- [x] Struktur: `{ type: string, value: int }`

### Task 3: Uppdatera frontend

- [x] Lägg till visning i listings-komponenten
- [x] Format: "X kr - Typ" per valuating
- [x] Placera bredvid nuvarande värdering

### Task 4: Verifiera

- [x] Kör API och verifiera response
- [x] Verifiera frontend-visning

---

# Task: Spara scraping-loggar i databasen

**Date:** 2026-02-27
**Status:** 📋 PENDING
**Spec:** /home/simon/repos/begbot-fix-permanent-logs/environments/supabase/migrations/

## Problem

- Loggar från scraping-körningar sparas endast i stdout/stderr
- Svårt att felsöka efter specifika körningar
- Vill kunna se loggarna i efterhand (30 dagars retention)

## Lösning

Spara viktiga loggmeddelanden (errors, warnings, info) till databasen.

## Uppskattad datamängd

- ~100 loggrader/körning × 288 körningar/dag × 30 dagar = ~864,000 rader
- Ca 40MB/månad (hanterbart)

## Tasks

### Task 1: Skapa databas-tabell

- [x] Skapa `environments/supabase/migrations/016_create_scraping_run_logs.sql`
- Kolumner: id, scraping_run_id, level, message, created_at
- Index på scraping_run_id och created_at
- Även lagt till migration i `internal/db/postgres.go` för automatisk körning

### Task 2: Lägg till Go-modell

- [x] Lägg till `ScrapingRunLog` struct i `internal/models/models.go`

### Task 3: Lägg till DB-metoder

- [x] `SaveScrapingRunLogs(ctx, []*ScrapingRunLog)` - batch insert
- [x] `GetScrapingRunLogs(ctx, runID)` - hämta för en körning
- [x] `DeleteOldScrapingRunLogs(ctx, beforeDate)` - städa upp (30 dagar)

### Task 4: Modifiera bot.log()

- [x] Spara loggar till DB (errors, warnings, viktiga info)
- Buffra och batcha för prestanda (spara vid körningsslut)

### Task 5: Lägg till API-endpoint

- [x] `GET /api/scraping-runs/:id/logs` i `cmd/api/main.go`

### Task 6: Städa upp gammal data

- [x] Lägg till `DeleteOldScrapingRunLogs()` anrop i `Run()`
- Kör vid varje körning, behåll 30 dagar

### Task 7: Verifiera

- [x] Koden kompilerar
- [ ] Kör bot och verifiera att loggar sparas
- [ ] Verifiera API endpoint
- [ ] Kolla att cleanup fungerar
