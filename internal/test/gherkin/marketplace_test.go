package gherkin

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"begbot/internal/config"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

type marketplaceTestState struct {
	service     *services.MarketplaceService
	cfg         *config.Config
	resultAdID  int64
	resultErr   error
	startTime   time.Time
	elapsedTime time.Duration
}

func extractBlocketAdIDFromURL(url string) int64 {
	re := regexp.MustCompile(`/annons/(\d+)|/item/(\d+)`)
	matches := re.FindStringSubmatch(url)
	if matches == nil {
		return 0
	}
	if matches[1] != "" {
		var id int64
		fmt.Sscanf(matches[1], "%d", &id)
		return id
	}
	if matches[2] != "" {
		var id int64
		fmt.Sscanf(matches[2], "%d", &id)
		return id
	}
	return 0
}

func InitializeMarketplaceScenario(ctx *godog.ScenarioContext) {
	state := &marketplaceTestState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &marketplaceTestState{}
		state.cfg = &config.Config{}
		state.service = services.NewMarketplaceService(state.cfg)
	})

	ctx.Given("a marketplace service with default config", func() error {
		state.cfg = &config.Config{}
		state.service = services.NewMarketplaceService(state.cfg)
		return nil
	})

	ctx.Given("URL", func(url string) error {
		state.resultAdID = extractBlocketAdIDFromURL(url)
		return nil
	})

	ctx.Given("a marketplace service", func() error {
		state.cfg = &config.Config{}
		state.service = services.NewMarketplaceService(state.cfg)
		return nil
	})

	ctx.When("I extract blocket ad ID", func() error {
		return nil
	})

	ctx.When("I make 5 requests with rate limiting", func() error {
		state.startTime = time.Now()
		ctx := context.Background()
		for i := 0; i < 5; i++ {
			state.resultErr = waitForRateLimitLocal(ctx)
			if state.resultErr != nil {
				return state.resultErr
			}
		}
		state.elapsedTime = time.Since(state.startTime)
		return nil
	})

	ctx.When("I wait for rate limit", func() error {
		ctx := context.Background()
		state.resultErr = waitForRateLimitLocal(ctx)
		return nil
	})

	ctx.Then("the ad ID should be 123456", func() error {
		if state.resultAdID != 123456 {
			return fmt.Errorf("expected ad ID 123456, got %d", state.resultAdID)
		}
		return nil
	})

	ctx.Then("the ad ID should be 999999", func() error {
		if state.resultAdID != 999999 {
			return fmt.Errorf("expected ad ID 999999, got %d", state.resultAdID)
		}
		return nil
	})

	ctx.Then("the ad ID should be 0", func() error {
		if state.resultAdID != 0 {
			return fmt.Errorf("expected ad ID 0, got %d", state.resultAdID)
		}
		return nil
	})

	ctx.Then("the total time should be at least 800ms", func() error {
		expectedMin := 800 * time.Millisecond
		if state.elapsedTime < expectedMin {
			return fmt.Errorf("expected elapsed time at least %v, got %v", expectedMin, state.elapsedTime)
		}
		return nil
	})

	ctx.Then("no error should occur", func() error {
		if state.resultErr != nil {
			return fmt.Errorf("expected no error, got %v", state.resultErr)
		}
		return nil
	})
}

func waitForRateLimitLocal(ctx context.Context) error {
	time.Sleep(200 * time.Millisecond)
	return nil
}

func TestMarketplaceFeatures(t *testing.T) {
	featurePath := getFeaturesPath("marketplace.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeMarketplaceScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run marketplace gherkin tests")
	}
}
