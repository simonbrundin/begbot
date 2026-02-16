# Shape: API Consolidation Review

## Status
Granskning klar - Alla endpoints implementerade ✓

## Arkitekturbeslut

**Endast Go API** — Denna applikation använder ett enda backend-API i Go. Inga Nuxt server routes eller proxy-endpoints.

### Motivering
- Enkelhet: Endast en kodbas för backend-logik
- Prestanda: Go är optimerat för hög throughput
- Underhåll: Gemensamt språk (Go) för hela stacken

### Verifiering
- `frontend/server/api/` — Tom katalog, inga Nuxt server routes
- Alla frontend-anrop går till Go API (`localhost:8081`)
- `nuxt.config.ts` pekar på `apiBase: 'http://localhost:8081'`

## Slutsats

**Inga Nuxt server routes finns att migrera.** Frontend använder redan Go API som enda backend.

## Implementerade ändringar

### Nya POST-handlers i cmd/api/main.go
- POST `/api/inventory` ✓
- POST `/api/listings` ✓  
- POST `/api/products` ✓
- POST `/api/transactions` ✓
- POST `/api/search-terms` ✓

### Query param-stöd
- GET `/api/listings?mine=true` ✓ (filtrerar på is_my_listing)

### Nya databasmetoder i internal/db/postgres.go
- `DeleteSearchTerm()` ✓

### Testresultat
- POST inventory: ✓ Skapar ny post
- POST listings: ✓ Skapar ny post
- POST products: ✓ Skapar ny post
- POST search-terms: ✓ Skapar ny post
- GET listings?mine=true: ✓ Returnerar endast egna listningar
- DELETE search-terms: ✓ Raderar post

### Observerat
- Transaction POST kräver ISO 8601 datumformat (med tid) - frontend behöver anpassas

## Go API Endpoints (cmd/api/main.go)

| Method | Endpoint | Handler | Status |
|--------|----------|---------|--------|
| GET | /api/health | healthHandler | OK |
| GET | /api/inventory | getInventory | OK |
| PUT | /api/inventory/{id} | inventoryItemHandler | OK |
| POST | /api/inventory | - | **Saknas** |
| GET | /api/listings | getListings | OK |
| PUT | /api/listings/{id} | listingItemHandler | OK |
| POST | /api/listings | - | **Saknas** |
| DELETE | /api/listings/{id} | listingItemHandler | OK |
| GET | /api/products | getProducts | OK |
| PUT | /api/products/{id} | productItemHandler | OK |
| POST | /api/products | - | **Saknas** |
| GET | /api/transactions | getTransactions | OK |
| POST | /api/transactions | - | **Saknas** |
| DELETE | /api/transactions/{id} | transactionItemHandler | OK |
| GET | /api/transaction-types | getTransactionTypes | OK |
| GET | /api/marketplaces | getMarketplaces | OK |
| GET | /api/search-terms | getSearchTerms | OK |
| POST | /api/search-terms | - | **Saknas** |
| PUT | /api/search-terms/{id} | searchTermItemHandler | OK |
| DELETE | /api/search-terms/{id} | searchTermItemHandler | OK |
| POST | /api/fetch-ads | fetchAdsHandler | OK |
| GET | /api/fetch-ads/status/{jobId} | fetchAdsStatusHandler | OK |

## Frontend API Calls (frontend/pages/*.vue)

### index.vue (Lager)
- GET `/api/inventory` ✓
- GET `/api/products` ✓
- PUT `/api/inventory/{id}` ✓
- **POST `/api/inventory`** ✗

### scraping.vue (Sökord)
- POST `/api/fetch-ads` ✓
- GET `/api/fetch-ads/status/{jobId}` ✓
- GET `/api/search-terms` ✓
- GET `/api/marketplaces` ✓
- **POST `/api/search-terms`** ✗
- PUT `/api/search-terms/{id}` ✓
- DELETE `/api/search-terms/{id}` ✓

### analytics.vue (Analyser)
- GET `/api/inventory` ✓

### transactions.vue (Transaktioner)
- GET `/api/transactions` ✓
- GET `/api/transaction-types` ✓
- **POST `/api/transactions`** ✗
- DELETE `/api/transactions/{id}` ✓

### listings.vue (Mina annonser)
- GET `/api/listings?mine=true` ⚠️ (query param hanteras inte)
- GET `/api/products` ✓
- GET `/api/marketplaces` ✓
- PUT `/api/listings/{id}` ✓
- **POST `/api/listings`** ✗
- DELETE `/api/listings/{id}` ✓

### products.vue (Produkter)
- GET `/api/products` ✓
- PUT `/api/products/{id}` ✓
- **POST `/api/products`** ✗

### ads.vue (Annonser)
- GET `/api/listings` ⚠️ (filtrerar på is_my_listing klient-sida)

## Gaps & Mismatches

### Saknade POST-handlers (7 st)
1. POST `/api/inventory` - index.vue
2. POST `/api/listings` - listings.vue
3. POST `/api/products` - products.vue
4. POST `/api/transactions` - transactions.vue
5. POST `/api/search-terms` - scraping.vue

### Query Param Issues (2 st)
1. GET `/api/listings?mine=true` - hanteras inte i Go
2. GET `/api/listings` utan filter - frontend filtrerar is_my_listing klient-sida i ads.vue

## Recommendations

### Hög prioritet
1. Lägg till POST-handlers för alla resurser som behöver skapas
2. Hantera `?mine=true` query param i getListings

### Medel prioritet
3. Överväg att lägga till DELETE-handler för products
4. Lägg till felhantering för 404 när resurs inte finns

### Låg prioritet
5. Standardisera API-response format
6. Lägg till OpenAPI/Swagger dokumentation
