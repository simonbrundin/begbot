package marketplaces

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		lastRequest:    now,
		dailyCallCount: 0,
		resetTime:      now.Add(24 * time.Hour),
	}
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

func (t *TraderaClient) SupportsQuery() bool { return true }

func (t *TraderaClient) FetchAdsFromQuery(ctx context.Context, query string) ([]RawAd, error) {
	return t.FetchAds(ctx, query)
}

func (t *TraderaClient) FetchAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error) {
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

	req, err := http.NewRequestWithContext(ctx, "POST", TraderaAPIBaseURL+SearchServicePath, strings.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://api.tradera.com/Search/Search")

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

	var searchResp SearchResponse
	if err := xml.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var ads []RawAd
	for _, item := range searchResp.Items {
		ad := RawAd{
			Link:        fmt.Sprintf("https://www.tradera.com/item/%d/%d/%s", item.CategoryID, item.ID, url.PathEscape(item.Name)),
			Title:       item.Name,
			Price:       float64(item.CurrentPrice) / 100,
			Marketplace: "tradera",
			AdText:      item.Description,
		}
		if item.ThumbnailURL != "" {
			ad.ImageURLs = []string{item.ThumbnailURL}
		}
		ads = append(ads, ad)
	}

	t.recordAPICall()
	return ads, nil
}

func (t *TraderaClient) FetchAdDetails(ctx context.Context, adURL string) (*AdDetails, error) {
	itemID, err := extractTraderaItemID(adURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract item ID: %w", err)
	}

	if t.cfg.AppID == "" || t.cfg.AppKey == "" {
		return nil, fmt.Errorf("Tradera API credentials not configured")
	}

	t.waitForRateLimit()

	soapBody := BuildGetItemSoapRequest(itemID, t.cfg.AppID, t.cfg.AppKey)

	req, err := http.NewRequestWithContext(ctx, "POST", TraderaAPIBaseURL+"/PublicService.asmx", strings.NewReader(soapBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://api.tradera.com/PublicService/GetItem")

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
		Title:      itemResp.Item.Name,
		Price:      float64(itemResp.Item.CurrentPrice) / 100,
		AdText:     itemResp.Item.Description,
		ImageURLs:  itemResp.Item.AllMedia,
		SellerName: itemResp.Item.Seller.Nick,
	}

	t.recordAPICall()
	return details, nil
}

func extractTraderaItemID(adURL string) (int, error) {
	parts := strings.Split(adURL, "/")
	for i, part := range parts {
		if part == "item" && i+1 < len(parts) {
			return strconv.Atoi(parts[i+1])
		}
	}
	return 0, fmt.Errorf("could not extract item ID from URL: %s", adURL)
}

type AdDetails struct {
	Title      string
	Price      float64
	AdText     string
	ImageURLs  []string
	SellerName string
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
	XMLName xml.Name     `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Items   []SearchItem `xml:"SearchResult>Items>Item"`
}

type SearchItem struct {
	ID           int    `xml:"Id"`
	Name         string `xml:"Name"`
	Description  string `xml:"Description"`
	CurrentPrice int    `xml:"CurrentPrice"`
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
	ID            int      `xml:"Id"`
	Name          string   `xml:"Name"`
	Description   string   `xml:"Description"`
	CurrentPrice  int      `xml:"CurrentPrice"`
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
