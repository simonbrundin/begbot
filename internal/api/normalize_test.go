package api

import (
	"math"
	"testing"

	"begbot/internal/models"
)

func makeConfig(productID int64, typeID int16, isActive bool, weight float64) models.ProductValuationTypeConfig {
	return models.ProductValuationTypeConfig{
		ProductID:       productID,
		ValuationTypeID: typeID,
		IsActive:        isActive,
		Weight:          weight,
	}
}

func totalActiveWeight(configs []models.ProductValuationTypeConfig) float64 {
	sum := 0.0
	for _, c := range configs {
		if c.IsActive {
			sum += c.Weight
		}
	}
	return sum
}

func TestNormalizeWeights_EqualDistribution(t *testing.T) {
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 0),
		makeConfig(1, 2, true, 0),
		makeConfig(1, 3, true, 0),
	}
	result := NormalizeWeights(configs)
	for _, c := range result {
		if !c.IsActive {
			continue
		}
		want := 100.0 / 3.0
		if math.Abs(c.Weight-want) > 0.001 {
			t.Errorf("expected weight ~%.4f, got %.4f", want, c.Weight)
		}
	}
	if total := totalActiveWeight(result); math.Abs(total-100.0) > 0.001 {
		t.Errorf("active weights should sum to 100, got %f", total)
	}
}

func TestNormalizeWeights_SumsTo100(t *testing.T) {
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 40),
		makeConfig(1, 2, true, 60),
		makeConfig(1, 3, false, 0),
	}
	result := NormalizeWeights(configs)
	if total := totalActiveWeight(result); math.Abs(total-100.0) > 0.001 {
		t.Errorf("active weights should sum to 100, got %f", total)
	}
	// Inactive type should have weight 0
	if result[2].Weight != 0 {
		t.Errorf("inactive type should have weight 0, got %f", result[2].Weight)
	}
}

func TestNormalizeWeights_DeactivateRedistributes(t *testing.T) {
	// Two types: one deactivated – remaining gets 100%
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 50),
		makeConfig(1, 2, false, 50),
	}
	result := NormalizeWeights(configs)
	if math.Abs(result[0].Weight-100.0) > 0.001 {
		t.Errorf("single active type should have weight 100, got %f", result[0].Weight)
	}
	if result[1].Weight != 0 {
		t.Errorf("inactive type should have weight 0, got %f", result[1].Weight)
	}
}

func TestNormalizeWeights_ReactivateGetsEqualShare(t *testing.T) {
	// Three types: one re-activated with weight 0 gets an equal share (all get 1/3).
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 60),
		makeConfig(1, 2, true, 40),
		makeConfig(1, 3, true, 0), // newly re-activated
	}
	result := NormalizeWeights(configs)
	if total := totalActiveWeight(result); math.Abs(total-100.0) > 0.001 {
		t.Errorf("active weights should sum to 100, got %f", total)
	}
	want := 100.0 / 3.0
	for idx, c := range result {
		if math.Abs(c.Weight-want) > 0.001 {
			t.Errorf("type %d: expected weight ~%.4f, got %.4f", idx+1, want, c.Weight)
		}
	}
}

func TestNormalizeWeights_ReactivateTwoTypesGetsEqual(t *testing.T) {
	// Regression: deactivate one of two types (→ 100%), reactivate it → should be 50/50.
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 100), // sole active type after the other was deactivated
		makeConfig(1, 2, true, 0),   // newly re-activated, weight was zeroed out
	}
	result := NormalizeWeights(configs)
	if math.Abs(result[0].Weight-50.0) > 0.001 {
		t.Errorf("expected type 1 weight ~50.0, got %.4f", result[0].Weight)
	}
	if math.Abs(result[1].Weight-50.0) > 0.001 {
		t.Errorf("expected type 2 weight ~50.0, got %.4f", result[1].Weight)
	}
}

func TestNormalizeWeights_NoActiveTypes(t *testing.T) {
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, false, 50),
		makeConfig(1, 2, false, 50),
	}
	result := NormalizeWeights(configs)
	for _, c := range result {
		if c.Weight != 0 {
			t.Errorf("all types inactive – expected weight 0, got %f", c.Weight)
		}
	}
}

func TestNormalizeWeights_SingleActiveType(t *testing.T) {
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 30),
	}
	result := NormalizeWeights(configs)
	if math.Abs(result[0].Weight-100.0) > 0.001 {
		t.Errorf("single active type should have weight 100, got %f", result[0].Weight)
	}
}

func TestNormalizeWeights_DoesNotMutateInput(t *testing.T) {
	configs := []models.ProductValuationTypeConfig{
		makeConfig(1, 1, true, 30),
		makeConfig(1, 2, true, 70),
	}
	original := configs[0].Weight
	NormalizeWeights(configs)
	if configs[0].Weight != original {
		t.Error("NormalizeWeights should not mutate the input slice")
	}
}
