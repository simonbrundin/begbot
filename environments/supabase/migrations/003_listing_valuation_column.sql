-- Migration: 003_listing_valuation_column
-- Created: 2026-02-03
-- Description: Add valuation column to listings for the compiled/summary valuation

ALTER TABLE listings ADD COLUMN IF NOT EXISTS valuation INTEGER;
