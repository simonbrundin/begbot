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

type AdDetails struct {
	Title             string
	Price             float64
	AdText            string
	ImageURLs         []string
	SellerName        string
	SellerRating      float64
	ShippingCost      *float64
	ShippingInsurance *float64
	ItemCondition     string
}

type MarketplaceFetcher interface {
	Name() string
	FetchAds(ctx context.Context, query string) ([]RawAd, error)
	FetchAdDetails(ctx context.Context, adURL string) (*AdDetails, error)
}
