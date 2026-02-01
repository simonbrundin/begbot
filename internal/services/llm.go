package services

import (
	"context"
	"encoding/json"
	"fmt"

	"begbot/internal/config"
)

type LLMService struct {
	cfg    *config.Config
	client *OpenRouterClient
}

func NewLLMService(cfg *config.Config) *LLMService {
	client := NewOpenRouterClient(cfg.LLM.APIKey, cfg.LLM.SiteURL, cfg.LLM.SiteName)
	return &LLMService{cfg: cfg, client: client}
}

type ProductInfo struct {
	Manufacturer string
	Model        string
	Category     string
	Storage      string
	Condition    string
	ShippingCost float64
	AdText       string
	NewPrice     float64
}

func (s *LLMService) ExtractProductInfo(ctx context.Context, adText, link string) (*ProductInfo, error) {
	prompt := fmt.Sprintf(`Analyze this marketplace ad and extract product information. Return ONLY a JSON object with these fields:
- manufacturer: The brand/manufacturer (e.g., "Apple", "Samsung", "Sony")
- model: The product name/model (e.g., "iPhone 16 Pro", "Galaxy S24")
- category: One of: phone, tablet, watch, headphones, case, charger, accessory, computer, component, other
- storage: Storage capacity if applicable (e.g., "256GB", "512GB")
- condition: Product condition (e.g., "nyskick", "bra skick", "godkänd")

Ad text: %s

JSON output:`, adText)

	model := s.client.GetModel("ExtractProductInfo", s.cfg.LLM.DefaultModel, s.cfg.LLM.Models)

	content, err := s.client.Chat(ctx, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	content = cleanupMarkdownJSON(content)

	var info ProductInfo
	if err := json.Unmarshal([]byte(content), &info); err != nil {
		info = ProductInfo{}
	}

	info.AdText = adText
	return &info, nil
}
