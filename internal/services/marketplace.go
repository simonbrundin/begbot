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
	cfg *config.Config
}

func NewMarketplaceService(cfg *config.Config) *MarketplaceService {
	return &MarketplaceService{cfg: cfg}
}

type RawAd struct {
	Link        string
	Title       string
	Price       float64
	AdText      string
	ImageURLs   []string
	AdDate      time.Time
	Marketplace string
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
	return &models.TradedItem{
		SourceLink: ad.Link,
		BuyPrice:   int(ad.Price),
		StatusID:   1,
	}
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

	ads, err := parseBlocketJSONLD(body)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d ads from Blocket", len(ads))
	return ads, nil
}

func parseBlocketJSONLD(body []byte) ([]RawAd, error) {
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

	var ads []RawAd
	for _, item := range structuredData.MainEntity.ItemListElement {
		if item.Item.URL == "" {
			continue
		}

		price, _ := strconv.ParseFloat(item.Item.Offers.Price, 64)

		ad := RawAd{
			Link:        item.Item.URL,
			Title:       item.Item.Name,
			Price:       price,
			AdText:      item.Item.Description,
			Marketplace: "blocket",
		}

		if ad.Title != "" && ad.Price > 0 {
			ads = append(ads, ad)
		}
	}

	return ads, nil
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
