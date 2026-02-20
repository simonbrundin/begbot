package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"begbot/internal/config"
)

func TestTraderaValuationMethod_ParsesAPIResponse(t *testing.T) {
	// Mock: first request = page load (returns cookies), second = API response
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			// Page load – set a session cookie
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123"})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html></html>"))
			return
		}
		// API response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"averagePrice":2045,"lowestPrice":160,"highestPrice":2900,"count":901}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = true
	cfg.Scraping.Tradera.Timeout = 5 * time.Second
	cfg.Scraping.Tradera.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &TraderaValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Apple", Model: "iPhone 13"}
	v, err := method.Valuate(context.Background(), pi)
	if err != nil {
		t.Fatalf("Valuate returned error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil valuation")
	}
	if v.Value != 2045 {
		t.Fatalf("expected value 2045, got %d", v.Value)
	}
	if v.Confidence < 0.84 {
		t.Fatalf("expected high confidence for 901 items, got %f", v.Confidence)
	}
}

func TestTraderaValuationMethod_ReturnsErrorWhenNoPrices(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html></html>"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"averagePrice":0,"lowestPrice":0,"highestPrice":0,"count":0}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = true
	cfg.Scraping.Tradera.Timeout = 5 * time.Second
	cfg.Scraping.Tradera.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &TraderaValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Okänt", Model: "Produkt"}
	v, err := method.Valuate(context.Background(), pi)
	if err == nil {
		t.Fatal("expected error when no prices found")
	}
	if v != nil {
		t.Fatal("expected nil valuation")
	}
}

func TestTraderaValuationMethod_DisabledConfig(t *testing.T) {
	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = false

	svc := &ValuationService{cfg: cfg}
	method := &TraderaValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Apple", Model: "iPhone 13"}
	v, err := method.Valuate(context.Background(), pi)
	if err != nil {
		t.Fatalf("expected no error for disabled config, got: %v", err)
	}
	if v != nil {
		t.Fatal("expected nil valuation for disabled config")
	}
}

func TestTraderaValuationMethod_CachesResult(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		// First request: page load (HTML). All subsequent requests: API JSON.
		if requestCount == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html></html>"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"averagePrice":%d,"lowestPrice":100,"highestPrice":300,"count":10}`, 1000+requestCount)))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = true
	cfg.Scraping.Tradera.Timeout = 5 * time.Second
	cfg.Scraping.Tradera.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &TraderaValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Test", Model: "Cache"}

	v1, _ := method.Valuate(context.Background(), pi)
	v2, _ := method.Valuate(context.Background(), pi)

	if v1.Value != v2.Value {
		t.Fatalf("expected cached result, got different values: %d vs %d", v1.Value, v2.Value)
	}
	// Should have made only 3 requests (page + 2 api queries), not 6
	if requestCount != 3 {
		t.Fatalf("expected 3 requests (cached second call), got %d", requestCount)
	}
}
