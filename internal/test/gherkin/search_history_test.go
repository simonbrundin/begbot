package gherkin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"begbot/internal/models"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

type searchHistoryTestState struct {
	service       *services.SearchHistoryService
	mockDB        *mockSearchHistoryDB
	resultHistory []models.SearchHistory
	resultCount   int
	resultErr     error
	resultEmpty   bool
}

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

func InitializeSearchHistoryScenario(ctx *godog.ScenarioContext) {
	state := &searchHistoryTestState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &searchHistoryTestState{
			mockDB: &mockSearchHistoryDB{history: []models.SearchHistory{}},
		}
		state.service = services.NewSearchHistoryService(state.mockDB)
	})

	ctx.Given("a search history service with mock database", func() error {
		state.mockDB = &mockSearchHistoryDB{history: []models.SearchHistory{}}
		state.service = services.NewSearchHistoryService(state.mockDB)
		return nil
	})

	ctx.Given("the database has search history records", func(table *godog.Table) error {
		now := time.Now()
		for _, row := range table.Rows {
			id := 0
			fmt.Sscanf(row.Cells[0].Value, "%d", &id)
			termID := 0
			fmt.Sscanf(row.Cells[1].Value, "%d", &termID)
			resultsFound := 0
			fmt.Sscanf(row.Cells[4].Value, "%d", &resultsFound)
			newAdsFound := 0
			fmt.Sscanf(row.Cells[5].Value, "%d", &newAdsFound)
			state.mockDB.history = append(state.mockDB.history, models.SearchHistory{
				ID:              int64(id),
				SearchTermID:    int64(termID),
				SearchTermDesc:  row.Cells[2].Value,
				URL:             row.Cells[3].Value,
				ResultsFound:    resultsFound,
				NewAdsFound:     newAdsFound,
				MarketplaceName: "Blocket",
				SearchedAt:      now,
			})
		}
		return nil
	})

	ctx.Given("the database has no search history", func() error {
		state.mockDB.history = []models.SearchHistory{}
		return nil
	})

	ctx.Given("the database has 5 search history records", func() error {
		now := time.Now()
		state.mockDB.history = []models.SearchHistory{}
		for i := 1; i <= 5; i++ {
			state.mockDB.history = append(state.mockDB.history, models.SearchHistory{
				ID:              int64(i),
				SearchTermID:    int64(i),
				SearchTermDesc:  fmt.Sprintf("Item %d", i),
				ResultsFound:    10,
				NewAdsFound:     1,
				MarketplaceName: "Blocket",
				SearchedAt:      now,
			})
		}
		return nil
	})

	ctx.Given("the database returns an error", func(errMsg string) error {
		state.mockDB.err = errors.New(errMsg)
		return nil
	})

	ctx.When("I record a search with term ID 1, description \"iPhone 15 Pro\", URL \"https://blocket.se/...\", results 10, new ads 3", func() error {
		ctx := context.Background()
		history, err := state.service.RecordSearch(ctx, 1, "iPhone 15 Pro", "https://blocket.se/...", 10, 3)
		state.resultErr = err
		if err == nil && history != nil {
			state.resultHistory = []models.SearchHistory{*history}
		}
		return nil
	})

	ctx.When("I get history for page 1 with page size 20", func() error {
		ctx := context.Background()
		state.resultHistory, state.resultCount, state.resultErr = state.service.GetHistory(ctx, 1, 20)
		return nil
	})

	ctx.When("I get page 1 with page size 2", func() error {
		ctx := context.Background()
		state.resultHistory, state.resultCount, state.resultErr = state.service.GetHistory(ctx, 1, 2)
		return nil
	})

	ctx.When("I get page 2 with page size 2", func() error {
		ctx := context.Background()
		state.resultHistory, _, state.resultErr = state.service.GetHistory(ctx, 2, 2)
		return nil
	})

	ctx.When("I get page 3 with page size 2", func() error {
		ctx := context.Background()
		state.resultHistory, _, state.resultErr = state.service.GetHistory(ctx, 3, 2)
		return nil
	})

	ctx.When("I get history for page 0 with page size 20", func() error {
		ctx := context.Background()
		state.resultHistory, state.resultCount, state.resultErr = state.service.GetHistory(ctx, 0, 20)
		return nil
	})

	ctx.When("I get history for page -1 with page size 20", func() error {
		ctx := context.Background()
		state.resultHistory, state.resultCount, state.resultErr = state.service.GetHistory(ctx, -1, 20)
		return nil
	})

	ctx.When("I get history for page 1 with page size 200", func() error {
		ctx := context.Background()
		_, _, state.resultErr = state.service.GetHistory(ctx, 1, 200)
		return nil
	})

	ctx.Then("the search should be saved successfully", func() error {
		if state.resultErr != nil {
			return fmt.Errorf("expected no error, got %v", state.resultErr)
		}
		return nil
	})

	ctx.Then("the search should have ID set", func() error {
		if len(state.resultHistory) == 0 || state.resultHistory[0].ID == 0 {
			return fmt.Errorf("expected ID to be set")
		}
		return nil
	})

	ctx.Then("the search term description should be \"iPhone 15 Pro\"", func() error {
		if len(state.resultHistory) == 0 || state.resultHistory[0].SearchTermDesc != "iPhone 15 Pro" {
			return fmt.Errorf("expected description 'iPhone 15 Pro', got '%s'", state.resultHistory[0].SearchTermDesc)
		}
		return nil
	})

	ctx.Then("results found should be 10", func() error {
		if len(state.resultHistory) == 0 || state.resultHistory[0].ResultsFound != 10 {
			return fmt.Errorf("expected results 10, got %d", state.resultHistory[0].ResultsFound)
		}
		return nil
	})

	ctx.Then("new ads found should be 3", func() error {
		if len(state.resultHistory) == 0 || state.resultHistory[0].NewAdsFound != 3 {
			return fmt.Errorf("expected new ads 3, got %d", state.resultHistory[0].NewAdsFound)
		}
		return nil
	})

	ctx.Then("I should receive 2 history records", func() error {
		if len(state.resultHistory) != 2 {
			return fmt.Errorf("expected 2 records, got %d", len(state.resultHistory))
		}
		return nil
	})

	ctx.Then("total count should be 2", func() error {
		if state.resultCount != 2 {
			return fmt.Errorf("expected count 2, got %d", state.resultCount)
		}
		return nil
	})

	ctx.Then("first record should have description \"iPhone 15\"", func() error {
		if len(state.resultHistory) == 0 || state.resultHistory[0].SearchTermDesc != "iPhone 15" {
			return fmt.Errorf("expected first record 'iPhone 15', got '%s'", state.resultHistory[0].SearchTermDesc)
		}
		return nil
	})

	ctx.Then("I should receive 0 history records", func() error {
		if len(state.resultHistory) != 0 {
			return fmt.Errorf("expected 0 records, got %d", len(state.resultHistory))
		}
		return nil
	})

	ctx.Then("total count should be 0", func() error {
		if state.resultCount != 0 {
			return fmt.Errorf("expected count 0, got %d", state.resultCount)
		}
		return nil
	})

	ctx.Then("I should receive 2 records", func() error {
		if len(state.resultHistory) != 2 {
			return fmt.Errorf("expected 2 records, got %d", len(state.resultHistory))
		}
		return nil
	})

	ctx.Then("total count should be 5", func() error {
		if state.resultCount != 5 {
			return fmt.Errorf("expected count 5, got %d", state.resultCount)
		}
		return nil
	})

	ctx.Then("first record on page 2 should have ID 3", func() error {
		if len(state.resultHistory) == 0 || state.resultHistory[0].ID != 3 {
			return fmt.Errorf("expected first record ID 3, got %d", state.resultHistory[0].ID)
		}
		return nil
	})

	ctx.Then("I should receive 1 record", func() error {
		if len(state.resultHistory) != 1 {
			return fmt.Errorf("expected 1 record, got %d", len(state.resultHistory))
		}
		return nil
	})

	ctx.Then("no error should occur", func() error {
		if state.resultErr != nil {
			return fmt.Errorf("expected no error, got %v", state.resultErr)
		}
		return nil
	})

	ctx.Then("an error should be returned", func() error {
		if state.resultErr == nil {
			return fmt.Errorf("expected error, got nil")
		}
		return nil
	})

	ctx.Then("it should indicate empty state", func() error {
		state.resultEmpty = len(state.resultHistory) == 0 && state.resultCount == 0
		if !state.resultEmpty {
			return fmt.Errorf("expected empty state")
		}
		return nil
	})
}

func getFeaturesPath(filename string) string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "..", "..", "..", "internal", "test", "gherkin", "features", filename)
}

func TestSearchHistoryFeatures(t *testing.T) {
	featurePath := getFeaturesPath("search_history.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeSearchHistoryScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run search history gherkin tests")
	}
}
