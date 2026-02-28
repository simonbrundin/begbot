package services

import (
	"begbot/internal/config"
	"begbot/internal/marketplaces"
	"begbot/internal/models"
	"begbot/internal/proxy"
	"begbot/internal/scraper"
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
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type MarketplaceService struct {
	cfg         *config.Config
	lastReqTime time.Time
	// traderaClient is an abstraction to allow testing/mocking
	traderaClient TraderaFetcher
	proxyProvider proxy.ProxyProvider
	// evomiScraper is an abstraction so tests can inject a fake scraper
	evomiScraper EvomiFetcher
	logger       *log.Logger
	// injectable function hooks for testing
	fetchDirect  func(ctx context.Context, searchURL string) ([]RawAd, bool, error)
	fetchFromURL func(ctx context.Context, searchURL string) ([]RawAd, error)
}

// EvomiFetcher abstracts a scraper capable of returning Tradera HTML.
type EvomiFetcher interface {
	FetchTraderaHTML(ctx context.Context, query string) (string, error)
}

// TraderaFetcher abstracts Tradera API client used to fetch ads.
type TraderaFetcher interface {
	FetchAds(ctx context.Context, query string) ([]marketplaces.RawAd, error)
}

func NewMarketplaceService(cfg *config.Config) *MarketplaceService {
	traderaClient := marketplaces.NewTraderaClient(&cfg.Scraping.Tradera)

	proxyProvider, err := proxy.NewProvider(proxy.ProxyConfig{
		Provider: cfg.Scraping.Proxy.Provider,
		APIKey:   cfg.Scraping.Proxy.APIKey,
		Country:  cfg.Scraping.Proxy.Country,
		Username: cfg.Scraping.Proxy.Username,
		Password: cfg.Scraping.Proxy.Password,
	})
	if err != nil {
		log.Printf("Failed to initialize proxy: %v, using no proxy", err)
		proxyProvider = &proxy.NoProxy{}
	}

	var evomiScraper *scraper.EvomiScraper
	if cfg.Scraping.Scraper.EvomiAPIKey != "" {
		evomiScraper, err = scraper.NewEvomiScraper(&cfg.Scraping.Scraper, proxyProvider)
		if err != nil {
			log.Printf("Failed to initialize Evomi scraper: %v", err)
		}
	}

	s := &MarketplaceService{
		cfg:           cfg,
		traderaClient: traderaClient,
		proxyProvider: proxyProvider,
		evomiScraper:  evomiScraper,
		logger:        log.Default(),
	}

	// default hooks point to real methods
	s.fetchDirect = s.fetchTraderaAdsDirect
	s.fetchFromURL = s.fetchTraderaAdsFromURL

	return s
}

type RawAd struct {
	Link              string
	Title             string
	Price             float64
	AdText            string
	ImageURLs         []string
	AdDate            time.Time
	Marketplace       string
	ShippingCost      *float64 // NULL if unknown, 0 if free, positive value if specified
	ShippingInsurance *float64 // NULL if unknown, positive value if specified (e.g., köpskydd)
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

	traderaAds, err := s.fetchTraderaAds(ctx, query)
	if err != nil {
		return nil, err
	}
	ads = append(ads, traderaAds...)

	blocketAds, err := s.fetchBlocketAds(ctx, query)
	if err != nil {
		return nil, err
	}
	ads = append(ads, blocketAds...)

	return ads, nil
}

func (s *MarketplaceService) fetchTraderaAds(ctx context.Context, query string) ([]RawAd, error) {
	if s.logger == nil {
		s.logger = log.Default()
	}
	// Preferred order: Tradera SOAP/API -> direct scrape (no proxy) -> Evomi Scraper -> proxy-based scrape
	searchURL := fmt.Sprintf("https://www.tradera.com/search?q=%s", strings.ReplaceAll(query, " ", "+"))

	// 1) Try Tradera SOAP API if available. In production we require config credentials,
	// but in tests a mock traderaClient may be injected without cfg — allow that.
	if s.traderaClient != nil {
		if s.cfg == nil || s.cfg.Scraping.Tradera.AppID != "" || s.cfg.Scraping.Tradera.AppKey != "" {
			s.logger.Printf("Trying Tradera API for query: %s", query)
			apiAds, apiErr := s.traderaClient.FetchAds(ctx, query)
			if apiErr != nil {
				s.logger.Printf("Tradera API failed: %v; attempting direct scrape", apiErr)
			} else if len(apiAds) > 0 {
				s.logger.Printf("Fetched %d ads from Tradera API", len(apiAds))
				return convertMarketplacesRawAds(apiAds), nil
			} else {
				s.logger.Printf("Tradera API returned no ads; attempting direct scrape")
			}
		} else {
			s.logger.Printf("Tradera client present but no app credentials in cfg; skipping API call and attempting direct scrape")
		}
	}

	// 2) Try direct (no-proxy) scrape and detect blocking
	ads, blocked, err := s.fetchDirect(ctx, searchURL)
	if err == nil && len(ads) > 0 && !blocked {
		s.logger.Printf("Fetched %d ads from Tradera via direct scrape", len(ads))
		return ads, nil
	}
	if blocked {
		s.logger.Printf("Direct scrape appears blocked or suspicious; switching to Evomi Scraper (if configured)")
	} else if err != nil {
		s.logger.Printf("Direct scrape failed: %v; attempting Evomi Scraper if available", err)
	} else {
		s.logger.Printf("Direct scrape returned no ads; falling back to other providers")
	}

	// 3) Try Evomi Scraper if available (partner)
	if s.evomiScraper != nil {
		s.logger.Printf("Trying partner scraper (Evomi) for query: %s", query)
		htmlContent, evErr := s.evomiScraper.FetchTraderaHTML(ctx, query)
		if evErr != nil {
			s.logger.Printf("Evomi scraper failed: %v", evErr)
		} else if htmlContent != "" {
			doc, docErr := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
			if docErr != nil {
				s.logger.Printf("Failed to parse Evomi HTML: %v", docErr)
			} else {
				evAds := s.ParseTraderaDoc(doc)
				if len(evAds) > 0 {
					s.logger.Printf("Fetched %d ads from Evomi Scraper API", len(evAds))
					return evAds, nil
				}
			}
		}
	}

	// 4) Last resort: use proxy provider transport (injectable for tests)
	s.logger.Printf("Falling back to proxy fetch for %s", searchURL)
	return s.fetchFromURL(ctx, searchURL)
}

const directResultThreshold = 10

// fetchTraderaAdsDirect attempts to fetch Tradera search page without using the proxy provider.
// It returns a slice of ads, a boolean indicating whether the response looked blocked/suspicious,
// and an error for network/parse problems.
func (s *MarketplaceService) fetchTraderaAdsDirect(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		// Use default transport without proxy
		Transport: &http.Transport{},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create direct request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("direct fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read direct response: %w", err)
	}

	// Detect blocking heuristics
	blocked := false
	statusCode := resp.StatusCode
	lowerBody := strings.ToLower(string(body))

	if statusCode == http.StatusForbidden || statusCode == http.StatusTooManyRequests || statusCode == http.StatusServiceUnavailable {
		blocked = true
		log.Printf("Direct fetch returned status %d, treating as blocked", statusCode)
	}

	antiBotMarkers := []string{"captcha", "recaptcha", "cloudflare", "please enable javascript", "rate limited", "rate limit"}
	for _, m := range antiBotMarkers {
		if strings.Contains(lowerBody, m) {
			blocked = true
			log.Printf("Direct fetch response contains anti-bot marker %q", m)
			break
		}
	}

	// Parse doc and count result links
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, blocked, fmt.Errorf("failed to parse direct HTML: %w", err)
	}
	found := doc.Find("a[data-link-type='next-link']").Length()
	if found < directResultThreshold {
		blocked = true
		log.Printf("Direct fetch found only %d result links (threshold %d) — treating as blocked/suspicious", found, directResultThreshold)
	}

	if blocked {
		return nil, true, nil
	}

	ads := s.ParseTraderaDoc(doc)
	return ads, false, nil
}

