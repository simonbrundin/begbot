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
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	ValuationTypeBlocket     = "Blocket"
	ValuationTypeManual      = "Manuell"
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
	Category    string                 `json:"category,omitempty"`
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

	svc.loadValuationTypesFromDB()

	svc.compiler = NewValuationCompiler(cfg, llmSvc)
	svc.RegisterMethod(&DatabaseValuationMethod{svc: svc})
	svc.RegisterMethod(&LLMNewPriceMethod{svc: svc})
	svc.RegisterMethod(&TraderaValuationMethod{svc: svc})
	svc.RegisterMethod(&BlocketValuationMethod{svc: svc})
	svc.RegisterMethod(&SoldAdsValuationMethod{svc: svc})

	return svc
}

var valuationTypeEnabled = make(map[string]bool)
var valuationTypeCacheTime time.Time
var valuationTypeCacheTTL = 5 * time.Minute

func (s *ValuationService) loadValuationTypesFromDB() {
	if s.database == nil {
		return
	}
	ctx := context.Background()
	types, err := s.database.GetValuationTypes(ctx)
	if err != nil {
		log.Printf("Failed to load valuation types: %v", err)
		return
	}
	for _, t := range types {
		valuationTypeEnabled[t.Name] = t.Enabled
	}
	valuationTypeCacheTime = time.Now()
	log.Printf("Loaded valuation types from database: %v", valuationTypeEnabled)
}

func (s *ValuationService) IsValuationTypeEnabled(typeName string) bool {
	if time.Since(valuationTypeCacheTime) > valuationTypeCacheTTL {
		s.loadValuationTypesFromDB()
	}
	return valuationTypeEnabled[typeName]
}

func SetValuationTypeEnabled(typeName string, enabled bool) {
	valuationTypeEnabled[typeName] = enabled
}

func (s *ValuationService) RegisterMethod(m ValuationMethod) {
	s.methods = append(s.methods, m)
	sort.Slice(s.methods, func(i, j int) bool {
		return s.methods[i].Priority() < s.methods[j].Priority()
	})
}

func (s *ValuationService) CollectAll(ctx context.Context, productID string, productInfo ProductInfo) ([]ValuationInput, error) {
	productInfo.ProductID = productID
	inputs, _ := s.CollectAllWithErrors(ctx, productInfo)
	return inputs, nil
}

type CollectResult struct {
	Input *ValuationInput
	Error string
}

func (s *ValuationService) CollectAllWithErrors(ctx context.Context, productInfo ProductInfo) ([]ValuationInput, []CollectResult) {
	var inputs []ValuationInput
	var results []CollectResult

	for _, method := range s.methods {
		select {
		case <-ctx.Done():
			return inputs, results
		default:
		}

		input, err := method.Valuate(ctx, productInfo)
		if err != nil {
			log.Printf("Valuation method %s failed: %v", method.Name(), err)
			results = append(results, CollectResult{Input: &ValuationInput{Type: method.Name()}, Error: err.Error()})
			continue
		}

		if input != nil {
			inputs = append(inputs, *input)
			results = append(results, CollectResult{Input: input})
		}
	}

	return inputs, results
}

func (s *ValuationService) Compile(ctx context.Context, inputs []ValuationInput) (*ValuationOutput, error) {
	return s.compiler.Compile(ctx, inputs)
}

