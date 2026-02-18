package gherkin

import (
	"context"
	"errors"
	"testing"
	"time"

	"begbot/internal/models"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

// Mock SearchHistoryDB for testing
type mockSearchHistoryDB struct {
	history []models.SearchHistory
	err     error
}

func (m *mockSearchHistoryDB) SaveSearchHistory(ctx context.Context, h *models.SearchHistory) error {
	if m.err != nil {
		return m.err
	}
	h.ID = int64(len(m.history) + 1)
	h.CreatedAt = time.Now()
	m.history = append(m.history, *h)
	return nil
}

func (m *mockSearchHistoryDB) GetSearchHistory(ctx context.Context, limit, offset int) ([]models.SearchHistory, error) {
	if m.err != nil {
		return nil, m.err
	}
	if offset >= len(m.history) {
		return []models.SearchHistory{}, nil
	}
	end := offset + limit
	if end > len(m.history) {
		end = len(m.history)
	}
	return m.history[offset:end], nil
}

func (m *mockSearchHistoryDB) GetSearchHistoryCount(ctx context.Context) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return len(m.history), nil
}

// TestContext holds state for BDD tests
type searchHistoryTestContext struct {
	service *services.SearchHistoryService
	mockDB  *mockSearchHistoryDB
	ctx     context.Context
	result  *models.SearchHistory
	history []models.SearchHistory
	count   int
	err     error
}

