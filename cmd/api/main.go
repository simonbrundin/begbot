package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"begbot/internal/api"
	"begbot/internal/auth"
	"begbot/internal/config"
	db "begbot/internal/db"
	"begbot/internal/models"
	"begbot/internal/services"

	"github.com/joho/godotenv"
)

var logger *log.Logger

func init() {
	f, err := os.OpenFile("/home/simon/repos/begbot/fetch.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		logger = log.New(f, "", log.LstdFlags)
		logger.Println("=== LOG FILE INITIALIZED ===")
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
		logger.Printf("Warning: could not open log file: %v", err)
	}
}

type Server struct {
	db                   *db.Postgres
	jobService           *services.JobService
	searchHistoryService *services.SearchHistoryService
	scheduler            *services.Scheduler
	messagingService     *services.MessagingService
}

func main() {
	godotenv.Load()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.NewPostgres(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	logger.Println("Running database migrations...")
	if err := database.Migrate(); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	marketplaceService := services.NewMarketplaceService(cfg)
	cacheService := services.NewCacheService(cfg)
	llmService := services.NewLLMService(cfg)
	valuationService := services.NewValuationService(cfg, database, llmService)
	botService := services.NewBotService(cfg, marketplaceService, cacheService, llmService, valuationService, database)
	messagingService := services.NewMessagingService(cfg, database, llmService)

	scheduler := services.NewScheduler(database, cfg, botService)
	if err := scheduler.Start(context.Background()); err != nil {
		logger.Printf("Warning: Failed to start scheduler: %v", err)
	}

	server := &Server{
		db:                   database,
		jobService:           services.NewJobService(),
		searchHistoryService: services.NewSearchHistoryService(database),
		scheduler:            scheduler,
		messagingService:     messagingService,
	}

	// Initialize auth middleware
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "https://fxhknzpqhrkpqothjvrx.supabase.co"
	}
	supabaseAnonKey := os.Getenv("SUPABASE_KEY")
	authMiddleware := auth.NewAuthMiddleware(supabaseURL, supabaseAnonKey)

	mux := http.NewServeMux()

	// Health endpoint - no auth required
	mux.HandleFunc("/api/health", server.healthHandler)

	// Protected endpoints - wrapped with auth middleware
	mux.Handle("/api/inventory", authMiddleware.Middleware(http.HandlerFunc(server.inventoryHandler)))
	mux.Handle("/api/inventory/", authMiddleware.Middleware(http.HandlerFunc(server.inventoryItemHandler)))
	mux.Handle("/api/listings", authMiddleware.Middleware(http.HandlerFunc(server.listingsHandler)))
	mux.Handle("/api/listings/", authMiddleware.Middleware(http.HandlerFunc(server.listingItemHandler)))
	mux.Handle("/api/products", authMiddleware.Middleware(http.HandlerFunc(server.productsHandler)))
	mux.Handle("/api/products/", authMiddleware.Middleware(http.HandlerFunc(server.productItemHandler)))
	mux.Handle("/api/transactions", authMiddleware.Middleware(http.HandlerFunc(server.transactionsHandler)))
	mux.Handle("/api/transactions/", authMiddleware.Middleware(http.HandlerFunc(server.transactionItemHandler)))
	mux.Handle("/api/transaction-types", authMiddleware.Middleware(http.HandlerFunc(server.getTransactionTypes)))
	mux.Handle("/api/marketplaces", authMiddleware.Middleware(http.HandlerFunc(server.getMarketplaces)))
	mux.Handle("/api/search-terms", authMiddleware.Middleware(http.HandlerFunc(server.searchTermsHandler)))
	mux.Handle("/api/search-terms/", authMiddleware.Middleware(http.HandlerFunc(server.searchTermItemHandler)))
	// Scraping runs history (protected)
	mux.Handle("/api/scraping-runs", authMiddleware.Middleware(http.HandlerFunc(server.scrapingRunsHandler)))

	// Cron jobs management
	// Expose status endpoint without auth so the UI can poll running jobs
	// even when no user session is present (read-only, safe to be public).
	mux.HandleFunc("/api/cron-jobs/status", server.cronJobsStatusHandler)
	// Protected endpoints - require auth
	mux.Handle("/api/cron-jobs", authMiddleware.Middleware(http.HandlerFunc(server.cronJobsHandler)))
	mux.Handle("/api/cron-jobs/", authMiddleware.Middleware(http.HandlerFunc(server.cronJobItemHandler)))
	mux.Handle("/api/cron-jobs/cancel", authMiddleware.Middleware(http.HandlerFunc(server.cronJobsCancelHandler)))
	mux.Handle("/api/fetch-ads", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.fetchAdsHandlerWithConfig(w, r, cfg)
	})))
	mux.HandleFunc("/api/fetch-ads/status/", server.fetchAdsStatusHandler)
	mux.HandleFunc("/api/fetch-ads/logs/", server.fetchAdsLogsHandler)
	mux.HandleFunc("/api/fetch-ads/cancel/", server.fetchAdsCancelHandler)
	mux.HandleFunc("/api/valuation-types", server.valuationTypesHandler)
	mux.HandleFunc("/api/valuations", server.valuationsHandler)
	mux.HandleFunc("/api/valuations/", server.valuationItemHandler)
	mux.HandleFunc("/api/valuations/collect", server.collectValuationsHandler)
	mux.HandleFunc("/api/valuations/compiled", server.compiledValuationsHandler)
	mux.HandleFunc("/api/trading-rules", server.tradingRulesHandler)
	mux.HandleFunc("/api/conversations", server.conversationsHandler)
	mux.HandleFunc("/api/conversations/", server.conversationItemHandler)
	mux.HandleFunc("/api/messages", server.messagesHandler)
	mux.HandleFunc("/api/messages/", server.messageItemHandler)

	headers := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isLocalOrigin(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8081"
	}
	logger.Printf("API server starting on %s", addr)
	if err := http.ListenAndServe(addr, headers(mux)); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) getInventory(w http.ResponseWriter, r *http.Request) {
	items, err := s.db.GetAllTradedItems(r.Context())
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (s *Server) inventoryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getInventory(w, r)
	case "POST":
		var item models.TradedItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.CombineErrors(
			api.ValidateNonNegative(int64(item.BuyPrice), "buy_price"),
			api.ValidateNonNegative(int64(item.BuyShippingCost), "buy_shipping_cost"),
			api.ValidateNonNegative(int64(item.StatusID), "status_id"),
		); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		if err := s.db.SaveTradedItem(r.Context(), &item); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(item)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) inventoryItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/inventory/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid ID")
		return
	}

	switch r.Method {
	case "PUT":
		var item models.TradedItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if item.StatusID != 0 {
			if err := s.db.UpdateTradedItemStatus(r.Context(), id, item.StatusID); err != nil {
				api.WriteServerError(w, err.Error())
				return
			}
		}
		item.ID = id
		json.NewEncoder(w).Encode(item)
	case "DELETE":
		w.WriteHeader(204)
	}
}

