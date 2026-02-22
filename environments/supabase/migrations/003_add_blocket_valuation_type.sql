-- Migration: 003_add_blocket_valuation_type
-- Created: 2026-02-22
-- Description: Add Blocket valuation type with enabled=true

-- Insert Blocket valuation type if it doesn't exist, or update to enabled
INSERT INTO valuation_types (name, enabled) VALUES ('Blocket', true)
ON CONFLICT (name) DO UPDATE SET enabled = true;
