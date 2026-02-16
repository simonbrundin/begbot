# Standards

## Applicable Standards

### Backend
- `backend/api.md` - API endpoint structure and conventions
- `global/coding-style.md` - Go code style
- `global/error-handling.md` - Error handling patterns

### Frontend
- `frontend/components.md` - Component structure
- `frontend/swedish-text.md` - All text in Swedish
- `frontend/css.md` - Tailwind CSS usage

### Global
- `global/tech-stack.md` - Technology choices
- `global/conventions.md` - Naming conventions

## SSE-Specific Standards

### Endpoint Naming
- Path: `/api/fetch-ads/logs/{job_id}`
- Method: GET
- Follow existing API patterns in `cmd/api/main.go`

### SSE Format
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

id: 1
data: {"timestamp":"2026-02-04T13:30:15Z","level":"info","message":"Starting ad fetch"}

event: complete
data: {"job_id":"abc123","status":"completed"}
```

### Frontend SSE Handling
```typescript
const eventSource = new EventSource(`/api/fetch-ads/logs/${jobId}`)
eventSource.onmessage = (event) => {
  const log = JSON.parse(event.data)
  logs.value.push(log)
}
eventSource.onerror = () => {
  // Handle error, attempt reconnect
}
```
