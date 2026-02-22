package services

import (
	"context"
	"testing"
	"time"

	"begbot/internal/config"
)

func TestFetchBlocketAdFromAPI(t *testing.T) {
	valuationTypeEnabled["Blocket"] = true
	defer func() { valuationTypeEnabled["Blocket"] = false }()

	cfg := &config.Config{}

	svc := NewMarketplaceService(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testAdIDs := []int64{
		124456789,
		124450000,
	}

	for _, adID := range testAdIDs {
		t.Run("", func(t *testing.T) {
			details, err := svc.fetchBlocketAdFromAPI(ctx, adID)
			if err != nil {
				t.Logf("API call for ad %d returned error (expected for invalid IDs): %v", adID, err)
				return
			}

			if details == nil {
				t.Logf("Ad %d not found", adID)
				return
			}

			if details.Title == "" {
				t.Error("Title should not be empty")
			}

			if details.AdText == "" {
				t.Error("AdText should not be empty")
			}

			if details.Price <= 0 {
				t.Error("Price should be greater than 0")
			}

			t.Logf("Ad %d: %s", adID, details.Title)
			t.Logf("Price: %d SEK", int(details.Price))
			t.Logf("Description length: %d chars", len(details.AdText))
		})

		time.Sleep(300 * time.Millisecond)
	}
}

func TestWaitForRateLimit(t *testing.T) {
	cfg := &config.Config{}
	svc := NewMarketplaceService(cfg)

	ctx := context.Background()

	start := time.Now()
	for i := 0; i < 5; i++ {
		err := svc.waitForRateLimit(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	elapsed := time.Since(start)
	expectedMin := time.Second / maxRequestsPerSecond * 4

	if elapsed < expectedMin {
		t.Errorf("Rate limiting not working: elapsed %v, expected at least %v", elapsed, expectedMin)
	}

	t.Logf("5 requests took %v (expected at least %v)", elapsed, expectedMin)
}

func TestExtractBlocketAdID(t *testing.T) {
	testCases := []struct {
		url      string
		expected int64
	}{
		{"https://www.blocket.se/annons/123456", 123456},
		{"https://www.blocket.se/item/999999", 999999},
		{"https://www.blocket.se/annons/123456?q=test", 123456},
		{"invalid", 0},
		{"https://www.blocket.se/other/123", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			result := extractBlocketAdID(tc.url)
			if result != tc.expected {
				t.Errorf("extractBlocketAdID(%s) = %d, want %d", tc.url, result, tc.expected)
			}
		})
	}
}

func TestExtractShippingCost(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *float64
	}{
		{
			name:     "frakt från 50 kr",
			input:    "frakt från 50 kr",
			expected: float64Ptr(50),
		},
		{
			name:     "frakt från 50 kr with non-breaking space",
			input:    "Frakt från 39\xa0kr + köpskydd 93\xa0kr",
			expected: float64Ptr(39),
		},
		{
			name:     "frakt från 50",
			input:    "frakt från 50",
			expected: float64Ptr(50),
		},
		{
			name:     "Frakt från 75 kr",
			input:    "Frakt från 75 kr",
			expected: float64Ptr(75),
		},
		{
			name:     "frakt fr 50 kr",
			input:    "frakt fr 50 kr",
			expected: float64Ptr(50),
		},
		{
			name:     "frakt fr. 50 kr",
			input:    "frakt fr. 50 kr",
			expected: float64Ptr(50),
		},
		{
			name:     "frakt 63 kr",
			input:    "frakt 63 kr",
			expected: float64Ptr(63),
		},
		{
			name:     "frakt 63",
			input:    "frakt 63",
			expected: float64Ptr(63),
		},
		{
			name:     "frakt 63:-",
			input:    "frakt 63:-",
			expected: float64Ptr(63),
		},
		{
			name:     "gratis frakt",
			input:    "gratis frakt",
			expected: float64Ptr(0),
		},
		{
			name:     "fri frakt",
			input:    "fri frakt",
			expected: float64Ptr(0),
		},
		{
			name:     "frakt ingår",
			input:    "frakt ingår",
			expected: float64Ptr(0),
		},
		{
			name:     "50 kr frakt",
			input:    "50 kr frakt",
			expected: float64Ptr(50),
		},
		{
			name:     "+ 50 kr frakt",
			input:    "+ 50 kr frakt",
			expected: float64Ptr(50),
		},
		{
			name:     "frakt: 50 kr",
			input:    "frakt: 50 kr",
			expected: float64Ptr(50),
		},
		{
			name:     "frakt:50kr",
			input:    "frakt:50kr",
			expected: float64Ptr(50),
		},
		{
			name:     "no shipping info",
			input:    "some random text without shipping",
			expected: nil,
		},
		{
			name:     "kan skickas",
			input:    "kan skickas",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractShippingCost(tc.input)
			if tc.expected == nil {
				if result != nil {
					t.Errorf("extractShippingCost(%q) = %v, want nil", tc.input, *result)
				}
				return
			}
			if result == nil {
				t.Errorf("extractShippingCost(%q) = nil, want %v", tc.input, *tc.expected)
				return
			}
			if *result != *tc.expected {
				t.Errorf("extractShippingCost(%q) = %v, want %v", tc.input, *result, *tc.expected)
			}
		})
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func TestExtractInsuranceCost(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *float64
	}{
		{
			name:     "köpskydd 73 kr",
			input:    "Frakt från 69 kr + köpskydd 73 kr",
			expected: float64Ptr(73),
		},
		{
			name:     "köpskydd 50 kr",
			input:    "köpskydd 50 kr",
			expected: float64Ptr(50),
		},
		{
			name:     "köpskydd 50",
			input:    "köpskydd 50",
			expected: float64Ptr(50),
		},
		{
			name:     "köpskydd: 75 kr",
			input:    "köpskydd: 75 kr",
			expected: float64Ptr(75),
		},
		{
			name:     "köpskydd 75:-",
			input:    "köpskydd 75:-",
			expected: float64Ptr(75),
		},
		{
			name:     "försäkring 100 kr",
			input:    "försäkring 100 kr",
			expected: float64Ptr(100),
		},
		{
			name:     "försäkring 100",
			input:    "försäkring 100",
			expected: float64Ptr(100),
		},
		{
			name:     "försäkring: 80 kr",
			input:    "försäkring: 80 kr",
			expected: float64Ptr(80),
		},
		{
			name:     "försäkring 80:-",
			input:    "försäkring 80:-",
			expected: float64Ptr(80),
		},
		{
			name:     "no insurance",
			input:    "frakt från 50 kr",
			expected: nil,
		},
		{
			name:     "only shipping text",
			input:    "kan skickas",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractInsuranceCost(tc.input)
			if tc.expected == nil {
				if result != nil {
					t.Errorf("extractInsuranceCost(%q) = %v, want nil", tc.input, *result)
				}
				return
			}
			if result == nil {
				t.Errorf("extractInsuranceCost(%q) = nil, want %v", tc.input, *tc.expected)
				return
			}
			if *result != *tc.expected {
				t.Errorf("extractInsuranceCost(%q) = %v, want %v", tc.input, *result, *tc.expected)
			}
		})
	}
}
