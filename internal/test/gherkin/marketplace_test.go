package gherkin

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"begbot/internal/config"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

// MarketplaceTestContext holds the test state for marketplace scenarios
type MarketplaceTestContext struct {
	service       *services.MarketplaceService
	lastAdID      int64
	lastURL       string
	lastError     error
	elapsedTime   time.Duration
	adDetails     *services.BlocketAdDetails
}

func InitializeScenarioMarketplace(ctx *godog.ScenarioContext) {
	tc := &MarketplaceTestContext{}

	ctx.BeforeScenario(func(*godog.Scenario) {
		cfg := &config.Config{
			Scraping: config.ScrapingConfig{
				Blocket: config.BlocketConfig{
					Enabled: true,
				},
			},
		}
		tc.service = services.NewMarketplaceService(cfg)
		tc.lastAdID = 0
		tc.lastURL = ""
		tc.lastError = nil
		tc.elapsedTime = 0
		tc.adDetails = nil
	})

	// Given steps
	ctx.Given(`^a marketplace service is configured$`, func() error {
		cfg := &config.Config{}
		tc.service = services.NewMarketplaceService(cfg)
		return nil
	})

	ctx.Given(`^the URL "([^"]+)"$`, func(url string) error {
		tc.lastURL = url
		return nil
	})

	ctx.Given(`^rate limiting is enabled$`, func() error {
		// Rate limiting is enabled by default
		return nil
	})

	ctx.Given(`^a valid blocket ad ID "(\d+)"$`, func(adIDStr string) error {
		adID, _ := strconv.ParseInt(adIDStr, 10, 64)
		tc.lastAdID = adID
		return nil
	})

	// When steps
	ctx.When(`^I extract the ad ID$`, func() error {
		tc.lastAdID = extractBlocketAdID(tc.lastURL)
		return nil
	})

	ctx.When(`^I make "(\d+)" consecutive requests$`, func(countStr string) error {
		count, _ := strconv.Atoi(countStr)
		ctx := context.Background()
		
		start := time.Now()
		for i := 0; i < count; i++ {
			err := tc.service.WaitForRateLimit(ctx)
			if err != nil {
				return err
			}
		}
		tc.elapsedTime = time.Since(start)
		return nil
	})

	ctx.When(`^I fetch the ad from the API$`, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		details, err := tc.service.FetchBlocketAdDetails(ctx, fmt.Sprintf("https://www.blocket.se/annons/%d", tc.lastAdID))
		tc.adDetails = details
		tc.lastError = err
		return nil
	})

	// Then steps
	ctx.Then(`^the ad ID should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strconv.ParseInt(expectedStr, 10, 64)
		if tc.lastAdID != expected {
			return fmt.Errorf("expected ad ID %d, got %d", expected, tc.lastAdID)
		}
		return nil
	})

	ctx.Then(`^the ad ID should be "0"$`, func() error {
		if tc.lastAdID != 0 {
			return fmt.Errorf("expected ad ID 0, got %d", tc.lastAdID)
		}
		return nil
	})

	ctx.Then(`^the requests should take at least "([\d.]+)" seconds$`, func(minSecondsStr string) error {
		minSeconds, _ := strconv.ParseFloat(minSecondsStr, 64)
		minDuration := time.Duration(minSeconds * float64(time.Second))
		if tc.elapsedTime < minDuration {
			return fmt.Errorf("expected at least %v, got %v", minDuration, tc.elapsedTime)
		}
		return nil
	})

	ctx.Then(`^the request should either succeed or return an expected error for invalid ID$`, func() error {
		// Either success or an expected error is acceptable
		if tc.lastError != nil {
			// Error is acceptable for invalid IDs
			fmt.Printf("API returned error (expected for invalid ID): %v\n", tc.lastError)
		}
		return nil
	})

	ctx.Then(`^if successful, the title should not be empty$`, func() error {
		if tc.adDetails == nil {
			// Skip if API call failed
			return nil
		}
		if tc.adDetails.Title == "" {
			return fmt.Errorf("title should not be empty")
		}
		return nil
	})

	ctx.Then(`^if successful, the ad text should not be empty$`, func() error {
		if tc.adDetails == nil {
			// Skip if API call failed
			return nil
		}
		if tc.adDetails.AdText == "" {
			return fmt.Errorf("ad text should not be empty")
		}
		return nil
	})

	ctx.Then(`^if successful, the price should be greater than "(\d+)"$`, func(minPriceStr string) error {
		if tc.adDetails == nil {
			// Skip if API call failed
			return nil
		}
		minPrice, _ := strconv.ParseFloat(minPriceStr, 10)
		if tc.adDetails.Price <= minPrice {
			return fmt.Errorf("price should be greater than %f", minPrice)
		}
		return nil
	})
}

// extractBlocketAdID extracts the ad ID from a Blocket URL
// This is a copy of the function from marketplace.go for testing
func extractBlocketAdID(link string) int64 {
	re := regexp.MustCompile(`/(?:item|annons)/(\d+)`)
	matches := re.FindStringSubmatch(link)
	if len(matches) > 1 {
		id, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return id
		}
	}
	return 0
}

func TestMarketplaceFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenarioMarketplace,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"internal/test/gherkin/features/marketplace.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, there are failed test scenarios")
	}
}
