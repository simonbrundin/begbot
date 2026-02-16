# Shape: Manual Ad Fetch with Progress Tracking

## Scope

Separera annonshämtning från serverstart och lägg till manuell trigger från frontend med statusvisning.

## Key Decisions

### 1. Architecture Choice: Job Queue Pattern

**Decision:** Use simple in-memory job tracking instead of full queue system

**Rationale:**
- Single-user application, no concurrent fetch jobs needed
- Keep it simple - only one user triggering at a time
- Can be enhanced later with Redis queue if needed

**Implementation:**
- `FetchJob` struct with status, progress, timestamp, result count, error
- In-memory map `jobID -> FetchJob` in API server
- SSE or polling for frontend updates

### 2. API Endpoint Design

**Decision:** REST endpoint with job ID response

```
POST /api/fetch-ads
Response: { "job_id": "uuid", "status": "running" }

GET /api/fetch-ads/status/{job_id}
Response: { "status": "running", "progress": 45, "total_queries": 3, "current_query": "iphone 14" }
```

### 3. Frontend State Management

**Decision:** Use reactive state in scraping.vue with polling

**Rationale:**
- Nuxt/Pinia already available
- Simple polling (every 1s) sufficient for this use case
- No need for WebSocket complexity

### 4. CLI Command Structure

**Decision:** Separate `cmd/fetchads/main.go` that mirrors `cmd/main.go` but only runs bot once

**Rationale:**
- Clear separation of concerns
- Can be run via cron or manually
- No startup-fetch behavior in API server

## Context

### Existing Code to Reuse
- `botService.Run()` in `internal/services/bot.go`
- `marketplaceService.FetchAds()` in `internal/services/marketplace.go`
- `cacheService.Filter()` for deduplication
- Scraping page UI patterns in `frontend/pages/scraping.vue`

### Constraints
- Keep API server startup fast (no background jobs on start)
- Allow user to start backend for browsing data without triggering fetches
- Provide feedback during potentially long-running fetch operation

## Out of Scope

- Scheduled/cron fetches (future enhancement)
- Concurrent fetch jobs (single user = single job)
- Persistent job queue (in-memory sufficient for now)
