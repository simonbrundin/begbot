-- Migration: 010_add_shipping_insurance
-- Created: 2026-02-22
-- Description: Add shipping_insurance column to listings and buy_shipping_insurance to traded_items

ALTER TABLE listings ADD COLUMN IF NOT EXISTS shipping_insurance SMALLINT;
ALTER TABLE traded_items ADD COLUMN IF NOT EXISTS buy_shipping_insurance INTEGER DEFAULT 0;
