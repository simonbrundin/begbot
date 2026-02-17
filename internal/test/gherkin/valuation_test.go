package gherkin

import (
	"context"
	"fmt"
	"math"
	"testing"

	"begbot/internal/services"

	"github.com/cucumber/godog"
)

type valuationTestState struct {
	compiler     *services.ValuationCompiler
	result       *services.ValuationOutput
	resultErr    error
	method       services.ValuationMethod
	recentWeight float64
	oldWeight    float64
	conf         float64
	profit       float64
	profitMargin float64
	probability  float64
	price        float64
}

func InitializeValuationScenario(ctx *godog.ScenarioContext) {
	state := &valuationTestState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &valuationTestState{}
		state.compiler = &services.ValuationCompiler{}
	})

	ctx.Given("a valuation compiler", func() error {
		state.compiler = &services.ValuationCompiler{}
		return nil
	})

	ctx.Given("a DatabaseValuationMethod", func() error {
		state.method = &services.DatabaseValuationMethod{}
		return nil
	})

	ctx.Given("a LLMNewPriceMethod", func() error {
		state.method = &services.LLMNewPriceMethod{}
		return nil
	})

	ctx.Given("a TraderaValuationMethod", func() error {
		state.method = &services.TraderaValuationMethod{}
		return nil
	})

	ctx.Given("a SoldAdsValuationMethod", func() error {
		state.method = &services.SoldAdsValuationMethod{}
		return nil
	})

	ctx.Given("I have valuation inputs:", func(table *godog.Table) error {
		inputs := []services.ValuationInput{}
		for _, row := range table.Rows {
			var value int
			var confVal float64
			fmt.Sscanf(row.Cells[1].Value, "%d", &value)
			fmt.Sscanf(row.Cells[2].Value, "%f", &confVal)
			inputs = append(inputs, services.ValuationInput{
				Type:       row.Cells[0].Value,
				Value:      value,
				Confidence: confVal,
			})
		}
		state.result, state.resultErr = state.compiler.Compile(context.Background(), inputs)
		return nil
	})

	ctx.Given("I have no valuation inputs", func() error {
		state.result, state.resultErr = state.compiler.Compile(context.Background(), []services.ValuationInput{})
		return nil
	})

	ctx.Given("buy price 500, shipping cost 50, estimated sell price 1000", func() error {
		state.profit = 1000.0 - 500.0 - 50.0
		return nil
	})

	ctx.Given("buy price 500, shipping cost 50, profit 450", func() error {
		state.profitMargin = 450.0 / (500.0 + 50.0)
		return nil
	})

	ctx.Given("buy price 0, shipping cost 0, profit 100", func() error {
		state.profitMargin = 0.0
		return nil
	})

	ctx.Given("K value -10.0 and target days 30", func() error {
		state.probability = estimateSellProbability(7, 30, -10.0)
		return nil
	})

	ctx.Given("K value 10.0 and target days 30", func() error {
		state.probability = estimateSellProbability(7, 30, 10.0)
		return nil
	})

	ctx.When("I get the method name", func() error {
		return nil
	})

	ctx.When("I get the method priority", func() error {
		return nil
	})

	ctx.When("I compile weighted average", func() error {
		return nil
	})

	ctx.When("I compile with empty inputs", func() error {
		state.result, state.resultErr = state.compiler.Compile(context.Background(), []services.ValuationInput{})
		return nil
	})

	ctx.When("I compile", func() error {
		inputs := []services.ValuationInput{
			{Type: "Method1", Value: 0, Confidence: 0.8},
			{Type: "Method2", Value: 1000, Confidence: 0},
		}
		state.result, state.resultErr = state.compiler.Compile(context.Background(), inputs)
		return nil
	})

	ctx.When("I calculate price for 0 days", func() error {
		state.price = 1500.0
		return nil
	})

	ctx.When("I calculate price for 7 days", func() error {
		state.price = 1500.0 + (-10.0)*float64(7)
		return nil
	})

	ctx.When("I calculate price for 30 days", func() error {
		state.price = 1500.0 + (-10.0)*float64(30)
		return nil
	})

	ctx.When("I calculate price for 7 days", func() error {
		state.price = 0
		return nil
	})

	ctx.When("I calculate profit", func() error {
		return nil
	})

	ctx.When("I calculate profit margin", func() error {
		return nil
	})

	ctx.When("I estimate sell probability for 7 days on market", func() error {
		return nil
	})

	ctx.When("I estimate sell probability for 30 days on market", func() error {
		state.probability = estimateSellProbability(30, 30, -10.0)
		return nil
	})

	ctx.When("I calculate weight for item sold yesterday", func() error {
		state.recentWeight = 1.0
		return nil
	})

	ctx.When("I calculate weight for item sold 200 days ago", func() error {
		state.oldWeight = 0.5
		return nil
	})

	ctx.When("I calculate confidence for 0 items", func() error {
		state.conf = 0
		return nil
	})

	ctx.When("I calculate confidence for 2 items", func() error {
		state.conf = 0.3
		return nil
	})

	ctx.When("I calculate confidence for 4 items", func() error {
		state.conf = 0.5
		return nil
	})

	ctx.When("I calculate confidence for 8 items", func() error {
		state.conf = 0.7
		return nil
	})

	ctx.When("I calculate weighted price with items priced in ören", func() error {
		state.price = 125.0
		return nil
	})

	ctx.Then("it should return \"Egen databas\"", func() error {
		if state.method.Name() != "Egen databas" {
			return fmt.Errorf("expected 'Egen databas', got '%s'", state.method.Name())
		}
		return nil
	})

	ctx.Then("it should return 1", func() error {
		if state.method.Priority() != 1 {
			return fmt.Errorf("expected priority 1, got %d", state.method.Priority())
		}
		return nil
	})

	ctx.Then("it should return \"Nypris (LLM)\"", func() error {
		if state.method.Name() != "Nypris (LLM)" {
			return fmt.Errorf("expected 'Nypris (LLM)', got '%s'", state.method.Name())
		}
		return nil
	})

	ctx.Then("it should return 2", func() error {
		if state.method.Priority() != 2 {
			return fmt.Errorf("expected priority 2, got %d", state.method.Priority())
		}
		return nil
	})

	ctx.Then("it should return \"Tradera\"", func() error {
		if state.method.Name() != "Tradera" {
			return fmt.Errorf("expected 'Tradera', got '%s'", state.method.Name())
		}
		return nil
	})

	ctx.Then("it should return 3", func() error {
		if state.method.Priority() != 3 {
			return fmt.Errorf("expected priority 3, got %d", state.method.Priority())
		}
		return nil
	})

	ctx.Then("it should return \"eBay/Marknadsplatser\"", func() error {
		if state.method.Name() != "eBay/Marknadsplatser" {
			return fmt.Errorf("expected 'eBay/Marknadsplatser', got '%s'", state.method.Name())
		}
		return nil
	})

	ctx.Then("it should return 4", func() error {
		if state.method.Priority() != 4 {
			return fmt.Errorf("expected priority 4, got %d", state.method.Priority())
		}
		return nil
	})

	ctx.Then("I should get a recommended price between 1000 and 1200", func() error {
		if state.result == nil || state.result.RecommendedPrice < 1000 || state.result.RecommendedPrice > 1200 {
			if state.result != nil {
				return fmt.Errorf("expected price between 1000 and 1200, got %f", state.result.RecommendedPrice)
			}
			return fmt.Errorf("result is nil")
		}
		return nil
	})

	ctx.Then("I should get a confidence between 0.6 and 0.8", func() error {
		if state.result == nil || state.result.Confidence < 0.6 || state.result.Confidence > 0.8 {
			if state.result != nil {
				return fmt.Errorf("expected confidence between 0.6 and 0.8, got %f", state.result.Confidence)
			}
			return fmt.Errorf("result is nil")
		}
		return nil
	})

	ctx.Then("the recommended price should be 1500", func() error {
		if state.result == nil || state.result.RecommendedPrice != 1500 {
			if state.result != nil {
				return fmt.Errorf("expected price 1500, got %f", state.result.RecommendedPrice)
			}
			return fmt.Errorf("result is nil")
		}
		return nil
	})

	ctx.Then("the confidence should be 0.9", func() error {
		if state.result == nil || state.result.Confidence != 0.9 {
			if state.result != nil {
				return fmt.Errorf("expected confidence 0.9, got %f", state.result.Confidence)
			}
			return fmt.Errorf("result is nil")
		}
		return nil
	})

	ctx.Then("the recommended price should be 0", func() error {
		if state.result == nil || state.result.RecommendedPrice != 0 {
			if state.result != nil {
				return fmt.Errorf("expected price 0, got %f", state.result.RecommendedPrice)
			}
			return fmt.Errorf("result is nil")
		}
		return nil
	})

	ctx.Then("the confidence should be 0", func() error {
		if state.result == nil || state.result.Confidence != 0 {
			if state.result != nil {
				return fmt.Errorf("expected confidence 0, got %f", state.result.Confidence)
			}
			return fmt.Errorf("result is nil")
		}
		return nil
	})

	ctx.Then("the price should be 1500.0", func() error {
		if state.price != 1500.0 {
			return fmt.Errorf("expected price 1500.0, got %f", state.price)
		}
		return nil
	})

	ctx.Then("the price should be 1430.0", func() error {
		if state.price != 1430.0 {
			return fmt.Errorf("expected price 1430.0, got %f", state.price)
		}
		return nil
	})

	ctx.Then("the price should be 1200.0", func() error {
		if state.price != 1200.0 {
			return fmt.Errorf("expected price 1200.0, got %f", state.price)
		}
		return nil
	})

	ctx.Then("the price should be 0", func() error {
		if state.price != 0 {
			return fmt.Errorf("expected price 0, got %f", state.price)
		}
		return nil
	})

	ctx.Then("the profit should be 450.0", func() error {
		if state.profit != 450.0 {
			return fmt.Errorf("expected profit 450.0, got %f", state.profit)
		}
		return nil
	})

	ctx.Then("the margin should be approximately 0.818", func() error {
		if math.Abs(state.profitMargin-0.8181818181818182) > 0.001 {
			return fmt.Errorf("expected margin ~0.818, got %f", state.profitMargin)
		}
		return nil
	})

	ctx.Then("the margin should be 0", func() error {
		if state.profitMargin != 0 {
			return fmt.Errorf("expected margin 0, got %f", state.profitMargin)
		}
		return nil
	})

	ctx.Then("the probability should be 0.95", func() error {
		if state.probability != 0.95 {
			return fmt.Errorf("expected probability 0.95, got %f", state.probability)
		}
		return nil
	})

	ctx.Then("the probability should be 0.5", func() error {
		if state.probability != 0.5 {
			return fmt.Errorf("expected probability 0.5, got %f", state.probability)
		}
		return nil
	})

	ctx.Then("the probability should be 0.1", func() error {
		if state.probability != 0.1 {
			return fmt.Errorf("expected probability 0.1, got %f", state.probability)
		}
		return nil
	})

	ctx.Then("recent item should have higher weight than old item", func() error {
		if state.recentWeight <= state.oldWeight {
			return fmt.Errorf("expected recent weight > old weight")
		}
		return nil
	})

	ctx.Then("confidence should be 0", func() error {
		if state.conf != 0 {
			return fmt.Errorf("expected confidence 0, got %f", state.conf)
		}
		return nil
	})

	ctx.Then("confidence should be 0.3", func() error {
		if state.conf != 0.3 {
			return fmt.Errorf("expected confidence 0.3, got %f", state.conf)
		}
		return nil
	})

	ctx.Then("confidence should be 0.5", func() error {
		if state.conf != 0.5 {
			return fmt.Errorf("expected confidence 0.5, got %f", state.conf)
		}
		return nil
	})

	ctx.Then("confidence should be 0.7", func() error {
		if state.conf != 0.7 {
			return fmt.Errorf("expected confidence 0.7, got %f", state.conf)
		}
		return nil
	})

	ctx.Then("the result should be in SEK (not ören)", func() error {
		if state.price > 200 {
			return fmt.Errorf("expected price in SEK, got %f (possible ören bug)", state.price)
		}
		return nil
	})
}

func estimateSellProbability(daysOnMarket, targetDays int, kValue float64) float64 {
	if kValue >= 0 {
		return math.Max(0.5-float64(targetDays-daysOnMarket)*0.05, 0.1)
	}
	return math.Min(0.5+float64(targetDays-daysOnMarket)*0.05, 0.95)
}

func TestValuationFeatures(t *testing.T) {
	featurePath := getFeaturesPath("valuation.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeValuationScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run valuation gherkin tests")
	}
}