func (s *MarketplaceService) ParseTraderaDoc(doc *goquery.Document) []RawAd {
	type ItemData struct {
		Link   string
		Title  string
		Price  float64
		Image  string
		ItemID string
	}

	items := make(map[string]ItemData)

	doc.Find("a[data-link-type='next-link']").Each(func(i int, sel *goquery.Selection) {
		link, _ := sel.Attr("href")
		if link == "" || !strings.Contains(link, "/item/") {
			return
		}

		if !strings.HasPrefix(link, "http") {
			link = "https://www.tradera.com" + link
		}

		if _, exists := items[link]; exists {
			return
		}

		ariaDesc, _ := sel.Attr("aria-describedby")
		itemID := ""
		if ariaDesc != "" {
			parts := strings.Fields(ariaDesc)
			for _, p := range parts {
				if strings.HasSuffix(p, "-price") {
					itemID = strings.TrimSuffix(p, "-price")
					break
				}
			}
		}

		title := strings.TrimSpace(sel.Find(".item-card_title__okrrK").Text())
		if title == "" {
			title = strings.TrimSpace(sel.Find("[class*='title']").First().Text())
		}
		if title == "" {
			titleAttr, _ := sel.Attr("title")
			title = strings.TrimSpace(titleAttr)
		}

		var imageURL string
		sel.Find("img").Each(func(_ int, imgSel *goquery.Selection) {
			if src, ok := imgSel.Attr("src"); ok && src != "" && !strings.Contains(src, "data:") {
				if strings.HasPrefix(src, "//") {
					src = "https:" + src
				}
				imageURL = src
				return
			}
		})

		items[link] = ItemData{
			Link:   link,
			Title:  title,
			Image:  imageURL,
			ItemID: itemID,
		}
	})

	doc.Find("[data-testid='bin-price'], [data-testid='price']").Each(func(i int, sel *goquery.Selection) {
		priceText := strings.TrimSpace(sel.Text())
		price := parsePrice(priceText)
		if price == 0 {
			return
		}

		priceID, _ := sel.Attr("id")
		if priceID == "" {
			return
		}

		itemID := strings.TrimSuffix(priceID, "-price")

		for link, item := range items {
			if item.ItemID == itemID {
				item.Price = price
				items[link] = item
				break
			}
		}
	})

	var ads []RawAd
	for _, item := range items {
		ad := RawAd{
			Link:        item.Link,
			Title:       item.Title,
			Price:       item.Price,
			Marketplace: "tradera",
		}
		if item.Image != "" {
			ad.ImageURLs = []string{item.Image}
		}
		ads = append(ads, ad)
	}

	return ads
}

