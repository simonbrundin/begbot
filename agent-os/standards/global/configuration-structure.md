# Configuration Structure

**Rule:**
- Använd nested structs för varje sektion
- time.Duration för timeouts
- yaml:"snake_case" tags matchar YAML nycklar

**Exempel:**
```go
type Config struct {
    Database  DatabaseConfig  `yaml:"database"`
    Scraping  ScrapingConfig  `yaml:"scraping"`
}

type ScrapingConfig struct {
    Tradera TraderaConfig `yaml:"tradera"`
    Blocket BlocketConfig `yaml:"blocket"`
}

type TraderaConfig struct {
    Enabled bool          `yaml:"enabled"`
    Timeout time.Duration `yaml:"timeout"` // "30s" i YAML
}

func Load(path string) (*Config, error)
```
