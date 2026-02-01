package services

import (
	"context"
	"log"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

type BotService struct {
	cfg                *config.Config
	marketplaceService *MarketplaceService
	cacheService       *CacheService
	llmService         *LLMService
	valuationService   *ValuationService
	database           *db.Postgres
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

func (s *BotService) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Println("Starting Begbot...")

	queries := []string{"iphone 13", "iphone 14", "iphone 15"}

	for _, query := range queries {
		if err := s.processQuery(ctx, query); err != nil {
			log.Printf("Error processing query %s: %v", query, err)
		}
	}

	log.Println("Begbot finished.")
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
		log.Printf("Failed to extract product info: %v", err)
		return err
	}

	item.BuyShippingCost = int(productInfo.ShippingCost)

	validatedProduct, err := s.ValidateListing(ctx, ad)
	if err != nil {
		log.Printf("Failed to validate listing: %v", err)
		return err
	}

	if validatedProduct == nil {
		return nil
	}

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
		log.Printf("Failed to evaluate item: %v", err)
		return err
	}

	// Save listing for validated product
	productID := validatedProduct.ID
	price := item.BuyPrice
	marketplaceID := int64(1) // Blocket
	now := time.Now()

	listing := &models.Listing{
		ProductID:       &productID,
		Price:           &price,
		Link:            ad.Link,
		Description:     ad.AdText,
		MarketplaceID:   &marketplaceID,
		Status:          "active",
		PublicationDate: &now,
		IsMyListing:     false,
	}

	if err := s.database.SaveListing(ctx, listing); err != nil {
		log.Printf("Failed to save listing: %v", err)
		return err
	}
	log.Printf("Saved listing for %s at %d SEK", validatedProduct.Name, item.BuyPrice)

	if candidate.ShouldBuy {
		log.Printf("RECOMMENDATION: Buy %s for %d SEK (estimated profit: %d SEK)", item.SourceLink, candidate.TotalCost, candidate.EstimatedSell-candidate.TotalCost)
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
		log.Printf("Failed to extract product info: %v", err)
		return nil, err
	}

	if productInfo.Category == "" {
		log.Printf("No category detected for listing: %s", ad.Link)
		return nil, nil
	}

	product, err := s.database.FindProduct(ctx, productInfo.Manufacturer, productInfo.Model, productInfo.Category)
	if err != nil {
		log.Printf("Failed to find product: %v", err)
		return nil, err
	}

	if product == nil {
		log.Printf("Product not in catalog: %s %s (%s) - skipping", productInfo.Manufacturer, productInfo.Model, productInfo.Category)
		return nil, nil
	}

	if !product.Enabled {
		log.Printf("Product not enabled: %s %s (%s) - skipping", productInfo.Manufacturer, productInfo.Model, productInfo.Category)
		return nil, nil
	}

	return product, nil
}
