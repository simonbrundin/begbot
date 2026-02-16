# Scraping Cancel Button - Plan

## Overview
Lägga till en avbryt-knapp på `/scraping`-sidan som låter användaren avbryta pågående eller väntande scrapningsjobb.

## Requirements

### User Requirements
- Knappen ska visas när ett jobb är i status "pending" eller "running"
- Vid klick avbryts jobbet omedelbart
- Redan insamlad data ska sparas
- Jobbet får status "cancelled"
- Användaren ser tydlig feedback om att jobbet avbrutits

### Technical Requirements
- Thread-safe cancellation via context/channel
- Spara partial data som redan scrapats
- Clean up resources (goroutines, connections)
- Frontend ska sluta polla vid cancellation

## Implementation Tasks

### Task 1: Save spec documentation ✅
Skapa spec-mapp med plan.md, shape.md, standards.md, references.md.

### Task 2: Add cancelled status to backend
**Files:** `internal/services/job.go`

1. Lägg till `JobStatusCancelled` konstant
2. Lägg till `CancelChan` i `FetchJob` struct
3. Lägg till `CancelJob()` metod i `JobService`
4. Uppdatera `CreateJob()` för att initiera CancelChan

### Task 3: Implement cancellation in BotService
**Files:** `internal/services/bot.go`

1. Lägg till context cancellation check i Run() metoden
2. Kolla CancelChan mellan varje sökterm
3. Vid cancellation: Spara partial data, logga, returnera gracefully
4. Se till att job markeras som cancelled, inte failed

### Task 4: Add cancel API endpoint
**Files:** `cmd/api/main.go`

1. Lägg till `POST /api/fetch-ads/cancel/{job_id}` endpoint
2. Validera att jobbet finns och kan avbrytas (pending/running)
3. Anropa `jobService.CancelJob()`
4. Returnera 200 vid success, 400/404 vid fel

### Task 5: Add cancel button to frontend
**Files:** `frontend/pages/scraping.vue`

1. Lägg till "Avbryt"-knapp bredvid progress-indikatorn
2. Knappen visas bara för pending/running jobb
3. Vid klick: anropa cancel API, visa loading state
4. Uppdatera `fetchStatusText` computed för att visa "Avbrutet"
5. Uppdatera polling-logik för att hantera cancelled status

### Task 6: Testing
- [ ] Testa avbryt pending jobb
- [ ] Testa avbryt running jobb under pågående scraping
- [ ] Verifiera att partial data sparas
- [ ] Verifiera att UI uppdateras korrekt
- [ ] Testa att avbryta redan avslutat jobb (ska ge fel)

## Success Criteria
- ✅ Knapp visas korrekt för pending/running jobb
- ✅ Jobb avbryts omedelbart vid klick
- ✅ Partial data sparas till databas
- ✅ Jobb får status "cancelled"
- ✅ UI slutar polla och visar "Avbrutet"
- ✅ Inga memory leaks eller dangling goroutines

## Timeline
Uppskattad tid: 2-3 timmar