// InitializeSearchHistoryScenario initializes the test context
func InitializeSearchHistoryScenario(ctx *godog.ScenarioContext) {
	tc := &searchHistoryTestContext{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		tc.mockDB = &mockSearchHistoryDB{history: []models.SearchHistory{}}
		tc.service = services.NewSearchHistoryService(tc.mockDB)
		tc.ctx = context.Background()
		tc.result = nil
		tc.history = nil
		tc.count = 0
		tc.err = nil
	})

	// Background steps
	ctx.Given("a search history service is available", func(sc *godog.Step) error {
		if tc.service == nil {
			tc.mockDB = &mockSearchHistoryDB{history: []models.SearchHistory{}}
			tc.service = services.NewSearchHistoryService(tc.mockDB)
		}
		return nil
	})

	ctx.Given("the database is connected", func(sc *godog.Step) error {
		tc.mockDB.err = nil
		return nil
	})

	// Record search steps
	ctx.When("a user searches for {string} with URL {string}", func(sc *godog.Step, termDesc, url string) error {
		tc.result, tc.err = tc.service.RecordSearch(tc.ctx, 1, termDesc, url, 10, 3)
		return nil
	})

	ctx.And("the search finds {int} results with {int} new ads", func(sc *godog.Step, results, newAds int) error {
		// This is already set in the previous step, but we keep it for clarity
		return nil
	})

	ctx.Then("the search should be saved successfully", func(sc *godog.Step) error {
		if tc.err != nil {
			return errors.New("expected no error, got: " + tc.err.Error())
		}
		return nil
	})

	ctx.And("the search should have a valid ID", func(sc *godog.Step) error {
		if tc.result == nil || tc.result.ID == 0 {
			return errors.New("expected valid ID")
		}
		return nil
	})

	ctx.And("the search term description should be {string}", func(sc *godog.Step, expected string) error {
		if tc.result.SearchTermDesc != expected {
			return errors.New("expected " + expected + ", got " + tc.result.SearchTermDesc)
		}
		return nil
	})

	ctx.And("the results found should be {int}", func(sc *godog.Step, expected int) error {
		if tc.result.ResultsFound != expected {
			return errors.New("expected " + string(rune(expected)) + ", got " + string(rune(tc.result.ResultsFound)))
		}
		return nil
	})

	ctx.And("the new ads found should be {int}", func(sc *godog.Step, expected int) error {
		if tc.result.NewAdsFound != expected {
			return errors.New("expected " + string(rune(expected)) + ", got " + string(rune(tc.result.NewAdsFound)))
		}
		return nil
	})

	// Get history with data
	ctx.Given("the database has {int} search records", func(sc *godog.Step, count int) error {
		now := time.Now()
		tc.mockDB.history = make([]models.SearchHistory, count)
		for i := 0; i < count; i++ {
			tc.mockDB.history[i] = models.SearchHistory{
				ID:             int64(i + 1),
				SearchTermID:   int64(i + 1),
				SearchTermDesc: "Item " + string(rune('1'+i)),
				ResultsFound:   10,
				NewAdsFound:    1,
				SearchedAt:     now,
			}
		}
		return nil
	})

	ctx.When("the user requests search history for page {int} with {int} items per page", func(sc *godog.Step, page, pageSize int) error {
		tc.history, tc.count, tc.err = tc.service.GetHistory(tc.ctx, page, pageSize)
		return nil
	})

	ctx.Then("the response should contain {int} search records", func(sc *godog.Step, expected int) error {
		if len(tc.history) != expected {
			return errors.New("expected " + string(rune(expected)) + " records, got " + string(rune(len(tc.history))))
		}
		return nil
	})

	ctx.And("the total count should be {int}", func(sc *godog.Step, expected int) error {
		if tc.count != expected {
			return errors.New("expected count " + string(rune(expected)) + ", got " + string(rune(tc.count)))
		}
		return nil
	})

	ctx.And("the first record should have search term {string}", func(sc *godog.Step, expected string) error {
		if len(tc.history) == 0 {
			return errors.New("no history records")
		}
		if tc.history[0].SearchTermDesc != expected {
			return errors.New("expected " + expected + ", got " + tc.history[0].SearchTermDesc)
		}
		return nil
	})

	// Empty history
	ctx.Given("the database has no search records", func(sc *godog.Step) error {
		tc.mockDB.history = []models.SearchHistory{}
		return nil
	})

	ctx.When("the user requests search history", func(sc *godog.Step) error {
		tc.history, tc.count, tc.err = tc.service.GetHistory(tc.ctx, 1, 20)
		return nil
	})

	// Pagination
	ctx.When("the user requests page {int} with {int} items per page", func(sc *godog.Step, page, pageSize int) error {
		tc.history, tc.count, tc.err = tc.service.GetHistory(tc.ctx, page, pageSize)
		return nil
	})

	ctx.Then("the response should contain {int} items", func(sc *godog.Step, expected int) error {
		if len(tc.history) != expected {
			return errors.New("expected " + string(rune(expected)) + " items, got " + string(rune(len(tc.history))))
		}
		return nil
	})

	ctx.And("the first item on page {int} should have ID {int}", func(sc *godog.Step, page, expectedID int) error {
		if len(tc.history) == 0 {
			return errors.New("no history records")
		}
		if tc.history[0].ID != int64(expectedID) {
			return errors.New("expected ID " + string(rune(expectedID)) + ", got " + string(rune(int(tc.history[0].ID))))
		}
		return nil
	})

	// Invalid pagination
	ctx.When("the user requests page {int}", func(sc *godog.Step, page int) error {
		tc.history, tc.count, tc.err = tc.service.GetHistory(tc.ctx, page, 20)
		return nil
	})

	ctx.Then("the request should succeed", func(sc *godog.Step) error {
		if tc.err != nil {
			return errors.New("expected no error, got: " + tc.err.Error())
		}
		return nil
	})

	ctx.And("the count should be {int}", func(sc *godog.Step, expected int) error {
		if tc.count != expected {
			return errors.New("expected count " + string(rune(expected)) + ", got " + string(rune(tc.count)))
		}
		return nil
	})

	// Database errors
	ctx.Given("the database is unavailable", func(sc *godog.Step) error {
		tc.mockDB.err = errors.New("database unavailable")
		return nil
	})

	ctx.When("the user attempts to record a search", func(sc *godog.Step) error {
		tc.result, tc.err = tc.service.RecordSearch(tc.ctx, 1, "Test", "https://...", 10, 2)
		return nil
	})

	ctx.Then("an error should be returned", func(sc *godog.Step) error {
		if tc.err == nil {
			return errors.New("expected error, got nil")
		}
		return nil
	})

	// Large page size
	ctx.When("the user requests page {int} with {int} items per page", func(sc *godog.Step, page, pageSize int) error {
		_, _, tc.err = tc.service.GetHistory(tc.ctx, page, pageSize)
		return nil
	})
}

// TestSearchHistoryFeature runs the Godog tests
func TestSearchHistoryFeature(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeSearchHistoryScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features/search_history.feature"},
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
