-- Migration: 007_add_good_buys_column
-- Created: 2026-02-18
-- Description: Add column for tracking good buys (listings that met trading rules)

ALTER TABLE scraping_runs ADD COLUMN IF NOT EXISTS total_good_buys INTEGER DEFAULT 0;
