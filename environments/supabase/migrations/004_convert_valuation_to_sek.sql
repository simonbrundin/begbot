-- Migration: Convert listing valuation from öre to SEK
-- Description: Update existing valuation values to be stored in SEK instead of öre

UPDATE listings
SET valuation = valuation / 100
WHERE valuation IS NOT NULL AND valuation > 0;
