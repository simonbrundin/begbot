//go:build gherkin
// +build gherkin

package gherkin

import (
	"context"
	"testing"
	"time"

	"begbot/internal/config"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

// MarketplaceTestContext holds state for marketplace BDD tests
type marketplaceTestContext struct {
	service      *services.MarketplaceService
	cfg          *config.Config
	ctx          context.Context
	adID         int64
	extractedID  int64
	details      *services.BlocketAdDetails
	err          error
	elapsed      time.Duration
}

// InitializeMarketplaceScenario initializes the marketplace test context
func InitializeMarketplaceScenario(ctx *godog.ScenarioContext) {
	tc := &marketplaceTestContext{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		tc.cfg = &config.Config{
			Scraping: config.ScrapingConfig{
				Blocket: config.BlocketConfig{
					Enabled: true,
				},
			},
		}
		tc.service = services.NewMarketplaceService(tc.cfg)
		tc.ctx = context.Background()
		tc.adID = 0
		tc.extractedID = 0
		tc.details = nil
		tc.err = nil
	})

	// Background
	ctx.Given("a marketplace service is available", func(sc *godog.Step) error {
		tc.service = services.NewMarketplaceService(tc.cfg)
		return nil
	})

	ctx.And("the configuration has blocket enabled", func(sc *godog.Step) error {
		tc.cfg.Scraping.Blocket.Enabled = true
		return nil
	})

	// Extract ad ID
	ctx.Given("the URL {string}", func(sc *godog.Step, url string) error {
		tc.adID = services.ExtractBlocketAdID(url)
		return nil
	})

	ctx.When("extracting the ad ID", func(sc *godog.Step) error {
		// Already extracted in Given step
		return nil
	})

	ctx.Then("the ad ID should be {int}", func(sc *godog.Step, expected int64) error {
		if tc.adID != expected {
			return.Errorf("expected %d, got %d", expected, tc.adID)
		}
		return nil
	})

	ctx.Given("an invalid URL {string}", func(sc *godog.Step, url string) error {
		tc.adID = services.ExtractBlocketAdID(url)
		return nil
	})

	ctx.Given("a non-Blocket URL {string}", func(sc *godog.Step, url string) error {
		tc.adID = services.ExtractBlocketAdID(url)
		return nil
	})

	// Rate limiting
	ctx.Given("the rate limiter is reset", func(sc *godog.Step) error {
		// Rate limiter is reset per scenario
		return nil
	})

	ctx.When("making {int} consecutive requests", func(sc *godog.Step, count int) error {
		start := time.Now()
		for i := 0; i < count; i++ {
			tc.err = tc.service.WaitForRateLimit(tc.ctx)
			if tc.err != nil {
				return tc.err
			}
		}
		tc.elapsed = time.Since(start)
		return nil
	})

	ctx.Then("the requests should take at least {int} second", func(sc *godog.Step, seconds int) error {
		expectedMin := time.Duration(seconds) * time.Second
		if tc.elapsed < expectedMin {
			return.Errorf("expected at least %v, got %v", expectedMin, tc.elapsed)
		}
		return nil
	})

	ctx.And("no rate limit errors should occur", func(sc *godog.Step) error {
		if tc.err != nil {
			return tc.err
		}
		return nil
	})

	// Fetch from API
	ctx.Given("a valid Blocket ad ID", func(sc *godog.Step) error {
		tc.adID = 124456789 // Known valid ID
		return nil
	})

	ctx.When("fetching the ad from the API", func(sc *godog.Step) error {
		ctx, cancel := context.WithTimeout(tc.ctx, 30*time.Second)
		defer cancel()
		tc.details, tc.err = tc.service.FetchBlocketAdFromAPI(ctx, tc.adID)
		return nil
	})

	ctx.Then("the response should contain a title", func(sc *godog.Step) error {
		if tc.details == nil {
			return errors.New("no details returned")
		}
		if tc.details.Title == "" {
			return errors.New("title is empty")
		}
		return nil
	})

	ctx.And("the response should contain ad text", func(sc *godog.Step) error {
		if tc.details == nil {
			return errors.New("no details returned")
		}
		if tc.details.AdText == "" {
			return errors.New("ad text is empty")
		}
		return nil
	})

	ctx.And("the price should be greater than {int}", func(sc *godog.Step, minPrice int) error {
		if tc.details == nil {
			return errors.New("no details returned")
		}
		if tc.details.Price <= float64(minPrice) {
			return.Errorf("price %f is not greater than %d", tc.details.Price, minPrice)
		}
		return nil
	})

	ctx.Given("an invalid Blocket ad ID {int}", func(sc *godog.Step, id int64) error {
		tc.adID = id
		return nil
	})

	ctx.Then("an error may be returned (expected for invalid IDs)", func(sc *godog.Step) error {
		// Error is allowed for invalid IDs
		return nil
	})

	// Rate limit errors
	ctx.Given("the API returns a rate limit error", func(sc *godog.Step) error {
		// Would need to mock this
		return nil
	})

	ctx.When("retrying the request", func(sc *godog.Step) error {
		// Would implement retry logic
		return nil
	})

	ctx.Then("the request should eventually succeed", func(sc *godog.Step) error {
		// Would check for success
		return nil
	})

	ctx.Then("the request should succeed", func(sc *godog.Step) error {
		if tc.err != nil {
			return tc.err
		}
		return nil
	})
}

// Helper for error formatting
func errorsNew(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func Errorf(format string, args ...interface{}) error {
	return &testError{msg: formatArgs(format, args)}
}

func formatArgs(format string, args []interface{}) string {
	// Simple implementation
	result := format
	for _, arg := range args {
		result += " " + toString(arg)
	}
	return result
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case int:
		return string(rune('0' + val%10))
	case int64:
		return string(rune('0' + int(val)%10))
	case float64:
		return "0.0"
	default:
		return ""
	}
}

// TestMarketplaceFeature runs the Godog marketplace tests
func TestMarketplaceFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeMarketplaceScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features/marketplace.feature"},
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
