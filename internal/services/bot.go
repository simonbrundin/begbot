package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

var botLogger *log.Logger

var damageKeywords = []string{
	"spricka", "sprucken", "spruckit",
	"trasig", "trasigt",
	"buckla", "bucklat",
	"skärmfel", "skärm sprucken", "skärm trasig",
	"vattenskada", "vattenskadad",
	"oxidation", "oxiderad",
	"fungerar inte", "fungerar ej", "fungerar inte",
	"defekt", "fel på",
	"repa", "repor",
	"intryckt", "intryckt hörn",
}

func hasDamageInReasoning(reasoning string) bool {
	reasoningLower := strings.ToLower(reasoning)
	for _, keyword := range damageKeywords {
		if strings.Contains(reasoningLower, keyword) {
			return true
		}
	}
	return false
}

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
	cfg                 *config.Config
	marketplaceService  *MarketplaceService
	cacheService        *CacheService
	llmService          *LLMService
	valuationService    *ValuationService
	database            *db.Postgres
	jobService          *JobService
	jobID               string
	scrapingRunID       int64
	currentSearchTermID int64
	searchTermsOverride []models.SearchTerm
	newProductsCount    int
	emailedAdsCount     int
	logBuffer           []*models.ScrapingRunLog
}

type ProcessAdResult struct {
	Saved      bool
	NewProduct bool
	EmailSent  bool
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

func (s *BotService) SetSearchTermsOverride(terms []models.SearchTerm) {
	s.searchTermsOverride = terms
}

func (s *BotService) ValuationService() *ValuationService {
	return s.valuationService
}

func (s *BotService) log(level LogLevel, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] %s", level, message)

	if s.jobService != nil && s.jobID != "" {
		s.jobService.AddLog(s.jobID, level, message)
	}

	if s.scrapingRunID > 0 && (level == LogLevelError || level == LogLevelWarning || shouldSaveInfoLog(message)) {
		s.logBuffer = append(s.logBuffer, &models.ScrapingRunLog{
			ScrapingRunID: s.scrapingRunID,
			Level:         string(level),
			Message:       message,
			CreatedAt:     time.Now(),
		})
	}
}

func shouldSaveInfoLog(message string) bool {
	importantPhrases := []string{
		"=== STARTING",
		"=== COMPLETED",
		"Found",
		"ads for",
		"Processing search term",
		"Processing new ad",
		"Skipping duplicate",
		"Email sent",
		"Listing saved",
		"Product created",
		"Search term",
		"Trading rules",
	}
	for _, phrase := range importantPhrases {
		if strings.Contains(message, phrase) {
			return true
		}
	}
	return false
}

