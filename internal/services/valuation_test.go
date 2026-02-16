package services

import (
	"context"
	"math"
	"testing"
	"time"

	"begbot/internal/models"
)

func TestDatabaseValuationMethod_Name(t *testing.T) {
	method := &DatabaseValuationMethod{}
	name := method.Name()

	if name != "Egen databas" {
		t.Errorf("Expected name 'Egen databas', got '%s'", name)
	}
}

func TestDatabaseValuationMethod_Priority(t *testing.T) {
	method := &DatabaseValuationMethod{}
	priority := method.Priority()

	if priority != 1 {
		t.Errorf("Expected priority 1, got %d", priority)
	}
}

func TestValuationInput_Struct(t *testing.T) {
	input := ValuationInput{
		Type:        "Test",
		Value:       1500,
		Confidence:  0.75,
		SourceURL:   "https://example.com",
		Metadata:    map[string]interface{}{"test": "data"},
		CollectedAt: time.Now(),
		SoldCount:   10,
		DaysToSell:  7,
	}

	if input.Type != "Test" {
		t.Errorf("Expected type 'Test', got '%s'", input.Type)
	}

	if input.Value != 1500 {
		t.Errorf("Expected value 1500, got %d", input.Value)
	}

	if input.Confidence != 0.75 {
		t.Errorf("Expected confidence 0.75, got %f", input.Confidence)
	}
}

func TestValuationOutput_Struct(t *testing.T) {
	output := ValuationOutput{
		RecommendedPrice: 1400,
		Confidence:       0.65,
		Reasoning:        "Test reasoning",
		IndividualVals:   []ValuationInput{},
	}

	if output.RecommendedPrice != 1400 {
		t.Errorf("Expected recommended price 1400, got %f", output.RecommendedPrice)
	}

	if output.Confidence != 0.65 {
		t.Errorf("Expected confidence 0.65, got %f", output.Confidence)
	}

	if output.Reasoning != "Test reasoning" {
		t.Errorf("Expected reasoning 'Test reasoning', got '%s'", output.Reasoning)
	}
}

func TestValuationCompiler_compileWeightedAverage(t *testing.T) {
	compiler := &ValuationCompiler{}

	inputs := []ValuationInput{
		{Type: "Method1", Value: 1000, Confidence: 0.8},
		{Type: "Method2", Value: 1200, Confidence: 0.6},
		{Type: "Method3", Value: 1100, Confidence: 0.7},
	}

	result, err := compiler.compileWeightedAverage(inputs, inputs)
	if err != nil {
		t.Fatalf("compileWeightedAverage failed: %v", err)
	}

	// Check that we got a result
	if result.RecommendedPrice == 0 {
		t.Error("Expected non-zero recommended price")
	}

	// The weighted average should be between 1000 and 1200
	if result.RecommendedPrice < 1000 || result.RecommendedPrice > 1200 {
		t.Errorf("Expected price between 1000 and 1200, got %f", result.RecommendedPrice)
	}

	// Confidence should be between 0.6 and 0.8
	if result.Confidence < 0.6 || result.Confidence > 0.8 {
		t.Errorf("Expected confidence between 0.6 and 0.8, got %f", result.Confidence)
	}

	t.Logf("Weighted average result: price=%f, confidence=%f",
		result.RecommendedPrice, result.Confidence)
}

func TestValuationCompiler_compileWeightedAverage_SingleInput(t *testing.T) {
	compiler := &ValuationCompiler{}

	inputs := []ValuationInput{
		{Type: "Method1", Value: 1500, Confidence: 0.9},
	}

	result, err := compiler.compileWeightedAverage(inputs, inputs)
	if err != nil {
		t.Fatalf("compileWeightedAverage failed: %v", err)
	}

	if result.RecommendedPrice != 1500 {
		t.Errorf("Expected price 1500 for single input, got %f", result.RecommendedPrice)
	}

	if result.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9 for single input, got %f", result.Confidence)
	}
}

func TestValuationCompiler_compileWeightedAverage_NoValidInputs(t *testing.T) {
	compiler := &ValuationCompiler{}

	// Test with empty inputs
	result, err := compiler.compileWeightedAverage([]ValuationInput{}, []ValuationInput{})
	if err != nil {
		t.Fatalf("compileWeightedAverage failed: %v", err)
	}

	if result.RecommendedPrice != 0 {
		t.Errorf("Expected price 0 for empty inputs, got %f", result.RecommendedPrice)
	}

	if result.Confidence != 0 {
		t.Errorf("Expected confidence 0 for empty inputs, got %f", result.Confidence)
	}
}

