# Real-time Scraper Log Streaming

## Overview
Visa realtidsloggar från scrapern på /scraping-sidan när scrapern körs. Användaren ska se samma statusuppdateringar som visas i loggarna.

## Current State
- Frontend pollar `/api/fetch-ads/status/{job_id}` varje sekund
- Status innehåller: progress, current_query, ads_found, error
- BotService loggar detaljerad info till `fetch.log`
- Användaren ser bara grundläggande progress, inte detaljerade loggmeddelanden

## Desired State
- Strömma loggmeddelanden i realtid från backend till frontend
- Visa loggarna i en scrollbar logg-konsol på /scraping-sidan
- Loggarna ska visas när scrapern är igång

## Technical Approach

### Backend: Server-Sent Events (SSE)
- Lägg till ny endpoint: `GET /api/fetch-ads/logs/{job_id}`
- Använd SSE för att strömma loggmeddelanden
- Loggmeddelanden sparas i JobService under varje jobb
- BotService rapporterar loggmeddelanden via JobService

### Frontend: SSE Client
- Anslut till SSE-endpointen när hämtning startar
- Visa loggmeddelanden i en terminal-liknande komponent
- Scrolla automatiskt till senaste meddelandet
- Stäng anslutningen när jobbet är klart

## Data Flow

```
BotService (log) → JobService (store) → SSE Handler (stream) → Frontend (display)
```

## UI/UX

### Log Display Component
- Placeras under progress-baren på /scraping
- Max-height: 300px med overflow-y: auto
- Monospace font för terminal-känsla
- Timestamps visas för varje logg
- Olika färger för olika loggnivåer (info, warning, error)
- Auto-scroll till senaste meddelande

### Example Log Messages
```
[13:30:15] Starting ad fetch for job abc123
[13:30:16] Found 5 search terms
[13:30:17] Processing: iPhone 15 Pro
[13:30:18] Found 12 ads for iPhone 15 Pro
[13:30:19] Processing new ad: https://... (price: 8990 SEK)
[13:30:20] Product identified: Apple iPhone 15 Pro (smartphones)
[13:30:21] Saved listing at 8990 SEK
[13:30:22] Completed: iPhone 15 Pro
[13:30:23] All search terms processed. Total ads: 12
```

## Implementation Tasks

1. **Backend: Extend Job Model**
   - Lägg till `Logs []LogEntry` i FetchJob struct
   - Lägg till `AddLog(jobID, level, message)` metod i JobService

2. **Backend: Update BotService Logging**
   - Ersätt `log.Printf()` med `jobService.AddLog()` i BotService
   - Behåll fil-loggning som backup

3. **Backend: SSE Endpoint**
   - Skapa `GET /api/fetch-ads/logs/{job_id}` handler
   - Sätt rätta headers för SSE (Content-Type: text/event-stream)
   - Strömma nya loggar när de tillkommer

4. **Frontend: SSE Hook**
   - Skapa `useScraperLogs(jobId)` composable
   - Anslut till SSE-endpoint
   - Hantera omanslutning vid fel

5. **Frontend: Log Component**
   - Skapa `ScraperLogConsole` komponent
   - Visa loggmeddelanden med timestamps
   - Auto-scroll funktionalitet
   - Styling för olika loggnivåer

6. **Frontend: Integration**
   - Använd `useScraperLogs` i scraping.vue
   - Visa log-komponenten när isFetching är true
   - Stäng SSE när komponenten unmountas

## Files to Modify

### Backend
- `internal/services/job.go` - Add log storage
- `internal/services/bot.go` - Use job logging
- `cmd/api/main.go` - Add SSE endpoint

### Frontend
- `frontend/pages/scraping.vue` - Add log component
- `frontend/composables/useScraperLogs.ts` - New composable
- `frontend/components/ScraperLogConsole.vue` - New component

## Success Criteria
- [ ] Loggar visas i realtid på /scraping när scrapern körs
- [ ] Varje loggmeddelande har timestamp
- [ ] Loggarna scrollar automatiskt till senaste
- [ ] Olika loggnivåer har olika färger (info, warning, error)
- [ ] SSE-anslutningen stängs korrekt när jobbet är klart
- [ ] Fungerar i både dev och prod miljö
