CREATE TABLE IF NOT EXISTS scrape_runs (
    id              SERIAL PRIMARY KEY,
    county_id       TEXT NOT NULL REFERENCES county_connectors(county_id),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'running',
    records_found   INT DEFAULT 0,
    records_inserted INT DEFAULT 0,
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scrape_runs_county_id ON scrape_runs(county_id);
CREATE INDEX IF NOT EXISTS idx_scrape_runs_status    ON scrape_runs(status);
CREATE INDEX IF NOT EXISTS idx_scrape_runs_started_at ON scrape_runs(started_at DESC);
