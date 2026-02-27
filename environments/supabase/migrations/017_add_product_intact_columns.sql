-- Add columns for product intactness check
ALTER TABLE listings ADD COLUMN IF NOT EXISTS is_intact BOOLEAN;
ALTER TABLE listings ADD COLUMN IF NOT EXISTS intact_check_reasoning TEXT;
