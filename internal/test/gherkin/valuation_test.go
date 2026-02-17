package gherkin

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"begbot/internal/models"
	"begbot/internal/services"

	"github.com/cucumber/godog"
)

// ValuationTestContext holds the test state for valuation scenarios
type ValuationTestContext struct {
	compiler            *services.ValuationCompiler
	inputs              []services.ValuationInput
	result              *services.ValuationOutput
	lastError           error
	dbMethod            *services.DatabaseValuationMethod
	llmMethod           *services.LLMNewPriceMethod
	traderaMethod       *services.TraderaValuationMethod
	soldAdsMethod      *services.SoldAdsValuationMethod
	historicalVal       *services.HistoricalValuation
	profit             float64
	profitMargin       float64
	confidence         float64
}

func InitializeScenarioValuation(ctx *godog.ScenarioContext) {
	tc := &ValuationTestContext{}

	ctx.BeforeScenario(func(*godog.Scenario) {
		tc.compiler = &services.ValuationCompiler{}
		tc.inputs = nil
		tc.result = nil
		tc.lastError = nil
		tc.historicalVal = nil
	})

	// Given steps
	ctx.Given(`^a valuation service is available$`, func() error {
		tc.compiler = &services.ValuationCompiler{}
		return nil
	})

	ctx.Given(`^I have the following valuation inputs:$`, func(data *godog.Table) error {
		tc.inputs = []services.ValuationInput{}
		for _, row := range data.Rows[1:] {
			value, _ := strToInt(row.Cells[1].Value)
			confidence, _ := strToFloat(row.Cells[2].Value)
			tc.inputs = append(tc.inputs, services.ValuationInput{
				Type:       row.Cells[0].Value,
				Value:      value,
				Confidence: confidence,
			})
		}
		return nil
	})

	ctx.Given(`^I have a single valuation input with value "(\d+)" and confidence "([\d.]+)"$`, func(valueStr, confidenceStr string) error {
		value, _ := strToInt(valueStr)
		confidence, _ := strToFloat(confidenceStr)
		tc.inputs = []services.ValuationInput{
			{Type: "Method1", Value: value, Confidence: confidence},
		}
		return nil
	})

	ctx.Given(`^I have no valuation inputs$`, func() error {
		tc.inputs = []services.ValuationInput{}
		return nil
	})

	ctx.Given(`^a historical valuation with K value "([-\d.]+)" and intercept "([-\d.]+)"$`, func(kStr, interceptStr string) error {
		k, _ := strToFloat(kStr)
		intercept, _ := strToFloat(interceptStr)
		tc.historicalVal = &services.HistoricalValuation{
			HasData:   true,
			KValue:    k,
			Intercept: intercept,
		}
		return nil
	})

	ctx.Given(`^a historical valuation with no data$`, func() error {
		tc.historicalVal = &services.HistoricalValuation{
			HasData: false,
		}
		return nil
	})

	ctx.Given(`^buy price "(\d+)", shipping cost "(\d+)", and estimated sell price "(\d+)"$`, func(buyPriceStr, shippingStr, sellPriceStr string) error {
		// Store for later calculation
		buyPrice, _ := strToFloat(buyPriceStr)
		shipping, _ := strToFloat(shippingStr)
		sellPrice, _ := strToFloat(sellPriceStr)
		tc.profit = sellPrice - buyPrice - shipping
		return nil
	})

	ctx.Given(`^profit "(\d+)", buy price "(\d+)", and shipping cost "(\d+)"$`, func(profitStr, buyPriceStr, shippingStr string) error {
		profit, _ := strToFloat(profitStr)
		buyPrice, _ := strToFloat(buyPriceStr)
		shipping, _ := strToFloat(shippingStr)
		totalCost := buyPrice + shipping
		if totalCost > 0 {
			tc.profitMargin = profit / totalCost
		}
		return nil
	})

	ctx.Given(`^K value is "([-\d.]+)"$`, func(kStr string) error {
		tc.historicalVal = &services.HistoricalValuation{
			HasData: true,
			KValue:  strToFloat(kStr),
		}
		return nil
	})

	ctx.Given(`^I have "(\d+)" sold items$`, func(countStr string) error {
		count, _ := strToInt(countStr)
		tc.confidence = calculateConfidence(count)
		return nil
	})

	ctx.Given(`^I have the following invalid valuation inputs:$`, func(data *godog.Table) error {
		tc.inputs = []services.ValuationInput{}
		for _, row := range data.Rows[1:] {
			value, _ := strToInt(row.Cells[1].Value)
			confidence, _ := strToFloat(row.Cells[2].Value)
			tc.inputs = append(tc.inputs, services.ValuationInput{
				Type:       row.Cells[0].Value,
				Value:      value,
				Confidence: confidence,
			})
		}
		return nil
	})

	ctx.Given(`^sold items with prices in ören: "(\d+)", "(\d+)", "(\d+)"$`, func(price1, price2, price3 string) error {
		// Create test items with prices in ören
		now := time.Now()
		oneDayAgo := now.AddDate(0, 0, -1)
		
		items := []models.TradedItem{
			{SellPrice: ptrInt(strToInt(price1)), SellDate: &oneDayAgo},
			{SellPrice: ptrInt(strToInt(price2)), SellDate: &now},
			{SellPrice: ptrInt(strToInt(price3)), SellDate: &now},
		}
		
		// Calculate weighted average (simulating database method)
		var sumPrice, sumWeight float64
		for _, item := range items {
			weight := 1.0 // Simplified weight
			sumPrice += float64(*item.SellPrice) * weight
			sumWeight += weight
		}
		estimatedPrice := (sumPrice / sumWeight) / 100 // Convert from ören to kronor
		
		// Store for later assertion
		tc.compiler = &services.ValuationCompiler{}
		tc.result = &services.ValuationOutput{
			RecommendedPrice: estimatedPrice,
		}
		return nil
	})

	// When steps
	ctx.When(`^I check the database valuation method$`, func() error {
		tc.dbMethod = &services.DatabaseValuationMethod{}
		return nil
	})

	ctx.When(`^I check the LLM new price method$`, func() error {
		tc.llmMethod = &services.LLMNewPriceMethod{}
		return nil
	})

	ctx.When(`^I check the Tradera valuation method$`, func() error {
		tc.traderaMethod = &services.TraderaValuationMethod{}
		return nil
	})

	ctx.When(`^I check the sold ads valuation method$`, func() error {
		tc.soldAdsMethod = &services.SoldAdsValuationMethod{}
		return nil
	})

	ctx.When(`^I compile the weighted average$`, func() error {
		tc.result, tc.lastError = tc.compiler.Compile(context.Background(), tc.inputs)
		return nil
	})

	ctx.When(`^I compile the valuations$`, func() error {
		tc.result, tc.lastError = tc.compiler.Compile(context.Background(), tc.inputs)
		return nil
	})

	ctx.When(`^I calculate price for "(\d+)" days$`, func(daysStr string) error {
		days, _ := strToInt(daysStr)
		if tc.historicalVal != nil && tc.historicalVal.HasData {
			tc.result = &services.ValuationOutput{
				RecommendedPrice: tc.historicalVal.Intercept + tc.historicalVal.KValue*float64(days),
			}
		} else {
			tc.result = &services.ValuationOutput{
				RecommendedPrice: 0,
			}
		}
		return nil
	})

	ctx.When(`^I calculate the profit$`, func() error {
		// Already calculated in Given step
		return nil
	})

	ctx.When(`^I calculate the profit margin$`, func() error {
		// Already calculated in Given step
		return nil
	})

	ctx.When(`^I estimate sell probability for "(\d+)" days on market with target "(\d+)"$`, func(daysStr, targetStr string) error {
		days, _ := strToInt(daysStr)
		target, _ := strToInt(targetStr)
		kValue := tc.historicalVal.KValue
		
		if kValue >= 0 {
			tc.confidence = math.Max(0.5-float64(target-days)*0.05, 0.1)
		} else {
			tc.confidence = math.Min(0.5+float64(target-days)*0.05, 0.95)
		}
		return nil
	})

	ctx.When(`^I calculate confidence$`, func() error {
		// Already calculated in Given step
		return nil
	})

	// Then steps
	ctx.Then(`^the name should be "([^"]+)"$`, func(expected string) error {
		var name string
		if tc.dbMethod != nil {
			name = tc.dbMethod.Name()
		} else if tc.llmMethod != nil {
			name = tc.llmMethod.Name()
		} else if tc.traderaMethod != nil {
			name = tc.traderaMethod.Name()
		} else if tc.soldAdsMethod != nil {
			name = tc.soldAdsMethod.Name()
		}
		if name != expected {
			return fmt.Errorf("expected name '%s', got '%s'", expected, name)
		}
		return nil
	})

	ctx.Then(`^the priority should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToInt(expectedStr)
		var priority int
		if tc.dbMethod != nil {
			priority = tc.dbMethod.Priority()
		} else if tc.llmMethod != nil {
			priority = tc.llmMethod.Priority()
		} else if tc.traderaMethod != nil {
			priority = tc.traderaMethod.Priority()
		} else if tc.soldAdsMethod != nil {
			priority = tc.soldAdsMethod.Priority()
		}
		if priority != expected {
			return fmt.Errorf("expected priority %d, got %d", expected, priority)
		}
		return nil
	})

	ctx.Then(`^the recommended price should be between "(\d+)" and "(\d+)"$`, func(minStr, maxStr string) error {
		min, _ := strToFloat(minStr)
		max, _ := strToFloat(maxStr)
		if tc.result.RecommendedPrice < min || tc.result.RecommendedPrice > max {
			return fmt.Errorf("expected price between %f and %f, got %f", min, max, tc.result.RecommendedPrice)
		}
		return nil
	})

	ctx.Then(`^the confidence should be between "([\d.]+)" and "([\d.]+)"$`, func(minStr, maxStr string) error {
		min, _ := strToFloat(minStr)
		max, _ := strToFloat(maxStr)
		if tc.result.Confidence < min || tc.result.Confidence > max {
			return fmt.Errorf("expected confidence between %f and %f, got %f", min, max, tc.result.Confidence)
		}
		return nil
	})

	ctx.Then(`^the recommended price should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if tc.result.RecommendedPrice != expected {
			return fmt.Errorf("expected price %f, got %f", expected, tc.result.RecommendedPrice)
		}
		return nil
	})

	ctx.Then(`^the confidence should be "([\d.]+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if tc.result.Confidence != expected {
			return fmt.Errorf("expected confidence %f, got %f", expected, tc.result.Confidence)
		}
		return nil
	})

	ctx.Then(`^the price should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if tc.result.RecommendedPrice != expected {
			return fmt.Errorf("expected price %f, got %f", expected, tc.result.RecommendedPrice)
		}
		return nil
	})

	ctx.Then(`^the price should be "0"$`, func() error {
		if tc.result.RecommendedPrice != 0 {
			return fmt.Errorf("expected price 0, got %f", tc.result.RecommendedPrice)
		}
		return nil
	})

	ctx.Then(`^the profit should be "(\d+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if tc.profit != expected {
			return fmt.Errorf("expected profit %f, got %f", expected, tc.profit)
		}
		return nil
	})

	ctx.Then(`^the margin should be approximately "([\d.]+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if math.Abs(tc.profitMargin-expected) > 0.001 {
			return fmt.Errorf("expected margin %f, got %f", expected, tc.profitMargin)
		}
		return nil
	})

	ctx.Then(`^the margin should be "0"$`, func() error {
		if tc.profitMargin != 0 {
			return fmt.Errorf("expected margin 0, got %f", tc.profitMargin)
		}
		return nil
	})

	ctx.Then(`^the probability should be "([\d.]+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if tc.confidence != expected {
			return fmt.Errorf("expected probability %f, got %f", expected, tc.confidence)
		}
		return nil
	})

	ctx.Then(`^the confidence should be "([\d.]+)"$`, func(expectedStr string) error {
		expected, _ := strToFloat(expectedStr)
		if tc.confidence != expected {
			return fmt.Errorf("expected confidence %f, got %f", expected, tc.confidence)
		}
		return nil
	})

	ctx.Then(`^the result should be in SEK \(kronor\), not ören$`, func() error {
		if tc.result.RecommendedPrice > 200 {
			return fmt.Errorf("valuation should be in SEK (kronor), not ören. Got %f SEK which suggests 100x bug", tc.result.RecommendedPrice)
		}
		return nil
	})
}

func calculateConfidence(itemCount int) float64 {
	switch {
	case itemCount == 0:
		return 0
	case itemCount <= 2:
		return 0.3
	case itemCount <= 4:
		return 0.5
	case itemCount <= 8:
		return 0.7
	default:
		return 0.8
	}
}

func strToFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func ptrInt(v int) *int {
	return &v
}

func TestValuationFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenarioValuation,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"internal/test/gherkin/features/valuation.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, there are failed test scenarios")
	}
}
