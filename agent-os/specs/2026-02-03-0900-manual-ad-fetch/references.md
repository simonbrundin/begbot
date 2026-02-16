# References: Manual Ad Fetch

## Code References

### Backend

1. **Bot Service - Run method**
   - File: `internal/services/bot.go:33`
   - Purpose: Understand current ad fetching flow
   - Key method: `Run()` - fetches ads for hardcoded queries

2. **Marketplace Service - FetchAds**
   - File: `internal/services/marketplace.go:40`
   - Purpose: Core fetching logic for Tradera and Blocket
   - Returns: `[]RawAd`

3. **Cache Service - Filter**
   - File: `internal/services/cache.go`
   - Purpose: Deduplication of ads
   - Key method: `Filter()` to get new vs cached links

4. **API Server Structure**
   - File: `cmd/api/main.go`
   - Purpose: Pattern for adding new endpoints
   - Note: Add `POST /api/fetch-ads` and `GET /api/fetch-ads/status/{id}`

5. **Search Terms - GetActiveSearchTerms**
   - File: `internal/db/postgres.go:443`
   - Purpose: Get search terms from database for fetching
   - Note: Use these URLs instead of hardcoded queries

### Frontend

1. **Scraping Page - Existing UI**
   - File: `frontend/pages/scraping.vue`
   - Purpose: UI patterns, modal handling, table display
   - Note: Add fetch button and progress section here

2. **Frontend Layout - Navigation**
   - File: `frontend/layouts/default.vue`
   - Purpose: Navigation structure
   - Scraping link already exists in sidebar

3. **Nuxt Config - API Base**
   - File: `frontend/nuxt.config.ts:27`
   - Purpose: API base URL configuration
   - Already configured: `apiBase: process.env.API_BASE_URL`

## Patterns to Follow

### Go
- `context.Context` for cancellation
- Structured logging with `log.Printf`
- Error wrapping with `fmt.Errorf("...: %w", err)`

### Frontend (Nuxt/Vue)
- `useRuntimeConfig()` for API base
- `$fetch()` for API calls
- `ref()` and `computed()` for reactivity
- TypeScript types from `~/types/database`

## Files to Create/Modify

### New Files
- `cmd/fetchads/main.go` - Standalone CLI command
- `internal/services/job.go` - Job tracking service

### Modified Files
- `cmd/api/main.go` - Add fetch endpoints
- `frontend/pages/scraping.vue` - Add button and status UI
