package services

import "testing"

func TestParsePriceVariants(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"2 828 kr", 2828},
		{"2 828 kr", 2828},
		{"1\u00A0234 kr", 1234},
		{"3,5", 3.5},
		{"1.234,56", 1234.56},
		{"1,234.56", 1234.56},
		{"2830", 2830},
		{"Pris: 2 000 kr", 2000},
		{"", 0},
	}

	for _, c := range cases {
		got := parsePrice(c.in)
		if got != c.want {
			t.Fatalf("parsePrice(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
