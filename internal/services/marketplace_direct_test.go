package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func buildTraderaHTML(count int, injectAntiBot bool) string {
	var b strings.Builder
	b.WriteString("<html><body>")
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("item-card-123%d", i)
		b.WriteString(fmt.Sprintf(`<a data-link-type="next-link" href="/item/123%d" aria-describedby="%s-price">`, i, id))
		b.WriteString(fmt.Sprintf(`<span class="item-card_title__okrrK">Title %d</span>`, i))
		b.WriteString(fmt.Sprintf(`<img src="//img.example/%d.jpg"/>`, i))
		b.WriteString(`</a>`)
	}
	for i := 1; i <= count; i++ {
		id := fmt.Sprintf("item-card-123%d", i)
		b.WriteString(fmt.Sprintf(`<div id="%s-price" data-testid="bin-price">2 828 kr</div>`, id))
	}
	if injectAntiBot {
		b.WriteString("<div>Please enable JavaScript to view this site. CAPTCHA</div>")
	}
	b.WriteString("</body></html>")
	return b.String()
}

func TestDirectFetchSuccess(t *testing.T) {
	html := buildTraderaHTML(12, false)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	s := &MarketplaceService{}
	ctx := context.Background()
	// use hook
	s.fetchDirect = s.fetchTraderaAdsDirect
	ads, blocked, err := s.fetchDirect(ctx, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blocked {
		t.Fatalf("expected not blocked, got blocked")
	}
	if len(ads) != 12 {
		t.Fatalf("expected 12 ads, got %d", len(ads))
	}
	for _, ad := range ads {
		if ad.Price != 2828 {
			t.Errorf("expected price 2828, got %v for ad %v", ad.Price, ad.Link)
		}
		if ad.Title == "" {
			t.Errorf("expected non-empty title for %v", ad.Link)
		}
		if ad.Marketplace != "tradera" {
			t.Errorf("expected marketplace tradera, got %s", ad.Marketplace)
		}
	}
}

func TestDirectFetchBlockedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte("Forbidden"))
	}))
	defer srv.Close()

	s := &MarketplaceService{}
	ctx := context.Background()
	s.fetchDirect = s.fetchTraderaAdsDirect
	ads, blocked, err := s.fetchDirect(ctx, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Fatalf("expected blocked=true for 403 response")
	}
	if ads != nil {
		t.Fatalf("expected no ads when blocked, got %d", len(ads))
	}
}

func TestDirectFetchBlockedAntiBotMarker(t *testing.T) {
	html := buildTraderaHTML(12, true)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	s := &MarketplaceService{}
	ctx := context.Background()
	s.fetchDirect = s.fetchTraderaAdsDirect
	ads, blocked, err := s.fetchDirect(ctx, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Fatalf("expected blocked due to anti-bot marker")
	}
	if ads != nil {
		t.Fatalf("expected no ads when blocked, got %d", len(ads))
	}
}

func TestDirectFetchInsufficientLinks(t *testing.T) {
	html := buildTraderaHTML(2, false)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	s := &MarketplaceService{}
	ctx := context.Background()
	s.fetchDirect = s.fetchTraderaAdsDirect
	ads, blocked, err := s.fetchDirect(ctx, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Fatalf("expected blocked due to insufficient links")
	}
	if ads != nil {
		t.Fatalf("expected no ads when blocked, got %d", len(ads))
	}
}