func convertMarketplacesRawAds(ads []marketplaces.RawAd) []RawAd {
	result := make([]RawAd, len(ads))
	for i, ad := range ads {
		result[i] = RawAd{
			Link:              ad.Link,
			Title:             ad.Title,
			Price:             ad.Price,
			AdText:            ad.AdText,
			ImageURLs:         ad.ImageURLs,
			AdDate:            ad.AdDate,
			Marketplace:       ad.Marketplace,
			ShippingCost:      ad.ShippingCost,
			ShippingInsurance: ad.ShippingInsurance,
		}
	}
	return result
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
	if ad.ShippingInsurance != nil {
		item.BuyShippingInsurance = int(*ad.ShippingInsurance)
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
	transport, err := s.proxyProvider.GetTransport()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy transport: %w", err)
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
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

	seenLinks := make(map[string]bool)

	doc.Find("a[data-link-type='next-link']").Each(func(i int, sel *goquery.Selection) {
		link, _ := sel.Attr("href")
		if link == "" || !strings.Contains(link, "/item/") {
			return
		}

		if seenLinks[link] {
			return
		}
		seenLinks[link] = true

		if !strings.HasPrefix(link, "http") {
			link = "https://www.tradera.com" + link
		}

		title := strings.TrimSpace(sel.Find(".item-card_title__okrrK").Text())
		if title == "" {
			title = strings.TrimSpace(sel.Find("[class*='title']").First().Text())
		}
		if title == "" {
			titleAttr, _ := sel.Attr("title")
			title = strings.TrimSpace(titleAttr)
		}

		priceText := strings.TrimSpace(sel.Find(".item-card_priceDetails__TzN1U").Text())
		if priceText == "" {
			priceText = strings.TrimSpace(sel.Find("[class*='price']").First().Text())
		}
		price := parsePrice(priceText)

		var imageURLs []string
		sel.Find("img").Each(func(_ int, imgSel *goquery.Selection) {
			if src, ok := imgSel.Attr("src"); ok && src != "" && !strings.Contains(src, "data:") {
				if strings.HasPrefix(src, "//") {
					src = "https:" + src
				}
				imageURLs = []string{src}
				return
			}
		})

		ad := RawAd{
			Link:        link,
			Title:       title,
			Price:       price,
			Marketplace: "tradera",
			ImageURLs:   imageURLs,
		}

		ads = append(ads, ad)
	})

	log.Printf("Found %d ads from Tradera", len(ads))
	return ads, nil
}

func (s *MarketplaceService) fetchBlocketAdsFromURL(ctx context.Context, searchURL string) ([]RawAd, error) {
	transport, err := s.proxyProvider.GetTransport()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy transport: %w", err)
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
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
				// Copy shipping and insurance costs from API response
				if apiAd.ShippingCost != nil {
					ads[i].ShippingCost = apiAd.ShippingCost
				}
				if apiAd.ShippingInsurance != nil {
					ads[i].ShippingInsurance = apiAd.ShippingInsurance
				}
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

	// Look for shipping info in the HTML using the specific Blocket structure
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Map to store shipping cost by URL (nil = unknown, 0 = free, >0 = specified)
	shippingCosts := make(map[string]*float64)
	insuranceCosts := make(map[string]*float64)
	shippingFoundCount := 0

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
			insuranceCost := extractInsuranceCost(shippingText)
			shippingCosts[link] = shippingCost
			insuranceCosts[link] = insuranceCost
			shippingFoundCount++
			if shippingCost != nil {
				if insuranceCost != nil {
					log.Printf("[Blocket] Found shipping for %s: text=%q, cost=%.0f kr, insurance=%.0f kr", link, shippingText, *shippingCost, *insuranceCost)
				} else {
					log.Printf("[Blocket] Found shipping for %s: text=%q, cost=%.0f kr", link, shippingText, *shippingCost)
				}
			} else {
				log.Printf("[Blocket] Found shipping text but no cost for %s: %q", link, shippingText)
			}
		}
	})

	log.Printf("[Blocket] Shipping text found for %d items", shippingFoundCount)

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

		// Get insurance cost from HTML parsing if available
		var insuranceCostPtr *float64
		if cost, ok := insuranceCosts[item.Item.URL]; ok {
			insuranceCostPtr = cost
		}

		// Extract image URLs from JSON-LD
		var imageURLs []string
		if item.Item.Image != "" {
			imageURLs = []string{item.Item.Image}
		}

		ad := RawAd{
			Link:              item.Item.URL,
			Title:             item.Item.Name,
			Price:             price,
			AdText:            "",
			Marketplace:       "blocket",
			ShippingCost:      shippingCostPtr,
			ShippingInsurance: insuranceCostPtr,
			ImageURLs:         imageURLs,
		}

		if ad.Title != "" && ad.Price > 0 {
			ads = append(ads, ad)
		}
	}

	// Log summary of shipping costs
	shippingCount := 0
	freeShippingCount := 0
	insuranceCount := 0
	for _, ad := range ads {
		if ad.ShippingCost != nil {
			shippingCount++
			if *ad.ShippingCost == 0 {
				freeShippingCount++
			}
		}
		if ad.ShippingInsurance != nil {
			insuranceCount++
		}
	}
	log.Printf("[Blocket] Total: %d ads, %d with shipping cost (%d free), %d with insurance", len(ads), shippingCount, freeShippingCount, insuranceCount)

	return ads, nil
}

