package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	App       AppConfig       `yaml:"app"`
	Scraping  ScrapingConfig  `yaml:"scraping"`
	LLM       LLMConfig       `yaml:"llm"`
	Valuation ValuationConfig `yaml:"valuation"`
	Email     EmailConfig     `yaml:"email"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type AppConfig struct {
	LogLevel string        `yaml:"log_level"`
	CacheTTL time.Duration `yaml:"cache_ttl"`
}

type ScrapingConfig struct {
	Tradera TraderaConfig `yaml:"tradera"`
	Blocket BlocketConfig `yaml:"blocket"`
	Proxy   ProxyConfig   `yaml:"proxy"`
	Scraper ScraperConfig `yaml:"scraper"`
}

type ScraperConfig struct {
	Provider    string `yaml:"provider"`
	EvomiAPIKey string `yaml:"evomi_api_key"`
}

type ProxyConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Country  string `yaml:"country"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type TraderaConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	BaseURL string        `yaml:"base_url"`
	AppID   string        `yaml:"app_id"`
	AppKey  string        `yaml:"app_key"`
	// EnrichLimit controls how many GetItem enrichment calls are made per
	// FetchAds invocation. Keep small to avoid hitting API quotas.
	EnrichLimit int `yaml:"enrich_limit"`
	// EnrichCacheTTL controls how long GetItem responses are cached locally.
	EnrichCacheTTL time.Duration `yaml:"enrich_cache_ttl"`
	// SaveOnlyBuyNow: when true, only save Tradera ads that have a confirmed buy-now price
	SaveOnlyBuyNow bool `yaml:"save_only_buy_now"`
	// EnrichTimeoutSeconds: timeout for per-ad enrichment attempts (seconds)
	EnrichTimeoutSeconds int `yaml:"enrich_timeout_seconds"`
	// EnrichMaxRetries: max retry attempts for API enrichment (default 3)
	EnrichMaxRetries int `yaml:"enrich_max_retries"`
	// EnrichBackoffSeconds: base backoff time in seconds between retries (default 2)
	EnrichBackoffSeconds int `yaml:"enrich_backoff_seconds"`
}

type BlocketConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	BaseURL string        `yaml:"base_url"`
}

type LLMConfig struct {
	Provider     string            `yaml:"provider"`
	APIKey       string            `yaml:"api_key"`
	SiteURL      string            `yaml:"site_url"`
	SiteName     string            `yaml:"site_name"`
	Timeout      time.Duration     `yaml:"timeout"`
	DefaultModel string            `yaml:"default_model"`
	Models       map[string]string `yaml:"models"`
}

type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host"`
	SMTPPort     string   `yaml:"smtp_port"`
	SMTPUsername string   `yaml:"smtp_username"`
	SMTPPassword string   `yaml:"smtp_password"`
	From         string   `yaml:"from"`
	Recipients   []string `yaml:"recipients"`
}

type ValuationConfig struct {
	TargetSellDays  int     `yaml:"target_sell_days"`
	MinProfitMargin float64 `yaml:"min_profit_margin"`
	SafetyMargin    float64 `yaml:"safety_margin"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Scraping.Tradera.Timeout == 0 {
		cfg.Scraping.Tradera.Timeout = 30 * time.Second
	}
	if cfg.Scraping.Tradera.EnrichLimit == 0 {
		cfg.Scraping.Tradera.EnrichLimit = 10
	}
	if cfg.Scraping.Tradera.EnrichCacheTTL == 0 {
		cfg.Scraping.Tradera.EnrichCacheTTL = 24 * time.Hour
	}
	// Default: enforce save-only-buy-now and short enrichment timeout
	if cfg.Scraping.Tradera.EnrichTimeoutSeconds == 0 {
		cfg.Scraping.Tradera.EnrichTimeoutSeconds = 5
	}
	// default SaveOnlyBuyNow is true to opt into the new stricter behavior
	// unless explicitly disabled in config
	// Note: projects upgrading can set this to false to preserve old behavior
	if !cfg.Scraping.Tradera.SaveOnlyBuyNow {
		cfg.Scraping.Tradera.SaveOnlyBuyNow = true
	}
	if cfg.Scraping.Blocket.Timeout == 0 {
		cfg.Scraping.Blocket.Timeout = 30 * time.Second
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("DATABASE_SSLMODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.App.LogLevel = v
	}
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_SITE_URL"); v != "" {
		cfg.LLM.SiteURL = v
	}
	if v := os.Getenv("LLM_SITE_NAME"); v != "" {
		cfg.LLM.SiteName = v
	}
	if v := os.Getenv("LLM_DEFAULT_MODEL"); v != "" {
		cfg.LLM.DefaultModel = v
	}
	if v := os.Getenv("SMTP_HOST"); v != "" {
		cfg.Email.SMTPHost = v
	}
	if v := os.Getenv("SMTP_PORT"); v != "" {
		cfg.Email.SMTPPort = v
	}
	if v := os.Getenv("SMTP_USERNAME"); v != "" {
		cfg.Email.SMTPUsername = v
	}
	if v := os.Getenv("SMTP_PASSWORD"); v != "" {
		cfg.Email.SMTPPassword = v
	}
	if v := os.Getenv("SMTP_FROM"); v != "" {
		cfg.Email.From = v
	}
	if v := os.Getenv("TRADERA_APP_ID"); v != "" {
		cfg.Scraping.Tradera.AppID = v
	}
	if v := os.Getenv("TRADERA_APP_KEY"); v != "" {
		cfg.Scraping.Tradera.AppKey = v
	}
	if v := os.Getenv("TRADERA_ENRICH_LIMIT"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n >= 0 {
			cfg.Scraping.Tradera.EnrichLimit = n
		}
	}
	if v := os.Getenv("PROXY_PROVIDER"); v != "" {
		cfg.Scraping.Proxy.Provider = v
	}
	if v := os.Getenv("PROXY_API_KEY"); v != "" {
		cfg.Scraping.Proxy.APIKey = v
	}
	if v := os.Getenv("PROXY_COUNTRY"); v != "" {
		cfg.Scraping.Proxy.Country = v
	}
	if v := os.Getenv("PROXY_USERNAME"); v != "" {
		cfg.Scraping.Proxy.Username = v
	}
	if v := os.Getenv("PROXY_PASSWORD"); v != "" {
		cfg.Scraping.Proxy.Password = v
	}
	if v := os.Getenv("SCRAPER_PROVIDER"); v != "" {
		cfg.Scraping.Scraper.Provider = v
	}
	if v := os.Getenv("EVOMI_SCRAPER_API_KEY"); v != "" {
		cfg.Scraping.Scraper.EvomiAPIKey = v
	}
}
