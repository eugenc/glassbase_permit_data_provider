CREATE TABLE IF NOT EXISTS county_connectors (
    id              SERIAL PRIMARY KEY,
    county_id       TEXT NOT NULL UNIQUE,
    county_name     TEXT NOT NULL,
    state           TEXT NOT NULL,
    url             TEXT NOT NULL,
    source_type     TEXT NOT NULL,
    connector_config JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'active',
    last_generated_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_county_connectors_status ON county_connectors(status);
CREATE INDEX IF NOT EXISTS idx_county_connectors_state  ON county_connectors(state);
