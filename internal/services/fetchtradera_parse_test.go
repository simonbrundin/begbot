package services

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestFetchTraderaAdsFromURL_ParseBuyNowAndCurrentPrices(t *testing.T) {
	html := `<html><body>` +
		`<a data-link-type="next-link" href="/item/100" aria-describedby="item-100-price"><span class="item-card_title__okrrK">Item BuyNow</span><div class="item-card_priceDetails__TzN1U"><div id="item-100-bin-price" data-testid="bin-price">1 299 kr</div><div id="item-100-price" data-testid="price">1 000 kr</div></div></a>` +
		`<a data-link-type="next-link" href="/item/101" aria-describedby="item-101-price"><span class="item-card_title__okrrK">Item CurrentOnly</span><div class="item-card_priceDetails__TzN1U"><div id="item-101-price" data-testid="price">850 kr</div></div></a>` +
		`</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	m := &MarketplaceService{}
	ads := m.ParseTraderaDoc(doc)
	if len(ads) != 2 {
		t.Fatalf("expected 2 ads, got %d", len(ads))
	}

	var buyNowFound, currentFound bool
	for _, a := range ads {
		if strings.Contains(a.Link, "/item/100") {
			buyNowFound = true
			if !a.HasBuyNow {
				t.Fatalf("expected HasBuyNow=true for item/100")
			}
			if a.BuyNowPrice != 1299 {
				t.Fatalf("expected BuyNowPrice=1299 for item/100, got %.0f", a.BuyNowPrice)
			}
			if a.Price != 1299 {
				t.Fatalf("expected Price to prefer BuyNowPrice for item/100, got %.0f", a.Price)
			}
		}
		if strings.Contains(a.Link, "/item/101") {
			currentFound = true
			if a.HasBuyNow {
				t.Fatalf("expected HasBuyNow=false for item/101")
			}
			if a.CurrentPrice != 850 {
				t.Fatalf("expected CurrentPrice=850 for item/101, got %.0f", a.CurrentPrice)
			}
			if a.Price != 850 {
				t.Fatalf("expected Price=CurrentPrice for item/101, got %.0f", a.Price)
			}
		}
	}

	if !buyNowFound || !currentFound {
		t.Fatalf("did not find both sample items in parsed ads")
	}
}

func TestParseTraderaDoc_LoadsFromTestdataSamples(t *testing.T) {
	samples := []struct {
		path          string
		wantHasBuyNow bool
	}{
		{"testdata/bin_price_selector.html", true},
		{"testdata/price_details_complex.html", true},
	}

	for _, s := range samples {
		data, err := os.ReadFile(s.path)
		if err != nil {
			t.Fatalf("failed to read sample %s: %v", s.path, err)
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
		if err != nil {
			t.Fatalf("failed to parse sample %s: %v", s.path, err)
		}
		m := &MarketplaceService{}
		ads := m.ParseTraderaDoc(doc)
		if len(ads) == 0 {
			t.Fatalf("expected at least 1 ad for sample %s", s.path)
		}
		// ensure first ad has buy-now as expected
		if ads[0].HasBuyNow != s.wantHasBuyNow {
			t.Fatalf("sample %s: expected HasBuyNow=%v, got %v", s.path, s.wantHasBuyNow, ads[0].HasBuyNow)
		}
	}
}
