package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"begbot/internal/models"
	"begbot/internal/services"
)

// fakeSearchTermSvc records calls and returns a predictable SearchTerm.
type fakeSearchTermSvc struct {
	called   bool
	lastDesc string
	lastURL  string
	lastMP   *int64
}

func (f *fakeSearchTermSvc) CreateSearchTerm(ctx context.Context, description, url string, marketplaceID *int64) (*models.SearchTerm, error) {
	f.called = true
	f.lastDesc = description
	f.lastURL = url
	f.lastMP = marketplaceID
	// Return a created term (CreatedAt/UpdatedAt not set - handler doesn't assert them)
	return &models.SearchTerm{ID: 123, Description: description, URL: services.NormalizeSearchURL(url, marketplaceID), MarketplaceID: marketplaceID, IsActive: true}, nil
}

func TestSearchTermsHandler_NormalizesTradera(t *testing.T) {
	// Arrange
	srv := &Server{}
	fake := &fakeSearchTermSvc{}
	srv.searchTermSvc = fake

	payload := map[string]interface{}{
		"description":    "lego",
		"url":            "lego star wars",
		"marketplace_id": 2,
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/search-terms", bytes.NewReader(b))
	rr := httptest.NewRecorder()

	// Act
	srv.searchTermsHandler(rr, req)

	// Assert
	if rr.Code != 201 {
		t.Fatalf("expected 201 got %d, body=%s", rr.Code, rr.Body.String())
	}
	if !fake.called {
		t.Fatalf("expected CreateSearchTerm called")
	}
	if fake.lastURL != "lego star wars" {
		t.Fatalf("handler should forward raw url to service; got %q", fake.lastURL)
	}
	// Verify the returned JSON contains the normalized URL
	var created models.SearchTerm
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	want := "https://www.tradera.com/search?q=lego+star+wars"
	if created.URL != want {
		t.Fatalf("expected normalized URL %q got %q", want, created.URL)
	}
}