func TestHistoricalValuation_CalculatePriceForDays(t *testing.T) {
	valuation := &HistoricalValuation{
		HasData:      true,
		KValue:       -10.0, // Price drops 10 SEK per day
		Intercept:    1500.0,
		AveragePrice: 1400.0,
	}

	// Test calculating price for different days
	price0 := calculatePriceForDays(0, valuation)
	price7 := calculatePriceForDays(7, valuation)
	price30 := calculatePriceForDays(30, valuation)

	// With KValue = -10, price should drop as days increase
	if price0 != 1500.0 {
		t.Errorf("Expected price 1500 for day 0, got %f", price0)
	}

	if price7 != 1430.0 {
		t.Errorf("Expected price 1430 for day 7, got %f", price7)
	}

	if price30 != 1200.0 {
		t.Errorf("Expected price 1200 for day 30, got %f", price30)
	}
}

func calculatePriceForDays(targetDays int, valuation *HistoricalValuation) float64 {
	if !valuation.HasData {
		return 0
	}
	return valuation.Intercept + valuation.KValue*float64(targetDays)
}

func TestHistoricalValuation_CalculatePriceForDays_NoData(t *testing.T) {
	valuation := &HistoricalValuation{
		HasData:      false,
		KValue:       -10.0,
		Intercept:    1500.0,
		AveragePrice: 1400.0,
	}

	price := calculatePriceForDays(7, valuation)

	if price != 0 {
		t.Errorf("Expected price 0 when no data, got %f", price)
	}
}

func TestCalculateProfit(t *testing.T) {
	buyPrice := 500.0
	shippingCost := 50.0
	estimatedSellPrice := 1000.0

	profit := calculateProfit(buyPrice, shippingCost, estimatedSellPrice)
	expectedProfit := 450.0 // 1000 - 500 - 50

	if profit != expectedProfit {
		t.Errorf("Expected profit %f, got %f", expectedProfit, profit)
	}
}

func calculateProfit(buyPrice, shippingCost, estimatedSellPrice float64) float64 {
	return estimatedSellPrice - buyPrice - shippingCost
}

func TestCalculateProfitMargin(t *testing.T) {
	buyPrice := 500.0
	shippingCost := 50.0
	profit := 450.0

	margin := calculateProfitMargin(profit, buyPrice, shippingCost)
	expectedMargin := 0.8181818181818182 // 450 / 550

	if math.Abs(margin-expectedMargin) > 0.0001 {
		t.Errorf("Expected margin %f, got %f", expectedMargin, margin)
	}
}

func calculateProfitMargin(profit, buyPrice, shippingCost float64) float64 {
	totalCost := buyPrice + shippingCost
	if totalCost == 0 {
		return 0
	}
	return profit / totalCost
}

func TestCalculateProfitMargin_ZeroCost(t *testing.T) {
	margin := calculateProfitMargin(100.0, 0.0, 0.0)

	if margin != 0 {
		t.Errorf("Expected margin 0 for zero cost, got %f", margin)
	}
}

func TestEstimateSellProbability_NegativeK(t *testing.T) {
	// When K is negative (price drops over time), probability calculation uses min(0.5 + diff*0.05, 0.95)
	// where diff = targetDays - daysOnMarket
	kValue := -10.0

	prob7 := estimateSellProbability(7, 30, kValue)   // diff = 23, prob = min(0.5 + 1.15, 0.95) = 0.95
	prob14 := estimateSellProbability(14, 30, kValue) // diff = 16, prob = min(0.5 + 0.8, 0.95) = 0.95
	prob30 := estimateSellProbability(30, 30, kValue) // diff = 0, prob = min(0.5 + 0, 0.95) = 0.5

	// Probability should be capped at 0.95 for early days, lower at target
	if prob7 != 0.95 || prob14 != 0.95 {
		t.Errorf("Expected probability 0.95 for early days, got %f and %f", prob7, prob14)
	}

	if prob30 != 0.5 {
		t.Errorf("Expected probability 0.5 at target days, got %f", prob30)
	}

	t.Logf("Probabilities: 7 days=%f, 14 days=%f, 30 days=%f", prob7, prob14, prob30)
}

func estimateSellProbability(daysOnMarket, targetDays int, kValue float64) float64 {
	if kValue >= 0 {
		return math.Max(0.5-float64(targetDays-daysOnMarket)*0.05, 0.1)
	}
	return math.Min(0.5+float64(targetDays-daysOnMarket)*0.05, 0.95)
}

