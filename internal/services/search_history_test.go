package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"begbot/internal/models"
)

type mockSearchHistoryDBForTest struct {
	history []models.SearchHistory
	err     error
}

func (m *mockSearchHistoryDBForTest) SaveSearchHistory(ctx context.Context, h *models.SearchHistory) error {
	if m.err != nil {
		return m.err
	}
	h.ID = int64(len(m.history) + 1)
	h.CreatedAt = time.Now()
	m.history = append(m.history, *h)
	return nil
}

func (m *mockSearchHistoryDBForTest) GetSearchHistory(ctx context.Context, limit, offset int) ([]models.SearchHistory, error) {
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

func (m *mockSearchHistoryDBForTest) GetSearchHistoryCount(ctx context.Context) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return len(m.history), nil
}

func TestSearchHistoryServiceReal_RecordSearch_Success(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{},
		err:     nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()
	history, err := service.RecordSearch(ctx, 1, "iPhone 15 Pro", "https://blocket.se/...", 10, 3)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if history == nil {
		t.Fatal("Expected history to be returned")
	}

	if history.SearchTermID != 1 {
		t.Errorf("Expected SearchTermID 1, got %d", history.SearchTermID)
	}

	if history.SearchTermDesc != "iPhone 15 Pro" {
		t.Errorf("Expected SearchTermDesc 'iPhone 15 Pro', got '%s'", history.SearchTermDesc)
	}

	if history.ResultsFound != 10 {
		t.Errorf("Expected ResultsFound 10, got %d", history.ResultsFound)
	}

	if history.NewAdsFound != 3 {
		t.Errorf("Expected NewAdsFound 3, got %d", history.NewAdsFound)
	}

	if history.URL != "https://blocket.se/..." {
		t.Errorf("Expected URL 'https://blocket.se/...', got '%s'", history.URL)
	}

	if history.ID == 0 {
		t.Error("Expected history ID to be set")
	}
}

func TestSearchHistoryServiceReal_GetHistory_WithData(t *testing.T) {
	now := time.Now()
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{
			{
				ID:              1,
				SearchTermID:    1,
				SearchTermDesc:  "iPhone 15",
				URL:             "https://blocket.se/iphone",
				ResultsFound:    15,
				NewAdsFound:     5,
				MarketplaceID:   nil,
				MarketplaceName: "Blocket",
				SearchedAt:      now.Add(-time.Hour),
			},
			{
				ID:              2,
				SearchTermID:    2,
				SearchTermDesc:  "MacBook Pro",
				URL:             "https://blocket.se/macbook",
				ResultsFound:    8,
				NewAdsFound:     2,
				MarketplaceID:   nil,
				MarketplaceName: "Blocket",
				SearchedAt:      now,
			},
		},
		err: nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()
	history, count, err := service.GetHistory(ctx, 1, 20)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 history records, got %d", len(history))
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	if history[0].SearchTermDesc != "iPhone 15" {
		t.Errorf("Expected first record to be iPhone 15, got '%s'", history[0].SearchTermDesc)
	}
}

func TestSearchHistoryServiceReal_GetHistory_Empty(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{},
		err:     nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()
	history, count, err := service.GetHistory(ctx, 1, 20)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected 0 history records, got %d", len(history))
	}

	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestSearchHistoryServiceReal_GetHistory_Pagination(t *testing.T) {
	now := time.Now()
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{
			{ID: 1, SearchTermID: 1, SearchTermDesc: "Item 1", ResultsFound: 10, NewAdsFound: 1, SearchedAt: now},
			{ID: 2, SearchTermID: 2, SearchTermDesc: "Item 2", ResultsFound: 10, NewAdsFound: 1, SearchedAt: now},
			{ID: 3, SearchTermID: 3, SearchTermDesc: "Item 3", ResultsFound: 10, NewAdsFound: 1, SearchedAt: now},
			{ID: 4, SearchTermID: 4, SearchTermDesc: "Item 4", ResultsFound: 10, NewAdsFound: 1, SearchedAt: now},
			{ID: 5, SearchTermID: 5, SearchTermDesc: "Item 5", ResultsFound: 10, NewAdsFound: 1, SearchedAt: now},
		},
		err: nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()

	page1, count1, err := service.GetHistory(ctx, 1, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(page1) != 2 {
		t.Errorf("Expected 2 items on page 1, got %d", len(page1))
	}
	if count1 != 5 {
		t.Errorf("Expected total count 5, got %d", count1)
	}

	page2, _, err := service.GetHistory(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("Expected 2 items on page 2, got %d", len(page2))
	}
	if page2[0].ID != 3 {
		t.Errorf("Expected second page to start at ID 3, got %d", page2[0].ID)
	}

	page3, _, err := service.GetHistory(ctx, 3, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(page3) != 1 {
		t.Errorf("Expected 1 item on page 3, got %d", len(page3))
	}
}

func TestSearchHistoryServiceReal_GetHistory_InvalidPage(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{},
		err:     nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()

	_, count, err := service.GetHistory(ctx, 0, 20)
	if err != nil {
		t.Fatalf("Expected no error for page 0, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	_, count, err = service.GetHistory(ctx, -1, 20)
	if err != nil {
		t.Fatalf("Expected no error for negative page, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestSearchHistoryServiceReal_GetHistory_MaxPageSize(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{},
		err:     nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()

	_, _, err := service.GetHistory(ctx, 1, 200)
	if err != nil {
		t.Fatalf("Expected no error for large pageSize, got %v", err)
	}
}

func TestSearchHistoryServiceReal_GetHistory_DBError(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: nil,
		err:     errors.New("database unavailable"),
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()
	_, _, err := service.GetHistory(ctx, 1, 20)

	if err == nil {
		t.Fatal("Expected error when DB is unavailable")
	}
}

func TestSearchHistoryServiceReal_RecordSearch_DBError(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: nil,
		err:     errors.New("database unavailable"),
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()
	_, err := service.RecordSearch(ctx, 1, "Test", "https://...", 10, 2)

	if err == nil {
		t.Fatal("Expected error when DB is unavailable")
	}
}

func TestSearchHistoryServiceReal_EmptyStateMessage(t *testing.T) {
	mockDB := &mockSearchHistoryDBForTest{
		history: []models.SearchHistory{},
		err:     nil,
	}
	service := NewSearchHistoryService(mockDB)

	ctx := context.Background()
	history, count, err := service.GetHistory(ctx, 1, 20)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	emptyState := len(history) == 0 && count == 0
	if !emptyState {
		t.Error("Expected empty state when no history exists")
	}
}
