package scraper

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureTable creates permits_{countyID} if it doesn't exist.
func EnsureTable(ctx context.Context, pool *pgxpool.Pool, countyID string, fields []string) error {
	tableName := tableNameFor(countyID)

	var extra strings.Builder
	for _, f := range fields {
		extra.WriteString(fmt.Sprintf("    %s TEXT,\n", sanitizeColName(f)))
	}

	createSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    id             SERIAL PRIMARY KEY,
    permit_number  TEXT NOT NULL,
    scraped_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    raw_data       JSONB NOT NULL DEFAULT '{}',
%s
    CONSTRAINT %s UNIQUE (permit_number)
);`, tableName, extra.String(), "uk_"+sanitizeColName(countyID)+"_permit")

	if _, err := pool.Exec(ctx, createSQL); err != nil {
		return err
	}

	idxName := "idx_" + sanitizeColName(countyID) + "_scraped_at"
	idxSQL := fmt.Sprintf(
		`CREATE INDEX IF NOT EXISTS %s ON %s(scraped_at DESC);`,
		idxName, tableName,
	)
	_, err := pool.Exec(ctx, idxSQL)
	return err
}

func tableNameFor(countyID string) string {
	return "permits_" + sanitizeColName(countyID)
}

// PermitTableName returns the PostgreSQL table name for a county's permits (sanitized).
func PermitTableName(countyID string) string {
	return tableNameFor(countyID)
}

func sanitizeColName(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "col"
	}
	return out
}
