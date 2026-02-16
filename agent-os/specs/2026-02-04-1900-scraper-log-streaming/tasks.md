# Tasks: Scraper Log Streaming

## Task 1: Save spec documentation
- [x] Create spec folder `2026-02-04-1900-scraper-log-streaming/`
- [x] Write `plan.md`
- [x] Write `shape.md`
- [x] Write `standards.md`
- [x] Write `references.md`

## Task 2: Extend Job Model with Log Storage
**File:** `internal/services/job.go`

- [ ] Add `LogEntry` struct with fields: `Timestamp`, `Level`, `Message`
- [ ] Add `Logs []LogEntry` field to `FetchJob` struct (max 1000 entries)
- [ ] Add `AddLog(jobID, level, message)` method to JobService
- [ ] Add `GetLogs(jobID)` method to JobService
- [ ] Implement circular buffer for logs (evict old when >1000)
- [ ] Ensure thread-safe access with mutex

## Task 3: Update BotService to Use Job Logging
**File:** `internal/services/bot.go`

- [ ] Replace `log.Printf()` calls with `jobService.AddLog()` in Run()
- [ ] Add logging for:
  - Job start
  - Search terms found
  - Each search term processing
  - Ads found per term
  - Each ad processing
  - Product identification
  - Listing saved
  - Job completion
- [ ] Keep file logging as backup (MultiWriter)

## Task 4: Create SSE Endpoint for Logs
**File:** `cmd/api/main.go`

- [ ] Add `fetchAdsLogsHandler` function
- [ ] Set SSE headers: Content-Type, Cache-Control, Connection
- [ ] Listen for new logs via channel/polling
- [ ] Send logs as SSE events
- [ ] Send `complete` event when job finishes
- [ ] Handle client disconnect
- [ ] Register route: `GET /api/fetch-ads/logs/`

## Task 5: Create SSE Composable
**File:** `frontend/composables/useScraperLogs.ts` (new)

- [ ] Create composable function `useScraperLogs(jobId)`
- [ ] Setup EventSource connection
- [ ] Reactive `logs` array
- [ ] Handle onmessage events
- [ ] Parse JSON log data
- [ ] Auto-reconnect on error (max 3 attempts)
- [ ] Cleanup on unmount
- [ ] Return: `{ logs, isConnected, error, reconnect }`

## Task 6: Create Log Console Component
**File:** `frontend/components/ScraperLogConsole.vue` (new)

- [ ] Create Vue component
- [ ] Accept `logs` prop (array of log entries)
- [ ] Display logs with monospace font
- [ ] Show timestamp, level, and message
- [ ] Different colors for info/warning/error levels
- [ ] Scrollable container with max-height: 300px
- [ ] Auto-scroll to bottom on new logs
- [ ] Add "clear logs" button
- [ ] Empty state when no logs

## Task 7: Integrate Log Console in Scraping Page
**File:** `frontend/pages/scraping.vue`

- [ ] Import `useScraperLogs` composable
- [ ] Import `ScraperLogConsole` component
- [ ] Call `useScraperLogs(currentJobId)` when fetch starts
- [ ] Add `<ScraperLogConsole :logs="logs" />` under progress bar
- [ ] Only show when `isFetching` is true
- [ ] Close SSE connection when job completes
- [ ] Handle reconnect on error

## Task 8: Test and Verify
- [ ] Start scraper and verify logs appear in real-time
- [ ] Verify timestamps are correct
- [ ] Verify different log levels have correct colors
- [ ] Verify auto-scroll works
- [ ] Verify SSE reconnects on connection drop
- [ ] Test with multiple search terms
- [ ] Test error scenarios

## Task 9: Polish and Final Review
- [ ] Review code style
- [ ] Add comments where needed
- [ ] Verify Swedish text everywhere
- [ ] Test in different browsers
- [ ] Run any existing tests