func (s *ValuationService) SaveValuations(ctx context.Context, productID string, inputs []ValuationInput) error {
	var firstErr error

	pid := int64(0)
	if len(productID) > 0 {
		parsed, err := strconv.ParseInt(productID, 10, 64)
		if err == nil {
			pid = parsed
		}
	}

	anyValuationSaved := false
	for _, input := range inputs {
		// Skip if this valuation type is not enabled
		if !valuationTypeEnabled[input.Type] {
			continue
		}

		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			metadataJSON = []byte("{}")
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
		})
		if err != nil {
			log.Printf("Failed to save valuation for method %s: %v", input.Type, err)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			anyValuationSaved = true
		}
	}

	if pid > 0 && (anyValuationSaved || len(inputs) == 0) {
		shouldEnable, checkErr := s.database.ShouldAutoEnableProduct(ctx, pid)
		if checkErr != nil {
			log.Printf("Failed to check auto-enable for product %d: %v", pid, checkErr)
		} else if shouldEnable {
			if err := s.database.SetProductEnabled(ctx, pid, true); err != nil {
				log.Printf("Failed to auto-enable product %d: %v", pid, err)
			} else {
				log.Printf("Auto-enabled product: ID=%d (passed trading rules)", pid)
			}
		}
	}

	return firstErr
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
	case ValuationTypeManual:
		return 5
	case ValuationTypeBlocket:
		return 6
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
	if m.svc == nil {
		return nil, nil
	}

	if !m.svc.IsValuationTypeEnabled("Tradera") {
		return nil, nil
	}

	// We'll support two queries when both manufacturer and model are present:
	//  - brand+model
	//  - model only
	// We will compute both prices (using cached results when available), log both via metadata
	// and return a single ValuationInput where Value is the integer average of the two.

	basePageURL := "https://www.tradera.com/valuation"
	timeout := 30 * time.Second

	if m.svc.cfg != nil {
		if m.svc.cfg.Scraping.Tradera.BaseURL != "" {
			basePageURL = m.svc.cfg.Scraping.Tradera.BaseURL
		}
		if m.svc.cfg.Scraping.Tradera.Timeout > 0 {
			timeout = m.svc.cfg.Scraping.Tradera.Timeout
		}
	}
	client := &http.Client{Timeout: timeout}

	// Helper to get a valuation response for a single query (with category refinement)
	getResultForQuery := func(ctx context.Context, client *http.Client, apiBaseURL string, cookies []*http.Cookie, q string) (*traderaValuationResponse, int, string, error) {
		first, err := traderaValuationSearch(ctx, client, apiBaseURL, q, 0, cookies)
		if err != nil {
			return nil, 0, "", err
		}
		result := first
		// Diagnostic log: record what the first (unfiltered) API response contained
		log.Printf("Tradera first-pass: query=%q count=%d median=%.2f average=%.2f lowest=%.2f highest=%.2f", q, first.Count, first.MedianPrice, first.AveragePrice, first.LowestPrice, first.HighestPrice)
		bestCategoryID := 0
		bestCategoryName := ""
		if len(first.CategoryHits) > 0 {
			bestHit := first.CategoryHits[0]
			for _, hit := range first.CategoryHits {
				if hit.Count > bestHit.Count {
					bestHit = hit
				}
			}
			if len(bestHit.Category) > 0 {
				deepest := bestHit.Category[len(bestHit.Category)-1]
				bestCategoryID = deepest.ID
				bestCategoryName = deepest.Name
				filtered, err := traderaValuationSearch(ctx, client, apiBaseURL, q, bestCategoryID, cookies)
				if err == nil && filtered.AveragePrice > 0 {
					result = filtered
					log.Printf("Tradera filtered: query=%q category=%d name=%q count=%d median=%.2f average=%.2f", q, bestCategoryID, bestCategoryName, filtered.Count, filtered.MedianPrice, filtered.AveragePrice)
				}
			}
		}
		return result, bestCategoryID, bestCategoryName, nil
	}

	// Build list of queries to run
	queries := []string{}
	if productInfo.Manufacturer != "" && productInfo.Model != "" {
		queries = append(queries, productInfo.Manufacturer+" "+productInfo.Model)
		queries = append(queries, productInfo.Model)
	} else if productInfo.AdText != "" {
		queries = append(queries, productInfo.AdText)
	}

	if len(queries) == 0 {
		return nil, nil
	}

	// Prepare results storage
	prices := make(map[string]int)
	counts := make(map[string]int)
	medians := make(map[string]int)
	averages := make(map[string]int)
	bestCats := make(map[string]map[string]interface{})
	confidences := make(map[string]float64)

	cacheTTL := 5 * time.Minute

	// Fetch page/cookies only if we need to call Tradera API for any non-cached query
	needFetch := false
	for _, q := range queries {
		cacheKey := "tradera-valuation:" + q
		traderaCache.mu.RLock()
		if e, ok := traderaCache.m[cacheKey]; ok {
			if time.Since(e.Collected) < cacheTTL {
				cached := e.Val
				prices[q] = cached.Value

				// Try to extract count/median/average from cached metadata
				if cached.Metadata != nil {
					if v, ok := cached.Metadata["count"]; ok {
						switch t := v.(type) {
						case int:
							counts[q] = t
						case float64:
							counts[q] = int(t)
						case int64:
							counts[q] = int(t)
						default:
							counts[q] = 1
						}
					} else {
						counts[q] = 1
					}

					if v, ok := cached.Metadata["median_price"]; ok {
						switch t := v.(type) {
						case int:
							medians[q] = t
						case float64:
							medians[q] = int(t)
						default:
							medians[q] = cached.Value
						}
					} else {
						medians[q] = cached.Value
					}

					if v, ok := cached.Metadata["average_price"]; ok {
						switch t := v.(type) {
						case int:
							averages[q] = t
						case float64:
							averages[q] = int(t)
						default:
							averages[q] = cached.Value
						}
					} else {
						averages[q] = cached.Value
					}

					// Confidence stored on the cached ValuationInput
					confidences[q] = cached.Confidence
				} else {
					counts[q] = 1
					medians[q] = cached.Value
					averages[q] = cached.Value
					confidences[q] = cached.Confidence
				}

				traderaCache.mu.RUnlock()
				continue
			}
		}
		traderaCache.mu.RUnlock()
		needFetch = true
	}

	var pageCookies []*http.Cookie
	var apiBaseURL string
	if needFetch {
		pageReq, err := http.NewRequestWithContext(ctx, "GET", basePageURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to build tradera page request: %w", err)
		}
		pageReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

		pageResp, err := client.Do(pageReq)
		if err != nil {
			return nil, fmt.Errorf("tradera page request failed: %w", err)
		}
		io.Copy(io.Discard, pageResp.Body)
		pageCookies = pageResp.Cookies()
		pageResp.Body.Close()

		parsedBase, err := url.Parse(basePageURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base url: %w", err)
		}
		apiBaseURL = fmt.Sprintf("%s://%s/valuationsearch", parsedBase.Scheme, parsedBase.Host)
	} else {
		// still need apiBaseURL for constructing source URLs
		parsedBase, _ := url.Parse(basePageURL)
		apiBaseURL = fmt.Sprintf("%s://%s/valuationsearch", parsedBase.Scheme, parsedBase.Host)
	}

	// For each query not cached, call Tradera API
	for _, q := range queries {
		if _, ok := prices[q]; ok {
			continue
		}
		result, bestCategoryID, bestCategoryName, err := getResultForQuery(ctx, client, apiBaseURL, pageCookies, q)
		if err != nil {
			// don't fail the whole Valuate if one query fails; log and continue
			log.Printf("Tradera valuation for query %q failed: %v", q, err)
			continue
		}

		// result prices are floats; pick median then average and convert to int
		var priceFloat float64
		if result.MedianPrice > 0 {
			priceFloat = result.MedianPrice
		} else if result.AveragePrice > 0 {
			priceFloat = result.AveragePrice
		}
		price := int(math.Round(priceFloat))
		if price <= 0 {
			log.Printf("Tradera: inga priser hittades för sökning: %s - response: %+v", q, result)
			continue
		}

		confidence := 0.5
		if result.Count >= 10 {
			confidence = 0.7
		}
		if result.Count >= 50 {
			confidence = 0.85
		}

		// Build ValuationInput for caching
		valuationURL := fmt.Sprintf("https://www.tradera.com/valuation?query=%s", url.QueryEscape(q))
		if bestCategoryID > 0 {
			valuationURL += fmt.Sprintf("&categoryId=%d", bestCategoryID)
		}

		vi := ValuationInput{
			Type:       m.Name(),
			Value:      price,
			Confidence: confidence,
			SourceURL:  valuationURL,
			Metadata: map[string]interface{}{
				"query":         q,
				"category_id":   bestCategoryID,
				"category_name": bestCategoryName,
				"count":         result.Count,
				"lowest_price":  result.LowestPrice,
				"highest_price": result.HighestPrice,
				"median_price":  result.MedianPrice,
				"average_price": result.AveragePrice,
			},
			CollectedAt: time.Now(),
		}

		// Cache the individual query result
		cacheKey := "tradera-valuation:" + q
		traderaCache.mu.Lock()
		traderaCache.m[cacheKey] = cacheEntry{Val: vi, Collected: time.Now()}
		traderaCache.mu.Unlock()

		prices[q] = price
		counts[q] = result.Count
		medians[q] = int(math.Round(result.MedianPrice))
		averages[q] = int(math.Round(result.AveragePrice))
		confidences[q] = confidence
		bestCats[q] = map[string]interface{}{"category_id": bestCategoryID, "category_name": bestCategoryName}
	}

	// If we have multiple prices, compute weighted average by counts when available
	var sum int
	var n int
	var totalCount int
	var weightedSum int
	for q, p := range prices {
		sum += p
		n++
		c := counts[q]
		totalCount += c
		weightedSum += p * c
	}
	if n == 0 {
		return nil, fmt.Errorf("inga priser hittades på tradera för de givna sökningarna")
	}
	// Prefer weighted average by number of ads when we have counts; otherwise simple mean across queries
	var avg int
	if totalCount > 0 {
		avg = int(math.Round(float64(weightedSum) / float64(totalCount)))
	} else {
		avg = sum / n
	}

	// Compute combined confidence weighted by result counts when available
	var combinedConfidence float64
	if totalCount > 0 {
		var confSum float64
		for q, c := range confidences {
			confSum += c * float64(counts[q])
		}
		combinedConfidence = confSum / float64(totalCount)
	} else {
		// fallback: average of confidences
		var confSum float64
		for _, c := range confidences {
			confSum += c
		}
		if len(confidences) > 0 {
			combinedConfidence = confSum / float64(len(confidences))
		} else {
			combinedConfidence = 0.6
		}
	}

	// Build metadata to include both raw prices for logging
	metadata := map[string]interface{}{"queries": []string{}}
	qList := make([]string, 0, len(queries))
	for _, q := range queries {
		qList = append(qList, q)
	}
	metadata["queries"] = qList
	breakdown := make(map[string]interface{})
	for q, p := range prices {
		srcURL := fmt.Sprintf("https://www.tradera.com/valuation?query=%s", url.QueryEscape(q))
		if bc, ok := bestCats[q]; ok {
			if id, ok2 := bc["category_id"].(int); ok2 && id > 0 {
				srcURL += fmt.Sprintf("&categoryId=%d", id)
			}
		}

		breakdown[q] = map[string]interface{}{
			"price":         p,
			"median_price":  medians[q],
			"average_price": averages[q],
			"count":         counts[q],
			"source_url":    srcURL,
		}
		if bc, ok := bestCats[q]; ok {
			breakdown[q].(map[string]interface{})["category"] = bc
		}
	}
	metadata["breakdown"] = breakdown

	// Construct combined ValuationInput (average saved to DB)
	vi := ValuationInput{
		Type:        m.Name(),
		Value:       avg,
		Confidence:  combinedConfidence, // combined confidence from queries
		SourceURL:   "",
		Metadata:    metadata,
		CollectedAt: time.Now(),
	}

	return &vi, nil
}

