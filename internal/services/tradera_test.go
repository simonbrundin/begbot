package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"begbot/internal/config"
)

func TestTraderaValuationMethod_ParsesSimpleJSON(t *testing.T) {
	// Mock Tradera server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"price":1234,"confidence":80}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = true
	cfg.Scraping.Tradera.Timeout = 5 * time.Second
	cfg.Scraping.Tradera.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &TraderaValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Nokia", Model: "3310", AdText: ""}
	ctx := context.Background()

	v, err := method.Valuate(ctx, pi)
	if err != nil {
		t.Fatalf("Valuate returned error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil valuation")
	}
	if v.Value != 1234 {
		t.Fatalf("expected value 1234, got %d", v.Value)
	}
	if v.Confidence < 0.79 || v.Confidence > 0.81 {
		t.Fatalf("expected confidence ~0.8, got %f", v.Confidence)
	}
}

func TestTraderaValuationMethod_ParsesNestedJSON(t *testing.T) {
	// Nested response with strings
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"valuation": "1299", "confidence": "90"}}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Tradera.Enabled = true
	cfg.Scraping.Tradera.Timeout = 5 * time.Second
	cfg.Scraping.Tradera.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &TraderaValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Sony", Model: "X", AdText: ""}
	ctx := context.Background()

	v, err := method.Valuate(ctx, pi)
	if err != nil {
		t.Fatalf("Valuate returned error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil valuation")
	}
	if v.Value != 1299 {
		t.Fatalf("expected value 1299, got %d", v.Value)
	}
	if v.Confidence < 0.89 || v.Confidence > 0.91 {
		t.Fatalf("expected confidence ~0.9, got %f", v.Confidence)
	}
}
