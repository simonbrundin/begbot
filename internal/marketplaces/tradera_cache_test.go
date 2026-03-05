package marketplaces

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"begbot/internal/config"
)

// Verify basic TTL semantics for the enrich cache
func TestEnrichCache_TTL(t *testing.T) {
	c := NewEnrichCache(100 * time.Millisecond)
	details := &AdDetails{Title: "cached", Price: 12.34}
	c.Set(42, details)

	if d, ok := c.Get(42); !ok || d == nil || d.Title != "cached" {
		t.Fatalf("expected cache hit for id=42")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)
	if d, ok := c.Get(42); ok || d != nil {
		t.Fatalf("expected cache miss after TTL expiry")
	}
}

// Verify TraderaClient.FetchAdDetails returns cached value and stores fetched
// value in cache after a network call.
func TestTraderaClient_FetchAdDetails_Caches(t *testing.T) {
	cfg := &config.TraderaConfig{
		Timeout:        1 * time.Second,
		AppID:          "",
		AppKey:         "",
		EnrichLimit:    1,
		EnrichCacheTTL: 1 * time.Hour,
	}
	tc := NewTraderaClient(cfg)

	// 1) Ensure cache hit works even when credentials are empty (cache check happens first)
	tc.cache.Set(123, &AdDetails{Title: "from-cache", Price: 99.9})
	d, err := tc.FetchAdDetails(context.Background(), "https://www.tradera.com/item/123/1/x")
	if err != nil {
		t.Fatalf("unexpected error fetching cached details: %v", err)
	}
	if d.Title != "from-cache" {
		t.Fatalf("expected cached title, got %q", d.Title)
	}

	// 2) Now test that a network fetch stores into cache. Prepare a fake transport
	// that returns a minimal SOAP GetItem response for item id 555.
	sampleResp := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <GetItemResponse>
      <GetItemResult>
        <Id>555</Id>
        <Name>Net Item</Name>
        <Description>desc</Description>
        <CurrentPrice>12345</CurrentPrice>
        <BuyNowPrice>54321</BuyNowPrice>
        <AllMedia>
          <string>https://i.example/img1.jpg</string>
        </AllMedia>
        <Seller>
          <Nick>seller1</Nick>
        </Seller>
      </GetItemResult>
    </GetItemResponse>
  </soap:Body>
</soap:Envelope>`

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})
	tc.client = &http.Client{Transport: rt, Timeout: 2 * time.Second}

	// Provide fake credentials so FetchAdDetails will proceed to network call
	tc.cfg.AppID = "a"
	tc.cfg.AppKey = "k"

	d2, err := tc.FetchAdDetails(context.Background(), "https://www.tradera.com/item/555/1/net")
	if err != nil {
		t.Fatalf("unexpected error fetching network details: %v", err)
	}
	if d2.Title != "Net Item" {
		t.Fatalf("unexpected title from network response: %q", d2.Title)
	}

	// Ensure result was cached
	if cached, ok := tc.cache.Get(555); !ok || cached == nil || cached.Title != "Net Item" {
		t.Fatalf("expected network response to be cached")
	}
}

// small helper to implement http.RoundTripper from a function
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