func traderaValuationSearch(ctx context.Context, client *http.Client, apiBaseURL string, query string, categoryID int, cookies []*http.Cookie) (*traderaValuationResponse, error) {
	u, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid tradera api url: %w", err)
	}
	q := u.Query()
	q.Set("query", query)
	if categoryID > 0 {
		q.Set("categoryId", fmt.Sprintf("%d", categoryID))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build tradera api request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tradera api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tradera api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tradera api response: %w", err)
	}

	var result traderaValuationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tradera response: %w", err)
	}
	return &result, nil
}

type traderaValuationResponse struct {
	AveragePrice float64              `json:"averagePrice"`
	MedianPrice  float64              `json:"medianPrice"`
	LowestPrice  float64              `json:"lowestPrice"`
	HighestPrice float64              `json:"highestPrice"`
	Count        int                  `json:"count"`
	CategoryHits []traderaCategoryHit `json:"categoryHits"`
}

type traderaCategoryHit struct {
	Count    int               `json:"count"`
	Category []traderaCategory `json:"category"`
}

type traderaCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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

type BlocketValuationMethod struct {
	svc *ValuationService
}

func (m *BlocketValuationMethod) Name() string {
	return ValuationTypeBlocket
}

func (m *BlocketValuationMethod) Priority() int {
	return 2
}

