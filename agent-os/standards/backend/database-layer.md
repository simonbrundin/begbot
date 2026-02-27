## Database Layer Pattern

### Direct SQL with pgx

Use `database/sql` with the pgx driver directly. No ORM.

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
    db *sql.DB
}

func NewPostgres(cfg config.DatabaseConfig) (*Postgres, error) {
    connStr := fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=10",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
    )

    db, err := sql.Open("pgx", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Connection pool settings
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &Postgres{db: db}, nil
}
```

### Connection Pool Defaults

- `MaxOpenConns`: 10
- `MaxIdleConns`: 5
- `ConnMaxLifetime`: 5 minutes
- `ConnectTimeout`: 10 seconds

### Migrations

Embed migrations as string slices in the database package:

```go
func (p *Postgres) Migrate() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS products (...)`,
        `CREATE TABLE IF NOT EXISTS colors (...)`,
    }
    for _, q := range queries {
        if _, err := p.db.Exec(q); err != nil {
            return err
        }
    }
    return nil
}
```

### PostgreSQL Array Handling

PostgreSQL arrays need custom parsing:

```go
func parseIntegerArray(raw interface{}) []int64 {
    if raw == nil {
        return nil
    }
    str, ok := raw.(string)
    if !ok {
        return nil
    }
    if str == "{}" || str == "" {
        return []int64{}
    }
    re := regexp.MustCompile(`\d+`)
    matches := re.FindAllString(str, -1)
    var result []int64
    for _, m := range matches {
        var n int64
        fmt.Sscanf(m, "%d", &n)
        result = append(result, n)
    }
    return result
}
```
