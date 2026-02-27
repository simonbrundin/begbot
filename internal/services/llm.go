package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"begbot/internal/config"
)

type LLMService struct {
	cfg          *config.Config
	client       *OpenRouterClient
	defaultModel string
	models       map[string]string
}

func NewLLMService(cfg *config.Config) *LLMService {
	var apiKey, siteURL, siteName string
	var defaultModel string
	var models map[string]string

	if cfg != nil {
		apiKey = cfg.LLM.APIKey
		siteURL = cfg.LLM.SiteURL
		siteName = cfg.LLM.SiteName
		defaultModel = cfg.LLM.DefaultModel
		models = cfg.LLM.Models
	}

	client := NewOpenRouterClient(apiKey, siteURL, siteName)
	return &LLMService{
		cfg:          cfg,
		client:       client,
		defaultModel: defaultModel,
		models:       models,
	}
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
	ProductID    string
}

type ProductIntactResult struct {
	IsIntact          bool
	HasMinorScratches bool
	IssuesFound       []string
	Reasoning         string
}

func (s *LLMService) ExtractProductInfo(ctx context.Context, title, adText, link string) (*ProductInfo, error) {
	prompt := fmt.Sprintf(`Analyze this marketplace ad and extract product information. Return ONLY a JSON object with these exact fields:
{
  "manufacturer": "brand name",
  "model": "product model",
  "category": "one of: phone, tablet, watch, headphones, case, charger, accessory, computer, component, other",
  "storage": "storage capacity if applicable",
  "condition": "product condition",
  "shipping_cost": 0
}

Ad title: %s
Ad text: %s

JSON output:`, title, adText)

	log.Printf("[LLM] Processing ad: %s", link)
	log.Printf("[LLM] Ad title: %s", title)
	log.Printf("[LLM] Ad text preview: %.200s...", adText)

	model := s.client.GetModel("ExtractProductInfo", s.defaultModel, s.models)

	content, err := s.client.Chat(ctx, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	content = cleanupMarkdownJSON(content)

	var info ProductInfo
	if err := json.Unmarshal([]byte(content), &info); err != nil {
		info = ProductInfo{}
		log.Printf("[LLM] Failed to parse LLM response as JSON: %v", err)
		log.Printf("[LLM] Raw response: %s", content)
	}

	log.Printf("[LLM] Extracted - Manufacturer: %q, Model: %q, Category: %q, Storage: %q, Condition: %q",
		info.Manufacturer, info.Model, info.Category, info.Storage, info.Condition)

	info.AdText = adText
	return &info, nil
}

func (s *LLMService) CompileValuations(ctx context.Context, valuations []ValuationInput, productName string) (*ValuationOutput, error) {
	prompt := fmt.Sprintf(`Given these valuations for "%s", suggest a selling price and safety margin.

Valuations:
%s

Return ONLY a JSON object with:
- recommended_price: Suggested selling price in öre (NOT SEK - multiply SEK price by 100)
- safety_margin: Safety margin percentage (0-100)
- reasoning: Brief explanation for the recommendation

JSON output:`, productName, formatValuationsForPrompt(valuations))

	model := s.client.GetModel("CompileValuations", s.defaultModel, s.models)

	content, err := s.client.Chat(ctx, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	content = cleanupMarkdownJSON(content)

	var output ValuationOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	output.Valuations = valuations
	return &output, nil
}

func formatValuationsForPrompt(vals []ValuationInput) string {
	var result string
	for _, v := range vals {
		result += fmt.Sprintf("- %s: %d öre", v.Type, v.Value)
		if v.SoldCount > 0 {
			result += fmt.Sprintf(" (baserat på %d sålda)", v.SoldCount)
		}
		result += "\n"
	}
	return result
}

func (s *LLMService) EvaluateProductIntact(ctx context.Context, adText, title string) (*ProductIntactResult, error) {
	prompt := fmt.Sprintf(`Analysera annonsen och avgör om produkten är HEL.

EXEMPEL PÅ SVAR:
- {"is_intact": true, "has_minor_scratches": false, "issues_found": [], "reasoning": "Inga skador nämndes"}
- {"is_intact": true, "has_minor_scratches": true, "issues_found": [], "reasoning": "Nämner mindre repor"}
- {"is_intact": false, "has_minor_scratches": false, "issues_found": ["trasig skärm"], "reasoning": "Skärmen är trasig"}

REGEL: is_intact ska vara TRUE om annonsen INTE nämner några problem/skador.
is_intact ska vara FALSE endast om det STÅR att produkten har problem.

Annonsens titel: %s
Annonsens text: %s

Svara med JSON (endast JSON, inget annat):`, title, adText)

	log.Printf("[LLM] Evaluating product intactness for: %s", title)

	model := s.client.GetModel("EvaluateProductIntact", s.defaultModel, s.models)

	content, err := s.client.Chat(ctx, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	content = cleanupMarkdownJSON(content)

	log.Printf("[LLM] Raw response for product intactness: %s", content)

	var result ProductIntactResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Printf("[LLM] Failed to parse LLM response as JSON: %v", err)
		log.Printf("[LLM] Raw response: %s", content)
		return &ProductIntactResult{
			IsIntact:  true,
			Reasoning: "Kunde inte analysera produktskicket, accepterar produkten",
		}, nil
	}

	log.Printf("[LLM] Parsed result: is_intact=%v, has_minor_scratches=%v, issues=%v, reasoning=%s",
		result.IsIntact, result.HasMinorScratches, result.IssuesFound, result.Reasoning)

	log.Printf("[LLM] Product intactness: is_intact=%v, has_minor_scratches=%v, issues=%v",
		result.IsIntact, result.HasMinorScratches, result.IssuesFound)

	return &result, nil
}