func (s *Server) getListings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	potentialOnly := r.URL.Query().Get("potential") == "true" || r.URL.Query().Get("good-value") == "true"
	logger.Printf("getListings: potentialOnly=%v", potentialOnly)

	var listings []db.ListingWithProfit
	var err error

	if potentialOnly {
		listings, err = s.db.GetPotentialListings(ctx)
	} else {
		listings, err = s.db.GetListingsWithProfit(ctx)
	}
	if err != nil {
		logger.Printf("GetListings error: %v", err)
		api.WriteServerError(w, err.Error())
		return
	}
	mineOnly := r.URL.Query().Get("mine") == "true"
	if mineOnly {
		filtered := make([]db.ListingWithProfit, 0, len(listings))
		for _, l := range listings {
			if l.Listing.IsMyListing {
				filtered = append(filtered, l)
			}
		}
		listings = filtered
	}
	logger.Printf("Returning %d listings", len(listings))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

func (s *Server) listingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getListings(w, r)
	case "POST":
		var listing models.Listing
		if err := json.NewDecoder(r.Body).Decode(&listing); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.CombineErrors(
			api.ValidateRequired(listing.Title, "title"),
			api.ValidateRequired(listing.Link, "link"),
			api.ValidateRequired(listing.Status, "status"),
		); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		if err := s.db.SaveListing(r.Context(), &listing); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(listing)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) listingItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/listings/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid ID")
		return
	}

	switch r.Method {
	case "PUT":
		var listing models.Listing
		if err := json.NewDecoder(r.Body).Decode(&listing); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if listing.Status != "" {
			if err := s.db.UpdateListingStatus(r.Context(), id, listing.Status); err != nil {
				api.WriteServerError(w, err.Error())
				return
			}
		}
		listing.ID = id
		json.NewEncoder(w).Encode(listing)
	case "DELETE":
		if err := s.db.DeleteListing(r.Context(), id); err != nil {
			if err == sql.ErrNoRows {
				api.WriteNotFound(w, "listing not found")
				return
			}
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) getProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.DB().QueryContext(ctx, `SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, new_price, enabled, created_at FROM products ORDER BY created_at DESC`)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var brand, name, category, modelVariant sql.NullString
		var sellPackagingCost, sellPostageCost int
		var newPrice sql.NullInt64
		var enabled sql.NullBool
		var createdAt sql.NullTime

		if err := rows.Scan(
			&p.ID,
			&brand,
			&name,
			&category,
			&modelVariant,
			&sellPackagingCost,
			&sellPostageCost,
			&newPrice,
			&enabled,
			&createdAt,
		); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}

		if brand.Valid {
			p.Brand = &brand.String
		}
		if name.Valid {
			p.Name = &name.String
		}
		if category.Valid {
			p.Category = &category.String
		}
		if modelVariant.Valid {
			p.ModelVariant = &modelVariant.String
		}
		p.SellPackagingCost = sellPackagingCost
		p.SellPostageCost = sellPostageCost
		if newPrice.Valid {
			newPriceVal := int(newPrice.Int64)
			p.NewPrice = &newPriceVal
		}
		if enabled.Valid {
			p.Enabled = &enabled.Bool
		}
		if createdAt.Valid {
			p.CreatedAt = &createdAt.Time
		}

		products = append(products, p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (s *Server) productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getProducts(w, r)
	case "POST":
		var product models.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.CombineErrors(
			api.ValidateRequired(product.Brand, "brand"),
			api.ValidateRequired(product.Name, "name"),
			api.ValidateRequired(product.Category, "category"),
		); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		if err := s.db.SaveProduct(r.Context(), &product); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) productItemHandler(w http.ResponseWriter, r *http.Request) {
	pathSuffix := r.URL.Path[len("/api/products/"):]

	// Route: /api/products/{id}/valuation-type-config
	if strings.HasSuffix(pathSuffix, "/valuation-type-config") {
		idStr := strings.TrimSuffix(pathSuffix, "/valuation-type-config")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			api.WriteBadRequest(w, "Invalid ID")
			return
		}
		s.productValuationTypeConfigHandler(w, r, id)
		return
	}

	idStr := pathSuffix
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid ID")
		return
	}

	switch r.Method {
	case "PUT":
		var product models.Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		product.ID = id
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	case "DELETE":
		w.WriteHeader(204)
	}
}

