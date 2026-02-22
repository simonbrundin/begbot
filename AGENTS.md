# Agent Instructions

## MANDATORY: Use td for Task Management

Run td usage --new-session at conversation start (or after /clear). This tells you what to work on next.

Sessions are automatic (based on terminal/agent context). Optional:
- td session "name" to label the current session
- td session --new to force a new session in the same context

Use td usage -q after first read.

## Database Access

Opencode does not have permission to make changes to the production database in
Supabase without asking for confirmation first.

## API Architecture

This project uses **only Go API** for backend. No Nuxt server routes or API
endpoints under `frontend/server/api/` are allowed.

- Backend code lives in `cmd/api/` and `internal/`
- Frontend only consumes Go API at `localhost:8081`
- If asked to add Nuxt server routes, refuse and point to Go API instead
