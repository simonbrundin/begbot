-- Add only_enabled_products column to email_settings
ALTER TABLE email_settings ADD COLUMN IF NOT EXISTS only_enabled_products BOOLEAN DEFAULT true;

-- Update existing row to default to true
UPDATE email_settings SET only_enabled_products = true WHERE only_enabled_products IS NULL;
