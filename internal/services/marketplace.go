package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"begbot/internal/config"
	"begbot/internal/models"

	"github.com/PuerkitoBio/goquery"
)

type MarketplaceService struct {
	cfg         *config.Config
	lastReqTime time.Time
}

func NewMarketplaceService(cfg *config.Config) *MarketplaceService {
	return &MarketplaceService{cfg: cfg}
}

type RawAd struct {
	Link         string
	Title        string
	Price        float64
	AdText       string
	ImageURLs    []string
	AdDate       time.Time
	Marketplace  string
	ShippingCost *float64 // NULL if unknown, 0 if free, positive value if specified
}

// FetchAdDetails fetches detailed information from an individual ad page
func (s *MarketplaceService) FetchAdDetails(ctx context.Context, adURL string) (*BlocketAdDetails, error) {
	adID := extractBlocketAdID(adURL)
	if adID == 0 {
		return nil, fmt.Errorf("could not extract ad ID from URL: %s", adURL)
	}
	return s.fetchBlocketAdFromAPI(ctx, adID)
}

func parseBlocketAdPage(body []byte, adURL string) (*RawAd, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ad HTML: %w", err)
	}

	ad := &RawAd{
		Link:        adURL,
		Marketplace: "blocket",
	}

	// Extract title
	ad.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	if ad.Title == "" {
		ad.Title = strings.TrimSpace(doc.Find("[data-test='subject']").First().Text())
	}

	// Extract description - look for body/description section
	description := ""
	doc.Find("[data-test='body'], .body, .description, [itemprop='description']").Each(func(i int, s *goquery.Selection) {
		if description == "" {
			description = strings.TrimSpace(s.Text())
		}
	})

	// If no description found, try to get any text content
	if description == "" {
		// Try to find main content area
		doc.Find("main, article, .main-content, #main-content").Each(func(i int, s *goquery.Selection) {
			if description == "" {
				description = strings.TrimSpace(s.Text())
			}
		})
	}

	// Last resort: get all text from body
	if description == "" {
		description = strings.TrimSpace(doc.Find("body").Text())
	}

	ad.AdText = description

	// Extract price
	priceText := doc.Find("[data-test='price'], .price").First().Text()
	ad.Price = parsePrice(priceText)

	// Extract shipping cost from ad page
	shippingText := ""
	doc.Find("[data-test='shipping-section'], .shipping-info").Each(func(i int, s *goquery.Selection) {
		if shippingText == "" {
			shippingText = strings.ToLower(strings.TrimSpace(s.Text()))
		}
	})

	if shippingText == "" {
		doc.Find("p, span, div").Each(func(i int, s *goquery.Selection) {
			if shippingText == "" {
				text := strings.ToLower(strings.TrimSpace(s.Text()))
				if strings.Contains(text, "frakt") || strings.Contains(text, "skickas") {
					shippingText = text
				}
			}
		})
	}

	if shippingText != "" {
		ad.ShippingCost = extractShippingCost(shippingText)
	}

	return ad, nil
}

func (s *MarketplaceService) FetchAds(ctx context.Context, query string) ([]RawAd, error) {
	var ads []RawAd

	if s.cfg.Scraping.Tradera.Enabled {
		traderaAds, err := s.fetchTraderaAds(ctx, query)
		if err != nil {
			return nil, err
		}
		ads = append(ads, traderaAds...)
	}

	if s.cfg.Scraping.Blocket.Enabled {
		blocketAds, err := s.fetchBlocketAds(ctx, query)
		if err != nil {
			return nil, err
		}
		ads = append(ads, blocketAds...)
	}

	return ads, nil
}

func (s *MarketplaceService) fetchTraderaAds(ctx context.Context, query string) ([]RawAd, error) {
	url := fmt.Sprintf("https://www.tradera.com/search?q=%s", strings.ReplaceAll(query, " ", "+"))
	return s.fetchTraderaAdsFromURL(ctx, url)
}

func (s *MarketplaceService) fetchBlocketAds(ctx context.Context, query string) ([]RawAd, error) {
	url := fmt.Sprintf("https://blocket.se/recommerce/forsale/search?q=%s", strings.ReplaceAll(query, " ", "+"))
	return s.fetchBlocketAdsFromURL(ctx, url)
}

func (s *MarketplaceService) ConvertToPotentialItem(ad RawAd) *models.TradedItem {
	item := &models.TradedItem{
		SourceLink: ad.Link,
		BuyPrice:   int(ad.Price),
		StatusID:   1,
	}
	if ad.ShippingCost != nil {
		item.BuyShippingCost = int(*ad.ShippingCost)
	}
	return item
}

func (s *MarketplaceService) FetchAdsFromURL(ctx context.Context, marketplace string, searchURL string) ([]RawAd, error) {
	switch marketplace {
	case "blocket":
		return s.fetchBlocketAdsFromURL(ctx, searchURL)
	case "tradera":
		return s.fetchTraderaAdsFromURL(ctx, searchURL)
	default:
		return nil, nil
	}
}

