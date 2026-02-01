# References: Supabase Production Database

## Existing Code References

### 1. Database Schema
**File:** `/home/simon/repos/begbot/schema_improved.sql`

**What it contains:**
- Complete PostgreSQL schema with all tables
- Foreign key relationships
- Indexes for performance
- Initial seed data for `trade_statuses`

**How to use:**
- Copy entire SQL and run in Supabase SQL Editor
- No modifications needed

### 2. Database Layer (Go)
**File:** `/home/simon/repos/begbot/internal/db/postgres.go`

**What it contains:**
- `Postgres` struct with `*sql.DB`
- `NewPostgres()` function that creates connection from config
- `Migrate()` function (we may skip this since we import schema directly)
- CRUD methods: `SaveTradedItem`, `UpdateTradedItemStatus`, `GetTradedItemByID`, etc.
- Helper method `CalculateProfit()`

**Key connection string pattern:**
```go
connStr := fmt.Sprintf(
    "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
    cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
)
```

**How to use:**
- This code works with Supabase without changes
- Just update `config.yaml` with Supabase connection details

### 3. Data Models (Go)
**File:** `/home/simon/repos/begbot/internal/models/models.go`

**What it contains:**
- Struct definitions matching database tables
- `Product`, `TradedItem`, `Listing`, `Transaction`, etc.
- Proper JSON and DB tags
- Nullable fields using pointers

**How to use:**
- No changes needed
- Models already align with schema

### 4. Configuration
**File:** `/home/simon/repos/begbot/internal/config/config.go`

**What it contains:**
- `DatabaseConfig` struct with Host, Port, User, Password, Name, SSLMode
- YAML-based configuration loading
- Nested config structure

**How to use:**
- Create or update `config.yaml` with Supabase details:
```yaml
database:
  host: db.xxxxxx.supabase.co
  port: 5432
  user: postgres
  password: your_password
  name: postgres
  sslmode: require
```

## External References

### Supabase Documentation
- Getting Started: https://supabase.com/docs
- Database Setup: https://supabase.com/docs/guides/database
- Authentication: https://supabase.com/docs/guides/auth
- SQL Editor: https://supabase.com/docs/guides/database/query-editor

### PostgreSQL Driver
- `github.com/lib/pq`: https://github.com/lib/pq
- Already imported in code, no changes needed
