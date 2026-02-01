package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"begbot/internal/config"
	"begbot/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(cfg config.DatabaseConfig) (*Postgres, error) {
	encodedPassword := url.QueryEscape(cfg.Password)
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=10",
		cfg.User, encodedPassword, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
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
			description TEXT,
			marketplace_id SMALLINT REFERENCES marketplaces(id),
			status TEXT DEFAULT 'draft',
			publication_date TIMESTAMPTZ,
			sold_date TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			is_my_listing BOOLEAN DEFAULT FALSE
		)`,
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
		INSERT INTO listings (product_id, price, link, condition_id, shipping_cost, description, marketplace_id, status, publication_date, sold_date, is_my_listing)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`
	return p.db.QueryRowContext(ctx, query,
		listing.ProductID, listing.Price, listing.Link, listing.ConditionID, listing.ShippingCost,
		listing.Description, listing.MarketplaceID, listing.Status, listing.PublicationDate, listing.SoldDate, listing.IsMyListing,
	).Scan(&listing.ID)
}

func (p *Postgres) UpdateListingStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE listings SET status = $1 WHERE id = $2`
	_, err := p.db.ExecContext(ctx, query, status, id)
	return err
}

func (p *Postgres) GetListingByProductID(ctx context.Context, productID int64) (*models.Listing, error) {
	query := `
		SELECT id, product_id, price, link, condition_id, shipping_cost, description,
			marketplace_id, status, publication_date, sold_date, created_at, is_my_listing
		FROM listings WHERE product_id = $1 AND status = 'active'
	`
	var listing models.Listing
	err := p.db.QueryRowContext(ctx, query, productID).Scan(
		&listing.ID, &listing.ProductID, &listing.Price, &listing.Link, &listing.ConditionID,
		&listing.ShippingCost, &listing.Description, &listing.MarketplaceID, &listing.Status,
		&listing.PublicationDate, &listing.SoldDate, &listing.CreatedAt, &listing.IsMyListing,
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
		INSERT INTO products (brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return p.db.QueryRowContext(ctx, query, product.Brand, product.Name, product.Category, product.ModelVariant, product.SellPackagingCost, product.SellPostageCost, product.Enabled).Scan(&product.ID)
}

func (p *Postgres) GetProductByName(ctx context.Context, brand, name string) (*models.Product, error) {
	query := `
		SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, enabled, created_at
		FROM products WHERE brand = $1 AND name = $2
	`
	var product models.Product
	err := p.db.QueryRowContext(ctx, query, brand, name).Scan(
		&product.ID, &product.Brand, &product.Name, &product.Category, &product.ModelVariant, &product.SellPackagingCost, &product.SellPostageCost, &product.Enabled, &product.CreatedAt,
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
		SELECT id, brand, name, category, model_variant, sell_packaging_cost, sell_postage_cost, enabled, created_at
		FROM products WHERE brand = $1 AND name = $2 AND category = $3
	`
	var product models.Product
	err := p.db.QueryRowContext(ctx, query, brand, name, category).Scan(
		&product.ID, &product.Brand, &product.Name, &product.Category, &product.ModelVariant, &product.SellPackagingCost, &product.SellPostageCost, &product.Enabled, &product.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *Postgres) GetOrCreateProduct(ctx context.Context, brand, name, category, modelVariant string, packagingCost, postageCost int) (*models.Product, error) {
	product, err := p.GetProductByName(ctx, brand, name)
	if err != nil {
		return nil, err
	}
	if product != nil {
		return product, nil
	}

	newProduct := models.Product{
		Brand:             brand,
		Name:              name,
		Category:          category,
		ModelVariant:      modelVariant,
		SellPackagingCost: packagingCost,
		SellPostageCost:   postageCost,
		Enabled:           false,
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

func (p *Postgres) ListingExistsByLink(ctx context.Context, link string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM listings WHERE link = $1)`
	var exists bool
	err := p.db.QueryRowContext(ctx, query, link).Scan(&exists)
	return exists, err
}

func (p *Postgres) GetAllListings(ctx context.Context) ([]models.Listing, error) {
	query := `
		SELECT id, product_id, price, link, condition_id, shipping_cost, description,
			marketplace_id, status, publication_date, sold_date, created_at, is_my_listing
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
		err := rows.Scan(
			&listing.ID, &listing.ProductID, &listing.Price, &listing.Link, &listing.ConditionID,
			&listing.ShippingCost, &listing.Description, &listing.MarketplaceID, &listing.Status,
			&listing.PublicationDate, &listing.SoldDate, &listing.CreatedAt, &listing.IsMyListing,
		)
		if err != nil {
			return nil, err
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
