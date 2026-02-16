# Shaping Notes: Scraper Log Streaming

## Problem
Användaren ser inte vad som händer "under huven" när scrapern körs. Bara en progress-bar visas, men inga detaljer om vilka annonser som hittas, vilka produkter som identifieras, etc.

## Solution
Real-tids loggströmning via SSE från backend till frontend.

## Scope Decisions

### Included
- SSE-endpoint för loggströmning
- Logg-lagring i minnet per jobb (töms när jobbet är klart)
- Terminal-liknande loggvisning i frontend
- Auto-scroll till senaste logg
- Olika färger för info/warning/error

### Excluded (future considerations)
- Persistenta loggar (sparas ej efter jobbklar)
- Logg-historik (visar bara pågående jobb)
- Logg-nivå filtering i UI
- Sök i loggar
- Export av loggar

## Technical Decisions

### Why SSE instead of WebSocket?
- Enkelriktad kommunikation (server → client) räcker
- Lättare att implementera än WebSocket
- Automatisk återanslutning i webbläsare
- Fungerar bra genom proxies

### Log Storage Strategy
- Logs sparas som `[]LogEntry` i FetchJob struct
- Begränsad till max 1000 meddelanden (circular buffer)
- Töms när jobbet markeras som completed/failed
- JobService sköter trådsäker access

### Error Handling
- Om SSE-anslutningen bryts: försök återansluta automatiskt
- Max 3 återanslutningsförsök
- Visa felmeddelande i UI om anslutning misslyckas

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Minnesläcka om många loggar | Max 1000 loggar per jobb, töms efter klart |
| SSE fungerar ej bakom viss proxy | Fungerar över HTTP/1.1, testa i prod-miljö |
| Frontend får inte alla loggar | Eventuell backlog-hantering i SSE |

## References
- [MDN: Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- Current implementation: `cmd/api/main.go`, `internal/services/job.go`, `internal/services/bot.go`