func (s *Server) productValuationTypeConfigHandler(w http.ResponseWriter, r *http.Request, productID int64) {
	ctx := r.Context()
	switch r.Method {
	case "GET":
		configs, err := s.db.GetProductValuationTypeConfigs(ctx, productID)
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		if configs == nil {
			configs = []models.ProductValuationTypeConfig{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	case "PUT":
		var payload struct {
			Configs []models.ProductValuationTypeConfig `json:"configs"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		configs := payload.Configs
		// Validate at least one active type
		activeCount := 0
		for _, c := range configs {
			if c.IsActive {
				activeCount++
			}
		}
		if len(configs) > 0 && activeCount == 0 {
			api.WriteBadRequest(w, "Minst en värderingstyp måste vara aktiv")
			return
		}
		if err := s.db.UpsertProductValuationTypeConfigs(ctx, productID, configs); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(configs)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) tradingRulesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rules, err := s.db.GetTradingRules(r.Context())
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		api.WriteSuccess(w, rules)
		return
	case "PUT", "POST":
		var payload models.Economics
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		// Basic validation: non-negative values
		if payload.MinProfitSEK != nil && *payload.MinProfitSEK < 0 {
			api.WriteValidationError(w, []api.ValidationError{{Field: "min_profit_sek", Message: "must be non-negative"}})
			return
		}
		if payload.MinDiscount != nil && *payload.MinDiscount < 0 {
			api.WriteValidationError(w, []api.ValidationError{{Field: "min_discount", Message: "must be non-negative"}})
			return
		}
		if err := s.db.SaveTradingRules(r.Context(), &payload); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		// Return current rules from DB
		rules, err := s.db.GetTradingRules(r.Context())
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		api.WriteSuccess(w, rules)
		return
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) getTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.DB().QueryContext(ctx, `SELECT id, date, amount, transaction_type FROM transactions ORDER BY date DESC`)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.Date, &t.Amount, &t.TransactionType); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		transactions = append(transactions, t)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (s *Server) transactionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getTransactions(w, r)
	case "POST":
		var transaction models.Transaction
		if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.CombineErrors(
			api.ValidateNonNegative(int64(transaction.Amount), "amount"),
		); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		if err := s.db.SaveTransaction(r.Context(), &transaction); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) transactionItemHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		w.WriteHeader(204)
	}
}

func (s *Server) getTransactionTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.DB().QueryContext(ctx, `SELECT id, name FROM transaction_types`)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	defer rows.Close()

	var types []models.TransactionType
	for rows.Next() {
		var t models.TransactionType
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		types = append(types, t)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

func (s *Server) getMarketplaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.DB().QueryContext(ctx, `SELECT id, name, link FROM marketplaces`)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	defer rows.Close()

	var marketplaces []models.Marketplace
	for rows.Next() {
		var m models.Marketplace
		if err := rows.Scan(&m.ID, &m.Name, &m.Link); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		marketplaces = append(marketplaces, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marketplaces)
}

func (s *Server) getSearchTerms(w http.ResponseWriter, r *http.Request) {
	terms, err := s.db.GetAllSearchTerms(r.Context())
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	logger.Printf("Returning %d search terms", len(terms))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terms)
}

func (s *Server) searchTermsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getSearchTerms(w, r)
	case "POST":
		var term models.SearchTerm
		if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.CombineErrors(
			api.ValidateRequired(term.Description, "description"),
			api.ValidateRequired(term.URL, "url"),
		); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		if err := s.db.SaveSearchTerm(r.Context(), &term); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(term)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) searchTermItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/search-terms/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid ID")
		return
	}

	switch r.Method {
	case "PUT":
		var term models.SearchTerm
		if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if err := s.db.UpdateSearchTermStatus(r.Context(), id, term.IsActive); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(term)
	case "DELETE":
		if err := s.db.DeleteSearchTerm(r.Context(), id); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(204)
	}
}

func (s *Server) cronJobsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		jobs, err := s.db.GetAllCronJobs(r.Context())
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	case "POST":
		var job models.CronJob
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.CombineErrors(
			api.ValidateRequired(job.Name, "name"),
			api.ValidateRequired(job.CronExpression, "cron_expression"),
		); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		if err := s.db.CreateCronJob(r.Context(), &job); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		if s.scheduler != nil {
			s.scheduler.RefreshJobs(r.Context())
		}
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) cronJobItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/cron-jobs/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid ID")
		return
	}

	switch r.Method {
	case "GET":
		job, err := s.db.GetCronJobByID(r.Context(), id)
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		if job == nil {
			api.WriteError(w, "Not found", "NOT_FOUND", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	case "PUT":
		var job models.CronJob
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		job.ID = id
		if err := s.db.UpdateCronJob(r.Context(), &job); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		if s.scheduler != nil {
			s.scheduler.RefreshJobs(r.Context())
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	case "DELETE":
		if err := s.db.DeleteCronJob(r.Context(), id); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		if s.scheduler != nil {
			s.scheduler.RefreshJobs(r.Context())
		}
		w.WriteHeader(204)
	}
}

func (s *Server) cronJobsStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger.Printf("cronJobsStatusHandler invoked: method=%s remote=%s auth=%s", r.Method, r.RemoteAddr, r.Header.Get("Authorization"))
	if r.Method != "GET" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	var runningJobs []int64
	if s.scheduler != nil {
		running := s.scheduler.GetRunningJobs()
		for id := range running {
			runningJobs = append(runningJobs, id)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running_jobs": runningJobs,
	})
}

func (s *Server) cronJobsCancelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	var request struct {
		JobID int64 `json:"job_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.WriteBadRequest(w, "Invalid request body")
		return
	}

	if request.JobID == 0 {
		api.WriteBadRequest(w, "job_id is required")
		return
	}

	success := false
	if s.scheduler != nil {
		success = s.scheduler.CancelJob(request.JobID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
	})
}

func (s *Server) searchHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	history, count, err := s.searchHistoryService.GetHistory(r.Context(), page, pageSize)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	type PaginatedResponse struct {
		Data       []models.SearchHistory `json:"data"`
		TotalCount int                    `json:"total_count"`
		Page       int                    `json:"page"`
		PageSize   int                    `json:"page_size"`
		TotalPages int                    `json:"total_pages"`
	}

	totalPages := count / pageSize
	if count%pageSize > 0 {
		totalPages++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedResponse{
		Data:       history,
		TotalCount: count,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (s *Server) scrapingRunsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	offset := (page - 1) * pageSize
	runs, err := s.db.GetScrapingRuns(r.Context(), pageSize, offset)
	if err != nil {
		logger.Printf("ERROR GetScrapingRuns: %v", err)
		api.WriteServerError(w, err.Error())
		return
	}

	count, err := s.db.GetScrapingRunsCount(r.Context())
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	type PaginatedResponse struct {
		Data       []models.ScrapingRun `json:"data"`
		TotalCount int                  `json:"total_count"`
		Page       int                  `json:"page"`
		PageSize   int                  `json:"page_size"`
		TotalPages int                  `json:"total_pages"`
	}

	totalPages := count / pageSize
	if count%pageSize > 0 {
		totalPages++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedResponse{
		Data:       runs,
		TotalCount: count,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

func (s *Server) compiledValuationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		api.WriteBadRequest(w, "product_id required")
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "invalid product_id")
		return
	}

	// Get compiled valuation for product
	// Note: This would require the valuation service to be injected into the server
	// For now, return a stub response with basic data
	valuations, err := s.db.GetValuationsByProductID(context.Background(), productID)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	api.WriteSuccess(w, map[string]interface{}{
		"product_id": productID,
		"valuations": valuations,
		"compiled_result": map[string]interface{}{
			"recommended_price": 0,
			"confidence":        0.0,
			"reasoning":         "Compiled valuation not fully implemented in API yet",
		},
	})
}

func (s *Server) fetchAdsHandler(w http.ResponseWriter, r *http.Request) {
	// This will be called with config in main function
	api.WriteError(w, "Internal server error", "INTERNAL_ERROR", 500)
	return
}

func (s *Server) fetchAdsHandlerWithConfig(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method != "POST" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	jobID := generateJobID()
	job := s.jobService.CreateJob(jobID)

	go func() {
		log.Printf("Starting fetch job %s", jobID)
		marketplaceService := services.NewMarketplaceService(cfg)
		cacheService := services.NewCacheService(cfg)
		llmService := services.NewLLMService(cfg)
		valuationService := services.NewValuationService(cfg, s.db, llmService)
		botService := services.NewBotServiceWithJob(cfg, marketplaceService, cacheService, llmService, valuationService, s.db, s.jobService, jobID)

		if err := botService.Run(); err != nil {
			log.Printf("Job %s failed: %v", jobID, err)
			s.jobService.FailJob(jobID, err.Error())
		} else {
			log.Printf("Job %s completed successfully", jobID)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": job.ID,
		"status": job.Status,
	})
}

func (s *Server) fetchAdsStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract jobID from path, handling both "/api/fetch-ads/status/ID" and "/api/fetch-ads/statusID"
	path := r.URL.Path
	prefix := "/api/fetch-ads/status"
	jobID := strings.TrimPrefix(path, prefix)
	jobID = strings.TrimPrefix(jobID, "/")
	if jobID == "" {
		api.WriteBadRequest(w, "Job ID required")
		return
	}

	job := s.jobService.GetJob(jobID)
	if job == nil {
		api.WriteNotFound(w, "Job")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":            job.ID,
		"status":            job.Status,
		"progress":          job.Progress,
		"total_queries":     job.TotalQueries,
		"completed_queries": job.CompletedQueries,
		"current_query":     job.CurrentQuery,
		"ads_found":         job.AdsFound,
		"error":             job.Error,
		"started_at":        job.StartedAt,
		"completed_at":      job.CompletedAt,
	})
}

func (s *Server) fetchAdsLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	// Extract jobID from path, handling both "/api/fetch-ads/logs/ID" and "/api/fetch-ads/logsID"
	path := r.URL.Path
	prefix := "/api/fetch-ads/logs"
	jobID := strings.TrimPrefix(path, prefix)
	jobID = strings.TrimPrefix(jobID, "/")
	if jobID == "" {
		api.WriteBadRequest(w, "Job ID required")
		return
	}

	job := s.jobService.GetJob(jobID)
	if job == nil {
		api.WriteNotFound(w, "Job")
		return
	}

	logger.Printf("SSE connection for job %s from %s", jobID, r.RemoteAddr)

	// Set SSE headers - CORS is handled by middleware
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Flush headers immediately
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Send existing logs first
	logs := s.jobService.GetLogs(jobID)
	for _, log := range logs {
		data, _ := json.Marshal(log)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	// Subscribe to new logs
	logChan := s.jobService.SubscribeToLogs(jobID)
	if logChan == nil {
		return
	}
	defer s.jobService.UnsubscribeFromLogs(jobID, logChan)

	// Stream new logs until job completes or client disconnects
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case log, ok := <-logChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(log)
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func (s *Server) fetchAdsCancelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	// Extract jobID from path
	path := r.URL.Path
	prefix := "/api/fetch-ads/cancel"
	jobID := strings.TrimPrefix(path, prefix)
	jobID = strings.TrimPrefix(jobID, "/")
	if jobID == "" {
		api.WriteBadRequest(w, "Job ID required")
		return
	}

	// Check if job exists
	job := s.jobService.GetJob(jobID)
	if job == nil {
		api.WriteNotFound(w, "Job")
		return
	}

	// Try to cancel the job
	if cancelled := s.jobService.CancelJob(jobID); !cancelled {
		api.WriteError(w, "Job cannot be cancelled", "INVALID_STATE", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": "cancelled",
	})
}

func generateJobID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *Server) valuationTypesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	types, err := s.db.GetValuationTypes(ctx)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

func (s *Server) valuationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case "GET":
		productIDStr := r.URL.Query().Get("product_id")
		if productIDStr == "" {
			api.WriteBadRequest(w, "product_id required")
			return
		}
		productID, err := strconv.ParseInt(productIDStr, 10, 64)
		if err != nil {
			api.WriteBadRequest(w, "invalid product_id")
			return
		}
		valuations, err := s.db.GetValuationsByProductID(ctx, productID)
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(valuations)
	case "POST":
		var v models.Valuation
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}

		logger.Printf("POST /api/valuations payload=%+v", v)
		if errs := api.ValidateNonNegative(int64(v.Valuation), "valuation"); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		// For now, create valuations without listing ID (backward compatibility)
		if err := s.db.CreateValuation(ctx, &v, nil); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}

		logger.Printf("POST /api/valuations created id=%d product_id=%d valuation=%d", v.ID, v.ProductID, v.Valuation)
		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) valuationItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/valuations/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "Invalid ID")
		return
	}

	switch r.Method {
	case "PUT":
		var payload struct {
			Valuation int `json:"valuation"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}

		logger.Printf("PUT /api/valuations/%d payload=%+v", id, payload)
		if errs := api.ValidateNonNegative(int64(payload.Valuation), "valuation"); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		rows, err := s.db.UpdateValuation(r.Context(), id, payload.Valuation)
		if err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		logger.Printf("PUT /api/valuations/%d rows_affected=%d", id, rows)
		if rows == 0 {
			api.WriteError(w, "Not found", "NOT_FOUND", 404)
			return
		}
		api.WriteSuccess(w, map[string]interface{}{"id": id, "valuation": payload.Valuation, "rows_affected": rows})
	case "DELETE":
		w.WriteHeader(204)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) collectValuationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	type CollectValuationRequest struct {
		ProductID   int64  `json:"product_id"`
		ProductInfo string `json:"product_info"`
	}

	var req CollectValuationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
		return
	}

	api.WriteSuccess(w, map[string]interface{}{
		"message":    "Valuation collection not fully implemented in API yet",
		"product_id": req.ProductID,
	})
}

// Conversation handlers
func (s *Server) conversationsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getConversations(w, r)
	case "POST":
		s.createConversation(w, r)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) getConversations(w http.ResponseWriter, r *http.Request) {
	needsReview := r.URL.Query().Get("needs_review") == "true"

	var conversations []models.ConversationWithDetails
	var err error

	if needsReview {
		conversations, err = s.db.GetConversationsNeedingReview(r.Context())
	} else {
		conversations, err = s.db.GetAllConversations(r.Context())
	}

	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(conversations)
}

func (s *Server) createConversation(w http.ResponseWriter, r *http.Request) {
	type CreateConversationRequest struct {
		ListingID     int64 `json:"listing_id"`
		MarketplaceID int64 `json:"marketplace_id"`
	}

	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
		return
	}

	conv, err := s.messagingService.CreateConversation(r.Context(), req.ListingID, req.MarketplaceID)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(conv)
}

func (s *Server) conversationItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
	parts := strings.Split(idStr, "/")

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "id", Message: "invalid ID"}})
		return
	}

	if len(parts) > 1 && parts[1] == "messages" {
		s.getConversationMessages(w, r, id)
		return
	}

	switch r.Method {
	case "GET":
		s.getConversation(w, r, id)
	case "PUT":
		s.updateConversation(w, r, id)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) getConversation(w http.ResponseWriter, r *http.Request, id int64) {
	conv, err := s.db.GetConversationByID(r.Context(), id)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	if conv == nil {
		api.WriteError(w, "Conversation not found", "NOT_FOUND", 404)
		return
	}
	json.NewEncoder(w).Encode(conv)
}

func (s *Server) updateConversation(w http.ResponseWriter, r *http.Request, id int64) {
	type UpdateConversationRequest struct {
		Status string `json:"status"`
	}

	var req UpdateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
		return
	}

	if err := s.db.UpdateConversationStatus(r.Context(), id, req.Status); err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	api.WriteSuccess(w, map[string]interface{}{"status": "updated"})
}

func (s *Server) getConversationMessages(w http.ResponseWriter, r *http.Request, conversationID int64) {
	messages, err := s.db.GetMessagesByConversationID(r.Context(), conversationID)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	json.NewEncoder(w).Encode(messages)
}

// Message handlers
func (s *Server) messagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		s.createMessage(w, r)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	type CreateMessageRequest struct {
		ListingID      *int64 `json:"listing_id,omitempty"`
		ConversationID *int64 `json:"conversation_id,omitempty"`
		MessageType    string `json:"message_type"` // "initial", "reply", or "incoming"
		Content        string `json:"content,omitempty"`
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
		return
	}

	var msg *models.Message
	var err error

	switch req.MessageType {
	case "initial":
		if req.ListingID == nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "listing_id", Message: "required for initial messages"}})
			return
		}
		msg, err = s.messagingService.GenerateInitialMessage(r.Context(), *req.ListingID)

	case "reply":
		if req.ConversationID == nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "conversation_id", Message: "required for reply messages"}})
			return
		}
		msg, err = s.messagingService.GenerateReplyMessage(r.Context(), *req.ConversationID)

	case "incoming":
		if req.ConversationID == nil || req.Content == "" {
			api.WriteValidationError(w, []api.ValidationError{{Field: "conversation_id", Message: "required"}, {Field: "content", Message: "required"}})
			return
		}
		msg, err = s.messagingService.ReceiveMessage(r.Context(), *req.ConversationID, req.Content)

	default:
		api.WriteValidationError(w, []api.ValidationError{{Field: "message_type", Message: "must be 'initial', 'reply', or 'incoming'"}})
		return
	}

	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(msg)
}

func (s *Server) messageItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/messages/")
	parts := strings.Split(idStr, "/")

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "id", Message: "invalid ID"}})
		return
	}

	if len(parts) > 1 {
		action := parts[1]
		switch action {
		case "approve":
			s.approveMessage(w, r, id)
			return
		case "reject":
			s.rejectMessage(w, r, id)
			return
		}
	}

	switch r.Method {
	case "GET":
		s.getMessage(w, r, id)
	case "PUT":
		s.updateMessage(w, r, id)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) getMessage(w http.ResponseWriter, r *http.Request, id int64) {
	msg, err := s.db.GetMessageByID(r.Context(), id)
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
	if msg == nil {
		api.WriteError(w, "Message not found", "NOT_FOUND", 404)
		return
	}
	json.NewEncoder(w).Encode(msg)
}

func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request, id int64) {
	type UpdateMessageRequest struct {
		Content string `json:"content"`
	}

	var req UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
		return
	}

	if err := s.db.UpdateMessageContent(r.Context(), id, req.Content); err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	api.WriteSuccess(w, map[string]interface{}{"status": "updated"})
}

func (s *Server) approveMessage(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "POST" && r.Method != "PUT" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	if err := s.db.ApproveMessage(r.Context(), id); err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	api.WriteSuccess(w, map[string]interface{}{"status": "approved"})
}

func (s *Server) rejectMessage(w http.ResponseWriter, r *http.Request, id int64) {
	if r.Method != "POST" && r.Method != "PUT" {
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
		return
	}

	if err := s.db.RejectMessage(r.Context(), id); err != nil {
		api.WriteServerError(w, err.Error())
		return
	}

	api.WriteSuccess(w, map[string]interface{}{"status": "rejected"})
}

// isLocalOrigin returns true for origins from localhost or private network IPs,
// which is needed when dev.nu opens the browser via the LAN IP.
func isLocalOrigin(origin string) bool {
	host := strings.TrimPrefix(strings.TrimPrefix(origin, "https://"), "http://")
	if i := strings.LastIndex(host, ":"); i != -1 {
		host = host[:i]
	}
	return host == "localhost" ||
		strings.HasPrefix(host, "127.") ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.")
}

var _ = context.Background
