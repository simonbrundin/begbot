package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

const (
	OrenToKronorFactor       = 100
	MaxValuationRatio        = 10.0
	ValuationTypeDatabase    = "Egen databas"
	ValuationTypeTradera     = "Tradera"
	ValuationTypeMarketplace = "eBay/Marknadsplatser"
	ValuationTypeLLMNewPrice = "Nypris (LLM)"
)

type ValuationMethod interface {
	Name() string
	Priority() int
	Valuate(ctx context.Context, productInfo ProductInfo) (*ValuationInput, error)
}

type ValuationInput struct {
	Type        string                 `json:"type"`
	Value       int                    `json:"value"`
	Confidence  float64                `json:"confidence"`
	SourceURL   string                 `json:"source_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CollectedAt time.Time              `json:"collected_at"`
	SoldCount   int                    `json:"sold_count,omitempty"`
	DaysToSell  int                    `json:"days_to_sell,omitempty"`
}

type ValuationOutput struct {
	RecommendedPrice float64          `json:"recommended_price"`
	Confidence       float64          `json:"confidence"`
	Reasoning        string           `json:"reasoning"`
	IndividualVals   []ValuationInput `json:"individual_vals"`
	Valuations       []ValuationInput `json:"valuations,omitempty"`
}

type ValuationService struct {
	cfg          *config.Config
	database     *db.Postgres
	llmSvc       *LLMService
	methods      []ValuationMethod
	compiler     *ValuationCompiler
	defaultModel string
	models       map[string]string
}

// Simple in-memory cache for Tradera responses
var traderaCache = struct {
	mu sync.RWMutex
	m  map[string]cacheEntry
}{m: make(map[string]cacheEntry)}

type cacheEntry struct {
	Val       ValuationInput
	Collected time.Time
}

func NewValuationService(cfg *config.Config, database *db.Postgres, llmSvc *LLMService) *ValuationService {
	var defaultModel string
	var models map[string]string

	if cfg != nil {
		defaultModel = cfg.LLM.DefaultModel
		models = cfg.LLM.Models
	}

	svc := &ValuationService{
		cfg:          cfg,
		database:     database,
		llmSvc:       llmSvc,
		methods:      make([]ValuationMethod, 0),
		defaultModel: defaultModel,
		models:       models,
	}

	svc.compiler = NewValuationCompiler(cfg, llmSvc)
	svc.RegisterMethod(&DatabaseValuationMethod{svc: svc})
	svc.RegisterMethod(&LLMNewPriceMethod{svc: svc})
	svc.RegisterMethod(&TraderaValuationMethod{svc: svc})
	svc.RegisterMethod(&SoldAdsValuationMethod{svc: svc})

	return svc
}

func (s *ValuationService) RegisterMethod(m ValuationMethod) {
	s.methods = append(s.methods, m)
	sort.Slice(s.methods, func(i, j int) bool {
		return s.methods[i].Priority() < s.methods[j].Priority()
	})
}

func (s *ValuationService) CollectAll(ctx context.Context, productID string, productInfo ProductInfo) ([]ValuationInput, error) {
	var inputs []ValuationInput

	for _, method := range s.methods {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		input, err := method.Valuate(ctx, productInfo)
		if err != nil {
			log.Printf("Valuation method %s failed: %v", method.Name(), err)
			continue
		}

		if input != nil {
			inputs = append(inputs, *input)
		}
	}

	return inputs, nil
}

func (s *ValuationService) Compile(ctx context.Context, inputs []ValuationInput) (*ValuationOutput, error) {
	return s.compiler.Compile(ctx, inputs)
}

func (s *ValuationService) SaveValuations(ctx context.Context, productID string, inputs []ValuationInput) error {
	return s.SaveValuationsWithListingID(ctx, productID, inputs, nil)
}

func (s *ValuationService) SaveValuationsWithListingID(ctx context.Context, productID string, inputs []ValuationInput, listingID *int64) error {
	for _, input := range inputs {
		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			metadataJSON = []byte("{}")
		}

		pid := int64(0)
		if len(productID) > 0 {
			parsed, err := strconv.ParseInt(productID, 10, 64)
			if err == nil {
				pid = parsed
			}
		}

		vid := s.getValuationTypeID(input.Type)
		if vid == 0 {
			log.Printf("Unknown valuation type: %s, skipping", input.Type)
			continue
		}

		err = s.database.CreateValuation(ctx, &models.Valuation{
			ProductID:       &pid,
			ValuationTypeID: &vid,
			Valuation:       int(input.Value),
			Metadata:        metadataJSON,
		}, listingID)
		if err != nil {
			log.Printf("Failed to save valuation for method %s: %v", input.Type, err)
		}
	}

	return nil
}

func (s *ValuationService) getValuationTypeID(typeName string) int16 {
	switch typeName {
	case ValuationTypeDatabase:
		return 1
	case ValuationTypeTradera:
		return 2
	case ValuationTypeMarketplace:
		return 3
	case ValuationTypeLLMNewPrice:
		return 4
	default:
		return 0
	}
}

func (s *ValuationService) GetHistoricalValuation(ctx context.Context, productID string) (*HistoricalValuation, error) {
	items, err := s.database.GetSoldTradedItems(ctx, 100)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &HistoricalValuation{
			HasData:      false,
			AveragePrice: 0,
			KValue:       0,
		}, nil
	}

	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(items))

	for _, item := range items {
		var daysOnMarket int
		if item.BuyDate != nil && item.SellDate != nil {
			daysOnMarket = int(item.SellDate.Sub(*item.BuyDate).Hours() / 24)
		} else {
			daysOnMarket = 7
		}
		x := float64(daysOnMarket)
		y := float64(*item.SellPrice) / OrenToKronorFactor

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return &HistoricalValuation{
			HasData:      false,
			AveragePrice: 0,
			KValue:       0,
		}, nil
	}

	kValue := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - kValue*sumX) / n

	return &HistoricalValuation{
		HasData:      true,
		KValue:       kValue,
		Intercept:    intercept,
		AveragePrice: sumY / n,
	}, nil
}

func (s *ValuationService) CalculatePriceForDays(targetDays int, valuation *HistoricalValuation) float64 {
	if !valuation.HasData {
		return 0
	}
	return valuation.Intercept + valuation.KValue*float64(targetDays)
}

func (s *ValuationService) CalculateProfit(buyPrice, shippingCost, estimatedSellPrice float64) float64 {
	return estimatedSellPrice - buyPrice - shippingCost
}

func (s *ValuationService) CalculateProfitMargin(profit, buyPrice, shippingCost float64) float64 {
	totalCost := buyPrice + shippingCost
	if totalCost == 0 {
		return 0
	}
	return profit / totalCost
}

func (s *ValuationService) ShouldBuy(profitMargin float64) bool {
	if s.cfg == nil {
		return profitMargin >= 0.15
	}
	return profitMargin >= s.cfg.Valuation.MinProfitMargin
}

func (s *ValuationService) EstimateSellProbability(daysOnMarket, targetDays int, kValue float64) float64 {
	if kValue >= 0 {
		return math.Max(0.5-float64(targetDays-daysOnMarket)*0.05, 0.1)
	}
	return math.Min(0.5+float64(targetDays-daysOnMarket)*0.05, 0.95)
}

type HistoricalValuation struct {
	HasData      bool
	KValue       float64
	Intercept    float64
	AveragePrice float64
}

type DatabaseValuationMethod struct {
	svc *ValuationService
}

func (m *DatabaseValuationMethod) Name() string {
	return ValuationTypeDatabase
}

func (m *DatabaseValuationMethod) Priority() int {
	return 1
}

func (m *DatabaseValuationMethod) Valuate(ctx context.Context, productInfo ProductInfo) (*ValuationInput, error) {
	soldItems, err := m.svc.database.GetSoldTradedItems(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get sold items: %w", err)
	}

	if len(soldItems) == 0 {
		return nil, nil
	}

	var sumPrice, sumWeight float64
	for _, item := range soldItems {
		if item.SellPrice == nil {
			continue
		}
		weight := m.calculateWeight(item)
		sumPrice += float64(*item.SellPrice) * weight
		sumWeight += weight
	}

	estimatedPrice := (sumPrice / sumWeight) / OrenToKronorFactor
	confidence := m.calculateConfidence(soldItems)

	return &ValuationInput{
		Type:        m.Name(),
		Value:       int(estimatedPrice),
		Confidence:  confidence,
		SourceURL:   "",
		Metadata:    map[string]interface{}{"data_points": len(soldItems)},
		CollectedAt: time.Now(),
	}, nil
}

func (m *DatabaseValuationMethod) calculateWeight(item models.TradedItem) float64 {
	var daysSinceSold float64
	if item.SellDate != nil {
		daysSinceSold = time.Since(*item.SellDate).Hours() / 24
	}
	ageWeight := math.Exp(-daysSinceSold / 90)
	return ageWeight
}

func (m *DatabaseValuationMethod) calculateConfidence(items []models.TradedItem) float64 {
	if len(items) == 0 {
		return 0
	}

	if len(items) < 3 {
		return 0.3
	}
	if len(items) < 5 {
		return 0.5
	}
	if len(items) < 10 {
		return 0.7
	}

	var sum, sqDiff float64
	for _, item := range items {
		if item.SellPrice != nil {
			sum += float64(*item.SellPrice)
		}
	}
	mean := sum / float64(len(items))
	for _, item := range items {
		if item.SellPrice != nil {
			sqDiff += math.Pow(float64(*item.SellPrice)-mean, 2)
		}
	}
	stdDev := math.Sqrt(sqDiff / float64(len(items)))
	coefVar := stdDev / mean

	if coefVar < 0.1 {
		return 0.9
	}
	if coefVar < 0.2 {
		return 0.75
	}
	if coefVar < 0.3 {
		return 0.6
	}

	return 0.4
}

type LLMNewPriceMethod struct {
	svc *ValuationService
}

func (m *LLMNewPriceMethod) Name() string {
	return ValuationTypeLLMNewPrice
}

func (m *LLMNewPriceMethod) Priority() int {
	return 2
}

func (m *LLMNewPriceMethod) Valuate(ctx context.Context, productInfo ProductInfo) (*ValuationInput, error) {
	if productInfo.AdText == "" {
		return nil, fmt.Errorf("no product info available for LLM valuation")
	}

	prompt := fmt.Sprintf(`Estimate the NEW retail price in Swedish kronor (SEK) for this product:

Product information:
- Manufacturer: %s
- Model: %s
- Category: %s
- Condition: %s
- Storage: %s

Ad description: %s

Consider:
- Current market conditions
- Brand and model reputation
- Product age and condition
- Storage capacity (if applicable)

Return ONLY a JSON object:
{"price": 1500, "confidence": 75, "reasoning": "..."}

JSON output:`, productInfo.Manufacturer, productInfo.Model, productInfo.Category, productInfo.Condition, productInfo.Storage, productInfo.AdText)

	model := m.svc.llmSvc.client.GetModel("NewPrice", m.svc.defaultModel, m.svc.models)

	content, err := m.svc.llmSvc.client.Chat(ctx, model, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM API error: %w", err)
	}

	content = cleanupMarkdownJSON(content)

	type LLMResponse struct {
		Price      int    `json:"price"`
		Confidence int    `json:"confidence"`
		Reasoning  string `json:"reasoning"`
	}

	var response LLMResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &ValuationInput{
		Type:        m.Name(),
		Value:       response.Price,
		Confidence:  float64(response.Confidence),
		SourceURL:   "",
		Metadata:    map[string]interface{}{"reasoning": response.Reasoning},
		CollectedAt: time.Now(),
	}, nil
}

type TraderaValuationMethod struct {
	svc *ValuationService
}

func (m *TraderaValuationMethod) Name() string {
	return ValuationTypeTradera
}

func (m *TraderaValuationMethod) Priority() int {
	return 3
}

func (m *TraderaValuationMethod) Valuate(ctx context.Context, productInfo ProductInfo) (*ValuationInput, error) {
	if m.svc == nil || m.svc.cfg == nil {
		return nil, nil
	}

	if !m.svc.cfg.Scraping.Tradera.Enabled {
		return nil, nil
	}

	// Build query: prefer manufacturer+model, fall back to ad text
	query := ""
	if productInfo.Manufacturer != "" && productInfo.Model != "" {
		query = productInfo.Manufacturer + " " + productInfo.Model
	} else if productInfo.AdText != "" {
		query = productInfo.AdText
	}

	if query == "" {
		return nil, nil
	}

	timeout := m.svc.cfg.Scraping.Tradera.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	apiURL := "https://api.tradera.com/v1/valuation"
	if m.svc.cfg.Scraping.Tradera.BaseURL != "" {
		apiURL = m.svc.cfg.Scraping.Tradera.BaseURL
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid tradera api url: %w", err)
	}

	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	// Simple cache key includes base url and query
	cacheKey := u.String()
	// Cache TTL
	cacheTTL := 5 * time.Minute

	// Check cache
	traderaCache.mu.RLock()
	if e, ok := traderaCache.m[cacheKey]; ok {
		if time.Since(e.Collected) < cacheTTL {
			cached := e.Val
			traderaCache.mu.RUnlock()
			return &cached, nil
		}
	}
	traderaCache.mu.RUnlock()

	// Retry loop with exponential backoff
	var body []byte
	maxAttempts := 3
	backoff := 100 * time.Millisecond
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to build tradera request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("tradera request failed: %w", err)
			if attempt < maxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, lastErr
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read tradera response: %w", err)
			if attempt < maxAttempts {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		// Handle 429 with a longer wait and retry
		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxAttempts {
			time.Sleep(500 * time.Millisecond * time.Duration(attempt))
			continue
		}

		// For other non-OK statuses, treat as error
		lastErr = fmt.Errorf("tradera api returned status %d: %s", resp.StatusCode, string(body))
		if attempt < maxAttempts {
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return nil, lastErr
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse tradera response json: %w", err)
	}

	var priceFloat float64
	var confFloat float64

	keys := []string{"price", "valuation", "value", "estimated_price", "estimatedValue", "amount"}
	for _, k := range keys {
		if v, ok := parsed[k]; ok && v != nil {
			switch x := v.(type) {
			case float64:
				priceFloat = x
			case int:
				priceFloat = float64(x)
			case int64:
				priceFloat = float64(x)
			case string:
				if p, err := strconv.ParseFloat(x, 64); err == nil {
					priceFloat = p
				}
			}
			if priceFloat > 0 {
				break
			}
		}
	}

	if priceFloat == 0 {
		for _, parent := range []string{"data", "result", "valuation"} {
			if pmap, ok := parsed[parent].(map[string]interface{}); ok {
				for _, k := range keys {
					if v, ok := pmap[k]; ok && v != nil {
						switch x := v.(type) {
						case float64:
							priceFloat = x
						case int:
							priceFloat = float64(x)
						case int64:
							priceFloat = float64(x)
						case string:
							if p, err := strconv.ParseFloat(x, 64); err == nil {
								priceFloat = p
							}
						}
						if priceFloat > 0 {
							break
						}
					}
				}
			}
			if priceFloat > 0 {
				break
			}
		}
	}

	// Confidence: top-level
	if v, ok := parsed["confidence"]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			confFloat = x
		case int:
			confFloat = float64(x)
		case string:
			if c, err := strconv.ParseFloat(x, 64); err == nil {
				confFloat = c
			}
		}
	}

	// Confidence: nested
	if confFloat == 0 {
		for _, parent := range []string{"data", "result", "valuation"} {
			if pmap, ok := parsed[parent].(map[string]interface{}); ok {
				if v, ok := pmap["confidence"]; ok && v != nil {
					switch x := v.(type) {
					case float64:
						confFloat = x
					case int:
						confFloat = float64(x)
					case string:
						if c, err := strconv.ParseFloat(x, 64); err == nil {
							confFloat = c
						}
					}
					if confFloat > 0 {
						break
					}
				}
			}
		}
	}

	if confFloat > 1 {
		confFloat = confFloat / 100.0
	}

	if priceFloat <= 0 {
		return nil, fmt.Errorf("tradera returned no usable valuation for query: %s", query)
	}

	priceInt := int(math.Round(priceFloat))

	metadata := map[string]interface{}{"raw_response": parsed}

	vi := ValuationInput{
		Type:        m.Name(),
		Value:       priceInt,
		Confidence:  confFloat,
		SourceURL:   u.String(),
		Metadata:    metadata,
		CollectedAt: time.Now(),
	}

	// Save to cache
	traderaCache.mu.Lock()
	traderaCache.m[cacheKey] = cacheEntry{Val: vi, Collected: time.Now()}
	traderaCache.mu.Unlock()

	return &vi, nil
}

type SoldAdsValuationMethod struct {
	svc *ValuationService
}

func (m *SoldAdsValuationMethod) Name() string {
	return ValuationTypeMarketplace
}

func (m *SoldAdsValuationMethod) Priority() int {
	return 4
}

func (m *SoldAdsValuationMethod) Valuate(ctx context.Context, productInfo ProductInfo) (*ValuationInput, error) {
	// Stub implementation - scrape sold listings from marketplaces
	// This would scrape sold listings from various marketplaces
	// For now, return a placeholder value with low confidence

	// TODO: Implement actual marketplace scraping
	// - Parse product info from productInfo
	// - Search for sold listings on eBay, Blocket, etc.
	// - Calculate average price from sold listings

	// For demonstration, use a simple calculation based on product info
	basePrice := 0

	if productInfo.NewPrice > 0 {
		// Use 60-80% of new price as estimate
		basePrice = int(float64(productInfo.NewPrice) * 0.7)
	} else if productInfo.Manufacturer != "" && productInfo.Model != "" {
		// Use some heuristic based on product type
		// This is very basic - real implementation would need proper scraping
		if productInfo.Category == "phone" {
			basePrice = 2000 // placeholder
		} else if productInfo.Category == "tablet" {
			basePrice = 1500 // placeholder
		} else {
			basePrice = 1000 // placeholder
		}
	}

	return &ValuationInput{
		Type:       m.Name(),
		Value:      basePrice,
		Confidence: 0.2, // Low confidence for stub
		SourceURL:  "",
		Metadata: map[string]interface{}{
			"reasoning": "Marknadsplats-integration ej implementerad - scraping ej implementerat",
			"status":    "stub",
			"category":  productInfo.Category,
		},
		CollectedAt: time.Now(),
	}, nil
}

type ValuationCompiler struct {
	cfg    *config.Config
	llmSvc *LLMService
}

func NewValuationCompiler(cfg *config.Config, llmSvc *LLMService) *ValuationCompiler {
	return &ValuationCompiler{
		cfg:    cfg,
		llmSvc: llmSvc,
	}
}

func (c *ValuationCompiler) Compile(ctx context.Context, inputs []ValuationInput) (*ValuationOutput, error) {
	if len(inputs) == 0 {
		return &ValuationOutput{
			RecommendedPrice: 0,
			Confidence:       0,
			Reasoning:        "Inga värderingsmetoder tillgängliga",
			IndividualVals:   []ValuationInput{},
		}, nil
	}

	validInputs := make([]ValuationInput, 0)
	for _, input := range inputs {
		if input.Value > 0 && input.Confidence > 0 {
			validInputs = append(validInputs, input)
		}
	}

	if len(validInputs) == 0 {
		return &ValuationOutput{
			RecommendedPrice: 0,
			Confidence:       0,
			Reasoning:        "Inga giltiga värderingar tillgängliga",
			IndividualVals:   inputs,
		}, nil
	}

	if c.llmSvc != nil && len(validInputs) >= 2 {
		return c.compileWithLLM(ctx, validInputs, inputs)
	}

	return c.compileWeightedAverage(validInputs, inputs)
}

func (c *ValuationCompiler) compileWithLLM(ctx context.Context, validInputs []ValuationInput, allInputs []ValuationInput) (*ValuationOutput, error) {
	result, err := c.llmSvc.CompileValuations(ctx, validInputs, "")
	if err != nil {
		log.Printf("LLM compilation failed, falling back to weighted average: %v", err)
		return c.compileWeightedAverage(validInputs, allInputs)
	}

	result.IndividualVals = allInputs
	return result, nil
}

func (c *ValuationCompiler) compileWeightedAverage(inputs []ValuationInput, allInputs []ValuationInput) (*ValuationOutput, error) {
	var sumPrice, sumConfidence, totalWeight float64

	newPrice := c.extractLLMNewPrice(inputs)

	for _, input := range inputs {
		weight := input.Confidence * input.Confidence
		sumPrice += float64(input.Value) * weight
		sumConfidence += input.Confidence * weight
		totalWeight += weight

		if newPrice > 0 && input.Type == ValuationTypeDatabase {
			ratio := float64(input.Value) / newPrice
			if ratio > MaxValuationRatio {
				log.Printf("WARNING: Valuation for '%s' is %.0fx above new price (value=%d, newPrice=%.0f)",
					input.Type, ratio, input.Value, newPrice)
			}
		}
	}

	if totalWeight == 0 {
		return &ValuationOutput{
			RecommendedPrice: 0,
			Confidence:       0,
			Reasoning:        "Kan inte beräkna viktat genomsnitt",
			IndividualVals:   allInputs,
		}, nil
	}

	recommendedPrice := sumPrice / totalWeight
	avgConfidence := sumConfidence / totalWeight

	return &ValuationOutput{
		RecommendedPrice: recommendedPrice,
		Confidence:       avgConfidence,
		Reasoning:        fmt.Sprintf("Viktat genomsnitt baserat på %d metoder", len(inputs)),
		IndividualVals:   allInputs,
	}, nil
}

func (c *ValuationCompiler) extractLLMNewPrice(inputs []ValuationInput) float64 {
	for _, input := range inputs {
		if input.Type == ValuationTypeLLMNewPrice && input.Value > 0 {
			return float64(input.Value)
		}
	}
	return 0
}
