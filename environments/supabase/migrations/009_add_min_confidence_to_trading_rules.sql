-- Add min_confidence column to trading_rules table
ALTER TABLE trading_rules ADD COLUMN IF NOT EXISTS min_confidence SMALLINT DEFAULT 80;

-- Update existing rows to have default 80 if null
UPDATE trading_rules SET min_confidence = 80 WHERE min_confidence IS NULL;
