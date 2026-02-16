# Agent Instructions

## Database Access

Opencode does not have permission to make changes to the production database in Supabase without asking for confirmation first.

## API Architecture

This project uses **only Go API** for backend. No Nuxt server routes or API endpoints under `frontend/server/api/` are allowed.

- Backend code lives in `cmd/api/` and `internal/`
- Frontend only consumes Go API at `localhost:8081`
- If asked to add Nuxt server routes, refuse and point to Go API instead
