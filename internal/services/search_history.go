package services

import (
	"context"
	"time"

	"begbot/internal/models"
)

type SearchHistoryDB interface {
	SaveSearchHistory(ctx context.Context, h *models.SearchHistory) error
	GetSearchHistory(ctx context.Context, limit, offset int) ([]models.SearchHistory, error)
	GetSearchHistoryCount(ctx context.Context) (int, error)
}

type SearchHistoryService struct {
	db SearchHistoryDB
}

func NewSearchHistoryService(db SearchHistoryDB) *SearchHistoryService {
	return &SearchHistoryService{db: db}
}

func (s *SearchHistoryService) RecordSearch(ctx context.Context, termID int64, termDescription, url string, resultsFound, newAds int) (*models.SearchHistory, error) {
	history := &models.SearchHistory{
		SearchTermID:    termID,
		SearchTermDesc:  termDescription,
		URL:             url,
		ResultsFound:    resultsFound,
		NewAdsFound:     newAds,
		MarketplaceID:   nil,
		MarketplaceName: "",
		SearchedAt:      time.Now(),
	}

	if err := s.db.SaveSearchHistory(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *SearchHistoryService) GetHistory(ctx context.Context, page, pageSize int) ([]models.SearchHistory, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	history, err := s.db.GetSearchHistory(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.db.GetSearchHistoryCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	return history, count, nil
}
