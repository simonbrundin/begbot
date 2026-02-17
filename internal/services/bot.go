package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

var botLogger *log.Logger

func init() {
	// Log to file
	f, err := os.OpenFile("/home/simon/repos/begbot/bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		botLogger = log.New(io.MultiWriter(os.Stdout, f), "", log.LstdFlags)
	} else {
		botLogger = log.New(os.Stdout, "", log.LstdFlags)
	}
}

type BotService struct {
	cfg                *config.Config
	marketplaceService *MarketplaceService
	cacheService       *CacheService
	llmService         *LLMService
	valuationService   *ValuationService
	database           *db.Postgres
	jobService         *JobService
	jobID              string
}

func NewBotService(cfg *config.Config, marketplaceService *MarketplaceService, cacheService *CacheService, llmService *LLMService, valuationService *ValuationService, database *db.Postgres) *BotService {
	return &BotService{
		cfg:                cfg,
		marketplaceService: marketplaceService,
		cacheService:       cacheService,
		llmService:         llmService,
		valuationService:   valuationService,
		database:           database,
	}
}

func NewBotServiceWithJob(cfg *config.Config, marketplaceService *MarketplaceService, cacheService *CacheService, llmService *LLMService, valuationService *ValuationService, database *db.Postgres, jobService *JobService, jobID string) *BotService {
	return &BotService{
		cfg:                cfg,
		marketplaceService: marketplaceService,
		cacheService:       cacheService,
		llmService:         llmService,
		valuationService:   valuationService,
		database:           database,
		jobService:         jobService,
		jobID:              jobID,
	}
}

func (s *BotService) log(level LogLevel, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", level, message)

	if s.jobService != nil && s.jobID != "" {
		s.jobService.AddLog(s.jobID, level, message)
	}
}

func (s *BotService) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s.log(LogLevelInfo, "=== STARTING BEGBOT ===")

	searchTerms, err := s.database.GetActiveSearchTerms(ctx)
	if err != nil {
		s.log(LogLevelError, "Error getting search terms: %v", err)
		return fmt.Errorf("failed to get search terms: %w", err)
	}

	s.log(LogLevelInfo, "Found %d search terms", len(searchTerms))

	if len(searchTerms) == 0 {
		s.log(LogLevelWarning, "No active search terms found")
		return nil
	}

	tradingRules, err := s.database.GetTradingRules(ctx)
	if err != nil {
		s.log(LogLevelWarning, "Failed to get trading rules: %v", err)
		tradingRules = &models.Economics{
			MinProfitSEK: intPtr(0),
			MinDiscount:  intPtr(0),
		}
	}
	s.log(LogLevelInfo, "Trading rules: min_profit_sek=%d, min_discount=%d", ptrVal(tradingRules.MinProfitSEK), ptrVal(tradingRules.MinDiscount))

	if s.jobService != nil && s.jobID != "" {
		s.jobService.StartJob(s.jobID)
		s.jobService.UpdateProgress(s.jobID, 0, len(searchTerms), "")
	}

	totalAdsFound := 0
	totalListingsSaved := 0
	for i, term := range searchTerms {
		// Check for cancellation before each search term
		if s.jobService != nil && s.jobID != "" {
			job := s.jobService.GetJob(s.jobID)
			if job != nil {
				select {
				case <-job.CancelChan:
					s.log(LogLevelInfo, "Job cancelled, stopping after %d/%d search terms", i, len(searchTerms))
					return nil
				default:
				}
			}
		}

		s.log(LogLevelInfo, "Processing search term %d/%d: %s", i+1, len(searchTerms), term.Description)

		adsList, err := s.marketplaceService.FetchAdsFromURL(ctx, s.getMarketplaceName(term.MarketplaceID), term.URL)
		if err != nil {
			s.log(LogLevelError, "Error fetching ads for %s: %v", term.Description, err)
			continue
		}
		s.log(LogLevelInfo, "Found %d ads for %s", len(adsList), term.Description)
		totalAdsFound += len(adsList)

		newAdsCount := 0
		for _, ad := range adsList {
			exists, err := s.database.ListingExistsByLink(ctx, ad.Link)
			if err != nil {
				s.log(LogLevelError, "Error checking listing exists: %v", err)
				continue
			}
			if exists {
				s.log(LogLevelInfo, "Skipping duplicate: %s", ad.Link)
				continue
			}
			newAdsCount++
			s.log(LogLevelInfo, "Processing new ad: %s (price: %.0f SEK)", ad.Link, ad.Price)
			if err := s.processAd(ctx, ad); err != nil {
				s.log(LogLevelError, "Error processing ad %s: %v", ad.Link, err)
			} else {
				totalListingsSaved++
			}
		}

		marketplaceName := s.getMarketplaceName(term.MarketplaceID)
		history := &models.SearchHistory{
			SearchTermID:    term.ID,
			SearchTermDesc:  term.Description,
			URL:             term.URL,
			ResultsFound:    len(adsList),
			NewAdsFound:     newAdsCount,
			MarketplaceID:   term.MarketplaceID,
			MarketplaceName: marketplaceName,
			SearchedAt:      time.Now(),
		}
		if err := s.database.SaveSearchHistory(ctx, history); err != nil {
			s.log(LogLevelWarning, "Failed to save search history: %v", err)
		}

		if s.jobService != nil && s.jobID != "" {
			s.jobService.UpdateProgress(s.jobID, i+1, len(searchTerms), term.Description)
		}
	}

	if s.jobService != nil && s.jobID != "" {
		// Only complete if not already cancelled
		job := s.jobService.GetJob(s.jobID)
		if job != nil && job.Status != JobStatusCancelled {
			s.jobService.CompleteJob(s.jobID, totalAdsFound)
		}
	}

	s.log(LogLevelInfo, "=== BEGBOT FINISHED: Total ads found: %d, Listings saved: %d ===", totalAdsFound, totalListingsSaved)
	return nil
}

