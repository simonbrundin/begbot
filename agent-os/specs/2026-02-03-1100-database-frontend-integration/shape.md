# Shaping Notes

## Scope
Debug and fix why database data is not visible in the Nuxt 3 frontend.

## Key Decisions

### Tech Stack (from agent-os/product/tech-stack.md)
- **Backend**: Go
- **Frontend**: Nuxt 3 (Vue 3)
- **Database**: PostgreSQL (Supabase)
- **ORM**: pgx (raw SQL, no ORM)

### Data Flow
```
Supabase DB → Go Backend (port 8081) → Nuxt Frontend (port 3000)
```

## Context
- `simon dev` command runs both backend and frontend
- API base URL configured via environment variables
- Frontend uses `$fetch` wrapper for API calls

## Files to Examine
- `cmd/api/main.go` - API server entry point
- `internal/db/postgres.go` - Database connection
- `frontend/nuxt.config.ts` - API configuration
- Frontend pages: `index.vue`, `products.vue`, `listings.vue`, etc.
