# Shape: Supabase Production Database

## Scope
Create an online PostgreSQL database for the begbot trading system. The database will store products, listings, traded items, transactions, and related data as defined in the existing schema.

## Context
This is the foundational piece for the begbot application. Currently, the database schema exists as `schema_improved.sql` and there's Go code that models the database (`internal/models/models.go`, `internal/db/postgres.go`). The goal is to move from local/local-only database to a production-ready cloud database.

## Decisions

### Why Supabase
- Managed PostgreSQL with free tier
- Built-in authentication (email/password)
- Automatic backups and SSL
- Easy SQL Editor for schema import
- REST API available if needed later
- Compatible with existing Go code using `github.com/lib/pq`

### Authentication Approach
- Use Supabase Auth for email/password authentication
- This separates auth from database access
- Can be integrated later if we want user-specific data
- For now, serves as secure access control

### Database Connection
- Keep existing `github.com/lib/pq` driver
- Only update connection string in config
- No code changes to database layer needed

## Constraints
- Must use existing schema (`schema_improved.sql`) exactly
- Must follow currency storage standard (ints in SEK öre)
- Connection details must be in config, not hardcoded
- Use SSL for all connections
