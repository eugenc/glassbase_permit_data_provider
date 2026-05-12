CREATE TABLE IF NOT EXISTS repair_runs (
    id              SERIAL PRIMARY KEY,
    county_id       TEXT NOT NULL,
    repair_trigger  TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'running',
    claude_output   TEXT,
    commit_sha      TEXT,
    pr_url          TEXT,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ,
    error_message   TEXT
);

CREATE INDEX IF NOT EXISTS idx_repair_runs_county_id  ON repair_runs(county_id);
CREATE INDEX IF NOT EXISTS idx_repair_runs_status     ON repair_runs(status);
CREATE INDEX IF NOT EXISTS idx_repair_runs_started_at ON repair_runs(started_at DESC);
