-- Migration: 006_add_scraping_runs
-- Created: 2026-02-18
-- Description: Add table for tracking scraping run history

CREATE TABLE IF NOT EXISTS scraping_runs (
    id SERIAL PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    total_ads_found INTEGER DEFAULT 0,
    total_listings_saved INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scraping_runs_started_at ON scraping_runs(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_scraping_runs_status ON scraping_runs(status);