func (s *MarketplaceService) fetchTraderaAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tradera: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tradera returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tradera: %w", err)
	}

	var ads []RawAd

	doc.Find("a[data-test='item-card-link']").Each(func(i int, sel *goquery.Selection) {
		link, _ := sel.Attr("href")
		if link == "" {
			return
		}

		if !strings.HasPrefix(link, "http") {
			link = "https://www.tradera.com" + link
		}

		titleSel := sel.Find("[data-test='item-card-title']")
		title := strings.TrimSpace(titleSel.Text())

		priceSel := sel.Find("[data-test='item-card-price']")
		priceText := strings.TrimSpace(priceSel.Text())
		price := parsePrice(priceText)

		ad := RawAd{
			Link:        link,
			Title:       title,
			Price:       price,
			Marketplace: "tradera",
		}

		ads = append(ads, ad)
	})

	log.Printf("Found %d ads from Tradera", len(ads))
	return ads, nil
}

func (s *MarketplaceService) fetchBlocketAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blocket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blocket returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blocket response: %w", err)
	}

	ads, err := parseBlocketHTML(body)
	if err != nil {
		return nil, err
	}

	for i := range ads {
		adID := extractBlocketAdID(ads[i].Link)
		if adID > 0 {
			apiAd, err := s.fetchBlocketAdFromAPI(ctx, adID)
			if err == nil && apiAd != nil {
				ads[i].AdText = apiAd.AdText
			}
		}
	}

	log.Printf("Found %d ads from Blocket", len(ads))
	return ads, nil
}

func extractBlocketAdID(link string) int64 {
	re := regexp.MustCompile(`/(?:item|annons)/(\d+)`)
	matches := re.FindStringSubmatch(link)
	if len(matches) > 1 {
		id, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return id
		}
	}
	return 0
}

func parseBlocketHTML(body []byte) ([]RawAd, error) {
	// First, parse JSON-LD for basic info
	re := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*id="seoStructuredData"[^>]*>([^<]+)</script>`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no JSON-LD found")
	}

	jsonStr := string(matches[1])
	jsonStr = html.UnescapeString(jsonStr)

	var structuredData struct {
		MainEntity struct {
			ItemListElement []struct {
				Item struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					URL         string `json:"url"`
					Image       string `json:"image"`
					Offers      struct {
						Price         string `json:"price"`
						PriceCurrency string `json:"priceCurrency"`
						ItemCondition string `json:"itemCondition"`
					} `json:"offers"`
				} `json:"item"`
			} `json:"itemListElement"`
		} `json:"mainEntity"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &structuredData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-LD: %w", err)
	}

	// Parse HTML to extract shipping information
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Map to store shipping cost by URL (nil = unknown, 0 = free, >0 = specified)
	shippingCosts := make(map[string]*float64)

	// Look for shipping info in the HTML using the specific Blocket structure
	doc.Find("section article, [data-test='item-card'], .item-card").Each(func(i int, s *goquery.Selection) {
		// Try to find link
		linkElem := s.Find("a[href]")
		link, _ := linkElem.Attr("href")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = "https://www.blocket.se" + link
		}

		// Look for shipping text in the specific location (2nd div inside section)
		shippingText := ""

		// Look for specific Blocket shipping element pattern
		s.Find("div > div:nth-child(2) p, .shipping-info, [data-test='shipping-badge']").Each(func(j int, elem *goquery.Selection) {
			if shippingText == "" {
				text := strings.ToLower(strings.TrimSpace(elem.Text()))
				if strings.Contains(text, "frakt") || strings.Contains(text, "skickas") || strings.Contains(text, "kr") {
					shippingText = text
				}
			}
		})

		// Fallback: Look in all paragraphs and spans for shipping-related text
		if shippingText == "" {
			s.Find("p, span").Each(func(j int, elem *goquery.Selection) {
				if shippingText == "" {
					text := strings.ToLower(strings.TrimSpace(elem.Text()))
					if strings.Contains(text, "frakt") || strings.Contains(text, "skickas") {
						shippingText = text
					}
				}
			})
		}

		if link != "" && shippingText != "" {
			shippingCost := extractShippingCost(shippingText)
			shippingCosts[link] = shippingCost
		}
	})

	var ads []RawAd
	for _, item := range structuredData.MainEntity.ItemListElement {
		if item.Item.URL == "" {
			continue
		}

		price, _ := strconv.ParseFloat(item.Item.Offers.Price, 64)

		// Get shipping cost from HTML parsing if available
		var shippingCostPtr *float64
		if cost, ok := shippingCosts[item.Item.URL]; ok {
			shippingCostPtr = cost
		}

		ad := RawAd{
			Link:         item.Item.URL,
			Title:        item.Item.Name,
			Price:        price,
			AdText:       "",
			Marketplace:  "blocket",
			ShippingCost: shippingCostPtr,
		}

		if ad.Title != "" && ad.Price > 0 {
			ads = append(ads, ad)
		}
	}

	return ads, nil
}

