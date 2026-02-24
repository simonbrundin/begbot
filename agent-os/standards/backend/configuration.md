## Configuration Loading

### Environment Variables Only

All configuration must come from environment variables. No config files in production.

```go
type Config struct {
    Database  DatabaseConfig
    App       AppConfig
    LLM       LLMConfig
    Email     EmailConfig
}

func Load() (*Config, error) {
    cfg := &Config{
        Database: DatabaseConfig{
            Host:     getEnvOrDefault("DATABASE_HOST", "localhost"),
            Port:     getEnvAsInt("DATABASE_PORT", 5432),
            User:     getEnvOrDefault("DATABASE_USER", "postgres"),
            Password: os.Getenv("DATABASE_PASSWORD"),
            Name:     getEnvOrDefault("DATABASE_NAME", "begbot"),
            SSLMode:  getEnvOrDefault("DATABASE_SSLMODE", "disable"),
        },
        // ...
    }
    return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return defaultValue
}
```

### Required vs Optional

- Required values: Fail startup if missing (database, API keys)
- Optional values: Use sensible defaults
- Secrets: NEVER log or hardcode

### No YAML in Production

Remove YAML config loading. Environment variables are the single source of truth.
