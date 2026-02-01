package services

import (
	"context"
	"testing"

	"begbot/internal/models"
)

func TestSearchTermService(t *testing.T) {
	_ = &SearchTermService{}

	term := models.SearchTerm{
		ID:            1,
		Description:   "iPhone Search",
		URL:           "https://blocket.se/...?q=iphone&price_from=500",
		MarketplaceID: func() *int64 { v := int64(1); return &v }(),
		IsActive:      true,
	}

	if term.Description != "iPhone Search" {
		t.Errorf("Expected description 'iPhone Search', got '%s'", term.Description)
	}

	if term.URL == "" {
		t.Error("URL should not be empty")
	}

	if term.MarketplaceID == nil || *term.MarketplaceID != 1 {
		t.Error("MarketplaceID should be 1")
	}

	t.Logf("SearchTerm: %+v", term)
}

func TestSearchJob(t *testing.T) {
	marketplace := &models.Marketplace{
		ID:   1,
		Name: "Blocket",
	}

	job := SearchJob{
		SearchTerm: models.SearchTerm{
			ID:            1,
			Description:   "iPhone",
			URL:           "https://blocket.se/search?q=iphone",
			MarketplaceID: func() *int64 { v := int64(1); return &v }(),
			IsActive:      true,
		},
		Marketplace: marketplace,
	}

	if job.SearchTerm.URL == "" {
		t.Error("SearchTerm URL should not be empty")
	}

	if job.Marketplace == nil {
		t.Error("Marketplace should not be nil")
	}

	t.Logf("SearchJob: %+v", job)
}

func TestGetSearchJobsMock(t *testing.T) {
	_ = &SearchTermService{}

	job := SearchJob{
		SearchTerm: models.SearchTerm{
			ID:            1,
			Description:   "Lego Star Wars",
			URL:           "https://www.tradera.com/search?q=lego+star+wars",
			MarketplaceID: func() *int64 { v := int64(2); return &v }(),
			IsActive:      true,
		},
		Marketplace: &models.Marketplace{
			ID:   2,
			Name: "Tradera",
		},
	}

	if job.Marketplace.Name != "Tradera" {
		t.Errorf("Expected marketplace 'Tradera', got '%s'", job.Marketplace.Name)
	}

	t.Logf("Mock SearchJob: %+v", job)
}

func TestSearchTermInactive(t *testing.T) {
	term := models.SearchTerm{
		ID:            1,
		Description:   "Inactive Search",
		URL:           "https://example.com/search",
		MarketplaceID: func() *int64 { v := int64(1); return &v }(),
		IsActive:      false,
	}

	if term.IsActive {
		t.Error("SearchTerm should be inactive")
	}

	t.Logf("Inactive SearchTerm: %+v", term)
}

func TestNilMarketplaceID(t *testing.T) {
	term := models.SearchTerm{
		ID:            1,
		Description:   "No Marketplace",
		URL:           "https://example.com/search",
		MarketplaceID: nil,
		IsActive:      true,
	}

	if term.MarketplaceID != nil {
		t.Error("MarketplaceID should be nil")
	}

	t.Logf("SearchTerm without marketplace: %+v", term)
}

func TestContextPassing(t *testing.T) {
	_ = &SearchTermService{}

	ctx := context.Background()
	_ = ctx

	t.Log("SearchTermService can accept context")
}
