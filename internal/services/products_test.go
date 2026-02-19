package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"begbot/internal/models"
)

// Test that products with NULL created_at are handled correctly
// After fix: CreatedAt is *time.Time, so NULL becomes nil and is omitted from JSON
func TestProductModel_CreatedAt_ShouldBeNullNotZeroTime(t *testing.T) {
	brand := "TestBrand"
	name := "TestName"
	enabled := true
	product := models.Product{
		ID:        1,
		Brand:     &brand,
		Name:      &name,
		Enabled:   &enabled,
		CreatedAt: nil, // NULL from database
	}

	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Failed to marshal product: %v", err)
	}

	// After fix: created_at should be omitted from JSON when nil
	// This prevents frontend from showing "Invalid Date"
	if strings.Contains(string(jsonData), `"created_at"`) {
		t.Errorf("created_at should be omitted when nil, got: %s", jsonData)
	}

	t.Logf("Success: NULL created_at is omitted from JSON: %s", jsonData)
}

// Test with valid created_at - should be included
func TestProductModel_CreatedAt_WithValue(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	brand := "TestBrand"
	name := "TestName"
	enabled := true

	product := models.Product{
		ID:        1,
		Brand:     &brand,
		Name:      &name,
		Enabled:   &enabled,
		CreatedAt: &testTime,
	}

	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Failed to marshal product: %v", err)
	}

	// Should contain created_at with the value
	if !strings.Contains(string(jsonData), `"created_at"`) {
		t.Error("created_at should be present when not nil")
	}

	t.Logf("Valid created_at in JSON: %s", jsonData)
}

// Test that brand and name are properly displayed even when empty
func TestProductModel_EmptyBrandAndName_ShouldDisplay(t *testing.T) {
	testTime := time.Now()
	emptyStr := ""
	falseBool := false

	product := models.Product{
		ID:        1,
		Brand:     &emptyStr, // Empty - should be visible
		Name:      &emptyStr, // Empty - should be visible
		Enabled:   &falseBool,
		CreatedAt: &testTime,
	}

	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Failed to marshal product: %v", err)
	}

	if !strings.Contains(string(jsonData), `"brand":""`) {
		t.Error("Empty brand should be preserved in JSON")
	}

	if !strings.Contains(string(jsonData), `"name":""`) {
		t.Error("Empty name should be preserved in JSON")
	}

	t.Logf("Empty brand/name JSON: %s", jsonData)
}

// Test enabled field - it's currently boolean, frontend should convert to Ja/Nej
func TestProductEnabled_FrontendShouldShowJaNej(t *testing.T) {
	testTime := time.Now()
	brand := "TestBrand"
	name := "TestName"
	enabled := true

	product := models.Product{
		ID:        1,
		Brand:     &brand,
		Name:      &name,
		Enabled:   &enabled,
		CreatedAt: &testTime,
	}

	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Failed to marshal product: %v", err)
	}

	if !strings.Contains(string(jsonData), `"enabled":true`) {
		t.Error("enabled should be true")
	}

	t.Logf("Current JSON: %s", jsonData)
	t.Log("NOTE: Frontend must convert true→'Ja' and false→'Nej' for display")
}

// Edge case: Product with all NULL/zero values
func TestProductModel_AllNullFields(t *testing.T) {
	emptyStr := ""
	falseBool := false
	product := models.Product{
		ID:                0,
		Brand:             &emptyStr,
		Name:              &emptyStr,
		Category:          &emptyStr,
		ModelVariant:      nil,
		SellPackagingCost: 0,
		SellPostageCost:   0,
		NewPrice:          nil,
		Enabled:           &falseBool,
		CreatedAt:         nil, // NULL - will be omitted due to omitempty
	}

	jsonData, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("Failed to marshal product: %v", err)
	}

	t.Logf("All null fields JSON: %s", jsonData)

	// After fix: created_at should be omitted (not zero time)
	if strings.Contains(string(jsonData), "0001-01-01T00:00:00Z") {
		t.Errorf("BUG: created_at should not be zero time - causes 'Invalid Date' in frontend")
	}
}