func (s *BotService) processQuery(ctx context.Context, query string) error {
	log.Printf("Processing query: %s", query)

	ads, err := s.marketplaceService.FetchAds(ctx, query)
	if err != nil {
		return err
	}

	var links []string
	for _, ad := range ads {
		links = append(links, ad.Link)
	}

	newLinks, cachedLinks := s.cacheService.Filter(ctx, links)
	log.Printf("Found %d new ads, %d cached", len(newLinks), len(cachedLinks))

	for _, ad := range ads {
		if !s.isNewLink(ad.Link, newLinks) {
			continue
		}

		if err := s.processAd(ctx, ad); err != nil {
			log.Printf("Error processing ad %s: %v", ad.Link, err)
		}
	}

	return nil
}

func (s *BotService) getMarketplaceName(marketplaceID *int64) string {
	if marketplaceID == nil {
		return "blocket"
	}
	switch *marketplaceID {
	case 1:
		return "blocket"
	case 2:
		return "tradera"
	default:
		return "blocket"
	}
}

func intPtr(i int) *int {
	return &i
}

func ptrVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func (s *BotService) isNewLink(link string, newLinks []string) bool {
	for _, l := range newLinks {
		if l == link {
			return true
		}
	}
	return false
}

func (s *BotService) processAd(ctx context.Context, ad RawAd) error {
	item := s.marketplaceService.ConvertToPotentialItem(ad)

	productInfo, err := s.llmService.ExtractProductInfo(ctx, ad.AdText, ad.Link)
	if err != nil {
		s.log(LogLevelError, "Failed to extract product info: %v", err)
		return err
	}

	item.BuyShippingCost = int(productInfo.ShippingCost)

	validatedProduct, err := s.ValidateListing(ctx, ad)
	if err != nil {
		s.log(LogLevelError, "Failed to validate listing: %v", err)
		return err
	}

	if validatedProduct == nil {
		return nil
	}

	s.log(LogLevelInfo, "Product identified: %s %s (%s)", productInfo.Manufacturer, productInfo.Model, productInfo.Category)

	item.ProductID = &validatedProduct.ID
	if item.SellPackagingCost == nil {
		packagingCost := validatedProduct.SellPackagingCost
		item.SellPackagingCost = &packagingCost
	}
	if item.SellPostageCost == nil {
		postageCost := validatedProduct.SellPostageCost
		item.SellPostageCost = &postageCost
	}

	candidate, err := s.evaluateItem(ctx, item, productInfo)
	if err != nil {
		s.log(LogLevelError, "Failed to evaluate item: %v", err)
		return err
	}

	// Save listing for validated product
	productID := validatedProduct.ID
	price := item.BuyPrice
	marketplaceID := int64(1) // Blocket
	now := time.Now()

	// Collect all valuations from different methods
	valInputs, err := s.valuationService.CollectAll(ctx, strconv.FormatInt(productID, 10), *productInfo)
	if err != nil {
		s.log(LogLevelWarning, "Failed to collect valuations: %v", err)
		valInputs = nil
	}

	// Compile valuations into a final recommendation
	var compiledValuation int
	if len(valInputs) > 0 {
		output, err := s.valuationService.Compile(ctx, valInputs)
		if err != nil {
			s.log(LogLevelWarning, "Failed to compile valuations: %v", err)
			compiledValuation = candidate.EstimatedSell
		} else {
			compiledValuation = int(output.RecommendedPrice)
		}
	} else {
		compiledValuation = candidate.EstimatedSell
	}

	listing := &models.Listing{
		ProductID:       &productID,
		Price:           &price,
		Valuation:       compiledValuation,
		Link:            ad.Link,
		Title:           ad.Title,
		Description:     &ad.AdText,
		MarketplaceID:   &marketplaceID,
		Status:          "active",
		PublicationDate: &now,
		IsMyListing:     false,
	}

	if err := s.database.SaveListing(ctx, listing); err != nil {
		s.log(LogLevelError, "Failed to save listing: %v", err)
		return err
	}
	s.log(LogLevelInfo, "Saved listing for %s at %d SEK (valuation: %d SEK)", validatedProduct.Name, item.BuyPrice, compiledValuation)

	// Save individual valuations to the database
	if len(valInputs) > 0 {
		productIDStr := fmt.Sprintf("%d", productID)
		if err := s.valuationService.SaveValuationsWithListingID(ctx, productIDStr, valInputs, &listing.ID); err != nil {
			s.log(LogLevelWarning, "Failed to save valuations: %v", err)
		}
	}

	if candidate.ShouldBuy {
		s.log(LogLevelInfo, "RECOMMENDATION: Buy %s for %d SEK (profit: %d SEK)", item.SourceLink, candidate.TotalCost, candidate.EstimatedSell-candidate.TotalCost)
	}

	return nil
}