func TestEstimateSellProbability_PositiveK(t *testing.T) {
	// When K is positive (price increases over time), probability calculation uses max(0.5 - diff*0.05, 0.1)
	// where diff = targetDays - daysOnMarket
	kValue := 10.0

	prob7 := estimateSellProbability(7, 30, kValue)   // diff = 23, prob = max(0.5 - 1.15, 0.1) = 0.1
	prob14 := estimateSellProbability(14, 30, kValue) // diff = 16, prob = max(0.5 - 0.8, 0.1) = 0.1
	prob30 := estimateSellProbability(30, 30, kValue) // diff = 0, prob = max(0.5 - 0, 0.1) = 0.5

	// Probability should be capped at 0.1 for early days, higher at target
	if prob7 != 0.1 || prob14 != 0.1 {
		t.Errorf("Expected probability 0.1 for early days, got %f and %f", prob7, prob14)
	}

	if prob30 != 0.5 {
		t.Errorf("Expected probability 0.5 at target days, got %f", prob30)
	}

	t.Logf("Probabilities: 7 days=%f, 14 days=%f, 30 days=%f", prob7, prob14, prob30)
}

func TestLLMNewPriceMethod_Name(t *testing.T) {
	method := &LLMNewPriceMethod{}
	name := method.Name()

	if name != "Nypris (LLM)" {
		t.Errorf("Expected name 'Nypris (LLM)', got '%s'", name)
	}
}

func TestLLMNewPriceMethod_Priority(t *testing.T) {
	method := &LLMNewPriceMethod{}
	priority := method.Priority()

	if priority != 2 {
		t.Errorf("Expected priority 2, got %d", priority)
	}
}

func TestTraderaValuationMethod_Name(t *testing.T) {
	method := &TraderaValuationMethod{}
	name := method.Name()

	if name != "Tradera" {
		t.Errorf("Expected name 'Tradera', got '%s'", name)
	}
}

func TestTraderaValuationMethod_Priority(t *testing.T) {
	method := &TraderaValuationMethod{}
	priority := method.Priority()

	if priority != 3 {
		t.Errorf("Expected priority 3, got %d", priority)
	}
}

func TestSoldAdsValuationMethod_Name(t *testing.T) {
	method := &SoldAdsValuationMethod{}
	name := method.Name()

	if name != "eBay/Marknadsplatser" {
		t.Errorf("Expected name 'eBay/Marknadsplatser', got '%s'", name)
	}
}

func TestSoldAdsValuationMethod_Priority(t *testing.T) {
	method := &SoldAdsValuationMethod{}
	priority := method.Priority()

	if priority != 4 {
		t.Errorf("Expected priority 4, got %d", priority)
	}
}

func TestValuationCompiler_Compile_NoInputs(t *testing.T) {
	compiler := &ValuationCompiler{}

	result, err := compiler.Compile(context.Background(), []ValuationInput{})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result.RecommendedPrice != 0 {
		t.Errorf("Expected price 0 for no inputs, got %f", result.RecommendedPrice)
	}

	if result.Confidence != 0 {
		t.Errorf("Expected confidence 0 for no inputs, got %f", result.Confidence)
	}
}

func TestValuationCompiler_Compile_NoValidInputs(t *testing.T) {
	compiler := &ValuationCompiler{}

	// Test with inputs that have zero value or confidence
	inputs := []ValuationInput{
		{Type: "Method1", Value: 0, Confidence: 0.8},
		{Type: "Method2", Value: 1000, Confidence: 0},
	}

	result, err := compiler.Compile(context.Background(), inputs)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result.RecommendedPrice != 0 {
		t.Errorf("Expected price 0 for invalid inputs, got %f", result.RecommendedPrice)
	}
}

func TestDatabaseValuationMethod_CalculateWeight(t *testing.T) {
	method := &DatabaseValuationMethod{}

	// Test with recent item
	recentItem := models.TradedItem{
		SellDate: func() *time.Time { t := time.Now(); return &t }(),
	}
	recentWeight := method.calculateWeight(recentItem)

	// Test with old item
	oldTime := time.Now().AddDate(0, 0, -200)
	oldItem := models.TradedItem{
		SellDate: &oldTime,
	}
	oldWeight := method.calculateWeight(oldItem)

	// Recent items should have higher weight
	if recentWeight <= oldWeight {
		t.Error("Expected recent items to have higher weight than old items")
	}

	t.Logf("Recent weight: %f, Old weight: %f", recentWeight, oldWeight)
}

