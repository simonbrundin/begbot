package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"begbot/internal/config"
)

func TestBlocketValuationMethod_ParsesAPIResponse(t *testing.T) {
	valuationTypeEnabled["Blocket"] = true
	defer func() { valuationTypeEnabled["Blocket"] = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"docs": [
				{"id": "1", "price": {"amount": 1000}, "heading": "Test product 1"},
				{"id": "2", "price": {"amount": 1500}, "heading": "Test product 2"},
				{"id": "3", "price": {"amount": 2000}, "heading": "Test product 3"},
				{"id": "4", "price": {"amount": 1200}, "heading": "Test product 4"},
				{"id": "5", "price": {"amount": 1800}, "heading": "Test product 5"},
				{"id": "6", "price": {"amount": 2200}, "heading": "Test product 6"},
				{"id": "7", "price": {"amount": 1300}, "heading": "Test product 7"},
				{"id": "8", "price": {"amount": 1700}, "heading": "Test product 8"},
				{"id": "9", "price": {"amount": 1900}, "heading": "Test product 9"},
				{"id": "10", "price": {"amount": 2100}, "heading": "Test product 10"},
				{"id": "11", "price": {"amount": 1400}, "heading": "Test product 11"},
				{"id": "12", "price": {"amount": 1600}, "heading": "Test product 12"}
			]
		}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Blocket.Timeout = 5 * time.Second
	cfg.Scraping.Blocket.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &BlocketValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Test", Model: "Product"}
	v, err := method.Valuate(context.Background(), pi)
	if err != nil {
		t.Fatalf("Valuate returned error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil valuation")
	}
	if v.Value <= 0 {
		t.Fatalf("expected positive value, got %d", v.Value)
	}
	if v.Confidence < 0.5 {
		t.Fatalf("expected confidence >= 0.5 for 10+ items, got %f", v.Confidence)
	}
}

func TestBlocketValuationMethod_FiltersOutliersWithIQR(t *testing.T) {
	valuationTypeEnabled["Blocket"] = true
	defer func() { valuationTypeEnabled["Blocket"] = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"docs": [
				{"id": 1, "price": {"amount": 1000}, "heading": "Normal price 1"},
				{"id": 2, "price": {"amount": 1100}, "heading": "Normal price 2"},
				{"id": 3, "price": {"amount": 1200}, "heading": "Normal price 3"},
				{"id": 4, "price": {"amount": 1300}, "heading": "Normal price 4"},
				{"id": 5, "price": {"amount": 100}, "heading": "Outlier - too cheap"},
				{"id": 6, "price": {"amount": 10000}, "heading": "Outlier - too expensive"},
				{"id": 7, "price": {"amount": 1250}, "heading": "Normal price 5"},
				{"id": 8, "price": {"amount": 1150}, "heading": "Normal price 6"},
				{"id": 9, "price": {"amount": 1050}, "heading": "Normal price 7"},
				{"id": 10, "price": {"amount": 950}, "heading": "Normal price 8"},
				{"id": 11, "price": {"amount": 1350}, "heading": "Normal price 9"},
				{"id": 12, "price": {"amount": 1200}, "heading": "Normal price 10"}
			]
		}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Blocket.Timeout = 5 * time.Second
	cfg.Scraping.Blocket.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &BlocketValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Test", Model: "Product"}
	v, err := method.Valuate(context.Background(), pi)
	if err != nil {
		t.Fatalf("Valuate returned error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil valuation")
	}

	if v.Value > 2000 {
		t.Fatalf("expected value filtered from outliers, got %d", v.Value)
	}

	if v.Metadata != nil {
		if totalCount, ok := v.Metadata["total_count"].(float64); ok {
			if int(totalCount) != 12 {
				t.Errorf("expected total_count 12, got %d", int(totalCount))
			}
		}
		if filteredCount, ok := v.Metadata["filtered_count"].(float64); ok {
			if int(filteredCount) <= 0 {
				t.Errorf("expected positive filtered_count, got %d", int(filteredCount))
			}
			t.Logf("Filtered from 12 to %d prices using IQR", int(filteredCount))
		}
	}
}

func TestBlocketValuationMethod_DisabledConfig(t *testing.T) {
	valuationTypeEnabled["Blocket"] = false
	defer func() { valuationTypeEnabled["Blocket"] = true }()

	cfg := &config.Config{}

	svc := &ValuationService{cfg: cfg}
	method := &BlocketValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "Apple", Model: "iPhone 13"}
	v, err := method.Valuate(context.Background(), pi)
	if err != nil {
		t.Fatalf("expected no error for disabled config, got: %v", err)
	}
	if v != nil {
		t.Fatal("expected nil valuation for disabled config")
	}
}

func TestBlocketValuationMethod_ReturnsErrorWhenNoPrices(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"docs": []}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Blocket.Timeout = 5 * time.Second
	cfg.Scraping.Blocket.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &BlocketValuationMethod{svc: svc}

	pi := ProductInfo{Manufacturer: "NoSuchBrand12345", Model: "NoSuchModel12345"}
	v, err := method.Valuate(context.Background(), pi)
	if err == nil {
		t.Fatal("expected error when no prices found")
	}
	if v != nil {
		t.Fatal("expected nil valuation")
	}
}

