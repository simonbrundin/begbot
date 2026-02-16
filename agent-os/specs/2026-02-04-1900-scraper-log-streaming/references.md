# References

## Existing Code to Reference

### Job System
- `internal/services/job.go` - JobService with mutex-protected job storage
- `cmd/api/main.go` - API handlers, fetch-ads endpoints

### Current Status Polling
- `frontend/pages/scraping.vue` - Lines 200-246: fetchAds and polling logic
- `frontend/pages/scraping.vue` - Lines 19-42: Status display UI

### Bot Logging
- `internal/services/bot.go` - Lines 64-137: Run() method with logging
- `internal/services/bot.go` - Lines 100-128: Search term processing loop

### Log File
- `/home/simon/repos/begbot/fetch.log` - Example log format

## Similar Implementations in Codebase
None currently use SSE. Closest is the status polling in scraping.vue.

## External References
- [Go SSE Example](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
- [Vue 3 SSE Composable Pattern](https://vueuse.org/core/useEventSource/)