type blocketCacheEntry struct {
	Val       ValuationInput
	Collected time.Time
}

var blocketCache struct {
	mu sync.RWMutex
	m  map[string]blocketCacheEntry
}

var blocketCategoryCache struct {
	mu sync.RWMutex
	m  map[string]string
}

func init() {
	blocketCache.m = make(map[string]blocketCacheEntry)
	blocketCategoryCache.m = make(map[string]string)
}

func ClearBlocketCaches() {
	blocketCache.mu.Lock()
	blocketCache.m = make(map[string]blocketCacheEntry)
	blocketCache.mu.Unlock()
	blocketCategoryCache.mu.Lock()
	blocketCategoryCache.m = make(map[string]string)
	blocketCategoryCache.mu.Unlock()
}

type blocketSearchResponse struct {
	Docs []blocketAd `json:"docs"`
}

type blocketAd struct {
	ID      interface{}  `json:"id"`
	Price   blocketPrice `json:"price"`
	Heading string       `json:"heading"`
}

func (a blocketAd) GetID() string {
	switch v := a.ID.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return ""
	}
}

type blocketPrice struct {
	Amount int `json:"amount"`
}

type blocketAdAPIResponse struct {
	LoaderData struct {
		ItemRecommerce struct {
			ItemData struct {
				Category struct {
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
			} `json:"itemData"`
		} `json:"item-recommerce"`
	} `json:"loaderData"`
}

func detectBlocketCategory(ctx context.Context, client *http.Client, baseURL string, queries []string) string {
	if len(queries) == 0 {
		return ""
	}

	cacheKey := "blocket-category:" + queries[0]
	blocketCategoryCache.mu.RLock()
	if cat, ok := blocketCategoryCache.m[cacheKey]; ok {
		blocketCategoryCache.mu.RUnlock()
		return cat
	}
	blocketCategoryCache.mu.RUnlock()

	url := fmt.Sprintf("%s/v1/search?query=%s&limit=5", baseURL, url.QueryEscape(queries[0]))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Blocket: failed to create request for category detection: %v", err)
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Blocket: failed to fetch for category detection: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Blocket: category detection API returned status %d", resp.StatusCode)
		return ""
	}

	var result blocketSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Blocket: failed to parse search response for category: %v", err)
		return ""
	}

	if len(result.Docs) == 0 {
		return ""
	}

	categoryCounts := make(map[string]int)

	for _, ad := range result.Docs {
		adID := ad.GetID()
		if adID == "" {
			continue
		}

		adIDInt, err := strconv.ParseInt(adID, 10, 64)
		if err != nil {
			continue
		}

		adURL := fmt.Sprintf("%s/v1/ad/recommerce?id=%d", baseURL, adIDInt)
		adReq, err := http.NewRequestWithContext(ctx, "GET", adURL, nil)
		if err != nil {
			continue
		}
		adReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		adReq.Header.Set("Accept", "application/json")

		adResp, err := client.Do(adReq)
		if err != nil {
			continue
		}
		if adResp.StatusCode != http.StatusOK {
			adResp.Body.Close()
			continue
		}

		var adResult blocketAdAPIResponse
		if err := json.NewDecoder(adResp.Body).Decode(&adResult); err != nil {
			adResp.Body.Close()
			continue
		}
		adResp.Body.Close()

		cat := adResult.LoaderData.ItemRecommerce.ItemData.Category
		var categoryID string
		if cat.Parent.Parent.ID > 0 {
			categoryID = fmt.Sprintf("1.%d.%d", cat.Parent.Parent.ID, cat.Parent.ID)
		} else if cat.Parent.ID > 0 {
			categoryID = fmt.Sprintf("1.%d", cat.Parent.ID)
		}

		if categoryID != "" {
			categoryCounts[categoryID]++
		}
	}

	var mostCommon string
	var maxCount int
	for cat, count := range categoryCounts {
		if count > maxCount {
			maxCount = count
			mostCommon = cat
		}
	}

	if mostCommon != "" {
		log.Printf("Blocket: detected category %s (count: %d)", mostCommon, maxCount)
		blocketCategoryCache.mu.Lock()
		blocketCategoryCache.m[cacheKey] = mostCommon
		blocketCategoryCache.mu.Unlock()
	}

	return mostCommon
}

