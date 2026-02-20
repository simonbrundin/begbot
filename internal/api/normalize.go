package api

import "begbot/internal/models"

// NormalizeWeights redistributes weights for active valuation type configs so that
// the sum of weights for active types equals 100.0. Inactive types get weight 0.
//
// Rules:
//   - If no active types, returns unchanged.
//   - Active types with weight <= 0 are assigned an equal share of the total positive
//     weight (or 100/n if no positive weights exist), then all active weights are
//     normalised proportionally to sum to 100.
func NormalizeWeights(configs []models.ProductValuationTypeConfig) []models.ProductValuationTypeConfig {
	result := make([]models.ProductValuationTypeConfig, len(configs))
	copy(result, configs)

	// Collect indices of active configs and zero out inactive weights.
	var activeIdx []int
	for i, c := range result {
		if c.IsActive {
			activeIdx = append(activeIdx, i)
		} else {
			result[i].Weight = 0
		}
	}

	n := len(activeIdx)
	if n == 0 {
		return result
	}

	// Sum positive weights among active types and check if any are missing a weight.
	totalPositive := 0.0
	allHaveWeight := true
	for _, i := range activeIdx {
		if result[i].Weight > 0 {
			totalPositive += result[i].Weight
		} else {
			allHaveWeight = false
		}
	}

	if !allHaveWeight || totalPositive <= 0 {
		// Any active type without a weight (e.g. newly re-activated) → equal distribution.
		w := 100.0 / float64(n)
		for _, i := range activeIdx {
			result[i].Weight = w
		}
		return result
	}

	// All active types have positive weights – normalise proportionally to sum to 100.
	for _, i := range activeIdx {
		result[i].Weight = result[i].Weight / totalPositive * 100.0
	}

	return result
}
