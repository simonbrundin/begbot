package services

import (
	"context"

	"begbot/internal/db"
	"begbot/internal/models"
)

type SearchTermService struct {
	db *db.Postgres
}

func NewSearchTermService(db *db.Postgres) *SearchTermService {
	return &SearchTermService{db: db}
}

func (s *SearchTermService) CreateSearchTerm(ctx context.Context, description, url string, marketplaceID int64) (*models.SearchTerm, error) {
	term := &models.SearchTerm{
		Description:   description,
		URL:           url,
		MarketplaceID: &marketplaceID,
		IsActive:      true,
	}

	if err := s.db.SaveSearchTerm(ctx, term); err != nil {
		return nil, err
	}

	return term, nil
}

func (s *SearchTermService) GetActiveSearchTerms(ctx context.Context) ([]models.SearchTerm, error) {
	return s.db.GetActiveSearchTerms(ctx)
}

func (s *SearchTermService) DeactivateSearchTerm(ctx context.Context, id int64) error {
	return s.db.UpdateSearchTermStatus(ctx, id, false)
}

type SearchJob struct {
	SearchTerm  models.SearchTerm
	Marketplace *models.Marketplace
}

func (s *SearchTermService) GetSearchJobs(ctx context.Context) ([]SearchJob, error) {
	terms, err := s.db.GetActiveSearchTerms(ctx)
	if err != nil {
		return nil, err
	}

	var jobs []SearchJob
	for _, term := range terms {
		job := SearchJob{
			SearchTerm: term,
		}

		if term.MarketplaceID != nil {
			marketplace, err := s.db.GetMarketplaceByID(ctx, *term.MarketplaceID)
			if err != nil {
				return nil, err
			}
			job.Marketplace = marketplace
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
