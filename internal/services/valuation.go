package services

import (
	"context"
	"math"

	"begbot/internal/config"
	"begbot/internal/db"
)

type ValuationService struct {
	cfg      *config.Config
	database *db.Postgres
}

func NewValuationService(cfg *config.Config, database *db.Postgres) *ValuationService {
	return &ValuationService{
		cfg:      cfg,
		database: database,
	}
}

func (s *ValuationService) GetHistoricalValuation(ctx context.Context, marketplace string) (*HistoricalValuation, error) {
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
		if item.BuyDate == nil || item.SellDate == nil || item.SellPrice == nil {
			continue
		}
		daysOnMarket := int(item.SellDate.Sub(*item.BuyDate).Hours() / 24)
		x := float64(daysOnMarket)
		y := float64(*item.SellPrice)

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