func (m *BlocketValuationMethod) Valuate(ctx context.Context, productInfo ProductInfo) (*ValuationInput, error) {
	if m.svc == nil {
		return nil, nil
	}

	if !m.svc.IsValuationTypeEnabled("Blocket") {
		return nil, nil
	}

	baseURL := "https://blocket-api.se"
	timeout := 30 * time.Second

	if m.svc.cfg != nil {
		if m.svc.cfg.Scraping.Blocket.BaseURL != "" {
			baseURL = m.svc.cfg.Scraping.Blocket.BaseURL
		}
		if m.svc.cfg.Scraping.Blocket.Timeout > 0 {
			timeout = m.svc.cfg.Scraping.Blocket.Timeout
		}
	}
	client := &http.Client{Timeout: timeout}

	queries := []string{}
	if productInfo.Manufacturer != "" && productInfo.Model != "" {
		queries = append(queries, productInfo.Manufacturer+" "+productInfo.Model)
		queries = append(queries, productInfo.Model)
	} else if productInfo.Model != "" {
		queries = append(queries, productInfo.Model)
	} else if productInfo.AdText != "" {
		queries = append(queries, productInfo.AdText)
	}

	if len(queries) == 0 {
		return nil, nil
	}

	prices := make(map[string]int)
	counts := make(map[string]int)
	confidences := make(map[string]float64)

	cacheTTL := 5 * time.Minute

	// First: Try to get stored blocket_category from product in database
	blockletCategory := ""
	if m.svc.database != nil && productInfo.ProductID != "" {
		log.Printf("DEBUG Blocket: ProductID=%s, trying to fetch from database", productInfo.ProductID)
		productID, err := strconv.ParseInt(productInfo.ProductID, 10, 64)
		if err == nil {
			product, err := m.svc.database.GetProductByID(ctx, productID)
			if err == nil && product != nil {
				if product.BlocketCategory != nil && *product.BlocketCategory != "" {
					blockletCategory = *product.BlocketCategory
					log.Printf("DEBUG Blocket: Product %s has stored blocket_category=%s", productInfo.ProductID, blockletCategory)
				} else {
					log.Printf("DEBUG Blocket: Product %s has NO stored blocket_category (NULL or empty)", productInfo.ProductID)
				}
			} else {
				log.Printf("DEBUG Blocket: Product %s not found or error: %v", productInfo.ProductID, err)
			}
		}
	}

	// Second: Fallback to database lookup based on LLM category if not found
	if blockletCategory == "" && m.svc.database != nil && productInfo.Category != "" {
		cat, err := m.svc.database.GetBlocketCategoryByLLMCategory(ctx, productInfo.Category)
		if err != nil {
			log.Printf("Blocket: failed to lookup category: %v", err)
		} else if cat != nil {
			blockletCategory = cat.BlocketID
			log.Printf("Blocket: found category %s for LLM category %s", blockletCategory, productInfo.Category)
		}
	}

	// Third: Fallback to dynamic detection if still not found
	if blockletCategory == "" {
		log.Printf("Blocket: category not found, using dynamic detection")
		blockletCategory = detectBlocketCategory(ctx, client, baseURL, queries)
	} else if m.svc.database != nil && productInfo.Category != "" {
		log.Printf("Blocket: using category %s from database for product", blockletCategory)
	}

	for _, q := range queries {
		cacheKey := "blocket-valuation:" + q
		if blockletCategory != "" {
			cacheKey = "blocket-valuation:" + q + ":" + blockletCategory
		}

		log.Printf("DEBUG Blocket: Query=%s, category=%s, cacheKey=%s", q, blockletCategory, cacheKey)

		blocketCache.mu.RLock()
		if e, ok := blocketCache.m[cacheKey]; ok {
			if time.Since(e.Collected) < cacheTTL {
				cached := e.Val
				prices[q] = cached.Value
				if cached.Metadata != nil {
					if v, ok := cached.Metadata["count"]; ok {
						switch t := v.(type) {
						case int:
							counts[q] = t
						case float64:
							counts[q] = int(t)
						default:
							counts[q] = 1
						}
					} else {
						counts[q] = 1
					}
					confidences[q] = cached.Confidence
				} else {
					counts[q] = 1
					confidences[q] = cached.Confidence
				}
				blocketCache.mu.RUnlock()
				continue
			}
		}
		blocketCache.mu.RUnlock()

		result, err := blocketSearch(ctx, client, baseURL, q, blockletCategory)
		if err != nil {
			log.Printf("Blocket valuation for query %q failed: %v", q, err)
			continue
		}

		log.Printf("Blocket: got %d docs for query %q", len(result.Docs), q)

		if len(result.Docs) == 0 {
			log.Printf("Blocket: inga annonser hittades för sökning: %s", q)
			continue
		}

		allPrices := make([]int, 0, len(result.Docs))
		for _, ad := range result.Docs {
			if ad.Price.Amount > 0 {
				allPrices = append(allPrices, ad.Price.Amount)
			}
		}

		if len(allPrices) == 0 {
			continue
		}

		filteredPrices := filterOutliersIQR(allPrices)
		if len(filteredPrices) == 0 {
			continue
		}

		medianPrice := calculatePercentile(filteredPrices, 0.25)
		confidence := 0.5
		if len(filteredPrices) >= 10 {
			confidence = 0.7
		}
		if len(filteredPrices) >= 50 {
			confidence = 0.85
		}

		searchURL := fmt.Sprintf("https://www.blocket.se/recommerce/forsale/search?q=%s", url.QueryEscape(q))
		if blockletCategory != "" {
			searchURL += "&product_category=" + blockletCategory
		}

		vi := ValuationInput{
			Type:       m.Name(),
			Value:      medianPrice,
			Confidence: confidence,
			SourceURL:  searchURL,
			Category:   blockletCategory,
			Metadata: map[string]interface{}{
				"query":           q,
				"total_count":     len(result.Docs),
				"filtered_count":  len(filteredPrices),
				"all_prices":      allPrices,
				"filtered_prices": filteredPrices,
				"lower_bound":     calculateLowerBound(allPrices),
				"upper_bound":     calculateUpperBound(allPrices),
			},
			CollectedAt: time.Now(),
		}

		cacheKey = "blocket-valuation:" + q
		blocketCache.mu.Lock()
		blocketCache.m[cacheKey] = blocketCacheEntry{Val: vi, Collected: time.Now()}
		blocketCache.mu.Unlock()

		prices[q] = medianPrice
		counts[q] = len(filteredPrices)
		confidences[q] = confidence
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("inga priser hittades på Blocket för de givna sökningarna")
	}

	var sum int
	var totalCount int
	var weightedSum int
	for q, p := range prices {
		sum += p
		c := counts[q]
		totalCount += c
		weightedSum += p * c
	}

	var avg int
	if totalCount > 0 {
		avg = int(math.Round(float64(weightedSum) / float64(totalCount)))
	} else {
		avg = sum / len(prices)
	}

	var combinedConfidence float64
	if totalCount > 0 {
		var confSum float64
		for q, c := range confidences {
			confSum += c * float64(counts[q])
		}
		combinedConfidence = confSum / float64(totalCount)
	} else {
		var confSum float64
		for _, c := range confidences {
			confSum += c
		}
		if len(confidences) > 0 {
			combinedConfidence = confSum / float64(len(confidences))
		} else {
			combinedConfidence = 0.6
		}
	}

	metadata := map[string]interface{}{"queries": []string{}}
	qList := make([]string, 0, len(queries))
	for _, q := range queries {
		qList = append(qList, q)
	}
	metadata["queries"] = qList
	breakdown := make(map[string]interface{})
	for q, p := range prices {
		srcURL := fmt.Sprintf("https://www.blocket.se/recommerce/forsale/search?q=%s", url.QueryEscape(q))
		if blockletCategory != "" {
			srcURL += "&product_category=" + blockletCategory
		}
		breakdown[q] = map[string]interface{}{
			"price":      p,
			"count":      counts[q],
			"source_url": srcURL,
		}
	}
	metadata["breakdown"] = breakdown

	vi := ValuationInput{
		Type:        m.Name(),
		Value:       avg,
		Confidence:  combinedConfidence,
		SourceURL:   "",
		Metadata:    metadata,
		CollectedAt: time.Now(),
	}

	return &vi, nil
}

