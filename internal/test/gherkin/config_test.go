package gherkin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"begbot/internal/config"

	"github.com/cucumber/godog"
)

type configTestState struct {
	tmpDir      string
	configPath  string
	cfg         *config.Config
	resultErr   error
	modelsCount int
}

func InitializeConfigScenario(ctx *godog.ScenarioContext) {
	state := &configTestState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &configTestState{}
	})

	ctx.Given("temporary config directory", func() error {
		tmpDir, err := os.MkdirTemp("", "config-test")
		if err != nil {
			return err
		}
		state.tmpDir = tmpDir
		state.configPath = filepath.Join(tmpDir, "config.yaml")
		return nil
	})

	ctx.Given("a config file with LLM section containing:", func(table *godog.Table) error {
		llmConfig := ""
		for _, row := range table.Rows {
			if row.Cells[0].Value == "provider" {
				llmConfig += fmt.Sprintf("  provider: %q\n", row.Cells[1].Value)
			} else if row.Cells[0].Value == "api_key" {
				llmConfig += fmt.Sprintf("  api_key: %q\n", row.Cells[1].Value)
			} else if row.Cells[0].Value == "site_url" {
				llmConfig += fmt.Sprintf("  site_url: %q\n", row.Cells[1].Value)
			} else if row.Cells[0].Value == "site_name" {
				llmConfig += fmt.Sprintf("  site_name: %q\n", row.Cells[1].Value)
			} else if row.Cells[0].Value == "default_model" {
				llmConfig += fmt.Sprintf("  default_model: %q\n", row.Cells[1].Value)
			}
		}

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

llm:
%s
`, llmConfig)

		state.resultErr = os.WriteFile(state.configPath, []byte(configContent), 0644)
		return state.resultErr
	})

	ctx.Given("models section with:", func(table *godog.Table) error {
		modelsConfig := "  models:\n"
		for _, row := range table.Rows {
			modelsConfig += fmt.Sprintf("    %s: %q\n", row.Cells[0].Value, row.Cells[1].Value)
		}

		existingContent, err := os.ReadFile(state.configPath)
		if err != nil {
			return err
		}

		newContent := string(existingContent) + modelsConfig
		state.resultErr = os.WriteFile(state.configPath, []byte(newContent), 0644)
		return state.resultErr
	})

	ctx.Given("no models section", func() error {
		existingContent, err := os.ReadFile(state.configPath)
		if err != nil {
			return err
		}
		state.resultErr = os.WriteFile(state.configPath, existingContent, 0644)
		return state.resultErr
	})

	ctx.When("I load the config", func() error {
		state.cfg, state.resultErr = config.Load(state.configPath)
		if state.cfg != nil && state.cfg.LLM.Models != nil {
			state.modelsCount = len(state.cfg.LLM.Models)
		}
		return nil
	})

	ctx.Then("LLM provider should be \"openrouter\"", func() error {
		if state.cfg == nil || state.cfg.LLM.Provider != "openrouter" {
			return fmt.Errorf("expected provider 'openrouter', got '%s'", state.cfg.LLM.Provider)
		}
		return nil
	})

	ctx.Then("LLM api_key should be \"test-key\"", func() error {
		if state.cfg == nil || state.cfg.LLM.APIKey != "test-key" {
			return fmt.Errorf("expected api_key 'test-key', got '%s'", state.cfg.LLM.APIKey)
		}
		return nil
	})

	ctx.Then("LLM site_url should be \"http://localhost:3000\"", func() error {
		if state.cfg == nil || state.cfg.LLM.SiteURL != "http://localhost:3000" {
			return fmt.Errorf("expected site_url 'http://localhost:3000', got '%s'", state.cfg.LLM.SiteURL)
		}
		return nil
	})

	ctx.Then("LLM site_name should be \"Begbot\"", func() error {
		if state.cfg == nil || state.cfg.LLM.SiteName != "Begbot" {
			return fmt.Errorf("expected site_name 'Begbot', got '%s'", state.cfg.LLM.SiteName)
		}
		return nil
	})

	ctx.Then("LLM default_model should be \"deepseek/deepseek-v3.2\"", func() error {
		if state.cfg == nil || state.cfg.LLM.DefaultModel != "deepseek/deepseek-v3.2" {
			return fmt.Errorf("expected default_model 'deepseek/deepseek-v3.2', got '%s'", state.cfg.LLM.DefaultModel)
		}
		return nil
	})

	ctx.Then("there should be 4 models configured", func() error {
		if state.modelsCount != 4 {
			return fmt.Errorf("expected 4 models, got %d", state.modelsCount)
		}
		return nil
	})

	ctx.Then("models should be empty or nil", func() error {
		if state.cfg != nil && state.cfg.LLM.Models != nil && len(state.cfg.LLM.Models) != 0 {
			return fmt.Errorf("expected empty models, got %d", len(state.cfg.LLM.Models))
		}
		return nil
	})
}

func TestConfigFeatures(t *testing.T) {
	featurePath := getFeaturesPath("config.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeConfigScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run config gherkin tests")
	}
}