func TestBlocketValuationMethod_CachesResult(t *testing.T) {
	valuationTypeEnabled["Blocket"] = true
	defer func() { valuationTypeEnabled["Blocket"] = false }()

	ClearBlocketCaches()

	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"docs": [{"id": "1", "price": {"amount": 1000}, "heading": "Test"}]}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Blocket.Timeout = 5 * time.Second
	cfg.Scraping.Blocket.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &BlocketValuationMethod{svc: svc}

	pi := ProductInfo{Model: "CacheOnly"}

	v1, _ := method.Valuate(context.Background(), pi)
	v2, _ := method.Valuate(context.Background(), pi)

	if v1.Value != v2.Value {
		t.Fatalf("expected cached result, got different values: %d vs %d", v1.Value, v2.Value)
	}
	if requestCount > 5 {
		t.Fatalf("expected at most 5 requests, got %d", requestCount)
	}
}

func TestBlocketValuationMethod_OnlyModelQuery(t *testing.T) {
	valuationTypeEnabled["Blocket"] = true
	defer func() { valuationTypeEnabled["Blocket"] = false }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"docs": [
				{"id": "1", "price": {"amount": 500}, "heading": "Product only"},
				{"id": "2", "price": {"amount": 600}, "heading": "Product only"},
				{"id": "3", "price": {"amount": 700}, "heading": "Product only"}
			]
		}`))
	}))
	defer ts.Close()

	cfg := &config.Config{}
	cfg.Scraping.Blocket.Timeout = 5 * time.Second
	cfg.Scraping.Blocket.BaseURL = ts.URL

	svc := &ValuationService{cfg: cfg}
	method := &BlocketValuationMethod{svc: svc}

	pi := ProductInfo{Model: "ProductWithoutBrand"}
	v, err := method.Valuate(context.Background(), pi)
	if err != nil {
		t.Fatalf("Valuate returned error: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil valuation")
	}
	if v.Value <= 0 {
		t.Fatalf("expected positive value, got %d", v.Value)
	}
}

func TestCalculateQuartiles(t *testing.T) {
	tests := []struct {
		name   string
		prices []int
		wantQ1 float64
		wantQ3 float64
	}{
		{
			name:   "simple set",
			prices: []int{100, 200, 300, 400, 500, 600, 700, 800},
			wantQ1: 200,
			wantQ3: 600,
		},
		{
			name:   "odd set",
			prices: []int{100, 200, 300, 400, 500},
			wantQ1: 100,
			wantQ3: 300,
		},
		{
			name:   "single element",
			prices: []int{500},
			wantQ1: 500,
			wantQ3: 500,
		},
		{
			name:   "empty set",
			prices: []int{},
			wantQ1: 0,
			wantQ3: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q1, q3, iqr := calculateQuartiles(tt.prices)
			if q1 != tt.wantQ1 {
				t.Errorf("calculateQuartiles() q1 = %v, want %v", q1, tt.wantQ1)
			}
			if q3 != tt.wantQ3 {
				t.Errorf("calculateQuartiles() q3 = %v, want %v", q3, tt.wantQ3)
			}
			if tt.prices != nil && len(tt.prices) > 0 {
				expectedIQR := tt.wantQ3 - tt.wantQ1
				if iqr != expectedIQR {
					t.Errorf("calculateQuartiles() iqr = %v, want %v", iqr, expectedIQR)
				}
			}
		})
	}
}

func TestFilterOutliersIQR(t *testing.T) {
	tests := []struct {
		name    string
		prices  []int
		wantLen int
	}{
		{
			name:    "no outliers",
			prices:  []int{100, 200, 300, 400, 500},
			wantLen: 5,
		},
		{
			name:    "with outliers - only very high prices filtered",
			prices:  []int{100, 200, 300, 400, 500, 10000},
			wantLen: 5,
		},
		{
			name:    "all outliers - prices outside bounds",
			prices:  []int{1, 2, 3},
			wantLen: 3,
		},
		{
			name:    "empty",
			prices:  []int{},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterOutliersIQR(tt.prices)
			if len(result) != tt.wantLen {
				t.Errorf("filterOutliersIQR() len = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name   string
		prices []int
		want   int
	}{
		{
			name:   "odd length",
			prices: []int{1, 2, 3, 4, 5},
			want:   3,
		},
		{
			name:   "even length",
			prices: []int{1, 2, 3, 4},
			want:   2,
		},
		{
			name:   "single",
			prices: []int{5},
			want:   5,
		},
		{
			name:   "empty",
			prices: []int{},
			want:   0,
		},
		{
			name:   "unsorted",
			prices: []int{5, 1, 3, 2, 4},
			want:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedian(tt.prices)
			if result != tt.want {
				t.Errorf("calculateMedian() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestCalculatePercentile(t *testing.T) {
	tests := []struct {
		name       string
		prices     []int
		percentile float64
		want       int
	}{
		{
			name:       "25th percentile",
			prices:     []int{100, 200, 300, 400, 500, 600, 700, 800},
			percentile: 0.25,
			want:       275,
		},
		{
			name:       "50th percentile (median)",
			prices:     []int{100, 200, 300, 400, 500, 600, 700, 800},
			percentile: 0.50,
			want:       450,
		},
		{
			name:       "75th percentile",
			prices:     []int{100, 200, 300, 400, 500, 600, 700, 800},
			percentile: 0.75,
			want:       625,
		},
		{
			name:       "25th percentile odd length",
			prices:     []int{100, 200, 300, 400, 500},
			percentile: 0.25,
			want:       200,
		},
		{
			name:       "empty",
			prices:     []int{},
			percentile: 0.25,
			want:       0,
		},
		{
			name:       "single element",
			prices:     []int{500},
			percentile: 0.25,
			want:       500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercentile(tt.prices, tt.percentile)
			if result != tt.want {
				t.Errorf("calculatePercentile() = %d, want %d", result, tt.want)
			}
		})
	}
}
