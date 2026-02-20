-- Migration: 008_add_weight_to_valuation_config
-- Created: 2026-02-20
-- Description: Add weight column to product_valuation_type_config so weights can be
--              stored per product per valuation type (percentages summing to 100 for
--              active types).  Existing rows default to 0 and will be normalised to
--              equal distribution the first time the config is saved via the API.

ALTER TABLE product_valuation_type_config ADD COLUMN IF NOT EXISTS weight NUMERIC NOT NULL DEFAULT 0;
