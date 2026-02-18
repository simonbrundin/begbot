package gherkin

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cucumber/godog"
)

type adsFilterState struct {
	listings  []listingItem
	filtered  []listingItem
	filterErr error
}

type listingItem struct {
	id        int
	price     *int
	valuation *int
}

// InitializeAdsFilteringScenario sets up the step definitions for ads filtering
func InitializeAdsFilteringScenario(ctx *godog.ScenarioContext) {
	state := &adsFilterState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &adsFilterState{
			listings: []listingItem{},
		}
	})

	ctx.Given("I have a listing filter service", func() error {
		state.listings = []listingItem{}
		state.filtered = []listingItem{}
		state.filterErr = nil
		return nil
	})

	ctx.Given("the following listings:", func(table *godog.Table) error {
		state.listings = []listingItem{}
		for _, row := range table.Rows {
			id, _ := strconv.Atoi(row.Cells[0].Value)
			price := parseNullableInt(row.Cells[1].Value)
			valuation := parseNullableInt(row.Cells[2].Value)
			state.listings = append(state.listings, listingItem{
				id:        id,
				price:     price,
				valuation: valuation,
			})
		}
		return nil
	})

	ctx.Given("there are no listings", func() error {
		state.listings = []listingItem{}
		return nil
	})

	ctx.When("I filter by {string} tab", func(tab string) error {
		state.filtered = filterListings(state.listings, tab)
		return nil
	})

	ctx.Then("I should receive {int} listings", func(count int) error {
		if len(state.filtered) != count {
			return fmt.Errorf("expected %d listings, got %d", count, len(state.filtered))
		}
		return nil
	})
}

// filterListings filters listings based on the active tab
// "all" - returns all listings
// "good-value" - returns listings where price < valuation (any valuation)
func filterListings(listings []listingItem, tab string) []listingItem {
	if tab == "all" {
		return listings
	}

	if tab == "good-value" {
		result := []listingItem{}
		for _, item := range listings {
			// A listing is "good value" if it has a price AND valuation,
			// and price is less than valuation
			if item.price != nil && item.valuation != nil && *item.price < *item.valuation {
				result = append(result, item)
			}
		}
		return result
	}

	return listings
}

// parseNullableInt parses a string to int, returns nil for empty or invalid strings
func parseNullableInt(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func TestAdsFilteringFeatures(t *testing.T) {
	featurePath := getFeaturesPath("ads_filtering.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeAdsFilteringScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run ads filtering gherkin tests")
	}
}
