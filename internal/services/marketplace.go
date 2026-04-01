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
	cfg           *config.Config
	lastReqTime   time.Time
	traderaClient TraderaFetcher
	proxyProvider proxy.ProxyProvider
	evomiScraper  EvomiFetcher
	logger        *log.Logger
	fetchDirect   func(ctx context.Context, searchURL string) ([]RawAd, bool, error)
	fetchFromURL  func(ctx context.Context, searchURL string) ([]RawAd, error)
}

type EvomiFetcher interface {
	FetchTraderaHTML(ctx context.Context, query string) (string, error)
}

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
	ShippingCost      *float64
	ShippingInsurance *float64
}

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

	ad.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	if ad.Title == "" {
		ad.Title = strings.TrimSpace(doc.Find("[data-test='subject']").First().Text())
	}

	description := ""
	doc.Find("[data-test='body'], .body, .description, [itemprop='description']").Each(func(i int, s *goquery.Selection) {
		if description == "" {
			description = strings.TrimSpace(s.Text())
		}
	})

	if description == "" {
		doc.Find("main, article, .main-content, #main-content").Each(func(i int, s *goquery.Selection) {
			if description == "" {
				description = strings.TrimSpace(s.Text())
			}
		})
	}

	if description == "" {
		description = strings.TrimSpace(doc.Find("body").Text())
	}

	ad.AdText = description

	priceText := doc.Find("[data-test='price'], .price").First().Text()
	ad.Price = parsePrice(priceText)

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
	searchURL := fmt.Sprintf("https://www.tradera.com/search?q=%s", strings.ReplaceAll(query, " ", "+"))

	if s.traderaClient != nil {
		if s.cfg != nil && s.cfg.Scraping.Tradera.AppID != "" && s.cfg.Scraping.Tradera.AppKey != "" {
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

	ads, blocked, err := s.fetchDirect(ctx, searchURL)
	if err == nil && len(ads) > 0 && !blocked {
		s.logger.Printf("Fetched %d ads from Tradera via direct scrape", len(ads))
		return s.fetchPricesForZeroPriceItems(ctx, ads), nil
	}
	if blocked {
		s.logger.Printf("Direct scrape appears blocked or suspicious; switching to Evomi Scraper (if configured)")
	} else if err != nil {
		s.logger.Printf("Direct scrape failed: %v; attempting Evomi Scraper if available", err)
	} else {
		s.logger.Printf("Direct scrape returned no ads; falling back to other providers")
	}

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
					return s.fetchPricesForZeroPriceItems(ctx, evAds), nil
				}
			}
		}
	}

	s.logger.Printf("Falling back to proxy fetch for %s", searchURL)
	return s.fetchFromURL(ctx, searchURL)
}

const directResultThreshold = 10