func (s *BotService) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	s.log(LogLevelInfo, "=== STARTING BEGBOT ===")

	scrapingRun := &models.ScrapingRun{
		StartedAt: time.Now(),
		Status:    "running",
	}
	if err := s.database.SaveScrapingRun(ctx, scrapingRun); err != nil {
		s.log(LogLevelWarning, "Failed to create scraping run: %v", err)
	} else {
		s.scrapingRunID = scrapingRun.ID
	}

	runStatus := "completed"
	var errorMessage *string

	totalAdsFound := 0
	totalListingsSaved := 0
	totalNewAds := 0

	defer func() {
		if s.scrapingRunID > 0 {
			now := time.Now()
			run := &models.ScrapingRun{
				ID:                 s.scrapingRunID,
				CompletedAt:        &now,
				Status:             runStatus,
				TotalAdsFound:      totalAdsFound,
				NewAds:             totalNewAds,
				NewProducts:        s.newProductsCount,
				SavedProducts:      s.newProductsCount,
				TotalListingsSaved: totalListingsSaved,
				EmailedAds:         s.emailedAdsCount,
				ErrorMessage:       errorMessage,
			}

			finalCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var updErr error
			for attempt := 1; attempt <= 2; attempt++ {
				updErr = s.database.UpdateScrapingRun(finalCtx, run)
				if updErr == nil {
					break
				}
				s.log(LogLevelWarning, "Attempt %d: Failed to update scraping run: %v", attempt, updErr)
				time.Sleep(500 * time.Millisecond)
			}
			if updErr != nil {
				s.log(LogLevelWarning, "Failed to update scraping run after retries: %v", updErr)
			} else {
				s.log(LogLevelInfo, "Updated scraping run %d with status: %s", s.scrapingRunID, runStatus)
			}

			if len(s.logBuffer) > 0 {
				for attempt := 1; attempt <= 2; attempt++ {
					updErr = s.database.SaveScrapingRunLogs(finalCtx, s.logBuffer)
					if updErr == nil {
						break
					}
					s.log(LogLevelWarning, "Attempt %d: Failed to save scraping run logs: %v", attempt, updErr)
					time.Sleep(500 * time.Millisecond)
				}
			}

			beforeDate := now.AddDate(0, 0, -30)
			deleted, delErr := s.database.DeleteOldScrapingRunLogs(finalCtx, beforeDate)
			if delErr != nil {
				s.log(LogLevelWarning, "Failed to delete old scraping run logs: %v", delErr)
			} else if deleted > 0 {
				s.log(LogLevelInfo, "Deleted %d old scraping run logs", deleted)
			}
		}
	}()

	searchTerms, err := s.database.GetActiveSearchTerms(ctx)
	if err != nil {
		s.log(LogLevelError, "Error getting search terms: %v", err)
		errMsg := fmt.Sprintf("failed to get search terms: %v", err)
		errorMessage = &errMsg
		runStatus = "failed"
		return err
	}

	if len(s.searchTermsOverride) > 0 {
		searchTerms = s.searchTermsOverride
	}

	s.log(LogLevelInfo, "Found %d search terms", len(searchTerms))

	if len(searchTerms) == 0 {
		s.log(LogLevelWarning, "No active search terms found")
		errMsg := "no active search terms"
		errorMessage = &errMsg
		runStatus = "completed"
		return nil
	}

	emailSettings, err := s.database.GetEmailSettings(ctx)
	if err != nil {
		s.log(LogLevelWarning, "Failed to get email settings: %v", err)
		defaultProfit := 200
		defaultDiscount := 15
		trueVal := true
		emailSettings = &models.EmailSettings{
			IsActive:     &trueVal,
			MinProfitSEK: &defaultProfit,
			MinDiscount:  &defaultDiscount,
		}
	}
	s.log(LogLevelInfo, "Email settings: is_active=%v, only_enabled_products=%v, min_profit_sek=%d, min_discount=%d", ptrBoolVal(emailSettings.IsActive), ptrBoolVal(emailSettings.OnlyEnabledProducts), ptrVal(emailSettings.MinProfitSEK), ptrVal(emailSettings.MinDiscount))

	if s.jobService != nil && s.jobID != "" {
		s.jobService.StartJob(s.jobID)
		s.jobService.UpdateProgress(s.jobID, 0, len(searchTerms), "")
	}

	totalAdsFound = 0
	totalListingsSaved = 0
	totalNewAds = 0
	s.newProductsCount = 0
	s.emailedAdsCount = 0
	for i, term := range searchTerms {
		// Check for cancellation before each search term
		if s.jobService != nil && s.jobID != "" {
			job := s.jobService.GetJob(s.jobID)
			if job != nil {
				select {
				case <-job.CancelChan:
					s.log(LogLevelInfo, "Job cancelled, stopping after %d/%d search terms", i, len(searchTerms))
					errMsg := "job cancelled by user"
					errorMessage = &errMsg
					runStatus = "cancelled"
					return nil
				default:
				}
			}
		}

		s.log(LogLevelInfo, "Processing search term %d/%d: %s", i+1, len(searchTerms), term.Description)
		s.currentSearchTermID = term.ID

		adsList, err := s.marketplaceService.FetchAdsFromURL(ctx, s.getMarketplaceName(term.MarketplaceID), term.URL)
		if err != nil {
			s.log(LogLevelError, "Error fetching ads for %s: %v", term.Description, err)
			continue
		}
		s.log(LogLevelInfo, "Found %d ads for %s", len(adsList), term.Description)
		totalAdsFound += len(adsList)

		newAdsCount := 0
		for _, ad := range adsList {
			adCtx, adCancel := context.WithTimeout(ctx, 60*time.Second)
			exists, err := s.database.ListingExistsByLink(adCtx, ad.Link)
			if err != nil {
				s.log(LogLevelError, "Error checking listing exists: %v", err)
				adCancel()
				continue
			}
			if exists {
				s.log(LogLevelInfo, "Skipping duplicate: %s", ad.Link)
				adCancel()
				continue
			}
			newAdsCount++
			s.log(LogLevelInfo, "Processing new ad: %s (price: %.0f SEK)", ad.Link, ad.Price)
			result, err := s.processAd(adCtx, ad)
			adCancel()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					s.log(LogLevelWarning, "Timed out processing ad: %s - skipping and continuing", ad.Link)
				} else {
					s.log(LogLevelError, "Error processing ad %s: %v", ad.Link, err)
				}
			} else if result != nil {
				totalListingsSaved++
				if result.NewProduct {
					s.newProductsCount++
				}
				if result.EmailSent {
					s.emailedAdsCount++
				}
			}
		}

		totalNewAds += newAdsCount

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

	s.log(LogLevelInfo, "=== BEGBOT FINISHED: Total ads found: %d, New ads: %d, New products: %d, Listings saved: %d, Emailed: %d ===", totalAdsFound, totalNewAds, s.newProductsCount, totalListingsSaved, s.emailedAdsCount)
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

		_, err := s.processAd(ctx, ad)
		if err != nil {
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

func ptrBoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ptrBool(b bool) *bool {
	return &b
}

func (s *BotService) isNewLink(link string, newLinks []string) bool {
	for _, l := range newLinks {
		if l == link {
			return true
		}
	}
	return false
}

func (s *BotService) processAd(ctx context.Context, ad RawAd) (*ProcessAdResult, error) {
	result := &ProcessAdResult{}
	item := s.marketplaceService.ConvertToPotentialItem(ad)

	// Log shipping info from Blocket API
	if ad.ShippingCost != nil || ad.ShippingInsurance != nil {
		s.log(LogLevelInfo, "Blocket API shipping: cost=%v, insurance=%v", ad.ShippingCost, ad.ShippingInsurance)
	}

	productInfo, err := s.llmService.ExtractProductInfo(ctx, ad.Title, ad.AdText, ad.Link)
	if err != nil {
		s.log(LogLevelError, "Failed to extract product info: %v", err)
		return result, err
	}

	s.log(LogLevelInfo, "LLM extracted: Manufacturer=%q, Model=%q, Category=%q, Storage=%q",
		productInfo.Manufacturer, productInfo.Model, productInfo.Category, productInfo.Storage)

	// Only use LLM's shipping cost if Blocket API didn't provide one
	if item.BuyShippingCost == 0 && productInfo.ShippingCost > 0 {
		item.BuyShippingCost = int(productInfo.ShippingCost)
		s.log(LogLevelInfo, "Using LLM shipping cost: %.0f", productInfo.ShippingCost)
	}

	// Log final shipping cost
	s.log(LogLevelInfo, "Final shipping: cost=%d, insurance=%d", item.BuyShippingCost, item.BuyShippingInsurance)

	validationResult, err := s.ValidateListing(ctx, ad)
	if err != nil {
		s.log(LogLevelError, "Failed to validate listing: %v", err)
		return result, err
	}

	if validationResult == nil {
		s.log(LogLevelWarning, "Listing validation failed - no matching product found for: %s", ad.Link)
		return result, nil
	}

	intactResult, err := s.llmService.EvaluateProductIntact(ctx, ad.AdText, ad.Title)
	s.log(LogLevelInfo, "DEBUG: intactResult=%v, err=%v", intactResult, err)
	var isIntact *bool
	var intactReasoning *string
	if err != nil {
		s.log(LogLevelWarning, "Failed to evaluate product intactness: %v - accepting product", err)
	} else if hasDamageInReasoning(intactResult.Reasoning) {
		s.log(LogLevelWarning, "Product has damage mentioned in reasoning - skipping listing: %s, reasoning: %s",
			ad.Link, intactResult.Reasoning)
		return result, nil
	} else if len(intactResult.IssuesFound) > 0 {
		s.log(LogLevelWarning, "Product has issues - skipping listing: %s, issues: %v, reasoning: %s",
			ad.Link, intactResult.IssuesFound, intactResult.Reasoning)
		return result, nil
	} else {
		s.log(LogLevelInfo, "Product is intact - no issues found, continuing...")
		intactVal := true
		isIntact = &intactVal
		if intactResult.Reasoning != "" {
			intactReasoning = &intactResult.Reasoning
		}
	}

	if validationResult.IsNewProduct {
		s.log(LogLevelInfo, "New product detected: %s %s (%s) - collecting valuations before creation",
			validationResult.ProductInfo.Manufacturer, validationResult.ProductInfo.Model, validationResult.ProductInfo.Category)

		valInputs, err := s.valuationService.CollectAll(ctx, "", *validationResult.ProductInfo)
		if err != nil {
			s.log(LogLevelWarning, "Failed to collect valuations for new product: %v", err)
		}

		if len(valInputs) == 0 {
			s.log(LogLevelWarning, "No valuations collected for new product - skipping listing: %s", ad.Link)
			return result, nil
		}

		output, err := s.valuationService.Compile(ctx, valInputs)
		if err != nil {
			s.log(LogLevelWarning, "Failed to compile valuations for new product: %v", err)
			s.log(LogLevelWarning, "Skipping listing due to valuation failure: %s", ad.Link)
			return result, nil
		}

		product := &models.Product{
			Brand:    &validationResult.ProductInfo.Manufacturer,
			Name:     &validationResult.ProductInfo.Model,
			Category: &validationResult.ProductInfo.Category,
		}
		if err := s.database.CreateProduct(ctx, product); err != nil {
			s.log(LogLevelError, "Failed to create product: %v", err)
			return result, err
		}

		// Save valuations BEFORE checking save/auto-enable criteria
		if err := s.valuationService.SaveValuations(ctx, fmt.Sprintf("%d", product.ID), valInputs); err != nil {
			s.log(LogLevelWarning, "Failed to save valuations for new product: %v", err)
		}

		if err := s.saveBlocketCategoryFromValuations(ctx, product, valInputs); err != nil {
			s.log(LogLevelWarning, "Failed to save Blocket category for product %d: %v", product.ID, err)
		}

		// Now check if product should be saved
		shouldSave, checkErr := s.database.ShouldSaveProduct(ctx, product.ID)
		if checkErr != nil {
			s.log(LogLevelWarning, "Failed to check save criteria for new product %d: %v", product.ID, checkErr)
		}

		if !shouldSave {
			s.log(LogLevelWarning, "New product %s %s does not meet save criteria - deleting product and skipping listing",
				validationResult.ProductInfo.Manufacturer, validationResult.ProductInfo.Model)
			if delErr := s.database.DeleteProduct(ctx, product.ID); delErr != nil {
				s.log(LogLevelWarning, "Failed to delete product: %v", delErr)
			}
			return result, nil
		}

		// Check if product should be auto-enabled
		shouldEnable, checkErr := s.database.ShouldAutoEnableProduct(ctx, product.ID)
		if checkErr != nil {
			s.log(LogLevelWarning, "Failed to check auto-enable for new product %d: %v", product.ID, checkErr)
		}

		if shouldEnable {
			if err := s.database.SetProductEnabled(ctx, product.ID, true); err != nil {
				s.log(LogLevelWarning, "Failed to enable new product %d: %v", product.ID, err)
			} else {
				s.log(LogLevelInfo, "Auto-enabled new product: ID=%d (passed auto-enable criteria)", product.ID)
			}
		}

		validationResult.Product = product
		validationResult.ProductInfo.NewPrice = output.RecommendedPrice
		s.log(LogLevelInfo, "Created enabled product: ID=%d, Name=%s", product.ID, *product.Name)
	}

	s.log(LogLevelInfo, "Product identified: %s %s (%s)", validationResult.ProductInfo.Manufacturer, validationResult.ProductInfo.Model, validationResult.ProductInfo.Category)

	item.ProductID = &validationResult.Product.ID
	if item.SellPackagingCost == nil {
		packagingCost := validationResult.Product.SellPackagingCost
		item.SellPackagingCost = &packagingCost
	}
	if item.SellPostageCost == nil {
		postageCost := validationResult.Product.SellPostageCost
		item.SellPostageCost = &postageCost
	}

	candidate, err := s.evaluateItem(ctx, item, validationResult.ProductInfo)
	if err != nil {
		s.log(LogLevelError, "Failed to evaluate item: %v", err)
		return result, err
	}

	productID := validationResult.Product.ID
	price := item.BuyPrice
	marketplaceID := int64(1)
	now := time.Now()

	valInputs, err := s.valuationService.CollectAll(ctx, strconv.FormatInt(productID, 10), *validationResult.ProductInfo)
	if err != nil {
		s.log(LogLevelWarning, "Failed to collect valuations: %v", err)
		valInputs = nil
	}

	var compiledValuation int
	if len(valInputs) > 0 {
		output, err := s.valuationService.Compile(ctx, valInputs)
		if err != nil {
			s.log(LogLevelWarning, "Failed to compile valuations: %v", err)
			compiledValuation = candidate.EstimatedSell
		} else {
			compiledValuation = int(output.RecommendedPrice)
		}

		if err := s.saveBlocketCategoryFromValuations(ctx, validationResult.Product, valInputs); err != nil {
			s.log(LogLevelWarning, "Failed to save Blocket category for product %d: %v", productID, err)
		}
	} else {
		compiledValuation = candidate.EstimatedSell
	}

	listing := &models.Listing{
		ProductID:            &productID,
		Price:                &price,
		Link:                 ad.Link,
		Title:                ad.Title,
		Description:          &ad.AdText,
		MarketplaceID:        &marketplaceID,
		Status:               "active",
		PublicationDate:      &now,
		IsMyListing:          false,
		ShippingCost:         func() *int { v := int(item.BuyShippingCost); return &v }(),
		ShippingInsurance:    func() *int { v := item.BuyShippingInsurance; return &v }(),
		IsIntact:             isIntact,
		IntactCheckReasoning: intactReasoning,
	}

	if err := s.database.SaveListing(ctx, listing); err != nil {
		s.log(LogLevelError, "Failed to save listing: %v", err)
		return result, err
	}

	if len(ad.ImageURLs) > 0 {
		if err := s.database.SaveImageLinks(ctx, listing.ID, ad.ImageURLs); err != nil {
			s.log(LogLevelWarning, "Failed to save image links: %v", err)
		}
	}

	s.log(LogLevelInfo, "Saved listing for %s at %d SEK (valuation: %d SEK)", *validationResult.Product.Name, item.BuyPrice, compiledValuation)

	if len(valInputs) > 0 {
		productIDStr := fmt.Sprintf("%d", productID)
		if err := s.valuationService.SaveValuations(ctx, productIDStr, valInputs); err != nil {
			s.log(LogLevelWarning, "Failed to save valuations: %v", err)
		}
	}

	err = s.SendTradingRuleEmail(ctx, listing, validationResult.Product, s.currentSearchTermID)
	if err != nil {
		s.log(LogLevelWarning, "Failed to send trading rule email: %v", err)
	} else {
		result.EmailSent = true
	}

	if candidate.ShouldBuy {
		s.log(LogLevelInfo, "RECOMMENDATION: Buy %s for %d SEK (profit: %d SEK)", item.SourceLink, candidate.TotalCost, candidate.EstimatedSell-candidate.TotalCost)
	}

	result.Saved = true
	if validationResult.IsNewProduct {
		result.NewProduct = true
	}

	return result, nil
}

func (s *BotService) saveBlocketCategoryFromValuations(ctx context.Context, product *models.Product, valInputs []ValuationInput) error {
	for _, input := range valInputs {
		if input.Type == "Blocket" && input.Category != "" {
			product.BlocketCategory = &input.Category
			if err := s.database.UpdateProduct(ctx, product); err != nil {
				return err
			}
			s.log(LogLevelInfo, "Saved Blocket category %s for product ID=%d", input.Category, product.ID)
			break
		}
	}
	return nil
}

func (s *BotService) evaluateItem(ctx context.Context, item *models.TradedItem, productInfo *ProductInfo) (*models.TradedItemCandidate, error) {
	historicalValuation, err := s.valuationService.GetHistoricalValuation(ctx, "")
	if err != nil {
		return nil, err
	}

	estimatedSellPrice := s.valuationService.CalculatePriceForDays(s.cfg.Valuation.TargetSellDays, historicalValuation)

	totalCost := item.BuyPrice + item.BuyShippingCost + item.BuyShippingInsurance
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

type ValidateListingResult struct {
	Product      *models.Product
	IsNewProduct bool
	ProductInfo  *ProductInfo
}

func (s *BotService) ValidateListing(ctx context.Context, ad RawAd) (*ValidateListingResult, error) {
	productInfo, err := s.llmService.ExtractProductInfo(ctx, ad.Title, ad.AdText, ad.Link)
	if err != nil {
		s.log(LogLevelError, "Failed to extract product info: %v", err)
		return nil, err
	}

	if productInfo.Category == "" {
		s.log(LogLevelWarning, "No category detected for listing: %s", ad.Link)
		return nil, nil
	}

	s.log(LogLevelInfo, "Looking up product: brand=%q, name=%q, category=%q",
		productInfo.Manufacturer, productInfo.Model, productInfo.Category)

	products, err := s.database.FindProducts(ctx, productInfo.Manufacturer, productInfo.Model, productInfo.Category)
	if err != nil {
		s.log(LogLevelError, "Failed to find products: %v", err)
		return nil, err
	}

	if len(products) == 0 {
		s.log(LogLevelWarning, "Product NOT found in catalog: %s %s (%s) - caller will handle creation with valuations",
			productInfo.Manufacturer, productInfo.Model, productInfo.Category)
		return &ValidateListingResult{
			Product:      nil,
			IsNewProduct: true,
			ProductInfo:  productInfo,
		}, nil
	}

	s.log(LogLevelInfo, "Found %d candidate product(s) in catalog:", len(products))
	for i, p := range products {
		enabled := "disabled"
		if p.Enabled != nil && *p.Enabled {
			enabled = "enabled"
		}
		s.log(LogLevelInfo, "  [%d] ID=%d, Name=%s, Category=%s, Status=%s",
			i+1, p.ID, *p.Name, *p.Category, enabled)
	}

	var product *models.Product
	for _, p := range products {
		if p.Enabled != nil && *p.Enabled {
			product = p
			break
		}
	}

	if product == nil {
		s.log(LogLevelWarning, "All candidates are disabled - using first disabled product: %s %s",
			productInfo.Manufacturer, productInfo.Model)
		product = products[0]
	}

	s.log(LogLevelInfo, "Product selected: ID=%d, Name=%s, Category=%s",
		product.ID, *product.Name, *product.Category)

	return &ValidateListingResult{
		Product:      product,
		IsNewProduct: false,
		ProductInfo:  productInfo,
	}, nil
}

func (s *BotService) SendTradingRuleEmail(ctx context.Context, listing *models.Listing, product *models.Product, searchTermID int64) error {
	var emailSettings *models.EmailSettings
	var err error

	if s.database != nil {
		emailSettings, err = s.database.GetEmailSettings(ctx)
		if err != nil {
			s.log(LogLevelWarning, "Failed to get email settings: %v", err)
		}
	}

	if emailSettings == nil {
		defaultProfit := 200
		defaultDiscount := 15
		trueVal := true
		emailSettings = &models.EmailSettings{
			IsActive:            &trueVal,
			OnlyEnabledProducts: &trueVal,
			MinProfitSEK:        &defaultProfit,
			MinDiscount:         &defaultDiscount,
		}
	}

	// Skip if email is globally disabled
	if emailSettings.IsActive != nil && !*emailSettings.IsActive {
		s.log(LogLevelInfo, "Skipping email: email is disabled in settings")
		return nil
	}

	// Skip if only_enabled_products is true and product is disabled
	onlyEnabledProducts := true
	if emailSettings.OnlyEnabledProducts != nil {
		onlyEnabledProducts = *emailSettings.OnlyEnabledProducts
	}
	if onlyEnabledProducts && product != nil && product.Enabled != nil && !*product.Enabled {
		s.log(LogLevelInfo, "Skipping email: product %s is disabled and only_enabled_products is true", *product.Name)
		return nil
	}

	minProfitSEK := 0
	if emailSettings.MinProfitSEK != nil {
		minProfitSEK = *emailSettings.MinProfitSEK
	}

	minDiscount := 0
	if emailSettings.MinDiscount != nil {
		minDiscount = *emailSettings.MinDiscount
	}

	// Use computed product-level valuation with confidence
	computedValuation := listing.Valuation
	confidence := 0.0
	if listing.ProductID != nil && s.database != nil {
		if cv, conf, cvErr := s.database.ComputeWeightedValuationForProduct(ctx, *listing.ProductID); cvErr == nil && cv > 0 {
			computedValuation = cv
			confidence = conf
		}
	}

	// Get trading rules for min_confidence check
	minConfidence := 80
	if s.database != nil {
		if tradingRules, err := s.database.GetTradingRules(ctx); err == nil && tradingRules != nil && tradingRules.MinConfidence != nil {
			minConfidence = *tradingRules.MinConfidence
		}
	}

	// Skip if confidence is below minimum threshold
	if confidence > 0 && confidence < float64(minConfidence) {
		s.log(LogLevelInfo, "Listing does not pass email settings: confidence=%.2f (<%d%%)",
			confidence, minConfidence)
		return nil
	}

	profit := computedValuation - *listing.Price
	if listing.ShippingCost != nil {
		profit -= *listing.ShippingCost
	}
	if listing.ShippingInsurance != nil {
		profit -= *listing.ShippingInsurance
	}
	discountPercent := float64(profit) / float64(computedValuation) * 100

	if profit <= minProfitSEK || discountPercent <= float64(minDiscount) {
		s.log(LogLevelInfo, "Listing does not pass email settings: profit=%d (>%d), discount=%.2f%% (>%d%%)",
			profit, minProfitSEK, discountPercent, minDiscount)
		return nil
	}

	go func() {
		emailCfg := EmailConfig{
			SMTPHost:     s.cfg.Email.SMTPHost,
			SMTPPort:     s.cfg.Email.SMTPPort,
			SMTPUsername: s.cfg.Email.SMTPUsername,
			SMTPPassword: s.cfg.Email.SMTPPassword,
			From:         s.cfg.Email.From,
			Recipients:   s.cfg.Email.Recipients,
		}

		subject := "Ny annons som passar dina trading rules - " + listing.Title

		// Prepare template data
		priceStr := ""
		if listing.Price != nil {
			priceStr = fmt.Sprintf("%d kr", *listing.Price)
		}
		emailProfit := computedValuation
		if listing.Price != nil {
			emailProfit = computedValuation - *listing.Price
		}
		if listing.ShippingCost != nil {
			emailProfit -= *listing.ShippingCost
		}
		if listing.ShippingInsurance != nil {
			emailProfit -= *listing.ShippingInsurance
		}
		profitStr := fmt.Sprintf("%d kr", emailProfit)
		discountStr := fmt.Sprintf("%.0f%%", discountPercent)

		// Shipping cost
		shippingCostStr := "0 kr"
		if listing.ShippingCost != nil {
			shippingCostStr = fmt.Sprintf("%d kr", *listing.ShippingCost)
		}

		desc := ""
		if listing.Description != nil {
			desc = *listing.Description
		}

		brand := ""
		name := ""
		if product != nil {
			if product.Brand != nil {
				brand = *product.Brand
			}
			if product.Name != nil {
				name = *product.Name
			}
		}

		// Full product name (brand + name)
		productName := name
		if brand != "" && name != "" {
			productName = brand + " " + name
		} else if brand != "" {
			productName = brand
		}

		// Fetch image URLs from DB (can be multiple)
		var imageURLs []string
		if s.database != nil {
			if listing.ID != 0 {
				if imgs, err := s.database.GetImageLinks(ctx, listing.ID); err == nil {
					imageURLs = imgs
				}
			}
		}

		if len(imageURLs) == 0 {
			imageURLs = []string{""}
		}

		mailData := map[string]interface{}{
			"Title":           listing.Title,
			"Price":           priceStr,
			"Valuation":       fmt.Sprintf("%d kr", computedValuation),
			"Profit":          profitStr,
			"Discount":        discountStr,
			"SecurityPercent": "N/A",
			"ShippingCost":    shippingCostStr,
			"Product":         productName,
			"Description":     desc,
			"ImageURLs":       imageURLs,
			"Link":            listing.Link,
		}

		err := SendMailHTMLWithData(emailCfg, s.cfg.Email.Recipients, subject, "mail.html", mailData)
		if err != nil {
			s.log(LogLevelWarning, "Failed to send trading rule email: %v", err)
		} else {
			s.log(LogLevelInfo, "Sent trading rule email for listing: %s", listing.Link)

			// Log the sent email to database
			if s.database != nil {
				var marketplaceID *int16
				if listing.MarketplaceID != nil {
					mid := int16(*listing.MarketplaceID)
					marketplaceID = &mid
				}

				// Get confidence for the product
				var confidence *float64
				if listing.ProductID != nil {
					_, conf, err := s.database.ComputeWeightedValuationForProduct(context.Background(), *listing.ProductID)
					if err == nil && conf > 0 {
						confidence = &conf
					}
				}

				sentEmail := &models.SentEmail{
					ListingID:        &listing.ID,
					ListingTitle:     listing.Title,
					ListingLink:      listing.Link,
					ListingPrice:     listing.Price,
					ListingValuation: &computedValuation,
					Profit:           profit,
					DiscountPercent:  discountPercent,
					ProductID:        listing.ProductID,
					ProductName:      &productName,
					Brand:            product.Brand,
					Confidence:       confidence,
					ScrapingRunID:    &s.scrapingRunID,
					SearchTermID:     &searchTermID,
					MarketplaceID:    marketplaceID,
				}
				if dbErr := s.database.CreateSentEmail(context.Background(), sentEmail); dbErr != nil {
					s.log(LogLevelWarning, "Failed to log sent email: %v", dbErr)
				}
			}
		}
	}()

	return nil
}
