-- Migration script to add trade_statuses table and status_id column
-- Run this on existing databases before deploying the new code

-- Step 1: Create trade_statuses table
CREATE TABLE IF NOT EXISTS "trade_statuses" (
    "id" SMALLINT NOT NULL UNIQUE,
    "name" TEXT NOT NULL,
    PRIMARY KEY("id")
);

-- Step 2: Insert seed data
INSERT INTO "trade_statuses" ("id", "name") VALUES
    (1, 'potential'),
    (2, 'purchased'),
    (3, 'in_stock'),
    (4, 'listed'),
    (5, 'sold')
ON CONFLICT ("id") DO NOTHING;

-- Step 3: Add status_id column (without foreign key initially)
ALTER TABLE "traded_items" ADD COLUMN "status_id" SMALLINT DEFAULT 1;

-- Step 4: Migrate existing data from status text to status_id
UPDATE "traded_items" SET status_id = 1 WHERE status = 'potential';
UPDATE "traded_items" SET status_id = 2 WHERE status = 'purchased';
UPDATE "traded_items" SET status_id = 3 WHERE status = 'in_stock';
UPDATE "traded_items" SET status_id = 4 WHERE status = 'listed';
UPDATE "traded_items" SET status_id = 5 WHERE status = 'sold';

-- Step 5: Add foreign key constraint
ALTER TABLE "traded_items" ADD CONSTRAINT fk_traded_items_status_id
    FOREIGN KEY ("status_id") REFERENCES "trade_statuses"("id")
    ON UPDATE NO ACTION ON DELETE NO ACTION;

-- Step 6: Drop the old status column
ALTER TABLE "traded_items" DROP COLUMN "status";

-- Step 7: Create index
CREATE INDEX IF NOT EXISTS idx_traded_items_status_id ON traded_items(status_id);
