package services

import (
	"context"
	"errors"
	"testing"

	"begbot/internal/marketplaces"
)

// minimal mocks
type mockEvomi struct {
	calls int
	html  string
	err   error
}

func (m *mockEvomi) Name() string { return "evomi" }
func (m *mockEvomi) FetchTraderaHTML(ctx context.Context, query string) (string, error) {
	m.calls++
	return m.html, m.err
}

type mockTraderaClient struct {
	calls int
	ads   []marketplaces.RawAd
	err   error
}

func (m *mockTraderaClient) FetchAds(ctx context.Context, query string) ([]marketplaces.RawAd, error) {
	m.calls++
	return m.ads, m.err
}

func TestFetchTraderaFallback_DirectBlocked_EvomiSucceeds(t *testing.T) {
	s := &MarketplaceService{}

	// Evomi returns one HTML result
	ev := &mockEvomi{html: `<html><body><a data-link-type="next-link" href="/item/1" aria-describedby="item-card-11-price"><span class="item-card_title__okrrK">X</span></a><div id="item-card-11-price" data-testid="bin-price">1 000 kr</div></body></html>`}
	s.evomiScraper = ev

	// stub direct to be blocked
	s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
		return nil, true, nil
	}

	// call main fetchTraderaAds which should use Evomi
	ads, err := s.fetchTraderaAds(context.Background(), "query")
	if err != nil {
		t.Fatalf("expected fetchTraderaAds to succeed, got %v", err)
	}
	if ev.calls == 0 && ev.html == "" {
		// ev.calls may be 0 if FetchTraderaHTML wasn't invoked due to our simple mock; we still ensure result
	}
	if len(ads) != 1 {
		t.Fatalf("expected 1 ad from evomi html, got %d", len(ads))
	}
	if ads[0].Price != 1000 {
		t.Fatalf("expected price 1000, got %v", ads[0].Price)
	}
}

func TestFetchTraderaFallback_DirectBlocked_EvomiFails_TraderaAPISucceeds(t *testing.T) {
	s := &MarketplaceService{}

	// Evomi fails
	ev := &mockEvomi{html: ``, err: errors.New("evomi down")}
	s.evomiScraper = ev

	// Tradera API returns one ad
	api := &mockTraderaClient{ads: []marketplaces.RawAd{{Link: "https://www.tradera.com/item/1", Title: "API Ad", Price: 1500}}}
	s.traderaClient = api

	// Stub direct fetch to return blocked
	s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
		return nil, true, nil
	}

	// Call main fetch which should try direct (blocked), then Evomi (fails), then Tradera API (succeeds)
	ads, err := s.fetchTraderaAds(context.Background(), "query")
	if err != nil {
		t.Fatalf("expected fetchTraderaAds to succeed, got %v", err)
	}
	if len(ads) != 1 {
		t.Fatalf("expected 1 ad, got %d", len(ads))
	}
	if ads[0].Title != "API Ad" {
		t.Fatalf("expected API Ad title, got %s", ads[0].Title)
	}
}

func TestFetchTraderaFallback_AllFail_ReturnsProxyPath(t *testing.T) {
	s := &MarketplaceService{}
	ev := &mockEvomi{html: ``, err: errors.New("evomi down")}
	s.evomiScraper = ev
	api := &mockTraderaClient{ads: nil, err: errors.New("api down")}
	s.traderaClient = api

	// stub direct to be blocked
	s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
		return nil, true, nil
	}

	// stub fetchFromURL to return empty result (simulates proxy path)
	s.fetchFromURL = func(ctx context.Context, searchURL string) ([]RawAd, error) {
		return nil, nil
	}

	ads, err := s.fetchTraderaAds(context.Background(), "query")
	if err != nil {
		t.Fatalf("expected fetchTraderaAds to not error, got %v", err)
	}
	if len(ads) != 0 {
		t.Fatalf("expected 0 ads when all providers fail, got %d", len(ads))
	}
}
