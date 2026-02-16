package services

import (
	"context"
	"testing"
	"time"

	"begbot/internal/config"
)

func TestFetchBlocketAdFromAPI(t *testing.T) {
	cfg := &config.Config{
		Scraping: config.ScrapingConfig{
			Blocket: config.BlocketConfig{
				Enabled: true,
			},
		},
	}

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
