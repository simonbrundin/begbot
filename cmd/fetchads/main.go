package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/marketplaces"
	"begbot/internal/services"

	"github.com/joho/godotenv"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

// loggingRoundTripper wraps an underlying RoundTripper and logs outgoing requests.
type loggingRoundTripper struct{ rt http.RoundTripper }

func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("--- OUTGOING REQUEST: %s %s\n", req.Method, req.URL.String())
	for k, vv := range req.Header {
		for _, v := range vv {
			fmt.Printf("%s: %s\n", k, v)
		}
	}
	fmt.Printf("--- END HEADERS\n")
	if l.rt == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return l.rt.RoundTrip(req)
}

func main() {
	godotenv.Load()
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// quick probe mode: `fetchads probe [query] [itemID]`
	if len(os.Args) > 1 && os.Args[1] == "probe" {
		query := "iphone"
		itemID := 0
		if len(os.Args) > 2 {
			query = os.Args[2]
		}
		if len(os.Args) > 3 {
			if id, perr := strconv.Atoi(os.Args[3]); perr == nil {
				itemID = id
			}
		}
		// lightweight probe: run a search via Tradera API (if configured) and
		// optionally call GetItem for provided itemID to show headers/responses.
		trc := marketplaces.NewTraderaClient(&cfg.Scraping.Tradera)
		fmt.Printf("Probe: query=%q, probeItem=%d\n", query, itemID)
		ctx := context.Background()
		// wrap existing transport (or default) so we log both Search and GetItem requests
		trc.WrapTransport(func(base http.RoundTripper) http.RoundTripper {
			return loggingRoundTripper{rt: base}
		})

		ads, err := trc.FetchAds(ctx, query)
		if err != nil {
			fmt.Printf("FetchAds error: %v\n", err)
		} else {
			fmt.Printf("FetchAds returned %d ads (showing up to 5):\n", len(ads))
			for i, a := range ads {
				if i >= 5 {
					break
				}
				fmt.Printf(" - %s (price=%.2f, hasBuyNow=%v)\n", a.Link, a.Price, a.HasBuyNow)
			}
		}
		if itemID > 0 {
			url := fmt.Sprintf("https://www.tradera.com/item/%d/1/x", itemID)
			fmt.Printf("Probing GetItem for %s\n", url)
			details, derr := trc.FetchAdDetails(ctx, url)
			if derr != nil {
				fmt.Printf("FetchAdDetails error: %v\n", derr)
			} else if details != nil {
				fmt.Printf("GetItem title=%q price=%.2f buyNow=%.2f images=%v seller=%s\n",
					details.Title, details.Price, details.BuyNowPrice, details.ImageURLs, details.SellerName)
			}
		}
		return
	}

	var database *db.Postgres
	for i := 0; i < 3; i++ {
		database, err = db.NewPostgres(cfg.Database)
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}
	defer database.Close()

	log.Println("Running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully!")

	marketplaceService := services.NewMarketplaceService(cfg)
	cacheService := services.NewCacheService(cfg)
	llmService := services.NewLLMService(cfg)
	valuationService := services.NewValuationService(cfg, database, llmService)
	botService := services.NewBotService(cfg, marketplaceService, cacheService, llmService, valuationService, database)

	log.Println("Starting ad fetch...")
	if err := botService.Run(); err != nil {
		log.Fatalf("Ad fetch failed: %v", err)
	}

	log.Println("Ad fetch completed successfully!")
}