func extractShippingCost(text string) *float64 {
	// Normalize text: replace non-breaking spaces with regular spaces
	text = strings.ReplaceAll(text, "\u00a0", " ")
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
		`frakt[:\s]+från[:\s]+(\d+)\s*kr`,
		`från[:\s]+(\d+)\s*kr`,
		`frakt[:\s]+fr\.?[:\s]*(\d+)`,
		`frakt fr\.?[:\s]+(\d+)`,
		`frakt fr\.?[:\s]+(\d+)\s*kr`,
		// Also match "frakt från X" without "kr"
		`frakt[:\s]+från[:\s]+(\d+)(?:\s*kr)?`,
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

func extractInsuranceCost(text string) *float64 {
	// Normalize text: replace non-breaking spaces with regular spaces
	text = strings.ReplaceAll(text, "\u00a0", " ")
	textLower := strings.ToLower(text)

	if !strings.Contains(textLower, "köpskydd") && !strings.Contains(textLower, "försäkring") {
		return nil
	}

	patterns := []string{
		`köpskydd[:\s]+(\d+)`,
		`köpskydd[:\s]+(\d+)\s*kr`,
		`köpskydd[:\s]+(\d+):-`,
		`försäkring[:\s]+(\d+)`,
		`försäkring[:\s]+(\d+)\s*kr`,
		`försäkring[:\s]+(\d+):-`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(textLower)
		if len(matches) > 1 {
			cost, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && cost > 0 && cost < 10000 {
				return &cost
			}
		}
	}

	return nil
}

func parsePrice(priceStr string) float64 {
	if priceStr == "" {
		return 0
	}

	// Normalize known non-breaking spaces first
	s := strings.ReplaceAll(priceStr, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\u202f", " ")

	// Remove all Unicode space characters
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		sb.WriteRune(r)
	}
	cleaned := sb.String()

	// Keep only digits and decimal separators
	var numBuilder strings.Builder
	for _, r := range cleaned {
		if unicode.IsDigit(r) {
			numBuilder.WriteRune(r)
		} else if r == ',' || r == '.' {
			numBuilder.WriteRune(r)
		}
		// ignore other characters (kr, +, etc.)
	}

	numStr := numBuilder.String()
	if numStr == "" {
		return 0
	}

	// Handle separators. When both ',' and '.' present, decide by last occurrence:
	// - if last '.' comes after last ',' -> '.' is decimal separator (US style) => remove commas
	// - otherwise ',' is decimal separator (EU style) => remove dots and replace comma with dot
	if strings.Contains(numStr, ",") && strings.Contains(numStr, ".") {
		lastComma := strings.LastIndex(numStr, ",")
		lastDot := strings.LastIndex(numStr, ".")
		if lastDot > lastComma {
			// dot is decimal separator, remove commas
			numStr = strings.ReplaceAll(numStr, ",", "")
		} else {
			// comma is decimal separator
			numStr = strings.ReplaceAll(numStr, ".", "")
			numStr = strings.ReplaceAll(numStr, ",", ".")
		}
	} else if strings.Contains(numStr, ",") {
		// Single comma -> treat as decimal separator
		numStr = strings.ReplaceAll(numStr, ",", ".")
	} else {
		// Only dots or no separator -> remove dots (they are thousand separators)
		numStr = strings.ReplaceAll(numStr, ".", "")
	}

	// Final parse
	price, err := strconv.ParseFloat(numStr, 64)
	if err == nil {
		return price
	}

	// Fallback: extract digits only
	re := regexp.MustCompile(`\d+`)
	m := re.FindString(numStr)
	if m == "" {
		return 0
	}
	v, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0
	}
	return v
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
				Category    struct {
					ID     int64  `json:"id"`
					Value  string `json:"value"`
					Parent struct {
						ID     int64  `json:"id"`
						Value  string `json:"value"`
						Parent struct {
							ID    int64  `json:"id"`
							Value string `json:"value"`
						} `json:"parent"`
					} `json:"parent"`
				} `json:"category"`
				Images []struct {
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
			TransactableUIData struct {
				Sections struct {
					Sidebar struct {
						OptedIn struct {
							ShippingPrice struct {
								Text string `json:"text"`
							} `json:"shippingPrice"`
						} `json:"optedIn"`
					} `json:"sidebar"`
				} `json:"sections"`
			} `json:"transactableUiData"`
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
	BlocketCategoryID   string
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

	// Extract shipping and insurance costs from shippingPrice text
	var shippingCost *float64
	var insuranceCost *float64
	shippingPriceText := apiResp.LoaderData.ItemRecommerce.TransactableUIData.Sections.Sidebar.OptedIn.ShippingPrice.Text

	if shippingPriceText == "" {
		log.Printf("[Blocket API] Ad %d: no shippingPrice in response", adID)
	} else {
		// Normalize text (replace non-breaking spaces)
		normalizedText := strings.ReplaceAll(shippingPriceText, "\u00a0", " ")
		shippingCost = extractShippingCost(normalizedText)
		insuranceCost = extractInsuranceCost(normalizedText)
		log.Printf("[Blocket API] Ad %d: shippingPrice=%q, normalized=%q, shippingCost=%v, insuranceCost=%v",
			adID, shippingPriceText, normalizedText, shippingCost, insuranceCost)
	}

	// Extract category ID in format "1.93.3217" (parent.parent.id)
	cat := apiResp.LoaderData.ItemRecommerce.ItemData.Category
	blocketCategoryID := ""
	if cat.Parent.Parent.ID > 0 {
		blocketCategoryID = fmt.Sprintf("1.%d.%d", cat.Parent.Parent.ID, cat.Parent.ID)
	} else if cat.Parent.ID > 0 {
		blocketCategoryID = fmt.Sprintf("1.%d", cat.Parent.ID)
	}

	log.Printf("[Blocket API] Ad %d: category=%s (id=%d)", adID, cat.Value, cat.ID)

	return &BlocketAdDetails{
		RawAd: RawAd{
			Title:             apiResp.LoaderData.ItemRecommerce.ItemData.Title,
			AdText:            apiResp.LoaderData.ItemRecommerce.ItemData.Description,
			Price:             float64(apiResp.LoaderData.ItemRecommerce.ItemData.Price),
			Marketplace:       "blocket",
			ShippingCost:      shippingCost,
			ShippingInsurance: insuranceCost,
		},
		ConditionID:         conditionID,
		EligibleForShipping: &eligible,
		SellerPaysShipping:  &sellerPays,
		BuyNow:              &buyNow,
		Images:              images,
		BlocketCategoryID:   blocketCategoryID,
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
