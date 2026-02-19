-- Migration: 005_add_products_columns
-- Created: 2026-02-18
-- Description: Add missing columns to products table

ALTER TABLE products ADD COLUMN IF NOT EXISTS category TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS model_variant TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS new_price INTEGER;
ALTER TABLE products ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT false;
