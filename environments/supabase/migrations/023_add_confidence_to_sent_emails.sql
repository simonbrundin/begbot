-- Add confidence column to sent_emails
ALTER TABLE sent_emails ADD COLUMN IF NOT EXISTS confidence REAL;
