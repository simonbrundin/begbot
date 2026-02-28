-- Add min_ads column to auto_enable_settings
ALTER TABLE auto_enable_settings ADD COLUMN IF NOT EXISTS min_ads INTEGER DEFAULT 10;

-- Update existing row
UPDATE auto_enable_settings SET min_ads = 10 WHERE min_ads IS NULL;