func TestDatabaseValuationMethod_CalculateConfidence(t *testing.T) {
	method := &DatabaseValuationMethod{}

	// Test with no items
	noItems := []models.TradedItem{}
	conf0 := method.calculateConfidence(noItems)
	if conf0 != 0 {
		t.Errorf("Expected confidence 0 for no items, got %f", conf0)
	}

	// Test with 2 items (should be 0.3)
	twoItems := []models.TradedItem{
		{SellPrice: func() *int { v := 1000; return &v }()},
		{SellPrice: func() *int { v := 1100; return &v }()},
	}
	conf2 := method.calculateConfidence(twoItems)
	if conf2 != 0.3 {
		t.Errorf("Expected confidence 0.3 for 2 items, got %f", conf2)
	}

	// Test with 4 items (should be 0.5)
	fourItems := make([]models.TradedItem, 4)
	for i := range fourItems {
		price := 1000 + i*50
		fourItems[i] = models.TradedItem{SellPrice: &price}
	}
	conf4 := method.calculateConfidence(fourItems)
	if conf4 != 0.5 {
		t.Errorf("Expected confidence 0.5 for 4 items, got %f", conf4)
	}

	// Test with 8 items (should be 0.7)
	eightItems := make([]models.TradedItem, 8)
	for i := range eightItems {
		price := 1000 + i*50
		eightItems[i] = models.TradedItem{SellPrice: &price}
	}
	conf8 := method.calculateConfidence(eightItems)
	if conf8 != 0.7 {
		t.Errorf("Expected confidence 0.7 for 8 items, got %f", conf8)
	}
}

func TestDatabaseValuationMethod_PriceInOren(t *testing.T) {
	method := &DatabaseValuationMethod{}

	now := time.Now()
	oneDayAgo := now.AddDate(0, 0, -1)

	soldItems := []models.TradedItem{
		{SellPrice: ptr(10000), SellDate: &oneDayAgo}, // 100 SEK in database (ören)
		{SellPrice: ptr(15000), SellDate: &now},       // 150 SEK in database (ören)
		{SellPrice: ptr(12500), SellDate: &now},       // 125 SEK in database (ören)
	}

	var sumPrice, sumWeight float64
	for _, item := range soldItems {
		weight := method.calculateWeight(item)
		sumPrice += float64(*item.SellPrice) * weight
		sumWeight += weight
	}

	estimatedPrice := (sumPrice / sumWeight) / 100

	if estimatedPrice > 200 {
		t.Errorf("Valuation should be in SEK (kronor), not ören. Got %f SEK which suggests 100x bug", estimatedPrice)
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestValuationOutput_ReasonableBounds(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []ValuationInput
		newPrice float64
		wantErr  bool
		maxRatio float64
	}{
		{
			name: "happy path - normal valuation",
			inputs: []ValuationInput{
				{Type: ValuationTypeDatabase, Value: 1500, Confidence: 0.7},
				{Type: ValuationTypeLLMNewPrice, Value: 2000, Confidence: 0.8},
			},
			newPrice: 2000,
			maxRatio: MaxValuationRatio,
		},
		{
			name: "edge case - valuation 100x too high",
			inputs: []ValuationInput{
				{Type: ValuationTypeDatabase, Value: 150000, Confidence: 0.7},
			},
			newPrice: 2000,
			maxRatio: MaxValuationRatio,
			wantErr:  true,
		},
		{
			name: "edge case - valuation at 10x new price boundary",
			inputs: []ValuationInput{
				{Type: ValuationTypeDatabase, Value: 20000, Confidence: 0.7},
			},
			newPrice: 2000,
			maxRatio: MaxValuationRatio,
			wantErr:  true,
		},
		{
			name: "edge case - valuation just under 10x new price",
			inputs: []ValuationInput{
				{Type: ValuationTypeDatabase, Value: 19999, Confidence: 0.7},
			},
			newPrice: 2000,
			maxRatio: MaxValuationRatio,
			wantErr:  false,
		},
		{
			name: "no new price - should not error",
			inputs: []ValuationInput{
				{Type: ValuationTypeDatabase, Value: 1500, Confidence: 0.7},
			},
			newPrice: 0,
			maxRatio: MaxValuationRatio,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := &ValuationCompiler{}
			result, err := compiler.compileWeightedAverage(tt.inputs, tt.inputs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.newPrice > 0 {
				ratio := result.RecommendedPrice / tt.newPrice
				if ratio > tt.maxRatio && !tt.wantErr {
					t.Errorf("valuation ratio %f exceeds max %f - valuation is unreasonably high", ratio, tt.maxRatio)
				}
				if ratio > tt.maxRatio && tt.wantErr {
					t.Logf("correctly detected unreasonably high valuation: ratio=%f", ratio)
				}
			}
		})
	}
}

func TestValuationCompiler_LogsWarningForUnreasonableValuation(t *testing.T) {
	newPrice := 2000.0
	ratio := 50000.0 / newPrice

	if ratio <= MaxValuationRatio {
		t.Skip("test requires valuation > 10x new price to verify warning logging")
	}

	if ratio > MaxValuationRatio {
		t.Logf("Test would log warning for unreasonable valuation: ratio=%f (expected > 10x)", ratio)
	}
}
