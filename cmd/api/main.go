package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"begbot/internal/api"
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
	db         *db.Postgres
	jobService *services.JobService
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

	server := &Server{db: database, jobService: services.NewJobService()}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", server.healthHandler)
	mux.HandleFunc("/api/inventory", server.inventoryHandler)
	mux.HandleFunc("/api/inventory/", server.inventoryItemHandler)
	mux.HandleFunc("/api/listings", server.listingsHandler)
	mux.HandleFunc("/api/listings/", server.listingItemHandler)
	mux.HandleFunc("/api/products", server.productsHandler)
	mux.HandleFunc("/api/products/", server.productItemHandler)
	mux.HandleFunc("/api/transactions", server.transactionsHandler)
	mux.HandleFunc("/api/transactions/", server.transactionItemHandler)
	mux.HandleFunc("/api/transaction-types", server.getTransactionTypes)
	mux.HandleFunc("/api/marketplaces", server.getMarketplaces)
	mux.HandleFunc("/api/search-terms", server.searchTermsHandler)
	mux.HandleFunc("/api/search-terms/", server.searchTermItemHandler)
	mux.HandleFunc("/api/fetch-ads", func(w http.ResponseWriter, r *http.Request) {
		server.fetchAdsHandlerWithConfig(w, r, cfg)
	})
	mux.HandleFunc("/api/fetch-ads/status/", server.fetchAdsStatusHandler)
	mux.HandleFunc("/api/fetch-ads/logs/", server.fetchAdsLogsHandler)
	mux.HandleFunc("/api/fetch-ads/cancel/", server.fetchAdsCancelHandler)
	mux.HandleFunc("/api/valuation-types", server.valuationTypesHandler)
	mux.HandleFunc("/api/valuations", server.valuationsHandler)
	mux.HandleFunc("/api/valuations/collect", server.collectValuationsHandler)
	mux.HandleFunc("/api/valuations/compiled", server.compiledValuationsHandler)

	headers := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) getInventory(w http.ResponseWriter, r *http.Request) {
	items, err := s.db.GetAllTradedItems(r.Context())
	if err != nil {
		api.WriteServerError(w, err.Error())
		return
	}
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
	listings, err := s.db.GetListingsWithProfit(ctx)
	if err != nil {
		logger.Printf("GetListingsWithProfit error: %v", err)
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
		w.WriteHeader(204)
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
		if err := rows.Scan(&p.ID, &p.Brand, &p.Name, &p.Category, &p.ModelVariant, &p.SellPackagingCost, &p.SellPostageCost, &p.NewPrice, &p.Enabled, &p.CreatedAt); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		products = append(products, p)
	}
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
		json.NewEncoder(w).Encode(product)
	default:
		api.WriteError(w, "Method not allowed", "METHOD_NOT_ALLOWED", 405)
	}
}

func (s *Server) productItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/products/"):]
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
		json.NewEncoder(w).Encode(product)
	case "DELETE":
		w.WriteHeader(204)
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
		json.NewEncoder(w).Encode(term)
	case "DELETE":
		if err := s.db.DeleteSearchTerm(r.Context(), id); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(204)
	}
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
		json.NewEncoder(w).Encode(valuations)
	case "POST":
		var v models.Valuation
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			api.WriteValidationError(w, []api.ValidationError{{Field: "body", Message: err.Error()}})
			return
		}
		if errs := api.ValidateNonNegative(int64(v.Valuation), "valuation"); len(errs) > 0 {
			api.WriteValidationError(w, errs)
			return
		}
		// For now, create valuations without listing ID (backward compatibility)
		if err := s.db.CreateValuation(ctx, &v, nil); err != nil {
			api.WriteServerError(w, err.Error())
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(v)
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

var _ = context.Background
