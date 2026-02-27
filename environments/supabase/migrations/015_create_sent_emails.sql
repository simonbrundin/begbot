-- Migration: Create sent_emails table
-- Description: Creates table to track sent emails for trading rule matches

-- Create sent_emails table if it doesn't exist
CREATE TABLE IF NOT EXISTS sent_emails (
    id SERIAL PRIMARY KEY,
    listing_id INTEGER REFERENCES listings(id) ON DELETE SET NULL,
    listing_title TEXT NOT NULL,
    listing_link TEXT NOT NULL,
    listing_price INTEGER,
    listing_valuation INTEGER,
    profit INTEGER,
    discount_percent REAL,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_name TEXT,
    brand TEXT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scraping_run_id INTEGER REFERENCES scraping_runs(id) ON DELETE SET NULL,
    search_term_id INTEGER REFERENCES search_terms(id) ON DELETE SET NULL,
    marketplace_id SMALLINT REFERENCES marketplaces(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_sent_emails_sent_at ON sent_emails(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_sent_emails_listing_id ON sent_emails(listing_id);
CREATE INDEX IF NOT EXISTS idx_sent_emails_product_id ON sent_emails(product_id);
CREATE INDEX IF NOT EXISTS idx_sent_emails_scraping_run_id ON sent_emails(scraping_run_id);