func extractShippingCost(text string) *float64 {
	textLower := strings.ToLower(text)

	// Check for free shipping indicators - return 0 (free shipping)
	if strings.Contains(textLower, "gratis frakt") ||
		strings.Contains(textLower, "fri frakt") ||
		strings.Contains(textLower, "frakt ingår") {
		free := 0.0
		return &free
	}

	// Look for patterns like "frakt 63 kr", "+ 50 kr frakt", "frakt: 75kr"
	// Also handle "frakt från X kr" - use that price even if it's a minimum
	patterns := []string{
		`frakt[:\s]+(\d+)`,
		`frakt[:\s]+(\d+)\s*kr`,
		`(\d+)\s*kr[:\s]+frakt`,
		`frakt[:\s]+(\d+):-`,
		`\+(\d+)\s*kr[:\s]+frakt`,
		`frakt[:\s]+(\d+)[\s]*kr`,
		`frakt[:\s]+från[:\s]+(\d+)`,
		`från[:\s]+(\d+)\s*kr`,
		`frakt[:\s]+fr\.?[:\s]*(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(textLower)
		if len(matches) > 1 {
			cost, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && cost > 0 && cost < 10000 {
				// Sanity check: shipping cost should be less than 10000 kr
				return &cost
			}
		}
	}

	// If shipping is mentioned but no price found (e.g., "kan skickas"), return nil (unknown)
	if strings.Contains(textLower, "frakt") || strings.Contains(textLower, "skickas") {
		return nil
	}

	return nil
}

func parsePrice(priceStr string) float64 {
	if priceStr == "" {
		return 0
	}

	re := regexp.MustCompile(`[\d\s]+`)
	cleaned := re.FindString(priceStr)
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}

	return price
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

type BlocketAPIResponse struct {
	LoaderData struct {
		ItemRecommerce struct {
			ItemData struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Price       int    `json:"price"`
				Images      []struct {
					URI         string `json:"uri"`
					Width       int    `json:"width"`
					Height      int    `json:"height"`
					Description string `json:"description"`
				} `json:"images"`
				Extras []struct {
					ID      string `json:"id"`
					Value   string `json:"value"`
					ValueID int64  `json:"valueId"`
				} `json:"extras"`
			} `json:"itemData"`
			Meta struct {
				AdID int64 `json:"adId"`
			} `json:"meta"`
			TransactableData struct {
				EligibleForShipping bool `json:"eligibleForShipping"`
				SellerPaysShipping  bool `json:"sellerPaysShipping"`
				BuyNow              bool `json:"buyNow"`
			} `json:"transactableData"`
		} `json:"item-recommerce"`
	} `json:"loaderData"`
}

type BlocketAdDetails struct {
	RawAd
	ConditionID         *int64
	EligibleForShipping *bool
	SellerPaysShipping  *bool
	BuyNow              *bool
	Images              []string
}

func (s *MarketplaceService) fetchBlocketAdFromAPI(ctx context.Context, adID int64) (*BlocketAdDetails, error) {
	url := fmt.Sprintf("https://blocket-api.se/v1/ad/recommerce?id=%d", adID)

	if err := s.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Blocket API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Blocket API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var apiResp BlocketAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if apiResp.LoaderData.ItemRecommerce.ItemData.Title == "" {
		return nil, nil
	}

	var conditionID *int64
	for _, extra := range apiResp.LoaderData.ItemRecommerce.ItemData.Extras {
		if extra.ID == "condition" {
			conditionID = &extra.ValueID
			break
		}
	}

	eligible := apiResp.LoaderData.ItemRecommerce.TransactableData.EligibleForShipping
	sellerPays := apiResp.LoaderData.ItemRecommerce.TransactableData.SellerPaysShipping
	buyNow := apiResp.LoaderData.ItemRecommerce.TransactableData.BuyNow

	images := make([]string, 0, len(apiResp.LoaderData.ItemRecommerce.ItemData.Images))
	for _, img := range apiResp.LoaderData.ItemRecommerce.ItemData.Images {
		images = append(images, img.URI)
	}

	return &BlocketAdDetails{
		RawAd: RawAd{
			Title:       apiResp.LoaderData.ItemRecommerce.ItemData.Title,
			AdText:      apiResp.LoaderData.ItemRecommerce.ItemData.Description,
			Price:       float64(apiResp.LoaderData.ItemRecommerce.ItemData.Price),
			Marketplace: "blocket",
		},
		ConditionID:         conditionID,
		EligibleForShipping: &eligible,
		SellerPaysShipping:  &sellerPays,
		BuyNow:              &buyNow,
		Images:              images,
	}, nil
}

const maxRequestsPerSecond = 5
const minInterval = time.Second / maxRequestsPerSecond

func (s *MarketplaceService) waitForRateLimit(ctx context.Context) error {
	elapsed := time.Since(s.lastReqTime)
	if elapsed < minInterval {
		waitTime := minInterval - elapsed
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	s.lastReqTime = time.Now()
	return nil
}
