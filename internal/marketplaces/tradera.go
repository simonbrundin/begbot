package marketplaces

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"begbot/internal/config"
)

const (
	TraderaAPIBaseURL = "https://api.tradera.com/v3"
	SearchServicePath = "/searchservice.asmx"

	MaxAPICallsPerDay = 80
	MinRequestDelay   = 2 * time.Second
)

type TraderaClient struct {
	cfg            *config.TraderaConfig
	client         *http.Client
	cache          *EnrichCache
	lastRequest    time.Time
	dailyCallCount int
	resetTime      time.Time
}

func NewTraderaClient(cfg *config.TraderaConfig) *TraderaClient {
	now := time.Now()
	return &TraderaClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		cache:          NewEnrichCache(cfg.EnrichCacheTTL),
		lastRequest:    now,
		dailyCallCount: 0,
		resetTime:      now.Add(24 * time.Hour),
	}
}

// WrapTransport allows callers to wrap the underlying HTTP transport used by
// the TraderaClient. The wrapper receives the current base transport and
// should return a RoundTripper that will be used for future requests. Use
// this for diagnostics (logging) or custom proxying in probe/debug modes.
func (t *TraderaClient) WrapTransport(wrapper func(base http.RoundTripper) http.RoundTripper) {
	if t.client == nil {
		t.client = &http.Client{Timeout: t.cfg.Timeout}
	}
	base := http.DefaultTransport
	if t.client.Transport != nil {
		base = t.client.Transport
	}
	t.client.Transport = wrapper(base)
}

func (t *TraderaClient) waitForRateLimit() {
	now := time.Now()

	if now.After(t.resetTime) {
		t.dailyCallCount = 0
		t.resetTime = now.Add(24 * time.Hour)
	}

	if t.dailyCallCount >= MaxAPICallsPerDay {
		sleepDuration := time.Until(t.resetTime)
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
			t.dailyCallCount = 0
			t.resetTime = time.Now().Add(24 * time.Hour)
		}
	}

	timeSinceLastRequest := time.Since(t.lastRequest)
	if timeSinceLastRequest < MinRequestDelay {
		time.Sleep(MinRequestDelay - timeSinceLastRequest)
	}
}

func (t *TraderaClient) recordAPICall() {
	t.dailyCallCount++
	t.lastRequest = time.Now()
}

func (t *TraderaClient) Name() string {
	return "tradera"
}

// The existing TraderaClient implements MarketplaceProvider-style methods
// partially via FetchAds/FetchAdDetails. Add adapter methods that match the
// new provider interface so it can be used by the orchestrator.
func (t *TraderaClient) SupportsQuery() bool { return true }

func (t *TraderaClient) FetchAdsFromQuery(ctx context.Context, query string) ([]RawAd, error) {
	// Delegate to existing FetchAds which accepts a query string.
	return t.FetchAds(ctx, query)
}

func (t *TraderaClient) FetchAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error) {
	// The SOAP API expects queries, not full URLs. There's no meaningful
	// implementation here, but to satisfy the interface we attempt to
	// extract a query from the searchURL (if possible) and fall back to an
	// empty query which will return an error.
	// A more robust approach would be to let the orchestrator call the
	// HTML scrapers for URL-based fetching; this adapter remains API-focused.
	// Attempt naive extraction: look for q= param
	u, err := url.Parse(searchURL)
	if err != nil {
		return nil, fmt.Errorf("tradera: failed to parse search URL: %w", err)
	}
	q := u.Query().Get("q")
	if q == "" {
		return nil, fmt.Errorf("tradera: no query param found in URL: %s", searchURL)
	}
	return t.FetchAds(ctx, q)
}

// Note: FetchAdDetails is implemented further below (uses SOAP GetItem).

