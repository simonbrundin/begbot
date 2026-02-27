-- Create scraping_run_logs table to store logs from scraping runs
CREATE TABLE IF NOT EXISTS scraping_run_logs (
    id BIGSERIAL PRIMARY KEY,
    scraping_run_id BIGINT NOT NULL REFERENCES scraping_runs(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL CHECK (level IN ('info', 'warning', 'error')),
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fetching logs by scraping run
CREATE INDEX idx_scraping_run_logs_run_id ON scraping_run_logs(scraping_run_id);

-- Index for cleanup queries by date
CREATE INDEX idx_scraping_run_logs_created_at ON scraping_run_logs(created_at);
