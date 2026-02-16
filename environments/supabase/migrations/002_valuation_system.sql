-- Migration: 002_valuation_system
-- Created: 2026-02-03
-- Description: Valuation system tables for product pricing and valuation types

-- Valuation types: different sources of valuation data
CREATE TABLE IF NOT EXISTS valuation_types (
    id SMALLSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Insert default valuation types
INSERT INTO valuation_types (name) VALUES
    ('Egen databas'),
    ('Tradera'),
    ('eBay'),
    ('Nypris (LLM)')
ON CONFLICT (name) DO NOTHING;

-- Valuations: individual valuation entries for products
CREATE TABLE IF NOT EXISTS valuations (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    valuation_type_id SMALLINT REFERENCES valuation_types(id),
    valuation INTEGER NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient lookups
CREATE INDEX idx_valuations_product_id ON valuations(product_id);
CREATE INDEX idx_valuations_type_id ON valuations(valuation_type_id);
CREATE INDEX idx_valuations_created_at ON valuations(created_at DESC);
