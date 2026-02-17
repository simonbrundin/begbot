package gherkin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"begbot/internal/models"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

// SearchHistoryTestContext holds the test state for search history scenarios
type SearchHistoryTestContext struct {
	service         *services.SearchHistoryService
	mockDB          *mockSearchHistoryDB
	lastHistory     *models.SearchHistory
	histories       []models.SearchHistory
	lastError       error
	count           int
	lastDescription string
	lastID          int64
}

// mockSearchHistoryDB implements the SearchHistoryDB interface for testing
type mockSearchHistoryDB struct {
	history []models.SearchHistory
	err     error
}

func newMockSearchHistoryDB() *mockSearchHistoryDB {
	return &mockSearchHistoryDB{
		history: []models.SearchHistory{},
		err:     nil,
	}
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

func InitializeScenarioSearchHistory(ctx *godog.ScenarioContext) {
	tc := &SearchHistoryTestContext{}

	ctx.BeforeScenario(func(*godog.Scenario) {
		tc.mockDB = newMockSearchHistoryDB()
		tc.service = services.NewSearchHistoryService(tc.mockDB)
		tc.lastHistory = nil
		tc.histories = nil
		tc.lastError = nil
		tc.count = 0
		tc.lastDescription = ""
		tc.lastID = 0
	})

	// Given steps
	ctx.Given(`^a search history service with mock database$`, func() error {
		tc.mockDB = newMockSearchHistoryDB()
		tc.service = services.NewSearchHistoryService(tc.mockDB)
		return nil
	})

	ctx.Given(`^the database has the following search history:$`, func(data *godog.Table) error {
		tc.mockDB.history = []models.SearchHistory{}
		now := time.Now()
		for i, row := range data.Rows[1:] {
			id := i + 1
			searchTermID, _ := strToInt64(row.Cells[1].Value)
			tc.mockDB.history = append(tc.mockDB.history, models.SearchHistory{
				ID:              int64(id),
				SearchTermID:    searchTermID,
				SearchTermDesc:  row.Cells[2].Value,
				URL:             row.Cells[3].Value,
				ResultsFound:    strToInt(row.Cells[4].Value),
				NewAdsFound:     strToInt(row.Cells[5].Value),
				MarketplaceID:   nil,
				MarketplaceName: "Blocket",
				SearchedAt:      now.Add(-time.Duration(i) * time.Hour),
			})
		}
		return nil
	})

	ctx.Given(`^the database has no search history$`, func() error {
		tc.mockDB.history = []models.SearchHistory{}
		return nil
	})

	ctx.Given(`^the database has "(\d+)" search history records$`, func(countStr string) error {
		tc.mockDB.history = []models.SearchHistory{}
		count := strToInt(countStr)
		for i := 0; i < count; i++ {
			tc.mockDB.history = append(tc.mockDB.history, models.SearchHistory{
				ID:              int64(i + 1),
				SearchTermID:    int64(i + 1),
				SearchTermDesc:  fmt.Sprintf("Item %d", i+1),
				URL:             fmt.Sprintf("https://example.com/%d", i+1),
				ResultsFound:    10,
				NewAdsFound:     1,
				MarketplaceName: "Blocket",
				SearchedAt:      time.Now(),
			})
		}
		return nil
	})

	ctx.Given(`^the database returns error "([^"]+)"$`, func(errMsg string) error {
		tc.mockDB.err = errors.New(errMsg)
		return nil
	})

	// When steps
	ctx.When(`^I record a search with term ID "(\d+)", description "([^"]+)", URL "([^"]+)", results "(\d+)", new ads "(\d+)"$`, 
		func(termIDStr, desc, url, resultsStr, newAdsStr string) error {
		termID, _ := strToInt64(termIDStr)
		results, _ := strToInt(resultsStr)
		newAds, _ := strToInt(newAdsStr)
		
		tc.lastHistory, tc.lastError = tc.service.RecordSearch(context.Background(), termID, desc, url, results, newAds)
		return nil
	})

	ctx.When(`^I get search history for page "(\d+)" with page size "(\d+)"$`, func(pageStr, pageSizeStr string) error {
		page, _ := strToInt(pageStr)
		pageSize, _ := strToInt(pageSizeStr)
		tc.histories, tc.count, tc.lastError = tc.service.GetHistory(context.Background(), page, pageSize)
		return nil
	})

	ctx.When(`^I try to record a search with term ID "(\d+)"$`, func(termIDStr string) error {
		termID, _ := strToInt64(termIDStr)
		_, tc.lastError = tc.service.RecordSearch(context.Background(), termID, "Test", "https://...", 10, 2)
		return nil
	})

	ctx.When(`^I try to get search history$`, func() error {
		_, _, tc.lastError = tc.service.GetHistory(context.Background(), 1, 20)
		return nil
	})

	// Then steps
	ctx.Then(`^the search should be recorded successfully$`, func() error {
		if tc.lastError != nil {
			return fmt.Errorf("expected no error, got %v", tc.lastError)
		}
		if tc.lastHistory == nil {
			return errors.New("expected history to be returned")
		}
		return nil
	})

	ctx.Then(`^the search term ID should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt64(expectedStr)
		if tc.lastHistory.SearchTermID != expected {
			return fmt.Errorf("expected SearchTermID %d, got %d", expected, tc.lastHistory.SearchTermID)
		}
		return nil
	})

	ctx.Then(`^the search term description should be "([^"]+)"$`, func(expected string) error {
		if tc.lastHistory.SearchTermDesc != expected {
			return fmt.Errorf("expected SearchTermDesc '%s', got '%s'", expected, tc.lastHistory.SearchTermDesc)
		}
		return nil
	})

	ctx.Then(`^the results found should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		if tc.lastHistory.ResultsFound != expected {
			return fmt.Errorf("expected ResultsFound %d, got %d", expected, tc.lastHistory.ResultsFound)
		}
		return nil
	})

	ctx.Then(`^the new ads found should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		if tc.lastHistory.NewAdsFound != expected {
			return fmt.Errorf("expected NewAdsFound %d, got %d", expected, tc.lastHistory.NewAdsFound)
		}
		return nil
	})

	ctx.Then(`^the URL should be "([^"]+)"$`, func(expected string) error {
		if tc.lastHistory.URL != expected {
			return fmt.Errorf("expected URL '%s', got '%s'", expected, tc.lastHistory.URL)
		}
		return nil
	})

	ctx.Then(`^the history ID should be set$`, func() error {
		if tc.lastHistory.ID == 0 {
			return errors.New("expected history ID to be set")
		}
		return nil
	})

	ctx.Then(`^I should receive "(\d+)" history records$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		if len(tc.histories) != expected {
			return fmt.Errorf("expected %d history records, got %d", expected, len(tc.histories))
		}
		return nil
	})

	ctx.Then(`^the total count should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		if tc.count != expected {
			return fmt.Errorf("expected count %d, got %d", expected, tc.count)
		}
		return nil
	})

	ctx.Then(`^the first record should have description "([^"]+)"$`, func(expected string) error {
		if len(tc.histories) == 0 {
			return errors.New("no histories to check")
		}
		if tc.histories[0].SearchTermDesc != expected {
			return fmt.Errorf("expected first record to be '%s', got '%s'", expected, tc.histories[0].SearchTermDesc)
		}
		return nil
	})

	ctx.Then(`^the count should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		if tc.count != expected {
			return fmt.Errorf("expected count %d, got %d", expected, tc.count)
		}
		return nil
	})

	ctx.Then(`^I should receive "(\d+)" history records on page (\d+)$`, func(expectedStr, pageStr string) error {
		expected, _ := strToInt(expectedStr)
		if len(tc.histories) != expected {
			return fmt.Errorf("expected %d history records on page, got %d", expected, len(tc.histories))
		}
		return nil
	})

	ctx.Then(`^the record should start at ID "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt64(expectedStr)
		if len(tc.histories) == 0 {
			return errors.New("no histories to check")
		}
		if tc.histories[0].ID != expected {
			return fmt.Errorf("expected record to start at ID %d, got %d", expected, tc.histories[0].ID)
		}
		return nil
	})

	ctx.Then(`^the request should succeed$`, func() error {
		if tc.lastError != nil {
			return fmt.Errorf("expected no error, got %v", tc.lastError)
		}
		return nil
	})

	ctx.Then(`^I should receive an error$`, func() error {
		if tc.lastError == nil {
			return errors.New("expected error but got none")
		}
		return nil
	})

	ctx.Then(`^the system should detect empty state$`, func() error {
		isEmpty := len(tc.histories) == 0 && tc.count == 0
		if !isEmpty {
			return errors.New("expected empty state")
		}
		return nil
	})
}

func TestSearchHistoryFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenarioSearchHistory,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"internal/test/gherkin/features/search_history.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, there are failed test scenarios")
	}
}

// Helper functions
func strToInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func strToInt64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

func init() {
	// Register godog as test
	os.Chdir("/home/simon/repos/begbot")
}
