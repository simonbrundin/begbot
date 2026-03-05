package services

import (
	"context"
	"testing"

	"begbot/internal/marketplaces"
	"begbot/internal/models"
)

// Fake tradera client used in tests
type fakeTraderaClient struct{}

func (f *fakeTraderaClient) FetchAds(ctx context.Context, query string) ([]marketplaces.RawAd, error) {
	return nil, nil
}

func (f *fakeTraderaClient) FetchAdDetails(ctx context.Context, adURL string) (*marketplaces.AdDetails, error) {
	return &marketplaces.AdDetails{BuyNowPrice: 799.0, Price: 799.0}, nil
}

type fakeDB struct{}

func (f *fakeDB) GetEmailSettings(ctx context.Context) (*models.EmailSettings, error) {
	return nil, nil
}

// Implement only used methods with no-ops to satisfy calls in SendTradingRuleEmail path
func (f *fakeDB) ComputeWeightedValuationForProduct(ctx context.Context, productID int64) (int, float64, error) {
	return 2000, 0.0, nil
}
func (f *fakeDB) GetImageLinks(ctx context.Context, listingID int64) ([]string, error) {
	return []string{""}, nil
}
func (f *fakeDB) CreateSentEmail(ctx context.Context, sent *models.SentEmail) error { return nil }

// fakeLLM provides minimal LLMService methods used by BotService in tests
type fakeLLM struct{}

func (f *fakeLLM) ExtractProductInfo(ctx context.Context, title, adText, link string) (*ProductInfo, error) {
	return &ProductInfo{Category: ""}, nil
}

func (f *fakeLLM) EvaluateProductIntact(ctx context.Context, adText, title string) (*ProductIntactResult, error) {
	return &ProductIntactResult{IsIntact: true, Reasoning: ""}, nil
}

// minimal subset of db.Postgres used in BotService - we only need a few methods
// so create a small adapter by embedding the fake functions via interface conversion

func TestBotService_SkipsEmailForTraderaAuctionOnly(t *testing.T) {
	// Prepare BotService with fake dependencies
	bs := &BotService{}
	bs.database = nil // SendTradingRuleEmail uses database if present; we'll rely on defaults

	// Create a Tradera RawAd that is auction-only (no buy-now)
	ad := RawAd{
		Link:         "https://www.tradera.com/item/999",
		Title:        "Auction only",
		Price:        500,
		Marketplace:  "tradera",
		HasBuyNow:    false,
		BuyNowPrice:  0,
		CurrentPrice: 500,
	}

	// Create a dummy listing & product to pass to decision function
	listing := &models.Listing{ID: 1, Title: "Auction only", Price: func() *int { v := 500; return &v }()}
	product := &models.Product{ID: 1}

	// Directly test the decision helper: shouldSendTradingRuleEmail
	send := bs.shouldSendTradingRuleEmail(ad, listing, product)
	if send {
		t.Fatalf("expected shouldSendTradingRuleEmail=false for auction-only Tradera ad")
	}
}

func TestBotService_EnrichmentFindsBuyNowAndAllowsEmail(t *testing.T) {
	// Prepare BotService with fake dependencies and a fake tradera client that returns buy-now on FetchAdDetails
	bs := &BotService{}
	bs.database = nil

	// Attach marketplace service with fake client
	ms := &MarketplaceService{traderaClient: &fakeTraderaClient{}}
	bs.marketplaceService = ms

	// Create auction-only RawAd
	ad := RawAd{
		Link:         "https://www.tradera.com/item/123",
		Title:        "Might have buy now",
		Price:        500,
		Marketplace:  "tradera",
		HasBuyNow:    false,
		BuyNowPrice:  0,
		CurrentPrice: 500,
	}

	listing := &models.Listing{ID: 2, Title: "Might have buy now", Price: func() *int { v := 500; return &v }()}
	product := &models.Product{ID: 2}

	send := bs.shouldSendTradingRuleEmail(ad, listing, product)
	if !send {
		t.Fatalf("expected shouldSendTradingRuleEmail=true after enrichment found buy-now")
	}
}