func blocketSearch(ctx context.Context, client *http.Client, baseURL string, query string, category string) (*blocketSearchResponse, error) {
	// First: try API-style baseURL/v1/search?query=... (used in tests/mocks and internal API)
	apiBase := strings.TrimSuffix(baseURL, "/") + "/v1/search?query=" + url.QueryEscape(query)
	if category != "" {
		apiBase += "&product_category=" + url.QueryEscape(category)
	}
	log.Printf("Blocket API search URL: %s", apiBase)

	req, err := http.NewRequestWithContext(ctx, "GET", apiBase, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			// Try to parse as JSON API response
			var apiResp blocketSearchResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err == nil {
				log.Printf("Blocket API: got %d docs for query %q", len(apiResp.Docs), query)
				return &apiResp, nil
			}
			// If JSON decode failed, fall through to web scraping fallback
			log.Printf("Blocket API: failed to decode JSON, falling back to web scrape: %v", err)
		} else {
			log.Printf("Blocket API returned status %d, falling back to web scrape", resp.StatusCode)
		}
	} else {
		log.Printf("Blocket API request failed: %v, falling back to web scrape", err)
	}

	// Fallback: use blocket.se web search with category parameter (supports filtering!)
	searchURL := fmt.Sprintf("https://www.blocket.se/recommerce/forsale/search?q=%s", strings.ReplaceAll(query, " ", "+"))
	if category != "" {
		searchURL += "&product_category=" + category
	}
	log.Printf("Blocket web search URL: %s", searchURL)

	req2, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req2.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req2.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req2.Header.Set("Accept-Language", "sv-SE,sv;q=0.9,en;q=0.8")

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blocket returned status %d", resp2.StatusCode)
	}

	body, err := io.ReadAll(resp2.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Debug: log first 500 chars of response
	log.Printf("Blocket response (first 500 chars): %s", string(body[:min(500, len(body))]))

	// Parse HTML to get ads
	ads, err := parseBlocketSearchHTML(body)
	if err != nil {
		log.Printf("Blocket: HTML parsing failed: %v", err)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	log.Printf("Blocket: found %d ads for query=%s with category=%s", len(ads), query, category)

	return &blocketSearchResponse{Docs: ads}, nil
}

func parseBlocketSearchHTML(body []byte) ([]blocketAd, error) {
	// Parse JSON-LD for basic info - similar to marketplace.go
	re := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*id="seoStructuredData"[^>]*>([^<]+)</script>`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no JSON-LD found")
	}

	jsonStr := string(matches[1])
	// Unescape HTML entities
	jsonStr = strings.ReplaceAll(jsonStr, `\&quot;`, `"`)
	jsonStr = strings.ReplaceAll(jsonStr, `\&amp;`, `&`)
	jsonStr = strings.ReplaceAll(jsonStr, `\&lt;`, `<`)
	jsonStr = strings.ReplaceAll(jsonStr, `\&gt;`, `>`)

	var structuredData struct {
		MainEntity struct {
			ItemListElement []struct {
				Item struct {
					Name   string `json:"name"`
					ID     string `json:"@id"`
					URL    string `json:"url"`
					Offers struct {
						Price string `json:"price"`
					} `json:"offers"`
				} `json:"item"`
			} `json:"itemListElement"`
		} `json:"mainEntity"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &structuredData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	var ads []blocketAd
	for _, elem := range structuredData.MainEntity.ItemListElement {
		item := elem.Item
		if item.Offers.Price != "" {
			price, err := strconv.ParseFloat(item.Offers.Price, 64)
			if err == nil && price > 0 {
				ad := blocketAd{
					Heading: item.Name,
					Price:   blocketPrice{Amount: int(price)},
				}
				// Extract ID from URL
				if item.URL != "" {
					reID := regexp.MustCompile(`/item/(\d+)`)
					if m := reID.FindStringSubmatch(item.URL); len(m) > 1 {
						ad.ID = m[1]
					}
				}
				ads = append(ads, ad)
			}
		}
	}

	return ads, nil
}

func calculateQuartiles(prices []int) (float64, float64, float64) {
	sorted := make([]int, len(prices))
	copy(sorted, prices)
	sort.Ints(sorted)

	n := len(sorted)
	if n == 0 {
		return 0, 0, 0
	}

	q1Index := n / 4
	q3Index := 3 * n / 4

	var q1, q3 float64
	if q1Index > 0 && q1Index < n {
		q1 = float64(sorted[q1Index-1])
	} else if q1Index < n {
		q1 = float64(sorted[q1Index])
	}

	if q3Index > 0 && q3Index < n {
		q3 = float64(sorted[q3Index-1])
	} else if q3Index < n {
		q3 = float64(sorted[q3Index])
	}

	iqr := q3 - q1
	return q1, q3, iqr
}

func calculateLowerBound(prices []int) int {
	_, _, iqr := calculateQuartiles(prices)
	q1, _, _ := calculateQuartiles(prices)
	lower := q1 - 1.5*iqr
	if lower < 0 {
		lower = 0
	}
	return int(lower)
}

func calculateUpperBound(prices []int) int {
	_, _, iqr := calculateQuartiles(prices)
	_, q3, _ := calculateQuartiles(prices)
	upper := q3 + 1.5*iqr
	return int(upper)
}

func filterOutliersIQR(prices []int) []int {
	if len(prices) == 0 {
		return nil
	}

	lower := calculateLowerBound(prices)
	upper := calculateUpperBound(prices)

	var filtered []int
	for _, p := range prices {
		if p >= lower && p <= upper {
			filtered = append(filtered, p)
		}
	}

	return filtered
}

func calculateMedian(prices []int) int {
	if len(prices) == 0 {
		return 0
	}

	sorted := make([]int, len(prices))
	copy(sorted, prices)
	sort.Ints(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func calculatePercentile(prices []int, percentile float64) int {
	if len(prices) == 0 {
		return 0
	}

	sorted := make([]int, len(prices))
	copy(sorted, prices)
	sort.Ints(sorted)

	index := float64(len(sorted)-1) * percentile
	lower := int(index)
	upper := lower + 1
	weight := index - float64(lower)

	if upper >= len(sorted) {
		return sorted[lower]
	}
	return int(float64(sorted[lower])*(1-weight) + float64(sorted[upper])*weight)
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
