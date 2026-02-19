//go:build gherkin
// +build gherkin

package gherkin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"begbot/internal/config"

	"github.com/cucumber/godog"
)

// ConfigTestContext holds state for config BDD tests
type configTestContext struct {
	cfg          *config.Config
	err          error
	configPath   string
	tmpDir       string
}

// InitializeConfigScenario initializes the config test context
func InitializeConfigScenario(ctx *godog.ScenarioContext) {
	tc := &configTestContext{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		tc.cfg = nil
		tc.err = nil
		tc.configPath = ""
	})

	// Background
	ctx.Given("a configuration system is available", func(sc *godog.Step) error {
		return nil
	})

	// LLM with models
	ctx.Given("a config file with provider {string}", func(sc *godog.Step, provider string) error {
		d, err := os.MkdirTemp("", "cfgtest")
		if err != nil {
			return err
		}
		tc.tmpDir = d
		tc.configPath = filepath.Join(tc.tmpDir, "config.yaml")
		return writeTestConfig(tc.configPath, "provider: "+provider+"\n")
	})

	ctx.And("API key {string}", func(sc *godog.Step, apiKey string) error {
		// Would need to append to config
		return nil
	})

	ctx.And("site URL {string}", func(sc *godog.Step, siteURL string) error {
		// Would need to append to config
		return nil
	})

	ctx.And("site name {string}", func(sc *godog.Step, siteName string) error {
		// Would need to append to config
		return nil
	})

	ctx.And("default model {string}", func(sc *godog.Step, model string) error {
		// Would need to append to config
		return nil
	})

	ctx.And("models:", func(sc *godog.Step, table *godog.Table) error {
		// Would parse table and create config
		return nil
	})

	ctx.When("loading the configuration", func(sc *godog.Step) error {
		tc.cfg, tc.err = config.Load(tc.configPath)
		return nil
	})

	ctx.Then("the provider should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.LLM.Provider != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.LLM.Provider)
		}
		return nil
	})

	ctx.And("the API key should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.LLM.APIKey != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.LLM.APIKey)
		}
		return nil
	})

	ctx.And("the site URL should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.LLM.SiteURL != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.LLM.SiteURL)
		}
		return nil
	})

	ctx.And("the site name should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.LLM.SiteName != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.LLM.SiteName)
		}
		return nil
	})

	ctx.And("the default model should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.LLM.DefaultModel != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.LLM.DefaultModel)
		}
		return nil
	})

	ctx.And("there should be {int} models defined", func(sc *godog.Step, expected int) error {
		if len(tc.cfg.LLM.Models) != expected {
			return fmt.Errorf("expected %d models, got %d", expected, len(tc.cfg.LLM.Models))
		}
		return nil
	})

	// LLM without models
	ctx.And("no models defined", func(sc *godog.Step) error {
		// Would create config without models section
		return nil
	})

	ctx.Then("the models count should be {int}", func(sc *godog.Step, expected int) error {
		if tc.cfg.LLM.Models != nil && len(tc.cfg.LLM.Models) != expected {
			return fmt.Errorf("expected %d models, got %d", expected, len(tc.cfg.LLM.Models))
		}
		return nil
	})

	// Database settings
	ctx.Given("a config file with:", func(sc *godog.Step, table *godog.Table) error {
		// Would parse table and create config
		d, err := os.MkdirTemp("", "cfgtest")
		if err != nil {
			return err
		}
		tc.tmpDir = d
		tc.configPath = filepath.Join(tc.tmpDir, "config.yaml")
		
		configContent := "database:\n"
		for _, row := range table.Rows {
			configContent += "  " + row.Cells[0].Value + ": " + row.Cells[1].Value + "\n"
		}
		
		return os.WriteFile(tc.configPath, []byte(configContent), 0644)
	})

	ctx.Then("the database host should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.Database.Host != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.Database.Host)
		}
		return nil
	})

	ctx.And("the database port should be {int}", func(sc *godog.Step, expected int) error {
		if tc.cfg.Database.Port != expected {
			return fmt.Errorf("expected %d, got %d", expected, tc.cfg.Database.Port)
		}
		return nil
	})

	// Scraping settings
	ctx.Given("a config file with scraping:", func(sc *godog.Step, table *godog.Table) error {
		// Would parse table
		return nil
	})

	ctx.Then("tradera should be disabled", func(sc *godog.Step) error {
		if tc.cfg.Scraping.Tradera.Enabled != false {
			return errors.New("tradera should be disabled")
		}
		return nil
	})

	ctx.Then("blocklet should be disabled", func(sc *godog.Step) error {
		if tc.cfg.Scraping.Blocket.Enabled != false {
			return errors.New("blocklet should be disabled")
		}
		return nil
	})

	// Valuation settings
	ctx.Given("a config file with valuation:", func(sc *godog.Step, table *godog.Table) error {
		// Would parse table
		return nil
	})

	ctx.Then("the target sell days should be {int}", func(sc *godog.Step, expected int) error {
		if tc.cfg.Valuation.TargetSellDays != expected {
			return fmt.Errorf("expected %d, got %d", expected, tc.cfg.Valuation.TargetSellDays)
		}
		return nil
	})

	ctx.And("the minimum profit margin should be {float}", func(sc *godog.Step, expected float64) error {
		if tc.cfg.Valuation.MinProfitMargin != expected {
			return fmt.Errorf("expected %f, got %f", expected, tc.cfg.Valuation.MinProfitMargin)
		}
		return nil
	})

	ctx.And("the safety margin should be {float}", func(sc *godog.Step, expected float64) error {
		if tc.cfg.Valuation.SafetyMargin != expected {
			return fmt.Errorf("expected %f, got %f", expected, tc.cfg.Valuation.SafetyMargin)
		}
		return nil
	})

	// Email settings
	ctx.Given("a config file with email:", func(sc *godog.Step, table *godog.Table) error {
		// Would parse table
		return nil
	})

	ctx.Then("the SMTP host should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.Email.SMTPHost != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.Email.SMTPHost)
		}
		return nil
	})

	ctx.And("the SMTP port should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.Email.SMTPPort != expected {
			return fmt.Errorf("expected %s, got %s", expected, tc.cfg.Email.SMTPPort)
		}
		return nil
	})

	ctx.And("the from address should be {string}", func(sc *godog.Step, expected string) error {
		if tc.cfg.Email.From != expected {
			return.Errorf("expected %s, got %s", expected, tc.cfg.Email.From)
		}
		return nil
	})

	// Error cases
	ctx.Given("a non-existent config file", func(sc *godog.Step) error {
		tc.configPath = "/nonexistent/path/config.yaml"
		return nil
	})

	ctx.Then("an error should be returned", func(sc *godog.Step) error {
		if tc.err == nil {
			return errors.New("expected error, got nil")
		}
		return nil
	})

	ctx.Given("a config file with invalid YAML", func(sc *godog.Step) error {
		tc.tmpDir = t.TempDir()
		tc.configPath = filepath.Join(tc.tmpDir, "invalid.yaml")
		return os.WriteFile(tc.configPath, []byte("invalid: yaml: content:["), 0644)
	})
}

// Helper to write test config
func writeTestConfig(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// TestConfigFeature runs the Godog config tests
func TestConfigFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeConfigScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features/config.feature"},
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