func (t *TraderaClient) FetchAds(ctx context.Context, query string) ([]RawAd, error) {
	if t.cfg.AppID == "" || t.cfg.AppKey == "" {
		return nil, fmt.Errorf("Tradera API credentials not configured")
	}

	t.waitForRateLimit()

	searchReq := SearchRequest{
		XMLName:  xml.Name{Local: "SearchRequest"},
		Query:    query,
		PageInfo: PageInfo{PageNumber: 1, PageSize: 50},
		OrderBy:  "Relevance",
	}

	soapBody := BuildSearchSoapRequest(searchReq, t.cfg.AppID, t.cfg.AppKey)

	req, err := http.NewRequestWithContext(ctx, "POST", TraderaAPIBaseURL+SearchServicePath+"?appId="+t.cfg.AppID+"&appKey="+t.cfg.AppKey, strings.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	// Some SOAP endpoints expect the SOAPAction value to be quoted.
	req.Header.Set("SOAPAction", "\"http://api.tradera.com/Search\"")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Tradera API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	// Strip namespaces from XML before parsing
	bodyWithoutNS := stripNamespace(body)

	// Use simple regex to extract item data since XML parsing is complex
	re := regexp.MustCompile(`<Id>([^<]+)</Id><ShortDescription>([^<]*)</ShortDescription>.*?<MaxBid>([^<]*)</MaxBid>.*?<ThumbnailLink>([^<]*)</ThumbnailLink>.*?<CategoryId>([^<]*)</CategoryId>.*?<BidCount>([^<]*)</BidCount>`)
	matches := re.FindAllStringSubmatch(string(bodyWithoutNS), -1)
	var items []SearchItem
	for _, m := range matches {
		if len(m) >= 6 {
			id, _ := strconv.Atoi(m[1])
			catID, _ := strconv.Atoi(m[5])
			price, _ := strconv.Atoi(m[3])
			bidCount, _ := strconv.Atoi(m[6])
			items = append(items, SearchItem{
				ID:           id,
				Name:         m[2],
				Description:  m[2],
				CurrentPrice: price,
				ThumbnailURL: m[4],
				CategoryID:   catID,
				BidCount:     bidCount,
			})
		}
	}
	searchResp := SearchResponse{Items: items}

	var ads []RawAd
	for _, item := range searchResp.Items {
		current := float64(item.CurrentPrice) / 100
		link := fmt.Sprintf("https://www.tradera.com/item/%d/%d/%s", item.CategoryID, item.ID, url.PathEscape(item.Name))
		if item.CategoryID == 0 {
			link = fmt.Sprintf("https://www.tradera.com/item/0/%d/%s", item.ID, url.PathEscape(item.Name))
		}
		ad := RawAd{
			Link:         link,
			Title:        item.Name,
			Price:        current,
			Marketplace:  "tradera",
			AdText:       item.Description,
			CurrentPrice: current,
			HasBuyNow:    false,
			BuyNowPrice:  0,
		}
		if item.ThumbnailURL != "" {
			ad.ImageURLs = []string{item.ThumbnailURL}
		}

		// If Search item signals auction (via BidCount or ItemType), we leave HasBuyNow=false.
		// If Search response doesn't include buy-now and BidCount==0, we may need to call GetItem
		// for definitive data. Defer enrichment to caller to avoid extra API calls here.

		ads = append(ads, ad)
	}

	t.recordAPICall()

	// Enrichment: for ads where HasBuyNow is false and either CurrentPrice==0
	// or BuyNow is unknown, call GetItem for a limited subset to fetch definitive info.
	// Respect configured enrichment limit when available
	maxEnrich := 10
	if t.cfg != nil && t.cfg.EnrichLimit > 0 {
		maxEnrich = t.cfg.EnrichLimit
	}
	enriched := 0
	for i := range ads {
		if enriched >= maxEnrich {
			break
		}
		a := &ads[i]
		// If buy-now already detected skip
		if a.HasBuyNow {
			continue
		}
		// If current price is zero or we want to verify buy-now, attempt enrichment
		if a.CurrentPrice == 0 || a.Price == 0 {
			if details, err := t.FetchAdDetails(ctx, a.Link); err == nil && details != nil {
				// Update current price
				if details.Price > 0 {
					a.CurrentPrice = details.Price
					a.Price = details.Price
				}
				// If GetItem returned a buy-now price, set HasBuyNow
				if details.BuyNowPrice > 0 {
					a.BuyNowPrice = details.BuyNowPrice
					a.HasBuyNow = true
					a.Price = details.BuyNowPrice
				}
				// If GetItem returned shipping info, set it
				if details.ShippingCost != nil {
					a.ShippingCost = details.ShippingCost
				}
			}
			enriched++
		}
	}

	return ads, nil
}

func (t *TraderaClient) FetchAdDetails(ctx context.Context, adURL string) (*AdDetails, error) {
	itemID, err := extractTraderaItemID(adURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract item ID: %w", err)
	}

	// Check cache first
	if t.cache != nil {
		if cached, ok := t.cache.Get(itemID); ok {
			return cached, nil
		}
	}

	if t.cfg.AppID == "" || t.cfg.AppKey == "" {
		return nil, fmt.Errorf("Tradera API credentials not configured")
	}

	t.waitForRateLimit()

	soapBody := BuildGetItemSoapRequest(itemID, t.cfg.AppID, t.cfg.AppKey)

	req, err := http.NewRequestWithContext(ctx, "POST", TraderaAPIBaseURL+"/PublicService.asmx"+"?"+url.Values{"appId": {t.cfg.AppID}, "appKey": {t.cfg.AppKey}}.Encode(), strings.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	// Quote SOAPAction to satisfy strict HTTP header parsing on Tradera's PublicService
	req.Header.Set("SOAPAction", "\"http://api.tradera.com/GetItem\"")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Tradera API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var itemResp GetItemResponse
	if err := xml.Unmarshal(body, &itemResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if itemResp.Item.ID == 0 {
		return nil, fmt.Errorf("item not found")
	}

	details := &AdDetails{
		Title:       itemResp.Item.Name,
		Price:       float64(itemResp.Item.CurrentPrice) / 100,
		AdText:      itemResp.Item.Description,
		ImageURLs:   itemResp.Item.AllMedia,
		SellerName:  itemResp.Item.Seller.Nick,
		BuyNowPrice: float64(itemResp.Item.BuyNowPrice) / 100,
	}
	// Add shipping cost if available
	if itemResp.Item.Shipping.ShippingCost > 0 {
		cost := float64(itemResp.Item.Shipping.ShippingCost) / 100
		details.ShippingCost = &cost
	}

	t.recordAPICall()

	// Store in cache if available
	if t.cache != nil {
		t.cache.Set(itemID, details)
	}
	return details, nil
}

func extractTraderaItemID(adURL string) (int, error) {
	parts := strings.Split(adURL, "/")
	for i, part := range parts {
		if part == "item" {
			// Try a few candidate positions after "item" to find a numeric ID
			for j := i + 1; j < len(parts) && j <= i+3; j++ {
				if parts[j] == "" {
					continue
				}
				if id, err := strconv.Atoi(parts[j]); err == nil {
					return id, nil
				}
			}
		}
	}
	// Fallback: try to find the first numeric sequence in the URL
	re := regexp.MustCompile(`(\d+)`)
	m := re.FindStringSubmatch(adURL)
	if len(m) > 1 {
		if id, err := strconv.Atoi(m[1]); err == nil {
			return id, nil
		}
	}
	return 0, fmt.Errorf("could not extract item ID from URL: %s", adURL)
}

type PageInfo struct {
	PageNumber int `xml:"PageNumber"`
	PageSize   int `xml:"PageSize"`
}

type SearchRequest struct {
	XMLName  xml.Name `xml:"SearchRequest"`
	Query    string   `xml:"Query"`
	PageInfo PageInfo `xml:"PageInfo"`
	OrderBy  string   `xml:"OrderBy"`
}

type SearchResponse struct {
	XMLName xml.Name     `xml:"Envelope"`
	Items   []SearchItem `xml:"Body>SearchResponse>SearchResult>Items>Item"`
}

type SearchItem struct {
	ID           int    `xml:"Id"`
	Name         string `xml:"Name"`
	Description  string `xml:"Description"`
	CurrentPrice int    `xml:"CurrentPrice"`
	// Note: SOAP Search response may not include buy-now info; keep ItemType
	// and BidCount so callers can decide if further GetItem is required.
	ThumbnailURL string `xml:"ThumbnailUrl"`
	CategoryID   int    `xml:"CategoryId"`
	EndDate      string `xml:"EndDate"`
	ItemType     string `xml:"ItemType"`
	BidCount     int    `xml:"BidCount"`
}

type GetItemResponse struct {
	XMLName xml.Name    `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Item    GetItemData `xml:"Body>GetItemResponse>GetItemResult"`
}

type GetItemData struct {
	ID           int    `xml:"Id"`
	Name         string `xml:"Name"`
	Description  string `xml:"Description"`
	CurrentPrice int    `xml:"CurrentPrice"`
	// Some Tradera GetItem responses include explicit buy-now price
	BuyNowPrice   int      `xml:"BuyNowPrice"`
	AllMedia      []string `xml:"AllMedia>string"`
	Seller        Seller   `xml:"Seller"`
	Shipping      Shipping `xml:"Shipping"`
	ItemCondition string   `xml:"ItemCondition"`
}

type Seller struct {
	Nick string `xml:"Nick"`
}

type Shipping struct {
	ShippingCost int `xml:"ShippingCost"`
}

func BuildSearchSoapRequest(req SearchRequest, appID, appKey string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <TraderaHeader xmlns="http://api.tradera.com">
      <AppId>%s</AppId>
      <AppKey>%s</AppKey>
    </TraderaHeader>
  </soap:Header>
  <soap:Body>
    <Search xmlns="http://api.tradera.com">
      <query>%s</query>
      <categoryId>0</categoryId>
      <pageNumber>%d</pageNumber>
      <orderBy>%s</orderBy>
    </Search>
  </soap:Body>
</soap:Envelope>`, appID, appKey, req.Query, req.PageInfo.PageNumber, req.OrderBy)
}

func BuildGetItemSoapRequest(itemID int, appID, appKey string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Header>
    <TraderaHeader xmlns="http://api.tradera.com">
      <AppId>%s</AppId>
      <AppKey>%s</AppKey>
    </TraderaHeader>
  </soap:Header>
  <soap:Body>
    <GetItem xmlns="http://api.tradera.com">
      <itemId>%d</itemId>
    </GetItem>
  </soap:Body>
</soap:Envelope>`, appID, appKey, itemID)
}

func stripNamespace(xmlContent []byte) []byte {
	content := string(xmlContent)
	content = strings.ReplaceAll(content, ` xmlns="http://api.tradera.com"`, "")
	content = strings.ReplaceAll(content, ` xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"`, "")
	content = strings.ReplaceAll(content, ` xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`, "")
	content = strings.ReplaceAll(content, ` xmlns:xsd="http://www.w3.org/2001/XMLSchema"`, "")
	return []byte(content)
}

func init() {
	_ = time.Duration(0)
}
