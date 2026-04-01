package marketplaces

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type MarketplaceService interface {
	FetchAds(ctx context.Context, query string) ([]RawAd, error)
}

func ConvertRawAd(from RawAd, to *RawAd) {
	to.Link = from.Link
	to.Title = from.Title
	to.Price = from.Price
	to.AdText = from.AdText
	to.ImageURLs = from.ImageURLs
	to.AdDate = from.AdDate
	to.Marketplace = from.Marketplace
	to.ShippingCost = from.ShippingCost
	to.ShippingInsurance = from.ShippingInsurance
}

func ExtractQueryFromURL(searchURL string) (string, error) {
	u, err := url.Parse(searchURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}
	q := u.Query().Get("q")
	if q == "" {
		return "", fmt.Errorf("no query parameter found in URL")
	}
	return strings.TrimSpace(q), nil
}

func convertMarketplacesRawAds(ads []RawAd) []RawAd {
	result := make([]RawAd, len(ads))
	for i, ad := range ads {
		ConvertRawAd(ad, &result[i])
	}
	return result
}
