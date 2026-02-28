package services

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"begbot/internal/marketplaces"
)

// End-to-end style tests that assert which provider path is used via call flags.
func TestFetchTraderaFallback_EndToEndPaths(t *testing.T) {
	// Scenario 1: direct succeeds -> only direct should be used
	{
		var buf bytes.Buffer
		s := &MarketplaceService{logger: log.New(&buf, "", 0)}
		ev := &mockEvomi{html: ``, err: errors.New("shouldn't be called")}
		// For the direct-success scenario we don't register a traderaClient so
		// the service will attempt a direct scrape first (matching the test
		// intent after normalization to API-first behavior in other tests).
		s.evomiScraper = ev
		s.traderaClient = nil

		directCalled := false
		s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
			directCalled = true
			return []RawAd{{Link: "d", Title: "Direct", Price: 111}}, false, nil
		}
		proxyCalled := false
		s.fetchFromURL = func(ctx context.Context, searchURL string) ([]RawAd, error) {
			proxyCalled = true
			return nil, nil
		}

		ads, err := s.fetchTraderaAds(context.Background(), "q")
		if err != nil {
			t.Fatalf("direct-success: unexpected error: %v", err)
		}
		if !directCalled {
			t.Fatalf("direct-success: expected direct to be called")
		}
		if ev.calls != 0 || proxyCalled {
			t.Fatalf("direct-success: expected no fallback providers called; evomi=%v api=nil proxy=%v", ev.calls, proxyCalled)
		}
		if len(ads) != 1 || ads[0].Title != "Direct" {
			t.Fatalf("direct-success: unexpected ads: %+v", ads)
		}
	}

	// Scenario 2: direct blocked -> evomi succeeds
	{
		var buf bytes.Buffer
		s := &MarketplaceService{logger: log.New(&buf, "", 0)}
		ev := &mockEvomi{html: `<html><body><a data-link-type="next-link" href="/item/1" aria-describedby="item-card-11-price"><span class="item-card_title__okrrK">E</span></a><div id="item-card-11-price" data-testid="bin-price">500 kr</div></body></html>`}
		// No Tradera API client for this scenario; we want direct->evomi path.
		s.evomiScraper = ev
		s.traderaClient = nil

		s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
			return nil, true, nil
		}
		proxyCalled := false
		s.fetchFromURL = func(ctx context.Context, searchURL string) ([]RawAd, error) {
			proxyCalled = true
			return nil, nil
		}

		ads, err := s.fetchTraderaAds(context.Background(), "q")
		if err != nil {
			t.Fatalf("evomi-path: unexpected error: %v", err)
		}
		if ev.calls == 0 {
			t.Fatalf("evomi-path: expected evomi to be called")
		}
		// Assert logs contain blocked reason and evomi success
		logs := buf.String()
		if !strings.Contains(logs, "Direct scrape appears blocked") {
			t.Fatalf("expected log to state direct scrape blocked; got logs=%q", logs)
		}
		if !strings.Contains(logs, "Fetched 1 ads from Evomi Scraper API") {
			t.Fatalf("expected evomi success log; got logs=%q", logs)
		}
		if proxyCalled {
			t.Fatalf("evomi-path: expected no proxy calls; proxy=%v", proxyCalled)
		}
		if len(ads) != 1 || ads[0].Price != 500 {
			t.Fatalf("evomi-path: unexpected ads: %+v", ads)
		}
	}

	// Scenario 3: direct blocked -> evomi fails -> tradera API succeeds
	{
		var buf bytes.Buffer
		s := &MarketplaceService{logger: log.New(&buf, "", 0)}
		ev := &mockEvomi{html: ``, err: errors.New("evomi down")}
		api := &mockTraderaClient{ads: []marketplaces.RawAd{{Link: "https://www.tradera.com/item/1", Title: "API", Price: 950}}, err: nil}
		s.evomiScraper = ev
		s.traderaClient = api

		s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
			return nil, true, nil
		}
		proxyCalled := false
		s.fetchFromURL = func(ctx context.Context, searchURL string) ([]RawAd, error) {
			proxyCalled = true
			return nil, nil
		}

		ads, err := s.fetchTraderaAds(context.Background(), "q")
		if err != nil {
			t.Fatalf("api-path: unexpected error: %v", err)
		}
		// With API-first ordering we expect the Tradera API to be invoked and
		// return results. Evomi should not be called in this scenario because
		// the API succeeded early.
		if api.calls == 0 {
			t.Fatalf("api-path: expected tradera API to be called")
		}
		if ev.calls != 0 {
			t.Fatalf("api-path: expected evomi NOT to be called when API succeeds; evomi=%v", ev.calls)
		}
		// logs should show evomi failure and tradera success
		logs := buf.String()
		if !strings.Contains(logs, "Trying Tradera API") {
			t.Fatalf("expected log to contain Tradera API attempt; got logs=%q", logs)
		}
		if !strings.Contains(logs, "Fetched 1 ads from Tradera API") {
			t.Fatalf("expected tradera api success log; got logs=%q", logs)
		}
		if proxyCalled {
			t.Fatalf("api-path: proxy should not be called when api succeeds")
		}
		if len(ads) != 1 || ads[0].Title != "API" {
			t.Fatalf("api-path: unexpected ads: %+v", ads)
		}
	}

	// Scenario 4: all fail -> proxy path used
	{
		var buf bytes.Buffer
		s := &MarketplaceService{logger: log.New(&buf, "", 0)}
		ev := &mockEvomi{html: ``, err: errors.New("evomi down")}
		api := &mockTraderaClient{ads: nil, err: errors.New("api down")}
		s.evomiScraper = ev
		s.traderaClient = api

		s.fetchDirect = func(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
			return nil, true, nil
		}
		proxyCalled := false
		s.fetchFromURL = func(ctx context.Context, searchURL string) ([]RawAd, error) {
			proxyCalled = true
			return []RawAd{{Link: "p", Title: "Proxy", Price: 700}}, nil
		}

		ads, err := s.fetchTraderaAds(context.Background(), "q")
		if err != nil {
			t.Fatalf("proxy-path: unexpected error: %v", err)
		}
		if ev.calls == 0 {
			t.Fatalf("proxy-path: expected evomi to be called")
		}
		if api.calls == 0 {
			t.Fatalf("proxy-path: expected tradera api to be called")
		}
		if !proxyCalled {
			t.Fatalf("proxy-path: expected proxy fetch to be called")
		}
		logs := buf.String()
		if !strings.Contains(logs, "Tradera API failed") && !strings.Contains(logs, "falling back to proxy") {
			t.Fatalf("expected tradera api failure / proxy fallback logs; got logs=%q", logs)
		}
		if len(ads) != 1 || ads[0].Title != "Proxy" {
			t.Fatalf("proxy-path: unexpected ads: %+v", ads)
		}
	}
}
