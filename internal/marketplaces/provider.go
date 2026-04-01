package marketplaces

import (
	"context"
	"time"
)

type RawAd struct {
	Link              string
	Title             string
	Price             float64
	AdText            string
	ImageURLs         []string
	AdDate            time.Time
	Marketplace       string
	ShippingCost      *float64
	ShippingInsurance *float64
}

type MarketplaceProvider interface {
	Name() string
	SupportsQuery() bool
	FetchAdsFromQuery(ctx context.Context, query string) ([]RawAd, error)
	FetchAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error)
}
