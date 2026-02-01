package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_LLMConfigWithModels(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
database:
  host: "localhost"
  port: 5432
  user: "test"
  password: "test"
  name: "testdb"
  sslmode: "require"

app:
  log_level: "info"
  cache_ttl: 1h

scraping:
  tradera:
    enabled: false
    timeout: 10s
  blocket:
    enabled: false
    timeout: 10s

llm:
  provider: "openrouter"
  api_key: "test-key"
  site_url: "http://localhost:3000"
  site_name: "Begbot"
  default_model: "deepseek/deepseek-v3.2"
  models:
    ExtractProductInfo: "anthropic/claude-sonnet-4-20250514"
    ValidateProduct: "google/gemini-2.5-pro"
    CheckProductCondition: "openai/gpt-4o-mini"
    EstimateNewPrice: "deepseek/deepseek-chat"

valuation:
  target_sell_days: 14
  min_profit_margin: 0.15
  safety_margin: 0.2

email:
  smtp_host: "localhost"
  smtp_port: "587"
  smtp_username: "test"
  smtp_password: "test"
  from: "test@example.com"
  recipients:
    - "test@example.com"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LLM.Provider != "openrouter" {
		t.Errorf("Provider = %v, want %v", cfg.LLM.Provider, "openrouter")
	}

	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("APIKey = %v, want %v", cfg.LLM.APIKey, "test-key")
	}

	if cfg.LLM.SiteURL != "http://localhost:3000" {
		t.Errorf("SiteURL = %v, want %v", cfg.LLM.SiteURL, "http://localhost:3000")
	}

	if cfg.LLM.SiteName != "Begbot" {
		t.Errorf("SiteName = %v, want %v", cfg.LLM.SiteName, "Begbot")
	}

	if cfg.LLM.DefaultModel != "deepseek/deepseek-v3.2" {
		t.Errorf("DefaultModel = %v, want %v", cfg.LLM.DefaultModel, "deepseek/deepseek-v3.2")
	}

	if len(cfg.LLM.Models) != 4 {
		t.Errorf("Models count = %v, want %v", len(cfg.LLM.Models), 4)
	}

	expectedModels := map[string]string{
		"ExtractProductInfo":    "anthropic/claude-sonnet-4-20250514",
		"ValidateProduct":       "google/gemini-2.5-pro",
		"CheckProductCondition": "openai/gpt-4o-mini",
		"EstimateNewPrice":      "deepseek/deepseek-chat",
	}

	for key, expectedValue := range expectedModels {
		if cfg.LLM.Models[key] != expectedValue {
			t.Errorf("Models[%s] = %v, want %v", key, cfg.LLM.Models[key], expectedValue)
		}
	}
}

func TestLoad_LLMConfigWithoutModels(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
database:
  host: "localhost"
  port: 5432
  user: "test"
  password: "test"
  name: "testdb"
  sslmode: "require"

app:
  log_level: "info"
  cache_ttl: 1h

scraping:
  tradera:
    enabled: false
    timeout: 10s
  blocket:
    enabled: false
    timeout: 10s

llm:
  provider: "openrouter"
  api_key: "test-key"
  default_model: "openai/gpt-4o"

valuation:
  target_sell_days: 14
  min_profit_margin: 0.15
  safety_margin: 0.2

email:
  smtp_host: "localhost"
  smtp_port: "587"
  smtp_username: "test"
  smtp_password: "test"
  from: "test@example.com"
  recipients:
    - "test@example.com"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LLM.Models != nil && len(cfg.LLM.Models) != 0 {
		t.Errorf("Models count = %v, want %v", len(cfg.LLM.Models), 0)
	}
}
