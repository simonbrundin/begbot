package services

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseTraderaAdPage(t *testing.T) {
	cases := []struct {
		name           string
		html           string
		wantBuyNow     float64
		wantCurrent    float64
		wantTitle      string
		wantImageCount int
	}{
		{
			name: "buy-now-with-data-testid",
			html: `<html><body><h1>Test Product</h1>
				<div data-testid="bin-price">1 299 kr</div>
				<img src="https://images.example.com/1.jpg"/>
				</body></html>`,
			wantBuyNow:     1299,
			wantCurrent:    1299,
			wantTitle:      "Test Product",
			wantImageCount: 1,
		},
		{
			name: "buy-now-in-class",
			html: `<html><body><h1>Another Product</h1>
				<div class="bin-price">2 500 kr</div>
				<img src="https://images.example.com/2.jpg"/>
				</body></html>`,
			wantBuyNow:     2500,
			wantCurrent:    2500,
			wantTitle:      "Another Product",
			wantImageCount: 1,
		},
		{
			name: "separate-buy-now-and-current",
			html: `<html><body><h1>Separate Prices</h1>
				<div data-testid="price">800 kr</div>
				<div data-testid="bin-price">1 500 kr</div>
				</body></html>`,
			wantBuyNow:     1500,
			wantCurrent:    800,
			wantTitle:      "Separate Prices",
			wantImageCount: 0,
		},
		{
			name: "buy-now-with-kop-nu-text",
			html: `<html><body><h1>Kop Nu Product</h1>
				<span>Köp nu 999 kr</span>
				</body></html>`,
			wantBuyNow:  999,
			wantCurrent: 999,
			wantTitle:   "Kop Nu Product",
		},
		{
			name: "no-buy-now",
			html: `<html><body><h1>Auction Only</h1>
				<div data-testid="price">100 kr</div>
				</body></html>`,
			wantBuyNow:  0,
			wantCurrent: 100,
			wantTitle:   "Auction Only",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			details, err := ParseTraderaAdPage([]byte(c.html))
			if err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			if details.BuyNowPrice != c.wantBuyNow {
				t.Errorf("BuyNowPrice = %v, want %v", details.BuyNowPrice, c.wantBuyNow)
			}
			if details.CurrentPrice != c.wantCurrent {
				t.Errorf("CurrentPrice = %v, want %v", details.CurrentPrice, c.wantCurrent)
			}
			if c.wantTitle != "" && details.Title != c.wantTitle {
				t.Errorf("Title = %q, want %q", details.Title, c.wantTitle)
			}
			if c.wantImageCount > 0 && len(details.ImageURLs) != c.wantImageCount {
				t.Errorf("ImageURLs count = %v, want %v", len(details.ImageURLs), c.wantImageCount)
			}
		})
	}
}

func TestParseTraderaAdPageIntegration(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<!DOCTYPE html>
		<html>
		<head><title>Test</title></head>
		<body>
			<h1>Mac Mini M4</h1>
			<div data-testid="bin-price">5 990 kr</div>
			<div data-testid="price">4 990 kr</div>
			<img src="https://images.tradera.com/1.jpg"/>
			<img src="https://images.tradera.com/2.jpg"/>
		</body>
		</html>
	`))
	if err != nil {
		t.Fatalf("failed to create doc: %v", err)
	}

	body, _ := doc.Html()
	details, err := ParseTraderaAdPage([]byte(body))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if details.BuyNowPrice != 5990 {
		t.Errorf("BuyNowPrice = %v, want 5990", details.BuyNowPrice)
	}
	if details.CurrentPrice != 4990 {
		t.Errorf("CurrentPrice = %v, want 4990", details.CurrentPrice)
	}
	if !strings.Contains(details.Title, "Mac Mini") {
		t.Errorf("Title = %v, should contain 'Mac Mini'", details.Title)
	}
	if len(details.ImageURLs) != 2 {
		t.Errorf("ImageURLs count = %v, want 2", len(details.ImageURLs))
	}
}
