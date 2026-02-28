package services

import (
	"context"
	"fmt"
	"strings"

	"begbot/internal/db"
	"begbot/internal/models"
)

type SearchTermService struct {
	db *db.Postgres
}

func NewSearchTermService(db *db.Postgres) *SearchTermService {
	return &SearchTermService{db: db}
}

// CreateSearchTerm creates a new search term. `marketplaceID` may be nil when
// the caller doesn't specify a marketplace. When present and pointing to the
// Tradera marketplace (id == 2) plain queries (not URL-like) are normalized
// into a Tradera search URL.
func (s *SearchTermService) CreateSearchTerm(ctx context.Context, description, url string, marketplaceID *int64) (*models.SearchTerm, error) {
	// Normalize input before creating model
	normalized := NormalizeSearchURL(url, marketplaceID)

	term := &models.SearchTerm{
		Description: description,
		URL:         normalized,
		IsActive:    true,
	}
	// Preserve marketplace_id nilability
	if marketplaceID != nil {
		term.MarketplaceID = marketplaceID
	}

	if err := s.db.SaveSearchTerm(ctx, term); err != nil {
		return nil, err
	}

	return term, nil
}

// NormalizeSearchURL normalizes a user-provided search input into a full
// search URL when possible. Currently supports Tradera (marketplace id == 2).
func NormalizeSearchURL(input string, marketplaceID *int64) string {
	if isLikelyURL(input) {
		return input
	}
	if marketplaceID != nil && *marketplaceID == 2 {
		return fmt.Sprintf("https://www.tradera.com/search?q=%s", strings.ReplaceAll(input, " ", "+"))
	}
	return input
}

// isLikelyURL performs a simple check to see if the input looks like a URL.
func isLikelyURL(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	// Very simple heuristic: if it contains spaces it's probably a query
	if strings.Contains(s, " ") {
		return false
	}
	// If contains a dot and no spaces, assume URL-like
	if strings.Contains(s, ".") {
		return true
	}
	return false
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
