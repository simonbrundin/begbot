package marketplaces

import (
	"context"
)

// MarketplaceProvider is the pluggable provider interface for different
// marketplace implementations. Implementations should support querying by
// free text and/or by a full search URL.
type MarketplaceProvider interface {
	Name() string
	// SupportsQuery returns true if the provider can accept a free-text query
	// and translate it to search results (e.g. API/SOAP clients).
	SupportsQuery() bool
	// FetchAdsFromQuery performs a search using a free-text query.
	FetchAdsFromQuery(ctx context.Context, query string) ([]RawAd, error)
	// FetchAdsFromURL performs a search using a full search URL (HTML scraping
	// or proxy-based fetchers typically implement this).
	FetchAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error)
	// FetchAdDetails returns detailed information for a single ad URL.
	FetchAdDetails(ctx context.Context, adURL string) (*AdDetails, error)
}
