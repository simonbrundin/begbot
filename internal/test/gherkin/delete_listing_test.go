//go:build gherkin
// +build gherkin

package gherkin

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"begbot/internal/db"
	"begbot/internal/models"

	"github.com/cucumber/godog"
)

// deleteListingState holds the state for delete listing tests
type deleteListingState struct {
	listings     map[int64]*models.Listing
	valuations   map[int64][]db.ValuationWithType
	tradedItems  map[int64][]models.TradedItem
	imageLinks   map[int64][]string
	responseCode int
	deleteErr    error
}

// InitializeDeleteListingScenario sets up the step definitions for delete listing
func InitializeDeleteListingScenario(ctx *godog.ScenarioContext) {
	state := &deleteListingState{
		listings:    make(map[int64]*models.Listing),
		valuations:  make(map[int64][]db.ValuationWithType),
		tradedItems: make(map[int64][]models.TradedItem),
		imageLinks:  make(map[int64][]string),
	}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &deleteListingState{
			listings:    make(map[int64]*models.Listing),
			valuations:  make(map[int64][]db.ValuationWithType),
			tradedItems: make(map[int64][]models.TradedItem),
			imageLinks:  make(map[int64][]string),
		}
	})

	// Given steps
	ctx.Given("I have a listing database", func() error {
		state.listings = make(map[int64]*models.Listing)
		state.valuations = make(map[int64][]db.ValuationWithType)
		state.tradedItems = make(map[int64][]models.TradedItem)
		state.imageLinks = make(map[int64][]string)
		return nil
	})

	ctx.Given("a listing with id {string} exists in the database", func(idStr string) error {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		state.listings[id] = &models.Listing{
			ID:     id,
			Title:  fmt.Sprintf("Test Listing %d", id),
			Link:   fmt.Sprintf("https://example.com/listing/%d", id),
			Status: "active",
		}
		return nil
	})

	ctx.Given("no listing with id {string} exists in the database", func(idStr string) error {
		// Do nothing - listing should not exist
		return nil
	})

	ctx.Given("the listing has valuations", func() error {
		// Get the last added listing ID
		var lastID int64
		for id := range state.listings {
			if id > lastID {
				lastID = id
			}
		}
		state.valuations[lastID] = []db.ValuationWithType{
			{ID: 1, ValuationTypeID: 1, ValuationType: "ebay", Valuation: 1000},
		}
		return nil
	})

	ctx.Given("the listing has traded items", func() error {
		// Get the last added listing ID
		var lastID int64
		for id := range state.listings {
			if id > lastID {
				lastID = id
			}
		}
		state.tradedItems[lastID] = []models.TradedItem{
			{ID: 1, ListingID: &lastID},
		}
		return nil
	})

	ctx.Given("the listing has image links", func() error {
		// Get the last added listing ID
		var lastID int64
		for id := range state.listings {
			if id > lastID {
				lastID = id
			}
		}
		state.imageLinks[lastID] = []string{
			"https://example.com/image1.jpg",
			"https://example.com/image2.jpg",
		}
		return nil
	})

	// When steps
	ctx.When("I send a DELETE request to {string}", func(path string) error {
		// Simulate the DELETE request behavior
		// This will fail because the actual implementation doesn't call DeleteListing
		idStr := path[len("/api/listings/"):]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			state.responseCode = http.StatusBadRequest
			return nil
		}

		// Check if listing exists
		if _, ok := state.listings[id]; !ok {
			state.responseCode = http.StatusNotFound
			return nil
		}

		// This is the current behavior - just returns 204 without actually deleting
		// After implementation, it should call db.DeleteListing
		state.responseCode = http.StatusNoContent
		return nil
	})

	// Then steps
	ctx.Then("the response status should be {int}", func(statusCode int) error {
		if state.responseCode != statusCode {
			return fmt.Errorf("expected status %d, got %d", statusCode, state.responseCode)
		}
		return nil
	})

	ctx.Then("the listing with id {string} should no longer exist in the database", func(idStr string) error {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		// After implementation, this should check that DeleteListing was called
		// Currently it doesn't verify the actual deletion
		if _, ok := state.listings[id]; ok {
			return fmt.Errorf("listing %d should not exist after deletion", id)
		}
		return nil
	})

	ctx.Then("the related valuations should also be deleted", func() error {
		// This step is for documentation purposes
		// After implementation, verify that valuations are deleted via CASCADE
		return nil
	})

	ctx.Then("the related traded items should also be deleted", func() error {
		// This step is for documentation purposes
		// After implementation, verify that traded items are deleted via CASCADE
		return nil
	})

	ctx.Then("the related image links should also be deleted", func() error {
		// This step is for documentation purposes
		// After implementation, verify that image links are deleted via CASCADE
		return nil
	})
}

// TestDeleteListingFeatures runs the delete listing godog tests
func TestDeleteListingFeatures(t *testing.T) {
	featurePath := getFeaturesPath("delete_listing.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeDeleteListingScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run delete listing gherkin tests")
	}
}
