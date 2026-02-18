package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"begbot/internal/config"
	"begbot/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(cfg config.DatabaseConfig) (*Postgres, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=10",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}

func (p *Postgres) Migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			brand TEXT,
			name TEXT,
			category TEXT,
			model_variant TEXT,
			sell_packaging_cost INTEGER DEFAULT 0,
			sell_postage_cost INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS colors (
			id SERIAL PRIMARY KEY,
			name TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS conditions (
			id SMALLINT PRIMARY KEY,
			title TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS trade_statuses (
			id SMALLINT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS marketplaces (
			id SMALLINT PRIMARY KEY,
			name TEXT,
			link TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS transaction_types (
			id SERIAL PRIMARY KEY,
			name TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			date TIMESTAMPTZ DEFAULT NOW(),
			amount INTEGER,
			transaction_type INTEGER REFERENCES transaction_types(id)
		)`,
		`CREATE TABLE IF NOT EXISTS trading_rules (
			id SERIAL PRIMARY KEY,
			min_profit_sek INTEGER,
			min_discount SMALLINT
		)`,
		`CREATE TABLE IF NOT EXISTS traded_items (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id),
			storage SMALLINT,
			color_id INTEGER REFERENCES colors(id),
			buy_price INTEGER,
			buy_shipping_cost INTEGER DEFAULT 0,
			buy_transaction_id INTEGER REFERENCES transactions(id),
			buy_date TIMESTAMPTZ,
			sell_price INTEGER,
			sell_packaging_cost INTEGER DEFAULT 0,
			sell_postage_cost INTEGER DEFAULT 0,
			sell_shipping_collected INTEGER DEFAULT 0,
			sell_transaction_id INTEGER REFERENCES transactions(id),
			sell_date TIMESTAMPTZ,
			status_id SMALLINT REFERENCES trade_statuses(id) DEFAULT 1,
			source_link TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			listing_id INTEGER REFERENCES listings(id)
		)`,
		`CREATE TABLE IF NOT EXISTS listings (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id),
			price INTEGER,
			link TEXT,
			condition_id SMALLINT REFERENCES conditions(id),
			shipping_cost SMALLINT,
			title TEXT,
			description TEXT,
			marketplace_id SMALLINT REFERENCES marketplaces(id),
			status TEXT DEFAULT 'draft',
			publication_date TIMESTAMPTZ,
			sold_date TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			is_my_listing BOOLEAN DEFAULT FALSE
		)`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS title TEXT`,
		`CREATE TABLE IF NOT EXISTS image_links (
			id SERIAL PRIMARY KEY,
			url TEXT,
			listing_id INTEGER REFERENCES listings(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_traded_items_product_id ON traded_items(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_traded_items_status_id ON traded_items(status_id)`,
		`CREATE INDEX IF NOT EXISTS idx_traded_items_listing_id ON traded_items(listing_id)`,
		`CREATE INDEX IF NOT EXISTS idx_listings_product_id ON listings(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_listings_status ON listings(status)`,
		`CREATE INDEX IF NOT EXISTS idx_listings_marketplace_id ON listings(marketplace_id)`,
		`CREATE TABLE IF NOT EXISTS search_terms (
			id SERIAL PRIMARY KEY,
			description TEXT,
			url TEXT,
			marketplace_id SMALLINT REFERENCES marketplaces(id),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_search_terms_marketplace_id ON search_terms(marketplace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_search_terms_is_active ON search_terms(is_active)`,
		`CREATE TABLE IF NOT EXISTS search_history (
			id SERIAL PRIMARY KEY,
			search_term_id INTEGER REFERENCES search_terms(id),
			search_term_desc TEXT,
			url TEXT,
			results_found INTEGER DEFAULT 0,
			new_ads_found INTEGER DEFAULT 0,
			marketplace_id SMALLINT REFERENCES marketplaces(id),
			marketplace_name TEXT,
			searched_at TIMESTAMPTZ DEFAULT NOW(),
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_search_history_search_term_id ON search_history(search_term_id)`,
		`CREATE INDEX IF NOT EXISTS idx_search_history_searched_at ON search_history(searched_at DESC)`,
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS category TEXT`,
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS model_variant TEXT`,
		`UPDATE products SET category = 'phone' WHERE category IS NULL`,
		`INSERT INTO trade_statuses (id, name) VALUES
			(1, 'potential'),
			(2, 'purchased'),
			(3, 'in_stock'),
			(4, 'listed'),
			(5, 'sold')
		ON CONFLICT (id) DO NOTHING`,
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT FALSE`,
		`CREATE TABLE IF NOT EXISTS valuation_types (
			id SMALLINT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS valuations (
			id SERIAL PRIMARY KEY,
			product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
			listing_id INTEGER REFERENCES listings(id) ON DELETE CASCADE,
			valuation_type_id SMALLINT REFERENCES valuation_types(id),
			valuation INTEGER NOT NULL,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_valuations_product_id ON valuations(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_valuations_type_id ON valuations(valuation_type_id)`,
		`INSERT INTO valuation_types (id, name) VALUES
			(1, 'Egen databas'),
			(2, 'Tradera'),
			(3, 'eBay'),
			(4, 'Nypris (LLM)')
		ON CONFLICT (id) DO NOTHING`,
		`ALTER TABLE valuations ADD COLUMN IF NOT EXISTS listing_id INTEGER REFERENCES listings(id) ON DELETE CASCADE`,
		`CREATE INDEX IF NOT EXISTS idx_valuations_listing_id ON valuations(listing_id)`,
		`UPDATE valuations SET listing_id = NULL WHERE listing_id IS NULL`,
		`CREATE TABLE IF NOT EXISTS scraping_runs (
			id SERIAL PRIMARY KEY,
			started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ,
			status VARCHAR(20) NOT NULL DEFAULT 'running',
			total_ads_found INTEGER DEFAULT 0,
			total_listings_saved INTEGER DEFAULT 0,
			error_message TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_scraping_runs_started_at ON scraping_runs(started_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_scraping_runs_status ON scraping_runs(status)`,
	}

	for i, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			log.Printf("Migration failed at query %d: %v", i, err)
			log.Printf("Query: %s", query)
			return fmt.Errorf("migration failed at query %d: %w", i, err)
		}
	}

	return nil
}

func (p *Postgres) SaveTradedItem(ctx context.Context, item *models.TradedItem) error {
	query := `
		INSERT INTO traded_items (
			product_id, storage, color_id,
			buy_price, buy_shipping_cost, buy_transaction_id, buy_date,
			sell_price, sell_packaging_cost, sell_postage_cost, sell_shipping_collected,
			sell_transaction_id, sell_date, status_id, source_link, listing_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`
	return p.db.QueryRowContext(ctx, query,
		item.ProductID, item.Storage, item.ColorID,
		item.BuyPrice, item.BuyShippingCost, item.BuyTransactionID, item.BuyDate,
		item.SellPrice, item.SellPackagingCost, item.SellPostageCost, item.SellShippingCollected,
		item.SellTransactionID, item.SellDate, item.StatusID, item.SourceLink, item.ListingID,
	).Scan(&item.ID)
}

func (p *Postgres) UpdateTradedItemStatus(ctx context.Context, id int64, statusID int16) error {
	query := `UPDATE traded_items SET status_id = $1 WHERE id = $2`
	_, err := p.db.ExecContext(ctx, query, statusID, id)
	return err
}

func (p *Postgres) GetTradedItemByID(ctx context.Context, id int64) (*models.TradedItem, error) {
	query := `
		SELECT id, product_id, storage, color_id,
			buy_price, buy_shipping_cost, buy_transaction_id, buy_date,
			sell_price, sell_packaging_cost, sell_postage_cost, sell_shipping_collected,
			sell_transaction_id, sell_date, status_id, source_link, created_at, listing_id
		FROM traded_items WHERE id = $1
	`
	var item models.TradedItem
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.ProductID, &item.Storage, &item.ColorID,
		&item.BuyPrice, &item.BuyShippingCost, &item.BuyTransactionID, &item.BuyDate,
		&item.SellPrice, &item.SellPackagingCost, &item.SellPostageCost, &item.SellShippingCollected,
		&item.SellTransactionID, &item.SellDate, &item.StatusID, &item.SourceLink, &item.CreatedAt, &item.ListingID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (p *Postgres) GetActiveTradedItems(ctx context.Context) ([]models.TradedItem, error) {
	query := `
		SELECT id, product_id, storage, color_id,
			buy_price, buy_shipping_cost, buy_transaction_id, buy_date,
			sell_price, sell_packaging_cost, sell_postage_cost, sell_shipping_collected,
			sell_transaction_id, sell_date, status_id, source_link, created_at, listing_id
		FROM traded_items WHERE status_id IN (2, 3)
		ORDER BY created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanTradedItems(rows)
}

func (p *Postgres) GetAllTradedItems(ctx context.Context) ([]models.TradedItem, error) {
	query := `
		SELECT id, product_id, storage, color_id,
			buy_price, buy_shipping_cost, buy_transaction_id, buy_date,
			sell_price, sell_packaging_cost, sell_postage_cost, sell_shipping_collected,
			sell_transaction_id, sell_date, status_id, source_link, created_at, listing_id
		FROM traded_items
		ORDER BY created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanTradedItems(rows)
}

func (p *Postgres) GetSoldTradedItems(ctx context.Context, limit int) ([]models.TradedItem, error) {
	query := `
		SELECT id, product_id, storage, color_id,
			buy_price, buy_shipping_cost, buy_transaction_id, buy_date,
			sell_price, sell_packaging_cost, sell_postage_cost, sell_shipping_collected,
			sell_transaction_id, sell_date, status_id, source_link, created_at, listing_id
		FROM traded_items WHERE status_id = 5
		ORDER BY sell_date DESC
		LIMIT $1
	`
	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return p.scanTradedItems(rows)
}

func (p *Postgres) SaveListing(ctx context.Context, listing *models.Listing) error {
	query := `
		INSERT INTO listings (product_id, price, valuation, link, condition_id, shipping_cost, title, description, marketplace_id, status, publication_date, sold_date, is_my_listing, eligible_for_shipping, seller_pays_shipping, buy_now)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`
	return p.db.QueryRowContext(ctx, query,
		listing.ProductID, listing.Price, listing.Valuation, listing.Link, listing.ConditionID, listing.ShippingCost,
		listing.Title, listToNullString(listing.Description), listing.MarketplaceID, listing.Status, listing.PublicationDate, listing.SoldDate, listing.IsMyListing,
		listing.EligibleForShipping, listing.SellerPaysShipping, listing.BuyNow,
	).Scan(&listing.ID)
}

func listToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func (p *Postgres) SaveImageLinks(ctx context.Context, listingID int64, urls []string) error {
	query := `INSERT INTO image_links (listing_id, url) VALUES ($1, $2)`
	for _, url := range urls {
		if _, err := p.db.ExecContext(ctx, query, listingID, url); err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) UpdateListingStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE listings SET status = $1 WHERE id = $2`
	_, err := p.db.ExecContext(ctx, query, status, id)
	return err
}

func (p *Postgres) DeleteListing(ctx context.Context, id int64) error {
	result, err := p.db.ExecContext(ctx, `DELETE FROM listings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (p *Postgres) GetListingByProductID(ctx context.Context, productID int64) (*models.Listing, error) {
	query := `
		SELECT id, product_id, price, valuation, link, condition_id, shipping_cost, title, description,
			marketplace_id, status, publication_date, sold_date, created_at, is_my_listing,
			eligible_for_shipping, seller_pays_shipping, buy_now
		FROM listings WHERE product_id = $1 AND status = 'active'
	`
	var listing models.Listing
	err := p.db.QueryRowContext(ctx, query, productID).Scan(
		&listing.ID, &listing.ProductID, &listing.Price, &listing.Valuation, &listing.Link, &listing.ConditionID,
		&listing.ShippingCost, &listing.Title, &listing.Description, &listing.MarketplaceID, &listing.Status,
		&listing.PublicationDate, &listing.SoldDate, &listing.CreatedAt, &listing.IsMyListing,
		&listing.EligibleForShipping, &listing.SellerPaysShipping, &listing.BuyNow,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &listing, nil
}

func (p *Postgres) SaveTransaction(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (date, amount, transaction_type)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return p.db.QueryRowContext(ctx, query, transaction.Date, transaction.Amount, transaction.TransactionType).Scan(&transaction.ID)
}

func (p *Postgres) SaveProduct(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, new_price, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	return p.db.QueryRowContext(ctx, query, product.Brand, product.Name, product.Category, product.ModelVariant, product.SellPackagingCost, product.SellPostageCost, product.NewPrice, product.Enabled).Scan(&product.ID)
}

func (p *Postgres) GetProductByName(ctx context.Context, brand, name string) (*models.Product, error) {
	query := `
		SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, new_price, enabled, created_at
		FROM products WHERE brand = $1 AND name = $2
	`
	var product models.Product
	err := p.db.QueryRowContext(ctx, query, brand, name).Scan(
		&product.ID, &product.Brand, &product.Name, &product.Category, &product.ModelVariant, &product.SellPackagingCost, &product.SellPostageCost, &product.NewPrice, &product.Enabled, &product.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *Postgres) GetProductByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
		SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, new_price, enabled, created_at
		FROM products WHERE id = $1
	`
	var product models.Product
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Brand, &product.Name, &product.Category, &product.ModelVariant, &product.SellPackagingCost, &product.SellPostageCost, &product.NewPrice, &product.Enabled, &product.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *Postgres) FindProduct(ctx context.Context, brand, name, category string) (*models.Product, error) {
	query := `
		SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, new_price, enabled, created_at
		FROM products WHERE brand = $1 AND name = $2 AND category = $3
	`
	var product models.Product
	err := p.db.QueryRowContext(ctx, query, brand, name, category).Scan(
		&product.ID, &product.Brand, &product.Name, &product.Category, &product.ModelVariant, &product.SellPackagingCost, &product.SellPostageCost, &product.NewPrice, &product.Enabled, &product.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *Postgres) GetOrCreateProduct(ctx context.Context, brand, name, category string, modelVariant *string, packagingCost, postageCost int) (*models.Product, error) {
	product, err := p.GetProductByName(ctx, brand, name)
	if err != nil {
		return nil, err
	}
	if product != nil {
		return product, nil
	}

	newProduct := models.Product{
		Brand:             &brand,
		Name:              &name,
		Category:          &category,
		ModelVariant:      modelVariant,
		SellPackagingCost: packagingCost,
		SellPostageCost:   postageCost,
	}
	if err := p.SaveProduct(ctx, &newProduct); err != nil {
		return nil, err
	}
	return &newProduct, nil
}

func (p *Postgres) CalculateProfit(item *models.TradedItem) int {
	sellTotal := 0
	if item.SellPrice != nil {
		sellTotal += *item.SellPrice
	}
	if item.SellShippingCollected != nil {
		sellTotal += *item.SellShippingCollected
	}

	buyTotal := item.BuyPrice + item.BuyShippingCost
	sellCost := 0
	if item.SellPackagingCost != nil {
		sellCost += *item.SellPackagingCost
	}
	if item.SellPostageCost != nil {
		sellCost += *item.SellPostageCost
	}

	return sellTotal - (buyTotal + sellCost)
}

func (p *Postgres) scanTradedItems(rows *sql.Rows) ([]models.TradedItem, error) {
	var items []models.TradedItem
	for rows.Next() {
		var item models.TradedItem
		if err := rows.Scan(
			&item.ID, &item.ProductID, &item.Storage, &item.ColorID,
			&item.BuyPrice, &item.BuyShippingCost, &item.BuyTransactionID, &item.BuyDate,
			&item.SellPrice, &item.SellPackagingCost, &item.SellPostageCost, &item.SellShippingCollected,
			&item.SellTransactionID, &item.SellDate, &item.StatusID, &item.SourceLink, &item.CreatedAt, &item.ListingID,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (p *Postgres) SaveSearchTerm(ctx context.Context, term *models.SearchTerm) error {
	query := `
		INSERT INTO search_terms (description, url, marketplace_id, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return p.db.QueryRowContext(ctx, query, term.Description, term.URL, term.MarketplaceID, term.IsActive).Scan(&term.ID, &term.CreatedAt, &term.UpdatedAt)
}

func (p *Postgres) GetAllSearchTerms(ctx context.Context) ([]models.SearchTerm, error) {
	query := `
		SELECT id, description, url, marketplace_id, is_active, created_at, updated_at
		FROM search_terms
		ORDER BY created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []models.SearchTerm
	for rows.Next() {
		var term models.SearchTerm
		if err := rows.Scan(&term.ID, &term.Description, &term.URL, &term.MarketplaceID, &term.IsActive, &term.CreatedAt, &term.UpdatedAt); err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}
	return terms, rows.Err()
}

func (p *Postgres) GetActiveSearchTerms(ctx context.Context) ([]models.SearchTerm, error) {
	query := `
		SELECT id, description, url, marketplace_id, is_active, created_at, updated_at
		FROM search_terms
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []models.SearchTerm
	for rows.Next() {
		var term models.SearchTerm
		if err := rows.Scan(&term.ID, &term.Description, &term.URL, &term.MarketplaceID, &term.IsActive, &term.CreatedAt, &term.UpdatedAt); err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}
	return terms, rows.Err()
}

func (p *Postgres) GetSearchTermByID(ctx context.Context, id int64) (*models.SearchTerm, error) {
	query := `
		SELECT id, description, url, marketplace_id, is_active, created_at, updated_at
		FROM search_terms WHERE id = $1
	`
	var term models.SearchTerm
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&term.ID, &term.Description, &term.URL, &term.MarketplaceID, &term.IsActive, &term.CreatedAt, &term.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &term, nil
}

func (p *Postgres) UpdateSearchTermStatus(ctx context.Context, id int64, isActive bool) error {
	query := `UPDATE search_terms SET is_active = $1, updated_at = NOW() WHERE id = $2`
	_, err := p.db.ExecContext(ctx, query, isActive, id)
	return err
}

func (p *Postgres) DeleteSearchTerm(ctx context.Context, id int64) error {
	query := `DELETE FROM search_terms WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, id)
	return err
}

func (p *Postgres) SaveSearchHistory(ctx context.Context, h *models.SearchHistory) error {
	query := `
		INSERT INTO search_history (search_term_id, search_term_desc, url, results_found, new_ads_found, marketplace_id, marketplace_name, searched_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	return p.db.QueryRowContext(ctx, query,
		h.SearchTermID, h.SearchTermDesc, h.URL, h.ResultsFound, h.NewAdsFound,
		h.MarketplaceID, h.MarketplaceName, h.SearchedAt,
	).Scan(&h.ID, &h.CreatedAt)
}

func (p *Postgres) GetSearchHistory(ctx context.Context, limit, offset int) ([]models.SearchHistory, error) {
	query := `
		SELECT id, search_term_id, search_term_desc, url, results_found, new_ads_found, 
		       marketplace_id, marketplace_name, searched_at, created_at
		FROM search_history
		ORDER BY searched_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := p.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.SearchHistory
	for rows.Next() {
		var h models.SearchHistory
		if err := rows.Scan(&h.ID, &h.SearchTermID, &h.SearchTermDesc, &h.URL, &h.ResultsFound, &h.NewAdsFound, &h.MarketplaceID, &h.MarketplaceName, &h.SearchedAt, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

func (p *Postgres) GetSearchHistoryCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM search_history`
	var count int
	err := p.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (p *Postgres) ListingExistsByLink(ctx context.Context, link string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM listings WHERE link = $1)`
	var exists bool
	err := p.db.QueryRowContext(ctx, query, link).Scan(&exists)
	return exists, err
}

func (p *Postgres) GetAllListings(ctx context.Context) ([]models.Listing, error) {
	query := `
		SELECT id, product_id, price, valuation, link, condition_id, shipping_cost, title, description,
			marketplace_id, status, publication_date, sold_date, created_at, is_my_listing,
			eligible_for_shipping, seller_pays_shipping, buy_now
		FROM listings
		ORDER BY created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []models.Listing
	for rows.Next() {
		var listing models.Listing
		var title, description sql.NullString
		err := rows.Scan(
			&listing.ID, &listing.ProductID, &listing.Price, &listing.Valuation, &listing.Link, &listing.ConditionID,
			&listing.ShippingCost, &title, &description, &listing.MarketplaceID, &listing.Status,
			&listing.PublicationDate, &listing.SoldDate, &listing.CreatedAt, &listing.IsMyListing,
			&listing.EligibleForShipping, &listing.SellerPaysShipping, &listing.BuyNow,
		)
		if err != nil {
			return nil, err
		}
		if title.Valid {
			listing.Title = title.String
		}
		if description.Valid {
			listing.Description = &description.String
		}
		listings = append(listings, listing)
	}
	return listings, rows.Err()
}

func (p *Postgres) GetListingCount(ctx context.Context) (int, error) {
	var count int
	err := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM listings").Scan(&count)
	return count, err
}

func (p *Postgres) GetMarketplaceByID(ctx context.Context, id int64) (*models.Marketplace, error) {
	query := `SELECT id, name, link FROM marketplaces WHERE id = $1`
	var m models.Marketplace
	err := p.db.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.Name, &m.Link)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (p *Postgres) GetValuationTypes(ctx context.Context) ([]models.ValuationType, error) {
	query := `SELECT id, name FROM valuation_types ORDER BY id`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []models.ValuationType
	for rows.Next() {
		var vt models.ValuationType
		if err := rows.Scan(&vt.ID, &vt.Name); err != nil {
			return nil, err
		}
		types = append(types, vt)
	}
	return types, rows.Err()
}

func (p *Postgres) GetValuationsByProductID(ctx context.Context, productID int64) ([]models.Valuation, error) {
	query := `
		SELECT id, product_id, valuation_type_id, valuation, COALESCE(metadata, '{}'::jsonb), created_at
		FROM valuations
		WHERE product_id = $1
		ORDER BY created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var valuations []models.Valuation
	for rows.Next() {
		var v models.Valuation
		var metadata []byte
		if err := rows.Scan(&v.ID, &v.ProductID, &v.ValuationTypeID, &v.Valuation, &metadata, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Metadata = json.RawMessage(metadata)
		valuations = append(valuations, v)
	}
	return valuations, rows.Err()
}

type ValuationWithType struct {
	ID              int64  `json:"id"`
	ValuationTypeID int16  `json:"valuation_type_id"`
	ValuationType   string `json:"valuation_type"`
	Valuation       int    `json:"valuation"`
}

func (p *Postgres) GetValuationsWithTypesByProductID(ctx context.Context, productID int64) ([]ValuationWithType, error) {
	query := `
		SELECT v.id, v.valuation_type_id, vt.name, v.valuation
		FROM valuations v
		JOIN valuation_types vt ON v.valuation_type_id = vt.id
		WHERE v.product_id = $1
		ORDER BY v.created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var valuations []ValuationWithType
	for rows.Next() {
		var v ValuationWithType
		if err := rows.Scan(&v.ID, &v.ValuationTypeID, &v.ValuationType, &v.Valuation); err != nil {
			return nil, err
		}
		valuations = append(valuations, v)
	}
	return valuations, rows.Err()
}

func (p *Postgres) GetValuationsWithTypesByListingID(ctx context.Context, listingID int64) ([]ValuationWithType, error) {
	query := `
		SELECT v.id, v.valuation_type_id, vt.name, v.valuation
		FROM valuations v
		JOIN valuation_types vt ON v.valuation_type_id = vt.id
		WHERE v.listing_id = $1
		ORDER BY v.created_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var valuations []ValuationWithType
	for rows.Next() {
		var v ValuationWithType
		if err := rows.Scan(&v.ID, &v.ValuationTypeID, &v.ValuationType, &v.Valuation); err != nil {
			return nil, err
		}
		valuations = append(valuations, v)
	}
	return valuations, rows.Err()
}

func (p *Postgres) GetValuationsForListing(ctx context.Context, listingID int64) ([]ValuationWithType, error) {
	// First get the listing to get the product ID
	listing, err := p.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found")
	}

	// Get listing-specific valuations (these take priority)
	listingVals, err := p.GetValuationsWithTypesByListingID(ctx, listingID)
	if err != nil {
		return nil, err
	}

	// For product-wide valuations, get only the latest one per type
	var productVals []ValuationWithType
	if listing != nil && listing.ProductID != nil {
		productVals, err = p.GetLatestValuationByTypeForProduct(ctx, *listing.ProductID)
		if err != nil {
			return nil, err
		}
	}

	// Combine and ensure only one valuation per type, preferring listing-specific
	result := make([]ValuationWithType, 0)
	valuationMap := make(map[int16]ValuationWithType)

	// First add all listing-specific valuations
	for _, v := range listingVals {
		valuationMap[v.ValuationTypeID] = v
	}

	// Then add product-wide valuations only if type not already covered
	for _, v := range productVals {
		if _, exists := valuationMap[v.ValuationTypeID]; !exists {
			valuationMap[v.ValuationTypeID] = v
		}
	}

	// Convert map to slice
	for _, v := range valuationMap {
		result = append(result, v)
	}

	return result, nil
}

func (p *Postgres) GetListingByID(ctx context.Context, id int64) (*models.Listing, error) {
	query := `
		SELECT id, product_id, price, valuation, link, condition_id, shipping_cost, title, description,
			marketplace_id, status, publication_date, sold_date, created_at, is_my_listing,
			eligible_for_shipping, seller_pays_shipping, buy_now
		FROM listings WHERE id = $1
	`
	var listing models.Listing
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&listing.ID, &listing.ProductID, &listing.Price, &listing.Valuation, &listing.Link, &listing.ConditionID,
		&listing.ShippingCost, &listing.Title, &listing.Description, &listing.MarketplaceID, &listing.Status,
		&listing.PublicationDate, &listing.SoldDate, &listing.CreatedAt, &listing.IsMyListing,
		&listing.EligibleForShipping, &listing.SellerPaysShipping, &listing.BuyNow,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &listing, nil
}

func (p *Postgres) CreateValuation(ctx context.Context, v *models.Valuation, listingID *int64) error {
	query := `
		INSERT INTO valuations (product_id, listing_id, valuation_type_id, valuation, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := p.db.QueryRowContext(ctx, query, v.ProductID, listingID, v.ValuationTypeID, v.Valuation, v.Metadata).
		Scan(&v.ID, &v.CreatedAt)
	return err
}

type ListingWithValuations struct {
	Listing    models.Listing
	Product    *models.Product
	Valuations []ValuationWithType
}

func (p *Postgres) GetListingsWithValuations(ctx context.Context) ([]ListingWithValuations, error) {
	listings, err := p.GetAllListings(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ListingWithValuations, 0, len(listings))
	for _, l := range listings {
		listingWithV := ListingWithValuations{Listing: l}
		if l.ProductID != nil {
			vals, err := p.GetValuationsForListing(ctx, l.ID)
			if err != nil {
				return nil, err
			}
			listingWithV.Valuations = vals

			product, err := p.GetProductByID(ctx, *l.ProductID)
			if err != nil {
				return nil, err
			}
			listingWithV.Product = product
		}
		result = append(result, listingWithV)
	}
	return result, nil
}

func (p *Postgres) GetLatestValuationByTypeForProduct(ctx context.Context, productID int64) ([]ValuationWithType, error) {
	// For now, return an empty slice to avoid duplicate valuations issue
	// TODO: Fix this properly by either:
	// 1. Deleting duplicate valuations from database, or
	// 2. Implementing proper deduplication logic
	return []ValuationWithType{}, nil
}

func (p *Postgres) GetTradingRules(ctx context.Context) (*models.Economics, error) {
	query := `SELECT id, min_profit_sek, min_discount FROM trading_rules LIMIT 1`
	var rules models.Economics
	err := p.db.QueryRowContext(ctx, query).Scan(&rules.ID, &rules.MinProfitSEK, &rules.MinDiscount)
	if err == sql.ErrNoRows {
		fmt.Println("GetTradingRules: No rules found in database, using defaults")
		return &models.Economics{
			MinProfitSEK: intPtr(0),
			MinDiscount:  intPtr(0),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	fmt.Printf("GetTradingRules: id=%d, min_profit_sek=%v, min_discount=%v\n", rules.ID, rules.MinProfitSEK, rules.MinDiscount)
	return &rules, nil
}

func intPtr(i int) *int {
	return &i
}

type ListingWithProfit struct {
	Listing         models.Listing
	Product         *models.Product
	Valuations      []ValuationWithType
	PotentialProfit int
	DiscountPercent float64
}

func (p *Postgres) GetListingsWithProfit(ctx context.Context) ([]ListingWithProfit, error) {
	listings, err := p.GetAllListings(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ListingWithProfit, 0, len(listings))

	for _, l := range listings {
		listingWithP := ListingWithProfit{Listing: l}

		if l.Price != nil {
			shippingCost := 0
			if l.ShippingCost != nil {
				shippingCost = *l.ShippingCost
			}
			profit := l.Valuation - *l.Price - shippingCost
			listingWithP.PotentialProfit = profit

			if l.Valuation > 0 {
				listingWithP.DiscountPercent = float64(profit) / float64(l.Valuation) * 100
			}
		}

		if l.ProductID != nil {
			vals, err := p.GetValuationsForListing(ctx, l.ID)
			if err != nil {
				return nil, err
			}
			listingWithP.Valuations = vals

			product, err := p.GetProductByID(ctx, *l.ProductID)
			if err != nil {
				return nil, err
			}
			listingWithP.Product = product
		}
		result = append(result, listingWithP)
	}
	return result, nil
}

func (p *Postgres) GetPotentialListings(ctx context.Context) ([]ListingWithProfit, error) {
	listings, err := p.GetAllListings(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ListingWithProfit, 0, len(listings))

	rules, err := p.GetTradingRules(ctx)
	if err != nil {
		return nil, err
	}
	minProfit := 0
	if rules.MinProfitSEK != nil {
		minProfit = *rules.MinProfitSEK
	}
	minDiscount := 0
	if rules.MinDiscount != nil {
		minDiscount = *rules.MinDiscount
	}
	_ = minProfit
	_ = minDiscount

	for _, l := range listings {
		listingWithP := ListingWithProfit{Listing: l}

		if l.Price != nil && l.Valuation > 0 {
			shippingCost := 0
			if l.ShippingCost != nil {
				shippingCost = *l.ShippingCost
			}
			profit := l.Valuation - *l.Price - shippingCost
			discountPercent := float64(profit) / float64(l.Valuation) * 100

			if profit >= minProfit && discountPercent >= float64(minDiscount) {
				listingWithP.PotentialProfit = profit
				listingWithP.DiscountPercent = discountPercent
				fmt.Printf("Listing %d passes: profit=%d, discount=%.1f%%\n", l.ID, profit, discountPercent)
			} else {
				continue
			}
		} else {
			continue
		}

		if l.ProductID != nil {
			vals, err := p.GetValuationsForListing(ctx, l.ID)
			if err != nil {
				return nil, err
			}
			listingWithP.Valuations = vals

			product, err := p.GetProductByID(ctx, *l.ProductID)
			if err != nil {
				return nil, err
			}
			listingWithP.Product = product
		}
		result = append(result, listingWithP)
	}
	return result, nil
}

func (p *Postgres) SaveScrapingRun(ctx context.Context, run *models.ScrapingRun) error {
	query := `
		INSERT INTO scraping_runs (started_at, completed_at, status, total_ads_found, total_listings_saved, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return p.db.QueryRowContext(ctx, query,
		run.StartedAt, run.CompletedAt, run.Status, run.TotalAdsFound, run.TotalListingsSaved, run.ErrorMessage,
	).Scan(&run.ID, &run.CreatedAt)
}

func (p *Postgres) UpdateScrapingRun(ctx context.Context, run *models.ScrapingRun) error {
	query := `
		UPDATE scraping_runs 
		SET completed_at = $1, status = $2, total_ads_found = $3, total_listings_saved = $4, error_message = $5
		WHERE id = $6
	`
	_, err := p.db.ExecContext(ctx, query,
		run.CompletedAt, run.Status, run.TotalAdsFound, run.TotalListingsSaved, run.ErrorMessage, run.ID,
	)
	return err
}

func (p *Postgres) GetScrapingRuns(ctx context.Context, limit, offset int) ([]models.ScrapingRun, error) {
	query := `
		SELECT id, started_at, completed_at, status, total_ads_found, total_listings_saved, total_good_buys, error_message, created_at
		FROM scraping_runs
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := p.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.ScrapingRun
	for rows.Next() {
		var run models.ScrapingRun
		if err := rows.Scan(&run.ID, &run.StartedAt, &run.CompletedAt, &run.Status, &run.TotalAdsFound, &run.TotalListingsSaved, &run.TotalGoodBuys, &run.ErrorMessage, &run.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (p *Postgres) GetScrapingRunsCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM scraping_runs`
	var count int
	err := p.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