func (s *MarketplaceService) fetchTraderaAdsDirect(ctx context.Context, searchURL string) ([]RawAd, bool, error) {
	client := &http.Client{
		Timeout:   30 * time.Second,
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

	priceCount := 0
	for _, item := range items {
		if item.Price > 0 {
			priceCount++
		}
	}

	if s.logger != nil {
		s.logger.Printf("[Tradera scraper] Found %d items, %d with prices from data-testid selectors", len(items), priceCount)
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
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

	if len(ads) > 0 {
		ads = s.fetchPricesForZeroPriceItems(ctx, ads)
	}

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

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	shippingCosts := make(map[string]*float64)
	insuranceCosts := make(map[string]*float64)
	shippingFoundCount := 0

	doc.Find("section article, [data-test='item-card'], .item-card").Each(func(i int, s *goquery.Selection) {
		linkElem := s.Find("a[href]")
		link, _ := linkElem.Attr("href")
		if link != "" && !strings.HasPrefix(link, "http") {
			link = "https://www.blocket.se" + link
		}

		shippingText := ""

		s.Find("div > div:nth-child(2) p, .shipping-info, [data-test='shipping-badge']").Each(func(j int, elem *goquery.Selection) {
			if shippingText == "" {
				text := strings.ToLower(strings.TrimSpace(elem.Text()))
				if strings.Contains(text, "frakt") || strings.Contains(text, "skickas") || strings.Contains(text, "kr") {
					shippingText = text
				}
			}
		})

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

		var shippingCostPtr *float64
		if cost, ok := shippingCosts[item.Item.URL]; ok {
			shippingCostPtr = cost
		}

		var insuranceCostPtr *float64
		if cost, ok := insuranceCosts[item.Item.URL]; ok {
			insuranceCostPtr = cost
		}

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
	text = strings.ReplaceAll(text, "\u00a0", " ")
	textLower := strings.ToLower(text)

	if strings.Contains(textLower, "gratis frakt") ||
		strings.Contains(textLower, "fri frakt") ||
		strings.Contains(textLower, "frakt ingår") {
		free := 0.0
		return &free
	}

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
		`frakt[:\s]+från[:\s]+(\d+)(?:\s*kr)?`,
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

	if strings.Contains(textLower, "frakt") || strings.Contains(textLower, "skickas") {
		return nil
	}

	return nil
}

func extractInsuranceCost(text string) *float64 {
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

	transport, err := s.proxyProvider.GetTransport()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy transport: %w", err)
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
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

	var shippingCost *float64
	var insuranceCost *float64
	shippingPriceText := apiResp.LoaderData.ItemRecommerce.TransactableUIData.Sections.Sidebar.OptedIn.ShippingPrice.Text

	if shippingPriceText == "" {
		log.Printf("[Blocket API] Ad %d: no shippingPrice in response", adID)
	} else {
		normalizedText := strings.ReplaceAll(shippingPriceText, "\u00a0", " ")
		shippingCost = extractShippingCost(normalizedText)
		insuranceCost = extractInsuranceCost(normalizedText)
		log.Printf("[Blocket API] Ad %d: shippingPrice=%q, normalized=%q, shippingCost=%v, insuranceCost=%v",
			adID, shippingPriceText, normalizedText, shippingCost, insuranceCost)
	}

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

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func normalizeWeight(weightStr string) float64 {
	re := regexp.MustCompile(`[\d.,]+`)
	matches := re.FindStringSubmatch(weightStr)
	if len(matches) == 0 {
		return 0
	}

	weightStr = matches[0]
	weightStr = strings.ReplaceAll(weightStr, ",", ".")

	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return 0
	}

	if strings.Contains(strings.ToLower(weightStr), "kg") {
		weight *= 1000
	}

	return weight
}

func weightToGrams(weightStr string) (int, error) {
	weight := normalizeWeight(weightStr)
	if weight == 0 {
		return 0, fmt.Errorf("could not parse weight: %s", weightStr)
	}
	return int(weight), nil
}

func extractUUIDFromSSN(ssn string) (string, error) {
	re := regexp.MustCompile(`\d{8}-\d{4}`)
	match := re.FindStringSubmatch(ssn)
	if len(match) > 0 {
		return match[0], nil
	}
	return "", fmt.Errorf("no UUID found in SSN: %s", ssn)
}

func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("%.2f kr", amount)
}

func formatTimeSince(t time.Time) string {
	elapsed := time.Since(t)
	if elapsed < time.Minute {
		return "less than a minute"
	}
	if elapsed < time.Hour {
		minutes := int(elapsed.Minutes())
		return fmt.Sprintf("%d minute%s", minutes, suffix(minutes))
	}
	if elapsed < 24*time.Hour {
		hours := int(elapsed.Hours())
		return fmt.Sprintf("%d hour%s", hours, suffix(hours))
	}
	days := int(elapsed.Hours() / 24)
	return fmt.Sprintf("%d day%s", days, suffix(days))
}

func suffix(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func reverseRunes(r []rune) []rune {
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return r
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}

func removeNonASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 128 {
			return r
		}
		return -1
	}, s)
}

func normalizeWhitespace(s string) string {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s
}

func removeNewlines(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, s)
}

func stringsEqualIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func formatPhoneNumber(phone string) string {
	phone = regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	if len(phone) == 10 {
		return fmt.Sprintf("%s-%s-%s", phone[:3], phone[3:6], phone[6:])
	}
	return phone
}

func extractNumbers(s string) string {
	return regexp.MustCompile(`\d+`).FindString(s)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func intToRoman(num int) string {
	if num <= 0 {
		return ""
	}
	values := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	result := ""
	for i := 0; i < len(values); i++ {
		for num >= values[i] {
			result += symbols[i]
			num -= values[i]
		}
	}
	return result
}

func romanToInt(s string) int {
	result := 0
	prev := 0
	for i := len(s) - 1; i >= 0; i-- {
		val := 0
		switch s[i] {
		case 'I':
			val = 1
		case 'V':
			val = 5
		case 'X':
			val = 10
		case 'L':
			val = 50
		case 'C':
			val = 100
		case 'D':
			val = 500
		case 'M':
			val = 1000
		}
		if val < prev {
			result -= val
		} else {
			result += val
		}
		prev = val
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func isPalindrome(s string) bool {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(s, "")
	if len(s) <= 1 {
		return true
	}
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		if s[i] != s[j] {
			return false
		}
	}
	return true
}

func wordCount(s string) int {
	return len(regexp.MustCompile(`\s+`).Split(strings.TrimSpace(s), -1))
}

func mostFrequentChar(s string) (rune, int) {
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}
	var mostFreq rune
	maxCount := 0
	for c, count := range freq {
		if count > maxCount {
			maxCount = count
			mostFreq = c
		}
	}
	return mostFreq, maxCount
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func kebabToCamel(kebab string) string {
	parts := strings.Split(kebab, "-")
	for i, part := range parts {
		if i == 0 {
			continue
		}
		if len(part) > 0 {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, "")
}

func camelToKebab(camel string) string {
	var result []rune
	for i, r := range camel {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = minInt(matrix[i-1][j]+1, matrix[i][j-1]+1)
			matrix[i][j] = minInt(matrix[i][j], matrix[i-1][j-1]+cost)
		}
	}
	return matrix[len(a)][len(b)]
}

func hammingDistance(a, b string) int {
	if len(a) != len(b) {
		return -1
	}
	distance := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			distance++
		}
	}
	return distance
}

