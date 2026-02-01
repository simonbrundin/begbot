-- Migration: 001_initial_schema
-- Created: 2026-02-01
-- Description: Initial database schema for begbot trading platform

-- Core tables
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    brand TEXT,
    name TEXT,
    sell_packaging_cost INTEGER DEFAULT 0,
    sell_postage_cost INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS colors (
    id SERIAL PRIMARY KEY,
    name TEXT
);

CREATE TABLE IF NOT EXISTS conditions (
    id SMALLSERIAL PRIMARY KEY,
    title TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS marketplaces (
    id SMALLSERIAL PRIMARY KEY,
    name TEXT,
    link TEXT
);

CREATE TABLE IF NOT EXISTS transaction_types (
    id SERIAL PRIMARY KEY,
    name TEXT
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    date TIMESTAMPTZ DEFAULT NOW(),
    amount INTEGER,
    transaction_type INTEGER REFERENCES transaction_types(id)
);

CREATE TABLE IF NOT EXISTS trading_rules (
    id SERIAL PRIMARY KEY,
    min_profit_sek INTEGER,
    min_discount SMALLINT
);

CREATE TABLE IF NOT EXISTS traded_items (
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
    status_id SMALLINT DEFAULT 1,
    source_link TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    listing_id INTEGER
);

CREATE TABLE IF NOT EXISTS listings (
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
);

CREATE TABLE IF NOT EXISTS image_links (
    id SERIAL PRIMARY KEY,
    url TEXT,
    listing_id INTEGER REFERENCES listings(id) ON DELETE CASCADE
);

CREATE INDEX idx_transactions_type ON transactions(transaction_type);
CREATE INDEX idx_traded_items_status_id ON traded_items(status_id);
CREATE INDEX idx_listings_product_id ON listings(product_id);
CREATE INDEX idx_listings_status ON listings(status);
CREATE INDEX idx_listings_marketplace_id ON listings(marketplace_id);
CREATE INDEX idx_image_links_listing_id ON image_links(listing_id);

INSERT INTO trade_statuses (id, name) VALUES
    (1, 'potential'), (2, 'purchased'), (3, 'in_stock'),
    (4, 'listed'), (5, 'sold')
ON CONFLICT (id) DO NOTHING;
