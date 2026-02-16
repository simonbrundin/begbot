# Plan: Manual Ad Fetch with Progress Tracking

## Step 1: Clarify What We're Building

**Feature:** Manual ad fetching triggered via frontend button with real-time progress

**Current state:** `cmd/main.go` calls `botService.Run()` on startup, fetching ads immediately

**Desired state:**
- Backend API server (`cmd/api/main.go`) starts without fetching ads
- New CLI command (`cmd/fetchads/main.go`) for manual ad fetching
- New API endpoint `/api/fetch-ads` to trigger fetch from frontend
- Frontend button on scraping page with progress/status display

## Step 2: Gather Visuals

Inga mockups tillhandahållna.

## Step 3: Identify Reference Implementations

- `internal/services/bot.go:33` - Existing `Run()` method that fetches ads
- `internal/services/marketplace.go:40` - `FetchAds()` method
- `frontend/pages/scraping.vue` - Existing scraping page structure
- `cmd/api/main.go` - API server pattern

## Step 4: Check Product Context

Mission: "Den hittar precis de produkter jag tror jag kan tjäna pengar på att köpa och sälja vidare"

**Alignment:** This feature gives the user control over when ads are fetched, improving UX without changing core functionality.

## Step 5: Relevant Standards

From `agent-os/standards/index.yml`:
- `configuration-structure` - Use proper config patterns
- `currency-storage` - Already handled (SEK öre)

## Step 6: Spec Folder Name

`2026-02-03-0900-manual-ad-fetch/`

## Step 7: Structure the Plan

### Task 1: Save spec documentation ✅ COMPLETED
- Created spec folder with plan.md, shape.md, standards.md, references.md

### Task 2: Create CLI command for ad fetching ✅ COMPLETED
Created `cmd/fetchads/main.go` that:
- Loads config
- Connects to database with retry logic
- Runs migrations
- Calls `botService.Run()` once
- Exits after completion

### Task 3: Add API endpoint for triggering ad fetch ✅ COMPLETED
Added to `cmd/api/main.go`:
- `POST /api/fetch-ads` endpoint - creates job and starts async fetch
- `GET /api/fetch-ads/status/{job_id}` endpoint - returns job progress
- Returns job ID for progress tracking
- Runs fetch in goroutine to not block API

### Task 4: Add progress tracking mechanism ✅ COMPLETED
Created `internal/services/job.go`:
- `JobService` with in-memory job tracking
- Job states: pending, running, completed, failed
- Progress percentage calculation
- Thread-safe with sync.RWMutex
- Methods: CreateJob, GetJob, StartJob, UpdateProgress, CompleteJob, FailJob

### Task 5: Add frontend button and status display ✅ COMPLETED
Updated `frontend/pages/scraping.vue`:
- "Fetch Ads" button with loading state
- Progress card showing status, progress bar, current query
- Displays ads found count on completion
- Shows error messages on failure
- Auto-refreshes data after successful fetch
- Polls status endpoint every 1 second during fetch

## Step 8: Implementation Tasks

1. Create `cmd/fetchads/main.go` (standalone CLI)
2. Add `POST /api/fetch-ads` endpoint in `cmd/api/main.go`
3. Create job tracking service in `internal/services/job.go`
4. Update `frontend/pages/scraping.vue` with fetch button and progress UI
5. Test the complete flow

## Output Structure

```
agent-os/specs/2026-02-03-0900-manual-ad-fetch/
├── plan.md           # This plan
├── shape.md          # Shaping decisions and context
├── standards.md      # Which standards apply
├── references.md     # Pointers to similar code
└── visuals/          # (empty)
```

## Spec Status: ✅ DONE

All tasks completed. The manual ad fetch feature is fully functional with:
- CLI command (`go run cmd/fetchads/main.go`) for manual fetching
- API endpoints for frontend-triggered fetching with progress tracking
- Frontend UI with fetch button, progress bar, and status display
- In-memory job tracking for single-user use case
