package services

import (
	"testing"
)

func TestProductIntactResult_Struct(t *testing.T) {
	result := ProductIntactResult{
		IsIntact:          true,
		HasMinorScratches: false,
		IssuesFound:       []string{},
		Reasoning:         "Produkten är i bra skick",
	}

	if !result.IsIntact {
		t.Error("Expected IsIntact to be true")
	}

	if result.HasMinorScratches {
		t.Error("Expected HasMinorScratches to be false")
	}

	if len(result.IssuesFound) != 0 {
		t.Errorf("Expected empty IssuesFound, got %v", result.IssuesFound)
	}
}

func TestProductIntactResult_WithIssues(t *testing.T) {
	result := ProductIntactResult{
		IsIntact:          false,
		HasMinorScratches: false,
		IssuesFound:       []string{"sprucken skärm", "vattenskada"},
		Reasoning:         "Produkten har flera skador",
	}

	if result.IsIntact {
		t.Error("Expected IsIntact to be false")
	}

	if len(result.IssuesFound) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(result.IssuesFound))
	}
}

func TestProductIntactResult_WithMinorScratches(t *testing.T) {
	result := ProductIntactResult{
		IsIntact:          true,
		HasMinorScratches: true,
		IssuesFound:       []string{},
		Reasoning:         "Endast mindre repor på baksidan",
	}

	if !result.IsIntact {
		t.Error("Expected IsIntact to be true (minor scratches are acceptable)")
	}

	if !result.HasMinorScratches {
		t.Error("Expected HasMinorScratches to be true")
	}
}

func TestProductIntactResult_IsValidForPurchase(t *testing.T) {
	tests := []struct {
		name     string
		result   ProductIntactResult
		expected bool
	}{
		{
			name: "intact without scratches",
			result: ProductIntactResult{
				IsIntact:          true,
				HasMinorScratches: false,
			},
			expected: true,
		},
		{
			name: "intact with minor scratches",
			result: ProductIntactResult{
				IsIntact:          true,
				HasMinorScratches: true,
			},
			expected: true,
		},
		{
			name: "not intact",
			result: ProductIntactResult{
				IsIntact:          false,
				HasMinorScratches: false,
				IssuesFound:       []string{"sprucken skärm"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.result.IsIntact
			if isValid != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, isValid)
			}
		})
	}
}
