# Shaping Notes: Port-based Process Kill

## Scope
Enhance `dev.nu` to kill processes by port number (8080 + 3000) instead of pattern matching.

## Current Behavior
```nushell
try { pkill -9 -f "go run ./cmd/api/main.go" } catch { }
try { pkill -9 -f "npm run dev" } catch { }
```

## Desired Behavior
```nushell
kill-dev-ports
```

## Decisions
1. **Port-based over pattern-based**: More reliable, works regardless of how process was started
2. **Use `lsof -i :PORT`**: Standard tool for finding processes by port
3. **Single function for both ports**: `kill-dev-ports` handles 8080 and 3000
4. **Keep graceful error handling**: No error if port is already free

## Context
- Backend runs on port 8080 (Go API)
- Frontend runs on port 3000 (Nuxt)
- Both are started in background with `&`
