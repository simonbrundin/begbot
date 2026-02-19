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
}

type TraderaConfig struct {
	Enabled bool          `yaml:"enabled"`
	Timeout time.Duration `yaml:"timeout"`
	BaseURL string        `yaml:"base_url"`
}

type BlocketConfig struct {
	Enabled bool          `yaml:"enabled"`
	Timeout time.Duration `yaml:"timeout"`
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

	applyEnvOverrides(&cfg)

	return &cfg, nil
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
}