func (s *BotService) evaluateItem(ctx context.Context, item *models.TradedItem, productInfo *ProductInfo) (*models.TradedItemCandidate, error) {
	historicalValuation, err := s.valuationService.GetHistoricalValuation(ctx, "")
	if err != nil {
		return nil, err
	}

	estimatedSellPrice := s.valuationService.CalculatePriceForDays(s.cfg.Valuation.TargetSellDays, historicalValuation)

	totalCost := item.BuyPrice + item.BuyShippingCost
	estimatedProfit := int(estimatedSellPrice) - totalCost
	profitMargin := float64(estimatedProfit) / float64(totalCost)
	shouldBuy := s.valuationService.ShouldBuy(profitMargin)

	return &models.TradedItemCandidate{
		Item:          *item,
		EstimatedSell: int(estimatedSellPrice),
		ShippingCost:  item.BuyShippingCost,
		TotalCost:     totalCost,
		ShouldBuy:     shouldBuy,
	}, nil
}

func (s *BotService) ValidateListing(ctx context.Context, ad RawAd) (*models.Product, error) {
	productInfo, err := s.llmService.ExtractProductInfo(ctx, ad.AdText, ad.Link)
	if err != nil {
		s.log(LogLevelError, "Failed to extract product info: %v", err)
		return nil, err
	}

	if productInfo.Category == "" {
		s.log(LogLevelWarning, "No category detected for listing: %s", ad.Link)
		return nil, nil
	}

	product, err := s.database.FindProduct(ctx, productInfo.Manufacturer, productInfo.Model, productInfo.Category)
	if err != nil {
		s.log(LogLevelError, "Failed to find product: %v", err)
		return nil, err
	}

	if product == nil {
		s.log(LogLevelInfo, "Product not in catalog: %s %s (%s) - skipping", productInfo.Manufacturer, productInfo.Model, productInfo.Category)
		return nil, nil
	}

	if !product.Enabled {
		s.log(LogLevelWarning, "Product not enabled: %s %s (%s) - skipping", productInfo.Manufacturer, productInfo.Model, productInfo.Category)
		return nil, nil
	}

	return product, nil
}
