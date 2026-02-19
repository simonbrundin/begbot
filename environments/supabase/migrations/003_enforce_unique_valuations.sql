-- Migration: 003_enforce_unique_valuations
-- Created: 2026-02-19
-- Description: Remove duplicate valuations per (product_id, valuation_type_id)

-- Safety: delete duplicate rows keeping the latest by created_at (and id as tiebreaker),
-- then add a UNIQUE constraint on (product_id, valuation_type_id).

BEGIN;

-- Remove duplicate valuations for the same product and valuation type
WITH ranked AS (
  SELECT id,
         ROW_NUMBER() OVER (PARTITION BY product_id, valuation_type_id ORDER BY created_at DESC, id DESC) AS rn
  FROM valuations
  WHERE product_id IS NOT NULL AND valuation_type_id IS NOT NULL
)
DELETE FROM valuations
WHERE id IN (SELECT id FROM ranked WHERE rn > 1);

-- Add unique constraint to prevent future duplicates. NULL product_id values are allowed.
ALTER TABLE valuations
  ADD CONSTRAINT uniq_valuations_product_type UNIQUE (product_id, valuation_type_id);

COMMIT;
