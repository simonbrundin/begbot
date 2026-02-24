-- Migration: Add new tracking columns to scraping_runs
-- Description: Adds columns for tracking new ads, new products, saved products, and emailed ads

DO $$ 
BEGIN
    -- Add new_ads column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'scraping_runs' AND column_name = 'new_ads'
    ) THEN
        ALTER TABLE scraping_runs ADD COLUMN new_ads INTEGER DEFAULT 0;
    END IF;

    -- Add new_products column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'scraping_runs' AND column_name = 'new_products'
    ) THEN
        ALTER TABLE scraping_runs ADD COLUMN new_products INTEGER DEFAULT 0;
    END IF;

    -- Add saved_products column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'scraping_runs' AND column_name = 'saved_products'
    ) THEN
        ALTER TABLE scraping_runs ADD COLUMN saved_products INTEGER DEFAULT 0;
    END IF;

    -- Add emailed_ads column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'scraping_runs' AND column_name = 'emailed_ads'
    ) THEN
        ALTER TABLE scraping_runs ADD COLUMN emailed_ads INTEGER DEFAULT 0;
    END IF;
END $$;
