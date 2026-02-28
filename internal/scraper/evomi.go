package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"begbot/internal/config"
	"begbot/internal/proxy"
)

type EvomiScraper struct {
	apiKey   string
	provider proxy.ProxyProvider
	client   *http.Client
}

func NewEvomiScraper(cfg *config.ScraperConfig, proxyProvider proxy.ProxyProvider) (*EvomiScraper, error) {
	if cfg.EvomiAPIKey == "" {
		return nil, fmt.Errorf("Evomi API key is required")
	}

	transport := &http.Transport{}
	if proxyProvider != nil {
		if t, err := proxyProvider.GetTransport(); err == nil && t != nil {
			transport = t
		}
	}

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	return &EvomiScraper{
		apiKey:   cfg.EvomiAPIKey,
		provider: proxyProvider,
		client:   client,
	}, nil
}

func (s *EvomiScraper) Name() string { return "evomi" }

func (s *EvomiScraper) FetchTraderaHTML(ctx context.Context, query string) (string, error) {
	searchURL := fmt.Sprintf("https://www.tradera.com/search?q=%s", url.QueryEscape(query))

	apiURL := fmt.Sprintf(
		"https://scrape.evomi.com/api/v1/scraper/realtime?api_key=%s&url=%s",
		s.apiKey, url.QueryEscape(searchURL))

	log.Printf("Fetching from Evomi Scraper API: %s", searchURL)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from Evomi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Evomi API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	var htmlContent string

	if strings.Contains(contentType, "application/json") {
		var scraperResp EvomiResponse
		if err := json.Unmarshal(body, &scraperResp); err != nil {
			log.Printf("Failed to parse JSON, treating as HTML: %v", err)
			htmlContent = string(body)
		} else if !scraperResp.Success {
			return "", fmt.Errorf("Evomi scrape failed: %s", scraperResp.Message)
		} else {
			htmlContent = scraperResp.Body
		}
	} else {
		htmlContent = string(body)
	}

	return htmlContent, nil
}

type EvomiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Body    string `json:"body"`
	Status  int    `json:"status"`
}
