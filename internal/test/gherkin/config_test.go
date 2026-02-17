package gherkin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"begbot/internal/config"

	"github.com/cucumber/godog"
)

// ConfigTestContext holds the test state for config scenarios
type ConfigTestContext struct {
	config      *config.Config
	lastError   error
	configPath  string
	tempDir     string
}

func InitializeScenarioConfig(ctx *godog.ScenarioContext) {
	tc := &ConfigTestContext{}

	ctx.BeforeScenario(func(*godog.Scenario) error {
		// Create temporary directory for config files
		tmpDir, err := os.MkdirTemp("", "gherkin-config-test")
		if err != nil {
			return err
		}
		tc.tempDir = tmpDir
		tc.config = nil
		tc.lastError = nil
		return nil
	})

	ctx.AfterScenario(func(*godog.Scenario, error) {
		// Clean up temporary directory
		if tc.tempDir != "" {
			os.RemoveAll(tc.tempDir)
		}
	})

	// Given steps
	ctx.Given(`^a temporary directory for config files$`, func() error {
		// Already created in BeforeScenario
		return nil
	})

	ctx.Given(`^a config file with LLM models:$`, func(modelsContent string) error {
		configContent := fmt.Sprintf(`
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
  %s

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
`, modelsContent)

		tc.configPath = filepath.Join(tc.tempDir, "config.yaml")
		if err := os.WriteFile(tc.configPath, []byte(configContent), 0644); err != nil {
			return err
		}
		return nil
	})

	ctx.Given(`^a config file with LLM but no models:$`, func(llmContent string) error {
		configContent := fmt.Sprintf(`
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
  %s

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
`, llmContent)

		tc.configPath = filepath.Join(tc.tempDir, "config.yaml")
		if err := os.WriteFile(tc.configPath, []byte(configContent), 0644); err != nil {
			return err
		}
		return nil
	})

	// When steps
	ctx.When(`^I load the configuration$`, func() error {
		tc.config, tc.lastError = config.Load(tc.configPath)
		return nil
	})

	// Then steps
	ctx.Then(`^the LLM provider should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.Provider != expected {
			return fmt.Errorf("expected provider '%s', got '%s'", expected, tc.config.LLM.Provider)
		}
		return nil
	})

	ctx.Then(`^the API key should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.APIKey != expected {
			return fmt.Errorf("expected API key '%s', got '%s'", expected, tc.config.LLM.APIKey)
		}
		return nil
	})

	ctx.Then(`^the site URL should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.SiteURL != expected {
			return fmt.Errorf("expected site URL '%s', got '%s'", expected, tc.config.LLM.SiteURL)
		}
		return nil
	})

	ctx.Then(`^the site name should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.SiteName != expected {
			return fmt.Errorf("expected site name '%s', got '%s'", expected, tc.config.LLM.SiteName)
		}
		return nil
	})

	ctx.Then(`^the default model should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.DefaultModel != expected {
			return fmt.Errorf("expected default model '%s', got '%s'", expected, tc.config.LLM.DefaultModel)
		}
		return nil
	})

	ctx.Then(`^the number of models should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		actual := len(tc.config.LLM.Models)
		if actual != expected {
			return fmt.Errorf("expected %d models, got %d", expected, actual)
		}
		return nil
	})

	ctx.Then(`^the ExtractProductInfo model should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.Models["ExtractProductInfo"] != expected {
			return fmt.Errorf("expected ExtractProductInfo '%s', got '%s'", expected, tc.config.LLM.Models["ExtractProductInfo"])
		}
		return nil
	})

	ctx.Then(`^the ValidateProduct model should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.Models["ValidateProduct"] != expected {
			return fmt.Errorf("expected ValidateProduct '%s', got '%s'", expected, tc.config.LLM.Models["ValidateProduct"])
		}
		return nil
	})

	ctx.Then(`^the CheckProductCondition model should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.Models["CheckProductCondition"] != expected {
			return fmt.Errorf("expected CheckProductCondition '%s', got '%s'", expected, tc.config.LLM.Models["CheckProductCondition"])
		}
		return nil
	})

	ctx.Then(`^the EstimateNewPrice model should be "([^"]+)"$`, func(expected string) error {
		if tc.config.LLM.Models["EstimateNewPrice"] != expected {
			return fmt.Errorf("expected EstimateNewPrice '%s', got '%s'", expected, tc.config.LLM.Models["EstimateNewPrice"])
		}
		return nil
	})
}

func TestConfigFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenarioConfig,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"internal/test/gherkin/features/config.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, there are failed test scenarios")
	}
}
