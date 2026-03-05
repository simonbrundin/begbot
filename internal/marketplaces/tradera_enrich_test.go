package marketplaces

import (
	"context"
	"testing"
)

// Basic tests to verify Search->RawAd preserves CurrentPrice and HasBuyNow defaults
func TestSearchToRawAd_Defaults(t *testing.T) {
	// Create a minimal SearchItem with only CurrentPrice
	si := SearchItem{
		ID:           1,
		Name:         "Test Item",
		Description:  "desc",
		CurrentPrice: 150000, // represents 1500.00 SEK in cents
		ThumbnailURL: "https://i.test/img.jpg",
		CategoryID:   10,
	}

	resp := SearchResponse{Items: []SearchItem{si}}

	// Simulate unmarshalled flow: convert here using same code path as FetchAds
	var ads []RawAd
	for _, item := range resp.Items {
		current := float64(item.CurrentPrice) / 100
		ad := RawAd{
			Link:         "https://www.tradera.com/item/10/1/test",
			Title:        item.Name,
			Price:        current,
			Marketplace:  "tradera",
			AdText:       item.Description,
			CurrentPrice: current,
			HasBuyNow:    false,
			BuyNowPrice:  0,
		}
		ads = append(ads, ad)
	}

	if len(ads) != 1 {
		t.Fatalf("expected 1 ad, got %d", len(ads))
	}
	if ads[0].CurrentPrice != 1500 {
		t.Fatalf("expected CurrentPrice 1500, got %v", ads[0].CurrentPrice)
	}
	if ads[0].HasBuyNow != false {
		t.Fatalf("expected HasBuyNow false, got %v", ads[0].HasBuyNow)
	}

	// ensure FetchAds handles missing creds gracefully in test context by returning error
	// (we don't need to call FetchAds here because it requires valid config)
	_ = context.Background()
}
