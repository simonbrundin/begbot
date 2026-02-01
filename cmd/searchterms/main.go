package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
	"begbot/internal/services"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	postgres, err := db.NewPostgres(cfg.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer postgres.Close()

	searchTermSvc := services.NewSearchTermService(postgres)
	llmSvc := services.NewLLMService(cfg)
	valuationSvc := services.NewValuationService(cfg, postgres)
	botSvc := services.NewBotService(cfg, nil, nil, llmSvc, valuationSvc, postgres)

	command := flag.String("cmd", "list", "Command to run: list, add, run, deactivate")
	description := flag.String("description", "", "Search term description")
	url := flag.String("url", "", "Search URL (copy from browser)")
	marketplaceID := flag.Int64("marketplace", 0, "Marketplace ID (1=blocket, 2=tradera)")
	id := flag.Int64("id", 0, "Search term ID")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch *command {
	case "list":
		listSearchTerms(ctx, searchTermSvc)
	case "add":
		addSearchTerm(ctx, searchTermSvc, *description, *url, *marketplaceID)
	case "run":
		runSearchTerms(ctx, searchTermSvc, postgres, botSvc)
	case "deactivate":
		deactivateSearchTerm(ctx, searchTermSvc, *id)
	default:
		fmt.Println("Unknown command. Available commands: list, add, run, deactivate")
		os.Exit(1)
	}
}

func listSearchTerms(ctx context.Context, svc *services.SearchTermService) {
	terms, err := svc.GetActiveSearchTerms(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list search terms: %v\n", err)
		os.Exit(1)
	}

	for _, term := range terms {
		marketplaceName := "unknown"
		if term.MarketplaceID != nil {
			marketplaceName = fmt.Sprintf("%d", *term.MarketplaceID)
		}
		fmt.Printf("ID: %d | Description: %s | Marketplace: %s | URL: %s\n", term.ID, term.Description, marketplaceName, term.URL)
	}
}

func addSearchTerm(ctx context.Context, svc *services.SearchTermService, description, url string, marketplaceID int64) {
	if description == "" || url == "" || marketplaceID == 0 {
		fmt.Println("Error: description, url, and marketplace are required")
		flag.Usage()
		os.Exit(1)
	}

	term, err := svc.CreateSearchTerm(ctx, description, url, marketplaceID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create search term: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created search term: ID=%d, Description=%s, Marketplace=%d\n", term.ID, term.Description, *term.MarketplaceID)
}

func runSearchTerms(ctx context.Context, svc *services.SearchTermService, postgres *db.Postgres, botSvc *services.BotService) {
	jobs, err := svc.GetSearchJobs(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get search jobs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d search jobs to run\n", len(jobs))

	marketplaceSvc := services.NewMarketplaceService(nil)

	for _, job := range jobs {
		if job.Marketplace == nil {
			continue
		}

		fmt.Printf("Searching %s on %s...\n", job.SearchTerm.Description, job.Marketplace.Name)

		ads, err := marketplaceSvc.FetchAdsFromURL(ctx, job.Marketplace.Name, job.SearchTerm.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch ads: %v\n", err)
			continue
		}

		saved := 0
		skipped := 0
		for _, ad := range ads {
			exists, err := postgres.ListingExistsByLink(ctx, ad.Link)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check listing: %v\n", err)
				continue
			}
			if exists {
				continue
			}

			product, err := botSvc.ValidateListing(ctx, ad)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to validate listing: %v\n", err)
				continue
			}
			if product == nil {
				skipped++
				continue
			}

			price := int(ad.Price)
			listing := &models.Listing{
				ProductID:     &product.ID,
				Link:          ad.Link,
				Price:         &price,
				Description:   ad.AdText,
				MarketplaceID: &job.Marketplace.ID,
				Status:        "draft",
			}

			if err := postgres.SaveListing(ctx, listing); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save listing: %v\n", err)
				continue
			}
			saved++
		}

		fmt.Printf("  Saved %d new listings, skipped %d\n", saved, skipped)
	}
}

func deactivateSearchTerm(ctx context.Context, svc *services.SearchTermService, id int64) {
	if id == 0 {
		fmt.Println("Error: id is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := svc.DeactivateSearchTerm(ctx, id); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to deactivate search term: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deactivated search term %d\n", id)
}
