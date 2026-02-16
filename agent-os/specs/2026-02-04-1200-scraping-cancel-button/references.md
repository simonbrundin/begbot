# Scraping Cancel Button - References

## Key Files

### Backend

**Job Service:**
- `internal/services/job.go` - Job management, statuses, logs
  - Lägg till `JobStatusCancelled` konstant (rad 14)
  - Lägg till `CancelChan` i `FetchJob` struct (rad 31-45)
  - Lägg till `CancelJob()` metod

**Bot Service:**
- `internal/services/bot.go` - Scraping logic
  - Modifiera `Run()` för att kolla cancel-channel
  - Hantera clean exit vid cancellation

**API Handlers:**
- `cmd/api/main.go:539-570` - fetchAdsHandlerWithConfig
- `cmd/api/main.go:572-602` - fetchAdsStatusHandler
- Lägg till cancel handler efter rad 602

### Frontend

**Scraping Page:**
- `frontend/pages/scraping.vue` - Main scraping page
  - Progress card: rad 19-42
  - Status text: computed `fetchStatusText` rad 196-210
  - Polling logik: rad 212-264
  - Lägg till cancel-knapp i progress card

**Composables:**
- `frontend/composables/useApi.ts` - API client
- `frontend/composables/useScraperLogs.ts` - Log streaming

## Existing Patterns

### Job Status Flow
```
CreateJob() → pending
StartJob()  → running
CompleteJob() → completed
FailJob()   → failed
CancelJob() → cancelled (NY)
```

### API Response Format
```json
{
  "job_id": "abc123",
  "status": "cancelled"
}
```

### Error Response Format
```json
{
  "error": "Job not found",
  "code": "NOT_FOUND"
}
```

## Similar Features

### Delete Search Term
**File:** `frontend/pages/scraping.vue:319-327`
```typescript
const deleteTerm = async (id: number) => {
  if (!confirm("Ta bort detta sökord?")) return;
  try {
    await api.delete(`/search-terms/${id}`);
    await fetchData();
  } catch (e) {
    console.error("Failed to delete term:", e);
  }
};
```

### Job Status Polling
**File:** `frontend/pages/scraping.vue:222-249`
- Pollar varje sekund
- Stoppas vid completed/failed
- Ska även stoppas vid cancelled
