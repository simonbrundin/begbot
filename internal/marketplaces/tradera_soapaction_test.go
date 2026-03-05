package marketplaces

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"begbot/internal/config"
)

// Verify that SOAPAction header is quoted for both Search and GetItem requests.
func TestSOAPActionHeader_Quoted(t *testing.T) {
	cfg := &config.TraderaConfig{
		Timeout: 1 * time.Second,
		AppID:   "appid",
		AppKey:  "appkey",
	}
	tc := NewTraderaClient(cfg)

	var capturedSearch string
	var capturedGetItem string

	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// capture SOAPAction header
		sa := req.Header.Get("SOAPAction")
		if strings.Contains(req.URL.Path, "searchservice.asmx") || strings.Contains(req.URL.Path, SearchServicePath) {
			capturedSearch = sa
			// return minimal SOAP SearchResponse
			body := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SearchResponse>
      <SearchResult>
        <Items>
          <Item>
            <Id>1</Id>
            <Name>Test</Name>
            <CurrentPrice>1000</CurrentPrice>
            <ThumbnailUrl>https://i.test/img.jpg</ThumbnailUrl>
            <CategoryId>10</CategoryId>
          </Item>
        </Items>
      </SearchResult>
    </SearchResponse>
  </soap:Body>
</soap:Envelope>`
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(body)), Header: make(http.Header)}, nil
		}

		if strings.Contains(req.URL.Path, "PublicService.asmx") || strings.Contains(req.URL.Path, "/PublicService.asmx") {
			capturedGetItem = sa
			body := `<?xml version="1.0" encoding="utf-8"?>
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
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(body)), Header: make(http.Header)}, nil
		}

		// default
		return &http.Response{StatusCode: 404, Body: ioutil.NopCloser(bytes.NewBufferString("")), Header: make(http.Header)}, nil
	})

	tc.client = &http.Client{Transport: rt, Timeout: 2 * time.Second}

	// Call FetchAds -> should call search path
	_, err := tc.FetchAds(context.Background(), "q")
	if err != nil {
		t.Fatalf("FetchAds failed: %v", err)
	}
	if capturedSearch == "" {
		t.Fatalf("search request did not include SOAPAction header")
	}
	if capturedSearch != `"http://api.tradera.com/Search"` {
		t.Fatalf("expected quoted SOAPAction for Search, got: %s", capturedSearch)
	}

	// Call FetchAdDetails -> should call PublicService path
	_, err = tc.FetchAdDetails(context.Background(), "https://www.tradera.com/item/555/1/x")
	if err != nil {
		t.Fatalf("FetchAdDetails failed: %v", err)
	}
	if capturedGetItem == "" {
		t.Fatalf("GetItem request did not include SOAPAction header")
	}
	if capturedGetItem != `"http://api.tradera.com/GetItem"` {
		t.Fatalf("expected quoted SOAPAction for GetItem, got: %s", capturedGetItem)
	}
}

// reuses roundTripperFunc helper from tradera_cache_test.go
