package services

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestParseTraderaDoc_Samples(t *testing.T) {
	cases := []struct {
		name string
		html string
		want int
	}{
		{
			name: "single-item-aria-price",
			html: `<html><body>` +
				`<a data-link-type="next-link" href="/item/1" aria-describedby="item-card-11-price"><span class="item-card_title__okrrK">Title A</span><img src="//images.example/a.jpg"/></a>` +
				`<div id="item-card-11-price" data-testid="bin-price">1 234 kr</div>` +
				`</body></html>`,
			want: 1,
		},
		{
			name: "title-attr-and-price-class",
			html: `<html><body>` +
				`<a data-link-type="next-link" href="/item/2" title="Title B"><img src="https://images.example/b.jpg"/></a>` +
				`<div id="other-2-price" data-testid="price">2 000 kr</div>` +
				`</body></html>`,
			// price id doesn't match aria so price association is skipped; we still expect the item
			want: 1,
		},
		{
			name: "multiple-items",
			html: `<html><body>` +
				`<a data-link-type="next-link" href="/item/3" aria-describedby="item-card-33-price"><span class="item-card_title__okrrK">C</span></a>` +
				`<div id="item-card-33-price" data-testid="bin-price">500 kr</div>` +
				`<a data-link-type="next-link" href="/item/4" aria-describedby="item-card-44-price"><span class="item-card_title__okrrK">D</span></a>` +
				`<div id="item-card-44-price" data-testid="price">750 kr</div>` +
				`</body></html>`,
			want: 2,
		},
	}

	for _, c := range cases {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(c.html))
		if err != nil {
			t.Fatalf("%s: failed to parse html: %v", c.name, err)
		}
		// ParseTraderaDoc is a method on MarketplaceService; construct a service for testing
		m := &MarketplaceService{}
		ads := m.ParseTraderaDoc(doc)
		if len(ads) != c.want {
			t.Fatalf("%s: expected %d ads, got %d; ads=%+v", c.name, c.want, len(ads), ads)
		}

		// Additional: load real-ish sample from testdata/evomi_sample_1.html and ensure parse works
		data, err := os.ReadFile("testdata/evomi_sample_1.html")
		if err != nil {
			t.Fatalf("failed to read testdata sample: %v", err)
		}
		doc2, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
		if err != nil {
			t.Fatalf("failed to parse testdata sample: %v", err)
		}
		m2 := &MarketplaceService{}
		ads2 := m2.ParseTraderaDoc(doc2)
		if len(ads2) != 2 {
			t.Fatalf("testdata sample: expected 2 ads, got %d", len(ads2))
		}
		// basic checks: links are absolute and titles non-empty when available
		for _, a := range ads {
			if !strings.HasPrefix(a.Link, "https://") {
				t.Fatalf("%s: expected absolute link, got %s", c.name, a.Link)
			}
		}
	}
}
