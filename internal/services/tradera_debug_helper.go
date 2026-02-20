package services

import (
	"begbot/internal/config"
	"context"
)

// RunTraderaValuation is a small exported helper used by local debug tools.
// It creates a ValuationService and runs the Tradera valuator for the
// provided product info. This keeps debugging code inside the services
// package so it can access unexported fields safely.
func RunTraderaValuation(cfg *config.Config, pi ProductInfo) (*ValuationInput, error) {
	svc := NewValuationService(cfg, nil, nil)
	method := &TraderaValuationMethod{svc: svc}
	return method.Valuate(context.Background(), pi)
}