func isAnagram(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	freq := make(map[rune]int)
	for _, c := range a {
		freq[c]++
	}
	for _, c := range b {
		freq[c]--
		if freq[c] < 0 {
			return false
		}
	}
	return true
}

func isPangram(s string) bool {
	alphabet := make(map[rune]bool)
	for _, c := range strings.ToLower(s) {
		if c >= 'a' && c <= 'z' {
			alphabet[c] = true
		}
	}
	return len(alphabet) == 26
}

func reverseWords(s string) string {
	words := strings.Fields(s)
	for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
		words[i], words[j] = words[j], words[i]
	}
	return strings.Join(words, " ")
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			if len(runes) > 1 {
				runes[1] = unicode.ToLower(runes[1])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

func removeDuplicates(s []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func groupByLength(words []string) map[int][]string {
	groups := make(map[int][]string)
	for _, word := range words {
		length := len(word)
		groups[length] = append(groups[length], word)
	}
	return groups
}

func extractTraderaPriceFromHTML(htmlContent string) float64 {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return 0
	}

	ogDesc, _ := doc.Find("meta[property='og:description']").Attr("content")
	if ogDesc != "" {
		price := parsePrice(ogDesc)
		if price > 0 {
			return price
		}
	}

	foundPrice := 0.0
	doc.Find("script[type='application/ld+json']").Each(func(i int, sel *goquery.Selection) {
		if foundPrice > 0 {
			return
		}
		jsonText := strings.TrimSpace(sel.Text())
		if strings.Contains(jsonText, `"price"`) && strings.Contains(jsonText, `"priceCurrency":"SEK"`) {
			foundPrice = extractPriceFromJSONLD(jsonText)
		}
	})

	return foundPrice
}

func extractPriceFromJSONLD(jsonText string) float64 {
	pricePattern := regexp.MustCompile(`"price"\s*:\s*"?([0-9]+(?:\.[0-9]+)?)"?`)
	matches := pricePattern.FindStringSubmatch(jsonText)
	if len(matches) > 1 {
		price, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return price
		}
	}
	return 0
}

func (s *MarketplaceService) fetchTraderaItemPrice(ctx context.Context, itemURL string) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", itemURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en-US;q=0.8,en;q=0.7")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("tradera item page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	htmlContent := string(body)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	ogDesc, _ := doc.Find("meta[property='og:description']").Attr("content")
	if ogDesc != "" {
		log.Printf("[DEBUG] og:description for %s: %s", itemURL, ogDesc)
	}

	return extractTraderaPriceFromHTML(htmlContent), nil
}

func (s *MarketplaceService) fetchPricesForZeroPriceItems(ctx context.Context, ads []RawAd) []RawAd {
	zeroPriceCount := 0
	for _, ad := range ads {
		if ad.Price == 0 {
			zeroPriceCount++
		}
	}

	if zeroPriceCount == 0 {
		return ads
	}

	if s.logger != nil {
		s.logger.Printf("[Tradera scraper] Fetching prices for %d items with zero price from individual item pages", zeroPriceCount)
	}

	fetchedCount := 0
	for i, ad := range ads {
		if ad.Price == 0 {
			price, err := s.fetchTraderaItemPrice(ctx, ad.Link)
			if err != nil {
				if s.logger != nil {
					s.logger.Printf("[Tradera scraper] Failed to fetch price for %s: %v", ad.Link, err)
				}
			} else if price > 0 {
				ads[i].Price = price
				fetchedCount++
				if s.logger != nil {
					s.logger.Printf("[Tradera scraper] Fetched price %d SEK for %s", int(price), ad.Link)
				}
			} else {
				if s.logger != nil {
					s.logger.Printf("[Tradera scraper] Price was 0 for %s (HTML might not contain price info)", ad.Link)
				}
			}

			time.Sleep(200 * time.Millisecond)
		}
	}

	if s.logger != nil {
		s.logger.Printf("[Tradera scraper] Fetched prices for %d/%d zero-price items", fetchedCount, zeroPriceCount)
	}

	return ads
}
