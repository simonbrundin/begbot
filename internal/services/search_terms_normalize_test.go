package services

import (
	"testing"
)

func TestNormalizeSearchURL_TraderaQuery(t *testing.T) {
	input := "lego star wars"
	mp := int64(2)
	got := NormalizeSearchURL(input, &mp)
	want := "https://www.tradera.com/search?q=lego+star+wars"
	if got != want {
		t.Fatalf("normalized URL wrong: got %q want %q", got, want)
	}
}

func TestNormalizeSearchURL_UrlLikeUnchanged(t *testing.T) {
	input := "https://example.com/search?q=foo"
	got := NormalizeSearchURL(input, nil)
	if got != input {
		t.Fatalf("expected URL unchanged, got %q", got)
	}
}

func TestNormalizeSearchURL_OtherMarketplaceUnchanged(t *testing.T) {
	input := "lego star wars"
	mp := int64(1)
	got := NormalizeSearchURL(input, &mp)
	if got != input {
		t.Fatalf("expected input preserved for non-Tradera marketplace, got %q", got)
	}
}
